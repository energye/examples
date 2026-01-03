package main

import (
	"github.com/energye/energy/v3/wv"
	_ "github.com/energye/examples/syso"
	"github.com/energye/examples/wv/windows/simple/app"
	"github.com/energye/lcl/lcl"
)

// StartWebview 启动Webview应用程序
// 该函数初始化Webview并创建一个新的Webview应用实例，然后启动该应用
func StartWebview() {
	wvApp := wv.NewApplication()
	wvApp.Start()
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
}
