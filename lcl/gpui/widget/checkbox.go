package widget

import (
	"github.com/energye/examples/lcl/gpui/core/math"
	"github.com/energye/examples/lcl/gpui/render/font"
)

// Checkbox is a checkbox control with checked/unchecked/indeterminate states.
type Checkbox struct {
	ControlSurface
	text          string
	font          *font.Font
	checked       bool
	indeterminate bool
	onChange      func(checked bool)
}

// NewCheckbox creates a new checkbox.
func NewCheckbox(text string) *Checkbox {
	cb := &Checkbox{
		ControlSurface: *NewControlSurface(),
		text:           text,
	}
	cb.SetOwner(cb)
	cb.interaction.SetTarget(cb)
	cb.SetFocusable(true)
	return cb
}

// Text returns the checkbox text.
func (cb *Checkbox) Text() string {
	if cb == nil {
		return ""
	}
	return cb.text
}

// SetText updates the checkbox text.
func (cb *Checkbox) SetText(text string) {
	if cb == nil || cb.text == text {
		return
	}
	cb.text = text
	cb.Invalidate()
}

// Font returns the checkbox font.
func (cb *Checkbox) Font() *font.Font {
	if cb == nil {
		return nil
	}
	return cb.font
}

// SetFont updates the checkbox font.
func (cb *Checkbox) SetFont(f *font.Font) {
	if cb == nil || cb.font == f {
		return
	}
	cb.font = f
	cb.Invalidate()
}

// Checked returns whether the checkbox is checked.
func (cb *Checkbox) Checked() bool {
	return cb != nil && cb.checked
}

// SetChecked sets the checked state.
func (cb *Checkbox) SetChecked(checked bool) {
	if cb == nil || cb.checked == checked {
		return
	}
	cb.checked = checked
	cb.indeterminate = false
	cb.SetStateFlag(StateChecked, checked)
	cb.Invalidate()
}

// Indeterminate returns whether the checkbox is in indeterminate state.
func (cb *Checkbox) Indeterminate() bool {
	return cb != nil && cb.indeterminate
}

// SetIndeterminate sets the indeterminate state.
func (cb *Checkbox) SetIndeterminate(indeterminate bool) {
	if cb == nil || cb.indeterminate == indeterminate {
		return
	}
	cb.indeterminate = indeterminate
	if indeterminate {
		cb.checked = false
		cb.SetStateFlag(StateChecked, false)
	}
	cb.Invalidate()
}

// SetOnChange sets the change callback.
func (cb *Checkbox) SetOnChange(handler func(checked bool)) {
	if cb == nil {
		return
	}
	cb.onChange = handler
}

// Measure returns the checkbox size.
func (cb *Checkbox) Measure(ctx *Context, constraints Constraints) math.Vec2 {
	if cb == nil {
		return math.Vec2{}
	}
	tk := cb.Tokens(ctx)
	comp := tk.Components.Checkbox
	f := cb.effectiveFont(ctx)

	boxSize := comp.Size
	gap := comp.Gap

	textW := float32(0)
	textH := boxSize
	if f != nil && cb.text != "" {
		textW = f.TextWidth(cb.text)
		textH = f.LineHeight()
		if textH < boxSize {
			textH = boxSize
		}
	}

	width := boxSize + gap + textW
	height := textH
	return ClampSize(math.NewVec2(width, height), constraints)
}

// Render draws the checkbox.
func (cb *Checkbox) Render(ctx *Context) {
	if cb == nil || ctx == nil || ctx.Renderer == nil || !cb.Visible() {
		return
	}

	tk := cb.Tokens(ctx)
	comp := tk.Components.Checkbox
	global := tk.Global
	bounds := cb.Bounds()

	boxSize := comp.Size
	gap := comp.Gap
	radius := comp.Radius
	borderWidth := comp.BorderWidth

	// Checkbox box position
	boxX := bounds.X
	boxY := bounds.Y + (bounds.H-boxSize)/2
	box := math.NewRect(boxX, boxY, boxSize, boxSize)

	// Determine colors
	bgColor := global.ColorBgContainer
	borderColor := global.ColorBorder
	checkColor := global.ColorPrimary

	if cb.HasState(StateDisabled) {
		bgColor = global.ColorBgContainer
		borderColor = global.ColorBorder
		checkColor = global.ColorTextDisabled
	} else if cb.HasState(StateHover) {
		borderColor = checkColor
	}

	// Draw box
	if cb.checked || cb.indeterminate {
		ctx.Renderer.FillRoundRect(box, radius, checkColor)
		// Draw checkmark or indeterminate mark
		if cb.checked {
			drawCheckmark(ctx, box, math.NewColor(1, 1, 1, 1))
		} else {
			drawIndeterminate(ctx, box, math.NewColor(1, 1, 1, 1))
		}
	} else {
		ctx.Renderer.FillRoundRect(box, radius, bgColor)
		ctx.Renderer.StrokeRoundRect(box, radius, borderWidth, borderColor)
	}
	cb.RenderMotionOverlay(ctx, bounds)

	// Draw text
	f := cb.effectiveFont(ctx)
	if f != nil && cb.text != "" {
		textColor := global.ColorText
		if cb.HasState(StateDisabled) {
			textColor = global.ColorTextDisabled
		}
		textX := boxX + boxSize + gap
		textY := bounds.Y + (bounds.H-f.LineHeight())/2
		ctx.Renderer.DrawText(cb.text, textX, textY, f, textColor)
	}
}

// HandleEvent handles checkbox interaction.
func (cb *Checkbox) HandleEvent(ctx *Context, event Event) bool {
	if cb == nil || !cb.Enabled() {
		return false
	}
	if cb.interaction == nil {
		cb.interaction = NewInteractionController(cb)
	}
	cb.interaction.SetOnClick(func(Event) {
		if cb.indeterminate {
			cb.indeterminate = false
			cb.checked = true
		} else {
			cb.checked = !cb.checked
		}
		cb.SetStateFlag(StateChecked, cb.checked)
		if cb.onChange != nil {
			cb.onChange(cb.checked)
		}
	})
	return cb.interaction.HandleEvent(ctx, event)
}

func (cb *Checkbox) effectiveFont(ctx *Context) *font.Font {
	if cb != nil && cb.font != nil {
		return cb.font
	}
	if ctx != nil {
		return ctx.Font
	}
	return nil
}

func drawCheckmark(ctx *Context, rect math.Rect, color math.Color) {
	// Draw a simple checkmark using lines
	cx := rect.X + rect.W/2
	cy := rect.Y + rect.H/2
	size := rect.W * 0.3

	// Checkmark: two line segments
	ctx.Renderer.DrawLine(
		cx-size*0.8, cy,
		cx-size*0.1, cy+size*0.7,
		2, color,
	)
	ctx.Renderer.DrawLine(
		cx-size*0.1, cy+size*0.7,
		cx+size*0.8, cy-size*0.5,
		2, color,
	)
}

func drawIndeterminate(ctx *Context, rect math.Rect, color math.Color) {
	// Draw a horizontal line
	cx := rect.X + rect.W/2
	cy := rect.Y + rect.H/2
	size := rect.W * 0.4

	ctx.Renderer.DrawLine(cx-size, cy, cx+size, cy, 2, color)
}
