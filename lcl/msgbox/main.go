package main

import (
	_ "github.com/energye/examples/syso"
	"github.com/energye/lcl/inits"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/pkgs/win"
	"github.com/energye/lcl/types"
)

func main() {
	inits.Init(nil, nil)
	lcl.Application.Initialize()
	lcl.Application.SetMainFormOnTaskBar(true)

	mainForm := lcl.Application.CreateForm()
	mainForm.SetCaption("Hello")
	mainForm.SetPosition(types.PoScreenCenter)
	mainForm.EnabledMaximize(false)
	mainForm.SetWidth(300)
	mainForm.SetHeight(200)

	lcl.ShowMessage("消息")
	if lcl.MessageDlg("消息", types.MtConfirmation, types.MbYes, types.MbNo) == types.MrYes {
		lcl.ShowMessage("你点击了“是")
	}
	if lcl.Application.MessageBox("消息", "标题", win.MB_OKCANCEL+win.MB_ICONINFORMATION) == types.IdOK {
		lcl.ShowMessage("你点击了“是")
	}

	lcl.Application.Run()
}
