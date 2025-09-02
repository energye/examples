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
	mainWindowId               int32 // 窗口ID
	windowId                   int
	browses                    []*Browser  // 当前的chrom列表
	addChromBtn                *wg.TButton // 添加浏览器按钮
	// 浏览器控制按钮
	backBtn    *BrowserControlButton
	forwardBtn *BrowserControlButton
	refreshBtn *BrowserControlButton
	addr       *gtkhelper.Entry
	// 窗口关闭锁，一个一个关闭
	browserCloseLock    sync.Mutex
	isWindowButtonClose bool // 点击的窗口关闭按钮
	isChromCloseing     bool // 当前是否有正在关闭的 chrom
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
	m.SetDoubleBuffered(true)
	size := m.Constraints()
	size.SetMinWidth(400)
	size.SetMinHeight(200)

	m.SetOnShow(func(sender lcl.IObject) {
	})

	m.SetOnCloseQuery(func(sender lcl.IObject, canClose *bool) {

	})
	m.SetOnResize(func(sender lcl.IObject) {
		m.UpdateBrowserBounds()
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

func SetWidgetStyle(widget *gtkhelper.Widget, css string) {
	provider := gtkhelper.NewCssProvider()
	defer provider.Unref()
	provider.LoadFromData(css)
	context := widget.GetStyleContext()
	context.AddProvider(provider, gtkhelper.STYLE_PROVIDER_PRIORITY_APPLICATION)
}
