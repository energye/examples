package src

import (
	. "github.com/energye/energy/v3/platform/notification/types"
	"github.com/energye/lcl/lcl"
)

type TMainForm struct {
	lcl.TEngForm
	notifService INotification
	statusLabel  lcl.ILabel
	logMemo      lcl.IMemo
}

var MainForm TMainForm

// setStatus 设置状态栏文本
func (m *TMainForm) setStatus(text string) {
	m.statusLabel.SetCaption(text)
}

// appendLog 追加日志文本
func (m *TMainForm) appendLog(text string) {
	lcl.RunOnMainThreadAsync(func(id uint32) {
		currentText := m.logMemo.Lines().Text()
		m.logMemo.SetText(currentText + text)

		// 滚动到底部
		//m.logMemo.Perform(types.EM_SCROLLCARET, 0, 0)
	})
}
