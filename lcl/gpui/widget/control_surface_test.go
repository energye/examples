package widget

import (
	"testing"

	"github.com/energye/examples/lcl/gpui/core/math"
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
