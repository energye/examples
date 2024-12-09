package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"github.com/energye/assetserve"
	_ "github.com/energye/examples/syso"
	"github.com/energye/lcl/api"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/tools"
	"github.com/energye/lcl/types"
	"github.com/energye/wv/darwin"
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
	contextMenu   lcl.IPopupMenu
}

var mainForm TMainForm

//go:embed assets
var resources embed.FS

func main() {
	httpServer()
	wv.Init(nil, resources)
	lcl.Application.Initialize()
	lcl.Application.SetScaled(true)
	mainForm.IForm = &lcl.TForm{}
	mainForm.url = "energy://test.com"
	mainForm.isMainWindow = true
	lcl.Application.CreateForm(&mainForm)
	lcl.Application.Run()
}

func (m *TMainForm) CreateContextMenu() {
	// TPopupMenu
	m.contextMenu = lcl.NewPopupMenu(m)
	item := lcl.NewMenuItem(m)
	item.SetCaption("退出(&E)")
	item.SetOnClick(func(lcl.IObject) {
		m.Close()
	})
	m.contextMenu.Items().Add(item)
	item = lcl.NewMenuItem(m)
	item.SetCaption("Test")
	item.SetOnClick(func(lcl.IObject) {
		fmt.Println("test")
	})
	m.contextMenu.Items().Add(item)

	// 将窗口设置一个弹出菜单，右键单击就可显示
	m.SetPopupMenu(m.contextMenu)
}

func (m *TMainForm) CreateMainMenu() {
	mainMenu := lcl.NewMainMenu(m)
	// 创建一级菜单
	fileClassA := lcl.NewMenuItem(m)
	fileClassA.SetCaption("文件(&F)") //菜单名称 alt + f
	aboutClassA := lcl.NewMenuItem(m)
	aboutClassA.SetCaption("关于(&A)")

	var createMenuItem = func(label, shortCut string, click func(lcl.IObject)) (result lcl.IMenuItem) {
		result = lcl.NewMenuItem(m)
		result.SetCaption(label)                          //菜单项显示的文字
		result.SetShortCut(api.DTextToShortCut(shortCut)) // 快捷键
		result.SetOnClick(click)                          // 触发事件，回调函数
		return
	}
	// 给一级菜单添加菜单项
	createItem := createMenuItem("新建(&N)", "Meta+N", func(lcl.IObject) {
		fmt.Println("单击了新建")
	})
	fileClassA.Add(createItem) // 把创建好的菜单项添加到 第一个菜单中
	openItem := createMenuItem("打开(&O)", "Meta+O", func(lcl.IObject) {
		fmt.Println("单击了打开")
	})
	fileClassA.Add(openItem)
	mainMenu.Items().Add(fileClassA)
	mainMenu.Items().Add(aboutClassA)
	if tools.IsDarwin() {
		// https://wiki.lazarus.freepascal.org/Mac_Preferences_and_About_Menu
		// 动态添加的，静态好像是通过设计器将顶级的菜单标题设置为应用程序名，但动态的就是另一种方式
		appMenu := lcl.NewMenuItem(m)
		// 动态添加的，设置一个Unicode Apple logo char
		appMenu.SetCaption(types.AppleLogoChar)
		subItem := lcl.NewMenuItem(m)

		subItem.SetCaption("关于")
		subItem.SetOnClick(func(sender lcl.IObject) {
			lcl.ShowMessage("About")
		})
		appMenu.Add(subItem)

		subItem = lcl.NewMenuItem(m)
		subItem.SetCaption("-")
		appMenu.Add(subItem)

		subItem = lcl.NewMenuItem(m)
		subItem.SetCaption("首选项")
		subItem.SetShortCut(api.DTextToShortCut("Meta+,"))
		subItem.SetOnClick(func(sender lcl.IObject) {
			lcl.ShowMessage("Preferences")
		})
		appMenu.Add(subItem)
		// 添加
		mainMenu.Items().Insert(0, appMenu)
	}
}

