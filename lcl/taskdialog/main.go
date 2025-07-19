package main

import (
	"fmt"
	. "github.com/energye/examples/syso"
	"github.com/energye/lcl/api"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
)

type TMainForm struct {
	lcl.TEngForm
	Btn1 lcl.IButton
}

var MainForm TMainForm

func init() {
	TestLoadLibPath()
}
func main() {
	lcl.Init(nil, nil)
	lcl.Application.Initialize()
	lcl.Application.SetMainFormOnTaskBar(true)
	lcl.Application.NewForm(&MainForm)
	lcl.Application.Run()
}

func (f *TMainForm) FormCreate(sender lcl.IObject) {
	f.ScreenCenter()
	f.SetCaption("taskDialog演示")

	f.Btn1 = lcl.NewButton(f)
	f.Btn1.SetParent(f)
	f.Btn1.SetCaption("TaskDialog")
	f.Btn1.SetLeft(10)
	f.Btn1.SetTop(10)
	f.Btn1.SetOnClick(f.OnBtn1Click)

}

func (f *TMainForm) OnFormDestroy(sender lcl.IObject) {

}

func (f *TMainForm) OnBtn1Click(sender lcl.IObject) {
	taskdlg := lcl.NewTaskDialog(f)
	defer taskdlg.Free()
	taskdlg.SetTitle("确认移除")
	taskdlg.SetCaption("询问")
	taskdlg.SetText("移除选择的项目？")
	taskdlg.SetExpandButtonCaption("展开按钮标题")
	taskdlg.SetExpandedText("展开的文本")

	taskdlg.SetFooterText("底部文本")

	rd := taskdlg.RadioButtons().AddToTaskDialogBaseButtonItem()
	rd.SetCaption("单选按钮1")
	rd = taskdlg.RadioButtons().AddToTaskDialogBaseButtonItem()
	rd.SetCaption("单选按钮2")
	rd = taskdlg.RadioButtons().AddToTaskDialogBaseButtonItem()
	rd.SetCaption("单选按钮3")

	taskdlg.SetCommonButtons(0) //rtl.Include(0, 0))
	btn := taskdlg.Buttons().AddToTaskDialogBaseButtonItem()
	btn.SetCaption("移除")
	btn.SetModalResult(types.MrYes)

	btn = taskdlg.Buttons().AddToTaskDialogBaseButtonItem()
	btn.SetCaption("保持")
	btn.SetModalResult(types.MrNo)

	taskdlg.SetMainIcon(types.TdiQuestion)
	if taskdlg.ExecuteToBool() {
		if taskdlg.ModalResult() == types.MrYes {
			api.ShowMessage("项目已移除")

			fmt.Println(taskdlg.RadioButton().Caption())
		}
	}
}
