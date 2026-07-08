// Demo: Ant Design Style GPU UI - Clean Example
// This demonstrates the clean separation between framework and business logic
package main

import (
	"fmt"
	"os"

	"github.com/energye/lcl/api/libname"

	"github.com/energye/examples/lcl/gpui/core/math"
	"github.com/energye/examples/lcl/gpui/render/pipeline"
	"github.com/energye/examples/lcl/gpui/style/token"
	"github.com/energye/examples/lcl/gpui/ui"
	"github.com/energye/examples/lcl/gpui/widget"
)

func main() {
	libname.UseWS = "gtk3"
	if ws := os.Getenv("GPUI_WS"); ws != "" {
		libname.UseWS = ws
	}
	// Create application (handles ALL framework initialization)
	app := ui.NewApplication("Ant Design Style GPU UI", 800, 600)

	// Setup UI (only business logic - no framework code!)
	app.OnSetup(func(engine *ui.Engine) {
		setupUI(engine)
	})

	// Run
	app.Run()
}

func setupUI(engine *ui.Engine) {
	tokens := token.Current()

	panel := widget.NewBox(pipeline.BoxStyle{
		Background:  tokens.Global.ColorBgContainer,
		BorderColor: tokens.Global.ColorBorder,
		BorderWidth: 1,
		Radius:      tokens.Global.RadiusLG,
		Shadows: []pipeline.Shadow{{
			Offset: math.NewVec2(0, 6),
			Blur:   14,
			Color:  math.NewColor(0, 0, 0, 0.10),
		}},
	})
	panel.SetPos(24, 24)
	panel.SetSize(420, 160)
	engine.AddWidget(panel)

	title := widget.NewText("Ant Design Framework Core")
	title.SetPos(48, 48)
	title.SetSize(360, 28)
	title.Font = engine.Font()
	title.Color = tokens.Global.ColorText
	title.SetEnabled(false)
	engine.AddWidget(title)

	status := widget.NewText("Box/Text primitives running on the new widget lifecycle")
	status.SetPos(48, 86)
	status.SetSize(360, 24)
	status.Font = engine.Font()
	status.Color = tokens.Global.ColorTextSecondary
	status.Ellipsis = true
	status.SetEnabled(false)
	engine.AddWidget(status)

	action := widget.NewBox(pipeline.BoxStyle{
		Background:  tokens.Global.ColorPrimary,
		BorderColor: tokens.Global.ColorPrimary,
		BorderWidth: 1,
		Radius:      tokens.Global.RadiusMD,
	})
	action.SetPos(48, 126)
	action.SetSize(128, 36)
	action.SetFocusable(true)
	action.OnClick = func() {
		status.Text = "Clicked: event dispatch, focus, and state are active"
		status.Invalidate()
		fmt.Println("framework box clicked")
	}
	engine.AddWidget(action)

	label := widget.NewText("Click")
	label.SetPos(87, 134)
	label.SetSize(80, 20)
	label.Font = engine.Font()
	label.Color = tokens.Global.ColorTextLight
	label.SetEnabled(false)
	engine.AddWidget(label)

	fmt.Println("✓ UI ready")
}
