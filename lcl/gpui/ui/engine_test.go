package ui

import (
	"testing"
	"time"

	"github.com/energye/examples/lcl/gpui/core/math"
	"github.com/energye/examples/lcl/gpui/layout"
	"github.com/energye/examples/lcl/gpui/motion"
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

func TestEngineUpdatesNestedLayoutAnimations(t *testing.T) {
	engine := NewEngine()
	panel := widget.NewLayoutContainer(layout.Style{Direction: layout.Row})
	control := widget.NewButton("Animated")
	control.AddTransition("opacity", 0, 100*time.Millisecond, motion.Linear)
	control.SetMotionTarget("opacity", 1)
	panel.Add(control)
	engine.AddWidget(panel)

	engine.updateAnimations(50 * time.Millisecond)

	if got := control.MotionValue("opacity", 0); got != 0.5 {
		t.Fatalf("opacity = %v, want 0.5 after nested animation update", got)
	}
}

func TestEngineUpdatesPortalAnimations(t *testing.T) {
	engine := NewEngine()
	control := widget.NewButton("Portal")
	control.AddTransition("opacity", 0, 100*time.Millisecond, motion.Linear)
	control.SetMotionTarget("opacity", 1)
	engine.PortalHost().Add(control, widget.PortalOptions{
		ID:     "popup",
		Kind:   overlay.LayerPopup,
		ZIndex: 10,
		Bounds: math.NewRect(0, 0, 80, 32),
	})

	engine.updateAnimations(50 * time.Millisecond)

	if got := control.MotionValue("opacity", 0); got != 0.5 {
		t.Fatalf("portal opacity = %v, want 0.5 after animation update", got)
	}
}
