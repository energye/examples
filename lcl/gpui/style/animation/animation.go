// Package animation provides animation and easing functions
package animation

import (
	"sync"
	"time"
)

// EasingFunc is a function that maps [0,1] to [0,1]
type EasingFunc func(t float32) float32

// Common easing functions
var (
	Linear = EasingFunc(func(t float32) float32 {
		return t
	})

	EaseInQuad = EasingFunc(func(t float32) float32 {
		return t * t
	})

	EaseOutQuad = EasingFunc(func(t float32) float32 {
		return t * (2 - t)
	})

	EaseInOutQuad = EasingFunc(func(t float32) float32 {
		if t < 0.5 {
			return 2 * t * t
		}
		return -1 + (4-2*t)*t
	})

	EaseInCubic = EasingFunc(func(t float32) float32 {
		return t * t * t
	})

	EaseOutCubic = EasingFunc(func(t float32) float32 {
		t1 := t - 1
		return t1*t1*t1 + 1
	})

	EaseInOutCubic = EasingFunc(func(t float32) float32 {
		if t < 0.5 {
			return 4 * t * t * t
		}
		return (t-1)*(2*t-2)*(2*t-2) + 1
	})

	// Ant Design uses ease-out for most animations
	EaseOut   = EaseOutCubic
	EaseIn    = EaseInCubic
	EaseInOut = EaseInOutCubic
)

// Animation represents an animation
type Animation struct {
	mu sync.Mutex

	from     float32
	to       float32
	duration time.Duration
	easing   EasingFunc

	startTime time.Time
	running   bool
	reverse   bool
	value     float32

	// Loop support
	loop   bool
	paused bool
}

// NewAnimation creates a new animation
func NewAnimation(from, to float32, duration time.Duration, easing EasingFunc) *Animation {
	return &Animation{
		from:     from,
		to:       to,
		duration: duration,
		easing:   easing,
		value:    from,
	}
}

// NewLoopAnimation creates a looping animation
func NewLoopAnimation(from, to float32, duration time.Duration, easing EasingFunc) *Animation {
	return &Animation{
		from:     from,
		to:       to,
		duration: duration,
		easing:   easing,
		value:    from,
		loop:     true,
	}
}

// PlayForward plays the animation forward
func (a *Animation) PlayForward() {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.reverse = false
	a.running = true
	a.paused = false
	a.startTime = time.Now()
}

// PlayReverse plays the animation in reverse
func (a *Animation) PlayReverse() {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.reverse = true
	a.running = true
	a.paused = false
	a.startTime = time.Now()
}

// Stop stops the animation
func (a *Animation) Stop() {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.running = false
	a.paused = false
}

// Pause pauses the animation
func (a *Animation) Pause() {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.paused = true
}

// Resume resumes the animation
func (a *Animation) Resume() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.paused {
		a.paused = false
		a.startTime = time.Now()
	}
}

// Reset resets the animation to initial state
func (a *Animation) Reset() {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.running = false
	a.paused = false
	a.value = a.from
}

// Value returns the current value
func (a *Animation) Value() float32 {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.running || a.paused {
		return a.value
	}

	elapsed := time.Since(a.startTime)
	t := float32(elapsed) / float32(a.duration)

	if t >= 1.0 {
		if a.loop {
			// Loop: restart animation
			a.startTime = time.Now()
			t = 0
		} else {
			// Stop at end
			a.running = false
			t = 1.0
		}
	}

	// Apply easing
	t = a.easing(t)

	// Interpolate
	if a.reverse {
		a.value = a.to + (a.from-a.to)*t
	} else {
		a.value = a.from + (a.to-a.from)*t
	}

	return a.value
}

// IsRunning returns whether the animation is running
func (a *Animation) IsRunning() bool {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.running && !a.paused
}

// IsPaused returns whether the animation is paused
func (a *Animation) IsPaused() bool {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.paused
}

// SetLoop sets whether the animation should loop
func (a *Animation) SetLoop(loop bool) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.loop = loop
}

// Animator manages multiple animations
type Animator struct {
	animations map[string]*Animation
	mu         sync.RWMutex
}

// NewAnimator creates a new animator
func NewAnimator() *Animator {
	return &Animator{
		animations: make(map[string]*Animation),
	}
}

// Create creates a new animation
func (a *Animator) Create(id string, from, to float32, duration time.Duration, easing EasingFunc) *Animation {
	a.mu.Lock()
	defer a.mu.Unlock()

	anim := NewAnimation(from, to, duration, easing)
	a.animations[id] = anim
	return anim
}

// CreateLoop creates a new looping animation
func (a *Animator) CreateLoop(id string, from, to float32, duration time.Duration, easing EasingFunc) *Animation {
	a.mu.Lock()
	defer a.mu.Unlock()

	anim := NewLoopAnimation(from, to, duration, easing)
	a.animations[id] = anim
	return anim
}

// Get returns an animation by ID
func (a *Animator) Get(id string) *Animation {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.animations[id]
}

// Update updates all animations (call this in the render loop)
func (a *Animator) Update() {
	a.mu.RLock()
	defer a.mu.RUnlock()

	for _, anim := range a.animations {
		if anim.running && !anim.paused {
			anim.Value()
		}
	}
}

// Interpolator interpolates between two values
type Interpolator struct {
	from, to, current float32
	speed             float32
}

// NewInterpolator creates a new interpolator
func NewInterpolator(from, to, speed float32) *Interpolator {
	return &Interpolator{
		from:    from,
		to:      to,
		current: from,
		speed:   speed,
	}
}

// SetTarget sets the target value
func (i *Interpolator) SetTarget(target float32) {
	i.from = i.current
	i.to = target
}

// Update updates the interpolation
func (i *Interpolator) Update() {
	i.current += (i.to - i.current) * i.speed
}

// Value returns the current value
func (i *Interpolator) Value() float32 {
	return i.current
}

// ColorInterpolator interpolates between two colors
type ColorInterpolator struct {
	from, to, current [4]float32
	speed             float32
}

// NewColorInterpolator creates a new color interpolator
func NewColorInterpolator(fromR, fromG, fromB, fromA, toR, toG, toB, toA, speed float32) *ColorInterpolator {
	return &ColorInterpolator{
		from:    [4]float32{fromR, fromG, fromB, fromA},
		to:      [4]float32{toR, toG, toB, toA},
		current: [4]float32{fromR, fromG, fromB, fromA},
		speed:   speed,
	}
}

// SetTarget sets the target color
func (ci *ColorInterpolator) SetTarget(r, g, b, a float32) {
	ci.from = ci.current
	ci.to = [4]float32{r, g, b, a}
}

// Update updates the interpolation
func (ci *ColorInterpolator) Update() {
	for i := 0; i < 4; i++ {
		ci.current[i] += (ci.to[i] - ci.current[i]) * ci.speed
	}
}

// Color returns the current color components
func (ci *ColorInterpolator) Color() (r, g, b, a float32) {
	return ci.current[0], ci.current[1], ci.current[2], ci.current[3]
}
