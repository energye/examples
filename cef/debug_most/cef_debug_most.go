package main

import (
	"fmt"
	"github.com/energye/cef/cef"
	cefTypes "github.com/energye/cef/types"
	"github.com/energye/examples/cef/debug_most/application"
	"github.com/energye/examples/cef/debug_most/contextmenu"
	"github.com/energye/examples/cef/debug_most/cookie"
	"github.com/energye/examples/cef/debug_most/devtools"
	"github.com/energye/examples/cef/debug_most/scheme"
	"github.com/energye/examples/cef/debug_most/utils"
	. "github.com/energye/examples/syso"
	"github.com/energye/lcl/api"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/rtl"
	"github.com/energye/lcl/tool"
	"github.com/energye/lcl/types"
	"github.com/energye/lcl/types/messages"
	"path/filepath"
	"unsafe"
)

type BrowserWindow struct {
	lcl.TEngForm
	mainWindowId int32 // 主窗口ID
	timer        lcl.ITimer
	windowParent cef.ICEFWinControl
	chromium     cef.IChromium
	canClose     bool
	ChildForm    lcl.IForm
}

var (
	BW   BrowserWindow
	help string //= "true" // go build -ldflags="-X main.help=true"
)

func init() {
	TestLoadLibPath()
}

func main() {
	//全局初始化 每个应用都必须调用的
	cef.Init(nil, nil)
	app := application.NewApplication()
	if tool.IsDarwin() {
		app.SetUseMockKeyChain(true)
		app.InitLibLocationFromArgs()
		// MacOS
		cef.AddCrDelegate()
		scheduler := cef.NewWorkScheduler(nil)
		cef.SetGlobalCEFWorkSchedule(scheduler)

		app.SetOnScheduleMessagePumpWork(nil)
		app.SetExternalMessagePump(true)
		app.SetMultiThreadedMessageLoop(false)
		if app.ProcessType() != cefTypes.PtBrowser {
			// MacOS 多进程时，需要调用StartSubProcess来启动子进程
			subStart := app.StartSubProcess()
			fmt.Println("subStart:", subStart, app.ProcessType())
			app.Free()
			return
		}
	} else { // MacOS不需要设置CEF框架目录，它是一个固定的目录结构
		if help == "true" {
			subexe := filepath.Join(utils.RootPath(), "helper", "helper.exe")
			app.SetBrowserSubprocessPath(subexe)
		}
	}
	// 主进程启动
	mainStart := app.StartMainProcess()
	fmt.Println("mainStart:", mainStart, app.ProcessType())
	if mainStart {
		// 结束应用后释放资源
		api.SetReleaseCallback(func() {
			fmt.Println("Release")
			app.Free()
		})
		// LCL窗口
		lcl.Application.Initialize()
		lcl.Application.SetMainFormOnTaskBar(true)
		lcl.Application.NewForm(&BW)
		lcl.Application.Run()
	}
	fmt.Println("app free")
}

