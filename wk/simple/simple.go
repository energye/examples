package main

import (
	"fmt"
	_ "github.com/energye/examples/syso"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
	"github.com/energye/wk/wk"
	"io/ioutil"
	"unsafe"
)

type TMainForm struct {
	lcl.TForm
	webviewParent wk.IWkWebViewParent
	webview       wk.IWkWebview
	canClose      bool
}

var MainForm TMainForm

func main() {
	//os.Setenv("JSC_SIGNAL_FOR_GC", "SIGUSR")
	wk.Init(nil, nil)
	lcl.Application.Initialize()
	lcl.Application.SetScaled(true)
	lcl.Application.CreateForm(&MainForm)
	lcl.Application.Run()

}

func (m *TMainForm) FormCreate(sender lcl.IObject) {
	fmt.Println("main create")
	m.SetCaption("Main")
	// gtk3 需要设置一次较小的宽高, 然后在 OnShow 里设置默认宽高
	m.SetWidth(100)
	m.SetHeight(100)

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

	m.webviewParent = wk.NewWkWebViewParent(m)
	m.webviewParent.SetParent(m)
	m.webviewParent.SetAlign(types.AlClient)

	m.webview = wk.NewWkWebview(m)
	m.webview.SetOnContextMenu(func(sender wk.IObject, contextMenu wk.IWkContextMenu, defaultAction wk.WkAction) bool {
		fmt.Println("defaultAction:", defaultAction)
		return false
	})
	m.webview.SetOnWebProcessTerminated(func(sender wk.IObject, reason wk.WebKitWebProcessTerminationReason) {
		fmt.Println("SetOnWebProcessTerminated reason:", reason)
		if reason == wk.WEBKIT_WEB_PROCESS_TERMINATED_BY_API { //  call m.webview.TerminateWebProcess()
			m.webview.FreeWebview()
			m.Close()
		}
	})
	m.webview.SetOnURISchemeRequest(func(sender wk.IObject, uriSchemeRequest wk.IWkURISchemeRequest) {
		fmt.Println("OnURISchemeRequest")
		fmt.Println("uri:", uriSchemeRequest.Uri())
		data, _ := ioutil.ReadFile("/home/yanghy/app/gopath/src/github.com/energye/workspace/examples/wk/simple/test.html")

		ins := wk.NewWkInputStream1(uintptr(unsafe.Pointer(&data[0])), int64(len(data)))
		uriSchemeRequest.Finish(ins, int64(len(data)), "text/html")
	})
	wkContext := wk.WkWebContextRef.Default()
	wkContext.RegisterURIScheme("energy", m.webview.AsSchemeRequestDelegate())

	// 所有webview事件或配置都在 CreateBrowser 之前
	m.webview.CreateBrowser()
	m.webviewParent.SetWebView(m.webview)

	m.SetOnShow(func(sender lcl.IObject) {
		//m.webview.LoadURL("https://energye.github.io")
		m.webview.LoadURL("energy://demo.com")
		// gtk3 需要设置一次较小的宽高, 然后在 OnShow 里设置默认宽高
		m.SetWidth(1024)
		m.SetHeight(600)
		m.ScreenCenter()
	})

	m.SetOnCloseQuery(func(sender lcl.IObject, canClose *bool) {
		*canClose = m.canClose
		if !m.canClose {
			m.canClose = true
			m.webview.Stop()
			m.webview.TerminateWebProcess()
			//m.webviewParent.FreeChild()
		}
	})
}

func (m *TMainForm) CreateParams(params *types.TCreateParams) {
	fmt.Println("调用此过程  TMainForm.CreateParams:", *params)

}
