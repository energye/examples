package main

import (
	. "github.com/energye/examples/syso"
	"github.com/energye/lcl/inits"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
	"github.com/energye/lcl/types/colors"
)

type TForm1 struct {
	lcl.TForm
	Button1 lcl.IXButton
}

var form1 TForm1

func main() {
	inits.Init(nil, nil)
	lcl.RunApp(&form1)
}

func (f *TForm1) FormCreate(sender lcl.IObject) {

	f.SetWidth(600)
	f.SetHeight(400)
	f.ScreenCenter()

	f.Button1 = lcl.NewXButton(f)
	f.Button1.SetParent(f)
	f.Button1.SetDrawMode(types.DimCenter)

	f.Button1.Picture().LoadFromFile("icon.png")

	f.Button1.SetBackColor(colors.ClAzure)
	f.Button1.SetNormalFontColor(colors.ClBlue)

	f.Button1.SetHoverColor(colors.ClLinen)
	f.Button1.SetHoverFontColor(colors.ClGreen)

	f.Button1.SetDownColor(colors.ClSilver)
	f.Button1.SetDownFontColor(colors.ClFuchsia)

	f.Button1.SetBorderWidth(1)
	f.Button1.SetBorderColor(colors.ClBrown)

	f.Button1.SetCaption("文字")
	//f.Button1.SetShowCaption(false)

	f.Button1.SetBounds(10, 10, 80, 40)

	f.Button1.SetOnClick(f.OnButton1Click)
}

func (f *TForm1) OnButton1Click(object lcl.IObject) {
	lcl.ShowMessage("Click")
}

func init() {
	Chdir("lcl/xbutton")
}
