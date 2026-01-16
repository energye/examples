package src

import (
	"github.com/energye/energy/v3/window"
	"github.com/energye/lcl/api"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
)

type TMainForm struct {
	window.TWindow
	ImgList  lcl.IImageList
	ActList  lcl.IActionList
	Tlbar    lcl.IToolBar
	Tlbtn    lcl.IToolButton
	Btn      lcl.IButton
	Chk      lcl.ICheckBox
	Act      lcl.IAction
	MainMenu lcl.IMainMenu
}

var MainForm TMainForm

func (m *TMainForm) FormCreate(sender lcl.IObject) {
	//m.Frameless()
	m.SetWindowTransparent()
	m.SwitchFrostedMaterial("NSAppearanceNameLightAqua")
	//m.SwitchFrostedMaterial("NSAppearanceNameDarkAqua")

	m.SetCaption("Hello")
	m.SetPosition(types.PoScreenCenter)
	//m.SetWidth(300)
	//m.SetHeight(200)

	m.initComponents()

	//box := lcl.NewPanel(m)
	//box.SetParent(m)
	//box.SetColor(colors.ClBlue)

}

func (f *TMainForm) OnActExecute(sender lcl.IObject) {
	api.ShowMessage("点击了action")
}

func (f *TMainForm) OnActUpdate(sender lcl.IObject) {
	lcl.AsAction(sender).SetEnabled(f.Chk.Checked())
}

func (f *TMainForm) initComponents() {
	f.ImgList = lcl.NewImageList(f)

	if lcl.Application.Icon().Handle() != 0 {
		f.ImgList.AddIcon(lcl.Application.Icon())
	}

	f.ActList = lcl.NewActionList(f)
	f.ActList.SetImages(f.ImgList)

	// 顶部工具条
	f.Tlbar = lcl.NewToolBar(f)
	f.Tlbar.SetParent(f)
	f.Tlbar.SetImages(f.ImgList)

	f.Tlbtn = lcl.NewToolButton(f)
	f.Tlbtn.SetParent(f.Tlbar)

	f.Btn = lcl.NewButton(f)
	f.Btn.SetParent(f)
	f.Btn.SetLeft(80)
	f.Btn.SetTop(f.Tlbar.Top() + f.Tlbar.Height() + 10)

	f.Chk = lcl.NewCheckBox(f)
	f.Chk.SetParent(f)
	f.Chk.SetCaption("action状态演示")
	f.Chk.SetLeft(f.Btn.Left())
	f.Chk.SetTop(f.Btn.Top() + f.Btn.Height() + 10)
	f.Chk.SetChecked(true)

	// action
	f.Act = lcl.NewAction(f)
	f.Act.SetCaption("action")
	f.Act.SetImageIndex(0)
	f.Act.SetHint("这是一个提示|长提示了")
	f.Act.SetOnExecute(f.OnActExecute)
	f.Act.SetOnUpdate(f.OnActUpdate)

	// mainMenu
	f.MainMenu = lcl.NewMainMenu(f)
	f.MainMenu.SetImages(f.ImgList)

	menu := lcl.NewMenuItem(f)
	menu.SetCaption("菜单")
	f.MainMenu.Items().Add(menu)
	subMenu := lcl.NewMenuItem(f)
	subMenu.SetAction(f.Act)
	menu.Add(subMenu)

	f.Btn.SetAction(f.Act)
	f.Tlbtn.SetAction(f.Act)
}
