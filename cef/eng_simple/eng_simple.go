package main

import (
	cef2 "github.com/energye/cef/cef"
	"github.com/energye/energy/v3/cef"
	"github.com/energye/energy/v3/window"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
)

type TForm struct {
	window.TWindow
	Browser *cef.TBrowser
}

var Form TForm

func main() {
	app := cef.Init()
	app.SetOnBeforeChildProcessLaunch(func(commandLine cef2.ICefCommandLine) {

	})

	cef.Run(&Form)
}

func (m *TForm) FormCreate(sender lcl.IObject) {
	println("FormCreate")
	m.InternalBeforeFormCreate()

	m.SetCaption("ENERGY - CEF Simple 测试示例")
	m.SetPosition(types.PoScreenCenter)
	m.SetWidth(1000)
	m.SetHeight(700)

	m.Browser = cef.NewBrowser(m)
	m.Browser.SetAlign(types.AlClient)
	m.Browser.SetParent(m)
	m.Browser.SetWindow(m)
	m.Browser.Chromium().SetDefaultUrl("https://energye.gitee.io")

	m.TWindow.FormCreate(sender)
}

func (m *TForm) OnShow(sender lcl.IObject) {
	println("OnShow")
	m.TWindow.OnShow(sender)
}
