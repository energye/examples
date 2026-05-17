package window

import (
	gtk3 "github.com/energye/energy/v3/platform/linux/gtk3/cgo"
	gtk3types "github.com/energye/energy/v3/platform/linux/types"
	"github.com/energye/examples/wv/assets"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
	"net/url"
	"unsafe"
)

var (
	btnSize = 32
)

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
	backBtn.button.SetOnClick(func(sender gtk3types.PGtkWidget, userData gtk3types.GPointer) {
		if browse := m.getActiveBrowse(); browse != nil && browse.webview.CanGoBack() {
			browse.webview.Stop()
			browse.webview.GoBack()
		}
	})
	forwardBtn.button.SetOnClick(func(sender gtk3types.PGtkWidget, userData gtk3types.GPointer) {
		if browse := m.getActiveBrowse(); browse != nil && browse.webview.CanGoForward() {
			browse.webview.Stop()
			browse.webview.GoForward()
		}
	})
	refreshBtn.button.SetOnClick(func(sender gtk3types.PGtkWidget, userData gtk3types.GPointer) {
		if browse := m.getActiveBrowse(); browse != nil {
			if browse.isLoading {
				browse.webview.Stop()
			} else {
				browse.webview.Reload()
			}
		}
	})

	// 地址栏
	addr := lcl.NewEdit(m)
	addr.SetParent(m.box)
	addr.SetLeft(32*4 + 10)
	addr.SetTop(5)
	addr.SetWidth(m.box.Width() - (addr.Left() + 5))
	addr.SetAlign(types.AlCustom)
	addr.SetAnchors(types.NewSet(types.AkLeft, types.AkTop, types.AkRight))
	addr.SetOnKeyUp(func(sender lcl.IObject, key *uint16, shift types.TShiftState) {
		tempKey := *key
		if tempKey == 13 || tempKey == 10 {
			targetUrl := addr.Text()
			println("entry.SetOnKeyPress text:", targetUrl)
			var scheme string
			if u, err := url.Parse(targetUrl); err != nil {
				return
			} else {
				if u.Scheme == "" {
					scheme = "http://"
				}
			}

			if browse := m.getActiveBrowse(); browse != nil {
				browse.webview.LoadURL(scheme + targetUrl)
			}
		}
	})
	m.addr = addr
	addrHandle := lcl.PlatformHandle(addr.Handle())
	addrEntry := gtk3.AsEntry(unsafe.Pointer(addrHandle.Gtk3Widget()))
	addrEntry.SetIconFromIconName(gtk3.ENTRY_ICON_PRIMARY, "search")
	SetWidgetStyle(addrEntry.ToWidget(), `entry { background: rgba(56, 57, 60, 1); color: #FFFFFF; caret-color: #FFFFFF;} entry:focus { background: rgba(128, 128, 128, 0.4); }`)
	//println("addrEntry", addrEntry.TypeFromInstance().Name())

}

type BrowserControlButton struct {
	button    *gtk3.Button
	image     *gtk3.Image
	imagePath string
	clickSH   gtk3types.ISignalHandlerID
	enable    bool
}

func (m *BrowserWindow) NewBrowserControlBtn(imagePath string) *BrowserControlButton {
	btn := new(BrowserControlButton)
	btn.imagePath = imagePath
	btn.button = gtk3.NewButton() // .ButtonNewWithLabel("button")
	btn.button.SetRelief(gtk3.RELIEF_NONE)
	btn.button.SetSizeRequest(btnSize, btnSize)
	btn.button.SetFocusOnClick(false)
	btnCss := gtk3.NewCssProvider()
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
	btnStyleCtx.AddProvider(btnCss, gtk3types.STYLE_PROVIDER_PRIORITY_APPLICATION)
	btnIcon := gtk3.NewImageFromFile(imagePath)
	btn.image = btnIcon
	btn.button.SetImage(btnIcon)
	return btn
}

func (m *BrowserControlButton) UpdateImage(newImagePath string) {
	if m.imagePath != newImagePath {
		m.imagePath = newImagePath
		m.image.SetFromFile(m.imagePath)
	}
}

func (m *BrowserControlButton) SetEnable(v bool) {
	m.enable = v
	m.button.SetSensitive(v)
}
