package main

import (
	"github.com/energye/examples/wv/darwin/notification/src"
	"github.com/energye/lcl/lcl"
)

// codesign --force --deep --sign - notification.app
// codesign -dv notification.app
func main() {
	lcl.Init(nil, nil)
	lcl.Application.Initialize()
	lcl.Application.SetMainFormOnTaskBar(true)
	lcl.Application.NewForm(&src.MainForm)
	lcl.Application.Run()
}
