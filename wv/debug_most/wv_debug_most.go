package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"github.com/energye/energy/v3/ipc"
	"github.com/energye/energy/v3/ipc/callback"
	_ "github.com/energye/examples/syso"
	"github.com/energye/examples/wv/debug_most/contextmenu"
	"github.com/energye/examples/wv/debug_most/cookie"
	"github.com/energye/examples/wv/debug_most/devtools"
	"github.com/energye/examples/wv/debug_most/scheme"
	"github.com/energye/examples/wv/debug_most/utils"
	"github.com/energye/lcl/api/exception"
	"github.com/energye/lcl/api/libname"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/tools/exec"
	"github.com/energye/lcl/types"
	"github.com/energye/lcl/types/messages"
	"github.com/energye/wv/wv"
	"path/filepath"
)

type TMainForm struct {
	lcl.TForm
	windowParent wv.IWVWindowParent
	browser      wv.IWVBrowser
}

var MainForm TMainForm
var load wv.IWVLoader

//go:embed assets
var assets embed.FS

func main() {
	utils.Assets = assets
	fmt.Println("Go ENERGY Run Main")
	wv.Init(nil, nil)
	exception.SetOnException(func(funcName, message string) {
		fmt.Println("ERROR funcName:", funcName, "message:", message)
	})
	// GlobalWebView2Loader
	load = wv.GlobalWebView2Loader()
	liblcl := libname.LibName
	webView2Loader, _ := filepath.Split(liblcl)
	webView2Loader = filepath.Join(webView2Loader, "WebView2Loader.dll")
	fmt.Println("当前目录:", exec.CurrentDir)
	fmt.Println("liblcl.dll目录:", liblcl)
	fmt.Println("WebView2Loader.dll目录:", webView2Loader)
	fmt.Println("用户缓存目录:", filepath.Join(exec.CurrentDir, "EnergyCache"))
	fmt.Println("自定义URL协议头:", scheme.SchemeName)
	load.SetUserDataFolder(filepath.Join(exec.CurrentDir, "EnergyCache"))
	load.SetLoaderDllPath(webView2Loader)
	scheme.LoaderOnCustomSchemes(load)
	r := load.StartWebView2()
	fmt.Println("StartWebView2", r)
	lcl.Application.Initialize()
	lcl.Application.SetMainFormOnTaskBar(true)
	lcl.Application.CreateForm(&MainForm)
	lcl.Application.Run()
}

type ProcessMessage struct {
	Name string      `json:"n"`
	Data interface{} `json:"d"`
	Id   int         `json:"i"`
}

