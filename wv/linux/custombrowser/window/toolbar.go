package window

import (
	"fmt"
	"github.com/energye/examples/wv/assets"
	"github.com/energye/examples/wv/linux/gtkhelper"
	"github.com/energye/lcl/api"
	"github.com/energye/lcl/types"
)

func (m *BrowserWindow) Toolbar() {
	headerBar, err := gtkhelper.NewHeaderBar()
	if err != nil {
		return
	}
	m.gtkToolbar = headerBar
	headerBar.SetName("browser-header-bar")
	SetWidgetStyle(headerBar.ToWidget(), `#browser-header-bar { background: rgba(56, 57, 60, 1);border: 0;margin: 0;}`)

	m.gtkWindow.SetTitlebar(headerBar)

	//headerBar.SetShowCloseButton(true)
	headerBar.SetVExpand(false)
	headerBar.SetVAlign(gtkhelper.ALIGN_CENTER)

	closeBtn := m.NewBrowserControlBtn(assets.GetResourcePath("btn-close.png"))
	m.closeBtn = closeBtn
	closeBtn.button.SetOnClick(func(sender *gtkhelper.Widget) {
		m.Close()
	})
	headerBar.PackEnd(closeBtn.button)

	maxBtn := m.NewBrowserControlBtn(assets.GetResourcePath("btn-max.png"))
	m.maxBtn = maxBtn
	maxBtn.button.SetOnClick(func(sender *gtkhelper.Widget) {
		m.Maximize()
	})
	headerBar.PackEnd(maxBtn.button)

	minBtn := m.NewBrowserControlBtn(assets.GetResourcePath("btn-min.png"))
	m.minBtn = minBtn
	minBtn.button.SetOnClick(func(sender *gtkhelper.Widget) {
		m.Minimize()
	})
	headerBar.PackEnd(minBtn.button)

	m.addrRightIcon = m.NewBrowserControlBtn(assets.GetResourcePath("addr-right-btn.png"))
	m.addrRightIcon.clickSH = m.addrRightIcon.button.SetOnClick(func(sender *gtkhelper.Widget) {
		fmt.Println("地址栏右侧图标")
		if browse := m.getActiveBrowse(); browse != nil {
			fmt.Println(browse.windowId)
			browse.webview.LoadURL("https://energye.github.io")
		}
	})
	headerBar.PackEnd(m.addrRightIcon.button)

	// 添加浏览器 button
	addBrowserBtn := m.NewBrowserControlBtn(assets.GetResourcePath("add.png"))
	m.addBrowserBtn = addBrowserBtn
	headerBar.PackEnd(addBrowserBtn.button)
	addBrowserBtn.button.SetOnClick(func(sender *gtkhelper.Widget) {
		println("IsMainThread:", api.MainThreadId() == api.CurrentThreadId())
		// 添加浏览器
		newBrowser := m.CreateBrowser("")
		m.OnCreateTabSheet(newBrowser)
		newBrowser.Create()
	})

}

func (m *BrowserWindow) UpdateToolbar() {
	if m.WindowState() == types.WsNormal {
		m.maxBtn.UpdateImage(assets.GetResourcePath("btn-max.png"))
	} else {
		m.maxBtn.UpdateImage(assets.GetResourcePath("btn-max-re.png"))
	}
}

func (m *BrowserWindow) Minimize() {
	m.SetWindowState(types.WsMinimized)
}

func (m *BrowserWindow) Maximize() {
	if m.WindowState() == types.WsNormal {
		m.SetWindowState(types.WsMaximized)
	} else {
		m.SetWindowState(types.WsNormal)
	}
}

type TabButton struct {
	button       *gtkhelper.EventBox
	box          *gtkhelper.Box
	icon         *gtkhelper.Image
	iconPath     string
	label        *gtkhelper.Label
	closeBtn     *gtkhelper.Button
	closeBtnIcon *gtkhelper.Image
	click        func()
	closeClick   func()
	styleCtx     *gtkhelper.StyleContext
	isActive     bool
	isClick      bool
}

func (m *TabButton) SetOnClick(fn func()) {
	m.click = fn
}

func (m *TabButton) SetOnCloseClick(fn func()) {
	m.closeClick = fn
}

func (m *TabButton) SetVisible(v bool) {
	m.button.SetVisible(v)
}

