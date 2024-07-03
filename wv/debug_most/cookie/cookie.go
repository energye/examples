package cookie

import (
	"fmt"
	"github.com/energye/wv/wv"
)

func Cookie(browser wv.IWVBrowser) {
	browser.SetOnGetCookiesCompleted(func(sender wv.IObject, result int32, cookieList wv.ICoreWebView2CookieList) {
		cookieList = wv.NewCoreWebView2CookieList(cookieList)
		defer cookieList.Free()
		count := int(cookieList.Count())
		for i := 0; i < count; i++ {
			cookie := cookieList.Items(uint32(i))
			cookie = wv.NewCoreWebView2Cookie(cookie)
			fmt.Println("count", cookie.Name(), cookie.Domain())
		}
	})
}
