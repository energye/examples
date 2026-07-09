package widget

import (
	"testing"
	"time"

	"github.com/energye/examples/lcl/gpui/core/math"
	"github.com/energye/examples/lcl/gpui/motion"
	"github.com/energye/examples/lcl/gpui/style/token"
)

func TestControlSurfaceMeasuresFromTokens(t *testing.T) {
	tokens := token.DefaultLight()
	control := NewControlSurface()
	control.SetControlSize(SizeLarge)

	size := control.Measure(&Context{Tokens: tokens}, Constraints{})
	if size.X != tokens.Alias.ControlHeightLG {
		t.Fatalf("width = %v, want min touch width %v", size.X, tokens.Alias.ControlHeightLG)
	}
	if size.Y != tokens.Alias.ControlHeightLG {
		t.Fatalf("height = %v, want %v", size.Y, tokens.Alias.ControlHeightLG)
	}
}

func TestControlSurfacePreferredSizeAndConstraints(t *testing.T) {
	control := NewControlSurface()
	control.SetPreferredSize(math.NewVec2(160, 48))

	size := control.Measure(nil, Constraints{Max: math.NewVec2(120, 40)})
	if size.X != 120 || size.Y != 40 {
		t.Fatalf("size = (%v,%v), want constrained 120x40", size.X, size.Y)
	}
}

func TestControlSurfacePointerClick(t *testing.T) {
	control := NewControlSurface()
	clicks := 0
	control.SetOnClick(func(Event) {
		clicks++
	})

	if !control.HandleEvent(nil, Event{Type: EventMouseDown, Button: 1}) {
		t.Fatal("mouse down should be handled")
	}
	if !control.HasState(StateActive) {
		t.Fatal("mouse down should set active state")
	}
	if !control.HandleEvent(nil, Event{Type: EventMouseUp, Button: 1}) {
		t.Fatal("mouse up should be handled")
	}
	if control.HasState(StateActive) {
		t.Fatal("mouse up should clear active state")
	}
	if clicks != 1 {
		t.Fatalf("clicks = %d, want 1", clicks)
	}
}

func TestControlSurfacePointerStartsRippleTimeline(t *testing.T) {
	control := NewControlSurface()
	control.SetBounds(math.NewRect(10, 20, 80, 32))

	control.HandleEvent(nil, Event{Type: EventMouseDown, LocalX: 12, LocalY: 8, Button: 1})

	timeline := control.Timeline()
	if timeline == nil {
		t.Fatal("mouse down should create control motion timeline")
	}
	if !timeline.Running() {
		t.Fatal("ripple transitions should be running after press")
	}
	if got := timeline.Get(controlRippleProgress).Value(); got != 0 {
		t.Fatalf("ripple progress = %v, want 0 at start", got)
	}
	if got := timeline.Get(controlRippleAlpha).Value(); got <= 0 {
		t.Fatalf("ripple alpha = %v, want positive at start", got)
	}
}

func TestControlSurfaceDoubleClickRestartsRippleTimeline(t *testing.T) {
	control := NewControlSurface()
	control.SetBounds(math.NewRect(10, 20, 80, 32))

	control.HandleEvent(nil, Event{Type: EventMouseDown, LocalX: 12, LocalY: 8, Button: 1})
	control.Timeline().Update(100 * time.Millisecond)
	if got := control.Timeline().Get(controlRippleProgress).Value(); got <= 0 {
		t.Fatalf("ripple progress should advance before double click, got %v", got)
	}

	control.HandleEvent(nil, Event{Type: EventDoubleClick, LocalX: 14, LocalY: 9, Button: 1})
	if got := control.Timeline().Get(controlRippleProgress).Value(); got != 0 {
		t.Fatalf("double click should restart ripple progress, got %v", got)
	}
	if got := control.Timeline().Get(controlRippleAlpha).Value(); got <= 0 {
		t.Fatalf("double click should restart ripple alpha, got %v", got)
	}
}

func TestBaseWidgetMotionTimelineIsOptIn(t *testing.T) {
	w := newRecordingWidget(math.NewRect(0, 0, 10, 10))
	if w.Timeline() != nil {
		t.Fatal("new widget should not allocate a timeline until motion is registered")
	}
	w.AddTransition("opacity", 0, time.Second, motion.Linear)
	w.SetMotionTarget("opacity", 1)
	if w.Timeline() == nil || !w.MotionRunning() {
		t.Fatal("registered transition should create a running timeline")
	}
}

