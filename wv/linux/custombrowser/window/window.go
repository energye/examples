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
	windowWidth  = 1200
	windowHeight = 900
)

type BrowserWindow struct {
	lcl.TEngForm
	box lcl.IPanel
	// gtk3 window
	gtkWindow *gtkhelper.Window
	// toolbar
	gtkToolbar    *gtkhelper.HeaderBar
	closeBtn      *BrowserControlButton
	maxBtn        *BrowserControlButton
	minBtn        *BrowserControlButton
	addBrowserBtn *BrowserControlButton
	// browser
	browserBar    lcl.IPanel
	gtkBrowserBar *gtkhelper.Fixed
	gtkBrowserBox *gtkhelper.Box
	browses       []*Browser // 当前的chrom列表
	// 浏览器控制按钮
	backBtn    *BrowserControlButton
	forwardBtn *BrowserControlButton
	refreshBtn *BrowserControlButton
	// addr          *gtkhelper.Entry
	addr          lcl.IEdit
	addrRightIcon *BrowserControlButton
}

func (m *BrowserWindow) FormCreate(sender lcl.IObject) {
	png := lcl.NewPortableNetworkGraphic()
	png.LoadFromFile(assets.GetResourcePath("window-icon_64x64.png"))
	m.Icon().Assign(png)
	png.Free()
	m.SetWidth(int32(windowWidth))
	m.SetHeight(int32(windowHeight))
	m.SetDoubleBuffered(true)
	m.WorkAreaCenter()
	size := m.Constraints()
	size.SetMinWidth(400)
	size.SetMinHeight(200)

	m.box = lcl.NewPanel(m)
	m.box.SetParent(m)
	m.box.SetWidth(m.Width())
	m.box.SetHeight(m.Height())
	m.box.SetBevelOuter(types.BvNone)
	m.box.SetAlign(types.AlClient)
	m.box.SetDoubleBuffered(true)
	m.box.SetColor(bgColor)

	m.SetOnShow(func(sender lcl.IObject) {
		newBrowser := m.CreateBrowser("")
		m.OnCreateTabSheet(newBrowser)
		newBrowser.Create()
	})
	m.SetOnCloseQuery(func(sender lcl.IObject, canClose *bool) {

	})
	m.SetOnResize(func(sender lcl.IObject) {
		//fmt.Println("SetOnResize")
		m.UpdateToolbar()
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
	println("headerBoxWidget", browserBarFixed.TypeFromInstance().Name())
	m.browserBar = browserBar
	m.gtkBrowserBar = browserBarFixed

	// window move resize event
	m.gtkWindow.SetOnConfigure(func(sender *gtkhelper.Widget, event *gtkhelper.EventConfigure) bool {
		//if browse := m.getActiveBrowse(); browse != nil {
		//	browserHandle := lcl.PlatformHandle(browse.webviewParent.Handle())
		//	browserFixed := gtkhelper.ToFixed(unsafe.Pointer(browserHandle.Gtk3Widget()))
		//	fmt.Println(browserFixed.GetSizeRequest())
		//	fmt.Println(m.box.Width())
		//	//browserFixed.SetSizeRequest(int(m.box.Width()), int(m.box.Height()))
		//	//browserFixed.QueueDraw()
		//}
		return false
	})

	m.Toolbar()
	m.BrowserControlBar()
}

func (m *BrowserWindow) OnCreateTabSheet(currentBrowse *Browser) {
	m.browses = append(m.browses, currentBrowse)
	currentBrowse.windowId = int32(len(m.browses))
	m.AddTabSheetBtn(currentBrowse)
}

func (m *BrowserWindow) AddTabSheetBtn(currentBrowse *Browser) {
	tabSheetBtn := m.NewTabButton(assets.GetResourcePath("icon.png"), "新建标签页")
	currentBrowse.tabSheetBtn = tabSheetBtn
	currentBrowse.isActive = true
	tabSheetBtn.Active(true)
	tabSheetBtn.SetOnClick(func() {
		currentBrowse.updateTabSheetActive(true)
		m.updateOtherTabSheetNoActive(currentBrowse)
	})
	tabSheetBtn.SetOnCloseClick(func() {
		currentBrowse.CloseBrowser()
	})
	if m.gtkToolbar != nil {
		m.gtkToolbar.PackStart(tabSheetBtn.button)
		m.gtkToolbar.ShowAll() // call show
	}
	m.updateOtherTabSheetNoActive(currentBrowse)
}

// 获得当前激活的 chrom
func (m *BrowserWindow) getActiveBrowse() *Browser {
	var result *Browser
	for _, chrom := range m.browses {
		if chrom.isActive {
			result = chrom
			break
		}
	}
	return result
}

func (m *BrowserWindow) updateOtherTabSheetNoActive(currentBrowse *Browser) {
	for _, browse := range m.browses {
		if browse != currentBrowse {
			browse.updateTabSheetActive(false)
		}
	}
}

func (m *BrowserWindow) removeTabSheetBrowse(browse *Browser) {
	var isCloseCurrentActive bool
	if active := m.getActiveBrowse(); active != nil && active == browse {
		isCloseCurrentActive = true
	}
	// 删除当前chrom, 使用 windowId - 1 是当前 chrom 所在下标
	idx := browse.windowId - 1
	// 删除
	m.browses = append(m.browses[:idx], m.browses[idx+1:]...)
	// 重新设置每个 chromium 的 windowID, 在下次删除时能对应上
	for id, chrom := range m.browses {
		chrom.windowId = int32(id + 1)
	}
	if len(m.browses) > 0 {
		// 判断关闭时tabSheet是否为当前激活的
		// 如果是当前激活的，激活最后一个
		if isCloseCurrentActive {
			// 激活最后一个
			lastChrom := m.browses[len(m.browses)-1]
			lastChrom.updateTabSheetActive(true)
			// 其它的不激活
			m.updateOtherTabSheetNoActive(lastChrom)
		}
	} else {
		// 没有 chrom 清空和还原控制按钮、地址栏
		m.resetControlBtn()
	}
}

// 清空地址栏 和 还原控制按钮
func (m *BrowserWindow) resetControlBtn() {
	m.addr.SetText("")
	m.backBtn.SetEnable(false)
	m.forwardBtn.SetEnable(false)
}

func (m *BrowserWindow) updateRefreshBtn(browse *Browser, isLoading bool) {
	if browse.isActive {
		if isLoading {
			browse.mainWindow.refreshBtn.UpdateImage(assets.GetResourcePath("stop.png"))
		} else {
			browse.mainWindow.refreshBtn.UpdateImage(assets.GetResourcePath("refresh.png"))
		}
	}
}

func SetWidgetStyle(widget *gtkhelper.Widget, css string) {
	provider := gtkhelper.NewCssProvider()
	defer provider.Unref()
	provider.LoadFromData(css)
	context := widget.GetStyleContext()
	context.AddProvider(provider, gtkhelper.STYLE_PROVIDER_PRIORITY_APPLICATION)
}
