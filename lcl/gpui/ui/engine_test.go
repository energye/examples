package ui

import (
	"testing"

	"github.com/energye/examples/lcl/gpui/core/math"
	"github.com/energye/examples/lcl/gpui/overlay"
	"github.com/energye/examples/lcl/gpui/render/pipeline"
	"github.com/energye/examples/lcl/gpui/widget"
)

func TestEngineTabFocusUsesPortalFocusTrap(t *testing.T) {
	engine := NewEngine()
	engine.SetSize(200, 200)

	rootControl := widget.NewBox(pipeline.BoxStyle{})
	rootControl.SetFocusable(true)
	engine.AddWidget(rootControl)

	portalControl := widget.NewBox(pipeline.BoxStyle{})
	portalControl.SetFocusable(true)
	engine.PortalHost().Add(portalControl, widget.PortalOptions{
		ID:        "modal",
		Kind:      overlay.LayerModal,
		ZIndex:    100,
		Bounds:    math.NewRect(20, 20, 80, 40),
		FocusTrap: true,
		HasMask:   true,
	})
	engine.PortalHost().Layout(engine.Context(), math.NewRect(0, 0, 200, 200))

	engine.HandleKeyDown(9, 0)
	if engine.PortalHost().FocusManager().Current() != portalControl {
		t.Fatal("tab should focus portal content while top portal traps focus")
	}
	if engine.FocusManager().Current() != nil {
		t.Fatal("root focus manager should not receive tab while portal traps focus")
	}
}

func TestEngineSynthesizesDoubleClick(t *testing.T) {
	engine := NewEngine()
	engine.SetSize(200, 200)

	button := widget.NewButton("Open")
	button.SetBounds(math.NewRect(0, 0, 80, 32))
	clicks := 0
	button.SetOnClick(func() {
		clicks++
	})
	engine.AddWidget(button)
	engine.Root().Layout(engine.Context(), math.NewRect(0, 0, 200, 200))

	engine.HandleMouseDown(10, 10, 0)
	engine.HandleMouseDown(10, 10, 0)
	if clicks != 1 {
		t.Fatalf("double click activations = %d, want 1", clicks)
	}
}
