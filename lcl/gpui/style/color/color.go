// Package color provides Ant Design style color system
package color

import "github.com/energye/examples/lcl/gpui/core/math"

// Primary colors
var (
	Primary     = math.NewColor(0.094, 0.565, 1.0, 1.0)  // #1890ff
	PrimaryHover = math.NewColor(0.251, 0.667, 1.0, 1.0) // #40a9ff
	PrimaryActive = math.NewColor(0.035, 0.427, 0.851, 1.0) // #096dd9
	PrimaryBg   = math.NewColor(0.937, 0.969, 1.0, 1.0)  // #e6f7ff
)

// Semantic colors
var (
	Success     = math.NewColor(0.322, 0.769, 0.102, 1.0) // #52c41a
	SuccessHover = math.NewColor(0.420, 0.824, 0.224, 1.0) // #6bc82e
	SuccessActive = math.NewColor(0.259, 0.706, 0.075, 1.0) // #42a50f

	Warning     = math.NewColor(0.980, 0.678, 0.078, 1.0) // #faad14
	WarningHover = math.NewColor(1.0, 0.749, 0.220, 1.0)  // #ffbf38
	WarningActive = math.NewColor(0.918, 0.620, 0.047, 1.0) // #ea9e0c

	Error       = math.NewColor(1.0, 0.302, 0.310, 1.0)   // #ff4d4f
	ErrorHover   = math.NewColor(1.0, 0.420, 0.427, 1.0)   // #ff6b6d
	ErrorActive  = math.NewColor(0.918, 0.224, 0.235, 1.0)  // #ea393c

	Info        = Primary
	InfoHover   = PrimaryHover
	InfoActive  = PrimaryActive
)

// Text colors
var (
	TextPrimary   = math.NewColor(0, 0, 0, 0.85)    // Main text
	TextSecondary = math.NewColor(0, 0, 0, 0.45)    // Secondary text
	TextDisabled  = math.NewColor(0, 0, 0, 0.25)    // Disabled text
	TextWhite     = math.NewColor(1, 1, 1, 1.0)     // White text
	TextInverse   = math.NewColor(1, 1, 1, 0.85)    // Inverse text
)

// Background colors
var (
	BgBase     = math.NewColor(1, 1, 1, 1.0)        // Base background
	BgLight    = math.NewColor(0.980, 0.980, 0.980, 1.0) // #fafafa
	BgDark     = math.NewColor(0, 0, 0, 0.04)       // Dark background
	BgDisabled = math.NewColor(0, 0, 0, 0.04)       // Disabled background
	BgMask     = math.NewColor(0, 0, 0, 0.45)       // Mask background
)

// Border colors
var (
	BorderBase     = math.NewColor(0.851, 0.851, 0.851, 1.0) // #d9d9d9
	BorderHover    = PrimaryHover
	BorderActive   = PrimaryActive
	BorderDisabled = math.NewColor(0, 0, 0, 0.15)
)

// Shadow colors
var (
	ShadowSM = math.NewColor(0, 0, 0, 0.05)
	ShadowMD = math.NewColor(0, 0, 0, 0.10)
	LG = math.NewColor(0, 0, 0, 0.15)
)

// Component-specific colors

// ButtonColors defines button color scheme
type ButtonColors struct {
	Default  ButtonStateColors
	Primary  ButtonStateColors
	Success  ButtonStateColors
	Warning  ButtonStateColors
	Danger   ButtonStateColors
}

// ButtonStateColors defines colors for a button state
type ButtonStateColors struct {
	Background math.Color
	Text       math.Color
	Border     math.Color
}

// DefaultButtonColors returns default button colors
func DefaultButtonColors() ButtonColors {
	return ButtonColors{
		Default: ButtonStateColors{
			Background: BgBase,
			Text:       TextPrimary,
			Border:     BorderBase,
		},
		Primary: ButtonStateColors{
			Background: Primary,
			Text:       TextWhite,
			Border:     Primary,
		},
		Success: ButtonStateColors{
			Background: Success,
			Text:       TextWhite,
			Border:     Success,
		},
		Warning: ButtonStateColors{
			Background: Warning,
			Text:       TextWhite,
			Border:     Warning,
		},
		Danger: ButtonStateColors{
			Background: Error,
			Text:       TextWhite,
			Border:     Error,
		},
	}
}

// InputColors defines input color scheme
type InputColors struct {
	Background  math.Color
	Border      math.Color
	Text        math.Color
	Placeholder math.Color
	Focus       math.Color
	Error       math.Color
}

// DefaultInputColors returns default input colors
func DefaultInputColors() InputColors {
	return InputColors{
		Background:  BgBase,
		Border:      BorderBase,
		Text:        TextPrimary,
		Placeholder: TextDisabled,
		Focus:       Primary,
		Error:       Error,
	}
}