func TestControlSurfaceLoadingMotionLoops(t *testing.T) {
	control := NewControlSurface()
	control.SetLoadingMotion(true)

	spin := control.Timeline().Get(controlLoadingSpin)
	if spin == nil {
		t.Fatal("loading motion should register spinner transition")
	}
	if !spin.Running() || !spin.Loop() {
		t.Fatal("spinner transition should run as a loop")
	}

	control.Timeline().Update(1200 * time.Millisecond)
	if !spin.Running() {
		t.Fatal("spinner transition should still be running after wrapping")
	}

	control.SetLoadingMotion(false)
	if spin.Running() || spin.Value() != 0 {
		t.Fatalf("loading motion should reset when disabled, running=%v value=%v", spin.Running(), spin.Value())
	}
}

func TestControlSurfaceFocusRingUsesTokenFallback(t *testing.T) {
	control := NewControlSurface()
	control.SetFocusable(true)
	control.Focus()

	ring, ok := control.ResolveFocusRing(nil, math.NewRect(10, 12, 80, 32), 6)
	if !ok {
		t.Fatal("focused control should resolve a focus ring")
	}
	if ring.Rect.X != 8 || ring.Rect.Y != 10 || ring.Rect.W != 84 || ring.Rect.H != 36 {
		t.Fatalf("focus ring rect = %#v, want expanded bounds", ring.Rect)
	}
	if ring.Radius != 8 || ring.Width != 2 {
		t.Fatalf("focus ring radius/width = %v/%v, want 8/2", ring.Radius, ring.Width)
	}
	if ring.Color.A <= 0 {
		t.Fatal("focus ring should use a visible token-derived color")
	}
}

func TestControlSurfaceAnimatedColorTransitionsToTarget(t *testing.T) {
	control := NewControlSurface()
	first := math.NewColor(0, 0, 0, 1)
	second := math.NewColor(1, 0, 0, 1)

	if got := control.AnimatedColor(nil, "border", first); got != first {
		t.Fatalf("initial animated color = %#v, want first target", got)
	}
	if got := control.AnimatedColor(nil, "border", second); got != first {
		t.Fatalf("retargeted color should start from current value, got %#v want %#v", got, first)
	}
	control.Timeline().Update(50 * time.Millisecond)
	mid := control.AnimatedColor(nil, "border", second)
	if mid.R <= 0 || mid.R >= 1 {
		t.Fatalf("mid transition red channel = %v, want between 0 and 1", mid.R)
	}
	control.Timeline().Update(100 * time.Millisecond)
	if got := control.AnimatedColor(nil, "border", second); got != second {
		t.Fatalf("completed animated color = %#v, want %#v", got, second)
	}
}

func TestControlSurfaceDisabledFocusRingHidden(t *testing.T) {
	control := NewControlSurface()
	control.SetFocusable(true)
	control.Focus()
	control.SetEnabled(false)

	if _, ok := control.ResolveFocusRing(nil, math.NewRect(0, 0, 80, 32), 6); ok {
		t.Fatal("disabled control should not resolve a focus ring")
	}
}

func TestControlSurfaceKeyboardClick(t *testing.T) {
	control := NewControlSurface()
	control.SetFocusable(true)
	control.Focus()
	clicks := 0
	control.SetOnClick(func(Event) {
		clicks++
	})

	if !control.HandleEvent(nil, Event{Type: EventKeyDown, Key: keyEnter}) {
		t.Fatal("focused enter should activate control")
	}
	if clicks != 1 {
		t.Fatalf("clicks = %d, want 1", clicks)
	}
}

func TestControlSurfaceDisabledDoesNotActivate(t *testing.T) {
	control := NewControlSurface()
	control.SetEnabled(false)
	clicks := 0
	control.SetOnClick(func(Event) {
		clicks++
	})

	if control.HandleEvent(nil, Event{Type: EventMouseDown, Button: 1}) {
		t.Fatal("disabled control should not handle mouse down")
	}
	if clicks != 0 {
		t.Fatalf("clicks = %d, want 0", clicks)
	}
}
