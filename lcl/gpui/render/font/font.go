// Package font provides high-quality font rendering with texture atlas
package font

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"sync"
	"unsafe"

	"github.com/energye/examples/lcl/gpui/core/gl"
	"github.com/energye/examples/lcl/gpui/core/math"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
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
	glyphs     map[rune]*GlyphInfo
	fontSize   float64
	lineHeight float32
	ascent     float32
	descent    float32

	// Font face for on-demand rasterization
	face      font.Face
	dirty     bool
	dirtyRect image.Rectangle
}

const (
	atlasSize    = 2048
	glyphPadding = 2
	maxGlyphs    = 4096
)

// NewFont creates a new font from TTF data
func NewFont(ttfData []byte, fontSize float64) (*Font, error) {
	// Parse font (supports both TTF and TTC)
	f, err := parseFont(ttfData)
	if err != nil {
		return nil, err
	}

	// Create face
	face, err := opentype.NewFace(f, &opentype.FaceOptions{
		Size:    fontSize,
		DPI:     96,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return nil, fmt.Errorf("new face: %w", err)
	}

	// Calculate metrics
	m := face.Metrics()
	ascent := fixed26_6ToFloat32(m.Ascent)
	descent := fixed26_6ToFloat32(m.Descent)

	fontObj := &Font{
		texWidth:   atlasSize,
		texHeight:  atlasSize,
		glyphs:     make(map[rune]*GlyphInfo),
		fontSize:   fontSize,
		lineHeight: ascent + descent,
		ascent:     ascent,
		descent:    descent,
		face:       face,
		dirty:      true,
	}

	// Pre-rasterize common characters
	fontObj.preRasterize()

	// Upload to GPU
	if err := fontObj.uploadToGPU(); err != nil {
		face.Close()
		return nil, err
	}

	return fontObj, nil
}

// parseFont parses font data (supports TTF and TTC)
func parseFont(ttfData []byte) (*opentype.Font, error) {
	// Try single font first
	f, err := opentype.Parse(ttfData)
	if err == nil {
		return f, nil
	}

	// Try collection (TTC)
	collection, err2 := opentype.ParseCollection(ttfData)
	if err2 != nil {
		return nil, fmt.Errorf("parse font: %w (collection: %v)", err, err2)
	}

	if collection.NumFonts() == 0 {
		return nil, fmt.Errorf("font collection is empty")
	}

	f, err = collection.Font(0)
	if err != nil {
		return nil, fmt.Errorf("get font from collection: %w", err)
	}

	return f, nil
}

// preRasterize pre-rasterizes common characters
func (f *Font) preRasterize() {
	cellSize := int(f.fontSize*1.5) + 4
	cols := atlasSize / cellSize
	rows := atlasSize / cellSize
	maxGlyphs := cols * rows

	// Create atlas image
	atlasImg := image.NewRGBA(image.Rect(0, 0, atlasSize, atlasSize))
	draw.Draw(atlasImg, atlasImg.Bounds(), image.Transparent, image.Point{}, draw.Src)

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

	idx := 0
	for _, r := range chars {
		if idx >= maxGlyphs {
			break
		}

		gInfo, ok := f.rasterizeGlyph(atlasImg, r, idx, cols, cellSize)
		if !ok {
			continue
		}

		f.glyphs[r] = gInfo
		idx++
	}

	// Store atlas image for upload
	f.dirty = true
}

// rasterizeGlyph rasterizes a single glyph
func (f *Font) rasterizeGlyph(atlas *image.RGBA, r rune, idx, cols, cellSize int) (*GlyphInfo, bool) {
	col := idx % cols
	row := idx / cols

	cellX := col * cellSize
	cellY := row * cellSize

	// Get glyph advance
	adv, ok := f.face.GlyphAdvance(r)
	if !ok {
		if r != ' ' {
			adv, ok = f.face.GlyphAdvance(' ')
			if !ok {
				return nil, false
			}
		} else {
			return nil, false
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
			}, true
		}
		return nil, false
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
		}, true
	}

	// Clamp to cell size
	if glyphW > cellSize-glyphPadding*2 {
		glyphW = cellSize - glyphPadding*2
	}
	if glyphH > cellSize-glyphPadding*2 {
		glyphH = cellSize - glyphPadding*2
	}

	destX := cellX + glyphPadding
	destY := cellY + glyphPadding

	// Draw glyph
	draw.Draw(atlas, image.Rect(destX, destY, destX+glyphW, destY+glyphH),
		mask, maskp, draw.Over)

	return &GlyphInfo{
		U0:      float32(destX) / float32(atlasSize),
		V0:      float32(destY) / float32(atlasSize),
		U1:      float32(destX+glyphW) / float32(atlasSize),
		V1:      float32(destY+glyphH) / float32(atlasSize),
		Advance: fixed26_6ToFloat32(adv),
		Width:   float32(glyphW),
		Height:  float32(glyphH),
	}, true
}

