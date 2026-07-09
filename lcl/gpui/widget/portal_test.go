package widget

import (
	"testing"
	"time"

	"github.com/energye/examples/lcl/gpui/core/math"
	"github.com/energye/examples/lcl/gpui/overlay"
)

func TestPortalHostLayoutPlacesContentInViewport(t *testing.T) {
	host := NewPortalHost(nil)
	content := newRecordingWidget(math.Rect{})
	content.SetPreferredSize(math.NewVec2(80, 40))
	host.Add(content, PortalOptions{
		ID:        "popup",
		ZIndex:    10,
		Anchor:    math.NewRect(20, 20, 40, 20),
		Placement: overlay.BottomLeft,
		Offset:    math.NewVec2(0, 4),
		Clamp:     true,
	})

	host.Layout(nil, math.NewRect(0, 0, 200, 200))
	portal, ok := host.Portal("popup")
	if !ok {
		t.Fatal("portal was not added")
	}
	assertWidgetRect(t, portal.Layer.Bounds, 20, 44, 80, 40)
	assertWidgetRect(t, content.Bounds(), 0, 0, 80, 40)
}

func TestPortalHostRoutesPointerToTopmostPortal(t *testing.T) {
	host := NewPortalHost(nil)
	low := newRecordingWidget(math.Rect{})
	high := newRecordingWidget(math.Rect{})
	host.Add(low, PortalOptions{ID: "low", ZIndex: 10, Bounds: math.NewRect(0, 0, 100, 100)})
	host.Add(high, PortalOptions{ID: "high", ZIndex: 20, Bounds: math.NewRect(10, 10, 100, 100)})
	host.Layout(nil, math.NewRect(0, 0, 300, 300))

	handled := host.HandleEvent(nil, Event{Type: EventMouseDown, X: 20, Y: 20, Button: 1})
	if !handled {
		t.Fatal("topmost portal should handle pointer")
	}
	if len(high.events) != 1 {
		t.Fatalf("high events = %d, want 1", len(high.events))
	}
	if len(low.events) != 0 {
		t.Fatalf("low events = %d, want 0", len(low.events))
	}
	event := high.events[0]
	if event.X != 10 || event.Y != 10 || event.LocalX != 10 || event.LocalY != 10 {
		t.Fatalf("event = X:%v Y:%v LocalX:%v LocalY:%v, want all 10", event.X, event.Y, event.LocalX, event.LocalY)
	}
}

func TestPortalHostDismissOutside(t *testing.T) {
	host := NewPortalHost(nil)
	content := newRecordingWidget(math.Rect{})
	dismissed := ""
	host.Add(content, PortalOptions{
		ID:             "popup",
		ZIndex:         10,
		Bounds:         math.NewRect(20, 20, 80, 40),
		CloseOnOutside: true,
		OnDismiss: func(id string) {
			dismissed = id
		},
	})
	host.Layout(nil, math.NewRect(0, 0, 200, 200))

	if !host.HandleEvent(nil, Event{Type: EventMouseDown, X: 150, Y: 150, Button: 1}) {
		t.Fatal("outside dismiss should consume pointer down")
	}
	if dismissed != "popup" {
		t.Fatalf("dismissed = %q, want popup", dismissed)
	}
	if _, ok := host.Portal("popup"); ok {
		t.Fatal("dismissed portal should be removed")
	}
}

func TestPortalHostMaskConsumesOutsidePointer(t *testing.T) {
	host := NewPortalHost(nil)
	content := newRecordingWidget(math.Rect{})
	host.Add(content, PortalOptions{
		ID:      "modal",
		ZIndex:  100,
		Bounds:  math.NewRect(40, 40, 80, 60),
		HasMask: true,
	})
	host.Layout(nil, math.NewRect(0, 0, 200, 200))

	if !host.HandleEvent(nil, Event{Type: EventMouseDown, X: 10, Y: 10, Button: 1}) {
		t.Fatal("modal mask should consume outside pointer events")
	}
	if len(content.events) != 0 {
		t.Fatal("outside mask click should not be delivered to modal content")
	}
}

