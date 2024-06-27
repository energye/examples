package cookie

import (
	"fmt"
	"github.com/energye/cef/cef"
	"github.com/energye/lcl/lcl"
	"time"
)

func Cookie(chromium cef.IChromium) {
	//获取cookie时触发
	chromium.SetOnCookiesVisited(func(sender cef.IObject, cookie cef.TCookie, count, total, id int32, deleteCookie, result *bool) {
		fmt.Println("SetOnCookiesVisited: ", count, total, id, deleteCookie)
		fmt.Println("cookie:", cookie)
	})
	//删除cookie时触发
	chromium.SetOnCookiesDeleted(func(sender lcl.IObject, numDeleted int32) {
		fmt.Println("SetOnCookiesDeleted:", numDeleted)
	})
	//设置cookie时触发
	chromium.SetOnCookieSet(func(sender lcl.IObject, success bool, ID int32) {
		fmt.Println("SetOnCookieSet: ", success, ID)
	})
	chromium.SetOnCookiesFlushed(func(sender lcl.IObject) {
		fmt.Println("OnCookiesFlushed")
	})
	chromium.SetOnCookieVisitorDestroyed(func(sender lcl.IObject, ID int32) {
		fmt.Println("OnCookieVisitorDestroyed")
	})
}

func CookieVisited(chromium cef.IChromium) {
	chromium.VisitURLCookies("https://www.baidu.com", true, 1)
	chromium.VisitAllCookies(1)
}

func DeleteCookie(chromium cef.IChromium) {
	chromium.DeleteCookies("", "", false)
}

func SetCookie(chromium cef.IChromium) {
	cookie := cef.TCookie{
		Name:       "example_cookie_name",
		Value:      "111",
		Domain:     "",
		Path:       "/",
		Creation:   cef.DateTimeToDTime(time.Now()),
		LastAccess: cef.DateTimeToDTime(time.Now()),
		Expires:    cef.DateTimeToDTime(time.Now()),
		Secure:     true,
		Httponly:   true,
		HasExpires: true,
		SameSite:   cef.CEF_COOKIE_SAME_SITE_UNSPECIFIED,
		Priority:   cef.CEF_COOKIE_PRIORITY_MEDIUM,
	}
	chromium.SetCookie("https://www.example.com", false, 1, cookie)
	chromium.SetCookie("https://www.example.com", false, 2, cookie)
	chromium.SetCookie("https://www.baidu.com", false, 3, cookie)

}
