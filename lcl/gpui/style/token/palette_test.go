package token

import (
	"testing"

	"github.com/energye/examples/lcl/gpui/core/math"
)

func TestGeneratePalette(t *testing.T) {
	// Test with Ant Design primary blue #1677ff
	primary := math.NewColorFromHex(0x1677ffff)
	palette := GeneratePalette(primary)

	// color-1 should be very light
	if palette[1].ToHSL().L < 0.9 {
		t.Fatalf("color-1 should be very light (L > 0.9), got L=%f", palette[1].ToHSL().L)
	}

	// color-10 should be dark
	if palette[10].ToHSL().L > 0.3 {
		t.Fatalf("color-10 should be dark (L < 0.3), got L=%f", palette[10].ToHSL().L)
	}

	// Lightness should decrease monotonically from color-1 to color-10
	for i := 2; i <= 10; i++ {
		prevL := palette[i-1].ToHSL().L
		currL := palette[i].ToHSL().L
		if currL > prevL {
			t.Fatalf("lightness should decrease: color-%d L=%f > color-%d L=%f", i, currL, i-1, prevL)
		}
	}

	// color-5 should preserve the hue
	seedHSL := primary.ToHSL()
	paletteHSL := palette[5].ToHSL()
	hueDiff := abs32(seedHSL.H - paletteHSL.H)
	if hueDiff > 5 {
		t.Fatalf("color-5 should preserve hue: seed H=%f, palette H=%f", seedHSL.H, paletteHSL.H)
	}
}

func TestGeneratePaletteFromSeed(t *testing.T) {
	palettes := GeneratePaletteFromSeed(
		math.NewColorFromHex(0x1677ffff), // primary
		math.NewColorFromHex(0x52c41aff), // success
		math.NewColorFromHex(0xfaad14ff), // warning
		math.NewColorFromHex(0xff4d4fff), // error
		math.NewColorFromHex(0x1677ffff), // info
	)

	if len(palettes) != 5 {
		t.Fatalf("expected 5 palettes, got %d", len(palettes))
	}

	for name, palette := range palettes {
		// Each palette should have 11 entries (0-10)
		if len(palette) != 11 {
			t.Fatalf("palette %s should have 11 entries, got %d", name, len(palette))
		}

		// Verify monotonic lightness
		for i := 2; i <= 10; i++ {
			prevL := palette[i-1].ToHSL().L
			currL := palette[i].ToHSL().L
			if currL > prevL {
				t.Fatalf("palette %s: lightness should decrease at index %d", name, i)
			}
		}
	}
}

func abs32(x float32) float32 {
	if x < 0 {
		return -x
	}
	return x
}
