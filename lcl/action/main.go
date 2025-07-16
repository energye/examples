package main

import (
	"github.com/energye/examples/lcl/action/src"
	. "github.com/energye/examples/syso"
	"github.com/energye/lcl/api/libname"
	"github.com/energye/lcl/lcl"
)

func main() {
	libname.LibName = "E:\\SWT\\gopath\\src\\github.com\\energye\\workspace\\gen\\gout\\liblcl.dll"
	lcl.Init(nil, nil)
	lcl.Application.Initialize()
	lcl.Application.SetMainFormOnTaskBar(true)
	lcl.Application.NewForm(&src.MainForm)
	lcl.Application.Run()
}

func init() {
	Chdir("lcl/action")
}
