package window

import (
	"fmt"
	"github.com/energye/cef/cef"
	cefTypes "github.com/energye/cef/types"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/tool"
	"github.com/energye/lcl/types"
	"github.com/energye/lcl/types/colors"
	"widget/wg"
)

type ChromiumAfterCreate func(newChromium *Chromium)

type Chromium struct {
	mainWindow   *BrowserWindow
	windowId     int32 // 窗口ID
	timer        lcl.ITimer
	windowParent cef.ICEFWinControl
	chromium     cef.IChromium
	canClose     bool
	oldWndPrc    uintptr
	afterCreate  ChromiumAfterCreate
	tabSheet     *wg.TButton
	isActive     bool
	currentURL   string
}

func (m *Chromium) createBrowser(sender lcl.IObject) {
	if m.timer == nil {
		return
	}
	m.timer.SetEnabled(false)
	rect := m.windowParent.Parent().ClientRect()
	init := m.chromium.Initialized()
	created := m.chromium.CreateBrowserWithWindowHandleRectStringRequestContextDictionaryValueBool(m.windowParent.Handle(), rect, "", nil, nil, false)
	fmt.Println("createBrowser rect:", rect, "init:", init, "create:", created)
	if !created {
		m.timer.SetEnabled(true)
	} else {
		m.windowParent.UpdateSize()
		m.timer.Free()
		m.timer = nil
	}
}

func (m *Chromium) resize(sender lcl.IObject) {
	if m.chromium != nil {
		m.chromium.NotifyMoveOrResizeStarted()
		if m.windowParent != nil {
			m.windowParent.UpdateSize()
		}
	}
}

func (m *Chromium) closeQuery(sender lcl.IObject, canClose *bool) {
	fmt.Println("closeQuery")
	*canClose = m.canClose
	if !m.canClose {
		m.canClose = true
		m.chromium.CloseBrowser(true)
	}
}

func (m *Chromium) chromiumClose(sender lcl.IObject, browser cef.ICefBrowser, aAction *cefTypes.TCefCloseBrowserAction) {
	fmt.Println("chromium.Close")
	if tool.IsDarwin() {
		m.windowParent.DestroyChildWindow()
		*aAction = cefTypes.CbaClose
	} else {
		*aAction = cefTypes.CbaDelay
		lcl.RunOnMainThreadAsync(func(id uint32) {
			m.windowParent.Free()
		})
	}
}

func (m *Chromium) chromiumBeforeClose(sender lcl.IObject, browser cef.ICefBrowser) {
	fmt.Println("chromium.BeforeClose")
	m.canClose = true
	if tool.IsDarwin() {
		//m.Close()
	} else {
		//rtl.PostMessage(m.Handle(), messages.WM_CLOSE, 0, 0)
	}
}

func (m *Chromium) SetOnAfterCreated(fn ChromiumAfterCreate) {
	m.afterCreate = fn
}

func (m *Chromium) updateTabSheetActive(isActive bool) {
	if m.tabSheet == nil {
		return
	}
	if isActive {
		activeColor := colors.RGBToColor(86, 88, 93)
		m.tabSheet.SetStartColor(activeColor)
		m.tabSheet.SetEndColor(activeColor)
		m.windowParent.SetVisible(true)
		m.isActive = true
		lcl.RunOnMainThreadAsync(func(id uint32) {
			m.mainWindow.addr.SetText(m.currentURL)
		})
	} else {
		notActiveColor := colors.RGBToColor(56, 57, 60)
		m.tabSheet.SetStartColor(notActiveColor)
		m.tabSheet.SetEndColor(notActiveColor)
		m.windowParent.SetVisible(false)
		m.isActive = false
	}
	m.tabSheet.Invalidate()
}

func (m *Chromium) closeBrowser() {
	m.windowParent.SetVisible(false)
	m.chromium.CloseBrowser(true)
	m.tabSheet.Free()
}

