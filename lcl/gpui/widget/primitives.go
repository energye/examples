package widget

import (
	"github.com/energye/examples/lcl/gpui/core/math"
	"github.com/energye/examples/lcl/gpui/render/font"
	"github.com/energye/examples/lcl/gpui/render/pipeline"
)

// Box is a basic visual primitive used to validate lifecycle and state.
type Box struct {
	BaseWidget
	style       pipeline.BoxStyle
	interaction *InteractionController
	onClick     func()
}

// BoxProps stores mutable box properties.
type BoxProps struct {
	Style   pipeline.BoxStyle
	OnClick func()
}

// NewBox creates a box primitive.
func NewBox(style pipeline.BoxStyle) *Box {
	b := &Box{BaseWidget: NewBaseWidget(), style: style}
	b.SetOwner(b)
	b.interaction = NewInteractionController(b)
	return b
}

// Style returns the box style.
func (b *Box) Style() pipeline.BoxStyle {
	if b == nil {
		return pipeline.BoxStyle{}
	}
	return b.style
}

// SetStyle updates the box style.
func (b *Box) SetStyle(style pipeline.BoxStyle) {
	if b == nil {
		return
	}
	b.style = style
	b.Invalidate()
}

// SetOnClick sets the click callback.
func (b *Box) SetOnClick(handler func()) {
	if b == nil {
		return
	}
	b.onClick = handler
}

// SetProps updates box properties as a unit.
func (b *Box) SetProps(props BoxProps) {
	if b == nil {
		return
	}
	b.style = props.Style
	b.onClick = props.OnClick
	b.Invalidate()
}

// Render draws the box.
func (b *Box) Render(ctx *Context) {
	if b == nil || ctx == nil || ctx.Renderer == nil || !b.Visible() {
		return
	}
	style := b.style
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
	if b.interaction == nil {
		b.interaction = NewInteractionController(b)
	}
	b.interaction.SetOnClick(func(Event) {
		if b.onClick != nil {
			b.onClick()
		}
	})
	if b.interaction.HandleEvent(ctx, event) {
		return true
	}
	return event.Type == EventMouseDown || event.Type == EventMouseUp
}

// Text is a basic text primitive.
type Text struct {
	BaseWidget
	text       string
	color      math.Color
	font       *font.Font
	align      pipeline.TextAlign
	ellipsis   bool
	maxLines   int
	lineHeight float32
}

// TextProps stores mutable text properties.
type TextProps struct {
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
	t := &Text{BaseWidget: NewBaseWidget(), text: text}
	t.SetOwner(t)
	return t
}

// Text returns current text content.
func (t *Text) Text() string {
	if t == nil {
		return ""
	}
	return t.text
}

// SetText updates text content and invalidates rendering.
func (t *Text) SetText(text string) {
	if t == nil || t.text == text {
		return
	}
	t.text = text
	t.Invalidate()
}

// SetColor updates text color.
func (t *Text) SetColor(color math.Color) {
	if t == nil || t.color == color {
		return
	}
	t.color = color
	t.Invalidate()
}

// SetFont updates text font.
func (t *Text) SetFont(font *font.Font) {
	if t == nil || t.font == font {
		return
	}
	t.font = font
	t.Invalidate()
}

// SetAlign updates horizontal text alignment.
func (t *Text) SetAlign(align pipeline.TextAlign) {
	if t == nil || t.align == align {
		return
	}
	t.align = align
	t.Invalidate()
}

// SetEllipsis toggles ellipsis rendering.
func (t *Text) SetEllipsis(ellipsis bool) {
	if t == nil || t.ellipsis == ellipsis {
		return
	}
	t.ellipsis = ellipsis
	t.Invalidate()
}

// SetMaxLines updates maximum rendered lines.
func (t *Text) SetMaxLines(maxLines int) {
	if t == nil || t.maxLines == maxLines {
		return
	}
	t.maxLines = maxLines
	t.Invalidate()
}

// SetLineHeight updates text line height.
func (t *Text) SetLineHeight(lineHeight float32) {
	if t == nil || t.lineHeight == lineHeight {
		return
	}
	t.lineHeight = lineHeight
	t.Invalidate()
}

// Props returns current text properties.
func (t *Text) Props() TextProps {
	if t == nil {
		return TextProps{}
	}
	return TextProps{
		Text:       t.text,
		Color:      t.color,
		Font:       t.font,
		Align:      t.align,
		Ellipsis:   t.ellipsis,
		MaxLines:   t.maxLines,
		LineHeight: t.lineHeight,
	}
}

// SetProps updates text properties as a unit.
func (t *Text) SetProps(props TextProps) {
	if t == nil {
		return
	}
	t.text = props.Text
	t.color = props.Color
	t.font = props.Font
	t.align = props.Align
	t.ellipsis = props.Ellipsis
	t.maxLines = props.MaxLines
	t.lineHeight = props.LineHeight
	t.Invalidate()
}

// Measure returns text size using the context or widget font.
func (t *Text) Measure(ctx *Context, constraints Constraints) math.Vec2 {
	if t == nil {
		return math.Vec2{}
	}
	f := t.effectiveFont(ctx)
	lineHeight := t.lineHeight
	if lineHeight <= 0 && f != nil {
		lineHeight = f.LineHeight()
	}
	if lineHeight <= 0 {
		lineHeight = 16
	}
	width := t.Bounds().W
	if f != nil && t.text != "" {
		width = f.TextWidth(t.text)
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
	if t == nil || ctx == nil || ctx.Renderer == nil || !t.Visible() || t.text == "" {
		return
	}
	f := t.effectiveFont(ctx)
	if f == nil {
		return
	}
	color := t.color
	if color.A == 0 {
		color = ctx.Tokens.Global.ColorText
	}
	ctx.Renderer.DrawTextInRect(t.text, t.Bounds(), pipeline.TextOptions{
		Font:       f,
		Color:      color,
		Align:      t.align,
		MaxLines:   t.maxLines,
		Ellipsis:   t.ellipsis,
		LineHeight: t.lineHeight,
	})
}

func (t *Text) effectiveFont(ctx *Context) *font.Font {
	if t != nil && t.font != nil {
		return t.font
	}
	if ctx != nil {
		return ctx.Font
	}
	return nil
}
