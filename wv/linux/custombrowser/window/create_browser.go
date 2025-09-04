package window

import (
	"fmt"
	"github.com/energye/examples/wv/assets"
	"github.com/energye/lcl/api"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
	wv "github.com/energye/wv/linux"
	wvTypes "github.com/energye/wv/types/linux"
	"strings"
)

type Browser struct {
	mainWindow                         *BrowserWindow
	windowId                           int32 // 窗口ID
	webviewParent                      wv.IWkWebviewParent
	webview                            wv.IWkWebview
	tabSheetBtn                        *TabButton
	isActive                           bool
	currentURL                         string
	currentTitle                       string
	siteFavIcon                        map[string]string
	isLoading, canGoBack, canGoForward bool
	isCloseing                         bool
}

func (m *BrowserWindow) CreateBrowser(defaultUrl string) *Browser {
	newBrowser := new(Browser)
	newBrowser.mainWindow = m
	if defaultUrl == "" {
		defaultHtmlPath := assets.GetResourcePath("default.html")
		newBrowser.currentURL = "file://" + defaultHtmlPath
	} else {
		newBrowser.currentURL = defaultUrl
	}

	newBrowser.webviewParent = wv.NewWebviewParent(m)
	newBrowser.webviewParent.SetParent(m.box)
	newBrowser.webviewParent.SetTop(m.browserBar.Height())
	newBrowser.webviewParent.SetLeft(5)
	newBrowser.webviewParent.SetWidth(m.box.Width() - 10)
	newBrowser.webviewParent.SetHeight(m.box.Height() - (m.browserBar.Height() + 5))
	newBrowser.webviewParent.SetAnchors(types.NewSet(types.AkLeft, types.AkTop, types.AkRight, types.AkBottom))
	newBrowser.webviewParent.SetDoubleBuffered(true)

	println("CreateBrowser top:", newBrowser.webviewParent.Top(), "left:", newBrowser.webviewParent.Left(), "width:", newBrowser.webviewParent.Width(), "height:", newBrowser.webviewParent.Height())

	newBrowser.webview = wv.NewWebview(m)
	newBrowser.webview.SetOnLoadChange(func(sender lcl.IObject, loadEvent wvTypes.WebKitLoadEvent) {
		title := newBrowser.webview.GetTitle()
		if title != "" {
			if isDefaultResourceHTML(title) {
				title = "新建标签页"
			}
			if newBrowser.isActive {
				newBrowser.tabSheetBtn.SetTitle(title)
			}
		}
		newBrowser.currentTitle = title
		fmt.Println("OnLoadChange wkLoadEvent:", loadEvent, "title:", title, "isMainThread:", api.MainThreadId() == api.CurrentThreadId())
		if loadEvent == wvTypes.WEBKIT_LOAD_FINISHED {
			fmt.Println("title:", title)
		}
	})
	newBrowser.webview.SetOnWebProcessTerminated(func(sender lcl.IObject, reason wvTypes.WebKitWebProcessTerminationReason) {
		fmt.Println("OnWebProcessTerminated reason:", reason)
		if reason == wvTypes.WEBKIT_WEB_PROCESS_TERMINATED_BY_API { //  call m.webview.TerminateWebProcess()

		}
	})
	newBrowser.webview.SetOnDecidePolicy(func(sender lcl.IObject, wkDecision wvTypes.WebKitPolicyDecision, type_ wvTypes.WebKitPolicyDecisionType) bool {
		fmt.Println("OnDecidePolicy type_:", type_, "isMainThread:", api.MainThreadId() == api.CurrentThreadId())
		tempDecision := wv.NewNavigationPolicyDecision(wkDecision)
		defer tempDecision.Free()
		var targetURL string
		if type_ == wvTypes.WEBKIT_POLICY_DECISION_TYPE_NEW_WINDOW_ACTION || type_ == wvTypes.WEBKIT_POLICY_DECISION_TYPE_NAVIGATION_ACTION {
			tempNavigationAction := wv.NewNavigationAction(tempDecision.GetNavigationAction())
			defer tempNavigationAction.Free()
			tempURIRequest := wv.NewURIRequest(tempNavigationAction.GetRequest())
			defer tempURIRequest.Free()
			targetURL = tempURIRequest.URI()
			fmt.Println("NewWindow URL:", targetURL)
			// new window
			if type_ == wvTypes.WEBKIT_POLICY_DECISION_TYPE_NEW_WINDOW_ACTION {
				lcl.RunOnMainThreadAsync(func(id uint32) {
					newBrowser := m.CreateBrowser(targetURL)
					m.OnCreateTabSheet(newBrowser)
					newBrowser.Create()
				})
			} else {
				if isDefaultResourceHTML(targetURL) {
					targetURL = ""
				}
				newBrowser.currentURL = targetURL
				if newBrowser.isActive {
					newBrowser.mainWindow.addr.SetText(targetURL)
				}
			}
		} else {
			tempResponsePolicyDecision := wv.NewResponsePolicyDecision(wkDecision)
			defer tempResponsePolicyDecision.Free()
			tempURIRequest := wv.NewURIRequest(tempResponsePolicyDecision.GetRequest())
			defer tempURIRequest.Free()
			targetURL = tempURIRequest.URI()
			fmt.Println("URL:", targetURL)
		}
		return true
	})
	setting := wv.NewSettings()
	setting.SetHardwareAccelerationPolicy(wvTypes.WEBKIT_HARDWARE_ACCELERATION_POLICY_ALWAYS)
	newBrowser.webview.SetSettings(setting)
	return newBrowser
}

func (m *Browser) Create() {
	if m.webview == nil {
		return
	}
	m.webview.CreateBrowser()
	m.webviewParent.SetWebview(m.webview)
	m.webview.LoadURL(m.currentURL)
}

func (m *Browser) updateTabSheetActive(isActive bool) {
	m.isActive = isActive
	if isActive {
		m.Show()
	} else {
		m.Hide()
	}
	m.tabSheetBtn.Active(isActive)
}

func (m *Browser) Show() {
	m.webviewParent.SetVisible(true)
}
func (m *Browser) Hide() {
	m.webviewParent.SetVisible(false)
}

// 过滤 掉一些特定的 url , 在浏览器首页加载时使用的
func isDefaultResourceHTML(v string) bool {
	return v == "about:blank" || v == "DevTools" ||
		(strings.Index(v, "file://") != -1 && strings.Index(v, "resources") != -1) ||
		strings.Index(v, "default.html") != -1 ||
		strings.Index(v, "view-source:file://") != -1 ||
		strings.Index(v, "devtools://") != -1
}
