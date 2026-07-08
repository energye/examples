// Package token provides Ant Design color palette generation
package token

import (
	"github.com/energye/examples/lcl/gpui/core/math"
)

// Palette holds 10 shades of a color (index 1-10)
// color-1 is lightest, color-10 is darkest
type Palette [11]math.Color

// GeneratePalette creates a 10-shade color palette from a seed color
// following Ant Design's HSL-based algorithm
func GeneratePalette(seed math.Color) Palette {
	hsl := seed.ToHSL()
	var palette Palette

	// Ant Design palette generation algorithm:
	// color-1: very light (background tints)
	// color-2 to color-4: light shades
	// color-5: the seed color itself
	// color-6 to color-9: progressively darker
	// color-10: darkest

	// Lightness steps for each index (Ant Design v5 approximate)
	lightnessSteps := [11]float32{
		0,    // placeholder
		0.97, // color-1: near white
		0.93, // color-2
		0.88, // color-3
		0.80, // color-4
		0.50, // color-5: base (will be adjusted)
		0.45, // color-6
		0.40, // color-7
		0.33, // color-8
		0.26, // color-9
		0.20, // color-10: darkest
	}

	// Saturation adjustments (slightly desaturate very light/dark shades)
	saturationSteps := [11]float32{
		0,    // placeholder
		0.65, // color-1
		0.75, // color-2
		0.85, // color-3
		0.95, // color-4
		1.00, // color-5: full saturation
		1.00, // color-6
		0.95, // color-7
		0.85, // color-8
		0.75, // color-9
		0.65, // color-10
	}

	for i := 1; i <= 10; i++ {
		l := lightnessSteps[i]
		s := hsl.S * saturationSteps[i]
		if s > 1 {
			s = 1
		}
		palette[i] = math.NewColorFromHSL(hsl.H, s, l, seed.A)
	}

	return palette
}

// GeneratePaletteFromSeed creates palettes for all semantic colors
func GeneratePaletteFromSeed(seed math.Color, success, warning, error, info math.Color) map[string]Palette {
	return map[string]Palette{
		"primary": GeneratePalette(seed),
		"success": GeneratePalette(success),
		"warning": GeneratePalette(warning),
		"error":   GeneratePalette(error),
		"info":    GeneratePalette(info),
	}
}
