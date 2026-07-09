// Package font provides high-quality font rendering with texture atlas
package font

import (
	"crypto/sha256"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unsafe"

	"github.com/energye/examples/lcl/gpui/core/gl"
	"github.com/energye/examples/lcl/gpui/core/math"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	xfont "golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

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
	style      FontStyle
	cacheKey   fontCacheKey
	ttFont     *truetype.Font
	fallbacks  []*Font

	// Font face for freetype on-demand rasterization.
	face xfont.Face
}

const (
	atlasSize    = 2048
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
	files  map[string][]byte
	parsed map[[sha256.Size]byte]*truetype.Font
	fonts  map[fontCacheKey]*Font
}{
	files:  make(map[string][]byte),
	parsed: make(map[[sha256.Size]byte]*truetype.Font),
	fonts:  make(map[fontCacheKey]*Font),
}

// SystemFont contains bytes and metadata for a discovered system font.
type SystemFont struct {
	Path       string
	Data       []byte
	LatinScore int
	CJKScore   int
	Validated  bool
}

// LatinProbeRunes are used to choose a primary UI font for ASCII text.
var LatinProbeRunes = []rune{'A', 'a', '0', ' '}

// CJKProbeRunes are used to prefer fonts that cover Chinese, Japanese, and Korean text.
var CJKProbeRunes = []rune{'中', '文', '日', '本', '語', 'あ', 'ア', '한', '글'}

