package cookie

import (
	"fmt"
	"github.com/energye/cef/cef"
	"github.com/energye/lcl/lcl"
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
