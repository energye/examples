package main

import (
	"embed"
	"fmt"
	"github.com/energye/assetserve"
	_ "github.com/energye/examples/syso"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
	"github.com/energye/wv/darwin"
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
	mainForm TMainForm
)

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

func (m *TMainForm) FormCreate(sender lcl.IObject) {
	fmt.Println("main create")
	icod, _ := resources.ReadFile("assets/icon.ico")
	m.Icon().LoadFromBytes(icod)
	m.SetCaption("Main")
	m.SetWidth(1024)
	m.SetHeight(768)
	m.SetDoubleBuffered(true)

	// TPopupMenu
	pm := lcl.NewPopupMenu(m)
	item := lcl.NewMenuItem(m)
	item.SetCaption("退出(&E)")
	item.SetOnClick(func(lcl.IObject) {
		m.Close()
	})
	pm.Items().Add(item)
	item = lcl.NewMenuItem(m)
	item.SetCaption("Test")
	item.SetOnClick(func(lcl.IObject) {
		fmt.Println("test")
	})
	pm.Items().Add(item)

	// 将窗口设置一个弹出菜单，右键单击就可显示
	m.SetPopupMenu(pm)

	m.webview = wv.NewWkWebview(m)
	m.webview.SetOnProcessMessage(func(sender wv.IObject, userContentController wv.WKUserContentController, name, message string) {
		fmt.Println("OnProcessMessage", name, "message:", message)
		pm.PopUp()
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
	userContentController.AddScriptMessageHandlerName(scriptMessageHandler, "processMessage")

	configuration := wv.WKWebViewConfigurationRef.New()
	configuration.SetUserContentController(userContentController.Data())

	URLSchemeHandler := wv.NewWKURLSchemeHandler(m.webview.AsWKURLSchemeHandlerDelegate())

	configuration.SetSuppressesIncrementalRendering(true)
	configuration.SetApplicationNameForUserAgent("energy.cn")
	configuration.SetURLSchemeHandlerForURLScheme(URLSchemeHandler.Data(), "energy")

	preference := wv.WKPreferencesRef.New()
	configuration.SetPreferences(preference.Data())

	preference.SetTabFocusesLinks(true)
	preference.SetFraudulentWebsiteWarningEnabled(true)

	navigationDelegate := wv.NewWKNavigationDelegate(m.webview.AsWKNavigationDelegate())
	UIDelegate := wv.NewWKUIDelegate(m.webview.AsWKUIDelegate())

	frame := &wv.TRect{}
	frame.SetWidth(m.Width())
	frame.SetHeight(m.Height())
	m.webview.InitWithFrameConfiguration(frame, configuration.Data())

	m.webview.SetNavigationDelegate(navigationDelegate.Data())
	m.webview.SetUIDelegate(UIDelegate.Data())

	// end
	m.webviewParent.SetWebview(m.webview.Data())

	m.SetOnShow(func(sender lcl.IObject) {
		fmt.Println("OnShow:", m.url)
		//m.webview.LoadURL("https://energye.github.io")
		//m.webview.LoadURL("http://localhost:22022/index.html")
		//m.webview.LoadURL(m.url)
		if m.url != "" {
			m.webview.LoadURL(m.url)
		}
		m.ScreenCenter()
	})
	m.webview.SetOnWebContentProcessDidTerminate(func(sender wv.IObject) {
		fmt.Println("OnWebContentProcessDidTerminate")
	})
	m.SetOnCloseQuery(func(sender lcl.IObject, canClose *bool) {
		fmt.Println("OnCloseQuery")
		//*canClose = m.canClose
		m.webview.StopLoading()
		m.webview.RemoveFromSuperview()
		//m.webview.Release()
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
	tempHtml, _ := resources.ReadFile("assets/index.html")
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
