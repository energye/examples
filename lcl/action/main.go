package main

import (
	"github.com/energye/examples/lcl/action/src"
	. "github.com/energye/examples/syso"
	"github.com/energye/lcl/api/libname"
	"github.com/energye/lcl/lcl"
)

func main() {
	libname.LibName = ""
	lcl.Init(nil, nil)
	lcl.Application.Initialize()
	lcl.Application.SetMainFormOnTaskBar(true)
	lcl.Application.CreateForm(&src.MainForm)
	lcl.Application.Run()
}

func init() {
	Chdir("lcl/action")
}
