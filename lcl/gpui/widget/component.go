package widget

import (
	"github.com/energye/examples/lcl/gpui/core/math"
	"github.com/energye/examples/lcl/gpui/render/pipeline"
	"github.com/energye/examples/lcl/gpui/style/token"
)

// ComponentSize describes the standard Ant Design control sizes.
type ComponentSize int

const (
	SizeSmall ComponentSize = iota
	SizeMiddle
	SizeLarge
)

// ComponentVariant describes the common visual surface variants.
type ComponentVariant int

const (
	VariantOutlined ComponentVariant = iota
	VariantFilled
	VariantBorderless
	VariantText
	VariantSolid
)

// ComponentStatus describes semantic validation or feedback status.
type ComponentStatus int

const (
	StatusDefault ComponentStatus = iota
	StatusError
	StatusWarning
	StatusSuccess
)

// ComponentBase stores common component-level styling inputs.
type ComponentBase struct {
	BaseWidget
	controlSize   ComponentSize
	variant       ComponentVariant
	status        ComponentStatus
	tokenOverride *token.Tokens
}

// NewComponentBase creates a visible enabled component base.
func NewComponentBase() ComponentBase {
	return ComponentBase{
		BaseWidget:  NewBaseWidget(),
		controlSize: SizeMiddle,
		variant:     VariantOutlined,
		status:      StatusDefault,
	}
}

// ControlSize returns the standard control size.
func (c *ComponentBase) ControlSize() ComponentSize {
	if c == nil {
		return SizeMiddle
	}
	return c.controlSize
}

// SetControlSize updates the standard control size.
func (c *ComponentBase) SetControlSize(size ComponentSize) {
	if c == nil {
		return
	}
	c.controlSize = size
	c.Invalidate()
}

// Variant returns the visual variant.
func (c *ComponentBase) Variant() ComponentVariant {
	if c == nil {
		return VariantOutlined
	}
	return c.variant
}

// SetVariant updates the visual variant.
func (c *ComponentBase) SetVariant(variant ComponentVariant) {
	if c == nil {
		return
	}
	c.variant = variant
	c.Invalidate()
}

// Status returns the semantic status.
func (c *ComponentBase) Status() ComponentStatus {
	if c == nil {
		return StatusDefault
	}
	return c.status
}

// SetStatus updates the semantic status.
func (c *ComponentBase) SetStatus(status ComponentStatus) {
	if c == nil {
		return
	}
	c.status = status
	c.Invalidate()
}

// SetTokenOverride sets component-local tokens.
func (c *ComponentBase) SetTokenOverride(tokens *token.Tokens) {
	if c == nil {
		return
	}
	c.tokenOverride = tokens
	c.Invalidate()
}

// Tokens resolves tokens from override, context, or global current tokens.
func (c *ComponentBase) Tokens(ctx *Context) token.Tokens {
	if c != nil && c.tokenOverride != nil {
		return *c.tokenOverride
	}
	return TokensFromContext(ctx)
}

// ControlMetrics stores geometry and typography derived from tokens.
type ControlMetrics struct {
	Height       float32
	FontSize     float32
	LineHeight   float32
	PaddingH     float32
	PaddingV     float32
	Radius       float32
	BorderWidth  float32
	IconGap      float32
	MinTouchSize math.Vec2
}

// ControlPalette stores common state-aware colors.
type ControlPalette struct {
	Text         math.Color
	Background   math.Color
	Border       math.Color
	Placeholder  math.Color
	HoverBorder  math.Color
	ActiveBorder math.Color
	FocusBorder  math.Color
	StatusColor  math.Color
}

// ControlStyle stores the resolved common style for a component.
type ControlStyle struct {
	Metrics ControlMetrics
	Palette ControlPalette
}

// ResolveControlStyle resolves metrics and palette from context, tokens, and state.
func (c *ComponentBase) ResolveControlStyle(ctx *Context) ControlStyle {
	if c == nil {
		base := NewComponentBase()
		return base.ResolveControlStyle(ctx)
	}
	tokens := c.Tokens(ctx)
	return ControlStyle{
		Metrics: c.ResolveControlMetrics(tokens),
		Palette: c.ResolveControlPalette(tokens),
	}
}

