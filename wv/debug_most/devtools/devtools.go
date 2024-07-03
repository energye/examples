package devtools

import "github.com/energye/wv/wv"

func OpenDevtools(browser wv.IWVBrowser) {
	browser.OpenDevToolsWindow()
}

func DevTools(browser wv.IWVBrowser) {
	browser.SetOnDevToolsProtocolEventReceived(func(sender wv.IObject, webView wv.ICoreWebView2, args wv.ICoreWebView2DevToolsProtocolEventReceivedEventArgs, eventName string, eventID int32) {

	})
	browser.SetOnCallDevToolsProtocolMethodCompleted(func(sender wv.IObject, errorCode int32, returnObjectAsJson string, executionID int32) {

	})
}
