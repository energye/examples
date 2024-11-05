package devtools

import (
	"fmt"
	"github.com/energye/wv/windows"
)

func OpenDevtools(browser wv.IWVBrowser) {
	browser.OpenDevToolsWindow()
}

func DevTools(browser wv.IWVBrowser) {
	browser.SetOnDevToolsProtocolEventReceived(func(sender wv.IObject, webView wv.ICoreWebView2, args wv.ICoreWebView2DevToolsProtocolEventReceivedEventArgs, eventName string, eventID int32) {
		fmt.Println("SetOnDevToolsProtocolEventReceived")
	})
	browser.SetOnCallDevToolsProtocolMethodCompleted(func(sender wv.IObject, errorCode int32, returnObjectAsJson string, executionID int32) {
		fmt.Println("SetOnCallDevToolsProtocolMethodCompleted errorCode:", errorCode, "returnObjectAsJson:", returnObjectAsJson, "executionID:", executionID)
	})
}

func ExecuteDevToolsMethod(browser wv.IWVBrowser) {
	browser.SubscribeToDevToolsProtocolEvent("Emulation.setUserAgentOverride", 110)
	ok := browser.CallDevToolsProtocolMethod("Runtime.evaluate", `{"expression":"alert(’test‘)"}`, 100)
	fmt.Println("ExecuteDevToolsMethod ok:", ok)
	ok = browser.CallDevToolsProtocolMethod("Emulation.setUserAgentOverride", `{"userAgent":"Mozilla/5.0 (Linux; Android 11; M2102K1G) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.101 Mobile Safari/537.36"}`, 110)
	fmt.Println("ExecuteDevToolsMethod ok:", ok)
}
