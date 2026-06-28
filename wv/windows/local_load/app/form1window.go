// ==============================================================================
// 📚 form1.go 用户代码文件
// 📌 该文件不存在时自动创建
// ✏️ 可在此文件中添加事件处理和业务逻辑
//    生成时间: 2025-12-15 22:42:55
// ==============================================================================

package app

import (
	"fmt"
	"github.com/energye/energy/v3/application"
	"github.com/energye/energy/v3/core"
	"github.com/energye/energy/v3/ipc"
	"github.com/energye/examples/wv/assets"
	"github.com/energye/lcl/api"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/rtl/version"
	"github.com/energye/lcl/types"
	"strconv"
)

// OnFormCreate 窗体初始化事件
func (m *TForm1Window) OnFormCreate(sender lcl.IObject) {
	println("OnFormCreate")
	// TODO 在此处添加窗体初始化代码
	m.SetShowInTaskBar(types.StAlways)
	m.Webview1.SetWindow(m)
	m.Webview1.SetAlign(types.AlNone)
	m.SetWidth(800)
	m.SetHeight(600)
	m.Webview1.SetTop(0)
	m.Webview1.SetLeft(0)
	m.Webview1.SetWidth(m.Width() - m.Webview1.Left()*2)
	m.Webview1.SetHeight(m.Height() - m.Webview1.Top()*2)
	m.Webview1.SetAnchors(types.NewSet(types.AkLeft, types.AkTop, types.AkRight, types.AkBottom))
	m.Webview1.SetOnLoadChange(func(url, title string, load core.TLoadChange) {
		fmt.Println("OnLoadChange:", url, title, load, m.BrowserId())
		if load == core.LcFinish {
			ipc.EmitOptions(&ipc.OptionsEvent{BrowserId: m.BrowserId(), Name: "native-text-change", Data: version.OSVersion.ToString()})
		}
	})
	m.Webview1.SetOnContextMenu(func(contextMenu *core.TContextMenuItem) {
		//contextMenu.Clear()
		contextMenu.Add("", core.CmkSeparator)
		contextMenu.Add("测试1", core.CmkCommand)
		test2, id := contextMenu.Add("测试2", core.CmkSub)
		fmt.Println("测试2:", id)
		_, id = test2.Add("测试2-测试", core.CmkCommand)
		fmt.Println("测试2-测试:", id)
		_, id = test2.Add("测试3-测试", core.CmkCommand)
		fmt.Println("测试3-测试:", id)
		contextMenu.Add("测试3", core.CmkCommand)
	})
	m.Webview1.SetOnContextMenuCommand(func(commandId int32, handle *bool) {
		fmt.Println("OnContextMenuCommand:", commandId)
		m.Webview1.ExecuteScriptCallback("document.title", func(result string, err string) {
			fmt.Println("ExecuteScriptCallback:", result, err)
		})
	})
	m.Webview1.SetOnPopupWindow(func(targetURL string) bool {
		fmt.Println("OnPopupWindow:", targetURL, api.CurrentThreadId() == api.MainThreadId())
		lcl.RunOnMainThreadAsync(func(id uint32) {
			newWindow := TForm1Window{}
			options := application.GApplication.Options
			options.DefaultURL = targetURL
			newWindow.SetOptions(options)
			lcl.Application.NewForm(&newWindow)
			newWindow.Show()
			Forms = append(Forms, &newWindow)
		})
		return true
	})
	m.Webview1.SetOnDragEnter(func(type_ core.TDragType, x, y int32) {
		fmt.Println("SetOnDragEnter --------------begin------------------", type_, x, y)
		ipc.Emit("drag-enter")
	})
	m.Webview1.SetOnDragLeave(func() {
		fmt.Println("SetOnDragLeave", "--------------zzz------------------")
	})
	m.Webview1.SetOnDragOver(func(data *core.TDragData, x, y int32) {
		da, err := strconv.Unquote("\"" + string(data.Data) + "\"")
		fmt.Println("SetOnDragOver --------------end------------------", x, y, da, err, data.Filenames)
		ipc.Emit("drag-over", da, data.Filenames)
	})
	//m.mainMenu()

	//

	//btn := lcl.NewButton(m)
	//btn.SetLeft(10)
	//btn.SetTop(100)
	//btn.SetCaption("原生按钮")
	//btn.SetParent(m)
	//txt := lcl.NewEdit(m)
	//txt.SetLeft(10)
	//txt.SetTop(200)
	//txt.SetText("原生文本框")
	//txt.SetParent(m)
	//txt.SetColor(colors.ClBlack)
	//txt.SetOnChange(func(sender lcl.IObject) {
	//	ipc.Emit("native-text-change", txt.Text())
	//})
	println("OnFormCreate end")

	tray := application.NewTrayIcon()

	trayMenu := tray.Menu()
	trayMenu.SetImageListEmbed(assets.Assets, []string{"resources/window-icon_64x64.png"})
	exit := trayMenu.AddMenuItem("退出").SetOnClick(func() {
		m.Close()
	})
	//exit.SetImage("window-icon_64x64.png")
	testdata, _ := assets.Assets.ReadFile("resources/window-icon_64x64.png")
	exit.SetBitmap(testdata)

	trayMenu.AddSeparator()
	//trayMenu.SetImageList([]string{"E:\\app\\workspace\\examples\\wv\\assets\\resources\\add.png"})
	testMenu := trayMenu.AddMenuItem("test")
	testMenu.SetOnMeasureItem(func(sender lcl.IObject, canvas lcl.ICanvas, width *int32, height *int32) {
		*height = 32
	})
	test2Menu := testMenu.AddSubMenuItem("test2")
	test2Menu.SetChecked(true)
	testMenu.AddSeparator()
	test2Menu = testMenu.AddSubMenuItem("test2222")
	test2Menu.SetRadio(true)
	test2Menu = testMenu.AddSubMenuItem("test3333")
	test2Menu.SetRadio(true)
	test2Menu.SetChecked(true)

	//tray.SetIcon("E:\\app\\workspace\\examples\\wv\\assets\\resources\\add.png")
	trayIconData, _ := assets.Assets.ReadFile("resources/add.png")
	tray.SetIconBytes(trayIconData)
	tray.SetOnMouseUp(func(button types.TMouseButton, shift types.TShiftState, X int32, Y int32) {
		fmt.Println("SetOnMouseUp")
	})
	tray.SetOnClick(func() {
		fmt.Println("SetOnClick")
	})
	tray.Show()

	m.SetOnThemeChange(func(isDark bool) {
		fmt.Println("OnThemeChange isDark:", isDark)
	})
}

