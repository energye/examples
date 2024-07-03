package scheme

import (
	"fmt"
	"github.com/energye/examples/wv/debug_most/utils"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
	"github.com/energye/wv/wv"
	"net/url"
)

var SchemeName = "myscheme"

func LoaderOnCustomSchemes(loader wv.IWVLoader) {
	loader.SetOnGetCustomSchemes(func(sender wv.IObject, customSchemes *wv.TWVCustomSchemeInfoArray) {
		fmt.Println("回调函数 WebView2Loader => SetOnGetCustomSchemes size:", len(*customSchemes))
		*customSchemes = append(*customSchemes, &wv.TWVCustomSchemeInfo{
			SchemeName:            SchemeName,
			TreatAsSecure:         true,
			AllowedDomains:        "https://*.baidu.com,https://*.yanghy.cn",
			HasAuthorityComponent: true,
		})
	})
}

func OnAfterCreated(browser wv.IWVBrowser) {
	browser.AddWebResourceRequestedFilter(SchemeName+"*", wv.COREWEBVIEW2_WEB_RESOURCE_CONTEXT_ALL)
}

func WebResourceRequested(browser wv.IWVBrowser) {
	var (
		embedAssetsStream  = lcl.NewMemoryStream()
		embedAssetsAdapter = lcl.NewStreamAdapter(embedAssetsStream, types.SoOwned)
	)
	// 自定义协议资源加载
	browser.SetOnWebResourceRequested(func(sender wv.IObject, webView wv.ICoreWebView2, args wv.ICoreWebView2WebResourceRequestedEventArgs) {
		args = wv.NewCoreWebView2WebResourceRequestedEventArgs(args)
		request := wv.NewCoreWebView2WebResourceRequestRef(args.Request())
		// 需要释放掉
		defer func() {
			request.Free()
			args.Free()
		}()
		// 重置 stream
		embedAssetsStream.SetPosition(0)
		embedAssetsStream.Clear()
		fmt.Println("回调函数 WVBrowser => SetOnWebResourceRequested")
		fmt.Println("回调函数 WVBrowser => TempURI:", request.URI(), request.Method())
		reqUrl, _ := url.Parse(request.URI())
		fmt.Println("回调函数 WVBrowser => 内置exe读取", reqUrl.Path)
		data, err := utils.Assets.ReadFile("assets" + reqUrl.Path)
		if err != nil {
			fmt.Println("加载本地资源-error:", err)
		}
		embedAssetsStream.LoadFromBytes(data)
		fmt.Println("回调函数 WVBrowser => stream", embedAssetsStream.Size())
		fmt.Println("回调函数 WVBrowser => adapter:", embedAssetsAdapter.StreamOwnership(), embedAssetsAdapter.Stream().Size())

		var response wv.ICoreWebView2WebResourceResponse
		environment := browser.CoreWebView2Environment()
		fmt.Println("回调函数 WVBrowser => Initialized():", environment.Initialized(), environment.BrowserVersionInfo())
		environment.CreateWebResourceResponse(embedAssetsAdapter, 200, "OK", "Content-Type: text/html", &response)
		args.SetResponse(response)
	})

}
