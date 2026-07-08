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
	ColorPrimaryPalette Palette
	ColorSuccessPalette Palette
	ColorWarningPalette Palette
	ColorErrorPalette   Palette
	ColorInfoPalette    Palette

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
	Button       ButtonToken
	Input        InputToken
	Card         CardToken
	Modal        ModalToken
	Checkbox     CheckboxToken
	Radio        RadioToken
	Switch       SwitchToken
	Select       SelectToken
	Tag          TagToken
	Tooltip      TooltipToken
	Table        TableToken
	Menu         MenuToken
	Tabs         TabsToken
	Badge        BadgeToken
	Avatar       AvatarToken
	Alert        AlertToken
	Progress     ProgressToken
	Pagination   PaginationToken
	Breadcrumb   BreadcrumbToken
	Steps        StepsToken
	Divider      DividerToken
	Collapse     CollapseToken
	Timeline     TimelineToken
	Message      MessageToken
	Notification NotificationToken
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

type CheckboxToken struct {
	Size        float32
	BorderWidth float32
	Radius      float32
	Gap         float32 // Gap between checkbox and label
}

type RadioToken struct {
	Size        float32
	BorderWidth float32
	Gap         float32
}

type SwitchToken struct {
	Height      float32
	MinWidth    float32
	InnerSize   float32
	InnerMargin float32
	Radius      float32
}

type SelectToken struct {
	Height       float32
	HeightSM     float32
	HeightLG     float32
	PaddingH     float32
	BorderWidth  float32
	Radius       float32
	DropdownMaxH float32
}

type TagToken struct {
	HeightSM    float32
	Height      float32
	HeightLG    float32
	PaddingH    float32
	BorderWidth float32
	Radius      float32
	FontSizeSM  float32
}

type TooltipToken struct {
	MaxWidth  float32
	PaddingH  float32
	PaddingV  float32
	Radius    float32
	ArrowSize float32
}

type TableToken struct {
	HeaderBg         math.Color
	HeaderColor      math.Color
	RowHoverBg       math.Color
	BorderColor      math.Color
	CellPaddingH     float32
	CellPaddingV     float32
	FontSize         float32
	HeaderFontWeight float32
}

type MenuToken struct {
	Height       float32
	ItemPaddingH float32
	ItemMarginB  float32
	IconSize     float32
	SubIconSize  float32
	Radius       float32
	FontSize     float32
}

type TabsToken struct {
	Height       float32
	CardHeight   float32
	FontSize     float32
	BarThickness float32
	Gap          float32
	PaddingH     float32
}

type BadgeToken struct {
	Height   float32
	HeightSM float32
	MinWidth float32
	Radius   float32
	FontSize float32
	DotSize  float32
}

type AvatarToken struct {
	SizeSM     float32
	Size       float32
	SizeLG     float32
	FontSizeSM float32
	FontSize   float32
	FontSizeLG float32
	Radius     float32
	GroupSpace float32
}

type AlertToken struct {
	PaddingH float32
	PaddingV float32
	Radius   float32
	FontSize float32
	IconSize float32
}

type ProgressToken struct {
	SizeSM             float32
	Size               float32
	FontSize           float32
	RailHeight         float32
	CircleTextFontSize float32
}

type PaginationToken struct {
	Height   float32
	HeightSM float32
	MinWidth float32
	FontSize float32
	Radius   float32
	Gap      float32
}

type BreadcrumbToken struct {
	FontSize        float32
	FontSizeIcon    float32
	SeparatorMargin float32
	LinkColor       math.Color
	SeparatorColor  math.Color
}

type StepsToken struct {
	Height        float32
	DotSize       float32
	IconSize      float32
	TitleFontSize float32
	DescFontSize  float32
	Gap           float32
}

type DividerToken struct {
	Thickness   float32
	MarginH     float32
	MarginV     float32
	TextPadding float32
}

