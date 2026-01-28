// ==============================================================================
// 📚 form1.go 用户代码文件
// 📌 该文件不存在时自动创建
// ✏️ 可在此文件中添加事件处理和业务逻辑
//    生成时间: 2025-12-15 22:42:55
// ==============================================================================

package app

import (
	"fmt"
	"github.com/energye/energy/v3/ipc"
	"github.com/energye/energy/v3/wv"
	"github.com/energye/lcl/api"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/rtl/version"
	"github.com/energye/lcl/types"
)

// OnFormCreate 窗体初始化事件
func (m *TForm1Window) OnFormCreate(sender lcl.IObject) {
	println("OnFormCreate")
	// TODO 在此处添加窗体初始化代码
	m.SetShowInTaskBar(types.StAlways)
	m.Webview1.SetWindow(m)
	m.Webview1.SetAlign(types.AlNone)
	m.Webview1.SetTop(0)
	m.Webview1.SetLeft(0)
	m.Webview1.SetWidth(m.Width() - m.Webview1.Left()*2)
	m.Webview1.SetHeight(m.Height() - m.Webview1.Top()*2)
	m.Webview1.SetAnchors(types.NewSet(types.AkLeft, types.AkTop, types.AkRight, types.AkBottom))
	m.Webview1.SetOnLoadChange(func(url, title string, load wv.TLoadChange) {
		fmt.Println("OnLoadChange:", url, title, load, m.BrowserId())
		if load == wv.LcFinish {
			ipc.Emit("native-text-change", version.OSVersion.ToString())
		}
	})
	m.Webview1.SetOnContextMenu(func(contextMenu *wv.TContextMenuItem) {
		//contextMenu.Clear()
		contextMenu.Add("", wv.CmkSeparator)
		contextMenu.Add("测试1", wv.CmkCommand)
		test2, id := contextMenu.Add("测试2", wv.CmkSub)
		fmt.Println("测试2:", id)
		_, id = test2.Add("测试2-测试", wv.CmkCommand)
		fmt.Println("测试2-测试:", id)
		_, id = test2.Add("测试3-测试", wv.CmkCommand)
		fmt.Println("测试3-测试:", id)
		contextMenu.Add("测试3", wv.CmkCommand)
	})
	m.Webview1.SetOnContextMenuCommand(func(commandId int32) {
		fmt.Println("OnContextMenuCommand:", commandId)
		m.Webview1.ExecuteScriptCallback("document.title", func(result string, err string) {
			fmt.Println("ExecuteScriptCallback:", result, err)
		})
	})
	m.Webview1.SetOnPopupWindow(func(targetURL string) bool {
		fmt.Println("OnPopupWindow:", targetURL, api.CurrentThreadId() == api.MainThreadId())
		lcl.RunOnMainThreadAsync(func(id uint32) {
			newWindow := TForm1Window{}
			lcl.Application.NewForm(&newWindow)
			newWindow.Webview1.SetDefaultURL(targetURL)
			newWindow.Show()
			Forms = append(Forms, &newWindow)
		})
		return true
	})

	lcl.NewMainMenu(m)

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
