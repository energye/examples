package cookie

import (
	"fmt"
	"github.com/energye/wv/windows"
)

func Cookie(browser wv.IWVBrowser) {
	browser.SetOnGetCookiesCompleted(func(sender wv.IObject, result int32, cookieList wv.ICoreWebView2CookieList) {
		cookieList = wv.NewCoreWebView2CookieList(cookieList)
		defer cookieList.Free()
		count := int(cookieList.Count())
		for i := 0; i < count; i++ {
			cookie := cookieList.Items(uint32(i))
			cookie = wv.NewCoreWebView2Cookie(cookie)
			fmt.Println("name:", cookie.Name(), "value:", cookie.Value())
			cookie.Free()
		}
	})
}
