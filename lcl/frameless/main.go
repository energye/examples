package main

import (
	. "github.com/energye/examples/syso"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
)

var MainForm TMainForm

type TMainForm struct {
	lcl.TEngForm
}

func init() {
	TestLoadLibPath()
}

func main() {
	lcl.Init(nil, nil)
	lcl.Application.Initialize()
	lcl.Application.SetMainFormOnTaskBar(true)
	lcl.Application.NewForm(&MainForm)
	lcl.Application.Run()
}

func (m *TMainForm) FormCreate(sender lcl.IObject) {
	m.SetCaption("frameless")
	m.SetPosition(types.PoScreenCenter)
	m.SetBorderStyleToFormBorderStyle(types.BsNone)
	m.SetBorderStyleToBorderStyle()
}
