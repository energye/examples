package widget

import (
	"time"

	"github.com/energye/examples/lcl/gpui/core/math"
	"github.com/energye/examples/lcl/gpui/motion"
)

const (
	controlRippleProgress = "control.ripple.progress"
	controlRippleAlpha    = "control.ripple.alpha"
	controlLoadingSpin    = "control.loading.spin"
	controlColorPrefix    = "control.color."
)

// FocusRing describes a resolved focus indicator.
type FocusRing struct {
	Rect   math.Rect
	Radius float32
	Width  float32
	Color  math.Color
}

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

// SetLoadingMotion toggles the shared loading spinner motion.
func (c *ControlSurface) SetLoadingMotion(loading bool) {
	if c == nil {
		return
	}
	if loading {
		c.ensureLoadingTimeline()
		spin := c.EnsureTimeline().Get(controlLoadingSpin)
		if spin != nil && !spin.Running() {
			spin.Reset(0)
			spin.SetTarget(1)
		}
	} else if c.timeline != nil {
		if spin := c.timeline.Get(controlLoadingSpin); spin != nil {
			spin.Reset(0)
		}
	}
	c.Invalidate()
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
	style := c.ResolveAnimatedControlStyle(ctx, c.ResolveControlStyle(ctx))
	ctx.Renderer.DrawBox(c.Bounds(), style.BoxStyle())
	c.RenderMotionOverlay(ctx, c.Bounds())
	c.RenderFocusRing(ctx, c.Bounds(), style.Metrics.Radius)
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

// ResolveAnimatedControlStyle applies shared state color transitions to a style.
func (c *ControlSurface) ResolveAnimatedControlStyle(ctx *Context, style ControlStyle) ControlStyle {
	if c == nil {
		return style
	}
	style.Palette.Text = c.AnimatedColor(ctx, "text", style.Palette.Text)
	style.Palette.Background = c.AnimatedColor(ctx, "background", style.Palette.Background)
	style.Palette.Border = c.AnimatedColor(ctx, "border", style.Palette.Border)
	style.Palette.Placeholder = c.AnimatedColor(ctx, "placeholder", style.Palette.Placeholder)
	return style
}

// AnimatedColor transitions a named color toward target using token motion timing.
func (c *ControlSurface) AnimatedColor(ctx *Context, name string, target math.Color) math.Color {
	if c == nil || name == "" {
		return target
	}
	base := controlColorPrefix + name
	duration := c.Tokens(ctx).Global.MotionDurationFast
	if duration <= 0 {
		duration = 100 * time.Millisecond
	}
	r := c.ensureColorTransition(base+".r", target.R, duration)
	g := c.ensureColorTransition(base+".g", target.G, duration)
	b := c.ensureColorTransition(base+".b", target.B, duration)
	a := c.ensureColorTransition(base+".a", target.A, duration)
	return math.NewColor(r.Value(), g.Value(), b.Value(), a.Value())
}

func (c *ControlSurface) ensureColorTransition(name string, target float32, duration time.Duration) *motion.Transition {
	timeline := c.EnsureTimeline()
	transition := timeline.Get(name)
	if transition == nil {
		transition = c.AddTransition(name, target, duration, motion.EaseOut)
		return transition
	}
	if transition.Target() != target {
		transition.SetTarget(target)
		c.Invalidate()
	}
	return transition
}

// ResolveFocusRing returns the shared Ant Design-style focus ring.
func (c *ControlSurface) ResolveFocusRing(ctx *Context, bounds math.Rect, radius float32) (FocusRing, bool) {
	if c == nil || !c.Focused() || !c.Enabled() || !c.Visible() || bounds.W <= 0 || bounds.H <= 0 {
		return FocusRing{}, false
	}
	tokens := c.Tokens(ctx)
	color := tokens.Global.ColorPrimary.WithAlpha(0.55)
	return FocusRing{
		Rect:   bounds.Expand(2),
		Radius: radius + 2,
		Width:  2,
		Color:  color,
	}, true
}

// RenderFocusRing draws the shared focus indicator.
func (c *ControlSurface) RenderFocusRing(ctx *Context, bounds math.Rect, radius float32) {
	if c == nil || ctx == nil || ctx.Renderer == nil {
		return
	}
	ring, ok := c.ResolveFocusRing(ctx, bounds, radius)
	if !ok {
		return
	}
	ctx.Renderer.StrokeRoundRect(ring.Rect, ring.Radius, ring.Width, ring.Color)
}

// RenderLoadingSpinner draws a shared loading spinner.
func (c *ControlSurface) RenderLoadingSpinner(ctx *Context, center math.Vec2, radius float32, color math.Color) {
	if c == nil || ctx == nil || ctx.Renderer == nil || radius <= 0 || color.A <= 0 {
		return
	}
	c.ensureLoadingTimeline()
	spin := c.MotionValue(controlLoadingSpin, 0)
	start := spin * 360
	ctx.Renderer.DrawArc(center, radius, 2, start, start+270, color)
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

func (c *ControlSurface) ensureLoadingTimeline() {
	if c == nil {
		return
	}
	timeline := c.EnsureTimeline()
	if timeline.Get(controlLoadingSpin) != nil {
		return
	}
	spin := c.AddTransition(controlLoadingSpin, 0, 800*time.Millisecond, motion.Linear)
	spin.SetLoop(true)
	spin.SetTarget(1)
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
	if event.Type == EventDoubleClick {
		return true
	}
	if event.Type == EventKeyDown {
		return event.Key == keyEnter || event.Key == keySpace
	}
	return false
}
