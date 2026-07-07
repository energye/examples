package motion

import (
	"time"

	"github.com/energye/examples/lcl/gpui/core/math"
)

// ColorTransition interpolates colors over time.
type ColorTransition struct {
	r *Transition
	g *Transition
	b *Transition
	a *Transition
}

// NewColorTransition creates a color transition.
func NewColorTransition(value math.Color, duration time.Duration, easing Easing) *ColorTransition {
	return &ColorTransition{
		r: NewTransition(value.R, duration, easing),
		g: NewTransition(value.G, duration, easing),
		b: NewTransition(value.B, duration, easing),
		a: NewTransition(value.A, duration, easing),
	}
}

// SetTarget starts transitioning toward a color.
func (c *ColorTransition) SetTarget(target math.Color) {
	c.r.SetTarget(target.R)
	c.g.SetTarget(target.G)
	c.b.SetTarget(target.B)
	c.a.SetTarget(target.A)
}

// Update advances the transition.
func (c *ColorTransition) Update(dt time.Duration) {
	c.r.Update(dt)
	c.g.Update(dt)
	c.b.Update(dt)
	c.a.Update(dt)
}

// Value returns the current color.
func (c *ColorTransition) Value() math.Color {
	return math.NewColor(c.r.Value(), c.g.Value(), c.b.Value(), c.a.Value())
}

// Running reports whether any channel is running.
func (c *ColorTransition) Running() bool {
	return c.r.Running() || c.g.Running() || c.b.Running() || c.a.Running()
}
