package main

import (
	"github.com/energye/energy/v3/wv"
	"github.com/energye/lcl/lcl"
)

func main() {
	app := wv.NewApplication()
	app.SetOptions(wv.Options{})
	app.SetOnWindowCreate(func(window wv.IBrowserWindow) {
		window.SetOnAfterCreated(func(sender lcl.IObject) {

		})
	})

	app.Run()
}
