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
	wv2 "github.com/energye/wv/windows"
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
	app.SetOnWindowCreate(func(window wv.IBrowserWindow) {
		window.WorkAreaCenter()
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
		window.SetOnNewWindowRequestedEvent(func(sender wv2.IObject, webview wv2.ICoreWebView2, args wv2.ICoreWebView2NewWindowRequestedEventArgs, callback *wv.NewWindowCallback) {
			//callback.SetHandled(true)
			newWindow := callback.NewWindow(wv.Options{
				Frameless: true,
			})
			newWindow.WorkAreaCenter()
			newWindow.SetOnClose(func(sender lcl.IObject, action *types.TCloseAction) {
				fmt.Println("new window close BrowserId:", newWindow.BrowserId(), "action:", *action)
				*action = types.CaFree
			})
		})
		rand.Seed(time.Now().UnixNano())
		var newBrowserWindow = wv.NewBrowserWindow(wv.Options{
			//DefaultURL: "https://www.baidu.com",
			DefaultURL: "http://localhost:22222/index.html",
			Caption:    "newBrowserWindow",
			Frameless:  true,
		})
		var subWindow = &SubForm{}
		subWindow.TForm = *(lcl.NewForm(nil).(*lcl.TForm))
		subWindow.SetBounds(rand.Int31n(300), rand.Int31n(300), 400, 200)
		subWindow.SetCaption("sub window")
		subWindow.SetOnShow(func(sender lcl.IObject) {
			fmt.Println("sub window show")
		})
		subWindow.SetShowInTaskBar(types.StAlways)

		btn := lcl.NewButton(window)
		btn.SetParent(window)
		btn.SetCaption("原生按钮")
		btn.SetBounds(rand.Int31n(70), rand.Int31n(70), 100, 35)
		btn.SetOnClick(func(sender lcl.IObject) {
			fmt.Println("SetOnClick")
			subWindow.SetBounds(rand.Int31n(300), rand.Int31n(300), 400, 200)
			if !newBrowserWindow.IsClosing() {
				newBrowserWindow.Show()
			}
			subWindow.Show()
		})
		//cs := window.Constraints()
		//cs.SetMinWidth(100)
		//cs.SetMinHeight(100)
		//cs.SetMaxWidth(800)
		//cs.SetMaxHeight(600)

	})
	app.SetOnWindowAfterCreate(func(window wv.IBrowserWindow) {
		fmt.Println("SetOnWindowAfterCreate")
	})

	ipc.On("test-ipc", func(context callback.IContext) {
		fmt.Println("BrowserId:", context.BrowserId(), "context:", context.Data())
		context.Result("返回", context.Data())
		ipc.EmitOptions(&ipc.OptionsEvent{
			BrowserId: context.BrowserId(),
			Name:      "ipcOnName",
			Data:      "数据",
			Callback: func(context callback.IContext) {
				fmt.Println("options ipcOnName data:", context.Data())
			},
		})
		ipc.Emit("ipcOnName", "数据-带有返回回调函数", func(context callback.IContext) {
			fmt.Println("ipcOnName data:", context.Data())
		})
		fmt.Println("test-ipc end")
	})
	ipc.On("CloseWindow", func(context callback.IContext) {
		if window := wv.GetBrowserWindow(context.BrowserId()); window != nil {
			window.Close()
		}
	})
	ipc.On("Restore", func(context callback.IContext) {
		if window := wv.GetBrowserWindow(context.BrowserId()); window != nil {
			window.Restore()
		}
	})
	ipc.On("Minimize", func(context callback.IContext) {
		if window := wv.GetBrowserWindow(context.BrowserId()); window != nil {
			window.Minimize()
		}
	})
	ipc.On("Maximize", func(context callback.IContext) {
		if window := wv.GetBrowserWindow(context.BrowserId()); window != nil {
			window.Maximize()
		}
	})
	ipc.On("FullScreen", func(context callback.IContext) {
		if window := wv.GetBrowserWindow(context.BrowserId()); window != nil {
			if window.IsFullScreen() {
				window.ExitFullScreen()
			} else {
				window.FullScreen()
			}
		}
	})
	//ipc.RemoveOn("test-name")
	startAssetsServer()
	app.Run()
}

type SubForm struct {
	lcl.TForm
}

func startAssetsServer() {
	server := assetserve.NewAssetsHttpServer()
	server.PORT = 22222               //服务端口号
	server.AssetsFSName = "resources" //必须设置目录名和资源文件夹同名
	//server.Assets = resources
	server.LocalAssets = filepath.Join(exec.CurrentDir, "wv", "simple", "resources")
	go server.StartHttpServer()
}
