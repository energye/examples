package window

import (
	"fmt"
	"github.com/energye/lcl/api"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
)

func (m *Window) HookWndProcMessage() {
	mainMenu := lcl.NewMainMenu(m)
	mainMenu.SetOnMeasureItem(func(sender lcl.IObject, aCanvas lcl.ICanvas, width, height *int32) {
		*height = 44
	})
	appMenu := lcl.NewMenuItem(m)
	// 动态添加的，设置一个Unicode Apple logo char
	appMenu.SetCaption(types.AppleLogoChar)
	subItem := lcl.NewMenuItem(m)

	subItem.SetCaption("关于")
	subItem.SetOnClick(func(sender lcl.IObject) {
		api.ShowMessage("ENERGY\nhttps://github.com/energye/energy")
	})
	appMenu.Add(subItem)

	subItem = lcl.NewMenuItem(m)
	subItem.SetCaption("-")
	appMenu.Add(subItem)

	subItem = lcl.NewMenuItem(m)
	subItem.SetCaption("首选项...")
	subItem.SetShortCut(api.TextToShortCut("Meta+,"))
	subItem.SetOnClick(func(sender lcl.IObject) {
		api.ShowMessage("Preferences")
	})
	appMenu.Add(subItem)
	// 添加
	mainMenu.Items().Insert(0, appMenu)
	// 一级菜单
	item := lcl.NewMenuItem(m)
	item.SetCaption("文件(&F)")

	subMenu := lcl.NewMenuItem(m)
	subMenu.SetCaption("新建(&N)")
	subMenu.SetShortCut(api.TextToShortCut("Ctrl+N"))
	subMenu.SetOnClick(func(lcl.IObject) {
		fmt.Println("单击了新建")
	})
	item.Add(subMenu)

	subMenu = lcl.NewMenuItem(m)
	subMenu.SetCaption("打开(&O)")
	subMenu.SetShortCut(api.TextToShortCut("Ctrl+O"))
	item.Add(subMenu)

	subMenu = lcl.NewMenuItem(m)
	subMenu.SetCaption("保存(&S)")
	subMenu.SetShortCut(api.TextToShortCut("Ctrl+S"))
	item.Add(subMenu)

	// 分割线
	subMenu = lcl.NewMenuItem(m)
	subMenu.SetCaption("-")
	item.Add(subMenu)

	subMenu = lcl.NewMenuItem(m)
	subMenu.SetCaption("历史记录...")
	item.Add(subMenu)

	mItem := lcl.NewMenuItem(m)
	mItem.SetCaption("第一个历史记录")
	subMenu.Add(mItem)

	subMenu = lcl.NewMenuItem(m)
	subMenu.SetCaption("-")
	item.Add(subMenu)

	subMenu = lcl.NewMenuItem(m)
	subMenu.SetCaption("退出(&Q)")
	subMenu.SetShortCut(api.TextToShortCut("Ctrl+Q"))
	subMenu.SetOnClick(func(lcl.IObject) {
		m.Close()
	})
	item.Add(subMenu)

	mainMenu.Items().Add(item)

	item = lcl.NewMenuItem(m)
	item.SetCaption("关于(&A)")

	subMenu = lcl.NewMenuItem(m)
	subMenu.SetCaption("帮助(&H)")
	item.Add(subMenu)
	mainMenu.Items().Add(item)

	trayicon := lcl.NewTrayIcon(m)
	pm := lcl.NewPopupMenu(m)
	item = lcl.NewMenuItem(m)
	item.SetCaption("显示(&S)")
	item.SetOnClick(func(lcl.IObject) {
		fmt.Println("show")
		m.BringToFront()
		m.Show()
	})
	pm.Items().Add(item)

	item = lcl.NewMenuItem(m)
	item.SetCaption("退出(&E)")
	item.SetOnClick(func(lcl.IObject) {
		m.Close()
	})
	pm.Items().Add(item)
	trayicon.SetPopUpMenu(pm)
	trayicon.SetVisible(true)

	png := lcl.NewPortableNetworkGraphic()
	png.LoadFromFile(getResourcePath("window-icon_64x64.png"))
	trayicon.Icon().Assign(png)
	png.Free()
}

//现代 macOS 工具栏开发最佳实践总结
//
//理解“统一工具栏”：从 macOS 11 (Big Sur) 开始，工具栏和标题栏在视觉上融合。使用 isNavigational 和 allowedAligned 属性来正确放置你的项。
//明确项的角色：
//导航类 (isNavigational = true)：如前进、后退、侧边栏切换。靠左放置。
//主要操作/搜索 (principalItem)：如搜索栏。居中放置。
//内容相关操作 (allowedAligned = .trailing)：如分享、排序、查看选项。靠右放置。
//灵活空间 (.flexibleSpace, .space)：用于布局和对齐。
//优先使用 SF Symbols：确保图标在不同主题和状态下的一致性。
//善用分组：对于相关的操作（如视图切换：列表、图标、分栏），使用 NSToolbarItemGroup 并以 collapsed 模式显示，以节省空间。
//响应式显示：正确设置 visibilityPriority，确保在窗口变窄时，最重要的项仍然可见，不重要的项会被自动隐藏到溢出菜单中。
