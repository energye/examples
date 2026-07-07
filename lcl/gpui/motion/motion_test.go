package motion

import (
	"testing"
	"time"

	"github.com/energye/examples/lcl/gpui/core/math"
)

func TestTransitionUpdate(t *testing.T) {
	tr := NewTransition(0, 100*time.Millisecond, Linear)
	tr.SetTarget(10)
	tr.Update(50 * time.Millisecond)
	if tr.Value() != 5 {
		t.Fatalf("value = %v, want 5", tr.Value())
	}
	if !tr.Running() {
		t.Fatal("transition should still be running")
	}
	tr.Update(50 * time.Millisecond)
	if tr.Value() != 10 {
		t.Fatalf("value = %v, want 10", tr.Value())
	}
	if tr.State() != Done {
		t.Fatalf("state = %v, want Done", tr.State())
	}
}

func TestTransitionRetargetsFromCurrentValue(t *testing.T) {
	tr := NewTransition(0, 100*time.Millisecond, Linear)
	tr.SetTarget(10)
	tr.Update(50 * time.Millisecond)
	tr.SetTarget(20)
	tr.Update(50 * time.Millisecond)
	if tr.Value() != 12.5 {
		t.Fatalf("value = %v, want 12.5", tr.Value())
	}
}

func TestTimeline(t *testing.T) {
	timeline := NewTimeline()
	timeline.Add("opacity", NewTransition(0, 100*time.Millisecond, Linear))
	if !timeline.SetTarget("opacity", 1) {
		t.Fatal("expected target set")
	}
	timeline.Update(100 * time.Millisecond)
	if timeline.Running() {
		t.Fatal("timeline should be done")
	}
	if got := timeline.Get("opacity").Value(); got != 1 {
		t.Fatalf("opacity = %v, want 1", got)
	}
}

func TestTimelineSkipsNilTransitions(t *testing.T) {
	timeline := NewTimeline()
	timeline.Add("missing", nil)
	timeline.Update(16 * time.Millisecond)
	if timeline.Running() {
		t.Fatal("timeline with nil transition should not be running")
	}
}

func TestColorTransition(t *testing.T) {
	ct := NewColorTransition(math.NewColor(0, 0, 0, 1), 100*time.Millisecond, Linear)
	ct.SetTarget(math.NewColor(1, 0.5, 0.25, 0.5))
	ct.Update(50 * time.Millisecond)
	got := ct.Value()
	if got.R != 0.5 || got.G != 0.25 || got.B != 0.125 || got.A != 0.75 {
		t.Fatalf("color = %#v", got)
	}
}

func TestEaseInOutBounds(t *testing.T) {
	if EaseInOut(-1) != 0 {
		t.Fatal("ease should clamp low")
	}
	if EaseInOut(2) != 1 {
		t.Fatal("ease should clamp high")
	}
}
