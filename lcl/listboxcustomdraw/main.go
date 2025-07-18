package main

import (
	. "github.com/energye/examples/syso"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
	"github.com/energye/lcl/types/colors"
)

func init() {
	TestLoadLibPath()
}
func main() {
	lcl.Init(nil, nil)
	lcl.Application.Initialize()
	lcl.Application.SetMainFormOnTaskBar(true)
	var mainForm lcl.TEngForm
	lcl.Application.NewForm(&mainForm)
	mainForm.SetCaption("Hello")
	mainForm.SetPosition(types.PoScreenCenter)
	mainForm.EnabledMaximize(false)
	mainForm.SetWidth(500)
	mainForm.SetHeight(400)

	var itemHeight int32 = 30
	listbox := lcl.NewListBox(&mainForm)
	listbox.SetParent(&mainForm)
	listbox.SetStyle(types.LbOwnerDrawFixed)
	listbox.SetAlign(types.AlClient)
	listbox.Items().Add("第一项")
	listbox.Items().Add("第二项")
	listbox.Items().Add("第三项")
	listbox.Items().Add("第四项")
	listbox.Items().Add("第五项")
	listbox.Items().Add("第六项")
	listbox.Items().Add("第七项")
	listbox.Items().Add("第八项")
	listbox.SetItemHeight(itemHeight)
	listbox.SetOnDrawItem(func(control lcl.IWinControl, index int32, aRect types.TRect, state types.TOwnerDrawState) {
		canvas := listbox.Canvas()
		s := listbox.Items().Strings(index)
		fw := canvas.TextWidthWithString(s)
		fh := canvas.TextHeightWithString(s)
		font := canvas.FontToFont()
		brush := canvas.BrushToBrush()
		pen := canvas.PenToPen()
		font.SetColor(colors.ClBlack)
		brush.SetColor(colors.ClBtnFace)
		canvas.FillRectWithRect(aRect)
		brush.SetColor(0x00FFF7F7)
		pen.SetColor(colors.ClSkyblue)
		canvas.RectangleWithIntX4(aRect.Left+1, aRect.Top+1, aRect.Right-1, aRect.Bottom-1)
		canvas.RectangleWithIntX4(aRect.Left, aRect.Top, aRect.Right, aRect.Bottom)
		if state.In(types.OdSelected) {
			brush.SetColor(0x00FFB2B5)
			canvas.RectangleWithIntX4(aRect.Left+1, aRect.Top+1, aRect.Right-1, aRect.Bottom-1)
			font.SetColor(colors.ClBlue)
			if state.In(types.OdFocused) {
				canvas.DrawFocusRect(aRect)
			}
		}
		canvas.TextOutWithIntX2String(aRect.Left+(aRect.Right-fw)/2, aRect.Top+(itemHeight-fh)/2, s)
	})
	lcl.Application.Run()
}
