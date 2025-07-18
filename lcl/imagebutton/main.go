package main

import (
	. "github.com/energye/examples/syso"
	"github.com/energye/lcl/api"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
)

func init() {
	TestLoadLibPath()

	Chdir("lcl/imagebutton")
}

func main() {
	lcl.Init(nil, nil)

	lcl.Application.Initialize()
	lcl.Application.SetMainFormOnTaskBar(true)

	var mainForm lcl.TEngForm
	lcl.Application.NewForm(&mainForm)
	mainForm.SetCaption("Hello")
	mainForm.SetPosition(types.PoScreenCenter)
	mainForm.SetDoubleBuffered(true) // 最好开启，以免闪烁
	mainForm.EnabledMaximize(false)
	mainForm.SetWidth(300)
	mainForm.SetHeight(400)

	mainForm.SetShowHint(true)

	btnClose := lcl.NewImageButton(&mainForm)
	btnClose.SetParent(&mainForm)

	btnClose.SetImageCount(4)
	btnClose.SetAutoSize(true)
	btnClose.Picture().LoadFromFile("btn_close.png")
	btnClose.SetLeft(mainForm.ClientWidth() - btnClose.Width() - 3)
	btnClose.SetHint("关闭")

	btnClose.SetOnClick(func(lcl.IObject) {
		api.ShowMessage("close")
	})

	btnMax := lcl.NewImageButton(&mainForm)
	btnMax.SetParent(&mainForm)
	btnMax.SetImageCount(4)
	btnMax.SetAutoSize(true)
	btnMax.Picture().LoadFromFile("btn_max.png")
	btnMax.SetLeft(btnClose.Left() - btnMax.Width())
	btnMax.SetHint("最大化")

	btnMin := lcl.NewImageButton(&mainForm)
	btnMin.SetParent(&mainForm)
	btnMin.SetImageCount(4)
	btnMin.SetAutoSize(true)
	btnMin.Picture().LoadFromFile("btn_min.png")
	btnMin.SetLeft(btnMax.Left() - btnMin.Width())
	btnMin.SetHint("最小化")

	btnSkin := lcl.NewImageButton(&mainForm)
	btnSkin.SetParent(&mainForm)
	btnSkin.SetImageCount(3)
	btnSkin.SetAutoSize(true)
	btnSkin.Picture().LoadFromFile("btn_skin.png")
	btnSkin.SetLeft(btnMin.Left() - btnSkin.Width())
	btnSkin.SetHint("皮肤")

	btnScan := lcl.NewImageButton(&mainForm)
	btnScan.SetParent(&mainForm)
	btnScan.SetImageCount(3)
	btnScan.SetAutoSize(true)
	btnScan.Picture().LoadFromFile("btn_scan.png")
	btnScan.SetTop((mainForm.ClientHeight() - btnScan.Height()) / 2)
	btnScan.SetLeft((mainForm.ClientWidth() - btnScan.Width()) / 2)
	btnScan.SetHint("全盘扫描")

	lcl.Application.Run()
}
