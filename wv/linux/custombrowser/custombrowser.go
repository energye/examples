package main

import (
	"fmt"
	. "github.com/energye/examples/syso"
	"github.com/energye/examples/wv/assets"
	"github.com/energye/lcl/api"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
	wv "github.com/energye/wv/linux"
	wvTypes "github.com/energye/wv/types/linux"
	"os"
)

type TMainForm struct {
	lcl.TEngForm
	url           string
	webviewParent wv.IWkWebviewParent
	webview       wv.IWkWebview
	canClose      bool
	isMainWindow  bool
}

var (
	mainForm TMainForm
)

func init() {
	TestLoadLibPath()
}

/*
Now requires GTK >= 3.24.24 and Glib2.0 >= 2.66
GTK3: dpkg -l | grep libgtk-3-0
Glib: dpkg -l | grep libglib2.0
ldd --version
*/
func main() {
	//window.CacheRoot = cacheRoot
	//window.SiteResource = siteResourceRoot
	wv.Init(nil, nil)
	lcl.Application.Initialize()
	lcl.Application.SetScaled(true)
	lcl.Application.NewForm(&mainForm)
	lcl.Application.Run()
}

func (m *TMainForm) FormCreate(sender lcl.IObject) {
	fmt.Println("main create")
	png := lcl.NewPortableNetworkGraphic()
	png.LoadFromFile(assets.GetResourcePath("window-icon_64x64.png"))
	m.Icon().Assign(png)
	png.Free()
	m.SetCaption("Main")
	// gtk3 需要设置一次较小的宽高, 然后在 OnShow 里设置默认宽高
	m.SetWidth(1024)
	m.SetHeight(600)
	m.ScreenCenter()
	m.SetDoubleBuffered(true)

	os.Setenv("WEBKIT_FORCE_COMPOSITING_MODE", "1")
	os.Setenv("WEBKIT_DISABLE_COMPOSITING_MODE", "0")
	os.Setenv("GDK_GL", "nvidia,mesa,sw") // 优先使用硬件GL， fallback到软件
	os.Setenv("WEBKIT_USE_SKIA", "1")     // 启用Skia渲染引擎（Ubuntu 22.04支持）

	// webview parent
	m.webviewParent = wv.NewWebviewParent(m)
	m.webviewParent.SetParent(m)
	m.webviewParent.SetAlign(types.AlClient)
	m.webviewParent.SetParentDoubleBuffered(true)

	var menuIdClose int32 = 10001
	m.webview = wv.NewWebview(m)
	m.webview.SetOnContextMenu(func(sender lcl.IObject, contextMenu wvTypes.WebKitContextMenu, defaultAction wvTypes.PWkAction) bool {
		fmt.Println("OnContextMenu defaultAction:", defaultAction)
		tempContextMenu := wv.NewContextMenu(contextMenu)
		defer tempContextMenu.Free()
		tempMenuItemSep := wv.ContextMenuItem.NewSeparator()
		defer tempMenuItemSep.Free()
		tempContextMenu.Append(tempMenuItemSep.Data())
		tempMenuItemClose := wv.ContextMenuItem.NewFromAction(defaultAction, "关闭", menuIdClose)
		defer tempMenuItemClose.Free()
		tempContextMenu.Append(tempMenuItemClose.Data())
		return false
	})
	m.webview.SetOnContextMenuCommand(func(sender lcl.IObject, menuID int32) {
		fmt.Println("OnContextMenuCommand menuID:", menuID)
		if menuID == menuIdClose {
			lcl.RunOnMainThreadAsync(func(id uint32) {
				m.Close()
			})
		}
	})
	m.webview.SetOnLoadChange(func(sender lcl.IObject, loadEvent wvTypes.WebKitLoadEvent) {
		fmt.Println("OnLoadChange wkLoadEvent:", loadEvent, "title:", m.webview.GetTitle())
		if loadEvent == wvTypes.WEBKIT_LOAD_FINISHED {
			title := m.webview.GetTitle()
			fmt.Println("title:", title)
			lcl.RunOnMainThreadAsync(func(id uint32) {
				m.SetCaption(title)
			})
		}
	})
	m.webview.SetOnWebProcessTerminated(func(sender lcl.IObject, reason wvTypes.WebKitWebProcessTerminationReason) {
		fmt.Println("OnWebProcessTerminated reason:", reason)
		if reason == wvTypes.WEBKIT_WEB_PROCESS_TERMINATED_BY_API { //  call m.webview.TerminateWebProcess()
			lcl.RunOnMainThreadAsync(func(id uint32) {
				m.Close()
			})
		}
	})
	m.webview.SetOnDecidePolicy(func(sender lcl.IObject, wkDecision wvTypes.WebKitPolicyDecision, type_ wvTypes.WebKitPolicyDecisionType) bool {
		fmt.Println("OnDecidePolicy type_:", type_, "IsMainThread:", api.MainThreadId() == api.CurrentThreadId())
		if type_ == wvTypes.WEBKIT_POLICY_DECISION_TYPE_NEW_WINDOW_ACTION || type_ == wvTypes.WEBKIT_POLICY_DECISION_TYPE_NAVIGATION_ACTION {
			tempDecision := wv.NewNavigationPolicyDecision(wkDecision)
			defer tempDecision.Free()
			// new window
			if type_ == wvTypes.WEBKIT_POLICY_DECISION_TYPE_NEW_WINDOW_ACTION {
				tempNavigationAction := wv.NewNavigationAction(tempDecision.GetNavigationAction())
				defer tempNavigationAction.Free()
				tempURIRequest := wv.NewURIRequest(tempNavigationAction.GetRequest())
				defer tempURIRequest.Free()
				newWindowURL := tempURIRequest.URI()
				fmt.Println("NewWindow URL:", newWindowURL)
				lcl.RunOnMainThreadAsync(func(id uint32) {
					window := NewWindow(newWindowURL)
					window.Show()
				})
			}
			tempDecision.Use()
		} else {
			// WEBKIT_POLICY_DECISION_TYPE_RESPONSE
			// 响应
			tempResponsePolicyDecision := wv.NewResponsePolicyDecision(wkDecision)
			defer tempResponsePolicyDecision.Free()
			tempURIRequest := wv.NewURIRequest(tempResponsePolicyDecision.GetRequest())
			defer tempURIRequest.Free()
			fmt.Println("Response URL:", tempURIRequest.URI())
			tempResponsePolicyDecision.Use()
		}
		return true
	})

	//m.webview.EnabledDevtools(true)
	wkContext := wv.WebContext.Default()
	wkContext.SetCacheModel(wvTypes.WEBKIT_CACHE_MODEL_DOCUMENT_VIEWER)

	setting := wv.NewSettings()
	setting.SetHardwareAccelerationPolicy(wvTypes.WEBKIT_HARDWARE_ACCELERATION_POLICY_ON_DEMAND)
	setting.SetEnableDeveloperExtras(true)
	setting.SetEnableWebgl(true)
	setting.SetEnableAccelerated2dCanvas(true)
	setting.SetEnablePageCache(true)
	m.webview.SetSettings(setting)

	// 所有webview事件或配置都在 CreateBrowser 之前
	m.webview.CreateBrowser()
	m.webviewParent.SetWebview(m.webview)

	m.SetOnActivate(func(sender lcl.IObject) {
	})
	m.SetOnShow(func(sender lcl.IObject) {
		//m.webview.LoadURL("https://element-plus.org/zh-CN/")
		if m.url == "" {
			m.webview.LoadURL("https://www.baidu.com")
		} else {
			m.webview.LoadURL(m.url)
		}
	})

	m.SetOnCloseQuery(func(sender lcl.IObject, canClose *bool) {
		*canClose = m.canClose
		fmt.Println("OnCloseQuery:", *canClose)
		if !m.canClose {
			m.canClose = true
			m.webview.Stop()
			m.webview.TerminateWebProcess()
			m.webview.Free()
			//m.webviewParent.FreeChild()
		}
		if *canClose && m.isMainWindow {
			os.Exit(0)
		}
	})
}

func (m *TMainForm) CreateParams(params *types.TCreateParams) {
	fmt.Println("调用此过程  TMainForm.CreateParams:", *params)
}

func NewWindow(url string) *TMainForm {
	var form = &TMainForm{url: url}
	lcl.Application.NewForm(form)
	return form
}
