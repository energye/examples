package main

import (
	"fmt"
	"github.com/energye/lcl/api"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
	"github.com/energye/lcl/types/colors"
)

var MainForm TMainForm

type TMainForm struct {
	lcl.TEngForm
}

func main() {
	lcl.Init(nil, nil)
	lcl.Application.Initialize()
	lcl.Application.SetMainFormOnTaskBar(true)
	lcl.Application.NewForm(&MainForm)
	lcl.Application.Run()
}

func (m *TMainForm) FormCreate(sender lcl.IObject) {
	fmt.Println(api.Widget().IsGTK3())
	m.SetCaption("frameless")
	m.SetPosition(types.PoScreenCenter)
	m.SetColor(colors.ClWhite)
	m.SetBorderStyleToFormBorderStyle(types.BsNone)
	box := lcl.NewCustomPanel(m)
	box.SetBevelInner(types.BvNone)
	box.SetBevelOuter(types.BvNone)
	//box.SetAlign(types.AlClient)
	box.SetWidth(m.Width())
	box.SetHeight(m.Height())
	box.SetColor(colors.ClBlack)
	box.SetParent(m)
}
