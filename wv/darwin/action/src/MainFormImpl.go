package src

import (
	"github.com/energye/energy/v3/window"
	"github.com/energye/lcl/api"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
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

	m.SetBackgroundColor(0, 0, 0, 0)
	m.Toolbar()
	//m.Frameless()
	//m.SetWindowTransparent()
	//m.SwitchFrostedMaterial("NSAppearanceNameLightAqua")
	//m.SwitchFrostedMaterial("NSAppearanceNameDarkAqua")

	m.SetPosition(types.PoScreenCenter)
	//m.SetWidth(300)
	//m.SetHeight(200)
	//m.SetColor(colors.ClBisque)

	//box := lcl.NewPanel(m)
	//box.SetParent(m)
	//box.SetColor(colors.ClBisque)
	//box.SetLeft(0)
	//box.SetTop(0)
	//box.SetWidth(150)
	//box.SetHeight(150)
	//box.SetAnchors(types.NewSet(types.AkTop, types.AkLeft, types.AkRight, types.AkBottom))

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

func (f *TMainForm) initComponents() {
	f.Btn = lcl.NewButton(f)
	f.Btn.SetParent(f)
	f.Btn.SetLeft(80)
	f.Btn.SetTop(10)
	f.Btn.SetCaption("Button")

	f.Chk = lcl.NewCheckBox(f)
	f.Chk.SetParent(f)
	f.Chk.SetCaption("action状态演示")
	f.Chk.SetLeft(f.Btn.Left())
	f.Chk.SetTop(f.Btn.Top() + f.Btn.Height() + 10)
	f.Chk.SetChecked(true)

	// mainMenu
	f.MainMenu = lcl.NewMainMenu(f)

	menu := lcl.NewMenuItem(f)
	menu.SetCaption("菜单")
	f.MainMenu.Items().Add(menu)
	subMenu := lcl.NewMenuItem(f)
	menu.Add(subMenu)
}
