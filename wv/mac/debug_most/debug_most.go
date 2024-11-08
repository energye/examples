package main

import (
	"embed"
	"fmt"
	"github.com/energye/assetserve"
	_ "github.com/energye/examples/syso"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
	"github.com/energye/wv/darwin"
	"os"
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
	mainForm.url = "file:///Users/yanghy/app/github.com/workspace/lib/wk/webkit2_mac/demo/test.html"
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

	m.webview = wv.NewWkWebview(m)

	// webview parent
	m.webviewParent = wv.NewWkWebviewParent(m)
	m.webviewParent.SetParent(m)
	m.webviewParent.SetAlign(types.AlClient)
	m.webviewParent.SetParentDoubleBuffered(true)

	//userContentController := wv.WKUserContentControllerRef.New()
	//wv.NewWKScriptMessageHandler()

	// end
	m.webviewParent.SetWebview(m.webview.Data())

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
