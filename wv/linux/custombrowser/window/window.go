package window

import "C"
import (
	"fmt"
	"github.com/energye/examples/wv/assets"
	"github.com/energye/examples/wv/linux/gtkhelper"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
	"github.com/energye/lcl/types/colors"
	"unsafe"
)

var (
	CacheRoot    string
	SiteResource string
	Window       BrowserWindow
	bgColor      = colors.RGBToColor(56, 57, 60)
	windowWidth  = 1024
	windowHeight = 768
)

type BrowserWindow struct {
	lcl.TEngForm
	box lcl.IPanel
	// gtk3 window
	gtkWindow     *gtkhelper.Window
	browserBar    lcl.IPanel
	gtkBrowserBar *gtkhelper.Fixed
	gtkBrowserBox *gtkhelper.Box
	browses       []*Browser            // 当前的chrom列表
	addBrowserBtn *BrowserControlButton // 添加浏览器按钮
	// 浏览器控制按钮
	backBtn       *BrowserControlButton
	forwardBtn    *BrowserControlButton
	refreshBtn    *BrowserControlButton
	addr          *gtkhelper.Entry
	addrRightIcon *BrowserControlButton
	//
}

func (m *BrowserWindow) FormCreate(sender lcl.IObject) {
	png := lcl.NewPortableNetworkGraphic()
	png.LoadFromFile(assets.GetResourcePath("window-icon_64x64.png"))
	m.Icon().Assign(png)
	png.Free()
	m.SetWidth(int32(windowWidth))
	m.SetHeight(int32(windowHeight))
	m.ScreenCenter()
	m.SetDoubleBuffered(true)
	size := m.Constraints()
	size.SetMinWidth(400)
	size.SetMinHeight(200)

	m.box = lcl.NewPanel(m)
	m.box.SetParent(m)
	m.box.SetWidth(int32(windowWidth))
	m.box.SetHeight(int32(windowHeight))
	m.box.SetBevelOuter(types.BvNone)
	m.box.SetAlign(types.AlClient)
	m.box.SetColor(bgColor)

	m.SetOnShow(func(sender lcl.IObject) {
		//go func() {
		//	time.Sleep(500)
		//	lcl.RunOnMainThreadAsync(func(id uint32) {
		//		m.SetWidth(1024)
		//		m.SetHeight(768)
		//		m.ScreenCenter()
		//	})
		//}()
	})
	m.SetOnCloseQuery(func(sender lcl.IObject, canClose *bool) {

	})
	m.SetOnResize(func(sender lcl.IObject) {
		//fmt.Println("SetOnResize")
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
	browserBar := lcl.NewPanel(m)
	browserBar.SetParent(m.box)
	browserBar.SetHeight(48)
	browserBar.SetWidth(m.box.Width())
	browserBar.SetBevelOuter(types.BvNone)
	browserBar.SetAnchors(types.NewSet(types.AkLeft, types.AkTop, types.AkRight))
	browserBarHandle := lcl.PlatformHandle(browserBar.Handle())
	browserBarFixed := gtkhelper.ToFixed(unsafe.Pointer(browserBarHandle.Gtk3Widget()))
	width, height := browserBarFixed.GetSizeRequest()
	println("headerBoxWidget", width, height, browserBarFixed.TypeFromInstance().Name())
	m.browserBar = browserBar
	m.gtkBrowserBar = browserBarFixed
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
