// Package font provides high-quality font rendering with texture atlas
package font

import (
	"crypto/sha256"
	_ "embed"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"sync"
	"unsafe"

	"github.com/energye/examples/lcl/gpui/core/gl"
	"github.com/energye/examples/lcl/gpui/core/math"

	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/sfnt"
	xfont "golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

//go:embed fonts/NotoSansCJK-Regular.ttc
var embeddedFontData []byte

// EmbeddedFontData returns the embedded default font data (Noto Sans CJK Regular).
// It covers Latin, CJK Unified Ideographs (Chinese, Japanese, Korean), Kana, and Hangul.
// Returns nil if the font file was not embedded at build time.
func EmbeddedFontData() []byte {
	return embeddedFontData
}

// GlyphInfo stores glyph metrics and atlas position
type GlyphInfo struct {
	// UV coordinates (0-1)
	U0, V0, U1, V1 float32
	// Metrics
	Advance  float32
	Width    float32
	Height   float32
	BearingX float32
	BearingY float32
}

// Font represents a font with texture atlas
type Font struct {
	mu sync.RWMutex

	texture    uint32
	texWidth   int
	texHeight  int
	atlas      *image.RGBA
	glyphs     map[rune]*GlyphInfo
	fontSize   float64
	lineHeight float32
	ascent     float32
	descent    float32
	letterGap  float32
	cellSize   int
	cols       int
	nextSlot   int
	// gpuSlot tracks how many glyphs have been uploaded to the GPU texture.
	// If gpuSlot < nextSlot, the GPU texture is stale and must be synced.
	gpuSlot    int
	style      FontStyle
	cacheKey   fontCacheKey
	fallbacks  []*Font

	// sfntFont is the parsed font (supports TTC/OTC/TTF/OTF via opentype.ParseCollection).
	sfntFont *opentype.Font
	// sfntBuf is a reusable buffer for sfnt operations (e.g. GlyphIndex).
	sfntBuf *sfnt.Buffer

	// Font face for on-demand glyph rasterization.
	face xfont.Face
}

const (
	// initialAtlasSize is the starting font atlas size (256x256 = 256KB RGBA).
	initialAtlasSize = 256
	// maxAtlasSize is the maximum atlas size before we stop growing.
	maxAtlasSize = 4096
	glyphPadding = 2
	maxGlyphs    = 4096
)

// FontStyle describes font style options.
type FontStyle struct {
	Size    float64
	Bold    bool
	Italic  bool
	DPI     float64
	Hinting xfont.Hinting
	// LetterSpacing adds extra pixels between glyph advances.
	LetterSpacing float32
}

type fontCacheKey struct {
	hash          [sha256.Size]byte
	size          float64
	dpi           float64
	bold          bool
	italic        bool
	hinting       xfont.Hinting
	letterSpacing float32
}

var globalFontCache = struct {
	sync.Mutex
	parsed map[[sha256.Size]byte]*opentype.Font
	fonts  map[fontCacheKey]*Font
}{
	parsed: make(map[[sha256.Size]byte]*opentype.Font),
	fonts:  make(map[fontCacheKey]*Font),
}

// NewFont creates a new font from TTF data
func NewFont(ttfData []byte, fontSize float64) (*Font, error) {
	return NewFontStyled(ttfData, FontStyle{
		Size:    fontSize,
		DPI:     96,
		Hinting: xfont.HintingFull,
	})
}

// NewFontStyled creates a new font with style options (bold/italic).
// It supports TTF, OTF, TTC, and OTC font formats.
func NewFontStyled(ttfData []byte, style FontStyle) (*Font, error) {
	style = normalizeFontStyle(style)
	key := cacheKeyFor(ttfData, style)

	globalFontCache.Lock()
	if cached := globalFontCache.fonts[key]; cached != nil {
		globalFontCache.Unlock()
		return cached, nil
	}
	globalFontCache.Unlock()

	// Parse font via opentype (supports TTF/OTF single fonts and TTC/OTC collections).
	parsedFont, err := parseFontAny(ttfData, key.hash)
	if err != nil {
		return nil, err
	}

	face, err := opentype.NewFace(parsedFont, &opentype.FaceOptions{
		Size:    style.Size,
		DPI:     style.DPI,
		Hinting: style.Hinting,
	})
	if err != nil {
		return nil, fmt.Errorf("create opentype face: %w", err)
	}

	// Calculate metrics
	m := face.Metrics()
	ascent := fixed26_6ToFloat32(m.Ascent)
	descent := fixed26_6ToFloat32(m.Descent)
	lineHeight := fixed26_6ToFloat32(m.Height)
	if lineHeight <= 0 {
		lineHeight = ascent + descent
	}
	cellSize := int(lineHeight+float32(style.Size*0.75)) + glyphPadding*4 + 4
	if cellSize < 8 {
		cellSize = 8
	}
	initSize := initialAtlasSize
	cols := initSize / cellSize
	if cols < 1 {
		cols = 1
	}

	fontObj := &Font{
		texWidth:   initSize,
		texHeight:  initSize,
		atlas:      image.NewRGBA(image.Rect(0, 0, initSize, initSize)),
		glyphs:     make(map[rune]*GlyphInfo),
		fontSize:   style.Size,
		lineHeight: lineHeight,
		ascent:     ascent,
		descent:    descent,
		letterGap:  style.LetterSpacing,
		cellSize:   cellSize,
		cols:       cols,
		style:      style,
		cacheKey:   key,
		sfntFont:   parsedFont,
		sfntBuf:    &sfnt.Buffer{},
		face:       face,
	}

	// No pre-rasterization — glyphs are rasterized on demand via GetGlyph/addGlyph.
	// Upload initial empty atlas to GPU so the texture exists for uploadRect later.
	_ = fontObj.uploadToGPU()

	globalFontCache.Lock()
	if cached := globalFontCache.fonts[key]; cached != nil {
		globalFontCache.Unlock()
		fontObj.Delete()
		return cached, nil
	}
	globalFontCache.fonts[key] = fontObj
	globalFontCache.Unlock()

	return fontObj, nil
}

// parseFontAny parses font data in any supported format (TTF, OTF, TTC, OTC).
// It first tries a single-font parse, then falls back to collection parsing.
func parseFontAny(data []byte, hash [sha256.Size]byte) (*opentype.Font, error) {
	// Check cache
	globalFontCache.Lock()
	if f := globalFontCache.parsed[hash]; f != nil {
		globalFontCache.Unlock()
		return f, nil
	}
	globalFontCache.Unlock()

	// Try single-font parse first
	f, err := opentype.Parse(data)
	if err == nil {
		globalFontCache.Lock()
		if cached := globalFontCache.parsed[hash]; cached != nil {
			globalFontCache.Unlock()
			return cached, nil
		}
		globalFontCache.parsed[hash] = f
		globalFontCache.Unlock()
		return f, nil
	}

	// Fall back to collection parse (TTC/OTC)
	coll, collErr := opentype.ParseCollection(data)
	if collErr != nil {
		return nil, fmt.Errorf("parse font: single=%v, collection=%v", err, collErr)
	}
	if coll.NumFonts() == 0 {
		return nil, fmt.Errorf("font collection is empty")
	}
	f, err = coll.Font(0)
	if err != nil {
		return nil, fmt.Errorf("get first font from collection: %w", err)
	}

	globalFontCache.Lock()
	if cached := globalFontCache.parsed[hash]; cached != nil {
		globalFontCache.Unlock()
		return cached, nil
	}
	globalFontCache.parsed[hash] = f
	globalFontCache.Unlock()
	return f, nil
}

// ValidateFontData verifies that font data is supported by the opentype parser.
func ValidateFontData(ttfData []byte) error {
	_, err := parseFontAny(ttfData, sha256.Sum256(ttfData))
	return err
}

// FontCoverageScore counts how many probe runes a font contains.
func FontCoverageScore(ttfData []byte, probes []rune) (int, error) {
	f, err := parseFontAny(ttfData, sha256.Sum256(ttfData))
	if err != nil {
		return 0, err
	}
	buf := &sfnt.Buffer{}
	score := 0
	for _, r := range probes {
		idx, err := f.GlyphIndex(buf, r)
		if err == nil && idx != 0 {
			score++
		}
	}
	return score, nil
}

func normalizeFontStyle(style FontStyle) FontStyle {
	if style.Size <= 0 {
		style.Size = 14
	}
	if style.DPI <= 0 {
		style.DPI = 96
	}
	return style
}

func cacheKeyFor(ttfData []byte, style FontStyle) fontCacheKey {
	return fontCacheKey{
		hash:          sha256.Sum256(ttfData),
		size:          style.Size,
		dpi:           style.DPI,
		bold:          style.Bold,
		italic:        style.Italic,
		hinting:       style.Hinting,
		letterSpacing: style.LetterSpacing,
	}
}

// rasterizeGlyph rasterizes a single glyph
func (f *Font) rasterizeGlyph(atlas *image.RGBA, r rune, slot int) (*GlyphInfo, image.Rectangle, bool) {
	col := slot % f.cols
	row := slot / f.cols

	cellX := col * f.cellSize
	cellY := row * f.cellSize
	cellRect := image.Rect(cellX, cellY, cellX+f.cellSize, cellY+f.cellSize).Intersect(atlas.Bounds())
	draw.Draw(atlas, cellRect, image.Transparent, image.Point{}, draw.Src)

	// Get glyph advance
	adv, ok := f.face.GlyphAdvance(r)
	if !ok {
		if r != ' ' {
			adv, ok = f.face.GlyphAdvance(' ')
			if !ok {
				return nil, image.Rectangle{}, false
			}
		} else {
			return nil, image.Rectangle{}, false
		}
	}

	// Get glyph image
	dr, mask, maskp, _, ok := f.face.Glyph(fixed.P(0, 0), r)
	if !ok {
		if r == ' ' {
			return &GlyphInfo{
				U0:      float32(cellX) / float32(f.texWidth),
				V0:      float32(cellY) / float32(f.texHeight),
				U1:      float32(cellX) / float32(f.texWidth),
				V1:      float32(cellY) / float32(f.texHeight),
				Advance: fixed26_6ToFloat32(adv),
			}, cellRect, true
		}
		return nil, image.Rectangle{}, false
	}

	glyphW := dr.Dx()
	glyphH := dr.Dy()

	if glyphW <= 0 || glyphH <= 0 {
		return &GlyphInfo{
			U0:      float32(cellX) / float32(f.texWidth),
			V0:      float32(cellY) / float32(f.texHeight),
			U1:      float32(cellX) / float32(f.texWidth),
			V1:      float32(cellY) / float32(f.texHeight),
			Advance: fixed26_6ToFloat32(adv),
		}, cellRect, true
	}

	alpha := glyphAlpha(mask, image.Rect(0, 0, glyphW, glyphH), maskp)
	if f.style.Bold {
		alpha = emboldenAlpha(alpha)
	}
	if f.style.Italic {
		alpha = italicAlpha(alpha)
	}
	glyphW = alpha.Rect.Dx()
	glyphH = alpha.Rect.Dy()

	destX := cellX + glyphPadding
	destY := cellY + glyphPadding
	if glyphW > f.cellSize-glyphPadding*2 {
		glyphW = f.cellSize - glyphPadding*2
	}
	if glyphH > f.cellSize-glyphPadding*2 {
		glyphH = f.cellSize - glyphPadding*2
	}

	glyphRect := image.Rect(destX, destY, destX+glyphW, destY+glyphH)
	drawAlphaMask(atlas, glyphRect, alpha, image.Point{})
	advance := fixed26_6ToFloat32(adv)
	if f.style.Bold {
		advance++
	}

	return &GlyphInfo{
		U0:       float32(destX) / float32(f.texWidth),
		V0:       float32(destY) / float32(f.texHeight),
		U1:       float32(destX+glyphW) / float32(f.texWidth),
		V1:       float32(destY+glyphH) / float32(f.texHeight),
		Advance:  advance,
		Width:    float32(glyphW),
		Height:   float32(glyphH),
		BearingX: float32(dr.Min.X),
		BearingY: float32(-dr.Min.Y),
	}, image.Rect(destX, destY, destX+glyphW, destY+glyphH), true
}

// uploadToGPU uploads the atlas to GPU
func (f *Font) uploadToGPU() error {
	if f == nil || f.atlas == nil {
		return fmt.Errorf("font atlas is not available")
	}
	if !fontTextureGLReady() {
		return fmt.Errorf("font texture upload requires initialized OpenGL texture functions")
	}
	if f.texture != 0 {
		gl.DeleteTextures(1, &f.texture)
		f.texture = 0
	}

	var tex uint32
	gl.GenTextures(1, &tex)
	if tex == 0 {
		return fmt.Errorf("font texture creation failed")
	}
	gl.BindTexture(gl.GL_TEXTURE_2D, tex)
	gl.TexParameteri(gl.GL_TEXTURE_2D, gl.GL_TEXTURE_MIN_FILTER, gl.GL_LINEAR)
	gl.TexParameteri(gl.GL_TEXTURE_2D, gl.GL_TEXTURE_MAG_FILTER, gl.GL_LINEAR)
	gl.TexParameteri(gl.GL_TEXTURE_2D, gl.GL_TEXTURE_WRAP_S, gl.GL_CLAMP_TO_EDGE)
	gl.TexParameteri(gl.GL_TEXTURE_2D, gl.GL_TEXTURE_WRAP_T, gl.GL_CLAMP_TO_EDGE)

	gl.TexImage2D(gl.GL_TEXTURE_2D, 0, int32(gl.GL_RGBA), int32(f.texWidth), int32(f.texHeight), 0,
		gl.GL_RGBA, gl.GL_UNSIGNED_BYTE, unsafePtr(f.atlas.Pix))

	f.texture = tex
	f.gpuSlot = f.nextSlot

	return nil
}

// SyncGPU ensures the GPU texture is up to date with the CPU atlas.
// Must be called with a current GL context (e.g. inside Render/BeginFrame).
func (f *Font) SyncGPU() {
	if f == nil {
		return
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.nextSlot == 0 || f.gpuSlot >= f.nextSlot {
		return
	}
	if f.texture == 0 || f.gpuSlot == 0 {
		// First upload or full re-upload
		if f.texture != 0 {
			gl.DeleteTextures(1, &f.texture)
			f.texture = 0
		}
		_ = f.uploadToGPU()
	} else {
		// Incremental upload of the entire atlas
		// (simpler than tracking per-glyph dirty rects)
		if f.texture != 0 {
			gl.DeleteTextures(1, &f.texture)
			f.texture = 0
		}
		_ = f.uploadToGPU()
	}
}

// Texture returns the font texture
func (f *Font) Texture() uint32 {
	if f == nil {
		return 0
	}
	return f.texture
}

// LineHeight returns the line height
func (f *Font) LineHeight() float32 {
	if f == nil {
		return 0
	}
	return f.lineHeight
}

// Ascent returns the ascent
func (f *Font) Ascent() float32 {
	if f == nil {
		return 0
	}
	return f.ascent
}

// Descent returns the descent (positive value indicating pixels below baseline).
func (f *Font) Descent() float32 {
	if f == nil {
		return 0
	}
	return f.descent
}

// LetterSpacing returns the extra spacing inserted after each glyph.
func (f *Font) LetterSpacing() float32 {
	if f == nil {
		return 0
	}
	return f.letterGap
}

// SetFallbacks replaces the fallback fonts used when a rune is missing.
func (f *Font) SetFallbacks(fallbacks ...*Font) {
	if f == nil {
		return
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.fallbacks = f.fallbacks[:0]
	for _, fallback := range fallbacks {
		if fallback != nil && fallback != f {
			f.fallbacks = append(f.fallbacks, fallback)
		}
	}
}

// Fallbacks returns a snapshot of configured fallback fonts.
func (f *Font) Fallbacks() []*Font {
	if f == nil {
		return nil
	}
	f.mu.RLock()
	defer f.mu.RUnlock()
	out := make([]*Font, len(f.fallbacks))
	copy(out, f.fallbacks)
	return out
}

// HasRune reports whether this font directly contains the rune.
func (f *Font) HasRune(r rune) bool {
	if f == nil || f.sfntFont == nil {
		return false
	}
	idx, err := f.sfntFont.GlyphIndex(f.sfntBuf, r)
	return err == nil && idx != 0
}

// GetGlyph returns glyph info for a rune
func (f *Font) GetGlyph(r rune) (*GlyphInfo, bool) {
	if f == nil {
		return nil, false
	}

	f.mu.RLock()
	g, ok := f.glyphs[r]
	f.mu.RUnlock()
	if !ok {
		// Try to add glyph on demand
		f.addGlyph(r)
		f.mu.RLock()
		g, ok = f.glyphs[r]
		f.mu.RUnlock()
	}

	if !ok {
		f.mu.RLock()
		g, ok = f.glyphs[' ']
		f.mu.RUnlock()
	}
	return g, ok
}

// ResolveGlyph returns the font atlas and glyph used to draw a rune, including fallbacks.
func (f *Font) ResolveGlyph(r rune) (*Font, *GlyphInfo, bool) {
	if f == nil {
		return nil, nil, false
	}
	if f.HasRune(r) || r == ' ' || r == '\t' {
		if g, ok := f.GetGlyph(r); ok {
			return f, g, true
		}
	}
	for _, fallback := range f.Fallbacks() {
		if fallback == nil || !fallback.HasRune(r) {
			continue
		}
		if g, ok := fallback.GetGlyph(r); ok {
			return fallback, g, true
		}
	}
	if g, ok := f.GetGlyph(r); ok {
		return f, g, true
	}
	return nil, nil, false
}

// RuneAdvance returns a rune's advance including configured letter spacing.
func (f *Font) RuneAdvance(r rune) float32 {
	if f == nil {
		return 0
	}
	_, g, ok := f.ResolveGlyph(r)
	if !ok {
		return 0
	}
	return g.Advance + f.letterGap
}

// addGlyph adds a glyph on demand, growing the atlas if needed.
func (f *Font) addGlyph(r rune) {
	if f == nil {
		return
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	// Check again under write lock
	if _, ok := f.glyphs[r]; ok {
		return
	}

	if f.face == nil {
		return
	}

	// Grow atlas if we've run out of slots
	if f.nextSlot >= f.maxSlots() {
		if !f.growAtlas() {
			return
		}
	}

	gInfo, _, ok := f.rasterizeGlyph(f.atlas, r, f.nextSlot)
	if !ok {
		return
	}

	f.glyphs[r] = gInfo
	f.nextSlot++

	// Note: GPU upload is deferred to SyncGPU(), called from DrawText
	// during the render pass. This way addGlyph works correctly even when
	// called during text layout (outside a GL context).
}

// growAtlas doubles the atlas height up to maxAtlasSize.
// After growth, UV coordinates of all existing glyphs are recomputed
// to account for the larger texture dimensions.
func (f *Font) growAtlas() bool {
	if f.texHeight >= maxAtlasSize {
		return false
	}
	oldHeight := f.texHeight
	newSize := f.texHeight * 2
	if newSize > maxAtlasSize {
		newSize = maxAtlasSize
	}

	newAtlas := image.NewRGBA(image.Rect(0, 0, f.texWidth, newSize))
	draw.Draw(newAtlas, f.atlas.Bounds(), f.atlas, image.Point{}, draw.Src)
	f.atlas = newAtlas
	f.texHeight = newSize

	// Recalculate cols based on new width
	f.cols = f.texWidth / f.cellSize
	if f.cols < 1 {
		f.cols = 1
	}

	// Fix UV coordinates of all existing glyphs.
	// UV is normalized (0-1). The pixel data was copied to the larger atlas
	// at the same pixel positions, so V = pixelY / texHeight must be
	// recomputed using the new texHeight.
	ratio := float32(oldHeight) / float32(f.texHeight)
	for _, g := range f.glyphs {
		g.V0 *= ratio
		g.V1 *= ratio
	}

	// Re-upload full atlas to GPU
	if f.texture != 0 {
		gl.DeleteTextures(1, &f.texture)
		f.texture = 0
	}
	_ = f.uploadToGPU()
	return true
}

func (f *Font) maxSlots() int {
	if f.cellSize <= 0 || f.cols <= 0 {
		return 0
	}
	rows := f.texHeight / f.cellSize
	slots := f.cols * rows
	if slots > maxGlyphs {
		return maxGlyphs
	}
	return slots
}

func (f *Font) uploadRect(rect image.Rectangle) {
	if f == nil || f.texture == 0 || rect.Empty() || !fontTextureUpdateGLReady() {
		return
	}

	rect = rect.Intersect(f.atlas.Bounds())
	if rect.Empty() {
		return
	}

	pixels := rgbaPatch(f.atlas, rect)
	if len(pixels) == 0 {
		return
	}

	gl.BindTexture(gl.GL_TEXTURE_2D, f.texture)
	gl.TexSubImage2D(
		gl.GL_TEXTURE_2D,
		0,
		int32(rect.Min.X),
		int32(rect.Min.Y),
		int32(rect.Dx()),
		int32(rect.Dy()),
		gl.GL_RGBA,
		gl.GL_UNSIGNED_BYTE,
		unsafePtr(pixels),
	)
}

func rgbaPatch(img *image.RGBA, rect image.Rectangle) []byte {
	if img == nil {
		return nil
	}
	rect = rect.Intersect(img.Bounds())
	if rect.Empty() {
		return nil
	}

	width := rect.Dx()
	height := rect.Dy()
	out := make([]byte, width*height*4)
	for y := 0; y < height; y++ {
		srcStart := img.PixOffset(rect.Min.X, rect.Min.Y+y)
		srcEnd := srcStart + width*4
		dstStart := y * width * 4
		copy(out[dstStart:dstStart+width*4], img.Pix[srcStart:srcEnd])
	}
	return out
}

// TextWidth calculates the width of a string
func (f *Font) TextWidth(text string) float32 {
	if f == nil {
		return 0
	}
	var w float32
	for _, r := range text {
		w += f.RuneAdvance(r)
	}
	return w
}

// MeasureText returns width and height of text
func (f *Font) MeasureText(text string) (float32, float32) {
	if f == nil {
		return 0, 0
	}
	return f.TextWidth(text), f.lineHeight
}

// Delete deletes the font texture
func (f *Font) Delete() {
	if f == nil {
		return
	}
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.texture != 0 && gl.DeleteTextures != nil {
		gl.DeleteTextures(1, &f.texture)
	}
	f.texture = 0
	if f.face != nil {
		f.face.Close()
		f.face = nil
	}
	globalFontCache.Lock()
	if cached := globalFontCache.fonts[f.cacheKey]; cached == f {
		delete(globalFontCache.fonts, f.cacheKey)
	}
	globalFontCache.Unlock()
}

// ForEachGlyph calls fn for each cached glyph while holding the read lock.
// Prefer this over Glyphs() for non-debug iteration (no map copy).
func (f *Font) ForEachGlyph(fn func(r rune, g *GlyphInfo)) {
	if f == nil || fn == nil {
		return
	}
	f.mu.RLock()
	defer f.mu.RUnlock()
	for r, g := range f.glyphs {
		fn(r, g)
	}
}

// Glyphs returns a copy of the glyph map (for debugging only).
// For production iteration use ForEachGlyph instead.
func (f *Font) Glyphs() map[rune]*GlyphInfo {
	if f == nil {
		return nil
	}
	f.mu.RLock()
	defer f.mu.RUnlock()
	out := make(map[rune]*GlyphInfo, len(f.glyphs))
	for r, g := range f.glyphs {
		out[r] = g
	}
	return out
}

func fixed26_6ToFloat32(x fixed.Int26_6) float32 {
	return float32(x) / 64.0
}

func unsafePtr(p []byte) uintptr {
	if len(p) == 0 {
		return 0
	}
	return uintptr(unsafe.Pointer(&p[0]))
}

func drawGlyphMask(dst *image.RGBA, rect image.Rectangle, mask image.Image, maskp image.Point) {
	drawAlphaMask(dst, rect, glyphAlpha(mask, rect.Sub(rect.Min), maskp), image.Point{})
}

func glyphAlpha(mask image.Image, rect image.Rectangle, maskp image.Point) *image.Alpha {
	alpha := image.NewAlpha(image.Rect(0, 0, rect.Dx(), rect.Dy()))
	if mask == nil || rect.Empty() {
		return alpha
	}
	if src, ok := mask.(*image.Alpha); ok {
		for y := 0; y < rect.Dy(); y++ {
			srcY := maskp.Y + y
			if srcY < src.Rect.Min.Y || srcY >= src.Rect.Max.Y {
				continue
			}
			for x := 0; x < rect.Dx(); x++ {
				srcX := maskp.X + x
				if srcX < src.Rect.Min.X || srcX >= src.Rect.Max.X {
					continue
				}
				alpha.Pix[alpha.PixOffset(x, y)] = src.Pix[src.PixOffset(srcX, srcY)]
			}
		}
		return alpha
	}
	for y := 0; y < rect.Dy(); y++ {
		for x := 0; x < rect.Dx(); x++ {
			_, _, _, a16 := mask.At(maskp.X+x, maskp.Y+y).RGBA()
			alpha.Pix[alpha.PixOffset(x, y)] = uint8(a16 >> 8)
		}
	}
	return alpha
}

func emboldenAlpha(src *image.Alpha) *image.Alpha {
	if src == nil || src.Rect.Empty() {
		return src
	}
	dst := image.NewAlpha(image.Rect(0, 0, src.Rect.Dx()+1, src.Rect.Dy()))
	for y := 0; y < src.Rect.Dy(); y++ {
		for x := 0; x < src.Rect.Dx(); x++ {
			a := src.Pix[src.PixOffset(src.Rect.Min.X+x, src.Rect.Min.Y+y)]
			if a == 0 {
				continue
			}
			for dx := 0; dx <= 1; dx++ {
				off := dst.PixOffset(x+dx, y)
				if a > dst.Pix[off] {
					dst.Pix[off] = a
				}
			}
		}
	}
	return dst
}

func italicAlpha(src *image.Alpha) *image.Alpha {
	if src == nil || src.Rect.Empty() {
		return src
	}
	h := src.Rect.Dy()
	maxShift := h / 4
	if maxShift < 1 {
		maxShift = 1
	}
	dst := image.NewAlpha(image.Rect(0, 0, src.Rect.Dx()+maxShift, h))
	for y := 0; y < h; y++ {
		shift := (h - 1 - y) * maxShift / h
		for x := 0; x < src.Rect.Dx(); x++ {
			a := src.Pix[src.PixOffset(src.Rect.Min.X+x, src.Rect.Min.Y+y)]
			if a == 0 {
				continue
			}
			off := dst.PixOffset(x+shift, y)
			if a > dst.Pix[off] {
				dst.Pix[off] = a
			}
		}
	}
	return dst
}

func drawAlphaMask(dst *image.RGBA, rect image.Rectangle, alpha *image.Alpha, maskp image.Point) {
	if dst == nil || alpha == nil || rect.Empty() {
		return
	}
	rect = rect.Intersect(dst.Bounds())
	if rect.Empty() {
		return
	}

	if alpha == nil {
		return
	}
	for y := 0; y < rect.Dy(); y++ {
		srcY := maskp.Y + y
		if srcY < alpha.Rect.Min.Y || srcY >= alpha.Rect.Max.Y {
			continue
		}
		for x := 0; x < rect.Dx(); x++ {
			srcX := maskp.X + x
			if srcX < alpha.Rect.Min.X || srcX >= alpha.Rect.Max.X {
				continue
			}
			a := alpha.Pix[alpha.PixOffset(srcX, srcY)]
			off := dst.PixOffset(rect.Min.X+x, rect.Min.Y+y)
			dst.Pix[off+0] = 255
			dst.Pix[off+1] = 255
			dst.Pix[off+2] = 255
			dst.Pix[off+3] = a
		}
	}
}

func fontTextureGLReady() bool {
	return gl.GenTextures != nil &&
		gl.DeleteTextures != nil &&
		gl.BindTexture != nil &&
		gl.TexParameteri != nil &&
		gl.TexImage2D != nil
}

func fontTextureUpdateGLReady() bool {
	return gl.BindTexture != nil && gl.TexSubImage2D != nil
}

// ColorToRGBA converts Color to color.RGBA
func ColorToRGBA(c math.Color) color.RGBA {
	return color.RGBA{
		R: uint8(c.R * 255),
		G: uint8(c.G * 255),
		B: uint8(c.B * 255),
		A: uint8(c.A * 255),
	}
}
