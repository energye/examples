package widget

import (
	"testing"

	"github.com/energye/examples/lcl/gpui/core/math"
	"github.com/energye/examples/lcl/gpui/layout"
)

func TestFlexLayoutAppliesChildBounds(t *testing.T) {
	root := NewFlex(layout.Row, 8)
	root.Style.Padding = layout.EdgeAll(4)
	root.Layout(nil, math.NewRect(0, 0, 200, 80))

	first := newRecordingWidget(math.Rect{})
	second := newRecordingWidget(math.Rect{})
	root.AddLayout(first, layout.Style{Width: layout.Px(50), Height: layout.Px(20)})
	root.AddLayout(second, layout.Style{Width: layout.Px(60), Height: layout.Px(24)})
	root.Layout(nil, math.NewRect(0, 0, 200, 80))

	assertWidgetRect(t, first.Bounds(), 4, 4, 50, 20)
	assertWidgetRect(t, second.Bounds(), 62, 4, 60, 24)
	if first.Parent() != root || second.Parent() != root {
		t.Fatal("layout children should keep the layout container as parent")
	}
}

func TestAlignCentersChild(t *testing.T) {
	root := NewAlign(layout.AlignCenter, layout.JustifyCenter)
	child := newRecordingWidget(math.Rect{})
	root.AddLayout(child, layout.Style{Width: layout.Px(40), Height: layout.Px(20)})

	root.Layout(nil, math.NewRect(0, 0, 120, 80))
	assertWidgetRect(t, child.Bounds(), 40, 30, 40, 20)
}

func TestWrapLayoutAppliesMultipleRows(t *testing.T) {
	root := NewWrap(10)
	for i := 0; i < 3; i++ {
		root.AddLayout(newRecordingWidget(math.Rect{}), layout.Style{Width: layout.Px(70), Height: layout.Px(20)})
	}

	root.Layout(nil, math.NewRect(0, 0, 150, 100))
	children := root.Children()
	assertWidgetRect(t, children[0].Bounds(), 0, 0, 70, 20)
	assertWidgetRect(t, children[1].Bounds(), 80, 0, 70, 20)
	assertWidgetRect(t, children[2].Bounds(), 0, 30, 70, 20)
}

func TestScrollAreaClampsAndRoutesScrolledCoordinates(t *testing.T) {
	root := NewScrollArea(layout.Style{Direction: layout.Column})
	child := newRecordingWidget(math.Rect{})
	root.AddLayout(child, layout.Style{Width: layout.Px(80), Height: layout.Px(200)})
	root.Layout(nil, math.NewRect(0, 0, 100, 60))

	root.SetScroll(0, 500)
	if root.Scroll().Y != 140 {
		t.Fatalf("scroll Y = %v, want 140", root.Scroll().Y)
	}
	if hit := root.HitTest(math.NewVec2(10, 10)); hit != child {
		t.Fatalf("hit = %T, want child under scrolled content", hit)
	}

	handled := root.HandleEvent(nil, Event{Type: EventMouseDown, X: 10, Y: 10, Button: 1})
	if !handled || len(child.events) != 1 {
		t.Fatal("scrolled child did not receive pointer event")
	}
	event := child.events[0]
	if event.X != 10 || event.Y != 150 || event.LocalX != 10 || event.LocalY != 150 {
		t.Fatalf("event coordinates = X:%v Y:%v LocalX:%v LocalY:%v, want 10,150,10,150", event.X, event.Y, event.LocalX, event.LocalY)
	}
}

func TestOverflowHiddenClipsWithoutScrolling(t *testing.T) {
	root := NewLayoutContainer(layout.Style{
		Direction: layout.Column,
		OverflowY: layout.OverflowHidden,
	})
	child := newRecordingWidget(math.Rect{})
	root.AddLayout(child, layout.Style{Width: layout.Px(80), Height: layout.Px(200)})
	root.Layout(nil, math.NewRect(0, 0, 100, 60))

	root.SetScroll(0, 100)
	if root.Scroll().Y != 0 {
		t.Fatalf("hidden overflow scroll Y = %v, want 0", root.Scroll().Y)
	}
}

