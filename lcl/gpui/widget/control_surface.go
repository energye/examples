package widget

import (
	"time"

	"github.com/energye/examples/lcl/gpui/core/math"
	"github.com/energye/examples/lcl/gpui/motion"
)

const (
	controlRippleProgress = "control.ripple.progress"
	controlRippleAlpha    = "control.ripple.alpha"
)

// ControlSurface is a reusable interactive visual shell for future components.
type ControlSurface struct {
	ComponentBase
	interaction   *InteractionController
	onClick       func(Event)
	rippleEnabled bool
	rippleOrigin  math.Vec2
}

// NewControlSurface creates an interactive component shell.
func NewControlSurface() *ControlSurface {
	c := &ControlSurface{
		ComponentBase: NewComponentBase(),
		rippleEnabled: true,
	}
	c.SetOwner(c)
	c.interaction = NewInteractionController(c)
	return c
}

// SetOnClick sets the activation callback.
func (c *ControlSurface) SetOnClick(handler func(Event)) {
	if c == nil {
		return
	}
	c.onClick = handler
}

// SetRippleEnabled toggles the shared Ant Design-style press wave.
func (c *ControlSurface) SetRippleEnabled(enabled bool) {
	if c == nil {
		return
	}
	c.rippleEnabled = enabled
	if !enabled {
		c.resetRipple()
	}
}

// RippleEnabled reports whether the shared press wave is enabled.
func (c *ControlSurface) RippleEnabled() bool {
	return c != nil && c.rippleEnabled
}

// Measure returns the standard token-derived control size.
func (c *ControlSurface) Measure(ctx *Context, constraints Constraints) math.Vec2 {
	if c == nil {
		return math.Vec2{}
	}
	style := c.ResolveControlStyle(ctx)
	size := c.PreferredSize()
	if size.X <= 0 {
		size.X = c.Bounds().W
	}
	if size.X <= 0 {
		size.X = style.Metrics.MinTouchSize.X
	}
	if size.Y <= 0 {
		size.Y = c.Bounds().H
	}
	if size.Y <= 0 {
		size.Y = style.Metrics.Height
	}
	return ClampSize(size, constraints)
}

// Render draws the token-resolved control surface.
func (c *ControlSurface) Render(ctx *Context) {
	if c == nil || ctx == nil || ctx.Renderer == nil || !c.Visible() {
		return
	}
	ctx.Renderer.DrawBox(c.Bounds(), c.ResolveControlStyle(ctx).BoxStyle())
	c.RenderMotionOverlay(ctx, c.Bounds())
}

// HandleEvent applies common interaction behavior.
func (c *ControlSurface) HandleEvent(ctx *Context, event Event) bool {
	if c == nil || !c.Enabled() {
		return false
	}
	if c.interaction == nil {
		c.interaction = NewInteractionController(c)
	}
	if isActivationStartEvent(event) {
		c.startRipple(event)
	}
	c.interaction.SetOnClick(c.onClick)
	return c.interaction.HandleEvent(ctx, event)
}

// RenderMotionOverlay draws shared control motion effects over a control.
func (c *ControlSurface) RenderMotionOverlay(ctx *Context, bounds math.Rect) {
	if c == nil || ctx == nil || ctx.Renderer == nil || !c.rippleEnabled {
		return
	}
	progress := c.MotionValue(controlRippleProgress, 1)
	alpha := c.MotionValue(controlRippleAlpha, 0)
	if alpha <= 0 || progress <= 0 || bounds.W <= 0 || bounds.H <= 0 {
		return
	}
	maxRadius := bounds.W
	if bounds.H > maxRadius {
		maxRadius = bounds.H
	}
	maxRadius *= 0.75
	color := c.Tokens(ctx).Global.ColorPrimary.WithAlpha(alpha)
	ctx.Renderer.PushClip(bounds)
	ctx.Renderer.FillCircle(c.rippleOrigin, maxRadius*progress, color)
	ctx.Renderer.PopClip()
}

func (c *ControlSurface) ensureRippleTimeline() {
	if c == nil {
		return
	}
	timeline := c.EnsureTimeline()
	if timeline.Get(controlRippleProgress) == nil {
		c.AddTransition(controlRippleProgress, 1, 260*time.Millisecond, motion.EaseOut)
	}
	if timeline.Get(controlRippleAlpha) == nil {
		c.AddTransition(controlRippleAlpha, 0, 260*time.Millisecond, motion.EaseOut)
	}
}

func (c *ControlSurface) startRipple(event Event) {
	if c == nil || !c.rippleEnabled || !c.Enabled() || c.HasState(StateDisabled) || c.HasState(StateLoading) {
		return
	}
	bounds := c.Bounds()
	if event.Type == EventKeyDown {
		c.rippleOrigin = bounds.Center()
	} else {
		c.rippleOrigin = math.NewVec2(bounds.X+event.LocalX, bounds.Y+event.LocalY)
	}
	c.ensureRippleTimeline()
	progress := c.EnsureTimeline().Get(controlRippleProgress)
	alpha := c.EnsureTimeline().Get(controlRippleAlpha)
	progress.Reset(0)
	alpha.Reset(0.18)
	progress.SetTarget(1)
	alpha.SetTarget(0)
	c.Invalidate()
}

func (c *ControlSurface) resetRipple() {
	if c == nil || c.timeline == nil {
		return
	}
	if progress := c.timeline.Get(controlRippleProgress); progress != nil {
		progress.Reset(1)
	}
	if alpha := c.timeline.Get(controlRippleAlpha); alpha != nil {
		alpha.Reset(0)
	}
	c.Invalidate()
}

func isActivationStartEvent(event Event) bool {
	if event.Type == EventMouseDown {
		return true
	}
	if event.Type == EventKeyDown {
		return event.Key == keyEnter || event.Key == keySpace
	}
	return false
}
