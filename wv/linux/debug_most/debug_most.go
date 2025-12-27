package main

import (
	"embed"
	"fmt"
	"github.com/energye/assetserve"
	. "github.com/energye/examples/syso"
	"github.com/energye/examples/wv/assets"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
	wv "github.com/energye/wv/linux"
	wvTypes "github.com/energye/wv/types/linux"
	"os"
	"path/filepath"
	"unsafe"
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
	mainForm  TMainForm
	wkContext wv.IWkWebContext
)

//go:embed assets
var resources embed.FS

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
	//os.Setenv("JSC_SIGNAL_FOR_GC", "SIGUSR")
	httpServer()
	// linux webkit2 > gtk3
	os.Setenv("--ws", "gtk3")
	wv.Init(nil, resources)

	load := wv.NewLoader(nil)
	load.SetLoaderWebKit2DllPath("/usr/lib/x86_64-linux-gnu/libwebkit2gtk-4.0.so.37")
	load.SetLoaderJavascriptCoreDllPath("/usr/lib/x86_64-linux-gnu/libjavascriptcoregtk-4.0.so.18")
	load.SetLoaderSoupDllPath("/usr/lib/x86_64-linux-gnu/libsoup-2.4.so.1")
	if load.StartWebKit2() {
		lcl.Application.Initialize()
		lcl.Application.SetScaled(true)
		mainForm.url = "energy://demo.com/test.html"
		mainForm.isMainWindow = true
		lcl.Application.NewForm(&mainForm)
		lcl.Application.Run()
	}
}

