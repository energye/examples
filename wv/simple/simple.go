package main

import (
	"embed"
	"fmt"
	"github.com/energye/assetserve"
	"github.com/energye/energy/v3/ipc"
	"github.com/energye/energy/v3/ipc/callback"
	"github.com/energye/energy/v3/wv"
	_ "github.com/energye/examples/syso"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/rtl/version"
	"github.com/energye/lcl/tools/exec"
	"github.com/energye/lcl/types"
	"math/rand"
	"path/filepath"
	"time"
)

//go:embed resources
var resources embed.FS

func main() {
	wv.Init(nil, nil)
	fmt.Println("version:", version.OSVersion.ToString())
	app := wv.NewApplication()
	icon, _ := resources.ReadFile("resources/icon.ico")
	app.SetOptions(wv.Options{
		Caption: "energy - webview2",
		//DefaultURL: "https://www.baidu.com",
		//DefaultURL: "https://ap2.baoleitech.com:2443/bl-viewer-v1/dicompat?hcode=3122401&puid=01J4183BSWGJCPSB60ZX5EETM9&accessToken=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjMiLCJuYW1lIjoiYWRtaW4iLCJlbWFpbCI6ImFkbWluQGhvdG1haWwuY29tIiwiZGlzcGxheU5hbWUiOiLnrqHnkIblkZgiLCJpYXQiOjE3MjIzMjc0MjMsImV4cCI6MTcyMjM0OTAyM30.RpjkzTQDVb0l6qDtky7tI_4gNOyJoA2kTB7PudDAdzM",
		DefaultURL: "http://localhost:22222/index.html",
		Windows: wv.Windows{
			ICON: icon,
		},
		//DisableContextMenu: true,
		//DisableDevTools: true,
		//Frameless: true,
		//DisableResize:   true,
		//DisableMinimize: true,
		//DisableMaximize: true,
		//DefaultWindowStatus: types.WsFullScreen,
		//MaxWidth:  1024,
		//MinHeight: 200,
		//DisableWebkitAppRegionDClk: true,
	})
	var mainWindow wv.IBrowserWindow
	app.SetOnWindowCreate(func(window wv.IBrowserWindow) {
		mainWindow = window
		window.ScreenCenter()
		window.SetOnBrowserAfterCreated(func(sender lcl.IObject) {
			fmt.Println("SetOnBrowserAfterCreated")
		})
		window.SetOnShow(func(sender lcl.IObject) {
			fmt.Println("SetOnShow")
		})
		window.SetOnClose(func(sender lcl.IObject, action *types.TCloseAction) {
			fmt.Println("SetOnClose action:", *action)
		})
		window.SetOnCloseQuery(func(sender lcl.IObject, canClose *bool) {
			fmt.Println("SetOnCloseQuery canClose:", *canClose)
		})
		var subWindow = &SubForm{}
		subWindow.IForm = lcl.NewForm(nil)
		subWindow.SetShowInTaskBar(types.StAlways)
		subWindow.SetBounds(rand.Int31n(300), rand.Int31n(300), 400, 200)
		subWindow.SetCaption("sub window")
		subWindow.SetOnShow(func(sender lcl.IObject) {
			fmt.Println("sub window show")
		})
		rand.Seed(time.Now().UnixNano())
		btn := lcl.NewButton(window)
		btn.SetParent(window)
		btn.SetCaption("原生按钮")
		btn.SetBounds(rand.Int31n(70), rand.Int31n(70), 100, 35)
		btn.SetOnClick(func(sender lcl.IObject) {
			fmt.Println("SetOnClick")
			subWindow.SetBounds(rand.Int31n(300), rand.Int31n(300), 400, 200)
			subWindow.Show()
			window.SetBorderStyleForFormBorderStyle(types.BsSizeable)
		})
		//cs := window.Constraints()
		//cs.SetMinWidth(100)
		//cs.SetMinHeight(100)
		//cs.SetMaxWidth(800)
		//cs.SetMaxHeight(600)

	})

	ipc.On("test-ipc", func(context callback.IContext) {
		fmt.Println("context:", context.Data())
		context.Result("返回", context.Data())
		ipc.Emit("ipcOnName", "数据")
		ipc.Emit("ipcOnName", "数据-带有返回回调函数", func(context callback.IContext) {
			fmt.Println("ipcOnName data:", context.Data())
		})
		fmt.Println("test-ipc end")
	})
	ipc.On("CloseWindow", func(context callback.IContext) {
		fmt.Println(mainWindow.WindowId())
		mainWindow.Close()
	})
	ipc.On("Restore", func(context callback.IContext) {
		mainWindow.Restore()
	})
	ipc.On("Minimize", func(context callback.IContext) {
		mainWindow.Minimize()
	})
	ipc.On("Maximize", func(context callback.IContext) {
		mainWindow.Maximize()
	})
	ipc.On("FullScreen", func(context callback.IContext) {
		if mainWindow.IsFullScreen() {
			mainWindow.ExitFullScreen()
		} else {
			mainWindow.FullScreen()
		}
	})
	//ipc.RemoveOn("test-name")
	startAssetsServer()
	app.Run()
}

type SubForm struct {
	lcl.IForm
}

func (m *SubForm) FormCreate(sender lcl.IObject) {
	m.SetBounds(100, 100, 300, 300)
	m.ScreenCenter()
	m.SetShowInTaskBar(types.StAlways)
}

func startAssetsServer() {
	server := assetserve.NewAssetsHttpServer()
	server.PORT = 22222               //服务端口号
	server.AssetsFSName = "resources" //必须设置目录名和资源文件夹同名
	//server.Assets = resources
	server.LocalAssets = filepath.Join(exec.CurrentDir, "wv", "simple", "resources")
	go server.StartHttpServer()
}
