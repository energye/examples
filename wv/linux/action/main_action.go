package main

import (
	"github.com/energye/energy/v3/application"
	_ "github.com/energye/examples/syso"
	"github.com/energye/examples/wv/linux/action/src"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types/colors"
	"os"
)

func main() {
	os.Setenv("--ws", "gtk3")
	application.GApplication = &application.Application{
		Options: application.Options{
			WindowTransparent:  true,
			WebviewTransparent: true,
			BackgroundColor:    colors.NewARGB(0, 0, 0, 0),
		},
	}
	lcl.Init(nil, nil)
	lcl.Application.Initialize()
	lcl.Application.SetMainFormOnTaskBar(true)
	lcl.Application.NewForm(&src.MainForm)
	lcl.Application.Run()
}