func TestLayoutContainerFocusRegistration(t *testing.T) {
	root := NewFlex(layout.Row, 0)
	child := newRecordingWidget(math.Rect{})
	root.AddLayout(child, layout.Style{Width: layout.Px(20), Height: layout.Px(20)})
	child.SetFocusable(true)

	root.FocusManager().Next()
	if root.FocusManager().Current() != child {
		t.Fatal("focus manager did not register child through layout container parent")
	}
}

func TestRemoveNestedContainerClearsFocusRegistration(t *testing.T) {
	root := NewFlex(layout.Row, 0)
	nested := NewFlex(layout.Row, 0)
	child := newRecordingWidget(math.Rect{})
	child.SetFocusable(true)
	nested.AddLayout(child, layout.Style{Width: layout.Px(20), Height: layout.Px(20)})
	root.Add(nested)

	root.FocusManager().Next()
	if root.FocusManager().Current() != child {
		t.Fatal("nested focusable child was not registered")
	}

	root.Remove(nested)
	if root.FocusManager().Current() != nil {
		t.Fatal("removing nested container should clear focused child")
	}
	root.FocusManager().Next()
	if root.FocusManager().Current() != nil {
		t.Fatal("removed nested child should not remain in focus order")
	}
}

func TestLayoutContainerCapturedMouseUpOutsideClearsActive(t *testing.T) {
	root := NewFlex(layout.Row, 0)
	button := NewButton("Save")
	clicks := 0
	button.SetOnClick(func() {
		clicks++
	})
	root.AddLayout(button, layout.Style{Width: layout.Px(80), Height: layout.Px(32)})
	root.Layout(nil, math.NewRect(0, 0, 200, 100))

	root.HandleEvent(nil, Event{Type: EventMouseDown, X: 10, Y: 10, Button: 1})
	if !button.HasState(StateActive) {
		t.Fatal("button should become active after mouse down")
	}
	root.HandleEvent(nil, Event{Type: EventMouseUp, X: 150, Y: 80, Button: 1})
	if button.HasState(StateActive) {
		t.Fatal("button should clear active after outside mouse up")
	}
	if clicks != 0 {
		t.Fatalf("clicks = %d, want 0", clicks)
	}
}

func assertWidgetRect(t *testing.T, got math.Rect, x, y, w, h float32) {
	t.Helper()
	if got.X != x || got.Y != y || got.W != w || got.H != h {
		t.Fatalf("rect = (%v,%v,%v,%v), want (%v,%v,%v,%v)", got.X, got.Y, got.W, got.H, x, y, w, h)
	}
}

// TestScrollPositionAPI verifies that SetOnScroll callback fires correctly.
func TestScrollPositionAPI(t *testing.T) {
	root := NewScrollArea(layout.Style{
		Direction:  layout.Column,
		OverflowX:  layout.OverflowScroll,
		OverflowY:  layout.OverflowScroll,
	})
	// Add a tall child to enable scrolling
	child := newRecordingWidget(math.Rect{})
	root.AddLayout(child, layout.Style{Width: layout.Px(200), Height: layout.Px(500)})
	root.Layout(nil, math.NewRect(0, 0, 200, 100))

	// Track scroll callbacks
	scrollEvents := []math.Vec2{}
	root.SetOnScroll(func(x, y float32) {
		scrollEvents = append(scrollEvents, math.NewVec2(x, y))
	})

	// SetScroll should trigger callback
	root.SetScroll(0, 50)
	if len(scrollEvents) != 1 {
		t.Fatalf("expected 1 scroll event, got %d", len(scrollEvents))
	}
	if scrollEvents[0].Y != 50 {
		t.Fatalf("scroll Y = %v, want 50", scrollEvents[0].Y)
	}

	// SetScroll to same position should NOT trigger callback
	root.SetScroll(0, 50)
	if len(scrollEvents) != 1 {
		t.Fatalf("expected 1 scroll event (no change), got %d", len(scrollEvents))
	}

	// SetScroll to different position should trigger callback
	root.SetScroll(0, 100)
	if len(scrollEvents) != 2 {
		t.Fatalf("expected 2 scroll events, got %d", len(scrollEvents))
	}
	if scrollEvents[1].Y != 100 {
		t.Fatalf("scroll Y = %v, want 100", scrollEvents[1].Y)
	}
}

