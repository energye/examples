package cookie

import (
	"fmt"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
	wv "github.com/energye/wv/windows"
)

func Cookie(browser wv.IWVBrowser) {
	browser.SetOnGetCookiesCompleted(func(sender lcl.IObject, errorCode types.HRESULT, result wv.ICoreWebView2CookieList) {
		result = wv.NewCoreWebView2CookieList(result)
		defer result.Free()
		count := int(result.Count())
		for i := 0; i < count; i++ {
			cookie := result.Items(uint32(i))
			cookie = wv.NewCoreWebView2Cookie(cookie)
			fmt.Println("name:", cookie.Name(), "value:", cookie.Value())
			cookie.Free()
		}
	})
}
