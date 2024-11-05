package main

import (
	"embed"
	"fmt"
	"github.com/energye/energy/v3/wv"
	_ "github.com/energye/examples/syso"
	"github.com/energye/lcl/api/exception"
	"github.com/energye/lcl/rtl/version"
	"github.com/energye/lcl/tools/exec"
	"net/http"
	_ "net/http/pprof"
	"path/filepath"
)

//go:embed resources
var resources embed.FS

func main() {
	go func() {
		http.ListenAndServe(":8080", nil)
	}()

	wv.Init(nil, nil)
	exception.SetOnException(func(funcName, message string) {
		fmt.Println("error funcName:", funcName, "message:", message)
	})
	fmt.Println("version:", version.OSVersion.ToString())
	app := wv.NewApplication()
	app.SetUserDataFolder(filepath.Join(exec.CurrentDir, "EnergyCache"))
	icon, _ := resources.ReadFile("resources/icon.ico")
	app.SetOptions(wv.Options{
		//Frameless:  true,
		Caption:    "energy - webview2",
		DefaultURL: "fs://energy/index.html",
		Windows: wv.Windows{
			ICON: icon,
		},
	})
	app.SetLocalLoad(wv.LocalLoad{
		Scheme:     "fs",
		Domain:     "energy",
		ResRootDir: "resources",
		FS:         resources,
	})

	app.SetOnWindowCreate(func(window wv.IBrowserWindow) {
		window.WorkAreaCenter()
		fmt.Println("SetOnWindowCreate")
	})
	app.SetOnWindowAfterCreate(func(window wv.IBrowserWindow) {
		fmt.Println("SetOnWindowAfterCreate")
	})

	app.Run()
	fmt.Println("run end")
}
