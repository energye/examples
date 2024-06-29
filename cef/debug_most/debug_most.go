package main

import (
	"fmt"
	"github.com/energye/cef/cef"
	"github.com/energye/examples/cef/debug_most/contextmenu"
	"github.com/energye/examples/cef/debug_most/cookie"
	"github.com/energye/examples/cef/debug_most/devtools"
	"github.com/energye/examples/cef/debug_most/v8context"
	_ "github.com/energye/examples/syso"
	"github.com/energye/lcl/api"
	"github.com/energye/lcl/api/exception"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/process"
	"github.com/energye/lcl/rtl"
	"github.com/energye/lcl/tools"
	"github.com/energye/lcl/tools/exec"
	"github.com/energye/lcl/types"
	"github.com/energye/lcl/types/messages"
	"os"
	"path/filepath"
	"unsafe"
)

type BrowserWindow struct {
	lcl.TForm
	mainWindowId int32 // 主窗口ID
	timer        lcl.ITimer
	windowParent cef.ICEFWinControl
	chromium     cef.IChromium
	canClose     bool
	ChildForm    lcl.IForm
}

var (
	BW BrowserWindow
)

func main() {
	//全局初始化 每个应用都必须调用的
	cef.GlobalInit(nil, nil)
	exception.SetOnException(func(funcName, message string) {
		fmt.Println("ERROR funcName:", funcName, "message:", message)
	})
	app := cef.NewCefApplication()
	app.SetEnableGPU(true)
	v8context.Context(app)
	cef.SetGlobalCEFApp(app)
	if tools.IsDarwin() {
		app.SetUseMockKeyChain(true)
		app.InitLibLocationFromArgs()
		// MacOS
		cef.AddCrDelegate()
		cef.GlobalWorkSchedulerCreate(nil)
		app.SetOnScheduleMessagePumpWork(nil)
		app.SetExternalMessagePump(true)
		app.SetMultiThreadedMessageLoop(false)
		if !process.Args.IsMain() {
			// MacOS 多进程时，需要调用StartSubProcess来启动子进程
			subStart := app.StartSubProcess()
			fmt.Println("subStart:", subStart, process.Args.ProcessType())
			app.Free()
			return
		}
	} else { // MacOS不需要设置CEF框架目录，它是一个固定的目录结构
		// 非MacOS需要指定CEF框架目录，执行文件在CEF目录不需要设置
		// 指定 CEF Framework
		frameworkDir := os.Getenv("ENERGY_HOME")
		app.SetFrameworkDirPath(frameworkDir)
		app.SetResourcesDirPath(frameworkDir)
		app.SetLocalesDirPath(filepath.Join(frameworkDir, "locales"))
	}
	// 主进程启动
	mainStart := app.StartMainProcess()
	fmt.Println("mainStart:", mainStart, process.Args.ProcessType())
	if mainStart {
		// 结束应用后释放资源
		api.SetReleaseCallback(func() {
			fmt.Println("Release")
			app.Free()
		})
		// LCL窗口
		lcl.Application.Initialize()
		lcl.Application.SetMainFormOnTaskBar(true)
		lcl.Application.CreateForm(&BW)
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
	//m.chromium.SetDefaultUrl("https://www.baidu.com")
	assetsHtml := filepath.Join(exec.CurrentDir, "cef", "debug_most", "assets", "index.html")
	assetsHtml = "D:\\gopath\\src\\workspace\\examples\\cef\\debug_most\\assets\\index.html"
	m.chromium.SetDefaultUrl(assetsHtml)
	if tools.IsWindows() {
		m.windowParent = cef.NewCEFWindowParent(m)
	} else {
		windowParent := cef.NewCEFLinkedWindowParent(m)
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

	m.chromium.SetOnLoadingProgressChange(func(sender cef.IObject, browser cef.ICefBrowser, progress float64) {
		fmt.Println("OnLoadingProgressChange:", progress)
	})
	m.chromium.SetOnLoadStart(func(sender cef.IObject, browser cef.ICefBrowser, frame cef.ICefFrame, transitionType cef.TCefTransitionType) {
		fmt.Println("OnLoadStart:", frame.GetUrl())
	})
	m.chromium.SetOnLoadEnd(func(sender cef.IObject, browser cef.ICefBrowser, frame cef.ICefFrame, httpStatusCode int32) {
		requestCtx := browser.GetHost().GetRequestContext()
		manager := requestCtx.GetCookieManager(nil)
		manager.VisitAllCookies(cef.NewCefCustomCookieVisitor(m.chromium.AsInterface(), 0).AsInterface())
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
	})
	m.chromium.SetOnBeforeBrowse(func(sender lcl.IObject, browser cef.ICefBrowser, frame cef.ICefFrame, request cef.ICefRequest,
		userGesture, isRedirect bool, result *bool) {
		fmt.Println("SetOnBeforeBrowser")
		m.windowParent.UpdateSize()
	})
	m.chromium.SetOnDragEnter(func(sender cef.IObject, browser cef.ICefBrowser, dragData cef.ICefDragData, mask cef.TCefDragOperations, outResult *bool) {
		if mask&cef.DRAG_OPERATION_LINK == cef.DRAG_OPERATION_LINK {
			var fileNameList cef.IStrings
			fmt.Println("SetOnDragEnter", mask&cef.DRAG_OPERATION_LINK, dragData.IsLink(), dragData.IsFile(), "GetFileName:", dragData.GetFileName(),
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
	m.chromium.SetOnBeforePopup(func(sender cef.IObject, browser cef.ICefBrowser, frame cef.ICefFrame, beforePopup cef.TBeforePopup, popupFeatures cef.TCefPopupFeatures,
		windowInfo *cef.TCefWindowInfo, settings *cef.TCefBrowserSettings) (
		client cef.ICefClient, extraInfo cef.ICefDictionaryValue, noJavascriptAccess, result bool) {
		fmt.Printf("beforePopup: %+v\n", beforePopup)
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
		windowInfo.X = 400
		windowInfo.Y = 10
		windowInfo.Width = 400
		windowInfo.Height = 400
		windowInfo.WindowName = "杨杨红红岩岩"
		//result = true
		return
	})

	m.chromium.SetOnRenderCompMsg(func(sender lcl.IObject, message *types.TMessage, lResult *types.LRESULT, aHandled *bool) {
		//fmt.Println("SetOnRenderCompMsg", *lResult, *aHandled)
		//*aHandled = true
	})

	m.chromium.SetOnDownloadUpdated(func(sender cef.IObject, browser cef.ICefBrowser, downloadItem cef.ICefDownloadItem, callback cef.ICefDownloadItemCallback) {
		fmt.Println("DownloadUpdated frameId", browser.GetMainFrame().GetIdentifier(), "Id:", downloadItem.GetId(), "originalUrl:", downloadItem.GetOriginalUrl(), "url:", downloadItem.GetUrl())
		fmt.Println("\t", downloadItem.GetTotalBytes(), "/", downloadItem.GetReceivedBytes(), "speed:", downloadItem.GetCurrentSpeed(), "fullPath:", downloadItem.GetFullPath())
	})

	m.chromium.SetOnBeforeResourceLoad(func(sender lcl.IObject, browser cef.ICefBrowser, frame cef.ICefFrame, request cef.ICefRequest, callback cef.ICefCallback, result *cef.TCefReturnValue) {
		fmt.Println("SetOnBeforeResourceLoad")
		headerMap := request.GetHeaderMap()
		fmt.Println("headerMap size:", headerMap.GetSize())
		for i := 0; i < int(headerMap.GetSize()); i++ {
			_ = headerMap.GetKey(uint32(i))
			_ = headerMap.GetValue(uint32(i))
			//fmt.Println("  key:", key, "val:", val)
		}
		//callback.Cont()
	})
	m.chromium.SetOnProcessMessageReceived(func(browser cef.ICefBrowser, frame cef.ICefFrame, sourceProcess cef.TCefProcessId, message cef.ICefProcessMessage, outResult *bool) {
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
			frame.SendProcessMessage(cef.PID_BROWSER, processMessage)
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
	created := m.chromium.CreateBrowserByWindowHandle(m.windowParent.Handle(), &rect, "", nil, nil, false)
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

func (m *BrowserWindow) chromiumClose(sender lcl.IObject, browser cef.ICefBrowser, aAction *cef.TCefCloseBrowserAction) {
	fmt.Println("chromiumClose id:", browser.GetIdentifier(), "mainWindowId:", m.mainWindowId)
	if browser.GetIdentifier() == m.mainWindowId {
		if tools.IsDarwin() {
			m.windowParent.DestroyChildWindow()
			*aAction = cef.CbaClose
		} else {
			*aAction = cef.CbaDelay
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
		if tools.IsDarwin() {
			m.Close()
		} else {
			rtl.PostMessage(m.Handle(), messages.WM_CLOSE, 0, 0)
		}
	}
}
