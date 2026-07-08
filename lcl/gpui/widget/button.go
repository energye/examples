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
		Text:           text,
		Kind:           ButtonDefault,
	}
	b.SetOwner(b)
	b.Interaction.SetTarget(b)
	b.SetFocusable(true)
	return b
}

// SetText updates button text.
func (b *Button) SetText(text string) {
	if b == nil {
		return
	}
	b.Text = text
	b.Invalidate()
}

// SetLoading toggles loading state.
func (b *Button) SetLoading(loading bool) {
	if b == nil {
		return
	}
	b.Loading = loading
	b.SetStateFlag(StateLoading, loading)
}

// SetOnClick sets the click callback.
func (b *Button) SetOnClick(handler func()) {
	if b == nil {
		return
	}
	b.OnClick = handler
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
			if b.Loading {
				width += style.Metrics.Height*0.45 + style.Metrics.IconGap
			}
		}
		if width < style.Metrics.Height {
			width = style.Metrics.Height
		}
	}
	if b.Block && constraints.Max.X > 0 {
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

	f := b.effectiveFont(ctx)
	if f == nil || b.Text == "" {
		return
	}
	textRect := bounds.Shrink(style.Metrics.PaddingH, style.Metrics.PaddingV)
	ctx.Renderer.DrawTextInRect(b.Text, textRect, pipeline.TextOptions{
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
	if b.Interaction == nil {
		b.Interaction = NewInteractionController(b)
	}
	b.Interaction.SetOnClick(func(Event) {
		if b.OnClick != nil {
			b.OnClick()
		}
	})
	return b.Interaction.HandleEvent(ctx, event)
}

func (b *Button) buttonStyle(ctx *Context) ControlStyle {
	base := b.ComponentBase
	base.variant = VariantOutlined
	switch b.Kind {
	case ButtonPrimary:
		base.variant = VariantSolid
	case ButtonText, ButtonLink:
		base.variant = VariantText
	}
	if b.Danger {
		base.status = StatusError
	}
	style := base.ResolveControlStyle(ctx)
	if b.Kind == ButtonLink {
		style.Palette.Background = math.Color{}
		style.Palette.Border = math.Color{}
		style.Palette.Text = style.Palette.StatusColor
		if b.HasState(StateHover) {
			style.Palette.Text = style.Palette.StatusColor.Lighten(0.08)
		}
	}
	if b.Danger && b.Kind != ButtonPrimary {
		style.Palette.Text = style.Palette.StatusColor
	}
	if b.Ghost {
		style.Palette.Background = math.Color{}
	}
	if b.Kind == ButtonText {
		style.Palette.Border = math.Color{}
	}
	return style
}

func (b *Button) effectiveFont(ctx *Context) *font.Font {
	if b != nil && b.Font != nil {
		return b.Font
	}
	if ctx != nil {
		return ctx.Font
	}
	return nil
}

func (b *Button) textWidth(ctx *Context) float32 {
	f := b.effectiveFont(ctx)
	if f == nil || b.Text == "" {
		return 0
	}
	return f.TextWidth(b.Text)
}
