package main

import (
	. "github.com/energye/examples/syso"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
	"github.com/energye/lcl/types/colors"
)

type TMainForm struct {
	lcl.TEngForm
	Button1  lcl.IXButton
	richEdit lcl.IRichEdit
}

var mainForm TMainForm

func init() {
	TestLoadLibPath()
}
func main() {
	lcl.Init(nil, nil)
	lcl.Application.Initialize()
	lcl.Application.SetMainFormOnTaskBar(true)

	lcl.Application.NewForm(&mainForm)

	lcl.Application.Run()
}

func (mainForm *TMainForm) FormCreate(sender lcl.IObject) {
	mainForm.SetCaption("Hello")
	mainForm.SetPosition(types.PoScreenCenter)
	mainForm.EnabledMaximize(false)
	mainForm.SetWidth(600)
	mainForm.SetHeight(400)
	mainForm.initMainMenu()

	mainForm.richEdit = lcl.NewRichEdit(mainForm)
	mainForm.richEdit.SetParent(mainForm)
	mainForm.richEdit.SetAlign(types.AlClient)
	mainForm.richEdit.Lines().Add("这是一段文字红色，粗体，斜體")
	mainForm.richEdit.SetSelStart(6)
	mainForm.richEdit.SetSelLength(2)
	mainForm.richEdit.SelAttributes().SetColor(colors.ClRed)

	mainForm.richEdit.SetSelStart(9)
	mainForm.richEdit.SetSelLength(2)

	mainForm.richEdit.SelAttributes().SetStyle(types.NewSet(types.FsBold))

	mainForm.richEdit.SetSelStart(12)
	mainForm.richEdit.SetSelLength(2)

	mainForm.richEdit.SelAttributes().SetStyle(types.NewSet(types.FsItalic))

	mainForm.richEdit.SetSelStart(15)
	mainForm.initRichEditPopupMenu()

	tlbar := lcl.NewToolBar(mainForm)
	tlbar.SetParent(mainForm)
	stabar := lcl.NewStatusBar(mainForm)
	stabar.SetParent(mainForm)
}

func (mainForm *TMainForm) initMainMenu() {
	mainMenu := lcl.NewMainMenu(mainForm)

	item := lcl.NewMenuItem(mainForm)
	item.SetCaption("&File")
	mainMenu.Items().Add(item)

	item = lcl.NewMenuItem(mainForm)
	item.SetCaption("&Help")
	mainMenu.Items().Add(item)
}

func (mainForm *TMainForm) initRichEditPopupMenu() {
	pm := lcl.NewPopupMenu(mainForm)
	item := lcl.NewMenuItem(mainForm)
	item.SetCaption("&Clear")
	pm.Items().Add(item)

	mainForm.richEdit.SetPopupMenu(pm)
}
