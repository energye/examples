package motion

import "time"

// State identifies transition state.
type State int

const (
	Idle State = iota
	Running
	Done
)

// Transition interpolates a float value over explicit time steps.
type Transition struct {
	from     float32
	to       float32
	value    float32
	elapsed  time.Duration
	duration time.Duration
	easing   Easing
	state    State
}

// NewTransition creates a transition.
func NewTransition(value float32, duration time.Duration, easing Easing) *Transition {
	if easing == nil {
		easing = Linear
	}
	return &Transition{
		from:     value,
		to:       value,
		value:    value,
		duration: duration,
		easing:   easing,
		state:    Idle,
	}
}

// SetTarget starts a transition from the current value to target.
func (t *Transition) SetTarget(target float32) {
	if t.to == target && t.state == Running {
		return
	}
	t.from = t.value
	t.to = target
	t.elapsed = 0
	if t.duration <= 0 {
		t.value = target
		t.state = Done
		return
	}
	t.state = Running
}

// Update advances the transition.
func (t *Transition) Update(dt time.Duration) {
	if t.state != Running {
		return
	}
	t.elapsed += dt
	progress := float32(t.elapsed) / float32(t.duration)
	if progress >= 1 {
		progress = 1
		t.state = Done
	}
	eased := t.easing(progress)
	t.value = t.from + (t.to-t.from)*eased
}

// Value returns the current value.
func (t *Transition) Value() float32 {
	return t.value
}

// Target returns the target value.
func (t *Transition) Target() float32 {
	return t.to
}

// State returns the current transition state.
func (t *Transition) State() State {
	return t.state
}

// Running reports whether the transition is running.
func (t *Transition) Running() bool {
	return t.state == Running
}
