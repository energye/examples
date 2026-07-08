package widget

import (
	"github.com/energye/examples/lcl/gpui/core/math"
	"github.com/energye/examples/lcl/gpui/render/font"
	"github.com/energye/examples/lcl/gpui/render/pipeline"
)

// ButtonKind describes the semantic Ant Design button variants.
type ButtonKind int

const (
	ButtonDefault ButtonKind = iota
	ButtonPrimary
	ButtonDashed
	ButtonText
	ButtonLink
)

// Button is the first real Ant Design-style control built on the phase 2 core.
type Button struct {
	ControlSurface
	text    string
	font    *font.Font
	kind    ButtonKind
	danger  bool
	ghost   bool
	block   bool
	loading bool
	onClick func()
}

// ButtonProps stores mutable button properties.
type ButtonProps struct {
	Text    string
	Font    *font.Font
	Kind    ButtonKind
	Danger  bool
	Ghost   bool
	Block   bool
	Loading bool
	OnClick func()
}

// NewButton creates a token-driven button.
func NewButton(text string) *Button {
	b := &Button{
		ControlSurface: *NewControlSurface(),
		text:           text,
		kind:           ButtonDefault,
	}
	b.SetOwner(b)
	b.interaction.SetTarget(b)
	b.SetFocusable(true)
	return b
}

// Text returns button text.
func (b *Button) Text() string {
	if b == nil {
		return ""
	}
	return b.text
}

// SetText updates button text.
func (b *Button) SetText(text string) {
	if b == nil || b.text == text {
		return
	}
	b.text = text
	b.Invalidate()
}

// Font returns button font.
func (b *Button) Font() *font.Font {
	if b == nil {
		return nil
	}
	return b.font
}

// SetFont updates button font.
func (b *Button) SetFont(font *font.Font) {
	if b == nil || b.font == font {
		return
	}
	b.font = font
	b.Invalidate()
}

// Kind returns button kind.
func (b *Button) Kind() ButtonKind {
	if b == nil {
		return ButtonDefault
	}
	return b.kind
}

// SetKind updates button kind.
func (b *Button) SetKind(kind ButtonKind) {
	if b == nil || b.kind == kind {
		return
	}
	b.kind = kind
	b.Invalidate()
}

// SetDanger toggles danger style.
func (b *Button) SetDanger(danger bool) {
	if b == nil || b.danger == danger {
		return
	}
	b.danger = danger
	b.Invalidate()
}

// SetGhost toggles ghost style.
func (b *Button) SetGhost(ghost bool) {
	if b == nil || b.ghost == ghost {
		return
	}
	b.ghost = ghost
	b.Invalidate()
}

// SetBlock toggles block width behavior.
func (b *Button) SetBlock(block bool) {
	if b == nil || b.block == block {
		return
	}
	b.block = block
	b.Invalidate()
}

// SetLoading toggles loading state.
func (b *Button) SetLoading(loading bool) {
	if b == nil {
		return
	}
	b.loading = loading
	b.SetStateFlag(StateLoading, loading)
}

// Loading reports whether the button is loading.
func (b *Button) Loading() bool {
	return b != nil && b.loading
}

// SetOnClick sets the click callback.
func (b *Button) SetOnClick(handler func()) {
	if b == nil {
		return
	}
	b.onClick = handler
}

// Props returns current button properties.
func (b *Button) Props() ButtonProps {
	if b == nil {
		return ButtonProps{}
	}
	return ButtonProps{
		Text:    b.text,
		Font:    b.font,
		Kind:    b.kind,
		Danger:  b.danger,
		Ghost:   b.ghost,
		Block:   b.block,
		Loading: b.loading,
		OnClick: b.onClick,
	}
}

// SetProps updates button properties as a unit.
func (b *Button) SetProps(props ButtonProps) {
	if b == nil {
		return
	}
	b.text = props.Text
	b.font = props.Font
	b.kind = props.Kind
	b.danger = props.Danger
	b.ghost = props.Ghost
	b.block = props.Block
	b.onClick = props.OnClick
	b.SetLoading(props.Loading)
	b.Invalidate()
}

