package main

import (
	"embed"
	"fmt"
	"github.com/energye/energy/v3/application"
	"github.com/energye/energy/v3/wv"
	_ "github.com/energye/examples/syso"
	"github.com/energye/examples/wv/windows/local_load/app"
	"github.com/energye/lcl/lcl"
)

//go:embed resources
var resources embed.FS

// StartWebview 启动Webview应用程序
// 该函数初始化Webview并创建一个新的Webview应用实例，然后启动该应用
func StartWebview() *wv.Application {
	wv.Init()
	wvApp := wv.NewWebviewApplication()
	icon, _ := resources.ReadFile("resources/icon.ico")
	wvApp.SetOptions(application.Options{
		Frameless:  true,
		Caption:    "energy - webview2",
		DefaultURL: "fs://energy/index.html",
		Windows: application.Windows{
			ICON: icon,
		},
	})
	wvApp.SetLocalLoad(application.LocalLoad{
		Scheme:     "fs",
		Domain:     "energy",
		ResRootDir: "resources",
		FS:         resources,
	})
	wvApp.Start()
	return wvApp
}

func main() {
	lcl.Init(nil, nil)

	StartWebview()

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
