package widget

import (
	"testing"

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
