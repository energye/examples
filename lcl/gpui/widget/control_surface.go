package widget

import "github.com/energye/examples/lcl/gpui/core/math"

// ControlSurface is a reusable interactive visual shell for future components.
type ControlSurface struct {
	ComponentBase
	Interaction *InteractionController
	OnClick     func(Event)
}

// NewControlSurface creates an interactive component shell.
func NewControlSurface() *ControlSurface {
	c := &ControlSurface{ComponentBase: NewComponentBase()}
	c.SetOwner(c)
	c.Interaction = NewInteractionController(c)
	return c
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
}

// HandleEvent applies common interaction behavior.
func (c *ControlSurface) HandleEvent(ctx *Context, event Event) bool {
	if c == nil || !c.Enabled() {
		return false
	}
	if c.Interaction == nil {
		c.Interaction = NewInteractionController(c)
	}
	c.Interaction.SetOnClick(c.OnClick)
	return c.Interaction.HandleEvent(ctx, event)
}
