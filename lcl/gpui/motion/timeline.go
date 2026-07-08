package motion

import "time"

// Animatable is implemented by widgets that have active animations.
type Animatable interface {
	// Timeline returns the widget's animation timeline.
	Timeline() *Timeline
}

// Timeline owns a set of named transitions.
type Timeline struct {
	transitions map[string]*Transition
}

// NewTimeline creates an empty timeline.
func NewTimeline() *Timeline {
	return &Timeline{transitions: make(map[string]*Transition)}
}

// Add adds or replaces a transition.
func (t *Timeline) Add(name string, transition *Transition) {
	if t == nil {
		return
	}
	t.transitions[name] = transition
}

// Get returns a transition by name.
func (t *Timeline) Get(name string) *Transition {
	if t == nil {
		return nil
	}
	return t.transitions[name]
}

// SetTarget sets a transition target.
func (t *Timeline) SetTarget(name string, target float32) bool {
	if t == nil {
		return false
	}
	transition := t.transitions[name]
	if transition == nil {
		return false
	}
	transition.SetTarget(target)
	return true
}

// Update advances all transitions.
func (t *Timeline) Update(dt time.Duration) {
	if t == nil {
		return
	}
	for _, transition := range t.transitions {
		if transition == nil {
			continue
		}
		transition.Update(dt)
	}
}

// Running reports whether any transition is running.
func (t *Timeline) Running() bool {
	if t == nil {
		return false
	}
	for _, transition := range t.transitions {
		if transition != nil && transition.Running() {
			return true
		}
	}
	return false
}
