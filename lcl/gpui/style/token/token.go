package token

import (
	"time"

	"github.com/energye/examples/lcl/gpui/core/math"
)

// Mode identifies the token color mode.
type Mode int

const (
	ModeLight Mode = iota
	ModeDark
)

// SeedToken is the minimal user-controlled theme input.
type SeedToken struct {
	ColorPrimary math.Color
	ColorSuccess math.Color
	ColorWarning math.Color
	ColorError   math.Color
	ColorInfo    math.Color

	FontFamily string
	FontSize   float32
	LineHeight float32

	SizeUnit     float32
	BorderRadius float32
}

// GlobalToken contains derived design primitives.
type GlobalToken struct {
	ColorPrimary       math.Color
	ColorPrimaryHover  math.Color
	ColorPrimaryActive math.Color
	ColorSuccess       math.Color
	ColorWarning       math.Color
	ColorError         math.Color
	ColorInfo          math.Color

	// 10-shade palettes for each semantic color
	// Index 1-10: color-1 (lightest) to color-10 (darkest)
	ColorPrimaryPalette   Palette
	ColorSuccessPalette   Palette
	ColorWarningPalette   Palette
	ColorErrorPalette     Palette
	ColorInfoPalette      Palette

	ColorText          math.Color
	ColorTextSecondary math.Color
	ColorTextDisabled  math.Color
	ColorTextLight     math.Color
	ColorBgBase        math.Color
	ColorBgContainer   math.Color
	ColorBgElevated    math.Color
	ColorBgMask        math.Color
	ColorBorder        math.Color
	ColorBorderHover   math.Color

	FontFamily string
	FontSize   float32
	FontSizeSM float32
	FontSizeLG float32
	LineHeight float32

	RadiusSM float32
	RadiusMD float32
	RadiusLG float32
	RadiusXL float32

	SpaceXXS float32
	SpaceXS  float32
	SpaceSM  float32
	SpaceMD  float32
	SpaceLG  float32
	SpaceXL  float32

	ShadowSM ShadowToken
	ShadowMD ShadowToken
	ShadowLG ShadowToken

	MotionDurationFast time.Duration
	MotionDurationMid  time.Duration
	MotionDurationSlow time.Duration
	MotionEaseOut      string
	MotionEaseInOut    string
}

// AliasToken contains semantic aliases used by components.
type AliasToken struct {
	ColorLink        math.Color
	ColorLinkHover   math.Color
	ColorLinkActive  math.Color
	ColorFillAlter   math.Color
	ColorFillContent math.Color
	ColorSplit       math.Color
	ControlHeightSM  float32
	ControlHeight    float32
	ControlHeightLG  float32
}

// ComponentTokens stores per-component defaults.
type ComponentTokens struct {
	Button ButtonToken
	Input  InputToken
	Card   CardToken
	Modal  ModalToken
}

type ButtonToken struct {
	HeightSM    float32
	Height      float32
	HeightLG    float32
	PaddingH    float32
	BorderWidth float32
	Radius      float32
}

type InputToken struct {
	Height      float32
	PaddingH    float32
	PaddingV    float32
	BorderWidth float32
	Radius      float32
}

type CardToken struct {
	Padding     float32
	Radius      float32
	Shadow      ShadowToken
	BorderWidth float32
}

type ModalToken struct {
	Width     float32
	HeaderH   float32
	Padding   float32
	Radius    float32
	MaskColor math.Color
	Shadow    ShadowToken
}

type ShadowToken struct {
	Offset math.Vec2
	Blur   float32
	Spread float32
	Color  math.Color
}

// Tokens is a complete derived token set.
type Tokens struct {
	Mode       Mode
	Seed       SeedToken
	Global     GlobalToken
	Alias      AliasToken
	Components ComponentTokens
}

// DefaultSeed returns Ant Design-like default seed tokens.
func DefaultSeed() SeedToken {
	return SeedToken{
		ColorPrimary: math.NewColor(0.086, 0.467, 1.0, 1),   // #1677ff
		ColorSuccess: math.NewColor(0.322, 0.769, 0.102, 1), // #52c41a
		ColorWarning: math.NewColor(0.980, 0.678, 0.078, 1), // #faad14
		ColorError:   math.NewColor(1.0, 0.302, 0.310, 1),   // #ff4d4f
		ColorInfo:    math.NewColor(0.086, 0.467, 1.0, 1),
		FontFamily:   "system-ui, -apple-system, Segoe UI, Roboto, Helvetica Neue, Arial, Noto Sans, sans-serif",
		FontSize:     14,
		LineHeight:   1.5715,
		SizeUnit:     4,
		BorderRadius: 6,
	}
}

// Derive creates a full token set from seed tokens.
func Derive(seed SeedToken, mode Mode) Tokens {
	global := deriveGlobal(seed, mode)
	alias := deriveAlias(global)
	return Tokens{
		Mode:       mode,
		Seed:       seed,
		Global:     global,
		Alias:      alias,
		Components: deriveComponents(global, alias),
	}
}

