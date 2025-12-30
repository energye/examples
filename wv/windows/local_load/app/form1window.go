// ==============================================================================
// 📚 form1.go 用户代码文件
// 📌 该文件不存在时自动创建
// ✏️ 可在此文件中添加事件处理和业务逻辑
//    生成时间: 2025-12-15 22:42:55
// ==============================================================================

package app

import (
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
)

// OnFormCreate 窗体初始化事件
func (m *TForm1Window) OnFormCreate(sender lcl.IObject) {
	// TODO 在此处添加窗体初始化代码
	m.BrowserWindow1.SetDefaultURL("fs://energy/index.html")
	m.BrowserWindow1.SetAlign(types.AlClient)
	m.WorkAreaCenter()
}

func (m *TForm1Window) OnShow(sender lcl.IObject) {
	// TODO 在此处添加窗体显示代码
	m.BrowserWindow1.CreateBrowser()
}

// OnCloseQuery 窗体关闭前询问事件
func (m *TForm1Window) OnCloseQuery(sender lcl.IObject, canClose *bool) {
	// TODO 在此处添加窗体关闭前询问代码
}

// OnClose 仅当 OnCloseQuery 中 CanClose 被设置为 True 后会触发
func (m *TForm1Window) OnClose(sender lcl.IObject, closeAction *types.TCloseAction) {
	// TODO 在此处添加窗体关闭代码
}
