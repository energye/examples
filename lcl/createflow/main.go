package main

import (
	"fmt"
	. "github.com/energye/examples/syso"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
)

type Form1 struct {
	lcl.TEngForm
}

type Form2 struct {
	lcl.TEngForm
}

type Form3 struct {
	lcl.TEngForm
}

var (
	form1 Form1
	form2 Form2
	form3 Form3
)

func init() {
	TestLoadLibPath()
	Chdir("lcl/createflow")
}

func main() {
	lcl.Init(nil, nil)
	lcl.Application.Initialize()
	lcl.Application.SetMainFormOnTaskBar(true)
	// 创建Form 按参数顺序执行 FormCreate 回调函数
	lcl.Application.NewForms(&form1, &form2, &form3)
	lcl.Application.Run()
}

func (m *Form1) FormCreate(sender lcl.IObject) {
	fmt.Println("FormCreate 1")
	m.SetCaption("form1")
	m.SetBounds(100, 100, 300, 300)
	btn1 := lcl.NewButton(m)
	btn1.SetParent(m)
	btn1.SetBounds(10, 10, 100, 35)
	btn1.SetCaption("show form2")
	btn1.SetOnClick(func(sender lcl.IObject) {
		form2.Show()
	})
	btn2 := lcl.NewButton(m)
	btn2.SetParent(m)
	btn2.SetBounds(120, 10, 100, 35)
	btn2.SetCaption("show form3")
	btn2.SetOnClick(func(sender lcl.IObject) {
		form3.Show()
	})
}

func (m *Form1) CreateParams(params *types.TCreateParams) {
	fmt.Println("CreateParams 1")
}

func (m *Form2) FormCreate(sender lcl.IObject) {
	fmt.Println("FormCreate 2")
	m.SetCaption("form2")
	m.SetBounds(200, 200, 300, 300)
	m.SetShowInTaskBar(types.StAlways)
}

func (m *Form2) CreateParams(params *types.TCreateParams) {
	fmt.Println("CreateParams 2")
}

func (m *Form3) FormCreate(sender lcl.IObject) {
	fmt.Println("FormCreate 3")
	m.SetCaption("form3")
	m.SetBounds(300, 300, 300, 300)
	m.SetShowInTaskBar(types.StNever)
}

func (m *Form3) CreateParams(params *types.TCreateParams) {
	fmt.Println("CreateParams 3")
}
