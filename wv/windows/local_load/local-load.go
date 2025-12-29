package main

import (
	"embed"
	"fmt"
	"github.com/energye/energy/v3/application"
	"github.com/energye/energy/v3/wv"
	_ "github.com/energye/examples/syso"
	"github.com/energye/lcl/api/exception"
	"github.com/energye/lcl/rtl/version"
	"github.com/energye/lcl/tool/exec"
	"path/filepath"
)

//go:embed resources
var resources embed.FS

func main() {
	wv.Init(nil, nil)
	exception.SetOnException(func(exception int32, message string) {
		fmt.Println("error exception:", exception, "message:", message)
	})
	fmt.Println("version:", version.OSVersion.ToString())
	app := wv.NewApplication()
	app.SetUserDataFolder(filepath.Join(exec.AppDir(), "EnergyCache"))
	icon, _ := resources.ReadFile("resources/icon.ico")
	app.SetOptions(application.Options{
		//Frameless:  true,
		Caption:    "energy - webview2",
		DefaultURL: "fs://energy/index.html",
		Windows: application.Windows{
			ICON: icon,
		},
	})
	app.SetLocalLoad(application.LocalLoad{
		Scheme:     "fs",
		Domain:     "energy",
		ResRootDir: "resources",
		FS:         resources,
	})

	fmt.Println("run end")
}
