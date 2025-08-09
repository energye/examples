package window

import (
	"fmt"
	"github.com/energye/cef/cef"
	cefTypes "github.com/energye/cef/types"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/tool"
	"github.com/energye/lcl/types"
	"github.com/energye/lcl/types/colors"
	"strings"
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
	isLoading    bool
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
		if url == "" {
			newChromium.chromium.LoadStringWithStringFrame(defaultHTML, browser.GetMainFrame())
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
			if title == "about:blank" || strings.Index(title, "data:text/html") != -1 {
				title = "新建标签页"
			}
			lcl.RunOnMainThreadAsync(func(id uint32) {
				newChromium.tabSheet.SetCaption(title)
				newChromium.tabSheet.SetHint(title)
				newChromium.tabSheet.Invalidate()
			})
		}
	})
	newChromium.chromium.SetOnLoadingStateChange(func(sender lcl.IObject, browser cef.ICefBrowser, isLoading bool, canGoBack bool, canGoForward bool) {
		newChromium.isLoading = isLoading
		if isLoading {
			lcl.RunOnMainThreadAsync(func(id uint32) {
				newChromium.mainWindow.refreshBtn.SetIcon(getImageResourcePath("stop.png"))
			})
		} else {
			lcl.RunOnMainThreadAsync(func(id uint32) {
				newChromium.mainWindow.refreshBtn.SetIcon(getImageResourcePath("refresh.png"))
			})
		}
	})
	newChromium.chromium.SetOnLoadStart(func(sender lcl.IObject, browser cef.ICefBrowser, frame cef.ICefFrame, transitionType cefTypes.TCefTransitionType) {
		tempUrl := browser.GetMainFrame().GetUrl()
		if tempUrl == "about:blank" || strings.Index(tempUrl, "data:text/html") != -1 {
			tempUrl = ""
		}
		newChromium.currentURL = tempUrl
		lcl.RunOnMainThreadAsync(func(id uint32) {
			m.addr.SetText(tempUrl)
			m.addr.SetFocus()
		})
	})
	return newChromium
}

