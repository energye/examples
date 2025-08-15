package main

import (
	"fmt"
	"github.com/energye/cef/cef"
	cefTypes "github.com/energye/cef/types"
	"github.com/energye/examples/cef/application"
	"github.com/energye/lcl/config"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/tool"
	"github.com/energye/lcl/types"
	"os"
	"path/filepath"
)

var (
	chromium         cef.IChromium
	wd, _            = os.Getwd()
	cacheRoot        = filepath.Join(wd, "EnergyCache")         // 浏览器缓存目录
	siteResourceRoot = filepath.Join(cacheRoot, "SiteResource") // 网站资源缓存目录
)

func main() {
	cef.Init(nil, nil)
	app := application.NewApplication()
	app.SetFrameworkDirPath(config.Get().FrameworkPath())
	app.SetMultiThreadedMessageLoop(false)
	app.SetExternalMessagePump(false)
	app.SetDisablePopupBlocking(true)
	app.SetRootCache(cacheRoot)
	app.SetCache(filepath.Join(cacheRoot, "cache"))
	if tool.IsLinux() {
		app.SetDisableZygote(true)
	}
	app.SetOnContextInitialized(func() {
		fmt.Println("OnContextInitialized")
		fmt.Println("  GetScreenDPI:", cef.MiscFunc.GetScreenDPI(), "GetDeviceScaleFactor:", cef.MiscFunc.GetDeviceScaleFactor())
		var handle cefTypes.TCefWindowHandle
		cef.MiscFunc.InitializeWindowHandle(&handle)
		rect := types.TRect{}
		chromium = cef.NewChromium(nil)
		chromium.SetDefaultUrl("https://www.baidu.com")
		chromium.SetOnBeforeClose(func(sender lcl.IObject, browser cef.ICefBrowser) {
			app.QuitMessageLoop()
		})
		var tabURL string
		chromium.SetOnLoadStart(func(sender lcl.IObject, browser cef.ICefBrowser, frame cef.ICefFrame, transitionType cefTypes.TCefTransitionType) {
			fmt.Println("OnLoadStart", browser.GetIdentifier())
			if tabURL != "" {
				frame.LoadUrl(tabURL)
				tabURL = ""
			}
		})
		chromium.SetOnBeforePopup(func(sender lcl.IObject, browser cef.ICefBrowser, frame cef.ICefFrame, popupId int32, targetUrl string, targetFrameName string, targetDisposition cefTypes.TCefWindowOpenDisposition, userGesture bool, popupFeatures cef.TCefPopupFeatures, windowInfo *cef.TCefWindowInfo, client *cef.IEngClient, settings *cef.TCefBrowserSettings, extraInfo *cef.ICefDictionaryValue, noJavascriptAccess *bool, result *bool) {
			browser.GetHost().ExecuteChromeCommand(cefTypes.IDC_NEW_TAB, cefTypes.CEF_WOD_CURRENT_TAB)
			tabURL = targetUrl
			fmt.Println("OnBeforePopup", tabURL)
			*result = true
		})
		chromium.SetOnOpenUrlFromTab(func(sender lcl.IObject, browser cef.ICefBrowser, frame cef.ICefFrame, targetUrl string, targetDisposition cefTypes.TCefWindowOpenDisposition, userGesture bool, outResult *bool) {
			fmt.Println("OpenUrlFromTab", tabURL)
			*outResult = true
		})
		chromium.CreateBrowserWithWindowHandleRectStringRequestContextDictionaryValueBool(handle, rect, "tiny browser", nil, nil, true)
	})
	app.SetOnGetDefaultClient(func(client *cef.IEngClient) {
		fmt.Println("OnGetDefaultClient:", chromium)
		if chromium != nil {
			*client = chromium.CefClient()
		}
	})
	ok := app.StartMainProcess()
	fmt.Println("StartMainProcess", ok)
	if ok {
		app.RunMessageLoop()
	}
}
