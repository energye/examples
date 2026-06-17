package main

import (
	"embed"
	cef2 "github.com/energye/cef/cef"
	"github.com/energye/energy/v3/application"
	"github.com/energye/energy/v3/cef"
	"github.com/energye/energy/v3/logger"
	"github.com/energye/energy/v3/window"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
)

type TForm struct {
	window.TWindow
	Browser *cef.TBrowser
}

var Form TForm

//go:embed resources
var resources embed.FS

func main() {
	logger.L().SetLevel(logger.DebugLevel)
	app := cef.Init()
	app.SetOnBeforeChildProcessLaunch(func(commandLine cef2.ICefCommandLine) {
		println("app.SetOnBeforeChildProcessLaunch")
	})
	app.SetLocalLoad(application.LocalLoad{
		Scheme:     "fs",
		Domain:     "energy",
		ResRootDir: "resources",
		FS:         resources,
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
	//m.Browser.Chromium().SetDefaultUrl("https://energye.gitee.io")
	m.Browser.Chromium().SetDefaultUrl("fs://energy/index-home.html")

	m.TWindow.FormCreate(sender)
}

func (m *TForm) OnShow(sender lcl.IObject) {
	println("OnShow")
	m.TWindow.OnShow(sender)
}