func (m *TabButton) SetTitle(s string) {
	m.label.SetText(s)
}

func (m *TabButton) Active(v bool) {
	m.isActive = v
	m.removeCss()
	if v {
		m.styleCtx.AddClass("active")
	}
}

func (m *TabButton) removeCss() {
	m.styleCtx.RemoveClass("active")
	m.styleCtx.RemoveClass("inactive")
	m.styleCtx.RemoveClass("click")
}

func (m *TabButton) UpdateImage(newImagePath string) {
	if m.iconPath != newImagePath {
		m.iconPath = newImagePath
		m.icon.SetFromFile(m.iconPath)
	}
}

func (m *BrowserWindow) NewTabButton(iconPath string, text string) *TabButton {
	tabButton := new(TabButton)
	button := gtkhelper.NewEventBox()
	tabButton.button = button
	button.SetHExpand(false)
	button.SetVExpand(false)
	button.SetSizeRequest(-1, 28)
	button.SetBorderWidth(0)
	button.SetVAlign(gtkhelper.ALIGN_CENTER)
	button.SetVisibleWindow(true)
	button.AddEvents(gtkhelper.POINTER_MOTION_MASK | gtkhelper.ENTER_NOTIFY_MASK | gtkhelper.LEAVE_NOTIFY_MASK)
	styleCtx := button.GetStyleContext()
	tabButton.styleCtx = styleCtx
	styleCtx.AddClass("tab")
	button.SetOnEnter(func(sender *gtkhelper.Widget, event *gtkhelper.EventCrossing) {
		if !tabButton.isActive {
			tabButton.removeCss()
		}
		styleCtx.AddClass("active")
		if tabButton.isClick {
			tabButton.isClick = false
		}
	})
	button.SetOnLeave(func(sender *gtkhelper.Widget, event *gtkhelper.EventCrossing) {
		if !tabButton.isActive {
			tabButton.removeCss()
		}
		if tabButton.isClick {
			styleCtx.AddClass("click")
			tabButton.isClick = false
		}
	})
	button.SetOnClick(func(sender *gtkhelper.Widget, event *gtkhelper.EventButton) {
		tabButton.isClick = true
		if tabButton.click != nil {
			tabButton.click()
		}
	})
	box := gtkhelper.NewBox(gtkhelper.ORIENTATION_HORIZONTAL, 4)
	box.SetVAlign(gtkhelper.ALIGN_CENTER)
	tabButton.box = box
	button.Add(box)

	icon := gtkhelper.NewImageFromFile(iconPath)
	icon.SetSizeRequest(16, 16)
	tabButton.icon = icon
	box.PackStart(icon, false, false, 4)

	label := gtkhelper.NewLabel(text)
	label.SetXAlign(0.0)
	label.SetEllipsize(gtkhelper.ELLIPSIZE_END)
	label.SetHExpand(false)
	label.SetVExpand(false)
	label.SetSizeRequest(45, -1)
	tabButton.label = label
	box.PackStart(label, true, true, 0)

	closeBtn := gtkhelper.NewButton()
	tabButton.closeBtn = closeBtn
	closeBtnIcon := gtkhelper.NewImageFromIconName("window-close-symbolic", gtkhelper.ICON_SIZE_MENU)
	tabButton.closeBtnIcon = closeBtnIcon
	closeBtn.SetImage(closeBtnIcon)
	closeBtn.SetSizeRequest(16, 16)
	closeBtnStyleCtx := closeBtn.GetStyleContext()
	closeBtnStyleCtx.AddClass("tab-close-button")
	closeBtn.SetOpacity(0.7)
	closeBtn.SetFocusOnClick(false)
	closeBtn.SetOnClick(func(sender *gtkhelper.Widget) {
		if tabButton.closeClick != nil {
			tabButton.closeClick()
		}
	})
	closeBtn.SetOnEnter(func(sender *gtkhelper.Widget, event *gtkhelper.EventCrossing) {
		styleCtx.RemoveClass("active")
		styleCtx.RemoveClass("inactive")
		styleCtx.AddClass("active")
	})
	closeBtn.SetOnLeave(func(sender *gtkhelper.Widget, event *gtkhelper.EventCrossing) {
		styleCtx.RemoveClass("inactive")
		styleCtx.RemoveClass("active")
	})
	box.PackEnd(closeBtn, false, false, 4)

	return tabButton
}
