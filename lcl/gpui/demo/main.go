// Demo: Ant Design Style GPU UI - All Controls Showcase
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
	app := ui.NewApplication("Ant Design Controls Showcase", 900, 700)

	app.OnSetup(func(engine *ui.Engine) {
		setupUI(engine)
	})

	app.Run()
}

func setupUI(engine *ui.Engine) {
	tokens := token.Current()

	// ========== Section 1: Panel with Title ==========
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
	panel.SetSize(852, 652)
	engine.AddWidget(panel)

	// Title
	title := widget.NewText("Ant Design Component Showcase")
	title.SetPos(48, 48)
	title.SetSize(400, 32)
	title.SetFont(engine.Font())
	title.SetColor(tokens.Global.ColorText)
	engine.AddWidget(title)

	// Subtitle
	subtitle := widget.NewText("GPU-accelerated UI framework with token-driven design system")
	subtitle.SetPos(48, 84)
	subtitle.SetSize(500, 24)
	subtitle.SetFont(engine.Font())
	subtitle.SetColor(tokens.Global.ColorTextSecondary)
	engine.AddWidget(subtitle)

	// ========== Section 2: Buttons ==========
	btnLabel := widget.NewText("Buttons")
	btnLabel.SetPos(48, 130)
	btnLabel.SetSize(200, 24)
	btnLabel.SetFont(engine.Font())
	btnLabel.SetColor(tokens.Global.ColorText)
	engine.AddWidget(btnLabel)

	// Default button
	btnDefault := widget.NewButton("Default")
	btnDefault.SetPos(48, 160)
	btnDefault.SetSize(96, 36)
	btnDefault.SetFont(engine.Font())
	btnDefault.SetOnClick(func() {
		fmt.Println("Default clicked")
	})
	engine.AddWidget(btnDefault)

	// Primary button
	btnPrimary := widget.NewButton("Primary")
	btnPrimary.SetKind(widget.ButtonPrimary)
	btnPrimary.SetPos(156, 160)
	btnPrimary.SetSize(96, 36)
	btnPrimary.SetFont(engine.Font())
	btnPrimary.SetOnClick(func() {
		fmt.Println("Primary clicked")
	})
	engine.AddWidget(btnPrimary)

	// Danger button
	btnDanger := widget.NewButton("Danger")
	btnDanger.SetKind(widget.ButtonPrimary)
	btnDanger.SetDanger(true)
	btnDanger.SetPos(264, 160)
	btnDanger.SetSize(96, 36)
	btnDanger.SetFont(engine.Font())
	engine.AddWidget(btnDanger)

	// Link button
	btnLink := widget.NewButton("Link")
	btnLink.SetKind(widget.ButtonLink)
	btnLink.SetPos(372, 160)
	btnLink.SetSize(96, 36)
	btnLink.SetFont(engine.Font())
	engine.AddWidget(btnLink)

	// Ghost button
	btnGhost := widget.NewButton("Ghost")
	btnGhost.SetGhost(true)
	btnGhost.SetPos(480, 160)
	btnGhost.SetSize(96, 36)
	btnGhost.SetFont(engine.Font())
	engine.AddWidget(btnGhost)

	// Disabled button
	btnDisabled := widget.NewButton("Disabled")
	btnDisabled.SetPos(588, 160)
	btnDisabled.SetSize(96, 36)
	btnDisabled.SetFont(engine.Font())
	btnDisabled.SetEnabled(false)
	engine.AddWidget(btnDisabled)

	// ========== Section 3: Checkbox ==========
	cbLabel := widget.NewText("Checkboxes")
	cbLabel.SetPos(48, 220)
	cbLabel.SetSize(200, 24)
	cbLabel.SetFont(engine.Font())
	cbLabel.SetColor(tokens.Global.ColorText)
	engine.AddWidget(cbLabel)

	cb1 := widget.NewCheckbox("Checked")
	cb1.SetPos(48, 250)
	cb1.SetSize(120, 24)
	cb1.SetFont(engine.Font())
	cb1.SetChecked(true)
	engine.AddWidget(cb1)

	cb2 := widget.NewCheckbox("Unchecked")
	cb2.SetPos(180, 250)
	cb2.SetSize(120, 24)
	cb2.SetFont(engine.Font())
	engine.AddWidget(cb2)

	cb3 := widget.NewCheckbox("Indeterminate")
	cb3.SetPos(320, 250)
	cb3.SetSize(140, 24)
	cb3.SetFont(engine.Font())
	cb3.SetIndeterminate(true)
	engine.AddWidget(cb3)

	cb4 := widget.NewCheckbox("Disabled")
	cb4.SetPos(480, 250)
	cb4.SetSize(120, 24)
	cb4.SetFont(engine.Font())
	cb4.SetEnabled(false)
	engine.AddWidget(cb4)

	// ========== Section 4: Radio ==========
	radioLabel := widget.NewText("Radio Buttons")
	radioLabel.SetPos(48, 295)
	radioLabel.SetSize(200, 24)
	radioLabel.SetFont(engine.Font())
	radioLabel.SetColor(tokens.Global.ColorText)
	engine.AddWidget(radioLabel)

	radio1 := widget.NewRadio("Selected")
	radio1.SetPos(48, 325)
	radio1.SetSize(120, 24)
	radio1.SetFont(engine.Font())
	radio1.SetChecked(true)
	engine.AddWidget(radio1)

	radio2 := widget.NewRadio("Unselected")
	radio2.SetPos(180, 325)
	radio2.SetSize(120, 24)
	radio2.SetFont(engine.Font())
	engine.AddWidget(radio2)

	radio3 := widget.NewRadio("Disabled")
	radio3.SetPos(320, 325)
	radio3.SetSize(120, 24)
	radio3.SetFont(engine.Font())
	radio3.SetEnabled(false)
	engine.AddWidget(radio3)

	// ========== Section 5: Switch ==========
	switchLabel := widget.NewText("Switches")
	switchLabel.SetPos(48, 370)
	switchLabel.SetSize(200, 24)
	switchLabel.SetFont(engine.Font())
	switchLabel.SetColor(tokens.Global.ColorText)
	engine.AddWidget(switchLabel)

	sw1 := widget.NewSwitch()
	sw1.SetPos(48, 400)
	sw1.SetSize(44, 22)
	sw1.SetChecked(true)
	engine.AddWidget(sw1)

	sw2 := widget.NewSwitch()
	sw2.SetPos(110, 400)
	sw2.SetSize(44, 22)
	engine.AddWidget(sw2)

	sw3 := widget.NewSwitch()
	sw3.SetPos(172, 400)
	sw3.SetSize(44, 22)
	sw3.SetEnabled(false)
	engine.AddWidget(sw3)

	// ========== Section 6: Tags ==========
	tagLabel := widget.NewText("Tags")
	tagLabel.SetPos(48, 445)
	tagLabel.SetSize(200, 24)
	tagLabel.SetFont(engine.Font())
	tagLabel.SetColor(tokens.Global.ColorText)
	engine.AddWidget(tagLabel)

	tagBlue := widget.NewTag("Blue")
	tagBlue.SetPos(48, 475)
	tagBlue.SetSize(60, 24)
	tagBlue.SetFont(engine.Font())
	tagBlue.SetColor(widget.TagBlue)
	engine.AddWidget(tagBlue)

	tagGreen := widget.NewTag("Success")
	tagGreen.SetPos(120, 475)
	tagGreen.SetSize(70, 24)
	tagGreen.SetFont(engine.Font())
	tagGreen.SetColor(widget.TagGreen)
	engine.AddWidget(tagGreen)

	tagRed := widget.NewTag("Error")
	tagRed.SetPos(202, 475)
	tagRed.SetSize(60, 24)
	tagRed.SetFont(engine.Font())
	tagRed.SetColor(widget.TagRed)
	engine.AddWidget(tagRed)

	tagOrange := widget.NewTag("Warning")
	tagOrange.SetPos(274, 475)
	tagOrange.SetSize(76, 24)
	tagOrange.SetFont(engine.Font())
	tagOrange.SetColor(widget.TagOrange)
	engine.AddWidget(tagOrange)

	tagClosable := widget.NewTag("Closable")
	tagClosable.SetPos(362, 475)
	tagClosable.SetSize(80, 24)
	tagClosable.SetFont(engine.Font())
	tagClosable.SetColor(widget.TagBlue)
	tagClosable.SetClosable(true)
	engine.AddWidget(tagClosable)

	// ========== Section 7: Focused Button (showing focus ring) ==========
	focusLabel := widget.NewText("Focus State (click button to see focus ring)")
	focusLabel.SetPos(48, 520)
	focusLabel.SetSize(300, 24)
	focusLabel.SetFont(engine.Font())
	focusLabel.SetColor(tokens.Global.ColorText)
	engine.AddWidget(focusLabel)

	btnFocused := widget.NewButton("Click to Focus")
	btnFocused.SetKind(widget.ButtonPrimary)
	btnFocused.SetPos(48, 550)
	btnFocused.SetSize(140, 36)
	btnFocused.SetFont(engine.Font())
	engine.AddWidget(btnFocused)

	// ========== Section 8: Status indicators ==========
	statusLabel := widget.NewText("Status Colors")
	statusLabel.SetPos(250, 520)
	statusLabel.SetSize(200, 24)
	statusLabel.SetFont(engine.Font())
	statusLabel.SetColor(tokens.Global.ColorText)
	engine.AddWidget(statusLabel)

	// Success tag
	tagSuccess := widget.NewTag("✓ Success")
	tagSuccess.SetPos(250, 550)
	tagSuccess.SetSize(80, 24)
	tagSuccess.SetFont(engine.Font())
	tagSuccess.SetColor(widget.TagGreen)
	engine.AddWidget(tagSuccess)

	// Error tag
	tagError := widget.NewTag("✗ Error")
	tagError.SetPos(342, 550)
	tagError.SetSize(70, 24)
	tagError.SetFont(engine.Font())
	tagError.SetColor(widget.TagRed)
	engine.AddWidget(tagError)

	// Warning tag
	tagWarning := widget.NewTag("⚠ Warning")
	tagWarning.SetPos(424, 550)
	tagWarning.SetSize(86, 24)
	tagWarning.SetFont(engine.Font())
	tagWarning.SetColor(widget.TagOrange)
	engine.AddWidget(tagWarning)

	// Info tag
	tagInfo := widget.NewTag("ℹ Info")
	tagInfo.SetPos(522, 550)
	tagInfo.SetSize(60, 24)
	tagInfo.SetFont(engine.Font())
	tagInfo.SetColor(widget.TagBlue)
	engine.AddWidget(tagInfo)

	fmt.Println("✓ UI ready - all controls rendered")
}
