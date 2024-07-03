package main

import (
	"embed"
	"encoding/json"
	"fmt"
	_ "github.com/energye/examples/syso"
	"github.com/energye/examples/wv/debug_most/contextmenu"
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
	"net/url"
	"path/filepath"
)

type TMainForm struct {
	lcl.TForm
	windowParent wv.IWVWindowParent
	browser      wv.IWVBrowser
}

type ProcessMessage struct {
	Name string        `json:"n"`
	Data []interface{} `json:"d"`
	Id   int           `json:"i"`
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

func (m *TMainForm) FormCreate(sender lcl.IObject) {
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
	m.browser.SetDefaultURL("myscheme://domain/index.html")
	m.browser.SetTargetCompatibleBrowserVersion("95.0.1020.44")
	fmt.Println("TargetCompatibleBrowserVersion:", m.browser.TargetCompatibleBrowserVersion())
	contextmenu.Contextmenu(m, m.browser)
	devtools.DevTools(m.browser)

	m.browser.SetOnAfterCreated(func(sender lcl.IObject) {
		fmt.Println("回调函数 WVBrowser => SetOnAfterCreated")
		m.windowParent.UpdateSize()
		scheme.AddWebResourceRequestedFilter(m.browser)
		// 1. 先植入 ipc js
		fmt.Println("AddScriptToExecuteOnDocumentCreated:", string(utils.IPCJavaScript))
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
		//args = wv.NewCoreWebView2NavigationStartingEventArgs(args)
		//defer args.Free()
		navBtns(true)
	})
	// 进程消息
	m.browser.SetOnWebMessageReceived(func(sender wv.IObject, webView wv.ICoreWebView2, args wv.ICoreWebView2WebMessageReceivedEventArgs) {
		fmt.Println("回调函数 WVBrowser => SetOnWebMessageReceived")
		args = wv.NewCoreWebView2WebMessageReceivedEventArgs(args)
		defer args.Free()
		fmt.Println("\tdata string:", args.WebMessageAsString())
		var message ProcessMessage
		_ = json.Unmarshal([]byte(args.WebMessageAsString()), &message)
		fmt.Printf("\tmessage: %+v\n ", message)
		if message.Name == "emit-name" {
			// messageId 不等于0表示有回调函数需要执行
			// 需要回调一个消息
			if message.Id != 0 {
				message.Name = "" // 不需要事件名
				message.Data = append(message.Data, "返回值")
				jsonData, _ := json.Marshal(message)
				m.browser.PostWebMessageAsString(string(jsonData))
			}
		} else if message.Name == "showDevtools" {
			devtools.OpenDevtools(m.browser)
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
	var (
		stream  lcl.IMemoryStream
		adapter lcl.IStreamAdapter
	)
	// 自定义协议资源加载
	m.browser.SetOnWebResourceRequested(func(sender wv.IObject, webView wv.ICoreWebView2, args wv.ICoreWebView2WebResourceRequestedEventArgs) {
		args = wv.NewCoreWebView2WebResourceRequestedEventArgs(args)
		request := wv.NewCoreWebView2WebResourceRequestRef(args.Request())
		// 需要释放掉
		defer func() {
			request.Free()
			args.Free()
		}()
		if stream != nil {
			fmt.Println("stream-position:", stream.Position())
			stream.Free()
		}
		if adapter != nil {
			fmt.Println("stream-RefCount:", adapter.RefCount())
			adapter.Free()
		}
		fmt.Println("回调函数 WVBrowser => SetOnWebResourceRequested")
		fmt.Println("回调函数 WVBrowser => TempURI:", request.URI(), request.Method())
		fmt.Println("回调函数 WVBrowser => 内置exe读取 index.html ")
		reqUrl, _ := url.Parse(request.URI())
		fmt.Println("reqUrl.Path:", reqUrl.Path)
		data, err := assets.ReadFile("assets" + reqUrl.Path)
		fmt.Println("加载本地资源:", err)
		stream = lcl.NewMemoryStream()
		stream.LoadFromBytes(data)
		fmt.Println("回调函数 WVBrowser => stream", stream.Size())
		adapter = lcl.NewStreamAdapter(stream, types.SoOwned)
		fmt.Println("回调函数 WVBrowser => adapter:", adapter.StreamOwnership(), adapter.Stream().Size())

		var response wv.ICoreWebView2WebResourceResponse
		environment := m.browser.CoreWebView2Environment()
		fmt.Println("回调函数 WVBrowser => Initialized():", environment.Initialized(), environment.BrowserVersionInfo())
		environment.CreateWebResourceResponse(adapter, 200, "OK", "Content-Type: text/html", &response)
		args.SetResponse(response)
	})

	m.browser.SetOnNavigationCompleted(func(sender wv.IObject, webView wv.ICoreWebView2, args wv.ICoreWebView2NavigationCompletedEventArgs) {
		fmt.Println("回调函数 WVBrowser => SetOnNavigationCompleted")
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
	m.windowParent.SetBrowser(m.browser)

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
