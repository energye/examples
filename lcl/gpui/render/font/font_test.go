package font

import "testing"

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