func TestPortalHostFocusRoutesKeyboard(t *testing.T) {
	host := NewPortalHost(nil)
	content := newRecordingWidget(math.Rect{})
	content.SetFocusable(true)
	host.Add(content, PortalOptions{ID: "popup", ZIndex: 10, Bounds: math.NewRect(0, 0, 80, 40)})
	host.Layout(nil, math.NewRect(0, 0, 200, 200))

	host.HandleEvent(nil, Event{Type: EventMouseDown, X: 10, Y: 10, Button: 1})
	if host.FocusManager().Current() != content {
		t.Fatal("pointer down should focus focusable portal content")
	}
	host.HandleEvent(nil, Event{Type: EventKeyDown, Key: 65})
	if len(content.events) != 2 {
		t.Fatalf("content events = %d, want pointer and key events", len(content.events))
	}
	if content.events[1].Type != EventKeyDown || content.events[1].Key != 65 {
		t.Fatal("focused portal content should receive key events")
	}
}

func TestPortalHostCapturedMouseUpOutsideClearsActive(t *testing.T) {
	host := NewPortalHost(nil)
	button := NewButton("Save")
	clicks := 0
	button.SetOnClick(func() {
		clicks++
	})
	host.Add(button, PortalOptions{ID: "popup", ZIndex: 10, Bounds: math.NewRect(20, 20, 80, 32)})
	host.Layout(nil, math.NewRect(0, 0, 200, 200))

	host.HandleEvent(nil, Event{Type: EventMouseDown, X: 30, Y: 30, Button: 1})
	if !button.HasState(StateActive) {
		t.Fatal("button should become active after portal mouse down")
	}
	host.HandleEvent(nil, Event{Type: EventMouseUp, X: 160, Y: 160, Button: 1})
	if button.HasState(StateActive) {
		t.Fatal("button should clear active after portal outside mouse up")
	}
	if clicks != 0 {
		t.Fatalf("clicks = %d, want 0", clicks)
	}
}

// TestPortalAnimationFade verifies that PortalAnimFade creates an animation progress.
func TestPortalAnimationFade(t *testing.T) {
	host := NewPortalHost(nil)
	content := newRecordingWidget(math.Rect{})
	content.SetPreferredSize(math.NewVec2(80, 40))

	host.Add(content, PortalOptions{
		ID:         "fade-popup",
		ZIndex:     10,
		Bounds:     math.NewRect(50, 50, 80, 40),
		Animation:  PortalAnimFade,
		AnimDuration: 200 * time.Millisecond,
	})

	portal, ok := host.Portal("fade-popup")
	if !ok {
		t.Fatal("portal was not added")
	}
	if portal.animation != PortalAnimFade {
		t.Fatalf("animation = %v, want PortalAnimFade", portal.animation)
	}
	if portal.animProgress == nil {
		t.Fatal("animProgress should not be nil for animated portal")
	}
	if !portal.entering {
		t.Fatal("portal should be entering after creation")
	}
	if portal.exiting {
		t.Fatal("portal should not be exiting after creation")
	}
}

// TestPortalAnimationSlideDown verifies that PortalAnimSlideDown creates correct animation.
func TestPortalAnimationSlideDown(t *testing.T) {
	host := NewPortalHost(nil)
	content := newRecordingWidget(math.Rect{})
	content.SetPreferredSize(math.NewVec2(80, 40))

	host.Add(content, PortalOptions{
		ID:         "slide-popup",
		ZIndex:     10,
		Bounds:     math.NewRect(50, 50, 80, 40),
		Animation:  PortalAnimSlideDown,
		AnimDuration: 150 * time.Millisecond,
	})

	portal, ok := host.Portal("slide-popup")
	if !ok {
		t.Fatal("portal was not added")
	}
	if portal.animation != PortalAnimSlideDown {
		t.Fatalf("animation = %v, want PortalAnimSlideDown", portal.animation)
	}
}

// TestPortalAnimationZoom verifies that PortalAnimZoom creates correct animation.
func TestPortalAnimationZoom(t *testing.T) {
	host := NewPortalHost(nil)
	content := newRecordingWidget(math.Rect{})
	content.SetPreferredSize(math.NewVec2(80, 40))

	host.Add(content, PortalOptions{
		ID:         "zoom-popup",
		ZIndex:     10,
		Bounds:     math.NewRect(50, 50, 80, 40),
		Animation:  PortalAnimZoom,
		AnimDuration: 250 * time.Millisecond,
	})

	portal, ok := host.Portal("zoom-popup")
	if !ok {
		t.Fatal("portal was not added")
	}
	if portal.animation != PortalAnimZoom {
		t.Fatalf("animation = %v, want PortalAnimZoom", portal.animation)
	}
}