// ResolveControlMetrics resolves shared geometry from tokens.
func (c *ComponentBase) ResolveControlMetrics(tokens token.Tokens) ControlMetrics {
	size := SizeMiddle
	if c != nil {
		size = c.controlSize
	}
	global := tokens.Global
	alias := tokens.Alias
	metrics := ControlMetrics{
		Height:       alias.ControlHeight,
		FontSize:     global.FontSize,
		LineHeight:   global.LineHeight,
		PaddingH:     global.SpaceSM,
		PaddingV:     global.SpaceXXS,
		Radius:       global.RadiusMD,
		BorderWidth:  1,
		IconGap:      global.SpaceXS,
		MinTouchSize: math.NewVec2(alias.ControlHeight, alias.ControlHeight),
	}

	switch size {
	case SizeSmall:
		metrics.Height = alias.ControlHeightSM
		metrics.FontSize = global.FontSizeSM
		metrics.PaddingH = global.SpaceXS
		metrics.Radius = global.RadiusSM
		metrics.MinTouchSize = math.NewVec2(alias.ControlHeightSM, alias.ControlHeightSM)
	case SizeLarge:
		metrics.Height = alias.ControlHeightLG
		metrics.FontSize = global.FontSizeLG
		metrics.PaddingH = global.SpaceMD
		metrics.Radius = global.RadiusLG
		metrics.MinTouchSize = math.NewVec2(alias.ControlHeightLG, alias.ControlHeightLG)
	}
	return metrics
}

// ResolveControlPalette resolves shared colors from tokens and state.
func (c *ComponentBase) ResolveControlPalette(tokens token.Tokens) ControlPalette {
	global := tokens.Global
	alias := tokens.Alias
	statusColor := controlStatusColor(tokens, StatusDefault)
	status := StatusDefault
	variant := VariantOutlined
	state := StateNormal
	if c != nil {
		status = c.status
		variant = c.variant
		state = c.State()
	}
	if status != StatusDefault {
		statusColor = controlStatusColor(tokens, status)
	}

	palette := ControlPalette{
		Text:         global.ColorText,
		Background:   global.ColorBgContainer,
		Border:       global.ColorBorder,
		Placeholder:  global.ColorTextDisabled,
		HoverBorder:  global.ColorBorderHover,
		ActiveBorder: global.ColorPrimaryActive,
		FocusBorder:  statusColor,
		StatusColor:  statusColor,
	}
	if status != StatusDefault {
		palette.Border = statusColor
		palette.HoverBorder = statusColor.Lighten(0.08)
		palette.ActiveBorder = statusColor.Darken(0.08)
	}

	switch variant {
	case VariantFilled:
		palette.Background = alias.ColorFillAlter
		palette.Border = math.Color{}
	case VariantBorderless, VariantText:
		palette.Background = math.Color{}
		palette.Border = math.Color{}
	case VariantSolid:
		palette.Background = statusColor
		palette.Border = statusColor
		palette.Text = global.ColorTextLight
		palette.HoverBorder = statusColor.Lighten(0.08)
		palette.ActiveBorder = statusColor.Darken(0.08)
	}

	if state&StateHover != 0 {
		palette.Border = palette.HoverBorder
		if variant == VariantSolid {
			palette.Background = statusColor.Lighten(0.04)
		}
	}
	if state&StateActive != 0 {
		palette.Border = palette.ActiveBorder
		if variant == VariantSolid {
			palette.Background = statusColor.Darken(0.06)
		}
	}
	if state&StateFocus != 0 {
		palette.Border = palette.FocusBorder
	}
	if state&StateDisabled != 0 {
		palette.Text = global.ColorTextDisabled
		palette.Placeholder = global.ColorTextDisabled
		palette.Background = alias.ColorFillAlter
		palette.Border = global.ColorBorder.WithAlpha(0.65)
	}

	return palette
}

// BoxStyle converts the resolved control style into a renderer box style.
func (s ControlStyle) BoxStyle() pipeline.BoxStyle {
	return pipeline.BoxStyle{
		Background:  s.Palette.Background,
		BorderColor: s.Palette.Border,
		BorderWidth: s.Metrics.BorderWidth,
		Radius:      s.Metrics.Radius,
	}
}

// TokensFromContext returns context tokens or the globally active tokens.
func TokensFromContext(ctx *Context) token.Tokens {
	if ctx == nil || ctx.Tokens.Seed.FontSize <= 0 {
		return token.Current()
	}
	return ctx.Tokens
}

func controlStatusColor(tokens token.Tokens, status ComponentStatus) math.Color {
	switch status {
	case StatusError:
		return tokens.Global.ColorError
	case StatusWarning:
		return tokens.Global.ColorWarning
	case StatusSuccess:
		return tokens.Global.ColorSuccess
	default:
		return tokens.Global.ColorPrimary
	}
}
