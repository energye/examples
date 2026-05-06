package main

import (
	_ "github.com/energye/examples/syso"
	"github.com/energye/examples/wv/windows/notification/src"
	"github.com/energye/lcl/api"
	"github.com/energye/lcl/lcl"
)

func main() {
	api.SetDebug(true)
	lcl.Init(nil, nil)
	lcl.Application.Initialize()
	lcl.Application.SetMainFormOnTaskBar(true)
	lcl.Application.NewForm(&src.MainForm)
	lcl.Application.Run()
}
