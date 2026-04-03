package main

import (
	"fmt"
	"github.com/energye/examples/lcl/action/src"
	_ "github.com/energye/examples/syso"
	"github.com/energye/lcl/api"
	"github.com/energye/lcl/lcl"
)

func main() {
	lcl.Init(nil, nil)
	fmt.Println(api.LibAbout(), api.Widget())
	api.SetDebug(true)
	api.SetOnException(func(exceptionID, message string) {
		fmt.Println("exceptionID:", exceptionID, "exceptionMessage:", message)
	})
	lcl.Application.Initialize()
	lcl.Application.SetMainFormOnTaskBar(true)
	lcl.Application.NewForm(&src.MainForm)
	lcl.Application.Run()

	println(1111)
}

func init() {
	Chdir("lcl/action")
}
