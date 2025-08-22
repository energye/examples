package main

import (
	"fmt"
	"github.com/energye/cef/cef"
	cefTypes "github.com/energye/cef/types"
	"github.com/energye/examples/cef/application"
	"github.com/energye/examples/cef/custombrowser/window"
	_ "github.com/energye/examples/syso"
	"github.com/energye/lcl/api"
	"github.com/energye/lcl/api/libname"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/tool"
	"os"
	"path/filepath"
)

func init() {
	var name string
	if tool.IsWindows() {
		name = "liblcl.dll"
	} else if tool.IsLinux() {
		name = "liblcl.so"
	}
	if name != "" {
		// 当前目录
		liblcl := filepath.Join(wd, name)
		if tool.IsExist(liblcl) {
			libname.LibName = liblcl
			return
		}
		// 测试编译输出目录
		if tool.IsWindows() {
			liblcl = filepath.Join("E:\\SWT\\gopath\\src\\github.com\\energye\\workspace\\gen\\gout", name)
		} else if tool.IsLinux() {
			liblcl = filepath.Join("/home/yanghy/app/gopath/src/github.com/energye/workspace/gen/gout", name)
		}
		if tool.IsExist(liblcl) {
			libname.LibName = liblcl
			return
		}
	}

}

var (
	wd, _            = os.Getwd()
	cacheRoot        = filepath.Join(wd, "EnergyCache")         // 浏览器缓存目录
	siteResourceRoot = filepath.Join(cacheRoot, "SiteResource") // 网站资源缓存目录
)

func main() {
	window.CacheRoot = cacheRoot
	window.SiteResource = siteResourceRoot
	//全局初始化 每个应用都必须调用的
	cef.Init(nil, nil)
	if tool.IsDarwin() {
		cef.AddCrDelegate()
	}

	app := application.NewApplication()
	app.SetLocale("zh-CN")
	app.SetRootCache(cacheRoot)
	app.SetCache(cacheRoot)
	//app.SetDeleteCache(true)

	if tool.IsDarwin() {
		app.InitLibLocationFromArgs()
		// MacOS不需要设置CEF框架目录，它是一个固定的目录结构
		app.SetUseMockKeyChain(true)
		app.SetExternalMessagePump(true)
		app.SetMultiThreadedMessageLoop(false)
		if app.ProcessType() == cefTypes.PtBrowser {
			scheduler := cef.NewWorkScheduler(nil)
			cef.SetGlobalCEFWorkSchedule(scheduler)
			app.SetOnScheduleMessagePumpWork(func(delayMs int64) {
				scheduler.ScheduleMessagePumpWork(delayMs)
			})
		} else {
			startSub := app.StartSubProcess()
			fmt.Println("startSub:", startSub)
			return
		}
	} else if tool.IsWindows() {
		// win32 使用 lcl 窗口
		app.SetExternalMessagePump(false)
		app.SetMultiThreadedMessageLoop(true)
	} else if tool.IsLinux() {
		if api.Widget().IsGTK2() {
			// gtk2 使用 lcl 窗口
			app.SetExternalMessagePump(false)
			app.SetMultiThreadedMessageLoop(true)
		} else if api.Widget().IsGTK3() {
			// gtk3 使用 vf 窗口
			println("当前 demo 为 CEF LCL GTK2, EXIT.")
			os.Exit(1)
		}
		// 这是一个解决“GPU不可用错误”问题的方法 linux
		// https://bitbucket.org/chromiumembedded/cef/issues/2964/gpu-is-not-usable-error-during-cef
		app.SetDisableZygote(true)
	}

	app.SetOnAlreadyRunningAppRelaunch(func(commandLine cef.ICefCommandLine, currentDirectory string, result *bool) {
		*result = true
	})
	// 主进程启动
	mainStart := app.StartMainProcess()
	if mainStart {
		CEFINfo(app)
		// 结束应用后释放资源
		api.SetReleaseCallback(func() {
			fmt.Println("Release")
			if tool.IsLinux() {
				api.WidgetSetFinalization()
			}
		})
		// LCL窗口
		lcl.Application.Initialize()
		lcl.Application.SetMainFormOnTaskBar(true)
		lcl.Application.NewForms(&window.BW)
		//lcl.Application.NewForms(&window.BW, &window.CW)
		lcl.Application.Run()
	}
}

func CEFINfo(app cef.ICefApplication) {
	// 输出版本信息
	println("ChromeVersion:", app.ChromeVersion())
	println("CefVersion:", app.LibCefVersion())
}
