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
	})
	app.SetOnWindowCreate(func(window wv.IBrowserWindow) {
		window.SetOnBrowserAfterCreated(func(sender lcl.IObject) {

		})
		window.SetOnShow(func(sender lcl.IObject) {

		})
		window.SetOnClose(func(sender lcl.IObject, action *types.TCloseAction) {
			fmt.Println("action:", *action)
		})
		window.SetOnCloseQuery(func(sender lcl.IObject, canClose *bool) {
			fmt.Println("canClose:", *canClose)
		})
	})

	app.Run()
}
