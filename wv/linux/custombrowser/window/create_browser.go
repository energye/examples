package window

import (
	"github.com/energye/lcl/lcl"
	wv "github.com/energye/wv/linux"
	"widget/wg"
)

type Browser struct {
	mainWindow                         *BrowserWindow
	windowId                           int32 // 窗口ID
	webviewParent                      wv.IWkWebviewParent
	webview                            wv.IWkWebview
	oldWndPrc                          uintptr
	tabSheetBtn                        *wg.TButton
	tabSheet                           lcl.IPanel
	isActive                           bool
	currentURL                         string
	currentTitle                       string
	siteFavIcon                        map[string]string
	isLoading, canGoBack, canGoForward bool
	isCloseing                         bool
}
