package main

import (
	"github.com/energye/energy/v3/application"
	"github.com/energye/examples/wv/darwin/action/src"
	"github.com/energye/lcl/lcl"
)

func main() {
	application.GApplication = &application.Application{Options: application.Options{Frameless: true}}
	lcl.Init(nil, nil)
	lcl.Application.Initialize()
	lcl.Application.SetMainFormOnTaskBar(true)
	lcl.Application.NewForm(&src.MainForm)
	lcl.Application.Run()
}
