package window

import (
	"github.com/energye/examples/wv/assets"
	"github.com/energye/examples/wv/linux/gtkhelper"
)

func (m *BrowserWindow) Toolbar() {
	headerBar, err := gtkhelper.NewHeaderBar()
	if err != nil {
		return
	}
	m.gtkWindow.SetTitlebar(headerBar)

	headerBar.SetShowCloseButton(true)
	headerBar.SetVExpand(false)
	headerBar.SetVAlign(gtkhelper.ALIGN_CENTER)

	// test
	tabBtn1 := m.NewTabButton("edit-delete-symbolic", "删除项目删除项目")
	headerBar.PackStart(tabBtn1.button)

	// 添加浏览器 button
	addBrowserBtn := m.NewBrowserControlBtn(assets.GetResourcePath("add.png"))
	m.addBrowserBtn = addBrowserBtn
	headerBar.PackEnd(addBrowserBtn.button)
	addBrowserBtn.button.SetOnClick(func(sender *gtkhelper.Widget) {
		// 添加浏览器
	})
}

type TabButton struct {
	button       *gtkhelper.EventBox
	box          *gtkhelper.Box
	icon         *gtkhelper.Image
	label        *gtkhelper.Label
	closeBtn     *gtkhelper.Button
	closeBtnIcon *gtkhelper.Image
	click        func()
	closeClick   func()
}

func (m *TabButton) SetOnClick(fn func()) {
	m.click = fn
}

func (m *TabButton) SetOnCloseClick(fn func()) {
	m.closeClick = fn
}

func (m *BrowserWindow) NewTabButton(iconName string, text string) *TabButton {
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
	styleCtx.AddClass("tab")
	styleCtx.AddClass("active")
	button.SetOnEnter(func(sender *gtkhelper.Widget, event *gtkhelper.EventCrossing) {
		println("event.SetOnEnter")
		styleCtx = sender.GetStyleContext()
		styleCtx.RemoveClass("active")
		styleCtx.RemoveClass("inactive")
		styleCtx.AddClass("hover")
	})
	button.SetOnLeave(func(sender *gtkhelper.Widget, event *gtkhelper.EventCrossing) {
		println("event.SetOnLeave")
		styleCtx = sender.GetStyleContext()
		styleCtx.RemoveClass("active")
		styleCtx.RemoveClass("inactive")
		styleCtx.AddClass("inactive")
	})
	button.SetOnClick(func(sender *gtkhelper.Widget, event *gtkhelper.EventButton) {
		println("event.SetOnClick")
		styleCtx = sender.GetStyleContext()
		styleCtx.RemoveClass("active")
		styleCtx.RemoveClass("inactive")
		styleCtx.AddClass("active")
		if tabButton.click != nil {
			tabButton.click()
		}
	})
	box := gtkhelper.NewBox(gtkhelper.ORIENTATION_HORIZONTAL, 4)
	box.SetVAlign(gtkhelper.ALIGN_CENTER)
	tabButton.box = box
	button.Add(box)

	icon := gtkhelper.NewImageFromIconName(iconName, gtkhelper.ICON_SIZE_MENU)
	icon.SetSizeRequest(16, 16)
	tabButton.icon = icon
	box.PackStart(icon, false, false, 4)

	label := gtkhelper.NewLabel(text)
	label.SetXAlign(0.0)
	label.SetEllipsize(gtkhelper.ELLIPSIZE_END)
	label.SetHExpand(true)
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
	box.PackEnd(closeBtn, false, false, 4)

	return tabButton
}