func (m *TMainForm) FormCreate(sender lcl.IObject) {
	fmt.Println("main create")
	icon, _ := resources.ReadFile("assets/icon.ico")
	m.Icon().LoadFromBytes(icon)
	m.SetCaption("Main")
	m.SetWidth(800)
	m.SetHeight(600)
	m.SetDoubleBuffered(true)

	if m.isMainWindow {
		m.CreateMainMenu()
	}
	m.CreateContextMenu()

	m.webview = wv.NewWkWebview(m)
	m.webview.SetOnProcessMessage(func(sender wv.IObject, userContentController wv.WKUserContentController, name, message string) {
		fmt.Println("OnProcessMessage", name, "message:", message)
		messageData := struct {
			Name string `json:"n"`
		}{}
		json.Unmarshal([]byte(message), &messageData)
		if messageData.Name == "contextmenu" {
			m.contextMenu.PopUp()
		}
	})
	m.webview.SetOnStartProvisionalNavigation(func(sender wv.IObject, navigation wv.WKNavigation) {
		fmt.Println("OnStartProvisionalNavigation")
	})
	m.webview.SetOnFinishNavigation(func(sender wv.IObject, navigation wv.WKNavigation) {
		fmt.Println("OnFinishNavigation")
	})
	m.webview.SetOnDecidePolicyForNavigationActionPreferences(func(sender wv.IObject, navigationAction wv.WKNavigationAction,
		actionPolicy *wv.WKNavigationActionPolicy, preferences *wv.WKWebpagePreferences) {
		fmt.Println("OnDecidePolicyForNavigationActionPreferences")
		//wkNavigationAction := wv.NewWKNavigationAction(navigationAction)
		//sourceFrameInfo := wv.NewWKFrameInfo(wkNavigationAction.SourceFrame())
		//sourceRequest := wv.NewNSURLRequest(sourceFrameInfo.Request())
		//targetFrameInfo := wv.NewWKFrameInfo(wkNavigationAction.TargetFrame())
		//targetRequest := wv.NewNSURLRequest(targetFrameInfo.Request())
		//if sourceRequest.IsValid() {
		//	url := wv.NewNSURL(sourceRequest.URL())
		//	fmt.Println("\tsource:", url.AbsoluteString())
		//	url.Free()
		//}
		//if targetRequest.IsValid() {
		//	url := wv.NewNSURL(targetRequest.URL())
		//	fmt.Println("\ttarget:", url.AbsoluteString())
		//	fmt.Println("\ttarget:", url.Scheme(), url.Path())
		//	url.Free()
		//}
		//request := wv.NewNSURLRequest(wkNavigationAction.Request())
		//if request.IsValid() {
		//	url := wv.NewNSURL(request.URL())
		//	fmt.Println("\trequest:", url.AbsoluteString())
		//	fmt.Println("\trequest:", url.Scheme(), url.Path())
		//	url.Free()
		//}
	})
	// 打开一个新窗口时触发事件
	m.webview.SetOnCreateWebView(m.OnCreateWebView)

	m.webview.SetOnStartURLSchemeTask(m.OnStartURLSchemeTask)
	m.webview.SetOnStopURLSchemeTask(m.OnStopURLSchemeTask)

	// webview parent
	m.webviewParent = wv.NewWkWebviewParent(m)
	m.webviewParent.SetParent(m)
	m.webviewParent.SetAlign(types.AlClient)
	m.webviewParent.SetParentDoubleBuffered(true)

	userContentController := wv.WKUserContentControllerRef.New()
	scriptMessageHandler := wv.NewWKScriptMessageHandler(m.webview.AsReceiveScriptMessageDelegate())
	// 自定义ipc进程消息对象(在js使用)
	userContentController.AddScriptMessageHandlerName(scriptMessageHandler, "processMessage")

	configuration := wv.WKWebViewConfigurationRef.New()
	configuration.SetUserContentController(userContentController.Data())

	URLSchemeHandler := wv.NewWKURLSchemeHandler(m.webview.AsWKURLSchemeHandlerDelegate())

	configuration.SetSuppressesIncrementalRendering(true)
	configuration.SetApplicationNameForUserAgent("energy.cn")
	// 自定义 url 协议
	configuration.SetURLSchemeHandlerForURLScheme(URLSchemeHandler.Data(), "energy")

	preference := wv.NewWKPreferences(configuration.Preferences()) //wv.WKPreferencesRef.New()
	configuration.SetPreferences(preference.Data())

	preference.SetTabFocusesLinks(true)
	preference.SetFraudulentWebsiteWarningEnabled(true)
	preference.EnableDevtools()

	navigationDelegate := wv.NewWKNavigationDelegate(m.webview.AsWKNavigationDelegate())
	uiDelegate := wv.NewWKUIDelegate(m.webview.AsWKUIDelegate())

	frame := &wv.TRect{}
	frame.SetWidth(m.Width())
	frame.SetHeight(m.Height())
	m.webview.InitWithFrameConfiguration(frame, configuration.Data())

	m.webview.SetNavigationDelegate(navigationDelegate.Data())
	m.webview.SetUIDelegate(uiDelegate.Data())

	// end
	m.webviewParent.SetWebview(m.webview.Data())

	m.SetOnShow(func(sender lcl.IObject) {
		fmt.Println("OnShow:", m.url)
		if m.url != "" {
			m.webview.LoadURL(m.url)
		}
		m.ScreenCenter()
	})
	m.webview.SetOnWebContentProcessDidTerminate(func(sender wv.IObject) {
		fmt.Println("OnWebContentProcessDidTerminate")
	})
	m.webview.SetOnWebViewDidClose(func(sender wv.IObject) {
		fmt.Println("OnWebViewDidClose")
	})
	m.SetOnCloseQuery(func(sender lcl.IObject, canClose *bool) {
		fmt.Println("OnCloseQuery")
		//*canClose = m.canClose
		m.webview.StopLoading()
		m.webview.RemoveFromSuperview()
		m.webview.Release()
		m.webviewParent.Free()
	})
}