var defaultFontPatterns = []string{
	// User-installed fonts.
	"~/.local/share/fonts/**/*.ttf",
	"~/.local/share/fonts/**/*.ttc",
	"~/.fonts/**/*.ttf",
	"~/.fonts/**/*.ttc",
	// Linux CJK TrueType families.
	"/usr/share/fonts/truetype/wqy/*.ttc",
	"/usr/share/fonts/truetype/wqy/*.ttf",
	"/usr/share/fonts/truetype/arphic/*.ttf",
	"/usr/share/fonts/truetype/arphic/*.ttc",
	"/usr/share/fonts/truetype/droid/*.ttf",
	"/usr/share/fonts/truetype/droid/*.ttc",
	"/usr/share/fonts/truetype/noto/*CJK*.ttc",
	"/usr/share/fonts/truetype/noto/*CJK*.ttf",
	"/usr/share/fonts/truetype/noto/NotoSans*.ttf",
	"/usr/share/fonts/truetype/noto/NotoSerif*.ttf",
	// Common fallbacks.
	"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
	"/usr/share/fonts/truetype/liberation/LiberationSans-Regular.ttf",
	// macOS.
	"/System/Library/Fonts/PingFang.ttc",
	"/System/Library/Fonts/STHeiti Light.ttc",
	"/System/Library/Fonts/Hiragino Sans GB.ttc",
	"/Library/Fonts/Arial Unicode.ttf",
	// Windows.
	"C:/Windows/Fonts/msyh.ttc",
	"C:/Windows/Fonts/msyh.ttf",
	"C:/Windows/Fonts/simsun.ttc",
	"C:/Windows/Fonts/simhei.ttf",
	"C:/Windows/Fonts/msgothic.ttc",
	"C:/Windows/Fonts/meiryo.ttc",
	"C:/Windows/Fonts/malgun.ttf",
	"C:/Windows/Fonts/arial.ttf",
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
func NewFontStyled(ttfData []byte, style FontStyle) (*Font, error) {
	style = normalizeFontStyle(style)
	key := cacheKeyFor(ttfData, style)

	globalFontCache.Lock()
	if cached := globalFontCache.fonts[key]; cached != nil {
		globalFontCache.Unlock()
		return cached, nil
	}
	globalFontCache.Unlock()

	// Parse font through golang/freetype (supports both TTF and TTC).
	f, err := parseFontCached(ttfData, key.hash)
	if err != nil {
		return nil, err
	}

	face := truetype.NewFace(f, &truetype.Options{
		Size:    style.Size,
		DPI:     style.DPI,
		Hinting: style.Hinting,
	})

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
	cols := atlasSize / cellSize
	if cols < 1 {
		cols = 1
	}

	fontObj := &Font{
		texWidth:   atlasSize,
		texHeight:  atlasSize,
		atlas:      image.NewRGBA(image.Rect(0, 0, atlasSize, atlasSize)),
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
		ttFont:     f,
		face:       face,
	}

	// Pre-rasterize common characters
	fontObj.preRasterize()

	// Upload to GPU
	if err := fontObj.uploadToGPU(); err != nil {
		face.Close()
		return nil, err
	}

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

// ValidateFontData verifies that font data is supported by the freetype parser.
func ValidateFontData(ttfData []byte) error {
	key := cacheKeyFor(ttfData, normalizeFontStyle(FontStyle{}))
	_, err := parseFontCached(ttfData, key.hash)
	return err
}

// FontCoverageScore counts how many probe runes a freetype-supported font contains.
func FontCoverageScore(ttfData []byte, probes []rune) (int, error) {
	key := cacheKeyFor(ttfData, normalizeFontStyle(FontStyle{}))
	f, err := parseFontCached(ttfData, key.hash)
	if err != nil {
		return 0, err
	}
	score := 0
	for _, r := range probes {
		if f.Index(r) != 0 {
			score++
		}
	}
	return score, nil
}

// LoadSystemFontData returns the primary freetype-supported system font.
//
// GPUI_FONT_PATHS can override or prepend candidates. It uses the OS path-list
// separator (":" on Unix, ";" on Windows). Glob patterns are accepted.
func LoadSystemFontData() (SystemFont, error) {
	fonts, err := LoadSystemFontSet()
	if err != nil {
		return SystemFont{}, err
	}
	return fonts[0], nil
}

// LoadSystemFontSet returns a primary Latin font followed by CJK-capable fallbacks.
func LoadSystemFontSet() ([]SystemFont, error) {
	paths := SystemFontCandidates()
	var candidates []SystemFont
	var firstValid SystemFont

	for _, path := range paths {
		data, err := ReadFontFile(path)
		if err != nil {
			continue
		}
		cjkScore, err := FontCoverageScore(data, CJKProbeRunes)
		if err != nil {
			continue
		}
		latinScore, err := FontCoverageScore(data, LatinProbeRunes)
		if err != nil {
			continue
		}
		candidate := SystemFont{Path: path, Data: data, LatinScore: latinScore, CJKScore: cjkScore, Validated: true}
		if firstValid.Data == nil {
			firstValid = candidate
		}
		candidates = append(candidates, candidate)
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no freetype-supported system font found")
	}

	primary := firstValid
	for _, candidate := range candidates {
		if candidate.LatinScore > primary.LatinScore ||
			(candidate.LatinScore == primary.LatinScore && candidate.CJKScore > primary.CJKScore) {
			primary = candidate
		}
	}

	out := []SystemFont{primary}
	for len(out) < 9 {
		bestIndex := -1
		var best SystemFont
		for i, candidate := range candidates {
			if candidate.Path == primary.Path || containsSystemFont(out[1:], candidate.Path) {
				continue
			}
			if candidate.CJKScore <= 0 {
				continue
			}
			if bestIndex < 0 || candidate.CJKScore > best.CJKScore ||
				(candidate.CJKScore == best.CJKScore && candidate.LatinScore > best.LatinScore) {
				bestIndex = i
				best = candidate
			}
		}
		if bestIndex < 0 {
			break
		}
		out = append(out, best)
	}
	return out, nil
}

func containsSystemFont(fonts []SystemFont, path string) bool {
	for _, font := range fonts {
		if font.Path == path {
			return true
		}
	}
	return false
}

// SystemFontCandidates returns configured and built-in font candidates.
func SystemFontCandidates() []string {
	patterns := make([]string, 0, len(defaultFontPatterns)+8)
	if env := os.Getenv("GPUI_FONT_PATHS"); env != "" {
		for _, item := range filepath.SplitList(env) {
			item = strings.TrimSpace(item)
			if item != "" {
				patterns = append(patterns, item)
			}
		}
	}
	patterns = append(patterns, defaultFontPatterns...)

	seen := make(map[string]bool, len(patterns))
	paths := make([]string, 0, len(patterns))
	for _, pattern := range patterns {
		for _, path := range expandFontPattern(pattern) {
			abs, err := filepath.Abs(path)
			if err == nil {
				path = abs
			}
			if !seen[path] {
				seen[path] = true
				paths = append(paths, path)
			}
		}
	}
	return paths
}

func expandFontPattern(pattern string) []string {
	pattern = expandHome(pattern)
	if strings.Contains(pattern, "**") {
		return globRecursive(pattern)
	}
	matches, err := filepath.Glob(pattern)
	if err == nil && len(matches) > 0 {
		return matches
	}
	if _, err := os.Stat(pattern); err == nil {
		return []string{pattern}
	}
	return nil
}

func expandHome(path string) string {
	if path == "~" {
		if home, err := os.UserHomeDir(); err == nil {
			return home
		}
	}
	if strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}

func globRecursive(pattern string) []string {
	parts := strings.SplitN(pattern, "**", 2)
	root := strings.TrimSuffix(parts[0], string(os.PathSeparator))
	if root == "" {
		root = "."
	}
	suffix := ""
	if len(parts) > 1 {
		suffix = strings.TrimPrefix(parts[1], string(os.PathSeparator))
	}

	var matches []string
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d == nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		candidate := path
		if suffix != "" {
			candidate = filepath.Join(path, suffix)
		}
		if ok, _ := filepath.Match(filepath.Base(pattern), filepath.Base(path)); ok || suffix == "" {
			if suffix == "" || strings.HasSuffix(path, strings.TrimPrefix(suffix, "*")) {
				matches = append(matches, path)
			}
			_ = candidate
		}
		return nil
	})
	return matches
}

// NewFontFromFile creates or reuses a font from a TTF/TTC file path.
func NewFontFromFile(path string, fontSize float64) (*Font, error) {
	return NewFontFileStyled(path, FontStyle{Size: fontSize, DPI: 96, Hinting: xfont.HintingFull})
}

