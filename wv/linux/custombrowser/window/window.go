package window

import "C"
import (
	"fmt"
	"github.com/energye/examples/wv/assets"
	"github.com/energye/examples/wv/linux/gtkhelper"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
	"github.com/energye/lcl/types/colors"
	"sync"
	"unsafe"
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
	// gtk3 window
	gtkWindow                  *gtkhelper.Window
	gtkControlBrowserBarWidget *gtkhelper.Fixed
	controlBrowserBar          lcl.IPanel
	box                        lcl.IPanel
	mainWindowId               int32 // 窗口ID
	windowId                   int
	browses                    []*Browser  // 当前的chrom列表
	addChromBtn                *wg.TButton // 添加浏览器按钮
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

	// Global CSS Style
	addCSSStyles()

	// Window
	gtkHandle := lcl.PlatformHandle(m.Handle())
	gtkWindowPtr := gtkHandle.Gtk3Window()
	m.gtkWindow = gtkhelper.ToGtkWindow(uintptr(gtkWindowPtr))
	fmt.Println("gtkWindowPtr:", gtkWindowPtr)
	// Browser Control
	// 把 LCL 的Panel转为 gtk3 控件
	controlBrowserBar := lcl.NewPanel(m)
	controlBrowserBar.SetParent(m)
	controlBrowserBar.SetHeight(48)
	controlBrowserBar.SetWidth(m.Width())
	controlBrowserBar.SetBevelOuter(types.BvNone)
	//controlBrowserBar.SetColor(colors.ClBlue)
	controlBrowserBar.SetAnchors(types.NewSet(types.AkLeft, types.AkTop, types.AkRight))
	controlBrowserBarHandle := lcl.PlatformHandle(controlBrowserBar.Handle())
	controlBrowserBarWidget := gtkhelper.ToFixed(unsafe.Pointer(controlBrowserBarHandle.Gtk3Widget()))
	width, height := controlBrowserBarWidget.GetSizeRequest()
	println("headerBoxWidget", width, height, controlBrowserBarWidget.TypeFromInstance().Name())
	m.controlBrowserBar = controlBrowserBar
	m.gtkControlBrowserBarWidget = controlBrowserBarWidget
	m.Toolbar()
	m.BrowserControlBar()
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
	closeBtn.SetFocusOnClick(false)
	closeBtn.SetOnClick(func(sender *gtkhelper.Widget) {
		println("close btn")
	})
	box.PackEnd(closeBtn, false, false, 4)

	return event.ToWidget()
}

func SetWidgetStyle(widget *gtkhelper.Widget, css string) {
	provider := gtkhelper.NewCssProvider()
	defer provider.Unref()
	provider.LoadFromData(css)
	context := widget.GetStyleContext()
	context.AddProvider(provider, gtkhelper.STYLE_PROVIDER_PRIORITY_APPLICATION)
}
