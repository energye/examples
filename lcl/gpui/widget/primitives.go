package widget

import (
	"github.com/energye/examples/lcl/gpui/core/math"
	"github.com/energye/examples/lcl/gpui/render/font"
	"github.com/energye/examples/lcl/gpui/render/pipeline"
)

// Box is a basic visual primitive used to validate lifecycle and state.
type Box struct {
	BaseWidget
	Style   pipeline.BoxStyle
	OnClick func()
}

// NewBox creates a box primitive.
func NewBox(style pipeline.BoxStyle) *Box {
	b := &Box{BaseWidget: NewBaseWidget(), Style: style}
	b.SetOwner(b)
	return b
}

// Render draws the box.
func (b *Box) Render(ctx *Context) {
	if b == nil || ctx == nil || ctx.Renderer == nil || !b.Visible() {
		return
	}
	style := b.Style
	if b.HasState(StateHover) && style.BorderColor.A > 0 {
		style.BorderColor = ctx.Tokens.Global.ColorPrimaryHover
	}
	ctx.Renderer.DrawBox(b.Bounds(), style)
}

// HandleEvent handles box click state.
func (b *Box) HandleEvent(ctx *Context, event Event) bool {
	if b == nil || !b.Enabled() {
		return false
	}
	if event.Type == EventMouseUp && b.OnClick != nil {
		b.OnClick()
		return true
	}
	return event.Type == EventMouseDown || event.Type == EventMouseUp
}

// Text is a basic text primitive.
type Text struct {
	BaseWidget
	Text       string
	Color      math.Color
	Font       *font.Font
	Align      pipeline.TextAlign
	Ellipsis   bool
	MaxLines   int
	LineHeight float32
}

// NewText creates a text primitive.
func NewText(text string) *Text {
	t := &Text{BaseWidget: NewBaseWidget(), Text: text}
	t.SetOwner(t)
	return t
}

// Measure returns text size using the context or widget font.
func (t *Text) Measure(ctx *Context, constraints Constraints) math.Vec2 {
	if t == nil {
		return math.Vec2{}
	}
	f := t.effectiveFont(ctx)
	lineHeight := t.LineHeight
	if lineHeight <= 0 && f != nil {
		lineHeight = f.LineHeight()
	}
	if lineHeight <= 0 {
		lineHeight = 16
	}
	width := t.Bounds().W
	if f != nil && t.Text != "" {
		width = f.TextWidth(t.Text)
	}
	if width <= 0 {
		width = t.PreferredSize().X
	}
	if width <= 0 {
		width = constraints.Max.X
	}
	return ClampSize(math.NewVec2(width, lineHeight), constraints)
}

// Render draws text constrained to widget bounds.
func (t *Text) Render(ctx *Context) {
	if t == nil || ctx == nil || ctx.Renderer == nil || !t.Visible() || t.Text == "" {
		return
	}
	f := t.effectiveFont(ctx)
	if f == nil {
		return
	}
	color := t.Color
	if color.A == 0 {
		color = ctx.Tokens.Global.ColorText
	}
	ctx.Renderer.DrawTextInRect(t.Text, t.Bounds(), pipeline.TextOptions{
		Font:       f,
		Color:      color,
		Align:      t.Align,
		MaxLines:   t.MaxLines,
		Ellipsis:   t.Ellipsis,
		LineHeight: t.LineHeight,
	})
}

func (t *Text) effectiveFont(ctx *Context) *font.Font {
	if t != nil && t.Font != nil {
		return t.Font
	}
	if ctx != nil {
		return ctx.Font
	}
	return nil
}
