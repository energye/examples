package main

import (
	"github.com/energye/energy/v3/wv"
	"github.com/energye/lcl/lcl"
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
	})

	app.Run()
}