func (m *TMainForm) FormCreate(sender lcl.IObject) {
	fmt.Println("main create")
	png := lcl.NewPortableNetworkGraphic()
	png.LoadFromFile(assets.GetResourcePath("window-icon_64x64.png"))
	m.Icon().Assign(png)
	png.Free()
	m.SetCaption("Main")
	// gtk3 需要设置一次较小的宽高, 然后在 OnShow 里设置默认宽高
	m.SetWidth(100)
	m.SetHeight(100)
	m.SetDoubleBuffered(true)

	mainMenu := lcl.NewMainMenu(m)
	item := lcl.NewMenuItem(m)
	item.SetCaption("文件(&F)")
	mainMenu.Items().Add(item)
	subItem := lcl.NewMenuItem(m)
	subItem.SetCaption("sub")
	item.Add(subItem)
	subItem.SetOnClick(func(sender lcl.IObject) {
		fmt.Println("sub-click")
	})

	var cookieManager wv.IWkCookieManager

	CookieManage := lcl.NewMenuItem(m)
	CookieManage.SetCaption("CookieManage")
	mainMenu.Items().Add(CookieManage)
	getAcceptPolicy := lcl.NewMenuItem(m)
	getAcceptPolicy.SetCaption("GetAcceptPolicy")
	CookieManage.Add(getAcceptPolicy)
	getAcceptPolicy.SetOnClick(func(sender lcl.IObject) {
		if cookieManager != nil {
			cookieManager.GetAcceptPolicy()
		}
	})
	addCookie := lcl.NewMenuItem(m)
	addCookie.SetCaption("AddCookie")
	CookieManage.Add(addCookie)
	addCookie.SetOnClick(func(sender lcl.IObject) {
		if cookieManager != nil {
			cookie := wv.Cookie.NewCookie("webkit2-custom-cookie-key", "value-data-energy-custom-cookie", "www.baidu.com", "/", 100000)
			defer cookie.Free()
			cookieManager.AddCookie(cookie.Data())
		}
	})
	getCookie := lcl.NewMenuItem(m)
	getCookie.SetCaption("GetCookie")
	CookieManage.Add(getCookie)
	getCookie.SetOnClick(func(sender lcl.IObject) {
		if cookieManager != nil {
			cookieManager.GetCookies("www.baidu.com")
		}
	})
	deleteCookie := lcl.NewMenuItem(m)
	deleteCookie.SetCaption("DeleteCookie")
	CookieManage.Add(deleteCookie)
	deleteCookie.SetOnClick(func(sender lcl.IObject) {
		fmt.Println("DeleteCookie")
		if cookieManager != nil {
			cookieManager.DeleteCookiesForDomain("www.baidu.com")
			// trigger OnDeleteCookieFinish
			cookie := wv.Cookie.NewCookie("webkit2-custom-cookie-key", "value-data-energy-custom-cookie", "www.baidu.com", "/", 100000)
			defer cookie.Free()
			cookieManager.DeleteCookie(cookie.Data())
		}
	})

	// webview parent
	m.webviewParent = wv.NewWebviewParent(m)
	m.webviewParent.SetParent(m)
	m.webviewParent.SetAlign(types.AlClient)
	m.webviewParent.SetParentDoubleBuffered(true)

	m.webview = wv.NewWebview(m)
	m.webview.SetOnContextMenu(func(sender lcl.IObject, contextMenu wvTypes.WebKitContextMenu, defaultAction wvTypes.PWkAction) bool {
		fmt.Println("OnContextMenu defaultAction:", defaultAction)
		tempContextMenu := wv.NewContextMenu(contextMenu)
		defer tempContextMenu.Free()
		tempMenuItemSep := wv.ContextMenuItem.NewSeparator()
		defer tempMenuItemSep.Free()
		tempContextMenu.Append(tempMenuItemSep.Data())
		tempMenuItemClose := wv.ContextMenuItem.NewFromAction(defaultAction, "关闭", 10001)
		defer tempMenuItemClose.Free()
		tempContextMenu.Append(tempMenuItemClose.Data())
		return false
	})
	m.webview.SetOnContextMenuCommand(func(sender lcl.IObject, menuID int32) {
		fmt.Println("OnContextMenuCommand menuID:", menuID)
		if menuID == 10001 {
			lcl.RunOnMainThreadAsync(func(id uint32) {
				m.Close()
			})
		}
	})
	m.webview.SetOnGetAcceptPolicyFinish(func(sender lcl.IObject, policy wvTypes.WebKitCookieAcceptPolicy, error_ string) {
		fmt.Println("OnGetAcceptPolicyFinish policy:", policy)
	})
	m.webview.SetOnGetCookiesFinish(func(sender lcl.IObject, cookieList wvTypes.PList, error_ string) {
		fmt.Println("OnGetCookiesFinish error_:", error_)
		tempCookieList := wv.NewCookieList(cookieList)
		defer tempCookieList.Free()
		size := tempCookieList.Length()
		fmt.Println("\tsize:", size)
		for i := 0; i < int(size); i++ {
			cookie := wv.NewCookie(tempCookieList.GetCookie(int32(i)))
			fmt.Println("\t cookie domain:", cookie.Domain())
			cookie.Free()
		}
	})
	m.webview.SetOnAddCookieFinish(func(sender lcl.IObject, result bool, error_ string) {
		fmt.Println("OnAddCookieFinish result:", result, "error:", error_)
	})
	m.webview.SetOnDeleteCookieFinish(func(sender lcl.IObject, result bool, error_ string) {
		fmt.Println("OnDeleteCookieFinish result:", result, "error:", error_)
	})
	m.webview.SetOnLoadChange(func(sender lcl.IObject, loadEvent wvTypes.WebKitLoadEvent) {
		fmt.Println("OnLoadChange wkLoadEvent:", loadEvent, "title:", m.webview.GetTitle())
		if loadEvent == wvTypes.WEBKIT_LOAD_FINISHED {
			if cookieManager == nil {
				cookieManager = m.webview.CookieManager()
				cookieManager.SetAcceptPolicy(wvTypes.WEBKIT_COOKIE_POLICY_ACCEPT_ALWAYS)
			}
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
	var headers = func(headers wvTypes.PSoupMessageHeaders) {
		tempHeaders := wv.NewHeaders(headers)
		defer tempHeaders.Free()
		headerList := tempHeaders.List()
		if headerList != nil {
			defer headerList.Free()
			count := headerList.Count()
			for i := 0; i < int(count); i++ {
				name := headerList.Names(int32(i))
				value := headerList.Values(name)
				fmt.Println("header name:", name, "value:", value)
			}
		}
	}
	m.webview.SetOnDecidePolicy(func(sender lcl.IObject, wkDecision wvTypes.WebKitPolicyDecision, type_ wvTypes.WebKitPolicyDecisionType) bool {
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
				lcl.RunOnMainThreadAsync(func(id uint32) {
					window := NewWindow(newWindowURL)
					window.Show()
				})
			}
			headers(tempURIRequest.Headers())
		} else {
			tempResponsePolicyDecision := wv.NewResponsePolicyDecision(wkDecision)
			defer tempResponsePolicyDecision.Free()
			tempURIRequest := wv.NewURIRequest(tempResponsePolicyDecision.GetRequest())
			defer tempURIRequest.Free()
			fmt.Println("URL:", tempURIRequest.URI())
			headers(tempURIRequest.Headers())
		}
		return true
	})
	m.webview.SetOnExecuteScriptFinished(func(sender lcl.IObject, jsValue wv.IWkJSValue) {
		fmt.Println("OnExecuteScriptFinished")
	})
	m.webview.SetOnURISchemeRequest(func(sender lcl.IObject, wkURISchemeRequest wvTypes.WebKitURISchemeRequest) {
		fmt.Println("OnURISchemeRequest")
		uriSchemeRequest := wv.NewURISchemeRequest(wkURISchemeRequest)
		defer uriSchemeRequest.Free()
		fmt.Println("uri:", uriSchemeRequest.Uri(), "method:", uriSchemeRequest.Method(), "path:", uriSchemeRequest.Path())
		path := uriSchemeRequest.Path()
		if path == "" {
			path = "index.html"
		}
		assetsPath := filepath.Join("assets", path)
		data, _ := resources.ReadFile(assetsPath)
		ins := wv.InputStream.New(uintptr(unsafe.Pointer(&data[0])), int64(len(data)))
		uriSchemeRequest.Finish(ins.Data(), int64(len(data)), "text/html")
		headers := wv.NewHeaders(uriSchemeRequest.Headers())
		headers.Append("test", "test")
		headList := headers.List()
		if headList != nil {
			fmt.Println("headList:", headList.Count())
			count := int(headList.Count())
			for i := 0; i < count; i++ {
				key := headList.Names(int32(i))
				val := headList.Values(key)
				fmt.Println("header name:", key, "value:", val)
			}
			headList.Free()
		}
		headers.Free()
	})
	var windowState = func(state int) {
		if state == 0 {
			m.SetWindowState(types.WsMinimized)
		} else if m.WindowState() == types.WsMaximized {
			m.SetWindowState(types.WsNormal)
		} else {
			m.SetWindowState(types.WsMaximized)
		}
	}
	m.webview.SetOnProcessMessage(func(sender lcl.IObject, jsValue wv.IWkJSValue, processId wvTypes.TWkProcessId) {
		fmt.Println("OnProcessMessage value-type:", jsValue.ValueType())
		switch jsValue.ValueType() {
		case wvTypes.JtString:
			fmt.Println("OnProcessMessageEvent 类型: [", jsValue.ValueType(), "] 返回结果: [", jsValue.StringValue(), "] JS异常: [", jsValue.ExceptionMessage(), "] processId: [", processId, "]")
		case wvTypes.JtInteger:
			fmt.Println("OnProcessMessageEvent 类型: [", jsValue.ValueType(), "] 返回结果: [", jsValue.IntegerValue(), "] JS异常: [", jsValue.ExceptionMessage(), "] processId: [", processId, "]")
		case wvTypes.JtBoolean:
			fmt.Println("OnProcessMessageEvent 类型: [", jsValue.ValueType(), "] 返回结果: [", jsValue.BooleanValue(), "] JS异常: [", jsValue.ExceptionMessage(), "] processId: [", processId, "]")
		}
		value := jsValue.StringValue()
		if value == "minimize" {
			lcl.RunOnMainThreadAsync(func(id uint32) {
				windowState(0)
			})
		} else if value == "maximize" {
			lcl.RunOnMainThreadAsync(func(id uint32) {
				windowState(1)
			})
		} else if value == "close" {
			lcl.RunOnMainThreadAsync(func(id uint32) {
				m.Close()
			})
		} else if value == "startdarg" {
			lcl.RunOnMainThreadAsync(func(id uint32) {
				m.webview.StartDrag(m)
				fmt.Println("startdarg end")
			})
		}
	})
	if wkContext == nil {
		wkContext = wv.WebContext.Default()
		wkContext.RegisterURIScheme("energy", m.webview.AsSchemeRequestDelegate())
	}
	m.webview.RegisterScriptCode(`let test = {"name": "zhangsan"}`)
	m.webview.RegisterScriptMessageHandler("processMessage")

	setting := wv.NewSettings()
	setting.SetEnableDeveloperExtras(true)
	setting.SetUserAgentWithApplicationDetails("energy.io", "3.0")
	setting.SetEnablePageCache(true)
	// SetHardwareAccelerationPolicy VMWare GPU ???不这样配置加载页面卡死，不知道是不是GPU问题
	// 需要动态判断当前系统环境是否支持？
	//setting.SetHardwareAccelerationPolicy(wvTypes.WEBKIT_HARDWARE_ACCELERATION_POLICY_NEVER)
	setting.SetHardwareAccelerationPolicy(wvTypes.WEBKIT_HARDWARE_ACCELERATION_POLICY_ALWAYS)
	m.webview.SetSettings(setting)

	// 所有webview事件或配置都在 CreateBrowser 之前
	m.webview.CreateBrowser()
	m.webviewParent.SetWebview(m.webview)

	m.SetOnShow(func(sender lcl.IObject) {
		fmt.Println("OnShow:", m.url)
		//m.webview.LoadURL("https://energye.github.io")
		//m.webview.LoadURL("http://localhost:22022/test.html")
		m.webview.LoadURL(m.url)
		// gtk3 需要设置一次较小的宽高, 然后在 OnShow 里设置默认宽高
		m.SetWidth(1024)
		m.SetHeight(600)
		m.ScreenCenter()
	})

	m.SetOnCloseQuery(func(sender lcl.IObject, canClose *bool) {
		*canClose = m.canClose
		fmt.Println("OnCloseQuery:", *canClose)
		if !m.canClose {
			m.canClose = true
			m.webview.Stop()
			m.webview.TerminateWebProcess()
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

func httpServer() {
	server := assetserve.NewAssetsHttpServer()
	server.PORT = 22022
	server.AssetsFSName = "assets" //必须设置目录名
	server.Assets = resources
	go server.StartHttpServer()
}
