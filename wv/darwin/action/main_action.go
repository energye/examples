package main

import (
	"github.com/energye/examples/lcl/action/src"
	. "github.com/energye/examples/syso"
	"github.com/energye/lcl/lcl"
)

func main() {
	lcl.Init(nil, nil)
	lcl.Application.Initialize()
	lcl.Application.SetMainFormOnTaskBar(true)
	lcl.Application.NewForm(&src.MainForm)
	lcl.Application.Run()

	println(1111)
}

func init() {
	Chdir("lcl/action")
}
