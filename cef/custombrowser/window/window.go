package window

import (
	"fmt"
	"github.com/energye/cef/cef"
	cefTypes "github.com/energye/cef/types"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/rtl"
	"github.com/energye/lcl/tool"
	"github.com/energye/lcl/types"
	"github.com/energye/lcl/types/colors"
	"github.com/energye/lcl/types/messages"
)

type BrowserWindow struct {
	lcl.TEngForm
	mainWindowId int32 // 主窗口ID
	timer        lcl.ITimer
	windowParent cef.ICEFWinControl
	chromium     cef.IChromium
	canClose     bool
	oldWndPrc    uintptr
}

var (
	BW BrowserWindow
)

func (m *BrowserWindow) FormCreate(sender lcl.IObject) {
	m.SetWidth(1024)
	m.SetHeight(768)
	m.SetDoubleBuffered(true)
	//m.SetColor(colors.ClYellow)
	m.SetColor(colors.RGBToColor(56, 57, 60))
	m.ScreenCenter()
	m.SetCaption("Energy3.0 - CEF simple")
	m.chromium = cef.NewChromium(m)
	assetsHtml := "https://www.baidu.com"
	fmt.Println("assetsHtml:", assetsHtml)
	m.chromium.SetDefaultUrl(assetsHtml)
	if tool.IsWindows() {
		m.windowParent = cef.NewWindowParent(m)
	} else {
		windowParent := cef.NewLinkedWindowParent(m)
		windowParent.SetChromium(m.chromium)
		m.windowParent = windowParent
	}
	m.windowParent.SetParent(m)
	//m.windowParent.SetAlign(types.AlClient)
	m.windowParent.SetWidth(m.Width() - 10)
	m.windowParent.SetTop(45)
	m.windowParent.SetLeft(5)
	m.windowParent.SetHeight(m.Height() - 50)
	m.windowParent.SetAnchors(types.NewSet(types.AkTop, types.AkLeft, types.AkRight, types.AkBottom))
	// 创建一个定时器, 用来createBrowser
	m.timer = lcl.NewTimer(m)
	m.timer.SetEnabled(false)
	m.timer.SetInterval(200)
	m.timer.SetOnTimer(m.createBrowser)
	// 在show时创建chromium browser
	if tool.IsLinux() {
		// Linux需要一个可见的表单来创建浏览器，因此我们需要使用 TForm。OnActivate事件而不是TForm.OnShow
		m.TForm.SetOnActivate(m.active)
	} else {
		m.TForm.SetOnShow(m.show)
	}
	m.TForm.SetOnResize(m.resize)
	m.windowParent.SetOnEnter(func(sender lcl.IObject) {
		m.chromium.Initialized()
		m.chromium.FrameIsFocused()
		m.chromium.SetFocus(true)
	})
	m.windowParent.SetOnExit(func(sender lcl.IObject) {
		m.chromium.SendCaptureLostEvent()
	})
	// 1. 关闭之前先调用chromium.CloseBrowser(true)，然后触发 chromium.SetOnClose
	m.TForm.SetOnCloseQuery(m.closeQuery)
	// 2. 触发后控制延迟关闭, 在UI线程中调用 windowParent.Free() 释放对象，然后触发 chromium.SetOnBeforeClose
	m.chromium.SetOnClose(m.chromiumClose)
	// 3. 触发后将canClose设置为true, 发送消息到主窗口关闭，触发 m.SetOnCloseQuery
	m.chromium.SetOnBeforeClose(m.chromiumBeforeClose)

	m.chromium.SetOnAfterCreated(func(sender lcl.IObject, browser cef.ICefBrowser) {
		if m.mainWindowId == 0 {
			m.mainWindowId = browser.GetIdentifier()
		}
		m.windowParent.UpdateSize()
	})
	m.chromium.SetOnBeforeBrowse(func(sender lcl.IObject, browser cef.ICefBrowser, frame cef.ICefFrame, request cef.ICefRequest,
		userGesture, isRedirect bool, result *bool) {
		m.windowParent.UpdateSize()
	})
	m.chromium.SetOnBeforePopup(func(sender lcl.IObject, browser cef.ICefBrowser, frame cef.ICefFrame,
		popupId int32, targetUrl string, targetFrameName string, targetDisposition cefTypes.TCefWindowOpenDisposition, userGesture bool,
		popupFeatures cef.TCefPopupFeatures, windowInfo *cef.TCefWindowInfo, client *cef.IEngClient, settings *cef.TCefBrowserSettings,
		extraInfo *cef.ICefDictionaryValue, noJavascriptAccess *bool, result *bool) {
		fmt.Printf("beforePopup: %+v\n", windowInfo)
		fmt.Printf("popupFeatures: %+v\n", popupFeatures)
		fmt.Println(browser.GetIdentifier())
		fmt.Println(frame.GetIdentifier(), frame.GetUrl())
		v8ctx := frame.GetV8Context()
		if v8ctx != nil {
			fmt.Println(frame.GetV8Context())
			fmt.Println(frame.GetV8Context().GetFrame().GetUrl())
		}
		settings.DefaultFontSize = 36
		settings.StandardFontFamily = "微软雅黑"
		windowInfo.Bounds = cef.TCefRect{X: 400, Y: 10, Width: 400, Height: 400}
		windowInfo.WindowName = "杨杨红红岩岩"
	})

}

func (m *BrowserWindow) FormAfterCreate(sender lcl.IObject) {
	m.HookWndProcMessage()
}

func (m *BrowserWindow) createBrowser(sender lcl.IObject) {
	if m.timer == nil {
		return
	}
	m.timer.SetEnabled(false)
	rect := m.ClientRect()
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

func (m *BrowserWindow) active(sender lcl.IObject) {
	fmt.Println("window active")
	m.createBrowser(sender)
}

func (m *BrowserWindow) show(sender lcl.IObject) {
	fmt.Println("window show")
	m.createBrowser(sender)
}

func (m *BrowserWindow) resize(sender lcl.IObject) {
	if m.chromium != nil {
		m.chromium.NotifyMoveOrResizeStarted()
		if m.windowParent != nil {
			m.windowParent.UpdateSize()
		}
	}
}
func (m *BrowserWindow) closeQuery(sender lcl.IObject, canClose *bool) {
	fmt.Println("closeQuery")
	*canClose = m.canClose
	if !m.canClose {
		m.canClose = true
		m.chromium.CloseBrowser(true)
	}
}

func (m *BrowserWindow) chromiumClose(sender lcl.IObject, browser cef.ICefBrowser, aAction *cefTypes.TCefCloseBrowserAction) {
	fmt.Println("chromiumClose id:", browser.GetIdentifier(), "mainWindowId:", m.mainWindowId)
	if browser.GetIdentifier() == m.mainWindowId {
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
}

func (m *BrowserWindow) chromiumBeforeClose(sender lcl.IObject, browser cef.ICefBrowser) {
	fmt.Println("chromiumBeforeClose id:", browser.GetIdentifier(), "mainWindowId:", m.mainWindowId)
	if browser.GetIdentifier() == m.mainWindowId {
		m.canClose = true
		if tool.IsDarwin() {
			m.Close()
		} else {
			rtl.PostMessage(m.Handle(), messages.WM_CLOSE, 0, 0)
		}
	}
}
