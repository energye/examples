package widget

import (
	"testing"
	"time"
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
