package window

import "C"
import (
	"fmt"
	"github.com/energye/examples/wv/assets"
	"github.com/energye/examples/wv/linux/gtkhelper"
	"github.com/energye/lcl/api"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
	"github.com/energye/lcl/types/colors"
	"sync"
	"widget/wg"
)

var (
	CacheRoot    string
	SiteResource string
	Window       BrowserWindow
	bgColor      = colors.RGBToColor(56, 57, 60)
)

type BrowserWindow struct {
	lcl.TEngForm
	box          lcl.IPanel
	mainWindowId int32 // 窗口ID
	windowId     int
	browses      []*Browser  // 当前的chrom列表
	addChromBtn  *wg.TButton // 添加浏览器按钮
	// 浏览器控制按钮
	backBtn    *wg.TButton
	forwardBtn *wg.TButton
	refreshBtn *wg.TButton
	// 地址栏
	addr         lcl.IMemo
	addrRightBtn *wg.TButton
	// 窗口控制按钮
	minBtn   *wg.TButton
	maxBtn   *wg.TButton
	closeBtn *wg.TButton
	// 标题栏相关
	titleHeight           int32 // 标题栏高度
	borderWidth           int32 // 边框宽
	isDown                bool  // 鼠标按下和抬起
	isTitleBar, isDarging bool  // 窗口标题栏
	borderHT              uintptr
	// 窗口关闭锁，一个一个关闭
	browserCloseLock    sync.Mutex
	isWindowButtonClose bool // 点击的窗口关闭按钮
	isChromCloseing     bool // 当前是否有正在关闭的 chrom
	oldWndPrc           uintptr
	normalBounds        types.TRect
	windowState         types.TWindowState
}

func (m *BrowserWindow) FormCreate(sender lcl.IObject) {
	png := lcl.NewPortableNetworkGraphic()
	png.LoadFromFile(assets.GetResourcePath("window-icon_64x64.png"))
	m.Icon().Assign(png)
	png.Free()

	m.SetWidth(1024)
	m.SetHeight(600)
	m.ScreenCenter()
	//m.SetDoubleBuffered(true)

	m.SetOnShow(func(sender lcl.IObject) {
	})

	m.SetOnCloseQuery(func(sender lcl.IObject, canClose *bool) {

	})

	//edit := lcl.NewEdit(m)
	//edit.SetParent(m)
	//edit.SetOnKeyPress(func(sender lcl.IObject, key *uint16) {
	//	fmt.Println("SetOnKeyPress key:", *key)
	//})
	//edit.SetOnKeyUp(func(sender lcl.IObject, key *uint16, shift types.TShiftState) {
	//	fmt.Println("SetOnKeyUp key:", *key)
	//})
	//edit.SetOnKeyDown(func(sender lcl.IObject, key *uint16, shift types.TShiftState) {
	//
	//})
	m.Toolbar()
}

func (m *BrowserWindow) Toolbar() {
	addCSSStyles()

	gtkHandle := lcl.PlatformHandle(m.Handle())
	gtkWindowPtr := gtkHandle.Gtk3Window()
	fmt.Println("gtkWindowPtr:", gtkWindowPtr)

	headerBar, err := gtkhelper.NewHeaderBar()
	if err != nil {
		return
	}
	headerBar.SetShowCloseButton(true)
	headerBar.SetName("custom-headerbar")
	headerBar.SetVExpand(false)
	headerBar.SetVAlign(gtkhelper.ALIGN_CENTER)

	gtkWindow := gtkhelper.ToGtkWindow(uintptr(gtkWindowPtr))
	gtkWindow.SetTitlebar(headerBar)

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

	headerBar.PackStart(btn)

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
	headerBar.SetCustomTitle(entry)

	//
	btn1 := m.NewButton("edit-delete-symbolic", "删除项目")
	headerBar.PackEnd(btn1)
	//btn2 := m.NewButton("edit-delete-symbolic", "删除项目")
	//headerBar.PackEnd(btn2)
}

func (m *BrowserWindow) NewButton(iconName string, text string) *gtkhelper.Widget {
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
	//SetWidgetStyle(closeBtn.ToWidget(), closeBtnCss)
	box.PackEnd(closeBtn, false, false, 4)

	//styleContext := event.GetStyleContext()
	//SetWidgetStyle(event.ToWidget(), tabBtnCss)
	return event.ToWidget()
}

func SetWidgetStyle(widget *gtkhelper.Widget, css string) {
	provider := gtkhelper.NewCssProvider()
	defer provider.Unref()
	provider.LoadFromData(css)
	context := widget.GetStyleContext()
	context.AddProvider(provider, gtkhelper.STYLE_PROVIDER_PRIORITY_APPLICATION)
}

func addCSSStyles() {
	provider := gtkhelper.NewCssProvider()
	defer provider.Unref()
	css := `
.tab {
	background-color: #f0f0f0;
	border: 1px solid #dddddd;
	border-bottom: none;
	border-radius: 4px 4px 0 0;
	margin-top: 2px;
	padding: 4px 8px;
	color: #333333;
	transition: all 0.2s ease;
}

.tab.active {
	background-color: #ffffff;
	border-top: 2px solid #0a84ff;
	margin-top: 1px;
	color: #000000;
}

.tab.inactive {
	background-color: #f8f8f8;
}

.tab-close-button {
	border-radius: 2px;
	border: none;
	background: transparent;
	padding: 2px;
	min-width: 16px;
	min-height: 16px;
	transition: background-color 0.1s;
}

.tab-close-button:hover {
	background-color: rgba(0, 0, 0, 0.1);
}

.tab-close-button:active {
	background-color: rgba(0, 0, 0, 0.2);
}
	`
	provider.LoadFromData(css)

	screen := gtkhelper.ScreenGetDefault()
	gtkhelper.AddProviderForScreen(screen, provider, gtkhelper.STYLE_PROVIDER_PRIORITY_APPLICATION)
}
