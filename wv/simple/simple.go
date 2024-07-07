package main

import (
	"fmt"
	"github.com/energye/energy/v3/wv"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
)

func main() {
	wv.Init(nil, nil)
	app := wv.NewApplication()
	app.SetOptions(wv.Options{
		DefaultURL: "https://www.baidu.com",
		//DisableContextMenu: true,
		//DisableDevTools: true,
	})
	app.SetOnWindowCreate(func(window wv.IBrowserWindow) {
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

	app.Run()
}
