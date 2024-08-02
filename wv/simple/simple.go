package main

import (
	"embed"
	"fmt"
	"github.com/energye/assetserve"
	"github.com/energye/energy/v3/ipc"
	"github.com/energye/energy/v3/ipc/callback"
	"github.com/energye/energy/v3/wv"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/pkgs/win"
	"github.com/energye/lcl/tools/exec"
	"github.com/energye/lcl/types"
	"github.com/energye/lcl/types/messages"
	wv2 "github.com/energye/wv/wv"
	"path/filepath"
)

//go:embed resources
var resources embed.FS

func main() {
	wv.Init(nil, nil)
	app := wv.NewApplication()
	app.SetOptions(wv.Options{
		Caption: "energy - webview2",
		//DefaultURL: "https://www.baidu.com",
		//DefaultURL: "https://ap2.baoleitech.com:2443/bl-viewer-v1/dicompat?hcode=3122401&puid=01J4183BSWGJCPSB60ZX5EETM9&accessToken=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjMiLCJuYW1lIjoiYWRtaW4iLCJlbWFpbCI6ImFkbWluQGhvdG1haWwuY29tIiwiZGlzcGxheU5hbWUiOiLnrqHnkIblkZgiLCJpYXQiOjE3MjIzMjc0MjMsImV4cCI6MTcyMjM0OTAyM30.RpjkzTQDVb0l6qDtky7tI_4gNOyJoA2kTB7PudDAdzM",
		DefaultURL: "http://localhost:22022/index.html",
		//DisableContextMenu: true,
		//DisableDevTools: true,
		Frameless: true,
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
		window.Browser().SetOnRenderCompMsg(func(sender wv2.IObject, message *types.TMessage, handled bool) {
			fmt.Println("SetOnRenderCompMsg")
		})
		btn := lcl.NewButton(window)
		btn.SetParent(window)
		btn.SetCaption("原生按钮")
		btn.SetOnClick(func(sender lcl.IObject) {
			fmt.Println("SetOnClick")
		})
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
	ipc.On("ShowTitleBar", func(context callback.IContext) {
		lcl.RunOnMainThreadSync(func() {
			if win.ReleaseCapture() {
				if mainWindow.WindowState() == types.WsNormal {
					win.PostMessage(mainWindow.Handle(), messages.WM_SYSCOMMAND, messages.SC_MAXIMIZE, 0)
				} else {
					win.SendMessage(mainWindow.Handle(), messages.WM_SYSCOMMAND, messages.SC_RESTORE, 0)
				}
			}
		})
	})
	//ipc.RemoveOn("test-name")
	startAssetsServer()
	app.Run()
}

func startAssetsServer() {
	server := assetserve.NewAssetsHttpServer()
	server.PORT = 22022               //服务端口号
	server.AssetsFSName = "resources" //必须设置目录名和资源文件夹同名
	//server.Assets = resources
	server.LocalAssets = filepath.Join(exec.CurrentDir, "wv", "simple", "resources")
	go server.StartHttpServer()
}
