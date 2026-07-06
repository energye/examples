package main

import (
	"embed"
	"fmt"
	cef2 "github.com/energye/cef/cef"
	"github.com/energye/energy/v3/application"
	"github.com/energye/energy/v3/cef"
	"github.com/energye/energy/v3/core"
	"github.com/energye/energy/v3/ipc"
	"github.com/energye/energy/v3/logger"
	"github.com/energye/examples/cef/eng_simple_vf/app"
	"github.com/energye/examples/wv/assets"
	"github.com/energye/lcl/api"
	"github.com/energye/lcl/api/libname"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
	"strconv"
	"time"
)

type TBrowserWindow struct {
	cef.TViewsBrowser
}

var MainBrowserWindow TBrowserWindow

//go:embed resources
var resources embed.FS

func init() {
	libname.UseWS = "gtk3"
}

func main() {
	logger.L().SetLevel(logger.DebugLevel)
	cefApp := cef.Init()
	//cefApp.SetLogSeverity(cefTypes.LOGSEVERITY_DEBUG)
	cefApp.SetOptions(application.Options{
		//Frameless: true,
		//WindowTransparent: true,
		//WebviewTransparent: true,
		//BackgroundColor:    colors.NewARGB(0, 0, 0, 0),
		//DefaultURL:"https://energye.gitee.io",
		DefaultURL: "fs://energy/index-home.html",
		//DefaultURL: "fs://energy/index-ipc.html",
		//DefaultURL: "fs://energy/index-drag.html",
		//DefaultURL:      "http://chrome.360.cn/html5_labs",
		//DefaultURL:      "https://www.baidu.com",
		AutoPopupWindow: true,
		Width:           800,
		Height:          600,
		//DisableResize:   true,
	})
	cefApp.SetOnBeforeChildProcessLaunch(func(commandLine cef2.ICefCommandLine) {
		println("app.SetOnBeforeChildProcessLaunch")
	})
	cefApp.SetLocalLoad(application.LocalLoad{
		Scheme:     "fs",
		Domain:     "energy",
		ResRootDir: "resources",
		FS:         resources,
	})

	ipc.BindEvent(&app.DemoBind{})
	ipc.BindEventPrefix("demo", &app.DemoBind{})
	ipc.On("test", func(context ipc.IContext) {
		fmt.Println("ipc-test:", context.BrowserId(), "data:", context.Data())
		context.Result("ResultData", 123, 888.99, true, time.Now().String())
		ipc.Emit("test", "测试数据")
	})

	getWindow := func(browserId uint32) cef.IViewsBrowser {
		return cefApp.GetWindow(browserId).(cef.IViewsBrowser)
	}
	ipc.On("minimize", func(context ipc.IContext) {
		fmt.Println("minimize", context.BrowserId())
		tempWindow := getWindow(context.BrowserId())
		if tempWindow != nil {
			tempWindow.Minimize()
		}
	})

	ipc.On("maximize", func(context ipc.IContext) {
		fmt.Println("maximize", context.BrowserId())
		tempWindow := getWindow(context.BrowserId())
		if tempWindow != nil {
			if tempWindow.IsMaximize() {
				tempWindow.Restore()
			} else {
				tempWindow.Maximize()
			}
		}
	})

	ipc.On("fullscreen", func(context ipc.IContext) {
		fmt.Println("fullscreen", context.BrowserId())
		tempWindow := getWindow(context.BrowserId())
		if tempWindow != nil {
			if tempWindow.IsFullScreen() {
				tempWindow.ExitFullScreen()
			} else {
				tempWindow.FullScreen()
			}
		}
	})

	ipc.On("close", func(context ipc.IContext) {
		fmt.Println("close", context.BrowserId())
		tempWindow := getWindow(context.BrowserId())
		if tempWindow != nil {
			lcl.RunOnMainThreadAsync(func(id uint32) {
				tempWindow.Close()
			})
		}
	})
	cefApp.SetEnableGPU(true)
	//cefApp.SetBrowserSubprocessPath("/home/yanghy/app/workspace/examples/cef/eng_simple_vf/helper/helper")
	cef.Run(&MainBrowserWindow)
}

