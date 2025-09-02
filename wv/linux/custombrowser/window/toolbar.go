package window

import (
	"github.com/energye/examples/wv/linux/gtkhelper"
)

func (m *BrowserWindow) Toolbar() {
	headerBar, err := gtkhelper.NewHeaderBar()
	if err != nil {
		return
	}
	headerBar.SetShowCloseButton(true)
	headerBar.SetName("custom-headerbar")
	headerBar.SetVExpand(false)
	headerBar.SetVAlign(gtkhelper.ALIGN_CENTER)

	m.gtkWindow.SetTitlebar(headerBar)

	//
	btn1 := m.NewButton("edit-delete-symbolic", "删除项目删除项目")
	headerBar.PackStart(btn1)

}