// TestScrollPositionAPIScrollTo verifies ScrollTo triggers callback.
func TestScrollPositionAPIScrollTo(t *testing.T) {
	root := NewScrollArea(layout.Style{
		Direction:  layout.Column,
		OverflowX:  layout.OverflowScroll,
		OverflowY:  layout.OverflowScroll,
	})
	child := newRecordingWidget(math.Rect{})
	root.AddLayout(child, layout.Style{Width: layout.Px(200), Height: layout.Px(500)})
	root.Layout(nil, math.NewRect(0, 0, 200, 100))

	scrollEvents := []math.Vec2{}
	root.SetOnScroll(func(x, y float32) {
		scrollEvents = append(scrollEvents, math.NewVec2(x, y))
	})

	// ScrollTo should trigger callback - scroll to a position within bounds
	// rect.Y=150, rect.H=50, viewport.H=100
	// rect.Y + rect.H = 200 > scroll.Y + viewport.H = 100
	// So scroll.Y = rect.Y + rect.H - viewport.H = 200 - 100 = 100
	root.ScrollTo(math.NewRect(0, 150, 50, 50))
	if len(scrollEvents) != 1 {
		t.Fatalf("expected 1 scroll event, got %d", len(scrollEvents))
	}
	if scrollEvents[0].Y != 100 {
		t.Fatalf("scroll Y = %v, want 100", scrollEvents[0].Y)
	}
}

// TestScrollPositionAPIMouseWheel verifies mouse wheel triggers callback.
func TestScrollPositionAPIMouseWheel(t *testing.T) {
	root := NewScrollArea(layout.Style{
		Direction:  layout.Column,
		OverflowX:  layout.OverflowScroll,
		OverflowY:  layout.OverflowScroll,
	})
	child := newRecordingWidget(math.Rect{})
	root.AddLayout(child, layout.Style{Width: layout.Px(200), Height: layout.Px(500)})
	root.Layout(nil, math.NewRect(0, 0, 200, 100))

	scrollEvents := []math.Vec2{}
	root.SetOnScroll(func(x, y float32) {
		scrollEvents = append(scrollEvents, math.NewVec2(x, y))
	})

	// Mouse wheel should trigger callback
	// DeltaY = -1 means scroll down (positive Y direction)
	root.HandleEvent(nil, Event{Type: EventMouseWheel, X: 10, Y: 10, DeltaY: -1})
	if len(scrollEvents) != 1 {
		t.Fatalf("expected 1 scroll event, got %d", len(scrollEvents))
	}
	// scroll.Y should be 30 (scrollSpeed=30, DeltaY=-1, so scroll.Y -= (-1)*30 = +30)
	if scrollEvents[0].Y != 30 {
		t.Fatalf("scroll Y = %v, want 30", scrollEvents[0].Y)
	}
}

// TestScrollPositionAPINilHandler verifies nil handler is safe.
func TestScrollPositionAPINilHandler(t *testing.T) {
	root := NewScrollArea(layout.Style{
		Direction:  layout.Column,
		OverflowX:  layout.OverflowScroll,
		OverflowY:  layout.OverflowScroll,
	})
	child := newRecordingWidget(math.Rect{})
	root.AddLayout(child, layout.Style{Width: layout.Px(200), Height: layout.Px(500)})
	root.Layout(nil, math.NewRect(0, 0, 200, 100))

	// No handler set - should not panic
	root.SetScroll(0, 50)
	root.ScrollTo(math.NewRect(0, 200, 50, 50))
}

// TestScrollPositionAPIGetters verifies Scroll and ContentSize getters.
func TestScrollPositionAPIGetters(t *testing.T) {
	root := NewScrollArea(layout.Style{
		Direction:  layout.Column,
		OverflowX:  layout.OverflowScroll,
		OverflowY:  layout.OverflowScroll,
	})
	child := newRecordingWidget(math.Rect{})
	root.AddLayout(child, layout.Style{Width: layout.Px(200), Height: layout.Px(500)})
	root.Layout(nil, math.NewRect(0, 0, 200, 100))

	// Initial scroll should be 0,0
	scroll := root.Scroll()
	if scroll.X != 0 || scroll.Y != 0 {
		t.Fatalf("initial scroll = (%v,%v), want (0,0)", scroll.X, scroll.Y)
	}

	// ContentSize should reflect child size
	contentSize := root.ContentSize()
	if contentSize.Y < 500 {
		t.Fatalf("content size Y = %v, want >= 500", contentSize.Y)
	}

	// Set scroll and verify
	root.SetScroll(0, 100)
	scroll = root.Scroll()
	if scroll.Y != 100 {
		t.Fatalf("scroll Y = %v, want 100", scroll.Y)
	}
}

