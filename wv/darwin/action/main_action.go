package main

import (
	"github.com/energye/energy/v3/application"
	_ "github.com/energye/examples/syso"
	"github.com/energye/examples/wv/darwin/action/src"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types/colors"
)

func main() {
	application.GApplication = &application.Application{
		Options: application.Options{
			//Frameless: true,
			WindowIsTransparent:  true,
			WebviewIsTransparent: true,
			BackgroundColor:      colors.NewARGB(0, 0, 0, 0),
			MacOS: application.MacOS{
				AppearanceNamed: application.NSAppearanceNameDarkAqua,
			},
		},
	}
	lcl.Init(nil, nil)
	lcl.Application.Initialize()
	lcl.Application.SetMainFormOnTaskBar(true)
	lcl.Application.NewForm(&src.MainForm)
	lcl.Application.Run()
}
