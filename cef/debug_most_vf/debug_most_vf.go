package main

import (
	"fmt"
	"github.com/energye/cef/cef"
	cefTypes "github.com/energye/cef/types"
	"github.com/energye/examples/cef/debug_most/application"
	. "github.com/energye/examples/syso"
	"github.com/energye/lcl/api/exception"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/tool"
)

func init() {
	TestLoadLibPath()
}

func main() {
	//全局初始化 每个应用都必须调用的
	cef.Init(nil, nil)
	exception.SetOnException(func(exception int32, message string) {
		fmt.Println("[ERROR] exception:", exception, "message:", message)
	})
	app := application.NewApplication()
	// CEF message loop
	app.SetExternalMessagePump(false)
	app.SetMultiThreadedMessageLoop(false)
	if tool.IsDarwin() {
		app.SetUseMockKeyChain(true)
		app.InitLibLocationFromArgs()
		// MacOS使用扩展消息泵
		cef.AddCrDelegate()
		scheduler := cef.NewWorkScheduler(nil)
		cef.SetGlobalCEFWorkSchedule(scheduler)
		app.SetOnScheduleMessagePumpWork(nil)
		if app.ProcessType() != cefTypes.PtBrowser {
			// MacOS 多进程时，需要调用StartSubProcess来启动子进程
			subStart := app.StartSubProcess()
			fmt.Println("subStart:", subStart, app.ProcessType())
			return
		}
	}
	if tool.IsLinux() {
		// 这是一个解决“GPU不可用错误”问题的方法 linux
		// https://bitbucket.org/chromiumembedded/cef/issues/2964/gpu-is-not-usable-error-during-cef
		app.SetDisableZygote(true)
	}
	app.SetOnContextInitialized(func() {
		fmt.Println("SetOnContextInitialized")
		component := lcl.NewComponent(nil)
		chromium := cef.NewChromium(component)
		windowComponent := cef.NewWindowComponent(component)
		viewComponent := cef.NewBrowserViewComponent(component)
		url := "https://gitee.com/energye/energy"
		windowComponent.SetOnWindowCreated(func(sender lcl.IObject, window cef.ICefWindow) {
			ok := chromium.CreateBrowserWithStringBrowserViewComponentRequestContextDictionaryValue(url, viewComponent, nil, nil)
			fmt.Println("SetOnWindowCreated CreateBrowserByBrowserViewComponent:", true)
			if ok {
				windowComponent.AddChildView(viewComponent.BrowserView())
				viewComponent.RequestFocus()
				windowComponent.Show()
			}
		})
		chromium.SetOnBeforeClose(func(sender lcl.IObject, browser cef.ICefBrowser) {
			app.QuitMessageLoop()
		})
		windowComponent.CreateTopLevelWindow()
	})
	mainStart := app.StartMainProcess()
	fmt.Println("mainStart:", mainStart, app.ProcessType())
	if mainStart {
		// 结束应用后释放资源
		app.RunMessageLoop()
		fmt.Println("app free")
	}
	///usr/share/lazarus/3.2.0/lcl/graphics.pp
}