func (m *TMainForm) CreateParams(params *types.TCreateParams) {
	fmt.Println("调用此过程  TMainForm.CreateParams:", *params)
}

func (m *TMainForm) OnCreateWebView(sender wv.IObject, configuration wv.WKWebViewConfiguration, navigationAction wv.WKNavigationAction,
	windowFeatures wv.WKWindowFeatures) wv.WkWebview {
	fmt.Println("OnCreateWebView")
	wkNavigationAction := wv.NewWKNavigationAction(navigationAction)
	sourceFrameInfo := wv.NewWKFrameInfo(wkNavigationAction.SourceFrame())
	sourceRequest := wv.NewNSURLRequest(sourceFrameInfo.Request())
	targetFrameInfo := wv.NewWKFrameInfo(wkNavigationAction.TargetFrame())
	targetRequest := wv.NewNSURLRequest(targetFrameInfo.Request())
	if sourceRequest.IsValid() {
		url := wv.NewNSURL(sourceRequest.URL())
		fmt.Println("\tsource:", url.AbsoluteString())
		url.Free()
	}
	if targetRequest.IsValid() {
		url := wv.NewNSURL(targetRequest.URL())
		fmt.Println("\ttarget:", url.AbsoluteString())
		fmt.Println("\ttarget:", url.Scheme(), url.Path())
		url.Free()
	}

	request := wv.NewNSURLRequest(wkNavigationAction.Request())
	if request.IsValid() {
		url := wv.NewNSURL(request.URL())
		fmt.Println("\trequest:", url.AbsoluteString())
		fmt.Println("\trequest:", url.Scheme(), url.Path())
		window := NewWindow(url.AbsoluteString())
		window.Show()
		url.Free()
	}
	return 0
}
func (m *TMainForm) OnStartURLSchemeTask(sender wv.IObject, urlSchemeTask wv.WKURLSchemeTask) {
	fmt.Println("OnStartURLSchemeTask")
	tempURLSchemeTask := wv.NewWKURLSchemeTask(urlSchemeTask)
	request := wv.NewNSURLRequest(tempURLSchemeTask.Request())
	tempNSURL := wv.NewNSURL(request.URL())
	tempUrl := tempNSURL.AbsoluteString()
	tempHost := tempNSURL.Host()
	tempPath := tempNSURL.Path()
	fmt.Println(tempUrl, tempHost, tempPath)
	if tempPath == "" {
		tempPath = "index.html"
	}
	tempHtml, _ := resources.ReadFile(filepath.Join("assets", tempPath))
	tempDataBytesLength := int32(len(tempHtml))

	tempHTTPHeader := request.AllHTTPHeaderFields()
	fmt.Println("tempHTTPHeader:", tempHTTPHeader)

	// 响应对象必须包含所请求资源的 MIME 类型
	response := wv.NSURLResponseRef.New()
	response.InitWithURLMIMETypeExpectedContentLengthTextEncodingName(tempNSURL.Data(), "text/html", tempDataBytesLength, "utf-8")

	tempURLSchemeTask.ReceiveResponse(response.Data())
	tempURLSchemeTask.ReceiveData(uintptr(unsafe.Pointer(&tempHtml[0])), tempDataBytesLength)
	tempURLSchemeTask.Finish()
	tempURLSchemeTask.Free()
	response.Free()
}

func (m *TMainForm) OnStopURLSchemeTask(sender wv.IObject, urlSchemeTask wv.WKURLSchemeTask) {
	fmt.Println("OnStopURLSchemeTask")
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
