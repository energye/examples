package main

import (
	"embed"
	"fmt"
	"github.com/energye/assetserve"
	_ "github.com/energye/examples/syso"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
	"github.com/energye/wv/linux"
	"os"
	"path/filepath"
	"unsafe"
)

type TMainForm struct {
	lcl.IForm
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

/*
Now requires GTK >= 3.24.24 and Glib2.0 >= 2.66
GTK3: dpkg -l | grep libgtk-3-0
Glib: dpkg -l | grep libglib2.0
ldd --version
*/
func main() {
	//os.Setenv("JSC_SIGNAL_FOR_GC", "SIGUSR")
	httpServer()
	wv.Init(nil, resources)
	lcl.Application.Initialize()
	lcl.Application.SetScaled(true)
	mainForm.IForm = &lcl.TForm{}
	mainForm.url = "energy://demo.com/test.html"
	mainForm.isMainWindow = true
	lcl.Application.CreateForm(&mainForm)
	lcl.Application.Run()
}

func (m *TMainForm) FormCreate(sender lcl.IObject) {
	fmt.Println("main create")
	icod, _ := resources.ReadFile("assets/icon.ico")
	m.Icon().LoadFromBytes(icod)
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
			cookie := wv.WkCookieRef.NewCookie("webkit2-custom-cookie-key", "value-data-energy-custom-cookie", "www.baidu.com", "/", 100000)
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
			cookie := wv.WkCookieRef.NewCookie("webkit2-custom-cookie-key", "value-data-energy-custom-cookie", "www.baidu.com", "/", 100000)
			defer cookie.Free()
			cookieManager.DeleteCookie(cookie.Data())
		}
	})

	// webview parent
	m.webviewParent = wv.NewWkWebviewParent(m)
	m.webviewParent.SetParent(m)
	m.webviewParent.SetAlign(types.AlClient)
	m.webviewParent.SetParentDoubleBuffered(true)

	m.webview = wv.NewWkWebview(m)
	m.webview.SetOnContextMenu(func(sender wv.IObject, contextMenu wv.WebKitContextMenu, defaultAction wv.PWkAction) bool {
		fmt.Println("OnContextMenu defaultAction:", defaultAction)
		tempContextMenu := wv.NewWkContextMenu(contextMenu)
		defer tempContextMenu.Free()
		tempMenuItemSep := wv.WkContextMenuItemRef.NewSeparator()
		defer tempMenuItemSep.Free()
		tempContextMenu.Append(tempMenuItemSep.Data())
		tempMenuItemClose := wv.WkContextMenuItemRef.NewFromAction(defaultAction, "关闭", 10001)
		defer tempMenuItemClose.Free()
		tempContextMenu.Append(tempMenuItemClose.Data())
		return false
	})
	m.webview.SetOnContextMenuCommand(func(sender wv.IObject, menuID int32) {
		fmt.Println("OnContextMenuCommand menuID:", menuID)
		if menuID == 10001 {
			lcl.RunOnMainThreadAsync(func(id uint32) {
				m.Close()
			})
		}
	})
	m.webview.SetOnGetAcceptPolicyFinish(func(sender wv.IObject, policy wv.WebKitCookieAcceptPolicy, error_ string) {
		fmt.Println("OnGetAcceptPolicyFinish policy:", policy)
	})
	m.webview.SetOnGetCookiesFinish(func(sender wv.IObject, wkCookieList wv.PList, error_ string) {
		fmt.Println("OnGetCookiesFinish error_:", error_)
		tempCookieList := wv.NewWkCookieList(wkCookieList)
		defer tempCookieList.Free()
		size := tempCookieList.Length()
		fmt.Println("\tsize:", size)
		for i := 0; i < int(size); i++ {
			cookie := wv.NewWkCookie(tempCookieList.GetCookie(int32(i)))
			fmt.Println("\t cookie domain:", cookie.Domain())
			cookie.Free()
		}
	})
	m.webview.SetOnAddCookieFinish(func(sender wv.IObject, result bool, error_ string) {
		fmt.Println("OnAddCookieFinish result:", result, "error:", error_)
	})
	m.webview.SetOnDeleteCookieFinish(func(sender wv.IObject, result bool, error_ string) {
		fmt.Println("OnDeleteCookieFinish result:", result, "error:", error_)
	})
	m.webview.SetOnLoadChange(func(sender wv.IObject, wkLoadEvent wv.WebKitLoadEvent) {
		fmt.Println("OnLoadChange wkLoadEvent:", wkLoadEvent, "title:", m.webview.GetTitle())
		if wkLoadEvent == wv.WEBKIT_LOAD_FINISHED {
			if cookieManager == nil {
				cookieManager = m.webview.CookieManager()
				cookieManager.SetAcceptPolicy(wv.WEBKIT_COOKIE_POLICY_ACCEPT_ALWAYS)
			}
			title := m.webview.GetTitle()
			fmt.Println("title:", title)
			lcl.RunOnMainThreadAsync(func(id uint32) {
				m.SetCaption(title)
			})
		}
	})
	m.webview.SetOnWebProcessTerminated(func(sender wv.IObject, reason wv.WebKitWebProcessTerminationReason) {
		fmt.Println("OnWebProcessTerminated reason:", reason)
		if reason == wv.WEBKIT_WEB_PROCESS_TERMINATED_BY_API { //  call m.webview.TerminateWebProcess()
			lcl.RunOnMainThreadAsync(func(id uint32) {
				m.Close()
			})
		}
	})
	var headers = func(headers wv.PSoupMessageHeaders) {
		tempHeaders := wv.NewWkHeaders(headers)
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
	m.webview.SetOnDecidePolicy(func(sender wv.IObject, wkDecision wv.WebKitPolicyDecision, type_ wv.WebKitPolicyDecisionType) bool {
		fmt.Println("OnDecidePolicy type_:", type_)
		tempDecision := wv.NewWkNavigationPolicyDecision(wkDecision)
		defer tempDecision.Free()
		if type_ == wv.WEBKIT_POLICY_DECISION_TYPE_NEW_WINDOW_ACTION || type_ == wv.WEBKIT_POLICY_DECISION_TYPE_NAVIGATION_ACTION {
			tempNavigationAction := wv.NewWkNavigationAction(tempDecision.GetNavigationAction())
			defer tempNavigationAction.Free()
			tempURIRequest := wv.NewWkURIRequest(tempNavigationAction.GetRequest())
			defer tempURIRequest.Free()
			fmt.Println("URL:", tempURIRequest.URI())
			// new window
			if type_ == wv.WEBKIT_POLICY_DECISION_TYPE_NEW_WINDOW_ACTION {
				lcl.RunOnMainThreadAsync(func(id uint32) {
					window := NewWindow(tempURIRequest.URI())
					window.Show()
				})
			}
			headers(tempURIRequest.Headers())
		} else {
			tempResponsePolicyDecision := wv.NewWkResponsePolicyDecision(wkDecision)
			defer tempResponsePolicyDecision.Free()
			tempURIRequest := wv.NewWkURIRequest(tempResponsePolicyDecision.GetRequest())
			defer tempURIRequest.Free()
			fmt.Println("URL:", tempURIRequest.URI())
			headers(tempURIRequest.Headers())
		}
		return true
	})
	m.webview.SetOnExecuteScriptFinished(func(sender wv.IObject, jsValue wv.IWkJSValue) {
		fmt.Println("OnExecuteScriptFinished")
	})
	m.webview.SetOnURISchemeRequest(func(sender wv.IObject, wkURISchemeRequest wv.WebKitURISchemeRequest) {
		fmt.Println("OnURISchemeRequest")
		uriSchemeRequest := wv.NewWkURISchemeRequest(wkURISchemeRequest)
		defer uriSchemeRequest.Free()
		fmt.Println("uri:", uriSchemeRequest.Uri(), "method:", uriSchemeRequest.Method(), "path:", uriSchemeRequest.Path())
		path := uriSchemeRequest.Path()
		if path == "" {
			path = "index.html"
		}
		assetsPath := filepath.Join("assets", path)
		data, _ := resources.ReadFile(assetsPath)
		ins := wv.WkInputStreamRef.New(uintptr(unsafe.Pointer(&data[0])), int64(len(data)))
		uriSchemeRequest.Finish(ins.Data(), int64(len(data)), "text/html")
		headers := wv.NewWkHeaders(uriSchemeRequest.Headers())
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
	m.webview.SetOnProcessMessage(func(sender wv.IObject, jsValue wv.IWkJSValue, processId wv.TWkProcessId) {
		fmt.Println("OnProcessMessage value-type:", jsValue.ValueType())
		switch jsValue.ValueType() {
		case wv.JtString:
			fmt.Println("OnProcessMessageEvent 类型: [", jsValue.ValueType(), "] 返回结果: [", jsValue.StringValue(), "] JS异常: [", jsValue.ExceptionMessage(), "] processId: [", processId, "]")
		case wv.JtInteger:
			fmt.Println("OnProcessMessageEvent 类型: [", jsValue.ValueType(), "] 返回结果: [", jsValue.IntegerValue(), "] JS异常: [", jsValue.ExceptionMessage(), "] processId: [", processId, "]")
		case wv.JtBoolean:
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
		wkContext = wv.WkWebContextRef.Default()
		wkContext.RegisterURIScheme("energy", m.webview.AsSchemeRequestDelegate())
	}
	m.webview.EnabledDevtools(true)
	m.webview.RegisterScriptCode(`let test = {"name": "zhangsan"}`)
	m.webview.RegisterScriptMessageHandler("processMessage")

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
	form.IForm = &lcl.TForm{}
	lcl.Application.CreateForm(form)
	return form
}

func httpServer() {
	server := assetserve.NewAssetsHttpServer()
	server.PORT = 22022
	server.AssetsFSName = "assets" //必须设置目录名
	server.Assets = resources
	go server.StartHttpServer()
}
