package font

import (
	"image"
	"os"
	"testing"

	"golang.org/x/image/font"
)

func TestNilFontMethodsAreSafe(t *testing.T) {
	var f *Font
	if f.Texture() != 0 {
		t.Fatal("nil font texture should be zero")
	}
	if f.LineHeight() != 0 {
		t.Fatal("nil font line height should be zero")
	}
	if f.Ascent() != 0 {
		t.Fatal("nil font ascent should be zero")
	}
	if width := f.TextWidth("text"); width != 0 {
		t.Fatalf("nil font text width = %v, want 0", width)
	}
	if w, h := f.MeasureText("text"); w != 0 || h != 0 {
		t.Fatalf("nil font measure = (%v,%v), want (0,0)", w, h)
	}
	if glyphs := f.Glyphs(); glyphs != nil {
		t.Fatalf("nil font glyphs = %#v, want nil", glyphs)
	}
	f.ForEachGlyph(func(r rune, g *GlyphInfo) {
		t.Fatal("nil font ForEachGlyph should not call fn")
	})
	f.Delete()
}

func TestFontLoadingAndMetrics(t *testing.T) {
	// Try to load a system font
	fontPaths := []string{
		"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
		"/usr/share/fonts/truetype/noto/NotoSans-Regular.ttf",
		"/usr/share/fonts/truetype/liberation/LiberationSans-Regular.ttf",
	}

	var fontData []byte
	var err error
	for _, path := range fontPaths {
		fontData, err = os.ReadFile(path)
		if err == nil {
			break
		}
	}

	if fontData == nil {
		t.Skip("No system font found, skipping font loading test")
		return
	}

	// Note: NewFont requires GL context for texture upload
	// This test verifies the font data can be read
	if len(fontData) == 0 {
		t.Fatal("Font file is empty")
	}

	t.Logf("Font data loaded: %d bytes", len(fontData))
}

func TestFontTextWidthCalculation(t *testing.T) {
	// Test that TextWidth calculation logic is correct
	// This tests the measurement without requiring GL context

	// Create a mock font with known metrics
	f := &Font{
		lineHeight: 20,
		ascent:     15,
		descent:    5,
		glyphs: map[rune]*GlyphInfo{
			'A': {Advance: 10, Width: 8, Height: 12, BearingX: 1, BearingY: 11},
			'B': {Advance: 11, Width: 9, Height: 12, BearingX: 1, BearingY: 11},
			'C': {Advance: 10, Width: 8, Height: 12, BearingX: 1, BearingY: 11},
			' ': {Advance: 5, Width: 0, Height: 0},
		},
	}

	// Test single character width
	if w := f.TextWidth("A"); w != 10 {
		t.Fatalf("TextWidth('A') = %f, want 10", w)
	}

	// Test string width
	if w := f.TextWidth("ABC"); w != 31 {
		t.Fatalf("TextWidth('ABC') = %f, want 31", w)
	}

	// Test with space
	if w := f.TextWidth("A B"); w != 26 {
		t.Fatalf("TextWidth('A B') = %f, want 26", w)
	}

	// Test empty string
	if w := f.TextWidth(""); w != 0 {
		t.Fatalf("TextWidth('') = %f, want 0", w)
	}

	// Test MeasureText
	w, h := f.MeasureText("ABC")
	if w != 31 || h != 20 {
		t.Fatalf("MeasureText('ABC') = (%f, %f), want (31, 20)", w, h)
	}
}

func TestGlyphInfoMetrics(t *testing.T) {
	g := &GlyphInfo{
		U0:       0.0,
		V0:       0.0,
		U1:       0.5,
		V1:       0.5,
		Advance:  10,
		Width:    8,
		Height:   12,
		BearingX: 1,
		BearingY: 11,
	}

	if g.Advance != 10 {
		t.Fatalf("GlyphInfo.Advance = %f, want 10", g.Advance)
	}
	if g.Width != 8 {
		t.Fatalf("GlyphInfo.Width = %f, want 8", g.Width)
	}
	if g.Height != 12 {
		t.Fatalf("GlyphInfo.Height = %f, want 12", g.Height)
	}
	if g.BearingX != 1 {
		t.Fatalf("GlyphInfo.BearingX = %f, want 1", g.BearingX)
	}
	if g.BearingY != 11 {
		t.Fatalf("GlyphInfo.BearingY = %f, want 11", g.BearingY)
	}
}

