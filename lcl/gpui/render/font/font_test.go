package font

import (
	"os"
	"testing"
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
		U0:      0.0,
		V0:      0.0,
		U1:      0.5,
		V1:      0.5,
		Advance: 10,
		Width:   8,
		Height:  12,
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
