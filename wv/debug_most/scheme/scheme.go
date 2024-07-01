package scheme

import (
	"fmt"
	"github.com/energye/wv/wv"
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

func AddWebResourceRequestedFilter(browser wv.IWVBrowser) {
	browser.AddWebResourceRequestedFilter(SchemeName+"*", wv.COREWEBVIEW2_WEB_RESOURCE_CONTEXT_ALL)
}
