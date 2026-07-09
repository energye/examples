package widget

import (
	"time"

	"github.com/energye/examples/lcl/gpui/core/math"
	"github.com/energye/examples/lcl/gpui/motion"
)

const switchThumbPosition = "switch.thumb.position"

// Switch is a toggle switch control.
type Switch struct {
	ControlSurface
	checked  bool
	loading  bool
	onChange func(checked bool)
}

// NewSwitch creates a new switch.
func NewSwitch() *Switch {
	s := &Switch{
		ControlSurface: *NewControlSurface(),
	}
	s.SetOwner(s)
	s.interaction.SetTarget(s)
	s.SetFocusable(true)
	return s
}

// Checked returns whether the switch is on.
func (s *Switch) Checked() bool {
	return s != nil && s.checked
}

// SetChecked sets the switch state.
func (s *Switch) SetChecked(checked bool) {
	if s == nil || s.checked == checked {
		return
	}
	s.checked = checked
	s.SetStateFlag(StateChecked, checked)
	s.setThumbTarget(checked)
	s.Invalidate()
}

// Loading returns whether the switch is loading.
func (s *Switch) Loading() bool {
	return s != nil && s.loading
}

// SetLoading sets the loading state.
func (s *Switch) SetLoading(loading bool) {
	if s == nil {
		return
	}
	s.loading = loading
	s.SetStateFlag(StateLoading, loading)
}

// SetOnChange sets the change callback.
func (s *Switch) SetOnChange(handler func(checked bool)) {
	if s == nil {
		return
	}
	s.onChange = handler
}

// Measure returns the switch size.
func (s *Switch) Measure(ctx *Context, constraints Constraints) math.Vec2 {
	if s == nil {
		return math.Vec2{}
	}
	tk := s.Tokens(ctx)
	comp := tk.Components.Switch
	return ClampSize(math.NewVec2(comp.MinWidth, comp.Height), constraints)
}

// Render draws the switch.
func (s *Switch) Render(ctx *Context) {
	if s == nil || ctx == nil || ctx.Renderer == nil || !s.Visible() {
		return
	}

	tk := s.Tokens(ctx)
	comp := tk.Components.Switch
	global := tk.Global
	bounds := s.Bounds()

	height := comp.Height
	width := comp.MinWidth
	innerSize := comp.InnerSize
	innerMargin := comp.InnerMargin
	radius := comp.Radius

	// Track colors
	trackColor := global.ColorBorder
	if s.checked {
		trackColor = global.ColorPrimary
	}
	if s.HasState(StateDisabled) {
		trackColor = global.ColorTextDisabled
	}
	if s.HasState(StateHover) && !s.HasState(StateDisabled) {
		if s.checked {
			trackColor = global.ColorPrimaryHover
		} else {
			trackColor = global.ColorBorderHover
		}
	}

	// Draw track
	track := math.NewRect(bounds.X, bounds.Y+(bounds.H-height)/2, width, height)
	ctx.Renderer.FillRoundRect(track, radius, trackColor)
	s.RenderMotionOverlay(ctx, track)

	// Draw thumb
	thumbY := bounds.Y + (bounds.H-innerSize)/2
	thumbProgress := s.MotionValue(switchThumbPosition, boolProgress(s.checked))
	thumbMinX := bounds.X + innerMargin
	thumbMaxX := bounds.X + width - innerSize - innerMargin
	thumbX := thumbMinX + (thumbMaxX-thumbMinX)*thumbProgress
	thumb := math.NewRect(thumbX, thumbY, innerSize, innerSize)
	ctx.Renderer.FillRoundRect(thumb, innerSize/2, math.NewColor(1, 1, 1, 1))
}

// HandleEvent handles switch interaction.
func (s *Switch) HandleEvent(ctx *Context, event Event) bool {
	if s == nil || !s.Enabled() || s.loading {
		return false
	}
	if s.interaction == nil {
		s.interaction = NewInteractionController(s)
	}
	s.interaction.SetOnClick(func(Event) {
		s.SetChecked(!s.checked)
		if s.onChange != nil {
			s.onChange(s.checked)
		}
	})
	return s.interaction.HandleEvent(ctx, event)
}

func (s *Switch) setThumbTarget(checked bool) {
	if s == nil {
		return
	}
	if s.EnsureTimeline().Get(switchThumbPosition) == nil {
		s.AddTransition(switchThumbPosition, boolProgress(!checked), 180*time.Millisecond, motion.EaseOut)
	}
	s.SetMotionTarget(switchThumbPosition, boolProgress(checked))
}

func boolProgress(value bool) float32 {
	if value {
		return 1
	}
	return 0
}
