package main

import (
	. "github.com/energye/examples/syso"
	"github.com/energye/lcl/api"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/pkgs/win"
	"github.com/energye/lcl/types"
)

func init() {
	TestLoadLibPath()
}
func main() {
	lcl.Init(nil, nil)
	lcl.Application.Initialize()
	lcl.Application.SetMainFormOnTaskBar(true)
	var mainForm lcl.TEngForm
	lcl.Application.NewForm(&mainForm)
	mainForm.SetCaption("Hello")
	mainForm.SetPosition(types.PoScreenCenter)
	mainForm.EnabledMaximize(false)
	mainForm.SetWidth(300)
	mainForm.SetHeight(200)

	api.ShowMessage("消息")
	if api.MessageDlg("消息", types.MtConfirmation, types.NewSet(types.MbYes), types.MbNo) == types.MrYes {
		api.ShowMessage("你点击了“是")
	}
	if lcl.Application.MessageBox(api.PasStr("消息"), api.PasStr("标题"), win.MB_OKCANCEL+win.MB_ICONINFORMATION) == types.IdOK {
		api.ShowMessage("你点击了“是")
	}

	lcl.Application.Run()
}
