package src

import (
	"fmt"
	"github.com/energye/energy/v3/pkgs/cocoa"
	"github.com/energye/energy/v3/window"
	"github.com/energye/lcl/api"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
	"unsafe"
)

type TMainForm struct {
	window.TWindow
	Btn      lcl.IButton
	Chk      lcl.ICheckBox
	MainMenu lcl.IMainMenu
}

var MainForm TMainForm

func (m *TMainForm) FormCreate(sender lcl.IObject) {
	//win32.SetWindowAlpha(hwnd, 100)
	//win32.SetWindowColorKey(hwnd, uint32(colors.ClBlue))
	//win32.SetWindowBlurBehind(hwnd, true)
	//win32.SetWindowDisplayAffinity(hwnd, win.WDA_EXCLUDEFROMCAPTURE)
	m.SetOptions()

	m.SetBackgroundColor(0, 0, 0, 0)
	//m.Frameless()
	//m.SetWindowTransparent()
	//m.SwitchFrostedMaterial("NSAppearanceNameLightAqua")
	//m.SwitchFrostedMaterial("NSAppearanceNameDarkAqua")

	m.SetPosition(types.PoScreenCenter)
	//m.SetWidth(300)
	//m.SetHeight(200)
	//m.SetColor(colors.ClBisque)

	box := lcl.NewPanel(m)
	box.SetParent(m)
	//box.SetColor(colors.ClBisque)
	box.SetLeft(0)
	box.SetTop(0)
	box.SetWidth(150)
	box.SetHeight(150)
	box.SetAnchors(types.NewSet(types.AkTop, types.AkLeft, types.AkRight, types.AkBottom))

	fmt.Println("box:", cocoa.GetObjectInheritanceChain(unsafe.Pointer(box.Handle())))

	m.initComponents()

}

func (m *TMainForm) OnShow(sender lcl.IObject) {
	//m.SetOptions()
}

func (f *TMainForm) OnActExecute(sender lcl.IObject) {
	api.ShowMessage("点击了action")
}

func (f *TMainForm) OnActUpdate(sender lcl.IObject) {
	lcl.AsAction(sender).SetEnabled(f.Chk.Checked())
}

func (m *TMainForm) initComponents() {
	m.Btn = lcl.NewButton(m)
	m.Btn.SetParent(m)
	m.Btn.SetLeft(80)
	m.Btn.SetTop(10)
	m.Btn.SetCaption("Button")

	m.Chk = lcl.NewCheckBox(m)
	m.Chk.SetParent(m)
	m.Chk.SetCaption("action状态演示")
	m.Chk.SetLeft(m.Btn.Left())
	m.Chk.SetTop(m.Btn.Top() + m.Btn.Height() + 10)
	m.Chk.SetChecked(true)

	// mainMenu
	m.MainMenu = lcl.NewMainMenu(m)

	menu := lcl.NewMenuItem(m)
	menu.SetCaption("菜单")
	m.MainMenu.Items().Add(menu)
	subMenu := lcl.NewMenuItem(m)
	menu.Add(subMenu)

	fmt.Println("Btn:", cocoa.GetObjectInheritanceChain(unsafe.Pointer(m.Btn.Handle())))
	fmt.Println("Chk:", cocoa.GetObjectInheritanceChain(unsafe.Pointer(m.Chk.Handle())))
}
