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
