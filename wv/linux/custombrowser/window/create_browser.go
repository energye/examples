package window

import (
	"github.com/energye/lcl/lcl"
	wv "github.com/energye/wv/linux"
)

type Browser struct {
	mainWindow                         *BrowserWindow
	windowId                           int32 // 窗口ID
	webviewParent                      wv.IWkWebviewParent
	webview                            wv.IWkWebview
	oldWndPrc                          uintptr
	tabSheetBtn                        *TabButton
	tabSheet                           lcl.IPanel
	isActive                           bool
	currentURL                         string
	currentTitle                       string
	siteFavIcon                        map[string]string
	isLoading, canGoBack, canGoForward bool
	isCloseing                         bool
}
