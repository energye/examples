package main

import (
	"fmt"
	. "github.com/energye/examples/syso"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
	"github.com/energye/lcl/types/colors"
	"math/rand"
	"time"
)

type TMainForm struct {
	lcl.TEngForm
}

type TForm1 struct {
	lcl.TEngForm
}

var MainForm TMainForm
var Form1 TForm1

func init() {
	TestLoadLibPath()
}
func main() {
	lcl.Init(nil, nil)
	lcl.Application.Initialize()
	lcl.Application.SetMainFormOnTaskBar(true)
	lcl.Application.SetScaled(true)
	lcl.Application.NewForms(&MainForm, &Form1)
	lcl.Application.Run()
}

func (m *TMainForm) FormCreate(sender lcl.IObject) {
	fmt.Println("main create")
	m.SetCaption("Main")
	m.SetPosition(types.PoScreenCenter)
	m.SetWidth(400)
	m.SetHeight(300)
	m.SetColor(colors.RGBToColor(56, 57, 60))
	m.SetBorderWidth(0)
	box := lcl.NewPanel(m)
	box.SetParent(m)
	box.SetTop(5)
	box.SetLeft(5)
	box.SetWidth(m.Width() - 10)
	box.SetHeight(m.Height() - 10)
	box.SetDoubleBuffered(true)
	box.SetBevelColor(colors.ClBlue)
	box.SetBevelWidth(5)
	box.SetAlign(types.AlClient)
	box.SetAnchors(types.NewSet(types.AkLeft, types.AkTop, types.AkRight, types.AkBottom))

	btn := lcl.NewButton(m)
	btn.SetParent(box)
	btn.SetCaption("test")
	btn.SetOnClick(func(sender lcl.IObject) {
		rand.Seed(time.Now().UnixNano())
		Form1.SetLeft(rand.Int31n(100))
		Form1.SetTop(rand.Int31n(100))
		Form1.Show()
		Form1.SetFocus()
	})
}

func (m *TMainForm) CreateParams(params *types.TCreateParams) {
	fmt.Println("调用此过程  TMainForm.CreateParams")
}

func (m *TForm1) FormCreate(sender lcl.IObject) {
	fmt.Println("form1 create")
	m.SetCaption("Form1")
}

func (m *TForm1) CreateParams(params *types.TCreateParams) {
	fmt.Println("调用此过程 TForm1.CreateParams")
}
