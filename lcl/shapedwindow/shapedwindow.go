package main

import (
	. "github.com/energye/examples/syso"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
	"github.com/energye/lcl/types/colors"
)

type TMainForm struct {
	lcl.TEngForm
	CloseBtn lcl.IButton
}

var mainForm TMainForm

func init() {
	TestLoadLibPath()
}
func main() {
	lcl.Init(nil, nil)
	lcl.Application.Initialize()
	lcl.Application.SetMainFormOnTaskBar(true)
	lcl.Application.NewForm(&mainForm)
	lcl.Application.Run()
}

func (m *TMainForm) FormCreate(sender lcl.IObject) {
	m.SetCaption("Shaped Window")
	m.SetWidth(200)
	m.SetHeight(200)
	m.SetBorderStyleToFormBorderStyle(types.BsNone)
	m.SetColor(colors.ClRed)
	m.WorkAreaCenter()

	m.CloseBtn = lcl.NewButton(m)
	m.CloseBtn.SetParent(m)
	m.CloseBtn.SetLeft(50)
	m.CloseBtn.SetTop(88)
	m.CloseBtn.SetCaption("关 闭")

	m.SetOnShow(func(sender lcl.IObject) {
		shape := lcl.NewBitmap()
		shape.SetWidth(200)
		shape.SetHeight(200)
		shape.Canvas().EllipseWithIntX4(0, 0, 200, 200)
		m.SetShapeWithBitmap(shape)
		shape.Free()
	})

	m.CloseBtn.SetOnClick(func(sender lcl.IObject) {
		m.Close()
	})
}