func deriveGlobal(seed SeedToken, mode Mode) GlobalToken {
	text := math.NewColor(0, 0, 0, 0.88)
	textSecondary := math.NewColor(0, 0, 0, 0.65)
	textDisabled := math.NewColor(0, 0, 0, 0.25)
	bgBase := math.NewColor(1, 1, 1, 1)
	bgContainer := bgBase
	bgElevated := bgBase
	border := math.NewColor(0.851, 0.851, 0.851, 1)
	fillAlter := math.NewColor(0, 0, 0, 0.02)
	_ = fillAlter

	if mode == ModeDark {
		text = math.NewColor(1, 1, 1, 0.85)
		textSecondary = math.NewColor(1, 1, 1, 0.65)
		textDisabled = math.NewColor(1, 1, 1, 0.30)
		bgBase = math.NewColor(0.086, 0.086, 0.086, 1)
		bgContainer = math.NewColor(0.122, 0.122, 0.122, 1)
		bgElevated = math.NewColor(0.165, 0.165, 0.165, 1)
		border = math.NewColor(1, 1, 1, 0.18)
	}

	return GlobalToken{
		ColorPrimary:       seed.ColorPrimary,
		ColorPrimaryHover:  seed.ColorPrimary.LightenHSL(0.08),
		ColorPrimaryActive: seed.ColorPrimary.DarkenHSL(0.10),
		ColorSuccess:       seed.ColorSuccess,
		ColorWarning:       seed.ColorWarning,
		ColorError:         seed.ColorError,
		ColorInfo:          seed.ColorInfo,

		// Generate 10-shade palettes for each semantic color
		ColorPrimaryPalette: GeneratePalette(seed.ColorPrimary),
		ColorSuccessPalette: GeneratePalette(seed.ColorSuccess),
		ColorWarningPalette: GeneratePalette(seed.ColorWarning),
		ColorErrorPalette:   GeneratePalette(seed.ColorError),
		ColorInfoPalette:    GeneratePalette(seed.ColorInfo),

		ColorText:          text,
		ColorTextSecondary: textSecondary,
		ColorTextDisabled:  textDisabled,
		ColorTextLight:     math.NewColor(1, 1, 1, 1),
		ColorBgBase:        bgBase,
		ColorBgContainer:   bgContainer,
		ColorBgElevated:    bgElevated,
		ColorBgMask:        math.NewColor(0, 0, 0, 0.45),
		ColorBorder:        border,
		ColorBorderHover:   seed.ColorPrimary.LightenHSL(0.10),
		FontFamily:         seed.FontFamily,
		FontSize:           seed.FontSize,
		FontSizeSM:         seed.FontSize - 2,
		FontSizeLG:         seed.FontSize + 2,
		LineHeight:         seed.LineHeight,
		RadiusSM:           seed.BorderRadius - 2,
		RadiusMD:           seed.BorderRadius,
		RadiusLG:           seed.BorderRadius + 2,
		RadiusXL:           seed.BorderRadius + 2, // Ant Design v5: borderRadiusOuter
		SpaceXXS:           seed.SizeUnit,
		SpaceXS:            seed.SizeUnit * 2,
		SpaceSM:            seed.SizeUnit * 3,
		SpaceMD:            seed.SizeUnit * 4,
		SpaceLG:            seed.SizeUnit * 6,
		SpaceXL:            seed.SizeUnit * 8,
		ShadowSM:           ShadowToken{Offset: math.NewVec2(0, 1), Blur: 2, Color: math.NewColor(0, 0, 0, 0.06)},
		ShadowMD:           ShadowToken{Offset: math.NewVec2(0, 6), Blur: 16, Color: math.NewColor(0, 0, 0, 0.08)},
		ShadowLG:           ShadowToken{Offset: math.NewVec2(0, 8), Blur: 24, Color: math.NewColor(0, 0, 0, 0.12)},
		MotionDurationFast: 100 * time.Millisecond,
		MotionDurationMid:  200 * time.Millisecond,
		MotionDurationSlow: 300 * time.Millisecond,
		MotionEaseOut:      "cubic-bezier(0.215, 0.61, 0.355, 1)",
		MotionEaseInOut:    "cubic-bezier(0.645, 0.045, 0.355, 1)",
	}
}

func deriveAlias(global GlobalToken) AliasToken {
	return AliasToken{
		ColorLink:        global.ColorPrimary,
		ColorLinkHover:   global.ColorPrimaryHover,
		ColorLinkActive:  global.ColorPrimaryActive,
		ColorFillAlter:   math.NewColor(0, 0, 0, 0.02),
		ColorFillContent: math.NewColor(0, 0, 0, 0.06),
		ColorSplit:       global.ColorBorder.WithAlpha(0.65),
		ControlHeightSM:  24,
		ControlHeight:    32,
		ControlHeightLG:  40,
	}
}

func deriveComponents(global GlobalToken, alias AliasToken) ComponentTokens {
	return ComponentTokens{
		Button: ButtonToken{
			HeightSM:    alias.ControlHeightSM,
			Height:      alias.ControlHeight,
			HeightLG:    alias.ControlHeightLG,
			PaddingH:    global.SpaceMD,
			BorderWidth: 1,
			Radius:      global.RadiusMD,
		},
		Input: InputToken{
			Height:      alias.ControlHeight,
			PaddingH:    global.SpaceSM,
			PaddingV:    global.SpaceXXS,
			BorderWidth: 1,
			Radius:      global.RadiusMD,
		},
		Card: CardToken{
			Padding:     global.SpaceLG,
			Radius:      global.RadiusLG,
			Shadow:      global.ShadowSM,
			BorderWidth: 1,
		},
		Modal: ModalToken{
			Width:     520,
			HeaderH:   56,
			Padding:   global.SpaceLG,
			Radius:    global.RadiusLG,
			MaskColor: global.ColorBgMask,
			Shadow:    global.ShadowLG,
		},
	}
}
