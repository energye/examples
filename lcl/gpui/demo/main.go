// Demo: Ant Design Style GPU UI - Clean Example
// This demonstrates the clean separation between framework and business logic
package main

import (
	"fmt"
	"os"

	"github.com/energye/lcl/api/libname"

	"github.com/energye/examples/lcl/gpui/style/color"
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
	font := engine.Font()

	// Title
	title := widget.NewLabel("Ant Design Style GPU UI", font)
	title.SetPos(20, 20)
	title.SetSize(400, 24)
	title.SetColor(color.Primary)
	engine.AddWidget(title)

	// TextBox
	textbox := widget.NewTextBox("Enter text here...", font)
	textbox.SetPos(20, 60)
	textbox.SetSize(300, 32)
	textbox.SetOnChange(func(text string) {
		fmt.Println("Text:", text)
	})
	engine.AddWidget(textbox)

	// Buttons
	buttons := []struct {
		name string
		typ  widget.ButtonType
		x    float32
	}{
		{"Primary", widget.ButtonPrimary, 20},
		{"Default", widget.ButtonDefault, 130},
		{"Success", widget.ButtonSuccess, 240},
		{"Warning", widget.ButtonWarning, 350},
		{"Danger", widget.ButtonDanger, 460},
	}

	for _, b := range buttons {
		btn := widget.NewButton(b.name, b.typ, font)
		btn.SetPos(b.x, 110)
		btn.SetSize(100, 32)
		btn.SetOnClick(func() {
			text := textbox.Text()
			title.SetText(fmt.Sprintf("%s: %s", b.name, text))
			fmt.Printf("%s clicked! Text: %s\n", b.name, text)
		})
		engine.AddWidget(btn)
	}

	// Set focus
	engine.SetFocus(textbox)

	fmt.Println("✓ UI ready")
}