func (m *BrowserWindow) FormCreate(sender lcl.IObject) {
	m.SetWidth(1024)
	m.SetHeight(768)
	m.ScreenCenter()
	m.SetCaption("Energy3.0 - CEF simple")
	m.chromium = cef.NewChromium(m)
	var assetsHtml string
	if tool.IsDarwin() {
		assetsHtml = filepath.Join("file://", utils.RootPath(), "debug_most", "assets", "index.html")
	} else {
		assetsHtml = filepath.Join(utils.RootPath(), "debug_most", "assets", "index.html")
	}
	fmt.Println("assetsHtml:", assetsHtml)
	m.chromium.SetDefaultUrl(assetsHtml)
	//m.chromium.SetDefaultUrl("https://www.baidu.com")
	if tool.IsWindows() {
		m.windowParent = cef.NewWindowParent(m)
	} else {
		windowParent := cef.NewLinkedWindowParent(m)
		windowParent.SetChromium(m.chromium)
		m.windowParent = windowParent
	}
	m.windowParent.SetParent(m)
	m.windowParent.SetAlign(types.AlClient)
	m.windowParent.SetAnchors(types.NewSet(types.AkTop, types.AkLeft, types.AkRight, types.AkBottom))
	// 创建一个定时器, 用来createBrowser
	m.timer = lcl.NewTimer(m)
	m.timer.SetEnabled(false)
	m.timer.SetInterval(200)
	m.timer.SetOnTimer(m.createBrowser)
	// 在show时创建chromium browser
	m.TForm.SetOnShow(m.show)
	m.TForm.SetOnActivate(m.active)
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

	// 上下文菜单
	contextmenu.ContextMenu(m.chromium)
	// cookie
	cookie.Cookie(m.chromium)
	// devtools
	devtools.DevTools(m.chromium)

	m.chromium.SetOnLoadingProgressChange(func(sender lcl.IObject, browser cef.ICefBrowser, progress float64) {
		fmt.Println("OnLoadingProgressChange:", progress)
	})
	m.chromium.SetOnLoadStart(func(sender lcl.IObject, browser cef.ICefBrowser, frame cef.ICefFrame, transitionType cefTypes.TCefTransitionType) {
		fmt.Println("OnLoadStart:", frame.GetUrl())
	})
	m.chromium.SetOnLoadEnd(func(sender lcl.IObject, browser cef.ICefBrowser, frame cef.ICefFrame, httpStatusCode int32) {
		requestCtx := browser.GetHost().GetRequestContext()
		manager := requestCtx.GetCookieManager(nil)
		// 使用 chromium 事件
		manager.VisitAllCookies(cef.AsEngCookieVisitor(cef.NewCustomCookieVisitor(m.chromium, 0)))
		// 使用 Eng 事件
		//manager.VisitAllCookies(cef.NewEngCookieVisitor())
		manager.FreeAndNil()
	})
	m.chromium.SetOnAfterCreated(func(sender lcl.IObject, browser cef.ICefBrowser) {
		fmt.Println("SetOnAfterCreated 1")
		lcl.RunOnMainThreadAsync(func(id uint32) {
			fmt.Println("SetOnAfterCreated 2")
		})
		fmt.Println("SetOnAfterCreated 3")
		if m.mainWindowId == 0 {
			m.mainWindowId = browser.GetIdentifier()
		}
		m.windowParent.UpdateSize()
		scheme.ChromiumAfterCreated(browser)
	})
	m.chromium.SetOnBeforeBrowse(func(sender lcl.IObject, browser cef.ICefBrowser, frame cef.ICefFrame, request cef.ICefRequest,
		userGesture, isRedirect bool, result *bool) {
		fmt.Println("SetOnBeforeBrowser")
		m.windowParent.UpdateSize()
	})
	m.chromium.SetOnDragEnter(func(sender lcl.IObject, browser cef.ICefBrowser, dragData cef.ICefDragData, mask cefTypes.TCefDragOperations, outResult *bool) {
		if mask&cefTypes.DRAG_OPERATION_LINK == cefTypes.DRAG_OPERATION_LINK {
			var fileNameList lcl.IStrings
			fmt.Println("SetOnDragEnter", mask&cefTypes.DRAG_OPERATION_LINK, dragData.IsLink(), dragData.IsFile(), "GetFileName:", dragData.GetFileName(),
				"GetFileNames:", dragData.GetFileNames(&fileNameList))
			if fileNameList != nil {
				count := int(fileNameList.Count())
				fileNames := make([]string, count, count)
				for i := 0; i < count; i++ {
					fileNames[i] = fileNameList.Strings(int32(i))
				}
				fileNameList.Free()
				fmt.Println("count:", count, "fileNames:", fileNames)
			}
			*outResult = false
		} else {
			*outResult = true
		}
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

	m.chromium.SetOnDownloadUpdated(func(sender lcl.IObject, browser cef.ICefBrowser, downloadItem cef.ICefDownloadItem, callback cef.ICefDownloadItemCallback) {
		fmt.Println("DownloadUpdated frameId", browser.GetMainFrame().GetIdentifier(), "Id:", downloadItem.GetId(), "originalUrl:", downloadItem.GetOriginalUrl(), "url:", downloadItem.GetUrl())
		fmt.Println("\t", downloadItem.GetTotalBytes(), "/", downloadItem.GetReceivedBytes(), "speed:", downloadItem.GetCurrentSpeed(), "fullPath:", downloadItem.GetFullPath())
	})

	m.chromium.SetOnBeforeResourceLoad(func(sender lcl.IObject, browser cef.ICefBrowser, frame cef.ICefFrame, request cef.ICefRequest, callback cef.ICefCallback, result *cefTypes.TCefReturnValue) {
		fmt.Println("SetOnBeforeResourceLoad")
		headerMap := cef.NewCustomStringMultimap()
		request.GetHeaderMap(headerMap)
		fmt.Println("headerMap size:", headerMap.GetSize())
		var key, val string
		for i := 0; i < int(headerMap.GetSize()); i++ {
			key = headerMap.GetKey(uint32(i))
			val = headerMap.GetValue(uint32(i))
			if key != "" {
				fmt.Println("  key:", key, "val:", val)
			}
		}
		headerMap.Free()
		//callback.Cont()
	})
	m.chromium.SetOnProcessMessageReceived(func(sender lcl.IObject, browser cef.ICefBrowser, frame cef.ICefFrame, sourceProcess cefTypes.TCefProcessId,
		message cef.ICefProcessMessage, outResult *bool) {
		fmt.Println("主进程 name:", message.GetName())
		defer message.FreeAndNil()
		if message.GetName() == "jsreturn" {

		} else if message.GetName() == "cookieVisited" {
			cookie.CookieVisited(m.chromium)
		} else if message.GetName() == "cookieDelete" {
			cookie.DeleteCookie(m.chromium)
		} else if message.GetName() == "setCookie" {
			cookie.SetCookie(m.chromium)
		} else if message.GetName() == "showDevtools" {
			devtools.ShowDevtools(m.chromium)
		} else if message.GetName() == "executeDevToolsMethod" {
			devtools.ExecuteDevToolsMethod(m.chromium)
		} else if message.GetName() == "executeJavaScript" {
			devtools.ExecuteJavaScript(m.chromium)
		} else {
			args := message.GetArgumentList()
			binArgs := args.GetBinary(0)
			fmt.Println("size:", binArgs.GetSize())
			messageDataBytes := make([]byte, int(binArgs.GetSize()))
			binArgs.GetData(uintptr(unsafe.Pointer(&messageDataBytes[0])), binArgs.GetSize(), 0)
			fmt.Println("data:", string(messageDataBytes))
			binArgs.FreeAndNil()
			args.FreeAndNil()

			// 消息发送到渲染进程
			dataBytes := []byte("OK收到: " + string(messageDataBytes))
			processMessage := cef.ProcessMessageRef.New("send-render")
			messageArgumentList := processMessage.GetArgumentList()
			dataBin := cef.BinaryValueRef.New(uintptr(unsafe.Pointer(&dataBytes[0])), uint32(len(dataBytes)))
			messageArgumentList.SetBinary(0, dataBin)
			frame.SendProcessMessage(cefTypes.PID_BROWSER, processMessage)
			dataBin.FreeAndNil()
			messageArgumentList.Clear()
			messageArgumentList.FreeAndNil()
			processMessage.FreeAndNil()
		}
	})
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
	fmt.Println("resize")
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
		//m.SetVisible(false)
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
