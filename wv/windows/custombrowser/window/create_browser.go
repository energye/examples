package window

import (
	"fmt"
	"github.com/energye/examples/wv/assets"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
	"github.com/energye/lcl/types/colors"
	wv "github.com/energye/wv/windows"
	"widget/wg"
)

type Browser struct {
	mainWindow                         *BrowserWindow
	windowId                           int32 // 窗口ID
	windowParent                       wv.IWVWindowParent
	browser                            wv.IWVBrowser
	oldWndPrc                          uintptr
	tabSheetBtn                        *wg.TButton
	tabSheet                           lcl.IPanel
	isActive                           bool
	currentURL                         string
	currentTitle                       string
	siteFavIcon                        map[string]string
	isLoading, canGoBack, canGoForward bool
	isCloseing                         bool
}

func (m *BrowserWindow) CreateBrowser(defaultUrl string) *Browser {
	newBrowser := &Browser{}

	{
		newBrowser.tabSheet = lcl.NewPanel(m)
		newBrowser.tabSheet.SetParent(m.box)
		newBrowser.tabSheet.SetBevelOuter(types.BvNone)
		newBrowser.tabSheet.SetDoubleBuffered(true)
		newBrowser.tabSheet.SetAnchors(types.NewSet(types.AkLeft, types.AkTop, types.AkRight, types.AkBottom))
		newBrowser.tabSheet.SetTop(90)
		newBrowser.tabSheet.SetLeft(5)
		newBrowser.tabSheet.SetWidth(m.box.Width() - 10)
		newBrowser.tabSheet.SetHeight(m.box.Height() - (newBrowser.tabSheet.Top() + 5))
	}

	newBrowser.windowParent = wv.NewWindowParent(m)
	newBrowser.windowParent.SetParent(newBrowser.tabSheet)
	//重新调整browser窗口的Parent属性
	//重新设置了上边距，宽，高
	newBrowser.windowParent.SetAlign(types.AlClient) //重置对齐,默认是整个客户端

	newBrowser.browser = wv.NewBrowser(m.box)
	newBrowser.browser.SetDefaultURL(defaultUrl)
	//m.browser.SetTargetCompatibleBrowserVersion("95.0.1020.44") // 设置
	fmt.Println("TargetCompatibleBrowserVersion:", newBrowser.browser.TargetCompatibleBrowserVersion())
	newBrowser.browser.SetOnAfterCreated(func(sender lcl.IObject) {
		fmt.Println("回调函数 WVBrowser => SetOnAfterCreated")
		newBrowser.windowParent.UpdateSize()
	})
	newBrowser.browser.SetOnDocumentTitleChanged(func(sender lcl.IObject) {
		fmt.Println("回调函数 WVBrowser => SetOnDocumentTitleChanged:", newBrowser.browser.DocumentTitle())
	})
	newBrowser.browser.SetOnNewWindowRequested(func(sender lcl.IObject, webView wv.ICoreWebView2, args wv.ICoreWebView2NewWindowRequestedEventArgs) {
		args = wv.NewCoreWebView2NewWindowRequestedEventArgs(args)
		// 阻止新窗口
		args.SetHandled(true)
		// 可以自己创建窗口

		// 当前页面打开新链接
		//m.browser.Navigate(args.URI())
		//free
		args.Free()
	})
	// 设置browser到window parent
	newBrowser.windowParent.SetBrowser(newBrowser.browser)

	return newBrowser
}

func (m *Browser) Create() {
	m.browser.CreateBrowserWithHandleBool(m.windowParent.Handle(), true)
}

func (m *Browser) CloseBrowse() {
	if m.isCloseing {
		return
	}
	m.isCloseing = true
	m.browser.Stop()
	m.browser.Free()
	m.windowParent.Free()
}

func (m *Browser) resize(sender lcl.IObject) {
	if m.windowParent != nil {
		m.windowParent.UpdateSize()
	}
}

func (m *Browser) updateTabSheetActive(isActive bool) {
	if isActive {
		activeColor := colors.RGBToColor(86, 88, 93)
		m.tabSheetBtn.SetStartColor(activeColor)
		m.tabSheetBtn.SetEndColor(activeColor)
		m.tabSheet.SetVisible(true)
		m.isActive = true
		m.mainWindow.SetAddrText(m.currentURL)
		m.mainWindow.updateWindowCaption(m.currentTitle)
		m.resize(nil)
	} else {
		notActiveColor := bgColor //colors.RGBToColor(56, 57, 60)
		m.tabSheetBtn.SetStartColor(notActiveColor)
		m.tabSheetBtn.SetEndColor(notActiveColor)
		m.tabSheet.SetVisible(false)
		m.isActive = false
	}
	m.tabSheetBtn.Invalidate()
	// 根据当前 chromium 浏览器加载状态更新浏览器控制按钮
	m.updateBrowserControlBtn()
}

// 根据当前 chromium 浏览器加载状态更新浏览器控制按钮
func (m *Browser) updateBrowserControlBtn() {
	m.mainWindow.backBtn.IsDisable = !m.canGoBack
	m.mainWindow.forwardBtn.IsDisable = !m.canGoForward
	backDisable := !m.canGoBack
	forwardDisable := !m.canGoForward
	lcl.RunOnMainThreadAsync(func(id uint32) {
		// 退回按钮
		if backDisable {
			// 禁用
			m.mainWindow.backBtn.SetIcon(assets.GetResourcePath("back_disable.png"))
		} else {
			m.mainWindow.backBtn.SetIcon(assets.GetResourcePath("back.png"))
		}
		m.mainWindow.backBtn.Invalidate()
		// 前进按钮
		if forwardDisable {
			// 禁用
			m.mainWindow.forwardBtn.SetIcon(assets.GetResourcePath("forward_disable.png"))
		} else {
			m.mainWindow.forwardBtn.SetIcon(assets.GetResourcePath("forward.png"))
		}
		m.mainWindow.forwardBtn.Invalidate()
	})
}
