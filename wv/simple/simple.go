package main

import (
	"embed"
	"fmt"
	"github.com/energye/assetserve"
	"github.com/energye/energy/v3/ipc"
	"github.com/energye/energy/v3/ipc/callback"
	"github.com/energye/energy/v3/wv"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
)

//go:embed resources
var resources embed.FS

func main() {
	wv.Init(nil, nil)
	app := wv.NewApplication()
	app.SetOptions(wv.Options{
		Caption: "energy - webview2",
		//DefaultURL: "https://www.baidu.com",
		DefaultURL: "http://localhost:22022",
		//DisableContextMenu: true,
		//DisableDevTools: true,
	})
	app.SetOnWindowCreate(func(window wv.IBrowserWindow) {
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
	})

	ipc.On("test-ipc", func(context callback.IContext) {
		fmt.Println("context:", context.Data())
		context.Result("返回", context.Data())
		ipc.Emit("ipcOnName", "数据")
		ipc.Emit("ipcOnName", "数据-带有返回回调函数", func(context callback.IContext) {
			fmt.Println("ipcOnName data:", context.Data())
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
	server.LocalAssets = "D:\\gopath\\src\\workspace\\examples\\wv\\simple\\resources"
	go server.StartHttpServer()
}