// NewFontFileStyled creates or reuses a styled font from a TTF/TTC file path.
func NewFontFileStyled(path string, style FontStyle) (*Font, error) {
	data, err := ReadFontFile(path)
	if err != nil {
		return nil, err
	}
	return NewFontStyled(data, style)
}

// ReadFontFile reads font bytes once per absolute path and reuses the data globally.
func ReadFontFile(path string) ([]byte, error) {
	if path == "" {
		return nil, fmt.Errorf("font path is empty")
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}

	globalFontCache.Lock()
	if data := globalFontCache.files[abs]; data != nil {
		globalFontCache.Unlock()
		return data, nil
	}
	globalFontCache.Unlock()

	data, err := os.ReadFile(abs)
	if err != nil {
		return nil, err
	}

	globalFontCache.Lock()
	if cached := globalFontCache.files[abs]; cached != nil {
		globalFontCache.Unlock()
		return cached, nil
	}
	globalFontCache.files[abs] = data
	globalFontCache.Unlock()
	return data, nil
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

func parseFontCached(ttfData []byte, hash [sha256.Size]byte) (*truetype.Font, error) {
	globalFontCache.Lock()
	if f := globalFontCache.parsed[hash]; f != nil {
		globalFontCache.Unlock()
		return f, nil
	}
	globalFontCache.Unlock()

	f, err := freetype.ParseFont(ttfData)
	if err != nil {
		return nil, fmt.Errorf("parse freetype font: %w", err)
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

// preRasterize pre-rasterizes common characters
func (f *Font) preRasterize() {
	maxGlyphs := f.maxSlots()

	// Characters to rasterize
	chars := make([]rune, 0, 512)

	// ASCII printable (32-126)
	for c := rune(32); c <= 126; c++ {
		chars = append(chars, c)
	}

	// Common CJK characters (subset for fast loading)
	cjkCommon := []rune{
		'你', '好', '世', '界', '测', '试', '输', '入', '文', '字',
		'请', '在', '这', '里', '点', '击', '按', '钮', '标', '签',
		'框', '窗', '口', '程', '序', '开', '发', '使', '用',
		'中', '英', '大', '小', '多', '少', '上', '下', '左', '右',
		'是', '的', '不', '了', '在', '人', '有', '我', '他', '这',
		'个', '们', '中', '来', '到', '时', '大', '地', '为', '子',
		'说', '生', '国', '年', '着', '就', '那', '和', '要', '她',
		'出', '也', '得', '里', '后', '自', '会', '家', '可', '下',
		'而', '过', '去', '天', '能', '对', '小', '多', '然', '于',
		'心', '学', '么', '之', '都', '好', '看', '起', '发', '当',
	}
	chars = append(chars, cjkCommon...)

	for _, r := range chars {
		if f.nextSlot >= maxGlyphs {
			break
		}
		if _, exists := f.glyphs[r]; exists {
			continue
		}

		gInfo, _, ok := f.rasterizeGlyph(f.atlas, r, f.nextSlot)
		if !ok {
			continue
		}

		f.glyphs[r] = gInfo
		f.nextSlot++
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
				U0:      float32(cellX) / float32(atlasSize),
				V0:      float32(cellY) / float32(atlasSize),
				U1:      float32(cellX) / float32(atlasSize),
				V1:      float32(cellY) / float32(atlasSize),
				Advance: fixed26_6ToFloat32(adv),
			}, cellRect, true
		}
		return nil, image.Rectangle{}, false
	}

	glyphW := dr.Dx()
	glyphH := dr.Dy()

	if glyphW <= 0 || glyphH <= 0 {
		return &GlyphInfo{
			U0:      float32(cellX) / float32(atlasSize),
			V0:      float32(cellY) / float32(atlasSize),
			U1:      float32(cellX) / float32(atlasSize),
			V1:      float32(cellY) / float32(atlasSize),
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
		U0:       float32(destX) / float32(atlasSize),
		V0:       float32(destY) / float32(atlasSize),
		U1:       float32(destX+glyphW) / float32(atlasSize),
		V1:       float32(destY+glyphH) / float32(atlasSize),
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

	gl.TexImage2D(gl.GL_TEXTURE_2D, 0, int32(gl.GL_RGBA), int32(atlasSize), int32(atlasSize), 0,
		gl.GL_RGBA, gl.GL_UNSIGNED_BYTE, unsafePtr(f.atlas.Pix))

	f.texture = tex

	return nil
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
	if f == nil || f.ttFont == nil {
		return false
	}
	return f.ttFont.Index(r) != 0
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

// addGlyph adds a glyph on demand
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

	if f.face == nil || f.nextSlot >= f.maxSlots() {
		return
	}

	gInfo, dirtyRect, ok := f.rasterizeGlyph(f.atlas, r, f.nextSlot)
	if !ok {
		return
	}

	f.glyphs[r] = gInfo
	f.nextSlot++

	if f.texture == 0 {
		_ = f.uploadToGPU()
		return
	}
	f.uploadRect(dirtyRect)
}

func (f *Font) maxSlots() int {
	if f.cellSize <= 0 || f.cols <= 0 {
		return 0
	}
	rows := atlasSize / f.cellSize
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
