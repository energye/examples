package widget

import (
	"testing"

	"github.com/energye/examples/lcl/gpui/core/math"
	"github.com/energye/examples/lcl/gpui/render/pipeline"
)

type recordingWidget struct {
	BaseWidget
	events []Event
}

func newRecordingWidget(rect math.Rect) *recordingWidget {
	w := &recordingWidget{BaseWidget: NewBaseWidget()}
	w.SetOwner(w)
	w.SetBounds(rect)
	return w
}

func (w *recordingWidget) HandleEvent(ctx *Context, event Event) bool {
	w.events = append(w.events, event)
	return true
}

func TestClampSize(t *testing.T) {
	got := ClampSize(math.NewVec2(20, 200), Constraints{
		Min: math.NewVec2(50, 40),
		Max: math.NewVec2(120, 100),
	})
	if got.X != 50 || got.Y != 100 {
		t.Fatalf("ClampSize() = %+v, want 50x100", got)
	}
}

func TestStateFlagsKeepEnabledAndFocusInSync(t *testing.T) {
	w := newRecordingWidget(math.NewRect(0, 0, 10, 10))

	w.SetStateFlag(StateDisabled, true)
	if w.Enabled() {
		t.Fatal("disabled state should disable widget")
	}
	if !w.HasState(StateDisabled) {
		t.Fatal("disabled state flag was not set")
	}

	w.SetStateFlag(StateDisabled, false)
	w.SetFocusable(true)
	w.Focus()
	if !w.Enabled() || !w.Focused() || !w.HasState(StateFocus) {
		t.Fatal("focus and enabled state were not synchronized")
	}
}

func TestHitTestReturnsConcreteOwner(t *testing.T) {
	box := NewBox(pipeline.BoxStyle{})
	box.SetBounds(math.NewRect(4, 6, 20, 30))

	hit := box.HitTest(math.NewVec2(10, 10))
	if hit != box {
		t.Fatalf("HitTest() = %T, want *Box owner", hit)
	}
}

func TestContainerDispatchesToTopmostChild(t *testing.T) {
	root := NewContainer()
	root.Layout(nil, math.NewRect(0, 0, 100, 100))
	bottom := newRecordingWidget(math.NewRect(10, 10, 50, 50))
	top := newRecordingWidget(math.NewRect(20, 20, 50, 50))
	root.Add(bottom)
	root.Add(top)

	handled := root.HandleEvent(nil, Event{Type: EventMouseDown, X: 30, Y: 30, Button: 1})
	if !handled {
		t.Fatal("expected topmost child to handle pointer event")
	}
	if len(top.events) != 1 {
		t.Fatalf("top events = %d, want 1", len(top.events))
	}
	if len(bottom.events) != 0 {
		t.Fatalf("bottom events = %d, want 0", len(bottom.events))
	}
}

func TestNestedContainerEventCoordinates(t *testing.T) {
	root := NewContainer()
	panel := NewContainer()
	child := newRecordingWidget(math.NewRect(10, 12, 20, 20))

	root.Layout(nil, math.NewRect(0, 0, 200, 200))
	panel.SetBounds(math.NewRect(30, 40, 100, 100))
	panel.Add(child)
	root.Add(panel)

	handled := root.HandleEvent(nil, Event{Type: EventMouseDown, X: 45, Y: 57, Button: 1})
	if !handled {
		t.Fatal("nested child did not handle event")
	}
	if len(child.events) != 1 {
		t.Fatalf("child events = %d, want 1", len(child.events))
	}
	event := child.events[0]
	if event.X != 15 || event.Y != 17 || event.LocalX != 5 || event.LocalY != 5 {
		t.Fatalf("event coordinates = X:%v Y:%v LocalX:%v LocalY:%v, want 15,17,5,5", event.X, event.Y, event.LocalX, event.LocalY)
	}
}

func TestFocusRegistrationAndRefresh(t *testing.T) {
	root := NewContainer()
	first := newRecordingWidget(math.NewRect(0, 0, 20, 20))
	second := newRecordingWidget(math.NewRect(24, 0, 20, 20))
	first.SetFocusable(true)
	root.Add(first)
	root.Add(second)

	root.FocusManager().Next()
	if root.FocusManager().Current() != first {
		t.Fatal("focus manager did not register focusable child added to container")
	}

	second.SetFocusable(true)
	root.FocusManager().Next()
	if root.FocusManager().Current() != second {
		t.Fatal("SetFocusable after Add should register with parent focus manager")
	}

	second.SetFocusable(false)
	root.RefreshFocus()
	if root.FocusManager().Current() != nil {
		t.Fatal("refresh should clear focus when focused widget is no longer focusable")
	}
}

func TestBoxClickUsesConcreteHandler(t *testing.T) {
	root := NewContainer()
	root.Layout(nil, math.NewRect(0, 0, 100, 100))
	box := NewBox(pipeline.BoxStyle{})
	box.SetBounds(math.NewRect(10, 10, 40, 24))
	clicked := false
	box.OnClick = func() {
		clicked = true
	}
	root.Add(box)

	root.HandleEvent(nil, Event{Type: EventMouseDown, X: 12, Y: 12, Button: 1})
	root.HandleEvent(nil, Event{Type: EventMouseUp, X: 12, Y: 12, Button: 1})
	if !clicked {
		t.Fatal("box click handler was not called")
	}
}

func TestContainerCapturedMouseUpReturnsToPressedChild(t *testing.T) {
	root := NewContainer()
	root.Layout(nil, math.NewRect(0, 0, 200, 120))
	child := newRecordingWidget(math.NewRect(10, 10, 40, 20))
	root.Add(child)

	root.HandleEvent(nil, Event{Type: EventMouseDown, X: 20, Y: 20, Button: 1})
	root.HandleEvent(nil, Event{Type: EventMouseUp, X: 160, Y: 90, Button: 1})

	if len(child.events) != 2 {
		t.Fatalf("child events = %d, want mouse down and captured mouse up", len(child.events))
	}
	if child.events[1].Type != EventMouseUp {
		t.Fatalf("second event = %v, want mouse up", child.events[1].Type)
	}
	if child.events[1].LocalX != 150 || child.events[1].LocalY != 80 {
		t.Fatalf("captured mouse up local = (%v,%v), want (150,80)", child.events[1].LocalX, child.events[1].LocalY)
	}
}