// uploadToGPU uploads the atlas to GPU
func (f *Font) uploadToGPU() error {
	// Create atlas image
	atlasImg := image.NewRGBA(image.Rect(0, 0, atlasSize, atlasSize))
	draw.Draw(atlasImg, atlasImg.Bounds(), image.Transparent, image.Point{}, draw.Src)

	// Re-rasterize all glyphs
	cellSize := int(f.fontSize*1.5) + 4
	cols := atlasSize / cellSize

	idx := 0
	for r := range f.glyphs {
		f.rasterizeGlyph(atlasImg, r, idx, cols, cellSize)
		idx++
	}

	// Upload to GPU
	var tex uint32
	gl.GenTextures(1, &tex)
	gl.BindTexture(gl.GL_TEXTURE_2D, tex)
	gl.TexParameteri(gl.GL_TEXTURE_2D, gl.GL_TEXTURE_MIN_FILTER, gl.GL_LINEAR)
	gl.TexParameteri(gl.GL_TEXTURE_2D, gl.GL_TEXTURE_MAG_FILTER, gl.GL_LINEAR)
	gl.TexParameteri(gl.GL_TEXTURE_2D, gl.GL_TEXTURE_WRAP_S, gl.GL_CLAMP_TO_EDGE)
	gl.TexParameteri(gl.GL_TEXTURE_2D, gl.GL_TEXTURE_WRAP_T, gl.GL_CLAMP_TO_EDGE)

	gl.TexImage2D(gl.GL_TEXTURE_2D, 0, int32(gl.GL_RGBA), int32(atlasSize), int32(atlasSize), 0,
		gl.GL_RGBA, gl.GL_UNSIGNED_BYTE, unsafePtr(atlasImg.Pix))

	f.texture = tex
	f.dirty = false

	return nil
}

// Texture returns the font texture
func (f *Font) Texture() uint32 {
	return f.texture
}

// LineHeight returns the line height
func (f *Font) LineHeight() float32 {
	return f.lineHeight
}

// Ascent returns the ascent
func (f *Font) Ascent() float32 {
	return f.ascent
}

// GetGlyph returns glyph info for a rune
func (f *Font) GetGlyph(r rune) (*GlyphInfo, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	g, ok := f.glyphs[r]
	if !ok {
		// Try to add glyph on demand
		f.mu.RUnlock()
		f.addGlyph(r)
		f.mu.RLock()
		g, ok = f.glyphs[r]
	}

	if !ok {
		g, ok = f.glyphs[' ']
	}
	return g, ok
}

// addGlyph adds a glyph on demand
func (f *Font) addGlyph(r rune) {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Check again under write lock
	if _, ok := f.glyphs[r]; ok {
		return
	}

	// Find empty slot
	idx := len(f.glyphs)
	cellSize := int(f.fontSize*1.5) + 4
	cols := atlasSize / cellSize

	// Create temporary image for this glyph
	glyphImg := image.NewRGBA(image.Rect(0, 0, atlasSize, atlasSize))
	draw.Draw(glyphImg, glyphImg.Bounds(), image.Transparent, image.Point{}, draw.Src)

	gInfo, ok := f.rasterizeGlyph(glyphImg, r, idx, cols, cellSize)
	if !ok {
		return
	}

	f.glyphs[r] = gInfo

	// Upload updated atlas
	// Note: In production, you'd want to use glTexSubImage2D for efficiency
	f.uploadToGPU()
}

// TextWidth calculates the width of a string
func (f *Font) TextWidth(text string) float32 {
	var w float32
	for _, r := range text {
		if g, ok := f.GetGlyph(r); ok {
			w += g.Advance
		}
	}
	return w
}

// MeasureText returns width and height of text
func (f *Font) MeasureText(text string) (float32, float32) {
	return f.TextWidth(text), f.lineHeight
}

// Delete deletes the font texture
func (f *Font) Delete() {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.texture != 0 {
		gl.DeleteTextures(1, &f.texture)
		f.texture = 0
	}
	if f.face != nil {
		f.face.Close()
		f.face = nil
	}
}

// Glyphs returns the glyph map (for debugging)
func (f *Font) Glyphs() map[rune]*GlyphInfo {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.glyphs
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

// ColorToRGBA converts Color to color.RGBA
func ColorToRGBA(c math.Color) color.RGBA {
	return color.RGBA{
		R: uint8(c.R * 255),
		G: uint8(c.G * 255),
		B: uint8(c.B * 255),
		A: uint8(c.A * 255),
	}
}
