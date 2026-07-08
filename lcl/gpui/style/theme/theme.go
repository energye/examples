// Package theme provides the theme system
package theme

import (
	"sync"
	"time"

	"github.com/energye/examples/lcl/gpui/core/math"
	"github.com/energye/examples/lcl/gpui/style/color"
)

// Theme represents a complete UI theme
type Theme struct {
	// Typography
	FontFamily string
	FontSizeSM  float64
	FontSizeMD  float64
	FontSizeLG  float64
	FontSizeXL  float64
	LineHeight  float32

	// Spacing
	SpaceXXS float32
	SpaceXS  float32
	SpaceSM  float32
	SpaceMD  float32
	SpaceLG  float32
	SpaceXL  float32
	SpaceXXL float32

	// Border radius
	RadiusSM   float32
	RadiusMD   float32
	RadiusLG   float32
	RadiusXL   float32
	RadiusFull float32

	// Shadows
	ShadowSM Shadow
	ShadowMD Shadow
	ShadowLG Shadow

	// Animation
	DurationFast   time.Duration
	DurationNormal time.Duration
	DurationSlow   time.Duration

	// Component styles
	Button ButtonTheme
	Input  InputTheme
	Label  LabelTheme
}

// Shadow represents a box shadow
type Shadow struct {
	OffsetX float32
	OffsetY float32
	Blur    float32
	Spread  float32
	Color   math.Color
}

// ButtonTheme defines button theme
type ButtonTheme struct {
	HeightSM  float32
	HeightMD  float32
	HeightLG  float32
	HeightXL  float32

	PaddingH  float32
	PaddingV  float32

	FontSize  float64
	Radius    float32
	BorderW   float32

	color.ButtonColors
}

// InputTheme defines input theme
type InputTheme struct {
	Height    float32
	PaddingH  float32
	PaddingV  float32
	FontSize  float64
	Radius    float32
	BorderW   float32

	color.InputColors
}

// LabelTheme defines label theme
type LabelTheme struct {
	FontSize  float64
	LineHeight float32
}

// AntDesignTheme returns the Ant Design theme
func AntDesignTheme() *Theme {
	return &Theme{
		// Typography
		FontFamily: "system-ui, -apple-system, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, 'Noto Sans', sans-serif",
		FontSizeSM:  12,
		FontSizeMD:  14,
		FontSizeLG:  16,
		FontSizeXL:  20,
		LineHeight:  1.5,

		// Spacing (4px base unit)
		SpaceXXS: 4,
		SpaceXS:  8,
		SpaceSM:  12,
		SpaceMD:  16,
		SpaceLG:  24,
		SpaceXL:  32,
		SpaceXXL: 48,

		// Border radius
		RadiusSM:   2,
		RadiusMD:   4,
		RadiusLG:   6,
		RadiusXL:   8,
		RadiusFull: 9999,

		// Shadows
		ShadowSM: Shadow{0, 1, 2, 0, color.ShadowSM},
		ShadowMD: Shadow{0, 2, 4, 0, color.ShadowMD},
		ShadowLG: Shadow{0, 4, 8, 0, color.LG},

		// Animation
		DurationFast:   150 * time.Millisecond,
		DurationNormal: 200 * time.Millisecond,
		DurationSlow:   300 * time.Millisecond,

		// Button
		Button: ButtonTheme{
			HeightSM: 24,
			HeightMD: 32,
			HeightLG: 40,
			HeightXL: 48,

			PaddingH: 16,
			PaddingV: 4,

			FontSize: 14,
			Radius:   4,
			BorderW:  1,

			ButtonColors: color.DefaultButtonColors(),
		},

		// Input
		Input: InputTheme{
			Height:   32,
			PaddingH: 12,
			PaddingV: 4,

			FontSize: 14,
			Radius:   4,
			BorderW:  1,

			InputColors: color.DefaultInputColors(),
		},

		// Label
		Label: LabelTheme{
			FontSize:   14,
			LineHeight: 1.5,
		},
	}
}

// CurrentTheme is the current active theme
var (
	currentThemeMu sync.RWMutex
	currentTheme   = AntDesignTheme()
)

// GetTheme returns the current theme (safe for concurrent use).
func GetTheme() *Theme {
	currentThemeMu.RLock()
	defer currentThemeMu.RUnlock()
	return currentTheme
}

// SetTheme sets the current theme (safe for concurrent use).
func SetTheme(t *Theme) {
	currentThemeMu.Lock()
	defer currentThemeMu.Unlock()
	currentTheme = t
}
