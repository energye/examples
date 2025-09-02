package window

import (
	"github.com/energye/examples/wv/assets"
	"github.com/energye/examples/wv/linux/gtkhelper"
)

var (
	browserWidgetAddrLeft = 125
)

func (m *BrowserWindow) UpdateBrowserBounds() {
	if m.addr != nil {
		newWidth := int(m.Width()) - (browserWidgetAddrLeft + 32*3)
		m.addr.SetSizeRequest(newWidth, -1)
	}
}

func (m *BrowserWindow) BrowserControlBar() {
	backBtn := m.NewBrowserControlBtn(assets.GetResourcePath("back.png"))
	m.gtkControlBrowserBarWidget.Put(backBtn.button, 10, 7)
	forwardBtn := m.NewBrowserControlBtn(assets.GetResourcePath("forward.png"))
	m.gtkControlBrowserBarWidget.Put(forwardBtn.button, 45, 7)
	refreshBtn := m.NewBrowserControlBtn(assets.GetResourcePath("refresh.png"))
	m.gtkControlBrowserBarWidget.Put(refreshBtn.button, 80, 7)
	m.backBtn = backBtn
	m.forwardBtn = forwardBtn
	m.refreshBtn = refreshBtn

	addr := gtkhelper.NewEntry()
	addr.SetPlaceholderText("输入网站地址")
	//newWidth := int(m.Width()) - (browserWidgetAddrLeft + 32*3)
	//addr.SetSizeRequest(newWidth, -1)
	//fmt.Println("newWidth:", newWidth)
	addr.SetHAlign(gtkhelper.ALIGN_CENTER)
	//addr.SetHExpand(true)
	addr.SetOnKeyRelease(func(sender *gtkhelper.Widget, key *gtkhelper.EventKey) bool {
		println("entry.SetOnKeyPress key:", key.KeyVal(), gtkhelper.KEY_Return, gtkhelper.KEY_KP_Enter)
		if key.KeyVal() == gtkhelper.KEY_Return || key.KeyVal() == gtkhelper.KEY_KP_Enter {
			println("entry.SetOnKeyPress text:", addr.GetText())
			return true
		}
		return false
	})
	m.addr = addr
	m.gtkControlBrowserBarWidget.Put(addr, browserWidgetAddrLeft, 5)
	m.gtkControlBrowserBarWidget.ShowAll()
}

type BrowserControlButton struct {
	button  *gtkhelper.Button
	image   *gtkhelper.Image
	clickSH *gtkhelper.SignalHandler
}

func (m *BrowserWindow) NewBrowserControlBtn(imagePath string) *BrowserControlButton {
	btn := new(BrowserControlButton)
	btn.button = gtkhelper.NewButton() // .ButtonNewWithLabel("button")
	btn.button.SetRelief(gtkhelper.RELIEF_NONE)
	btn.button.SetSizeRequest(32, 32)
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

	btnStyleCtx := btn.button.GetStyleContext()
	btnStyleCtx.AddProvider(btnCss, gtkhelper.STYLE_PROVIDER_PRIORITY_APPLICATION)
	btnIcon := gtkhelper.NewImageFromFile(imagePath)
	btn.button.SetImage(btnIcon)
	return btn
}
