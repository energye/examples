package main

import (
	"embed"
	"fmt"
	"github.com/energye/energy/v3/application"
	"github.com/energye/energy/v3/ipc"
	"github.com/energye/energy/v3/ipc/callback"
	"github.com/energye/energy/v3/wv"
	_ "github.com/energye/examples/syso"
	"github.com/energye/examples/wv/windows/local_load/app"
	"github.com/energye/lcl/lcl"
	"os"
	"time"
)

//go:embed resources
var resources embed.FS

func main() {
	// linux webkit2 > gtk3
	os.Setenv("--ws", "gtk3")
	wvApp := wv.Init(nil, nil)
	wvApp.SetOptions(application.Options{
		//Frameless:  true,
		Caption:    "energy - webview2",
		DefaultURL: "fs://energy/index-2.html",
		Windows:    application.Windows{},
		Frameless:  true,
		//DisableContextMenu: true,
	})
	wvApp.SetLocalLoad(application.LocalLoad{
		Scheme:     "fs",
		Domain:     "energy",
		ResRootDir: "resources",
		FS:         resources,
	})
	wvApp.Start()

	ipc.On("test", func(context callback.IContext) {
		fmt.Println("ipc-test:", context.BrowserId(), "data:", context.Data())
		context.Result("ResultData", 123, 888.99, true, time.Now().String())
		ipc.Emit("test", "测试数据")
	})

	ipc.On("minimize", func(context callback.IContext) {
		fmt.Println("minimize")
		app.Form1Window.Minimize()
	})

	ipc.On("maximize", func(context callback.IContext) {
		fmt.Println("maximize")
		app.Form1Window.Maximize()
	})

	ipc.On("fullscreen", func(context callback.IContext) {
		fmt.Println("fullscreen")
		if app.Form1Window.IsFullScreen() {
			app.Form1Window.ExitFullScreen()
		} else {
			app.Form1Window.FullScreen()
		}
	})

	ipc.On("close", func(context callback.IContext) {
		fmt.Println("close")
		lcl.RunOnMainThreadAsync(func(id uint32) {
			app.Form1Window.Close()
		})
	})

	// 初始化应用程序实例
	lcl.Application.Initialize()
	// 配置应用程序设置，使主窗体在Windows任务栏上显示
	lcl.Application.SetMainFormOnTaskBar(true)
	// 启用自动缩放功能以支持高DPI显示器
	lcl.Application.SetScaled(true)
	// 创建所有窗体
	lcl.Application.NewForms(&app.Form1Window)
	// 启动应用程序消息循环
	lcl.Application.Run()
	fmt.Println("run end")
}
