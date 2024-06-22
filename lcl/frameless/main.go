package main

import (
	_ "github.com/energye/examples/syso"
	"github.com/energye/lcl/inits"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
)

var MainForm TMainForm

type TMainForm struct {
	lcl.TForm
}

func main() {
	lcl.DEBUG = true
	inits.Init(nil, nil)
	lcl.Application.Initialize()
	lcl.Application.SetMainFormOnTaskBar(true)
	lcl.Application.CreateForm(&MainForm)
	lcl.Application.Run()
}

func (m *TMainForm) FormCreate(sender lcl.IObject) {
	m.SetCaption("frameless")
	m.SetPosition(types.PoScreenCenter)
	m.SetBorderStyleForFormBorderStyle(types.BsNone)
}
