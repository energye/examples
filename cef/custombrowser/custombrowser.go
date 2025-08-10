package main

import (
	"fmt"
	"github.com/energye/cef/cef"
	"github.com/energye/examples/cef/application"
	"github.com/energye/examples/cef/custombrowser/window"
	. "github.com/energye/examples/syso"
	"github.com/energye/lcl/api"
	"github.com/energye/lcl/api/exception"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/tool"
	"os"
	"path/filepath"
)

func init() {
	TestLoadLibPath()
}

var (
	wd, _     = os.Getwd()
	cacheRoot = filepath.Join(wd, "EnergyCache")
)

func main() {
	//全局初始化 每个应用都必须调用的
	cef.Init(nil, nil)
	exception.SetOnException(func(exception int32, message string) {
		fmt.Println("[ERROR] exception:", exception, "message:", message)
	})
	app := application.NewApplication()
	app.SetEnableGPU(true)
	app.SetLocale("zh-CN")
	if tool.IsWindows() {
		// win32 使用 lcl 窗口
		app.SetExternalMessagePump(false)
		app.SetMultiThreadedMessageLoop(true)
		app.SetRootCache(cacheRoot)
	}
	// 主进程启动
	mainStart := app.StartMainProcess()
	fmt.Println("mainStart:", mainStart, app.ProcessType())
	if mainStart {
		// 结束应用后释放资源
		api.SetReleaseCallback(func() {
			fmt.Println("Release")
			if tool.IsLinux() {
				api.WidgetSetFinalization()
			}
		})
		api.WidgetSetInitialization()
		// LCL窗口
		lcl.Application.Initialize()
		lcl.Application.SetMainFormOnTaskBar(true)
		lcl.Application.NewForm(&window.BW)
		lcl.Application.Run()
	}
}
