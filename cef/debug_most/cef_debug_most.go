package main

import (
	"fmt"
	"github.com/energye/cef/cef"
	cefTypes "github.com/energye/cef/types"
	"github.com/energye/examples/cef/application"
	"github.com/energye/examples/cef/debug_most/contextmenu"
	"github.com/energye/examples/cef/debug_most/cookie"
	"github.com/energye/examples/cef/debug_most/devtools"
	"github.com/energye/examples/cef/debug_most/scheme"
	"github.com/energye/examples/cef/debug_most/v8context"
	"github.com/energye/examples/cef/utils"
	. "github.com/energye/examples/syso"
	"github.com/energye/lcl/api"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/rtl"
	"github.com/energye/lcl/tool"
	"github.com/energye/lcl/types"
	"github.com/energye/lcl/types/messages"
	"os"
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
	BW               BrowserWindow
	help             string //= "true" // go build -ldflags="-X main.help=true"
	wd, _            = os.Getwd()
	cacheRoot        = filepath.Join(wd, "EnergyCache")         // 浏览器缓存目录
	siteResourceRoot = filepath.Join(cacheRoot, "SiteResource") // 网站资源缓存目录
)

func init() {
	TestLoadLibPath()
}

func main() {
	//全局初始化 每个应用都必须调用的
	cef.Init(nil, nil)
	if tool.IsDarwin() {
		cef.AddCrDelegate()
	}
	app := application.NewApplication()
	app.SetLogSeverity(cefTypes.LOGSEVERITY_VERBOSE)
	app.SetRootCache(cacheRoot)
	app.SetCache(cacheRoot)
	//fmt.Println("ProcessType:", app.ProcessType())
	v8context.Context(app)
	app.SetOnRegCustomSchemes(func(registrar cef.ICefSchemeRegistrarRef) {
		scheme.ApplicationOnRegCustomSchemes(registrar)
	})
	app.SetOnBeforeChildProcessLaunch(func(commandLine cef.ICefCommandLine) {
		fmt.Println("SetOnBeforeChildProcessLaunch")
		//commandLine.AppendSwitch("--enable-gpu-memory-buffer-compositor-resources")
		//commandLine.AppendSwitch("--enable-main-frame-before-activation")
	})
	if tool.IsDarwin() {
		app.InitLibLocationFromArgs()
		// MacOS不需要设置CEF框架目录，它是一个固定的目录结构
		app.SetUseMockKeyChain(true)
		app.SetExternalMessagePump(true)
		app.SetMultiThreadedMessageLoop(false)
		if app.ProcessType() == cefTypes.PtBrowser {
			scheduler := cef.NewWorkScheduler(nil)
			cef.SetGlobalCEFWorkSchedule(scheduler)
			//messagepump.GlobalCEFApp = app
			//messagepump.InitMessagePump()
			app.SetOnScheduleMessagePumpWork(func(delayMs int64) {
				//fmt.Println("IsMainThread:", messagepump.IsMainThread(), "delayMs:", delayMs)
				//fmt.Println("OnScheduleMessagePumpWork delayMs:", delayMs)
				scheduler.ScheduleMessagePumpWork(delayMs)
				//messagepump.OnScheduleMessagePumpWork(delayMs)
			})
		} else {
			startSub := app.StartSubProcess()
			fmt.Println("startSub:", startSub)
			return
		}
	} else if tool.IsLinux() {
		if api.Widget().IsGTK2() {
			// gtk2 使用 lcl 窗口
			app.SetExternalMessagePump(false)
			app.SetMultiThreadedMessageLoop(true)
		} else if api.Widget().IsGTK3() {
			// gtk3 使用 vf 窗口
			println("当前 demo 为 CEF LCL GTK2, EXIT.")
			os.Exit(1)
		}
		// 这是一个解决“GPU不可用错误”问题的方法 linux
		// https://bitbucket.org/chromiumembedded/cef/issues/2964/gpu-is-not-usable-error-during-cef
		app.SetDisableZygote(true)
	} else if tool.IsWindows() {
		// win32 使用 lcl 窗口
		app.SetExternalMessagePump(false)
		app.SetMultiThreadedMessageLoop(true)
	}

	// 主进程启动
	mainStart := app.StartMainProcess()
	fmt.Println("mainStart:", mainStart, app.ProcessType())
	if mainStart {
		// 结束应用后释放资源
		api.SetReleaseCallback(func() {
			fmt.Println("Release")
			if tool.IsLinux() {
				api.WidgetSetFinalization()
			}
		})
		// LCL窗口
		lcl.Application.Initialize()
		lcl.Application.SetMainFormOnTaskBar(true)
		lcl.Application.NewForm(&BW)
		lcl.Application.Run()
	}
}