var defaultHTML = "<!DOCTYPE html>\n<html lang=\"zh-CN\">\n<head>\n  <meta charset=\"UTF-8\">\n  <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\n  <link rel=\"stylesheet\" href=\"https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css\">\n  <style>\n    * {\n      margin: 0;\n      padding: 0;\n      box-sizing: border-box;\n      font-family: 'Segoe UI', 'SF Pro Display', -apple-system, BlinkMacSystemFont, sans-serif;\n    }\n\n    body {\n      background: linear-gradient(135deg, #0f172a 0%, #1e293b 100%);\n      color: #e2e8f0;\n      display: flex;\n      flex-direction: column;\n      align-items: center;\n      justify-content: center;\n      overflow: hidden;\n      position: relative;\n    }\n\n    /* 背景装饰元素 */\n    .background-elements {\n      position: absolute;\n      top: 0;\n      left: 0;\n      width: 100%;\n      height: 100%;\n      z-index: -1;\n      overflow: hidden;\n    }\n\n    .circle {\n      position: absolute;\n      border-radius: 50%;\n      background: radial-gradient(circle, rgba(56, 189, 248, 0.15) 0%, transparent 70%);\n    }\n\n    .circle:nth-child(1) {\n      width: 300px;\n      height: 300px;\n      top: 10%;\n      left: 15%;\n    }\n\n    .circle:nth-child(2) {\n      width: 200px;\n      height: 200px;\n      bottom: 20%;\n      right: 20%;\n    }\n\n    .circle:nth-child(3) {\n      width: 150px;\n      height: 150px;\n      top: 40%;\n      right: 30%;\n    }\n\n    /* 主内容区 */\n    .main-container {\n      text-align: center;\n      padding: 2rem;\n      max-width: 800px;\n      max-height: 600px;\n      z-index: 10;\n    }\n\n    .logo {\n      display: flex;\n      flex-direction: column;\n      align-items: center;\n      margin-bottom: 2.5rem;\n    }\n\n    .logo-icon {\n      width: 120px;\n      height: 120px;\n      background: linear-gradient(135deg, #3b82f6 0%, #0ea5e9 100%);\n      border-radius: 24px;\n      display: flex;\n      align-items: center;\n      justify-content: center;\n      box-shadow: 0 10px 25px rgba(14, 165, 233, 0.3);\n      margin-bottom: 1.5rem;\n      transform: rotate(45deg);\n      transition: all 0.3s ease;\n    }\n\n    .logo-icon:hover {\n      transform: rotate(0deg) scale(1.05);\n      box-shadow: 0 15px 30px rgba(14, 165, 233, 0.4);\n    }\n\n    .logo-icon span {\n      transform: rotate(-45deg);\n      font-size: 3.5rem;\n      font-weight: 700;\n      color: white;\n    }\n\n    .logo-text {\n      font-size: 3.5rem;\n      font-weight: 800;\n      background: linear-gradient(to right, #38bdf8, #3b82f6);\n      -webkit-background-clip: text;\n      -webkit-text-fill-color: transparent;\n      letter-spacing: -1px;\n    }\n\n\n    .tech-stack {\n      display: flex;\n      justify-content: center;\n      gap: 2.5rem;\n      margin: 2.5rem 0;\n    }\n\n    .tech-item {\n      display: flex;\n      flex-direction: column;\n      align-items: center;\n      gap: 0.8rem;\n    }\n\n    .tech-icon {\n      width: 70px;\n      height: 70px;\n      background: rgba(30, 41, 59, 0.7);\n      border-radius: 16px;\n      display: flex;\n      align-items: center;\n      justify-content: center;\n      font-size: 2.5rem;\n      color: #38bdf8;\n      box-shadow: 0 4px 15px rgba(0, 0, 0, 0.2);\n      transition: all 0.3s ease;\n    }\n\n    .tech-icon:hover {\n      transform: translateY(-5px);\n      background: rgba(56, 189, 248, 0.15);\n    }\n\n    .tech-name {\n      font-size: 1.1rem;\n      font-weight: 500;\n      color: #94a3b8;\n    }\n\n    .description {\n      background: rgba(15, 23, 42, 0.6);\n      backdrop-filter: blur(10px);\n      border-radius: 16px;\n      padding: 2rem;\n      border: 1px solid rgba(56, 189, 248, 0.1);\n      box-shadow: 0 10px 30px rgba(0, 0, 0, 0.3);\n      text-align: center;\n      line-height: 1.7;\n    }\n\n    .description p {\n      margin-bottom: 1.2rem;\n      font-size: 1.1rem;\n      color: #cbd5e1;\n    }\n\n    .highlight {\n      color: #38bdf8;\n      font-weight: 600;\n    }\n\n    .feature i {\n      color: #3b82f6;\n      margin-right: 0.5rem;\n    }\n\n    .footer {\n      margin-top: 3rem;\n      color: #64748b;\n      font-size: 0.9rem;\n    }\n\n    /* 动画效果 */\n    @keyframes float {\n      0%, 100% { transform: translateY(0); }\n      50% { transform: translateY(-10px); }\n    }\n\n    .floating {\n      animation: float 6s ease-in-out infinite;\n    }\n\n    /* 响应式设计 */\n    @media (max-width: 768px) {\n      .tech-stack {\n        flex-wrap: wrap;\n        gap: 1.5rem;\n      }\n\n      .logo-text {\n        font-size: 2.8rem;\n      }\n    }\n  </style>\n</head>\n<body>\n<!-- 背景装饰元素 -->\n<div class=\"background-elements\">\n  <div class=\"circle floating\" style=\"animation-delay: 0s;\"></div>\n  <div class=\"circle floating\" style=\"animation-delay: 2s;\"></div>\n  <div class=\"circle floating\" style=\"animation-delay: 4s;\"></div>\n</div>\n\n<div class=\"main-container\">\n  <div class=\"logo\">\n    <div class=\"logo-icon\">\n      <span>E</span>\n    </div>\n    <h1 class=\"logo-text\">ENERGY</h1>\n  </div>\n\n  <div class=\"tech-stack\">\n    <div class=\"tech-item\">\n      <div class=\"tech-icon\">\n        <i class=\"fab fa-golang\"></i>\n      </div>\n      <div class=\"tech-name\">Go语言</div>\n    </div>\n    <div class=\"tech-item\">\n      <div class=\"tech-icon\">\n        <i class=\"fas fa-window-restore\"></i>\n      </div>\n      <div class=\"tech-name\">LCL</div>\n    </div>\n    <div class=\"tech-item\">\n      <div class=\"tech-icon\">\n        <i class=\"fas fa-compass\"></i>\n      </div>\n      <div class=\"tech-name\">CEF</div>\n    </div>\n    <div class=\"tech-item\">\n      <div class=\"tech-icon\">\n        <i class=\"fas fa-bolt\"></i>\n      </div>\n      <div class=\"tech-name\">高性能</div>\n    </div>\n  </div>\n\n  <div class=\"description\">\n    <p>ENERGY是一个基于Go语言构建的现代化桌面应用框架，结合了<span class=\"highlight\">LCL（Lazarus组件库）</span>的原生UI能力和<span class=\"highlight\">CEF（Chromium Embedded Framework）</span>的Web渲染能力。</p>\n    <p>该框架旨在为开发者提供创建跨平台桌面应用程序的终极解决方案，兼具原生应用的性能和Web技术的灵活性。</p>\n  </div>\n\n  <div class=\"footer\">\n    <p>© 2025 ENERGY Framework | 输入网址或开始创建您的应用程序</p>\n  </div>\n</div>\n\n<script>\n  // 添加简单的交互效果\n  document.addEventListener('DOMContentLoaded', function() {\n    const techIcons = document.querySelectorAll('.tech-icon');\n\n    techIcons.forEach(icon => {\n      icon.addEventListener('mouseenter', () => {\n        icon.style.transform = 'translateY(-8px)';\n      });\n\n      icon.addEventListener('mouseleave', () => {\n        icon.style.transform = 'translateY(0)';\n      });\n    });\n  });\n</script>\n</body>\n</html>"
