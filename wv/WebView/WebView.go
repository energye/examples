package main

import (
	"embed"
	"fmt"
	"github.com/energye/energy/v3/application"
	"github.com/energye/energy/v3/ipc"
	"github.com/energye/energy/v3/window"
	"github.com/energye/energy/v3/wv"
	"github.com/energye/lcl/api"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
	"runtime"
)

//go:embed resources/*
var resources embed.FS

type TMainForm struct {
	window.TWindow
	Webview1 wv.IWebview
}

var MainForm TMainForm

func main() {
	api.SetDebug(true)
	wvApp := wv.Init(nil, nil)

	wvApp.SetOptions(application.Options{
		DefaultURL: "app://custom/index.html",
		Caption:    "Energy WebView 完整示例",
		Width:      1200,
		Height:     800,
	})

	wvApp.SetLocalLoad(application.LocalLoad{
		Scheme:     "app",
		Domain:     "custom",
		ResRootDir: "resources",
		FS:         resources,
	})

	ipc.On("get-system-info", func(context ipc.IContext) {
		fmt.Println("获取系统信息")
		context.Result(map[string]interface{}{
			"platform":  runtime.GOOS,
			"arch":      runtime.GOARCH,
			"browserId": context.BrowserId(),
			"goVersion": runtime.Version(),
		})
	})

	ipc.On("calculate", func(context ipc.IContext) {
		data := context.Data().([]interface{})
		params := data[0].(map[string]interface{})
		a := params["a"].(float64)
		b := params["b"].(float64)
		operator := params["operator"].(string)

		var result float64
		switch operator {
		case "+":
			result = a + b
		case "-":
			result = a - b
		case "*":
			result = a * b
		case "/":
			if b != 0 {
				result = a / b
			} else {
				context.Result(map[string]interface{}{
					"error": "除数不能为零",
				})
				return
			}
		}

		fmt.Printf("计算: %.2f %s %.2f = %.2f\n", a, operator, b, result)
		context.Result(map[string]interface{}{
			"result": result,
		})
	})

	ipc.On("show-message", func(context ipc.IContext) {
		data := context.Data().([]interface{})
		message := data[0].(string)
		fmt.Println("收到消息:", message)
		context.Result(map[string]interface{}{
			"success": true,
			"message": "消息已接收: " + message,
		})
	})

	ipc.On("get-user-list", func(context ipc.IContext) {
		users := []map[string]interface{}{
			{"id": 1, "name": "张三", "age": 25, "email": "zhangsan@example.com"},
			{"id": 2, "name": "李四", "age": 30, "email": "lisi@example.com"},
			{"id": 3, "name": "王五", "age": 28, "email": "wangwu@example.com"},
		}
		context.Result(users)
	})

	wv.Run(&MainForm)
}

func (m *TMainForm) FormCreate(sender lcl.IObject) {
	m.InternalBeforeFormCreate()

	m.SetCaption("Energy WebView 完整示例")
	m.SetPosition(types.PoScreenCenter)
	m.SetWidth(1200)
	m.SetHeight(800)

	m.Webview1 = wv.NewWebview(m)
	m.Webview1.SetParent(m)
	m.Webview1.SetAlign(types.AlClient)
	m.Webview1.SetWindow(m)

	m.Webview1.SetOnLoadChange(func(url, title string, load wv.TLoadChange) {
		switch load {
		case wv.LcStart:
			fmt.Println("开始加载:", url)
		case wv.LcLoading:
			fmt.Println("加载中:", url)
		case wv.LcFinish:
			fmt.Println("加载完成:", title)
		}
	})

	m.Webview1.SetOnContextMenu(func(contextMenu *wv.TContextMenuItem) {
		contextMenu.Add("刷新页面", wv.CmkCommand)
		contextMenu.Add("", wv.CmkSeparator)
		contextMenu.Add("开发者工具", wv.CmkCommand)
	})

	m.TWindow.FormCreate(sender)
}

func (m *TMainForm) OnShow(sender lcl.IObject) {
	m.WorkAreaCenter()
	m.Webview1.CreateBrowser()
}