func (m *BrowserWindow) createChromium(url string) *Chromium {
	newChromium := &Chromium{mainWindow: m}

	newChromium.chromium = cef.NewChromium(m)
	newChromium.chromium.SetDefaultUrl(url)
	if tool.IsWindows() {
		newChromium.windowParent = cef.NewWindowParent(m)
	} else {
		windowParent := cef.NewLinkedWindowParent(m)
		windowParent.SetChromium(newChromium.chromium)
		newChromium.windowParent = windowParent
	}
	newChromium.windowParent.SetParent(m.content)
	newChromium.windowParent.SetDoubleBuffered(true)
	newChromium.windowParent.SetAlign(types.AlClient)
	// 创建一个定时器, 用来createBrowser
	newChromium.timer = lcl.NewTimer(m)
	newChromium.timer.SetEnabled(false)
	newChromium.timer.SetInterval(200)
	newChromium.timer.SetOnTimer(newChromium.createBrowser)

	m.content.SetOnResize(newChromium.resize)
	m.content.SetOnEnter(func(sender lcl.IObject) {
		newChromium.chromium.Initialized()
		newChromium.chromium.FrameIsFocused()
		newChromium.chromium.SetFocus(true)
	})

	newChromium.windowParent.SetOnExit(func(sender lcl.IObject) {
		newChromium.chromium.SendCaptureLostEvent()
	})

	// 2. 触发后控制延迟关闭, 在UI线程中调用 windowParent.Free() 释放对象，然后触发 chromium.SetOnBeforeClose
	newChromium.chromium.SetOnClose(newChromium.chromiumClose)
	// 3. 触发后将canClose设置为true, 发送消息到主窗口关闭，触发 m.SetOnCloseQuery
	newChromium.chromium.SetOnBeforeClose(newChromium.chromiumBeforeClose)

	newChromium.chromium.SetOnAfterCreated(func(sender lcl.IObject, browser cef.ICefBrowser) {
		newChromium.windowId = browser.GetIdentifier()
		newChromium.windowParent.UpdateSize()
		if newChromium.afterCreate != nil {
			newChromium.afterCreate(newChromium)
		}
	})
	newChromium.chromium.SetOnBeforeBrowse(func(sender lcl.IObject, browser cef.ICefBrowser, frame cef.ICefFrame, request cef.ICefRequest,
		userGesture, isRedirect bool, result *bool) {
		newChromium.windowParent.UpdateSize()
	})
	newChromium.chromium.SetOnBeforePopup(func(sender lcl.IObject, browser cef.ICefBrowser, frame cef.ICefFrame,
		popupId int32, targetUrl string, targetFrameName string, targetDisposition cefTypes.TCefWindowOpenDisposition, userGesture bool,
		popupFeatures cef.TCefPopupFeatures, windowInfo *cef.TCefWindowInfo, client *cef.IEngClient, settings *cef.TCefBrowserSettings,
		extraInfo *cef.ICefDictionaryValue, noJavascriptAccess *bool, result *bool) {
		*result = true
		newChromium.chromium.LoadURLWithStringFrame(targetUrl, frame)
	})
	newChromium.chromium.SetOnTitleChange(func(sender lcl.IObject, browser cef.ICefBrowser, title string) {
		if newChromium.tabSheet != nil {
			if title == "about:blank" {
				return
			}
			lcl.RunOnMainThreadAsync(func(id uint32) {
				newChromium.tabSheet.SetCaption(title)
				newChromium.tabSheet.Invalidate()
			})
		}
	})
	newChromium.chromium.SetOnLoadStart(func(sender lcl.IObject, browser cef.ICefBrowser, frame cef.ICefFrame, transitionType cefTypes.TCefTransitionType) {
		tempUrl := frame.GetUrl()
		if tempUrl == "about:blank" {
			return
		}
		newChromium.currentURL = tempUrl
		lcl.RunOnMainThreadAsync(func(id uint32) {
			m.addr.SetText(tempUrl)
		})
	})
	newChromium.chromium.SetOnLoadEnd(func(sender lcl.IObject, browser cef.ICefBrowser, frame cef.ICefFrame, httpStatusCode int32) {
	})
	return newChromium
}

// 创建浏览器关联的 tab sheet
func (m *BrowserWindow) createTabSheet() {

}
