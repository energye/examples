package main

import (
	"fmt"
	. "github.com/energye/examples/syso"
	"github.com/energye/lcl/api"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
)

type TMainForm struct {
	lcl.TEngForm
	Button1 lcl.IButton
}

type TForm1 struct {
	lcl.TEngForm
	Button1 lcl.IButton
}

var (
	mainForm TMainForm
	form1    TForm1
)

func init() {
	TestLoadLibPath()
}

func main() {
	lcl.Init(nil, nil)
	lcl.RunApp(&mainForm, &form1)
}

func (m *TMainForm) FormCreate(sender lcl.IObject) {
	fmt.Println("TMainForm FormCreate")
	//m.SetOnWndProc(func(msg *types.TMessage) {
	//	m.InheritedWndProc(msg)
	//	fmt.Println("msg", msg)
	//})
	m.SetCaption("Hello")
	m.EnabledMaximize(false)
	m.WorkAreaCenter()
	m.SetWidth(600)
	m.SetHeight(600)
	m.Button1 = lcl.NewButton(m)
	m.Button1.SetParent(m)
	m.Button1.SetCaption("窗口1")
	m.Button1.SetLeft(50)
	m.Button1.SetTop(50)
	m.Button1.SetOnClick(m.OnButton1Click)
}

func (f *TMainForm) OnCloseQuery(Sender lcl.IObject, CanClose *bool) {
	var buttons types.TMsgDlgButtons
	buttons = types.NewSet(types.MbYes)
	*CanClose = api.MessageDlg("是否退出？", types.MtConfirmation, buttons, types.MbNo) == types.IdYes
}

func (f *TMainForm) OnButton1Click(object lcl.IObject) {
	form1.Show()
	fmt.Println("清除事件")
	f.Button1.SetOnClick(f.OnButton1Click)
	fmt.Println("更换事件")
	f.Button1.SetOnClick(f.OnButton2Click)
}

func (f *TMainForm) OnButton2Click(object lcl.IObject) {
	fmt.Println("换成button2click事件了啊")
}

// ---------- Form1 ----------------

func (f *TForm1) FormCreate(sender lcl.IObject) {
	fmt.Println("TForm1 FormCreate")
	f.Button1 = lcl.NewButton(f)
	fmt.Println("f.Button1:", f.Button1.Instance())
	f.Button1.SetParent(f)
	f.Button1.SetCaption("我是按钮")
	f.Button1.SetOnClick(f.OnButton1Click)
}

func (f *TForm1) OnButton1Click(object lcl.IObject) {
	api.ShowMessage("Click")
}
