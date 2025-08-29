package window

import "C"
import (
	"fmt"
	"github.com/energye/examples/wv/assets"
	"github.com/energye/examples/wv/linux/gtkhelper"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
	"github.com/energye/lcl/types/colors"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
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
	m.SetDoubleBuffered(true)

	gtkHandle := lcl.PlatformHandle(m.Handle())
	gtkWindowPtr := gtkHandle.Gtk3Window()
	fmt.Println("gtkWindowPtr:", gtkWindowPtr)

	headerBar, err := gtkhelper.HeaderBarNew()
	if err != nil {
		return
	}
	headerBar.SetShowCloseButton(true)

	gtkWindow := gtkhelper.ToGtkWindow(uintptr(gtkWindowPtr))
	gtkWindow.SetTitlebar(headerBar)

	btn, _ := gtkhelper.ButtonNewWithLabel("button")
	btn.SetOnClick()
	headerBar.PackStart(btn)

	m.SetOnShow(func(sender lcl.IObject) {
	})

	m.SetOnCloseQuery(func(sender lcl.IObject, canClose *bool) {

	})

}

func ToGtkWindow(gtkWindow uintptr) *gtk.Window {
	cObj := glib.ToGObject(unsafe.Pointer(gtkWindow))
	obj := glib.Object{GObject: cObj}
	window := new(gtk.Window)
	window.InitiallyUnowned = glib.InitiallyUnowned{Object: &obj}
	return window
}