func (m *TMainForm) FormCreate(sender lcl.IObject) {
	ipc.On("ipc-test", func(context callback.IEvent) {

	})

	m.SetCaption("Energy3.0 - webview2 simple")
	m.SetPosition(types.PoScreenCenter)
	m.SetWidth(1024)
	m.SetHeight(768)
	m.SetDoubleBuffered(true)
	back, forward, stop, refresh, addr := controlUI(m)

	m.windowParent = wv.NewWVWindowParent(m)
	m.windowParent.SetParent(m)
	//m.windowParent.SetWidth(200)
	//m.windowParent.SetHeight(200)
	//重新调整browser窗口的Parent属性
	//重新设置了上边距，宽，高
	m.windowParent.SetAlign(types.AlCustom) //重置对齐,默认是整个客户端
	m.windowParent.SetTop(30)
	m.windowParent.SetHeight(m.Height() - 25)
	m.windowParent.SetWidth(m.Width())
	m.windowParent.SetAnchors(types.NewSet(types.AkLeft, types.AkTop, types.AkRight, types.AkBottom))
	//m.windowParent.SetAlign(types.AlClient)

	m.browser = wv.NewWVBrowser(m)
	m.windowParent.SetBrowser(m.browser)
	m.browser.SetDefaultURL("myscheme://domain/index.html")
	//m.browser.SetTargetCompatibleBrowserVersion("95.0.1020.44")
	fmt.Println("TargetCompatibleBrowserVersion:", m.browser.TargetCompatibleBrowserVersion())
	contextmenu.Contextmenu(m, m.browser)
	devtools.DevTools(m.browser)
	cookie.Cookie(m.browser)
	scheme.WebResourceRequested(m.browser)
	m.browser.SetOnAfterCreated(func(sender lcl.IObject) {
		fmt.Println("回调函数 WVBrowser => SetOnAfterCreated")
		m.windowParent.UpdateSize()
		scheme.OnAfterCreated(m.browser)
		// 1. 先植入 ipc js
		//fmt.Println("AddScriptToExecuteOnDocumentCreated:", string(utils.IPCJavaScript))
		m.browser.CoreWebView2().AddScriptToExecuteOnDocumentCreated(string(utils.IPCJavaScript), m.browser)
		// 禁用devtools, 不能通过浏览默认方式打开，需要自己手动打开
		settings := m.browser.CoreWebView2Settings()
		settings.SetAreDevToolsEnabled(false)
	})
	var navBtns = func(aIsNavigating bool) {
		back.SetEnabled(m.browser.CanGoBack())
		forward.SetEnabled(m.browser.CanGoForward())
		refresh.SetEnabled(!aIsNavigating)
		stop.SetEnabled(aIsNavigating)
	}
	m.browser.SetOnExecuteScriptCompleted(func(sender wv.IObject, errorCode int32, resulIObjectAsJson string, executionID int32) {
		fmt.Println("回调函数 WVBrowser => SetOnExecuteScriptCompleted errorCode:", errorCode,
			"executionID:", executionID, "resulIObjectAsJson:", resulIObjectAsJson)
	})
	m.browser.SetOnNavigationStarting(func(sender wv.IObject, webView wv.ICoreWebView2, args wv.ICoreWebView2NavigationStartingEventArgs) {
		fmt.Println("回调函数 WVBrowser => SetOnNavigationStarting")
		navBtns(true)
		args = wv.NewCoreWebView2NavigationStartingEventArgs(args)
		defer args.Free()
		fmt.Println("url:", args.URI())
		headers := wv.NewCoreWebView2HttpRequestHeaders(args.RequestHeaders())
		defer headers.Free()
		headers.SetHeader("custom-energy", "custom-value")
		//var headersIter wv.ICoreWebView2HttpHeadersCollectionIterator
		//headers.GetHeaders(&headersIter)
		iterator := wv.NewCoreWebView2HttpHeadersCollectionIterator(headers.Iterator())
		if iterator != nil {
			defer iterator.Free()
			var (
				name  string
				value string
			)

			for {
				iterator.GetCurrentHeader(&name, &value)
				fmt.Println("\tname:", name, "value:", value)
				if !iterator.MoveNext() {
					break
				}
			}
		}

	})
	// 进程消息
	m.browser.SetOnWebMessageReceived(func(sender wv.IObject, webView wv.ICoreWebView2, args wv.ICoreWebView2WebMessageReceivedEventArgs) {
		fmt.Println("回调函数 WVBrowser => SetOnWebMessageReceived")
		args = wv.NewCoreWebView2WebMessageReceivedEventArgs(args)
		defer args.Free()
		var message ProcessMessage
		err := json.Unmarshal([]byte(args.WebMessageAsString()), &message)
		if err != nil {

		} else {

		}
		fmt.Printf("\tmessage: %+v\n", message)
		if message.Name == "emit-name" {
			// messageId 不等于0表示有回调函数需要执行
			// 需要回调一个消息
			if message.Id != 0 {
				message.Name = "" // 不需要事件名
				message.Data = "返回值"
				jsonData, _ := json.Marshal(message)
				m.browser.PostWebMessageAsString(string(jsonData))
			}
		} else if message.Name == "showDevtools" {
			devtools.OpenDevtools(m.browser)
		} else if message.Name == "executeDevToolsMethod" {
			devtools.ExecuteDevToolsMethod(m.browser)
		} else if message.Name == "cookieVisited" {
			m.browser.GetCookies("")
		} else if message.Name == "cookieDelete" {
			m.browser.DeleteAllCookies()
		} else if message.Name == "setCookie" {
			newCookie := m.browser.CreateCookie("mycustomcookie", "123456", "example.com", "/")
			m.browser.AddOrUpdateCookie(newCookie)
		}
	})
	m.browser.SetOnSourceChanged(func(sender wv.IObject, webView wv.ICoreWebView2, args wv.ICoreWebView2SourceChangedEventArgs) {
		fmt.Println("回调函数 WVBrowser => SetOnSourceChanged")
		addr.SetText(m.browser.Source())
	})
	m.browser.SetOnContentLoading(func(sender wv.IObject, webView wv.ICoreWebView2, args wv.ICoreWebView2ContentLoadingEventArgs) {
		fmt.Println("回调函数 WVBrowser => SetOnContentLoading")
	})
	m.browser.SetOnDocumentTitleChanged(func(sender lcl.IObject) {
		fmt.Println("回调函数 WVBrowser => SetOnDocumentTitleChanged:", m.browser.DocumentTitle())
	})
	m.browser.SetOnDownloadStateChanged(func(sender wv.IObject, downloadOperation wv.ICoreWebView2DownloadOperation, downloadID int32) {
		fmt.Println("SetOnDownloadStateChanged:", downloadOperation.BytesReceived(), "/", downloadOperation.TotalBytesToReceive())
	})
	m.browser.SetOnDownloadStarting(func(sender wv.IObject, webView wv.ICoreWebView2, args wv.ICoreWebView2DownloadStartingEventArgs) {
		args = wv.NewCoreWebView2DownloadStartingEventArgs(args)
		defer args.Free()
		fmt.Println("SetOnDownloadStarting:", args.ResultFilePath())
	})

	m.browser.SetOnNavigationCompleted(func(sender wv.IObject, webView wv.ICoreWebView2, args wv.ICoreWebView2NavigationCompletedEventArgs) {
		fmt.Println("回调函数 WVBrowser => SetOnNavigationCompleted")
		// 重置 stream
		//embedAssetsStream.Clear()
		fmt.Println("回调函数 WVBrowser => SetOnNavigationCompleted => stream.Clear()")
		navBtns(false)
		//webView = wv.NewCoreWebView2(webView)
		//addOk := webView.AddScriptToExecuteOnDocumentCreated("alert(1);", m.browser)
		//fmt.Println("AddScriptToExecuteOnDocumentCreated OK:", addOk)

		// 2. 使用植入 ipc js
		message := ProcessMessage{
			Name: "test",
			Data: []interface{}{"stringdata", true, 5555.66, 99999},
			Id:   0,
		}
		jsonData, _ := json.Marshal(message)
		m.browser.PostWebMessageAsString(string(jsonData))
		//message.Name = "test-return"
		//jsonData, _ = json.Marshal(message)
		//`window.wails.EventsNotify('` + template.JSEscapeString(string(payload)) + `');`
		//jsData := template.JSEscapeString(string(jsonData))
		//js := `window.energy.executeEvent('` + jsData + `');`
		//m.browser.ExecuteScript(js, 100)
	})

	m.SetOnShow(func(sender lcl.IObject) {
		if load.InitializationError() {
			fmt.Println("回调函数 => SetOnShow 初始化失败")
		} else {
			if load.Initialized() {
				fmt.Println("回调函数 => SetOnShow 初始化成功")
				m.browser.CreateBrowser(m.windowParent.Handle(), true)
			}
		}
	})
	m.SetOnWndProc(func(msg *types.TMessage) {
		m.InheritedWndProc(msg)
		switch msg.Msg {
		case messages.WM_SIZE, messages.WM_MOVE:
			m.browser.NotifyParentWindowPositionChanged()
		}
	})
}

