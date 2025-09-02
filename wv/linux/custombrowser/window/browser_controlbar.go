package window

import (
	"github.com/energye/examples/wv/assets"
	"github.com/energye/examples/wv/linux/gtkhelper"
)

var (
	browserWidgetAddrLeft = 125
	btnSize               = 32
	btnMargin             = 10
)

func (m *BrowserWindow) UpdateBrowserBounds() {
	println("UpdateBrowserBounds:", m.box.Width(), m.browserBar.Width())
	if m.addr != nil {
		newWidth := int(m.box.Width()) - (32*4 + 50)
		m.addr.SetSizeRequest(newWidth, -1)
		m.gtkBrowserBar.Move(m.addrRightIcon.button, int(m.box.Width())-32+10, 5)
	}
}

func (m *BrowserWindow) BrowserControlBar() {
	// 浏览器控制按钮
	backBtn := m.NewBrowserControlBtn(assets.GetResourcePath("back.png"))
	m.gtkBrowserBar.Put(backBtn.button, 10, 7)
	forwardBtn := m.NewBrowserControlBtn(assets.GetResourcePath("forward.png"))
	m.gtkBrowserBar.Put(forwardBtn.button, 32+20, 7)
	refreshBtn := m.NewBrowserControlBtn(assets.GetResourcePath("refresh.png"))
	m.gtkBrowserBar.Put(refreshBtn.button, 32*2+30, 7)
	m.backBtn = backBtn
	m.forwardBtn = forwardBtn
	m.refreshBtn = refreshBtn

	// 地址栏
	addr := gtkhelper.NewEntry()
	addr.SetName("browser-addr")
	addr.SetPlaceholderText("输入网站地址")
	addr.SetHAlign(gtkhelper.ALIGN_CENTER)
	addr.SetHExpand(true)
	addr.SetOnKeyRelease(func(sender *gtkhelper.Widget, key *gtkhelper.EventKey) bool {
		println("entry.SetOnKeyPress key:", key.KeyVal(), gtkhelper.KEY_Return, gtkhelper.KEY_KP_Enter)
		if key.KeyVal() == gtkhelper.KEY_Return || key.KeyVal() == gtkhelper.KEY_KP_Enter {
			println("entry.SetOnKeyPress text:", addr.GetText())
			return true
		}
		return false
	})
	SetWidgetStyle(addr.ToWidget(), `entry { background: rgba(56, 57, 60, 1); color: #FFFFFF;} entry:focus { background: rgba(128, 128, 128, 0.4); }`)
	m.addr = addr
	m.gtkBrowserBar.Put(addr, 32*4+10, 5)

	// 地址栏右侧图标
	m.addrRightIcon = m.NewBrowserControlBtn(assets.GetResourcePath("addr-right-btn.png"))
	m.addrRightIcon.clickSH = m.addrRightIcon.button.SetOnClick(func(sender *gtkhelper.Widget) {

	})
	m.gtkBrowserBar.Put(m.addrRightIcon.button, int(m.box.Width())-32+10, 5)

	m.UpdateBrowserBounds()
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
	btn.button.SetSizeRequest(btnSize, btnSize)
	btn.button.SetFocusOnClick(false)
	btnCss := gtkhelper.NewCssProvider()
	defer btnCss.Unref()
	btnCss.LoadFromData(`
button {
	background: transparent;
	border: none;
	padding: 2px;
}
button:hover {
	background: rgba(128, 128, 128, 0.2);
	border-radius: 2px;
}
button:active {
	background: rgba(128, 128, 128, 0.4);
}
`)

	btnStyleCtx := btn.button.GetStyleContext()
	btnStyleCtx.AddProvider(btnCss, gtkhelper.STYLE_PROVIDER_PRIORITY_APPLICATION)
	btnIcon := gtkhelper.NewImageFromFile(imagePath)
	btn.button.SetImage(btnIcon)
	return btn
}
