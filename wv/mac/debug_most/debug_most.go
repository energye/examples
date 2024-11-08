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

	CookieManage := lcl.NewMenuItem(m)
	CookieManage.SetCaption("CookieManage")
	mainMenu.Items().Add(CookieManage)
	getAcceptPolicy := lcl.NewMenuItem(m)
	getAcceptPolicy.SetCaption("GetAcceptPolicy")
	CookieManage.Add(getAcceptPolicy)
	getAcceptPolicy.SetOnClick(func(sender lcl.IObject) {
	})
	addCookie := lcl.NewMenuItem(m)
	addCookie.SetCaption("AddCookie")
	CookieManage.Add(addCookie)
	addCookie.SetOnClick(func(sender lcl.IObject) {
	})
	getCookie := lcl.NewMenuItem(m)
	getCookie.SetCaption("GetCookie")
	CookieManage.Add(getCookie)
	getCookie.SetOnClick(func(sender lcl.IObject) {
	})
	deleteCookie := lcl.NewMenuItem(m)
	deleteCookie.SetCaption("DeleteCookie")
	CookieManage.Add(deleteCookie)
	deleteCookie.SetOnClick(func(sender lcl.IObject) {
		fmt.Println("DeleteCookie")
	})

	// webview parent
	m.webviewParent = wv.NewWkWebviewParent(m)
	m.webviewParent.SetParent(m)
	m.webviewParent.SetAlign(types.AlClient)
	m.webviewParent.SetParentDoubleBuffered(true)

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