func (m *BrowserWindow) FormCreate(sender lcl.IObject) {
	m.SetWidth(1024)
	m.SetHeight(768)
	m.ScreenCenter()
	m.SetCaption("Energy3.0 - CEF simple")
	m.chromium = cef.NewChromium(m)
	var assetsHtml string
	if tool.IsDarwin() {
		assetsHtml = "file:///Users/yanghy/app/workspace/examples/cef/debug_most/assets/index.html"
		//assetsHtml = "https://www.baidu.com"
		//assetsHtml = "https://www.bilibili.com/"
		//assetsHtml = "https://www.google.com/"
		//assetsHtml = "https://www.lazarus-ide.org"
	} else if tool.IsLinux() {
		assetsHtml = "file:///home/yanghy/app/gopath/src/github.com/energye/workspace/examples/cef/debug_most/assets/index.html"
		//assetsHtml = "https://www.baidu.com"
	} else if tool.IsWindows() {
		assetsHtml = filepath.Join(utils.RootPath(), "debug_most", "assets", "index.html")
		assetsHtml = "file://E:\\SWT\\gopath\\src\\github.com\\energye\\workspace\\examples\\cef\\debug_most\\assets\\index.html"
	}
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
	m.windowParent.SetAlign(types.AlClient)
	m.windowParent.SetAnchors(types.NewSet(types.AkTop, types.AkLeft, types.AkRight, types.AkBottom))
	// 创建一个定时器, 用来createBrowser
	m.timer = lcl.NewTimer(m)
	m.timer.SetEnabled(false)
	m.timer.SetInterval(500)
	m.timer.SetOnTimer(m.createBrowser)
	// 在show时创建chromium browser
	if tool.IsLinux() || tool.IsDarwin() {
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
		fmt.Println("OnLoadEnd")
		requestCtx := browser.GetHost().GetRequestContext()
		manager := requestCtx.GetCookieManager(nil)
		// 使用 chromium 事件
		manager.VisitAllCookies(cef.AsEngCookieVisitor(cef.NewCustomCookieVisitor(m.chromium, 0).AsIntfCookieVisitor()))
		// 使用 Eng 事件
		//manager.VisitAllCookies(cef.NewEngCookieVisitor())
		manager.Release()
	})
	m.chromium.SetOnAfterCreated(func(sender lcl.IObject, browser cef.ICefBrowser) {
		fmt.Println("SetOnAfterCreated isMainThread:", api.CurrentThreadId() == api.MainThreadId())
		m.timer.SetEnabled(true)
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
	m.chromium.SetOnBeforePopup(func(sender lcl.IObject, browser cef.ICefBrowser, frame cef.ICefFrame, targetUrl string, targetFrameName string, targetDisposition cefTypes.TCefWindowOpenDisposition, userGesture bool, popupFeatures cef.TCefPopupFeatures, windowInfo *cef.TCefWindowInfo, client *cef.IEngClient, settings *cef.TCefBrowserSettings, extraInfo *cef.ICefDictionaryValue, noJavascriptAccess *bool, result *bool) {
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
		headerMap := cef.NewStringMultimapOwn()
		intfHeaderMap := cef.AsCefStringMultimapOwn(headerMap.AsIntfStringMultimap())
		request.GetHeaderMap(intfHeaderMap)
		fmt.Println("headerMap size:", intfHeaderMap.GetSize())
		var key, val string
		for i := 0; i < int(intfHeaderMap.GetSize()); i++ {
			key = intfHeaderMap.GetKey(uint32(i))
			val = intfHeaderMap.GetValue(uint32(i))
			if key != "" {
				fmt.Println("  key:", key, "val:", val)
			}
		}
		intfHeaderMap.Release()
		fmt.Println("headerMap END")
	})
	m.chromium.SetOnProcessMessageReceived(func(sender lcl.IObject, browser cef.ICefBrowser, frame cef.ICefFrame, sourceProcess cefTypes.TCefProcessId,
		message cef.ICefProcessMessage, outResult *bool) {
		fmt.Println("主进程 name:", message.GetName())
		defer message.Release()
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
			binArgs.Release()
			args.Release()

			// 消息发送到渲染进程
			dataBytes := []byte("OK收到: " + string(messageDataBytes))
			processMessage := cef.ProcessMessageRef.New("send-render")
			messageArgumentList := processMessage.GetArgumentList()
			dataBin := cef.BinaryValueRef.New(uintptr(unsafe.Pointer(&dataBytes[0])), uint32(len(dataBytes)))
			messageArgumentList.SetBinary(0, dataBin)
			frame.SendProcessMessage(cefTypes.PID_BROWSER, processMessage)
			dataBin.Release()
			messageArgumentList.Clear()
			messageArgumentList.Release()
			processMessage.Release()
		}
	})
}

func (m *BrowserWindow) createBrowser(sender lcl.IObject) {
	if m.timer == nil {
		return
	}
	m.timer.SetEnabled(false)
	rect := m.ClientRect()
	created := m.chromium.CreateBrowserWithWindowHandleRectStringRequestContextDictionaryValueBool(m.windowParent.Handle(), rect, "", nil, nil, false)
	init := m.chromium.Initialized()
	fmt.Println("createBrowser rect:", rect, "init:", init, "create:", created)
	if !created && !init {
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
		lcl.RunOnMainThreadAsync(func(id uint32) {
			m.chromium.CloseBrowser(true)
		})
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
		if tool.IsDarwin() || tool.IsLinux() {
			m.Close()
		} else {
			rtl.PostMessage(m.Handle(), messages.WM_CLOSE, 0, 0)
		}
	}
}
