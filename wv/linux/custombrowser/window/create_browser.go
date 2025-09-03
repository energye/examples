package window

import (
	"fmt"
	"github.com/energye/examples/wv/assets"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
	wv "github.com/energye/wv/linux"
	wvTypes "github.com/energye/wv/types/linux"
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
		fmt.Println("OnLoadChange wkLoadEvent:", loadEvent, "title:", title)
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
		fmt.Println("OnDecidePolicy type_:", type_)
		tempDecision := wv.NewNavigationPolicyDecision(wkDecision)
		defer tempDecision.Free()
		if type_ == wvTypes.WEBKIT_POLICY_DECISION_TYPE_NEW_WINDOW_ACTION || type_ == wvTypes.WEBKIT_POLICY_DECISION_TYPE_NAVIGATION_ACTION {
			tempNavigationAction := wv.NewNavigationAction(tempDecision.GetNavigationAction())
			defer tempNavigationAction.Free()
			tempURIRequest := wv.NewURIRequest(tempNavigationAction.GetRequest())
			defer tempURIRequest.Free()
			newWindowURL := tempURIRequest.URI()
			fmt.Println("NewWindow URL:", newWindowURL)
			// new window
			if type_ == wvTypes.WEBKIT_POLICY_DECISION_TYPE_NEW_WINDOW_ACTION {
				//lcl.RunOnMainThreadAsync(func(id uint32) {
				//	window := NewWindow(newWindowURL)
				//	window.Show()
				//})
			}
		} else {
			tempResponsePolicyDecision := wv.NewResponsePolicyDecision(wkDecision)
			defer tempResponsePolicyDecision.Free()
			tempURIRequest := wv.NewURIRequest(tempResponsePolicyDecision.GetRequest())
			defer tempURIRequest.Free()
			fmt.Println("URL:", tempURIRequest.URI())
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
	if isActive {
		m.isActive = true
	} else {
		m.isActive = false
	}
	m.webviewParent.SetVisible(m.isActive)
}
