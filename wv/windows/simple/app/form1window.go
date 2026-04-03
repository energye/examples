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
	m.Webview1.SetWindow(m)
}

func (m *TForm1Window) OnFormShow(sender lcl.IObject) {
	// TODO 在此处添加窗体显示代码
}

// OnCloseQuery 窗体关闭前询问事件
func (m *TForm1Window) OnFormCloseQuery(sender lcl.IObject, canClose *bool) bool {
	// TODO 在此处添加窗体关闭前询问代码
	return false
}

// OnClose 仅当 OnCloseQuery 中 CanClose 被设置为 True 后会触发
func (m *TForm1Window) OnFormClose(sender lcl.IObject, closeAction *types.TCloseAction) bool {
	// TODO 在此处添加窗体关闭代码
	return false
}
