package widget

import (
	"github.com/energye/examples/lcl/gpui/core/math"
	"github.com/energye/examples/lcl/gpui/render/font"
)

// Radio is a radio button control.
type Radio struct {
	ControlSurface
	text     string
	font     *font.Font
	checked  bool
	onChange func(checked bool)
}

// NewRadio creates a new radio button.
func NewRadio(text string) *Radio {
	r := &Radio{
		ControlSurface: *NewControlSurface(),
		text:           text,
	}
	r.SetOwner(r)
	r.interaction.SetTarget(r)
	r.SetFocusable(true)
	return r
}

// Text returns the radio text.
func (r *Radio) Text() string {
	if r == nil {
		return ""
	}
	return r.text
}

// SetText updates the radio text.
func (r *Radio) SetText(text string) {
	if r == nil || r.text == text {
		return
	}
	r.text = text
	r.Invalidate()
}

// Font returns the radio font.
func (r *Radio) Font() *font.Font {
	if r == nil {
		return nil
	}
	return r.font
}

// SetFont updates the radio font.
func (r *Radio) SetFont(f *font.Font) {
	if r == nil || r.font == f {
		return
	}
	r.font = f
	r.Invalidate()
}

// Checked returns whether the radio is checked.
func (r *Radio) Checked() bool {
	return r != nil && r.checked
}

// SetChecked sets the checked state.
func (r *Radio) SetChecked(checked bool) {
	if r == nil || r.checked == checked {
		return
	}
	r.checked = checked
	r.SetStateFlag(StateChecked, checked)
	r.Invalidate()
}

// SetOnChange sets the change callback.
func (r *Radio) SetOnChange(handler func(checked bool)) {
	if r == nil {
		return
	}
	r.onChange = handler
}

// Measure returns the radio size.
func (r *Radio) Measure(ctx *Context, constraints Constraints) math.Vec2 {
	if r == nil {
		return math.Vec2{}
	}
	tk := r.Tokens(ctx)
	comp := tk.Components.Radio
	f := r.effectiveFont(ctx)

	boxSize := comp.Size
	gap := comp.Gap

	textW := float32(0)
	textH := boxSize
	if f != nil && r.text != "" {
		textW = f.TextWidth(r.text)
		textH = f.LineHeight()
		if textH < boxSize {
			textH = boxSize
		}
	}

	width := boxSize + gap + textW
	height := textH
	return ClampSize(math.NewVec2(width, height), constraints)
}

// Render draws the radio button.
func (r *Radio) Render(ctx *Context) {
	if r == nil || ctx == nil || ctx.Renderer == nil || !r.Visible() {
		return
	}

	tk := r.Tokens(ctx)
	comp := tk.Components.Radio
	global := tk.Global
	bounds := r.Bounds()

	boxSize := comp.Size
	gap := comp.Gap
	borderWidth := comp.BorderWidth

	// Radio circle position
	cx := bounds.X + boxSize/2
	cy := bounds.Y + bounds.H/2
	radius := boxSize / 2

	// Determine colors
	bgColor := global.ColorBgContainer
	borderColor := global.ColorBorder
	dotColor := global.ColorPrimary

	if r.HasState(StateDisabled) {
		bgColor = global.ColorBgContainer
		borderColor = global.ColorBorder
		dotColor = global.ColorTextDisabled
	} else if r.HasState(StateHover) {
		borderColor = dotColor
	}

	// Draw circle
	circle := math.NewRect(cx-radius, cy-radius, boxSize, boxSize)
	ctx.Renderer.FillRoundRect(circle, radius, bgColor)
	ctx.Renderer.StrokeRoundRect(circle, radius, borderWidth, borderColor)

	// Draw dot if checked
	if r.checked {
		dotRadius := radius * 0.4
		dot := math.NewRect(cx-dotRadius, cy-dotRadius, dotRadius*2, dotRadius*2)
		ctx.Renderer.FillRoundRect(dot, dotRadius, dotColor)
	}

	// Draw text
	f := r.effectiveFont(ctx)
	if f != nil && r.text != "" {
		textColor := global.ColorText
		if r.HasState(StateDisabled) {
			textColor = global.ColorTextDisabled
		}
		textX := bounds.X + boxSize + gap
		textY := bounds.Y + (bounds.H-f.LineHeight())/2
		ctx.Renderer.DrawText(r.text, textX, textY, f, textColor)
	}
}

// HandleEvent handles radio interaction.
func (r *Radio) HandleEvent(ctx *Context, event Event) bool {
	if r == nil || !r.Enabled() {
		return false
	}
	if r.interaction == nil {
		r.interaction = NewInteractionController(r)
	}
	r.interaction.SetOnClick(func(Event) {
		if !r.checked {
			r.checked = true
			r.SetStateFlag(StateChecked, true)
			if r.onChange != nil {
				r.onChange(r.checked)
			}
		}
	})
	return r.interaction.HandleEvent(ctx, event)
}

func (r *Radio) effectiveFont(ctx *Context) *font.Font {
	if r != nil && r.font != nil {
		return r.font
	}
	if ctx != nil {
		return ctx.Font
	}
	return nil
}