func (m *TForm1Window) OnFormShow(sender lcl.IObject) {
	println("OnFormShow")
	// TODO 在此处添加窗体显示代码
	m.WorkAreaCenter()
	m.Webview1.CreateBrowser()
	println("OnFormShow end")
}

// OnFormCloseQuery 窗体关闭前询问事件
func (m *TForm1Window) OnFormCloseQuery(sender lcl.IObject, canClose *bool) bool {
	// TODO 在此处添加窗体关闭前询问代码
	fmt.Println("OnFormCloseQuery", m.BrowserId())
	return false
}

// OnFormClose 仅当 OnCloseQuery 中 CanClose 被设置为 True 后会触发
func (m *TForm1Window) OnFormClose(sender lcl.IObject, closeAction *types.TCloseAction) bool {
	// TODO 在此处添加窗体关闭代码
	fmt.Println("OnFormClose", m.BrowserId())
	return false
}

func (m *TForm1Window) mainMenu() {
	mainMenu := lcl.NewMainMenu(m)

	fileMenu := lcl.NewMenuItem(m)
	fileMenu.SetCaption("文件(&F)")
	mainMenu.Items().Add(fileMenu)

	subMenu := lcl.NewMenuItem(m)
	subMenu.SetCaption("新建(&N)")
	subMenu.SetShortCut(api.TextToShortCut("Ctrl+N"))
	subMenu.SetOnClick(func(lcl.IObject) {
		fmt.Println("新建")
	})
	fileMenu.Add(subMenu)

	subMenu = lcl.NewMenuItem(m)
	subMenu.SetCaption("打开(&O)")
	subMenu.SetShortCut(api.TextToShortCut("Ctrl+O"))
	subMenu.SetOnClick(func(sender lcl.IObject) {
		fmt.Println("打开")
	})
	fileMenu.Add(subMenu)

	subMenu = lcl.NewMenuItem(m)
	subMenu.SetCaption("保存(&S)")
	subMenu.SetShortCut(api.TextToShortCut("Ctrl+S"))
	subMenu.SetOnClick(func(sender lcl.IObject) {
		fmt.Println("保存")
	})
	fileMenu.Add(subMenu)

	subMenu = lcl.NewMenuItem(m)
	subMenu.SetCaption("-")
	fileMenu.Add(subMenu)

	subMenu = lcl.NewMenuItem(m)
	subMenu.SetCaption("退出(&Q)")
	subMenu.SetShortCut(api.TextToShortCut("Ctrl+Q"))
	subMenu.SetOnClick(func(lcl.IObject) {
		fmt.Println("退出")
		m.Close()
	})
	fileMenu.Add(subMenu)

	aboutMenu := lcl.NewMenuItem(m)
	aboutMenu.SetCaption("关于(&A)")
	mainMenu.Items().Add(aboutMenu)

	subMenu = lcl.NewMenuItem(m)
	subMenu.SetCaption("帮助(&H)")
	subMenu.SetOnClick(func(sender lcl.IObject) {
		fmt.Println("帮助")
	})
	aboutMenu.Add(subMenu)

}
