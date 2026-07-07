package overlay

import (
	"testing"

	"github.com/energye/examples/lcl/gpui/core/math"
)

func TestManagerOrdersAndHitTestsTopmost(t *testing.T) {
	manager := NewManager()
	manager.Add(Layer{ID: "low", ZIndex: 10, Bounds: math.NewRect(0, 0, 100, 100)})
	manager.Add(Layer{ID: "high", ZIndex: 20, Bounds: math.NewRect(20, 20, 100, 100)})

	layers := manager.Layers()
	if layers[0].ID != "low" || layers[1].ID != "high" {
		t.Fatalf("unexpected layer order: %v, %v", layers[0].ID, layers[1].ID)
	}

	layer, ok := manager.TopmostAt(30, 30)
	if !ok || layer.ID != "high" {
		t.Fatalf("topmost = %v/%v, want high/true", layer.ID, ok)
	}
}

func TestDismissTargets(t *testing.T) {
	manager := NewManager()
	manager.Add(Layer{ID: "popup1", ZIndex: 10, Bounds: math.NewRect(0, 0, 100, 100), Options: Options{CloseOnOutside: true}})
	manager.Add(Layer{ID: "popup2", ZIndex: 20, Bounds: math.NewRect(120, 0, 100, 100), Options: Options{CloseOnOutside: true}})

	targets := manager.DismissTargets(300, 300)
	if len(targets) != 2 || targets[0].ID != "popup2" || targets[1].ID != "popup1" {
		t.Fatalf("unexpected dismiss targets: %#v", targets)
	}

	targets = manager.DismissTargets(130, 10)
	if len(targets) != 0 {
		t.Fatalf("click inside top popup should dismiss none, got %d", len(targets))
	}
}

func TestPlacementFlipAndClamp(t *testing.T) {
	viewport := math.NewRect(0, 0, 200, 200)
	anchor := math.NewRect(40, 170, 40, 20)
	size := math.NewVec2(80, 60)

	rect := Place(anchor, size, viewport, BottomLeft, PlacementOptions{
		Offset: math.NewVec2(0, 4),
		Flip:   true,
		Clamp:  true,
	})

	if rect.Y >= anchor.Y {
		t.Fatalf("expected popup to flip above anchor, got y=%v anchorY=%v", rect.Y, anchor.Y)
	}
	if rect.X < viewport.X || rect.X+rect.W > viewport.X+viewport.W {
		t.Fatalf("expected clamped x inside viewport, got %#v", rect)
	}
}

func TestPlacementCenter(t *testing.T) {
	viewport := math.NewRect(0, 0, 200, 200)
	anchor := math.NewRect(50, 50, 100, 100)
	rect := Place(anchor, math.NewVec2(40, 20), viewport, Center, PlacementOptions{})
	if rect.X != 80 || rect.Y != 90 {
		t.Fatalf("center rect = (%v,%v), want (80,90)", rect.X, rect.Y)
	}
}
