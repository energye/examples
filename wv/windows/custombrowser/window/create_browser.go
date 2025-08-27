package window

import (
	"github.com/energye/examples/wv/assets"
	"github.com/energye/examples/wv/tool"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
	"github.com/energye/lcl/types/colors"
	wvTypes "github.com/energye/wv/types/windows"
	wv "github.com/energye/wv/windows"
	"net/url"
	"strings"
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
	newBrowser := &Browser{mainWindow: m, siteFavIcon: make(map[string]string)}

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
	if defaultUrl == "" {
		defaultHtmlPath := assets.GetResourcePath("default.html")
		newBrowser.browser.SetDefaultURL("file://" + defaultHtmlPath)
	} else {
		newBrowser.browser.SetDefaultURL(defaultUrl)
	}
	//m.browser.SetTargetCompatibleBrowserVersion("95.0.1020.44") // 设置
	println("TargetCompatibleBrowserVersion:", newBrowser.browser.TargetCompatibleBrowserVersion())
	newBrowser.browser.SetOnAfterCreated(func(sender lcl.IObject) {
		println("回调函数 WVBrowser => SetOnAfterCreated")
		newBrowser.windowParent.UpdateSize()
	})
	newBrowser.browser.SetOnDocumentTitleChanged(func(sender lcl.IObject) {
		title := newBrowser.browser.DocumentTitle()
		println("回调函数 WVBrowser => SetOnDocumentTitleChanged:", title)
		if newBrowser.tabSheetBtn != nil {
			if isDefaultResourceHTML(title) {
				title = "新建标签页"
			}

			lcl.RunOnMainThreadAsync(func(id uint32) {
				newBrowser.tabSheetBtn.SetCaption(title)
				newBrowser.tabSheetBtn.SetHint(title)
				newBrowser.tabSheetBtn.Invalidate()
			})
		}
		newBrowser.currentTitle = title
		if newBrowser.isActive {
			m.updateWindowCaption(title)
		}
	})

	var navBtns = func(aIsNavigating bool) {
		newBrowser.isLoading = aIsNavigating
		newBrowser.canGoBack = newBrowser.browser.CanGoBack()
		newBrowser.canGoForward = newBrowser.browser.CanGoForward()
		newBrowser.mainWindow.updateRefreshBtn(newBrowser, aIsNavigating)
		newBrowser.updateBrowserControlBtn()
	}

	newBrowser.browser.SetOnNotificationCloseRequested(func(sender lcl.IObject, notification wv.ICoreWebView2Notification, args lcl.IUnknown) {
		println("SetOnNotificationCloseRequested")
	})
	newBrowser.browser.SetOnNavigationStarting(func(sender lcl.IObject, webView wv.ICoreWebView2, args wv.ICoreWebView2NavigationStartingEventArgs) {
		navBtns(true)
		args = wv.NewCoreWebView2NavigationStartingEventArgs(args)
		targetURL := args.URI()
		if isDefaultResourceHTML(targetURL) {
			targetURL = ""
		}
		println("OnLoadStart URL:", targetURL)
		newBrowser.currentURL = targetURL
		if newBrowser.isActive {
			m.SetAddrText(targetURL)
		}
		args.Free()
	})
	newBrowser.browser.SetOnNavigationCompleted(func(sender lcl.IObject, webView wv.ICoreWebView2, args wv.ICoreWebView2NavigationCompletedEventArgs) {
		navBtns(false)
	})
	newBrowser.browser.SetOnNewWindowRequested(func(sender lcl.IObject, webView wv.ICoreWebView2, args wv.ICoreWebView2NewWindowRequestedEventArgs) {
		args = wv.NewCoreWebView2NewWindowRequestedEventArgs(args)
		// 阻止新窗口
		args.SetHandled(true)
		// 可以自己创建窗口
		targetURL := args.URI()
		lcl.RunOnMainThreadAsync(func(id uint32) {
			// 创建新的 tab
			newChromium := m.CreateBrowser(targetURL)
			m.OnChromiumCreateTabSheet(newChromium)
			newChromium.Create()
		})
		args.Free()
	})
	newBrowser.browser.SetOnFaviconChanged(func(sender lcl.IObject, webView wv.ICoreWebView2, args lcl.IUnknown) {
		webView = wv.NewCoreWebView2(webView)
		icoURL := webView.FaviconURI()
		println("SetOnFaviconChanged FaviconURI:", icoURL)
		ok := webView.GetFavicon(wvTypes.COREWEBVIEW2_FAVICON_IMAGE_FORMAT_PNG, newBrowser.browser)
		println("SetOnFaviconChanged FaviconURI ok:", ok)
		var host string
		if tempUrl, err := url.Parse(newBrowser.currentURL); err != nil {
			println("[ERROR] OnFavIconUrlChange ICON Parse URL:", err.Error())
			return
		} else {
			host = tempUrl.Host
		}

		if icoURL != "" {
			if tempURL, err := url.Parse(icoURL); err == nil {
				if _, ok := newBrowser.siteFavIcon[tempURL.Host]; !ok {
					tool.DownloadFavicon(SiteResource, host, icoURL, func(iconPath string) {
						println("DownloadFavicon:", iconPath)
						newBrowser.siteFavIcon[tempURL.Host] = iconPath
						// 在此保证更新一次图标到 tabSheetBtn
						lcl.RunOnMainThreadAsync(func(id uint32) {
							newBrowser.tabSheetBtn.SetIconFavorite(iconPath)
							newBrowser.tabSheetBtn.Invalidate()
						})
					})
				}
			}
		}
		webView.Free()
	})
	newBrowser.browser.SetOnGetFaviconCompleted(func(sender lcl.IObject, errorCode types.HRESULT, result lcl.IStreamAdapter) {
		println("SetOnGetFaviconCompleted errorCode:", errorCode)
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
	m.tabSheetBtn.Free()
	m.tabSheet.Free()
	m.mainWindow.removeTabSheetBrowse(m)
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

// 过滤 掉一些特定的 url , 在浏览器首页加载时使用的
func isDefaultResourceHTML(v string) bool {
	return v == "about:blank" || v == "DevTools" ||
		(strings.Index(v, "file://") != -1 && strings.Index(v, "resources") != -1) ||
		strings.Index(v, "default.html") != -1 ||
		strings.Index(v, "view-source:file://") != -1 ||
		strings.Index(v, "devtools://") != -1
}
