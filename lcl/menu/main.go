package main

import (
	"fmt"
	_ "github.com/energye/examples/syso"
	"github.com/energye/lcl/api"
	"github.com/energye/lcl/inits"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
	"github.com/energye/lcl/types/colors"
	"runtime"
)

type TMainForm struct {
	lcl.TForm
	mainMenu lcl.IMainMenu
}

var (
	mainForm TMainForm
)

func main() {
	inits.Init(nil, nil)
	lcl.RunApp(&mainForm)
}

func (f *TMainForm) FormCreate(sender lcl.IObject) {

	f.SetCaption("Menu example")
	f.ScreenCenter()
	//f.SetBorderStyleForFormBorderStyle(types.BsNone)

	// TMainMenu
	f.mainMenu = lcl.NewMainMenu(f)
	f.mainMenu.SetOnMeasureItem(func(sender lcl.IObject, aCanvas lcl.ICanvas, width, height *int32) {
		*height = 44
	})

	// macOS下专有的
	if runtime.GOOS == "darwin" {
		fmt.Println("darwin")
		// https://wiki.lazarus.freepascal.org/Mac_Preferences_and_About_Menu
		// 动态添加的，静态好像是通过设计器将顶级的菜单标题设置为应用程序名，但动态的就是另一种方式
		appMenu := lcl.NewMenuItem(f)
		// 动态添加的，设置一个Unicode Apple logo char
		appMenu.SetCaption(types.AppleLogoChar)
		subItem := lcl.NewMenuItem(f)
		// ----
		subItem.SetCaption("关于")
		subItem.SetOnClick(func(sender lcl.IObject) {
			lcl.ShowMessage("About")
		})
		appMenu.Add(subItem)
		// --
		subItem = lcl.NewMenuItem(f)
		subItem.SetCaption("-")
		appMenu.Add(subItem)

		// ---
		subItem = lcl.NewMenuItem(f)
		subItem.SetCaption("首选项...")
		subItem.SetShortCut(api.DTextToShortCut("Meta+,"))
		subItem.SetOnClick(func(sender lcl.IObject) {
			lcl.ShowMessage("Preferences")
		})
		appMenu.Add(subItem)
		// 添加
		f.mainMenu.Items().Insert(0, appMenu)
	}

	// 一级菜单
	item := lcl.NewMenuItem(f)
	item.SetCaption("文件(&F)")

	subMenu := lcl.NewMenuItem(f)
	subMenu.SetCaption("新建(&N)")
	subMenu.SetShortCut(api.DTextToShortCut("Ctrl+N"))
	subMenu.SetOnClick(func(lcl.IObject) {
		fmt.Println("单击了新建")
	})
	item.Add(subMenu)

	subMenu = lcl.NewMenuItem(f)
	subMenu.SetCaption("打开(&O)")
	subMenu.SetShortCut(api.DTextToShortCut("Ctrl+O"))
	item.Add(subMenu)

	subMenu = lcl.NewMenuItem(f)
	subMenu.SetCaption("保存(&S)")
	subMenu.SetShortCut(api.DTextToShortCut("Ctrl+S"))
	item.Add(subMenu)

	// 分割线
	subMenu = lcl.NewMenuItem(f)
	subMenu.SetCaption("-")
	item.Add(subMenu)

	subMenu = lcl.NewMenuItem(f)
	subMenu.SetCaption("历史记录...")
	item.Add(subMenu)

	m := lcl.NewMenuItem(f)
	m.SetCaption("第一个历史记录")
	subMenu.Add(m)

	subMenu = lcl.NewMenuItem(f)
	subMenu.SetCaption("-")
	item.Add(subMenu)

	subMenu = lcl.NewMenuItem(f)
	subMenu.SetCaption("退出(&Q)")
	subMenu.SetShortCut(api.DTextToShortCut("Ctrl+Q"))
	subMenu.SetOnClick(func(lcl.IObject) {
		f.Close()
	})
	item.Add(subMenu)

	f.mainMenu.Items().Add(item)

	item = lcl.NewMenuItem(f)
	item.SetCaption("关于(&A)")

	subMenu = lcl.NewMenuItem(f)
	subMenu.SetCaption("帮助(&H)")
	item.Add(subMenu)
	f.mainMenu.Items().Add(item)

	// TPopupMenu
	pm := lcl.NewPopupMenu(f)
	item = lcl.NewMenuItem(f)
	item.SetCaption("退出(&E)")
	item.SetOnClick(func(lcl.IObject) {
		f.Close()
	})
	pm.Items().Add(item)

	// 将窗口设置一个弹出菜单，右键单击就可显示
	f.SetPopupMenu(pm)

	f.SetOnPaint(f.OnFormPaint)
}

func (f *TMainForm) OnFormPaint(sender lcl.IObject) {
	///r := types.TRect{0, 0, f.Width(), f.Height()}
	//f.Canvas().TextRect(r, 0, 0, "右键弹出菜单")
	f.Canvas().Brush().SetStyle(types.BsClear)
	lcl.AsFont(f.Canvas().Font()).SetColor(colors.ClGreen)
	f.Canvas().TextOut(10, 80, "右键弹出菜单")
}