func (m *TBrowserWindow) OnFormCreate(sender lcl.IObject) {
	println("OnFormCreate", api.CurrentThreadId(), api.MainThreadId())
	//m.SetIsAlwaysOnTop(true)
	m.SetTitle("OnFormCreate")
	pngData, _ := resources.ReadFile("resources/icon.png")
	m.SetIcon(pngData)
	m.SetOnResourceRequest(func(url, path, method string, header map[string]string) (resource string, ok bool) {
		fmt.Println("Browser.SetOnResourceRequest:", url, path, method, header)
		return
	})

	m.SetOnPopupWindow(func(targetURL string) bool {
		fmt.Println("Browser.SetOnPopupWindow targetURL:", targetURL)
		return false
	})

	m.SetOnLoadChange(func(url, title string, load core.TLoadChange) {
		fmt.Println("Browser.SetOnLoadChange:", url, title, load)
	})

	m.SetOnDragEnter(func(type_ core.TDragType, x, y int32) {
		fmt.Println("SetOnDragEnter --------------begin------------------", type_, x, y)
		ipc.Emit("drag-enter")
	})
	m.SetOnDragLeave(func() {
		fmt.Println("SetOnDragLeave", "--------------zzz------------------")
	})
	m.SetOnDragOver(func(data *core.TDragData, x, y int32) {
		da, err := strconv.Unquote("\"" + string(data.Data) + "\"")
		fmt.Println("SetOnDragOver --------------end------------------", x, y, da, err, data.Filenames)
		ipc.Emit("drag-over", da, data.Filenames)
	})
	m.SetOnContextMenu(func(contextMenu *core.TContextMenuItem) {
		contextMenu.Clear()
		contextMenu.Add("", core.CmkSeparator)
		_, id := contextMenu.Add("测试1", core.CmkCommand)
		fmt.Println("测试1:", id)
		test2, id := contextMenu.Add("测试2", core.CmkSub)
		fmt.Println("测试2:", id)
		_, id = test2.Add("测试2-测试", core.CmkCommand)
		fmt.Println("测试2-测试:", id)
		_, id = test2.Add("测试3-测试", core.CmkCommand)
		fmt.Println("测试3-测试:", id)
		contextMenu.Add("测试3", core.CmkCommand)
	})
	m.SetOnContextMenuCommand(func(commandId int32, handle *bool) {
		fmt.Println("OnCont extMenuCommand:", commandId)
		m.ExecuteScriptCallback("document.title", func(result string, err string) {
			fmt.Println("ExecuteScriptCallback:", result, err)
		})
	})
	m.SetOnThemeChange(func(isDark bool) {
		fmt.Println("SetOnThemeChange isDark:", isDark)
	})
	lcl.RunOnMainThreadAsync(func(id uint32) {
		fmt.Println("RunOnMainThreadAsync")
	})
	tray := application.NewTrayIcon()
	trayMenu := tray.Menu()
	trayMenu.SetImageListEmbed(assets.Assets, []string{"resources/window-icon_64x64.png"})

	exit := trayMenu.AddMenuItem("退出").SetOnClick(func() {
		println("退出")
		m.Close()
	})
	//exit.SetImage("window-icon_64x64.png")
	testdata, _ := assets.Assets.ReadFile("resources/window-icon_64x64.png")
	exit.SetBitmap(testdata)

	trayMenu.AddSeparator()
	//trayMenu.SetImageList([]string{"E:\\app\\workspace\\examples\\wv\\assets\\resources\\add.png"})
	testMenu := trayMenu.AddMenuItem("test")
	testMenu.SetOnMeasureItem(func(sender lcl.IObject, canvas lcl.ICanvas, width *int32, height *int32) {
		*height = 32
	})
	test2Menu := testMenu.AddSubMenuItem("test2")
	test2Menu.SetChecked(true)
	testMenu.AddSeparator()
	test2Menu = testMenu.AddSubMenuItem("test2222")
	test2Menu.SetRadio(true)
	test2Menu = testMenu.AddSubMenuItem("test3333")
	test2Menu.SetRadio(true)
	test2Menu.SetChecked(true)

	//tray.SetIcon("E:\\app\\workspace\\examples\\wv\\assets\\resources\\add.png")
	trayIconData, _ := assets.Assets.ReadFile("resources/add.png")
	tray.SetIconBytes(trayIconData)
	tray.SetOnMouseUp(func(button types.TMouseButton, shift types.TShiftState, X int32, Y int32) {
		fmt.Println("SetOnMouseUp")
	})
	tray.SetOnClick(func() {
		fmt.Println("SetOnClick")
	})
	tray.Show()
}

func (m *TBrowserWindow) OnFormShow(sender lcl.IObject) {
	m.CenterWindow()
	println("OnFormShow")
}

func (m *TBrowserWindow) OnFormCloseQuery(sender lcl.IObject, canClose *bool) bool {
	println("OnFormCloseQuery")
	//*canClose = false
	return false
}

func (m *TBrowserWindow) OnFormClose(sender lcl.IObject, closeAction *types.TCloseAction) bool {
	println("OnFormClose")
	//*closeAction = types.CaMinimize
	return false
}