type CollapseToken struct {
	HeaderPaddingH  float32
	HeaderPaddingV  float32
	ContentPaddingH float32
	ContentPaddingV float32
	Radius          float32
	BorderWidth     float32
}

type TimelineToken struct {
	DotSize   float32
	LineWidth float32
	Gap       float32
	PaddingB  float32
}

type MessageToken struct {
	PaddingH    float32
	PaddingV    float32
	Radius      float32
	FontSize    float32
	IconSize    float32
	ContentMaxW float32
}

type NotificationToken struct {
	Width    float32
	PaddingH float32
	PaddingV float32
	Radius   float32
	FontSize float32
	IconSize float32
	MarginB  float32
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
	seed = normalizeSeed(seed)
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

func normalizeSeed(seed SeedToken) SeedToken {
	defaults := DefaultSeed()
	if isZeroColor(seed.ColorPrimary) {
		seed.ColorPrimary = defaults.ColorPrimary
	}
	if isZeroColor(seed.ColorSuccess) {
		seed.ColorSuccess = defaults.ColorSuccess
	}
	if isZeroColor(seed.ColorWarning) {
		seed.ColorWarning = defaults.ColorWarning
	}
	if isZeroColor(seed.ColorError) {
		seed.ColorError = defaults.ColorError
	}
	if isZeroColor(seed.ColorInfo) {
		seed.ColorInfo = defaults.ColorInfo
	}
	if seed.FontFamily == "" {
		seed.FontFamily = defaults.FontFamily
	}
	if seed.FontSize <= 0 {
		seed.FontSize = defaults.FontSize
	}
	if seed.LineHeight <= 0 {
		seed.LineHeight = defaults.LineHeight
	}
	if seed.SizeUnit <= 0 {
		seed.SizeUnit = defaults.SizeUnit
	}
	if seed.BorderRadius < 0 || seed.BorderRadius == 0 {
		seed.BorderRadius = defaults.BorderRadius
	}
	return seed
}

func isZeroColor(color math.Color) bool {
	return color.R == 0 && color.G == 0 && color.B == 0 && color.A == 0
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
		ColorPrimaryPalette: GeneratePaletteForMode(seed.ColorPrimary, mode),
		ColorSuccessPalette: GeneratePaletteForMode(seed.ColorSuccess, mode),
		ColorWarningPalette: GeneratePaletteForMode(seed.ColorWarning, mode),
		ColorErrorPalette:   GeneratePaletteForMode(seed.ColorError, mode),
		ColorInfoPalette:    GeneratePaletteForMode(seed.ColorInfo, mode),

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
		Checkbox: CheckboxToken{
			Size:        16,
			BorderWidth: 1,
			Radius:      global.RadiusSM,
			Gap:         global.SpaceXS,
		},
		Radio: RadioToken{
			Size:        16,
			BorderWidth: 1,
			Gap:         global.SpaceXS,
		},
		Switch: SwitchToken{
			Height:      22,
			MinWidth:    44,
			InnerSize:   16,
			InnerMargin: 3,
			Radius:      9999,
		},
		Select: SelectToken{
			Height:       alias.ControlHeight,
			HeightSM:     alias.ControlHeightSM,
			HeightLG:     alias.ControlHeightLG,
			PaddingH:     global.SpaceSM,
			BorderWidth:  1,
			Radius:       global.RadiusMD,
			DropdownMaxH: 256,
		},
		Tag: TagToken{
			HeightSM:    20,
			Height:      24,
			HeightLG:    28,
			PaddingH:    global.SpaceXXS,
			BorderWidth: 1,
			Radius:      global.RadiusSM,
			FontSizeSM:  global.FontSizeSM,
		},
		Tooltip: TooltipToken{
			MaxWidth:  250,
			PaddingH:  global.SpaceSM,
			PaddingV:  global.SpaceXXS,
			Radius:    global.RadiusSM,
			ArrowSize: 8,
		},
		Table: TableToken{
			HeaderBg:         global.ColorBgElevated,
			HeaderColor:      global.ColorText,
			RowHoverBg:       global.ColorBgContainer,
			BorderColor:      global.ColorBorder,
			CellPaddingH:     global.SpaceSM,
			CellPaddingV:     global.SpaceSM,
			FontSize:         global.FontSize,
			HeaderFontWeight: 600,
		},
		Menu: MenuToken{
			Height:       40,
			ItemPaddingH: global.SpaceMD,
			ItemMarginB:  global.SpaceXXS,
			IconSize:     14,
			SubIconSize:  10,
			Radius:       global.RadiusMD,
			FontSize:     global.FontSize,
		},
		Tabs: TabsToken{
			Height:       40,
			CardHeight:   40,
			FontSize:     global.FontSize,
			BarThickness: 2,
			Gap:          global.SpaceXL,
			PaddingH:     global.SpaceSM,
		},
		Badge: BadgeToken{
			Height:   20,
			HeightSM: 16,
			MinWidth: 20,
			Radius:   9999,
			FontSize: 12,
			DotSize:  6,
		},
		Avatar: AvatarToken{
			SizeSM:     24,
			Size:       32,
			SizeLG:     40,
			FontSizeSM: 12,
			FontSize:   14,
			FontSizeLG: 16,
			Radius:     global.RadiusMD,
			GroupSpace: -8,
		},
		Alert: AlertToken{
			PaddingH: global.SpaceMD,
			PaddingV: global.SpaceSM,
			Radius:   global.RadiusMD,
			FontSize: global.FontSize,
			IconSize: 16,
		},
		Progress: ProgressToken{
			SizeSM:             32,
			Size:               40,
			FontSize:           global.FontSize,
			RailHeight:         8,
			CircleTextFontSize: 24,
		},
		Pagination: PaginationToken{
			Height:   alias.ControlHeight,
			HeightSM: alias.ControlHeightSM,
			MinWidth: alias.ControlHeight,
			FontSize: global.FontSize,
			Radius:   global.RadiusMD,
			Gap:      global.SpaceXXS,
		},
		Breadcrumb: BreadcrumbToken{
			FontSize:        global.FontSize,
			FontSizeIcon:    global.FontSizeSM,
			SeparatorMargin: global.SpaceXXS,
			LinkColor:       global.ColorTextSecondary,
			SeparatorColor:  global.ColorTextDisabled,
		},
		Steps: StepsToken{
			Height:        32,
			DotSize:       8,
			IconSize:      16,
			TitleFontSize: global.FontSize,
			DescFontSize:  global.FontSizeSM,
			Gap:           global.SpaceSM,
		},
		Divider: DividerToken{
			Thickness:   1,
			MarginH:     0,
			MarginV:     global.SpaceLG,
			TextPadding: global.SpaceSM,
		},
		Collapse: CollapseToken{
			HeaderPaddingH:  global.SpaceMD,
			HeaderPaddingV:  global.SpaceSM,
			ContentPaddingH: global.SpaceMD,
			ContentPaddingV: global.SpaceMD,
			Radius:          global.RadiusMD,
			BorderWidth:     1,
		},
		Timeline: TimelineToken{
			DotSize:   8,
			LineWidth: 2,
			Gap:       global.SpaceSM,
			PaddingB:  global.SpaceXXS,
		},
		Message: MessageToken{
			PaddingH:    global.SpaceMD,
			PaddingV:    global.SpaceSM,
			Radius:      global.RadiusMD,
			FontSize:    global.FontSize,
			IconSize:    16,
			ContentMaxW: 480,
		},
		Notification: NotificationToken{
			Width:    384,
			PaddingH: global.SpaceMD,
			PaddingV: global.SpaceMD,
			Radius:   global.RadiusLG,
			FontSize: global.FontSize,
			IconSize: 24,
			MarginB:  global.SpaceSM,
		},
	}
}
