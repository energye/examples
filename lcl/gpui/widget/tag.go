package widget

import (
	"github.com/energye/examples/lcl/gpui/core/math"
	"github.com/energye/examples/lcl/gpui/render/font"
	"github.com/energye/examples/lcl/gpui/style/token"
)

// TagColor represents a predefined tag color.
type TagColor int

const (
	TagDefault TagColor = iota
	TagBlue
	TagGreen
	TagRed
	TagOrange
	TagCyan
	TagPurple
)

// Tag is a small label for categorizing or marking.
type Tag struct {
	BaseWidget
	text      string
	font      *font.Font
	color     TagColor
	closable  bool
	onClose   func()
	onClick   func()
	interaction *InteractionController
}

// NewTag creates a new tag.
func NewTag(text string) *Tag {
	t := &Tag{
		BaseWidget: NewBaseWidget(),
		text:       text,
		color:      TagDefault,
	}
	t.SetOwner(t)
	t.interaction = NewInteractionController(t)
	return t
}

// Text returns the tag text.
func (t *Tag) Text() string {
	if t == nil {
		return ""
	}
	return t.text
}

// SetText updates the tag text.
func (t *Tag) SetText(text string) {
	if t == nil || t.text == text {
		return
	}
	t.text = text
	t.Invalidate()
}

// Font returns the tag font.
func (t *Tag) Font() *font.Font {
	if t == nil {
		return nil
	}
	return t.font
}

// SetFont updates the tag font.
func (t *Tag) SetFont(f *font.Font) {
	if t == nil || t.font == f {
		return
	}
	t.font = f
	t.Invalidate()
}

// Color returns the tag color.
func (t *Tag) Color() TagColor {
	if t == nil {
		return TagDefault
	}
	return t.color
}

// SetColor updates the tag color.
func (t *Tag) SetColor(color TagColor) {
	if t == nil || t.color == color {
		return
	}
	t.color = color
	t.Invalidate()
}

// Closable returns whether the tag is closable.
func (t *Tag) Closable() bool {
	return t != nil && t.closable
}

// SetClosable toggles the close button.
func (t *Tag) SetClosable(closable bool) {
	if t == nil || t.closable == closable {
		return
	}
	t.closable = closable
	t.Invalidate()
}

// SetOnClose sets the close callback.
func (t *Tag) SetOnClose(handler func()) {
	if t == nil {
		return
	}
	t.onClose = handler
}

// SetOnClick sets the click callback.
func (t *Tag) SetOnClick(handler func()) {
	if t == nil {
		return
	}
	t.onClick = handler
}

// Measure returns the tag size.
func (t *Tag) Measure(ctx *Context, constraints Constraints) math.Vec2 {
	if t == nil {
		return math.Vec2{}
	}
	tk := TokensFromContext(ctx)
	comp := tk.Components.Tag
	f := t.effectiveFont(ctx)

	height := comp.Height
	paddingH := comp.PaddingH

	textW := float32(0)
	if f != nil && t.text != "" {
		textW = f.TextWidth(t.text)
	}

	width := paddingH*2 + textW
	if t.closable {
		width += height // Space for close button
	}
	return ClampSize(math.NewVec2(width, height), constraints)
}

// Render draws the tag.
func (t *Tag) Render(ctx *Context) {
	if t == nil || ctx == nil || ctx.Renderer == nil || !t.Visible() {
		return
	}

	tk := TokensFromContext(ctx)
	comp := tk.Components.Tag
	global := tk.Global
	bounds := t.Bounds()

	// Tag colors
	bgColor, textColor, borderColor := t.tagColors(global)

	// Draw background
	ctx.Renderer.FillRoundRect(bounds, comp.Radius, bgColor)
	ctx.Renderer.StrokeRoundRect(bounds, comp.Radius, comp.BorderWidth, borderColor)

	// Draw text
	f := t.effectiveFont(ctx)
	if f != nil && t.text != "" {
		textX := bounds.X + comp.PaddingH
		textY := bounds.Y + (bounds.H-f.LineHeight())/2
		ctx.Renderer.DrawText(t.text, textX, textY, f, textColor)
	}

	// Draw close button
	if t.closable {
		closeSize := comp.Height * 0.5
		closeX := bounds.X + bounds.W - comp.PaddingH - closeSize
		closeY := bounds.Y + (bounds.H-closeSize)/2
		ctx.Renderer.DrawLine(closeX, closeY, closeX+closeSize, closeY+closeSize, 1.5, textColor)
		ctx.Renderer.DrawLine(closeX+closeSize, closeY, closeX, closeY+closeSize, 1.5, textColor)
	}
}

// HandleEvent handles tag interaction.
func (t *Tag) HandleEvent(ctx *Context, event Event) bool {
	if t == nil || !t.Enabled() {
		return false
	}
	if t.interaction == nil {
		t.interaction = NewInteractionController(t)
	}
	t.interaction.SetOnClick(func(e Event) {
		// Check if click is on close button
		if t.closable {
			tk := TokensFromContext(ctx)
			comp := tk.Components.Tag
			closeSize := comp.Height * 0.5
			closeX := t.Bounds().X + t.Bounds().W - comp.PaddingH - closeSize
			if e.LocalX >= closeX {
				if t.onClose != nil {
					t.onClose()
				}
				return
			}
		}
		if t.onClick != nil {
			t.onClick()
		}
	})
	return t.interaction.HandleEvent(ctx, event)
}

func (t *Tag) effectiveFont(ctx *Context) *font.Font {
	if t != nil && t.font != nil {
		return t.font
	}
	if ctx != nil {
		return ctx.Font
	}
	return nil
}

func (t *Tag) tagColors(global token.GlobalToken) (bg, text, border math.Color) {
	switch t.color {
	case TagBlue:
		return global.ColorPrimaryPalette[1], global.ColorPrimaryPalette[6], global.ColorPrimaryPalette[3]
	case TagGreen:
		return global.ColorSuccessPalette[1], global.ColorSuccessPalette[6], global.ColorSuccessPalette[3]
	case TagRed:
		return global.ColorErrorPalette[1], global.ColorErrorPalette[6], global.ColorErrorPalette[3]
	case TagOrange:
		return global.ColorWarningPalette[1], global.ColorWarningPalette[6], global.ColorWarningPalette[3]
	case TagCyan:
		return math.NewColorFromHSL(180, 0.8, 0.95, 1), math.NewColorFromHSL(180, 0.8, 0.35, 1), math.NewColorFromHSL(180, 0.6, 0.8, 1)
	case TagPurple:
		return math.NewColorFromHSL(270, 0.8, 0.95, 1), math.NewColorFromHSL(270, 0.8, 0.35, 1), math.NewColorFromHSL(270, 0.6, 0.8, 1)
	default:
		return math.NewColor(0.95, 0.95, 0.95, 1), global.ColorText, math.NewColor(0.85, 0.85, 0.85, 1)
	}
}
