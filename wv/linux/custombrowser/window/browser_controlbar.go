package window

import (
	"github.com/energye/examples/wv/linux/gtkhelper"
	"github.com/energye/lcl/api"
)

func (m *BrowserWindow) BrowserControlBar() {
	btn := gtkhelper.NewButton() // .ButtonNewWithLabel("button")
	btn.SetRelief(gtkhelper.RELIEF_NONE)
	btnCss := gtkhelper.NewCssProvider()
	defer btnCss.Unref()
	btnCss.LoadFromData(`
button {
	background: transparent;
	border: none;
	padding: 2px; /* 减小点击区域内边距 */
}
button:hover {
	background: rgba(128, 128, 128, 0.2); /* 悬停时轻微灰色背景 */
	border-radius: 2px;
}
button:active {
	background: rgba(128, 128, 128, 0.4); /* 点击时加深背景 */
}
`)

	btnStyleCtx := btn.GetStyleContext()
	btnStyleCtx.AddProvider(btnCss, gtkhelper.STYLE_PROVIDER_PRIORITY_APPLICATION)

	btn.SetOnClick(func(sender *gtkhelper.Widget) {
		println("btn.SetOnClick", sender, "IsMainThread:", api.MainThreadId() == api.CurrentThreadId())
		//sh.Disconnect()
	})
	btnIcon := gtkhelper.NewImageFromIconName("open-menu-symbolic", gtkhelper.ICON_SIZE_BUTTON)
	btn.SetImage(btnIcon)

	m.gtkControlBrowserBarWidget.Put(btn, 0, 0)

	entry := gtkhelper.NewEntry()
	entry.SetPlaceholderText("输入网站地址")
	entry.SetSizeRequest(250, -1)
	entry.SetHAlign(gtkhelper.ALIGN_CENTER)
	entry.SetOnKeyRelease(func(sender *gtkhelper.Widget, key *gtkhelper.EventKey) bool {
		println("entry.SetOnKeyPress key:", key.KeyVal(), gtkhelper.KEY_Return, gtkhelper.KEY_KP_Enter)
		if key.KeyVal() == gtkhelper.KEY_Return || key.KeyVal() == gtkhelper.KEY_KP_Enter {
			println("entry.SetOnKeyPress text:", entry.GetText())
			return true
		}
		return false
	})
	//headerBar.SetCustomTitle(entry)
}
