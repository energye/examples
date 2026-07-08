package widget

import (
	"testing"

	"github.com/energye/examples/lcl/gpui/core/math"
	"github.com/energye/examples/lcl/gpui/render/pipeline"
)

func TestInteractionControllerPointerLifecycle(t *testing.T) {
	target := newRecordingWidget(math.NewRect(0, 0, 40, 20))
	controller := NewInteractionController(target)
	clicks := 0
	controller.SetOnClick(func(Event) {
		clicks++
	})

	if handled := controller.HandleEvent(nil, Event{Type: EventMouseMove}); handled {
		t.Fatal("mouse move should update hover without consuming event")
	}
	if !target.HasState(StateHover) {
		t.Fatal("mouse move should set hover state")
	}

	if !controller.HandleEvent(nil, Event{Type: EventMouseDown, Button: 1}) {
		t.Fatal("mouse down should be handled")
	}
	if !controller.Pressed() || !target.HasState(StateActive) {
		t.Fatal("mouse down should set pressed and active state")
	}

	if !controller.HandleEvent(nil, Event{Type: EventMouseUp, Button: 1}) {
		t.Fatal("mouse up after press should be handled")
	}
	if controller.Pressed() || target.HasState(StateActive) {
		t.Fatal("mouse up should clear pressed and active state")
	}
	if clicks != 1 {
		t.Fatalf("clicks = %d, want 1", clicks)
	}
}

func TestInteractionControllerMouseUpWithoutPressDoesNotActivate(t *testing.T) {
	target := newRecordingWidget(math.NewRect(0, 0, 40, 20))
	controller := NewInteractionController(target)
	clicks := 0
	controller.SetOnClick(func(Event) {
		clicks++
	})

	if controller.HandleEvent(nil, Event{Type: EventMouseUp, Button: 1}) {
		t.Fatal("mouse up without prior press should not be handled")
	}
	if clicks != 0 {
		t.Fatalf("clicks = %d, want 0", clicks)
	}
}

func TestInteractionControllerMouseUpOutsideDoesNotActivate(t *testing.T) {
	target := newRecordingWidget(math.NewRect(0, 0, 40, 20))
	controller := NewInteractionController(target)
	clicks := 0
	controller.SetOnClick(func(Event) {
		clicks++
	})

	controller.HandleEvent(nil, Event{Type: EventMouseDown, LocalX: 10, LocalY: 10, Button: 1})
	if !target.HasState(StateActive) {
		t.Fatal("target should become active after mouse down")
	}
	controller.HandleEvent(nil, Event{Type: EventMouseMove, LocalX: 80, LocalY: 10})
	if !target.HasState(StateActive) {
		t.Fatal("target should stay active while pointer capture is pressed")
	}
	controller.HandleEvent(nil, Event{Type: EventMouseUp, LocalX: 80, LocalY: 10, Button: 1})
	if target.HasState(StateActive) {
		t.Fatal("target should clear active when captured mouse up is released")
	}
	if clicks != 0 {
		t.Fatalf("clicks = %d, want 0", clicks)
	}
}

func TestInteractionControllerKeyboardActivationRequiresFocus(t *testing.T) {
	target := newRecordingWidget(math.NewRect(0, 0, 40, 20))
	target.SetFocusable(true)
	controller := NewInteractionController(target)
	clicks := 0
	controller.SetOnClick(func(Event) {
		clicks++
	})

	if controller.HandleEvent(nil, Event{Type: EventKeyDown, Key: keyEnter}) {
		t.Fatal("keyboard activation should require focus by default")
	}
	target.Focus()
	if !controller.HandleEvent(nil, Event{Type: EventKeyDown, Key: keySpace}) {
		t.Fatal("space should activate focused target")
	}
	if target.HasState(StateActive) {
		t.Fatal("keyboard activation should clear active state after activation")
	}
	if clicks != 1 {
		t.Fatalf("clicks = %d, want 1", clicks)
	}
}

func TestInteractionControllerDisabledAndLoading(t *testing.T) {
	target := newRecordingWidget(math.NewRect(0, 0, 40, 20))
	controller := NewInteractionController(target)
	clicks := 0
	controller.SetOnClick(func(Event) {
		clicks++
	})

	target.SetEnabled(false)
	if controller.HandleEvent(nil, Event{Type: EventMouseDown, Button: 1}) {
		t.Fatal("disabled target should not handle pointer activation")
	}
	if target.HasState(StateActive) {
		t.Fatal("disabled target should not become active")
	}

	target.SetEnabled(true)
	target.SetStateFlag(StateLoading, true)
	if !controller.HandleEvent(nil, Event{Type: EventMouseDown, Button: 1}) {
		t.Fatal("loading target should consume activation event")
	}
	if controller.Pressed() || target.HasState(StateActive) {
		t.Fatal("loading target should not stay pressed or active")
	}
	if clicks != 0 {
		t.Fatalf("clicks = %d, want 0", clicks)
	}
}

func TestInteractionControllerSetTargetClearsOldState(t *testing.T) {
	first := newRecordingWidget(math.NewRect(0, 0, 40, 20))
	second := newRecordingWidget(math.NewRect(0, 0, 40, 20))
	controller := NewInteractionController(first)

	controller.HandleEvent(nil, Event{Type: EventMouseMove})
	controller.HandleEvent(nil, Event{Type: EventMouseDown, Button: 1})
	if !first.HasState(StateHover) || !first.HasState(StateActive) {
		t.Fatal("first target should have transient interaction state")
	}

	controller.SetTarget(second)
	if first.HasState(StateHover) || first.HasState(StateActive) {
		t.Fatal("old target should have transient interaction state cleared")
	}
	if second.HasState(StateHover) || second.HasState(StateActive) {
		t.Fatal("new target should start without transient interaction state")
	}
}

func TestBoxClickRequiresPress(t *testing.T) {
	box := NewBox(pipeline.BoxStyle{})
	box.SetBounds(math.NewRect(0, 0, 40, 20))
	clicks := 0
	box.OnClick = func() {
		clicks++
	}

	if box.HandleEvent(nil, Event{Type: EventMouseUp, Button: 1}) && clicks != 0 {
		t.Fatal("box should not click from mouse up without prior press")
	}
	box.HandleEvent(nil, Event{Type: EventMouseDown, Button: 1})
	box.HandleEvent(nil, Event{Type: EventMouseUp, Button: 1})
	if clicks != 1 {
		t.Fatalf("clicks = %d, want 1", clicks)
	}
}
