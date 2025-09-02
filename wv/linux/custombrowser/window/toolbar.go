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
	btn1 := m.NewTabButton("edit-delete-symbolic", "删除项目删除项目")
	headerBar.PackStart(btn1)

}

func (m *BrowserWindow) NewTabButton(iconName string, text string) *gtkhelper.Widget {
	event := gtkhelper.NewEventBox()
	event.SetHExpand(false)
	event.SetVExpand(false)
	event.SetSizeRequest(-1, 28)
	event.SetBorderWidth(0)
	event.SetVAlign(gtkhelper.ALIGN_CENTER)
	event.SetVisibleWindow(true)
	event.AddEvents(gtkhelper.POINTER_MOTION_MASK | gtkhelper.ENTER_NOTIFY_MASK | gtkhelper.LEAVE_NOTIFY_MASK)
	styleCtx := event.GetStyleContext()
	styleCtx.AddClass("tab")
	styleCtx.AddClass("active")
	event.SetOnEnter(func(sender *gtkhelper.Widget, event *gtkhelper.EventCrossing) {
		println("event.SetOnEnter")
		styleCtx = sender.GetStyleContext()
		styleCtx.RemoveClass("active")
		styleCtx.RemoveClass("inactive")
		styleCtx.AddClass("hover")
	})
	event.SetOnLeave(func(sender *gtkhelper.Widget, event *gtkhelper.EventCrossing) {
		println("event.SetOnLeave")
		styleCtx = sender.GetStyleContext()
		styleCtx.RemoveClass("active")
		styleCtx.RemoveClass("inactive")
		styleCtx.AddClass("inactive")
	})
	event.SetOnClick(func(sender *gtkhelper.Widget, event *gtkhelper.EventButton) {
		println("event.SetOnClick")
		styleCtx = sender.GetStyleContext()
		styleCtx.RemoveClass("active")
		styleCtx.RemoveClass("inactive")
		styleCtx.AddClass("active")
	})

	box := gtkhelper.NewBox(gtkhelper.ORIENTATION_HORIZONTAL, 4)
	event.Add(box)
	box.SetVAlign(gtkhelper.ALIGN_CENTER)

	icon := gtkhelper.NewImageFromIconName(iconName, gtkhelper.ICON_SIZE_MENU)
	icon.SetSizeRequest(16, 16)
	box.PackStart(icon, false, false, 4)

	label := gtkhelper.NewLabel(text)
	label.SetXAlign(0.0)
	label.SetEllipsize(gtkhelper.ELLIPSIZE_END)
	label.SetHExpand(true)
	box.PackStart(label, true, true, 0)

	closeBtn := gtkhelper.NewButton()
	closeBtnIcon := gtkhelper.NewImageFromIconName("window-close-symbolic", gtkhelper.ICON_SIZE_MENU)
	closeBtn.SetImage(closeBtnIcon)
	closeBtn.SetSizeRequest(16, 16)
	closeBtnStyleCtx := closeBtn.GetStyleContext()
	closeBtnStyleCtx.AddClass("tab-close-button")
	closeBtn.SetOpacity(0.7)
	closeBtn.SetFocusOnClick(false)
	closeBtn.SetOnClick(func(sender *gtkhelper.Widget) {
		println("close btn")
	})
	box.PackEnd(closeBtn, false, false, 4)

	return event.ToWidget()
}