// Measure returns token-based button size.
func (b *Button) Measure(ctx *Context, constraints Constraints) math.Vec2 {
	if b == nil {
		return math.Vec2{}
	}
	style := b.buttonStyle(ctx)
	width := b.PreferredSize().X
	if width <= 0 {
		width = b.Bounds().W
	}
	if width <= 0 {
		width = style.Metrics.MinTouchSize.X
		if textWidth := b.textWidth(ctx); textWidth > 0 {
			width = textWidth + style.Metrics.PaddingH*2
			if b.loading {
				width += style.Metrics.Height*0.45 + style.Metrics.IconGap
			}
		}
		if width < style.Metrics.Height {
			width = style.Metrics.Height
		}
	}
	if b.block && constraints.Max.X > 0 {
		width = constraints.Max.X
	}

	height := b.PreferredSize().Y
	if height <= 0 {
		height = b.Bounds().H
	}
	if height <= 0 {
		height = style.Metrics.Height
	}
	return ClampSize(math.NewVec2(width, height), constraints)
}

// Render draws the button frame and centered label.
func (b *Button) Render(ctx *Context) {
	if b == nil || ctx == nil || ctx.Renderer == nil || !b.Visible() {
		return
	}
	style := b.buttonStyle(ctx)
	bounds := b.Bounds()
	ctx.Renderer.DrawBox(bounds, style.BoxStyle())

	// Draw focus ring when focused
	if b.HasState(StateFocus) {
		focusRing := bounds.Expand(2)
		ctx.Renderer.StrokeRoundRect(focusRing, style.Metrics.Radius+2, 2, ctx.Tokens.Global.ColorPrimary)
	}

	f := b.effectiveFont(ctx)
	if f == nil || b.text == "" {
		return
	}
	textRect := bounds.Shrink(style.Metrics.PaddingH, style.Metrics.PaddingV)
	ctx.Renderer.DrawTextInRect(b.text, textRect, pipeline.TextOptions{
		Font:       f,
		Color:      style.Palette.Text,
		Align:      pipeline.TextAlignCenter,
		MaxLines:   1,
		Ellipsis:   true,
		LineHeight: f.LineHeight(),
	})
}

// HandleEvent applies shared button interaction.
func (b *Button) HandleEvent(ctx *Context, event Event) bool {
	if b == nil || !b.Enabled() {
		return false
	}
	if b.interaction == nil {
		b.interaction = NewInteractionController(b)
	}
	b.interaction.SetOnClick(func(Event) {
		if b.onClick != nil {
			b.onClick()
		}
	})
	return b.interaction.HandleEvent(ctx, event)
}

func (b *Button) buttonStyle(ctx *Context) ControlStyle {
	base := b.ComponentBase
	base.variant = VariantOutlined
	switch b.kind {
	case ButtonPrimary:
		base.variant = VariantSolid
	case ButtonText, ButtonLink:
		base.variant = VariantText
	}
	if b.danger {
		base.status = StatusError
	}
	style := base.ResolveControlStyle(ctx)
	if b.kind == ButtonLink {
		style.Palette.Background = math.Color{}
		style.Palette.Border = math.Color{}
		style.Palette.Text = style.Palette.StatusColor
		if b.HasState(StateHover) {
			style.Palette.Text = style.Palette.StatusColor.Lighten(0.08)
		}
	}
	if b.danger && b.kind != ButtonPrimary {
		style.Palette.Text = style.Palette.StatusColor
	}
	if b.ghost {
		style.Palette.Background = math.Color{}
	}
	if b.kind == ButtonText {
		style.Palette.Border = math.Color{}
	}
	return style
}

func (b *Button) effectiveFont(ctx *Context) *font.Font {
	if b != nil && b.font != nil {
		return b.font
	}
	if ctx != nil {
		return ctx.Font
	}
	return nil
}

func (b *Button) textWidth(ctx *Context) float32 {
	f := b.effectiveFont(ctx)
	if f == nil || b.text == "" {
		return 0
	}
	return f.TextWidth(b.text)
}
