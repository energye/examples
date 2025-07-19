package main

import (
	"fmt"
	. "github.com/energye/examples/syso"
	"github.com/energye/lcl/api"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/pkgs/win"
	"github.com/energye/lcl/pkgs/win/errcode"
	"github.com/energye/lcl/types"
	"os"
	"syscall"
	"unsafe"
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
var (
	kernel32dll  = syscall.NewLazyDLL("kernel32.dll")
	_CreateMutex = kernel32dll.NewProc("CreateMutexW")
)

// 不知道为什么GetLastError无法获取，只能重新申明下
func CreateMutex(lpMutexAttributes *win.TSecurityAttributes, bInitialOwner bool, lpName string) (uintptr, uintptr, error) {
	return _CreateMutex.Call(uintptr(unsafe.Pointer(lpMutexAttributes)), win.CBool(bInitialOwner), win.CStr(lpName))
}

func init() {
	TestLoadLibPath()
}
func main() {
	lcl.Init(nil, nil)
	// 利用互斥来演示exe单一运行，当然不止这一种方法了
	// GetLastError 无法获取错误
	Mutex, _, err := CreateMutex(nil, true, "SingleRunTest")
	defer win.ReleaseMutex(Mutex)
	fmt.Println("Mutex:", Mutex, err)
	if errNo, ok := err.(syscall.Errno); ok && errNo == errcode.ERROR_ALREADY_EXISTS {
		win.MessageBox(0, "我已经在运行中啦！", "运行提示", win.MB_OK+win.MB_ICONINFORMATION)
		hwnd := win.FindWindow("", "Hello")
		if hwnd != 0 {
			win.ShowWindowAsync(hwnd, win.SW_RESTORE)
			win.SetForegroundWindow(hwnd)
		}
		os.Exit(1)
	}

	lcl.RunApp(&mainForm, &form1)
}

func (m *TMainForm) FormCreate(sender lcl.IObject) {
	fmt.Println("TMainForm FormCreate")
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

func (f *TMainForm) OnFormCloseQuery(Sender lcl.IObject, CanClose *bool) {
	*CanClose = api.MessageDlg("是否退出？", types.MtConfirmation, types.NewSet(types.MbYes), types.MbNo) == types.IdYes
}

func (f *TMainForm) OnButton1Click(object lcl.IObject) {
	form1.Show()
	fmt.Println("清除事件")
	//f.Button1.SetOnClick(nil)
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
