package widget

import (
	"testing"
	"time"

	"github.com/energye/examples/lcl/gpui/core/math"
)

func TestSwitchCheckedStartsThumbTransition(t *testing.T) {
	s := NewSwitch()
	s.SetChecked(true)

	timeline := s.Timeline()
	if timeline == nil {
		t.Fatal("switch checked change should create a motion timeline")
	}
	thumb := timeline.Get(switchThumbPosition)
	if thumb == nil {
		t.Fatal("switch thumb transition should be registered")
	}
	if !thumb.Running() {
		t.Fatal("switch thumb transition should run after checked change")
	}
	if thumb.Value() != 0 {
		t.Fatalf("thumb value = %v, want transition to start from unchecked position", thumb.Value())
	}
	thumb.Update(180 * time.Millisecond)
	if thumb.Value() != 1 {
		t.Fatalf("thumb value = %v, want checked position after transition", thumb.Value())
	}
}

func TestSwitchRapidClicksToggleEveryActivation(t *testing.T) {
	s := NewSwitch()
	s.SetBounds(math.NewRect(0, 0, 44, 22))
	changes := 0
	s.SetOnChange(func(bool) {
		changes++
	})

	s.HandleEvent(nil, Event{Type: EventMouseDown, LocalX: 10, LocalY: 10, Button: 1})
	if !s.Checked() {
		t.Fatal("first mouse down should switch on immediately")
	}
	s.HandleEvent(nil, Event{Type: EventMouseUp, LocalX: 10, LocalY: 10, Button: 1})
	if !s.Checked() {
		t.Fatal("mouse up should not toggle switch back off")
	}

	s.HandleEvent(nil, Event{Type: EventDoubleClick, LocalX: 10, LocalY: 10, Button: 1})
	if s.Checked() {
		t.Fatal("second rapid activation should switch off")
	}
	if changes != 2 {
		t.Fatalf("changes = %d, want 2 toggles", changes)
	}
}

func TestSwitchLoadingStartsSpinnerMotion(t *testing.T) {
	s := NewSwitch()
	s.SetLoading(true)

	spin := s.Timeline().Get(controlLoadingSpin)
	if spin == nil {
		t.Fatal("loading switch should register spinner transition")
	}
	if !spin.Running() || !spin.Loop() {
		t.Fatal("loading switch spinner should run as a loop")
	}
	if s.HandleEvent(nil, Event{Type: EventMouseDown, LocalX: 10, LocalY: 10, Button: 1}) {
		t.Fatal("loading switch should not handle activation")
	}
}