// TestPortalAnimationNone verifies that PortalAnimNone does not create animation.
func TestPortalAnimationNone(t *testing.T) {
	host := NewPortalHost(nil)
	content := newRecordingWidget(math.Rect{})
	content.SetPreferredSize(math.NewVec2(80, 40))

	host.Add(content, PortalOptions{
		ID:         "no-anim-popup",
		ZIndex:     10,
		Bounds:     math.NewRect(50, 50, 80, 40),
		Animation:  PortalAnimNone,
	})

	portal, ok := host.Portal("no-anim-popup")
	if !ok {
		t.Fatal("portal was not added")
	}
	if portal.animation != PortalAnimNone {
		t.Fatalf("animation = %v, want PortalAnimNone", portal.animation)
	}
	if portal.animProgress != nil {
		t.Fatal("animProgress should be nil for non-animated portal")
	}
	if portal.entering {
		t.Fatal("portal should not be entering for non-animated portal")
	}
}

// TestPortalAnimationRemoveStartsExit verifies that Remove starts exit animation.
func TestPortalAnimationRemoveStartsExit(t *testing.T) {
	host := NewPortalHost(nil)
	content := newRecordingWidget(math.Rect{})
	content.SetPreferredSize(math.NewVec2(80, 40))

	host.Add(content, PortalOptions{
		ID:         "exit-popup",
		ZIndex:     10,
		Bounds:     math.NewRect(50, 50, 80, 40),
		Animation:  PortalAnimFade,
		AnimDuration: 200 * time.Millisecond,
	})

	portal, ok := host.Portal("exit-popup")
	if !ok {
		t.Fatal("portal was not added")
	}

	// Start exit animation
	host.Remove("exit-popup")

	if !portal.exiting {
		t.Fatal("portal should be exiting after Remove")
	}
	if portal.entering {
		t.Fatal("portal should not be entering after Remove")
	}
	if portal.animProgress == nil {
		t.Fatal("animProgress should not be nil during exit")
	}
	// Portal should still exist (waiting for animation to complete)
	_, ok = host.Portal("exit-popup")
	if !ok {
		t.Fatal("portal should still exist during exit animation")
	}
}

// TestPortalAnimationImmediateRemove verifies that non-animated portals are removed immediately.
func TestPortalAnimationImmediateRemove(t *testing.T) {
	host := NewPortalHost(nil)
	content := newRecordingWidget(math.Rect{})
	content.SetPreferredSize(math.NewVec2(80, 40))

	host.Add(content, PortalOptions{
		ID:        "immediate-popup",
		ZIndex:    10,
		Bounds:    math.NewRect(50, 50, 80, 40),
		Animation: PortalAnimNone,
	})

	_, ok := host.Portal("immediate-popup")
	if !ok {
		t.Fatal("portal was not added")
	}

	// Remove immediately (no animation)
	host.Remove("immediate-popup")

	_, ok = host.Portal("immediate-popup")
	if ok {
		t.Fatal("portal should be removed immediately for non-animated portal")
	}
}

// TestPortalAnimationDefaultDuration verifies default duration is applied.
func TestPortalAnimationDefaultDuration(t *testing.T) {
	host := NewPortalHost(nil)
	content := newRecordingWidget(math.Rect{})
	content.SetPreferredSize(math.NewVec2(80, 40))

	host.Add(content, PortalOptions{
		ID:        "default-duration-popup",
		ZIndex:    10,
		Bounds:    math.NewRect(50, 50, 80, 40),
		Animation: PortalAnimFade,
		// AnimDuration not set - should use default 200ms
	})

	portal, ok := host.Portal("default-duration-popup")
	if !ok {
		t.Fatal("portal was not added")
	}
	if portal.animDuration != 200*time.Millisecond {
		t.Fatalf("animDuration = %v, want 200ms (default)", portal.animDuration)
	}
}
