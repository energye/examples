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

func assertWidgetRect(t *testing.T, got math.Rect, x, y, w, h float32) {
	t.Helper()
	if got.X != x || got.Y != y || got.W != w || got.H != h {
		t.Fatalf("rect = (%v,%v,%v,%v), want (%v,%v,%v,%v)", got.X, got.Y, got.W, got.H, x, y, w, h)
	}
}
