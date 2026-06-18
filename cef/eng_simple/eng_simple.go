package main

import (
	"embed"
	"fmt"
	cef2 "github.com/energye/cef/cef"
	"github.com/energye/energy/v3/application"
	"github.com/energye/energy/v3/cef"
	"github.com/energye/energy/v3/core"
	"github.com/energye/energy/v3/logger"
	"github.com/energye/energy/v3/window"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
	"strconv"
)

type TForm struct {
	window.TWindow
	Browser cef.IBrowser
}

var Form TForm

//go:embed resources
var resources embed.FS

func main() {
	logger.L().SetLevel(logger.DebugLevel)
	app := cef.Init()
	app.SetOptions(application.Options{
		//Frameless:         true,
		//WindowTransparent: true,
		//WebviewTransparent: true,
		//BackgroundColor:    colors.NewARGB(0, 0, 0, 0),
		AutoPopupWindow: true,
	})
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

	m.SetCaption("ENERGY - CEF Simple 测试示例 " + strconv.Itoa(cef2.CEFVersion))
	m.SetPosition(types.PoScreenCenter)
	m.SetWidth(1000)
	m.SetHeight(700)

	m.Browser = cef.NewBrowser(m)
	m.Browser.SetAlign(types.AlClient)
	m.Browser.SetParent(m)
	m.Browser.SetWindow(m)
	//m.Browser.Chromium().SetDefaultUrl("https://energye.gitee.io")
	//m.Browser.Chromium().SetDefaultUrl("fs://energy/index-home.html")
	m.Browser.Chromium().SetDefaultUrl("fs://energy/index-ipc.html")

	m.Browser.SetOnResourceRequest(func(url, path, method string, header map[string]string) (resource string, ok bool) {
		fmt.Println("Browser.SetOnResourceRequest:", url, path, method, header)
		return
	})

	m.Browser.SetOnPopupWindow(func(targetURL string) bool {
		return false
	})

	m.Browser.SetOnLoadChange(func(url, title string, load core.TLoadChange) {
		fmt.Println("Browser.SetOnLoadChange:", url, title, load)
	})

	m.TWindow.FormCreate(sender)
}

func (m *TForm) OnShow(sender lcl.IObject) {
	println("OnShow")
	m.TWindow.OnShow(sender)
}
