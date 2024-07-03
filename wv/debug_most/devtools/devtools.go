package devtools

import "github.com/energye/wv/wv"

func OpenDevtools(browser wv.IWVBrowser) {
	browser.OpenDevToolsWindow()
}
