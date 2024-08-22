package main

import (
	"embed"
	"fmt"
	"github.com/energye/energy/v3/wv"
	_ "github.com/energye/examples/syso"
	"github.com/energye/lcl/rtl/version"
)

//go:embed resources
var resources embed.FS

func main() {
	wv.Init(nil, nil)
	fmt.Println("version:", version.OSVersion.ToString())
	app := wv.NewApplication()
	icon, _ := resources.ReadFile("resources/icon.ico")
	app.SetOptions(wv.Options{
		Frameless:  true,
		Caption:    "energy - webview2",
		DefaultURL: "fs://energy/index.html",
		Windows: wv.Windows{
			ICON: icon,
		},
		LocalLoad: &wv.LocalLoad{
			Scheme:     "fs",
			Domain:     "energy",
			ResRootDir: "resources",
			FS:         resources,
		},
	})
	app.SetOnWindowCreate(func(window wv.IBrowserWindow) {
		window.WorkAreaCenter()

	})
	app.SetOnWindowAfterCreate(func(window wv.IBrowserWindow) {
		fmt.Println("SetOnWindowAfterCreate")
	})

	app.Run()
}
