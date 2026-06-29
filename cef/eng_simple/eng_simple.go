package main

import (
	"embed"
	"fmt"
	cef2 "github.com/energye/cef/cef"
	cefTypes "github.com/energye/cef/cef/types"
	"github.com/energye/energy/v3/application"
	"github.com/energye/energy/v3/cef"
	"github.com/energye/energy/v3/core"
	"github.com/energye/energy/v3/ipc"
	"github.com/energye/energy/v3/logger"
	"github.com/energye/energy/v3/window"
	"github.com/energye/examples/cef/eng_simple/app"
	"github.com/energye/lcl/api/libname"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
	"strconv"
	"time"
)

type TForm struct {
	window.TWindow
	Browser cef.IEmbeddedBrowser
}

var Form TForm

//go:embed resources
var resources embed.FS

func init() {
	libname.UseWS = "gtk3"
}

func main() {
	logger.L().SetLevel(logger.DebugLevel)
	cefApp := cef.Init()
	cefApp.SetLogSeverity(cefTypes.LOGSEVERITY_DEBUG)
	cefApp.SetOptions(application.Options{
		//Frameless: true,
		//WindowTransparent: true,
		//WebviewTransparent: true,
		//BackgroundColor:    colors.NewARGB(0, 0, 0, 0),
		//DefaultURL:"https://energye.gitee.io",
		//DefaultURL:"fs://energy/index-home.html",
		//DefaultURL:"fs://energy/index-ipc.html",
		DefaultURL: "fs://energy/index-drag.html",
		//DefaultURL:      "http://chrome.360.cn/html5_labs",
		AutoPopupWindow: true,
	})
	cefApp.SetOnBeforeChildProcessLaunch(func(commandLine cef2.ICefCommandLine) {
		println("app.SetOnBeforeChildProcessLaunch")
	})
	cefApp.SetLocalLoad(application.LocalLoad{
		Scheme:     "fs",
		Domain:     "energy",
		ResRootDir: "resources",
		FS:         resources,
	})

	ipc.BindEvent(&app.DemoBind{})
	ipc.BindEventPrefix("demo", &app.DemoBind{})
	ipc.On("test", func(context ipc.IContext) {
		fmt.Println("ipc-test:", context.BrowserId(), "data:", context.Data())
		context.Result("ResultData", 123, 888.99, true, time.Now().String())
		ipc.Emit("test", "测试数据")
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
	m.Browser = cef.NewEmbeddedBrowser(m)
	m.Browser.SetAlign(types.AlClient)
	m.Browser.SetParent(m)
	m.Browser.SetWindow(m)

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

	m.Browser.SetOnDragEnter(func(type_ core.TDragType, x, y int32) {
		fmt.Println("SetOnDragEnter --------------begin------------------", type_, x, y)
		ipc.Emit("drag-enter")
	})
	m.Browser.SetOnDragLeave(func() {
		fmt.Println("SetOnDragLeave", "--------------zzz------------------")
	})
	m.Browser.SetOnDragOver(func(data *core.TDragData, x, y int32) {
		da, err := strconv.Unquote("\"" + string(data.Data) + "\"")
		fmt.Println("SetOnDragOver --------------end------------------", x, y, da, err, data.Filenames)
		ipc.Emit("drag-over", da, data.Filenames)
	})
	m.Browser.SetOnContextMenu(func(contextMenu *core.TContextMenuItem) {
		//contextMenu.Clear()
		contextMenu.Add("", core.CmkSeparator)
		_, id := contextMenu.Add("测试1", core.CmkCommand)
		fmt.Println("测试1:", id)
		test2, id := contextMenu.Add("测试2", core.CmkSub)
		fmt.Println("测试2:", id)
		_, id = test2.Add("测试2-测试", core.CmkCommand)
		fmt.Println("测试2-测试:", id)
		_, id = test2.Add("测试3-测试", core.CmkCommand)
		fmt.Println("测试3-测试:", id)
		contextMenu.Add("测试3", core.CmkCommand)
	})
	m.Browser.SetOnContextMenuCommand(func(commandId int32, handle *bool) {
		fmt.Println("OnContextMenuCommand:", commandId)
		m.Browser.ExecuteScriptCallback("document.title", func(result string, err string) {
			fmt.Println("ExecuteScriptCallback:", result, err)
		})
	})
	m.SetOnThemeChange(func(isDark bool) {
		fmt.Println("SetOnThemeChange isDark:", isDark)
	})
	m.WorkAreaCenter()
	m.TWindow.FormCreate(sender)
}

func (m *TForm) OnShow(sender lcl.IObject) {
	println("OnShow")
	m.TWindow.OnShow(sender)
}
