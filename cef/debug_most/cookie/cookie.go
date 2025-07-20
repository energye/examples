package cookie

import (
	"fmt"
	"github.com/energye/cef/cef"
	cefTypes "github.com/energye/cef/types"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
)

func Cookie(chromium cef.IChromium) {
	//获取cookie时触发
	chromium.SetOnCookiesVisited(func(sender lcl.IObject, name string, value string, domain string, path string, secure bool, httponly bool, hasExpires bool,
		creation types.TDateTime, lastAccess types.TDateTime, expires types.TDateTime, count int32, total int32, iD int32, sameSite cefTypes.TCefCookieSameSite,
		priority cefTypes.TCefCookiePriority, deleteCookie *bool, result *bool) {
		fmt.Println("SetOnCookiesVisited: ", count, total, iD, deleteCookie)
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
	//cookie := cef.TCookie{
	//	Name:       "example_cookie_name",
	//	Value:      "111",
	//	Domain:     "",
	//	Path:       "/",
	//	Creation:   cef.DateTimeToDTime(time.Now()),
	//	LastAccess: cef.DateTimeToDTime(time.Now()),
	//	Expires:    cef.DateTimeToDTime(time.Now()),
	//	Secure:     true,
	//	Httponly:   true,
	//	HasExpires: true,
	//	SameSite:   cef.CEF_COOKIE_SAME_SITE_UNSPECIFIED,
	//	Priority:   cef.CEF_COOKIE_PRIORITY_MEDIUM,
	//}
	//fmt.Println("set cookie 1")
	//chromium.SetCookie("https://www.example.com", false, 1, cookie)
	//fmt.Println("set cookie 2")
	//chromium.SetCookie("https://www.example.com", false, 2, cookie)
	//fmt.Println("set cookie 3")
	//chromium.SetCookie("https://www.baidu.com", false, 3, cookie)
}