// 控制组件UI
// 地址栏和控制按钮创建
func controlUI(window *TMainForm) (goBack lcl.IButton, goForward lcl.IButton, stop lcl.IButton, refresh lcl.IButton, addrBox lcl.IComboBox) {
	//这里使用系统UI组件
	//创建panel做为地址栏的父组件
	addrPanel := lcl.NewPanel(window) //设置父组件
	addrPanel.SetParent(window)
	addrPanel.SetAnchors(types.NewSet(types.AkLeft, types.AkTop, types.AkRight)) //设置锚点定位，让宽高自动根据窗口调整大小
	addrPanel.SetHeight(30)
	addrPanel.SetWidth(window.Width())
	//创建 按钮-后退
	goBack = lcl.NewButton(addrPanel) //设置父组件
	goBack.SetParent(addrPanel)
	goBack.SetCaption("后退")
	goBack.SetBounds(5, 3, 35, 25)
	goForward = lcl.NewButton(addrPanel) //设置父组件
	goForward.SetParent(addrPanel)
	goForward.SetCaption("前进")
	goForward.SetBounds(45, 3, 35, 25)
	stop = lcl.NewButton(addrPanel) //设置父组件
	stop.SetParent(addrPanel)
	stop.SetCaption("停止")
	stop.SetBounds(90, 3, 35, 25)
	refresh = lcl.NewButton(addrPanel) //设置父组件
	refresh.SetParent(addrPanel)
	refresh.SetCaption("刷新")
	refresh.SetBounds(135, 3, 35, 25)

	//创建下拉框
	addrBox = lcl.NewComboBox(addrPanel)
	addrBox.SetParent(addrPanel)
	addrBox.SetLeft(180)                                                       //这里是设置左边距 上面按钮的宽度
	addrBox.SetTop(3)                                                          //
	addrBox.SetWidth(window.Width() - (230))                                   //宽度 减按钮的宽度
	addrBox.SetAnchors(types.NewSet(types.AkLeft, types.AkTop, types.AkRight)) //设置锚点定位，让宽高自动根据窗口调整大小
	addrBox.Items().Add("myscheme://domain/index.html")
	addrBox.Items().Add("https://gitee.com/energye/energy")
	addrBox.Items().Add("https://github.com/energye/energy")
	addrBox.Items().Add("https://www.baidu.com")
	addrBox.Items().Add("https://energy.yanghy.cn")
	addrBox.SetText("myscheme://domain/index.html")

	goUrl := lcl.NewButton(addrPanel) //设置父组件
	goUrl.SetParent(addrPanel)
	goUrl.SetCaption("GO")
	goUrl.SetBounds(window.Width()-45, 3, 40, 25)
	goUrl.SetAnchors(types.NewSet(types.AkTop, types.AkRight)) //设置锚点定位，让宽高自动根据窗口调整大小

	//给按钮增加事件
	goBack.SetOnClick(func(sender lcl.IObject) {
		window.browser.GoBack()
	})
	goForward.SetOnClick(func(sender lcl.IObject) {
		window.browser.GoForward()
	})
	stop.SetOnClick(func(sender lcl.IObject) {
		window.browser.Stop()
	})
	refresh.SetOnClick(func(sender lcl.IObject) {
		window.browser.Refresh()
	})
	goUrl.SetOnClick(func(sender lcl.IObject) {
		var url = addrBox.Text()
		if url != "" {
			window.browser.Navigate(url)
		}
	})
	return
}