// TestVirtualScrollBasic verifies basic virtual scrolling setup.
func TestVirtualScrollBasic(t *testing.T) {
	root := NewScrollArea(layout.Style{
		Direction:  layout.Column,
		OverflowX:  layout.OverflowScroll,
		OverflowY:  layout.OverflowScroll,
	})
	root.Layout(nil, math.NewRect(0, 0, 200, 100))

	// Enable virtual scrolling with 100 items of height 20
	root.SetVirtualScroll(20, 100)

	if !root.IsVirtualScrollEnabled() {
		t.Fatal("virtual scroll should be enabled")
	}
	if root.TotalItems() != 100 {
		t.Fatalf("total items = %d, want 100", root.TotalItems())
	}

	// Check initial visible range
	start, end := root.VisibleRange()
	if start != 0 {
		t.Fatalf("visible start = %d, want 0", start)
	}
	// viewport height = 100, item height = 20, so 5 items visible + 3 buffer = 8
	if end != 8 {
		t.Fatalf("visible end = %d, want 8", end)
	}
}

// TestVirtualScrollOnScroll verifies visible range updates on scroll.
func TestVirtualScrollOnScroll(t *testing.T) {
	root := NewScrollArea(layout.Style{
		Direction:  layout.Column,
		OverflowX:  layout.OverflowScroll,
		OverflowY:  layout.OverflowScroll,
	})
	// Add a tall child to enable scrolling
	child := newRecordingWidget(math.Rect{})
	root.AddLayout(child, layout.Style{Width: layout.Px(200), Height: layout.Px(2000)})
	root.Layout(nil, math.NewRect(0, 0, 200, 100))

	// Enable virtual scrolling with 100 items of height 20
	root.SetVirtualScroll(20, 100)

	visibleChanges := []int{}
	root.SetOnVisibleChanged(func(start, end int) {
		visibleChanges = append(visibleChanges, start)
	})

	// Scroll down
	root.SetScroll(0, 200)

	// Should have triggered visible range change
	if len(visibleChanges) == 0 {
		t.Fatal("expected visible range change callback")
	}

	// Check that visible range updated
	start, end := root.VisibleRange()
	if start < 7 {
		t.Fatalf("visible start = %d, want >= 7 (scrolled 200px / 20px per item = 10 items)", start)
	}
	if end <= start {
		t.Fatalf("visible end (%d) should be > start (%d)", end, start)
	}
}

// TestVirtualScrollCallback verifies OnVisibleChanged callback.
func TestVirtualScrollCallback(t *testing.T) {
	root := NewScrollArea(layout.Style{
		Direction:  layout.Column,
		OverflowX:  layout.OverflowScroll,
		OverflowY:  layout.OverflowScroll,
	})
	child := newRecordingWidget(math.Rect{})
	root.AddLayout(child, layout.Style{Width: layout.Px(200), Height: layout.Px(2000)})
	root.Layout(nil, math.NewRect(0, 0, 200, 100))

	root.SetVirtualScroll(20, 100)

	callCount := 0
	root.SetOnVisibleChanged(func(start, end int) {
		callCount++
	})

	// Multiple scrolls should trigger callback only when range changes
	root.SetScroll(0, 100)
	root.SetScroll(0, 100) // Same position - no change
	root.SetScroll(0, 200)

	// Should have at least 2 calls (initial + scroll to 200)
	if callCount < 2 {
		t.Fatalf("expected at least 2 visible range changes, got %d", callCount)
	}
}

// TestVirtualScrollDisabledByDefault verifies virtual scroll is off by default.
func TestVirtualScrollDisabledByDefault(t *testing.T) {
	root := NewScrollArea(layout.Style{
		Direction:  layout.Column,
		OverflowX:  layout.OverflowScroll,
		OverflowY:  layout.OverflowScroll,
	})

	if root.IsVirtualScrollEnabled() {
		t.Fatal("virtual scroll should be disabled by default")
	}
	if root.TotalItems() != 0 {
		t.Fatalf("total items = %d, want 0", root.TotalItems())
	}
	start, end := root.VisibleRange()
	if start != 0 || end != 0 {
		t.Fatalf("visible range = (%d,%d), want (0,0)", start, end)
	}
}
