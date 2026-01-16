package main

import (
	"github.com/energye/examples/wv/darwin/action/src"
	"github.com/energye/lcl/lcl"
)

func main() {
	lcl.Init(nil, nil)
	lcl.Application.Initialize()
	lcl.Application.SetMainFormOnTaskBar(true)
	lcl.Application.NewForm(&src.MainForm)
	lcl.Application.Run()
}