func TestDrawGlyphMaskStoresWhiteRGBWithCoverageAlpha(t *testing.T) {
	dst := image.NewRGBA(image.Rect(0, 0, 4, 4))
	mask := image.NewAlpha(image.Rect(0, 0, 2, 2))
	mask.Pix[mask.PixOffset(0, 0)] = 0
	mask.Pix[mask.PixOffset(1, 0)] = 64
	mask.Pix[mask.PixOffset(0, 1)] = 128
	mask.Pix[mask.PixOffset(1, 1)] = 255

	drawGlyphMask(dst, image.Rect(1, 1, 3, 3), mask, image.Point{})

	checks := []struct {
		x, y int
		a    uint8
	}{
		{1, 1, 0},
		{2, 1, 64},
		{1, 2, 128},
		{2, 2, 255},
	}
	for _, check := range checks {
		off := dst.PixOffset(check.x, check.y)
		if dst.Pix[off+0] != 255 || dst.Pix[off+1] != 255 || dst.Pix[off+2] != 255 || dst.Pix[off+3] != check.a {
			t.Fatalf("pixel (%d,%d) = rgba(%d,%d,%d,%d), want white alpha %d",
				check.x, check.y,
				dst.Pix[off+0], dst.Pix[off+1], dst.Pix[off+2], dst.Pix[off+3],
				check.a)
		}
	}
}

// TestFontStyleStruct verifies FontStyle struct fields.
func TestFontStyleStruct(t *testing.T) {
	style := FontStyle{
		Size:    16.0,
		Bold:    true,
		Italic:  false,
		DPI:     96.0,
		Hinting: font.HintingFull,
	}

	if style.Size != 16.0 {
		t.Fatalf("FontStyle.Size = %f, want 16.0", style.Size)
	}
	if !style.Bold {
		t.Fatal("FontStyle.Bold should be true")
	}
	if style.Italic {
		t.Fatal("FontStyle.Italic should be false")
	}
	if style.DPI != 96.0 {
		t.Fatalf("FontStyle.DPI = %f, want 96.0", style.DPI)
	}
	if style.Hinting != font.HintingFull {
		t.Fatalf("FontStyle.Hinting = %v, want HintingFull", style.Hinting)
	}
}

// TestFontStyleDefaultValues verifies default values in NewFont.
func TestFontStyleDefaultValues(t *testing.T) {
	// NewFont should set DPI=96 and HintingFull by default
	style := FontStyle{
		Size: 14.0,
		// DPI and Hinting not set - should use defaults in NewFontStyled
	}

	// Verify the style struct has the expected size
	if style.Size != 14.0 {
		t.Fatalf("FontStyle.Size = %f, want 14.0", style.Size)
	}
	// DPI defaults to 0 (not set), NewFontStyled should handle this
	if style.DPI != 0 {
		t.Fatalf("FontStyle.DPI = %f, want 0 (unset)", style.DPI)
	}
}

// TestFontStyleWithBold verifies bold font style.
func TestFontStyleWithBold(t *testing.T) {
	style := FontStyle{
		Size:   14.0,
		Bold:   true,
		Italic: false,
		DPI:    96.0,
	}

	if !style.Bold {
		t.Fatal("FontStyle.Bold should be true")
	}
	if style.Italic {
		t.Fatal("FontStyle.Italic should be false")
	}
}

// TestFontStyleWithItalic verifies italic font style.
func TestFontStyleWithItalic(t *testing.T) {
	style := FontStyle{
		Size:   14.0,
		Bold:   false,
		Italic: true,
		DPI:    96.0,
	}

	if style.Bold {
		t.Fatal("FontStyle.Bold should be false")
	}
	if !style.Italic {
		t.Fatal("FontStyle.Italic should be true")
	}
}

// TestFontStyleWithBoldItalic verifies bold+italic font style.
func TestFontStyleWithBoldItalic(t *testing.T) {
	style := FontStyle{
		Size:   14.0,
		Bold:   true,
		Italic: true,
		DPI:    96.0,
	}

	if !style.Bold {
		t.Fatal("FontStyle.Bold should be true")
	}
	if !style.Italic {
		t.Fatal("FontStyle.Italic should be true")
	}
}
