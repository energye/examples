package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/energye/cef/base"
	"github.com/energye/cef/cef"
	cefTypes "github.com/energye/cef/cef/types"
	"github.com/energye/energy/v3/platform/linux/gtk3"
	"github.com/energye/energy/v3/platform/linux/types"
	"github.com/energye/lcl/api/libname"
	"github.com/energye/lcl/lcl"
	lclTypes "github.com/energye/lcl/types"
)

var (
	globalApp cef.ICefApplication
	chromium  cef.IChromium
	mainWin   types.IWindow
	canClose  bool
	isClosing bool
)

func init() {
	libname.UseWS = "gtk3"
}

func main() {
	path := "/home/yanghy/.energy/chromium/linux_amd64_147.0.14"
	libname.LibName = filepath.Join(path, "libenergy-amd64-gtk3.so")

	lcl.Init()
	base.Init()
	gtk3.SetX11ErrorHandlers(nil, nil)

	globalApp = cef.NewApplication()
	base.SetGlobalCEFApplication(globalApp.Instance())
	globalApp.SetFrameworkDirPath(path)
	globalApp.SetResourcesDirPath(path)
	globalApp.SetLocalesDirPath(filepath.Join(path, "locales"))
	globalApp.SetMultiThreadedMessageLoop(false)
	globalApp.SetExternalMessagePump(false)

	wd, _ := os.Getwd()
	cacheRoot := filepath.Join(wd, "EnergyCache")
	globalApp.SetLogFile("debug.log")
	globalApp.SetLogSeverity(cefTypes.LOGSEVERITY_INFO)
	globalApp.SetRootCache(cacheRoot)
	globalApp.SetCache(filepath.Join(cacheRoot, "cache"))
	globalApp.SetDisableZygote(true)

	globalApp.SetOnContextInitialized(func() {
		fmt.Println("OnContextInitialized")
		chromium = cef.NewChromium(nil)
		chromium.SetDefaultUrl("https://www.baidu.com")
		chromium.SetOnAfterCreated(func(sender lcl.IObject, browser cef.ICefBrowser) {
			_ = sender
			fmt.Println("OnAfterCreated browserId:", browser.GetIdentifier())
			if chromium.Initialized() {
				chromium.UpdateXWindowVisibility(true)
				w, h := mainWin.GetSize()
				chromium.UpdateBrowserSize(0, 0, int32(w), int32(h))
			}
		})
		chromium.SetOnBeforeClose(func(sender lcl.IObject, browser cef.ICefBrowser) {
			_ = sender
			_ = browser
			fmt.Println("OnBeforeClose")
			canClose = true
		})
	})

	globalApp.SetOnGetDefaultClient(func(client *cef.IEngClient) {
		if chromium != nil {
			*client = chromium.CefClient()
		}
	})

	fmt.Println("StartMainProcess...")
	ok := globalApp.StartMainProcess()
	fmt.Println("StartMainProcess:", ok)
	if !ok {
		log.Fatal("StartMainProcess failed")
	}

	mainWin, _ = gtk3.NewWindow(types.WINDOW_TOPLEVEL)
	mainWin.SetTitle("GTKBrowser")
	mainWin.SetDefaultSize(1024, 768)

	mainWin.SetOnDestroy(func(sender types.PGtkWidget, userData types.GPointer) {
		if !canClose && !isClosing {
			isClosing = true
			fmt.Println("closing browser...")
			if chromium != nil {
				chromium.CloseBrowser(true)
			}
		}
		globalApp.QuitMessageLoop()
	})
	// configure-event → DoResize + NotifyMoveOrResizeStarted
	mainWin.SetOnConfigure(func(sender types.PGtkWidget, event types.PEventConfigure, userData types.GPointer) bool {
		_ = sender
		_ = event
		_ = userData
		if chromium != nil && chromium.Initialized() {
			w, h := mainWin.GetSize()
			chromium.UpdateBrowserSize(0, 0, int32(w), int32(h))
			chromium.NotifyMoveOrResizeStarted()
		}
		return false
	})

	fmt.Println("Show")
	gtk3.UseDefaultX11VisualForGtk(mainWin)
	mainWin.ShowAll()

	//    FChromium.CreateBrowser(TCefWindowHandle(FWindow), Rect(0,0,Width,Height))
	//    TCefWindowHandle on Linux = Pointer = GtkWidget*
	//    Lazarus CEF binding internally converts GtkWidget* → X11 XID
	if chromium != nil && !chromium.Initialized() {
		handle := cefTypes.TCefWindowHandle(mainWin.Instance())
		xid := gtk3.WindowX11ID(mainWin)
		fmt.Println("CreateBrowser handle:", fmt.Sprintf("0x%x  xid:0x%x", handle, xid))
		if !chromium.CreateBrowserWithWHandleRectStrRContextDValueBool(
			handle,
			lclTypes.TRect{},
			"",
			nil, nil, false,
		) {
			fmt.Println("CreateBrowser failed")
		}
	}
	gtk3.FlushDisplay(mainWin)

	fmt.Println("Run")
	globalApp.RunMessageLoop()
	fmt.Println("Application exit")
}
