package main

import (
	"fmt"
	"github.com/energye/cef/cef"
	"github.com/energye/examples/cef/application"
	"github.com/energye/examples/cef/custombrowser/window"
	. "github.com/energye/examples/syso"
	"github.com/energye/lcl/api"
	"github.com/energye/lcl/api/exception"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/tool"
	"os"
	"path/filepath"
)

func init() {
	TestLoadLibPath()
}

var (
	wd, _            = os.Getwd()
	cacheRoot        = filepath.Join(wd, "EnergyCache")         // 浏览器缓存目录
	siteResourceRoot = filepath.Join(cacheRoot, "SiteResource") // 网站资源缓存目录
)

func main() {
	window.CacheRoot = cacheRoot
	window.SiteResource = siteResourceRoot
	//全局初始化 每个应用都必须调用的
	cef.Init(nil, nil)
	exception.SetOnException(func(exception int32, message string) {
		fmt.Println("[ERROR] exception:", exception, "message:", message)
	})
	app := application.NewApplication()
	app.SetEnableGPU(true)
	app.SetLocale("zh-CN")
	if tool.IsWindows() {
		// win32 使用 lcl 窗口
		app.SetExternalMessagePump(false)
		app.SetMultiThreadedMessageLoop(true)
		app.SetRootCache(cacheRoot)
	}
	//	app.SetOnWebKitInitialized(func() {
	//		var myParamValue string
	//		v8Handler := cef.NewEngV8Handler()
	//		v8Handler.SetOnV8Execute(func(name string, object cef.ICefv8Value, arguments cef.ICefv8ValueArray, retval *cef.ICefv8Value, exception *string) bool {
	//			fmt.Println("v8Handler.Execute", name)
	//			var result bool
	//			if name == "GetMyParam" {
	//				result = true
	//				myParamValue = myParamValue + " " + time.Now().String()
	//				*retval = cef.V8ValueRef.NewString(myParamValue)
	//			} else if name == "SetMyParam" {
	//				if arguments.Count() > 0 {
	//					newValue := arguments.Get(0)
	//					fmt.Println("value is string:", newValue.IsString())
	//					fmt.Println("value:", newValue.GetStringValue())
	//					myParamValue = newValue.GetStringValue()
	//					newValue.Free()
	//				}
	//				result = true
	//			}
	//			return result
	//		})
	//		var jsCode = `
	//            let test;
	//            if (!test) {
	//                test = {};
	//            }
	//            (function () {
	//                test.__defineGetter__('myparam', function () {
	//                    native function GetMyParam();
	//                    return GetMyParam();
	//                });
	//                test.__defineSetter__('myparam', function (b) {
	//                    native function SetMyParam();
	//					b = b + ' TEST';
	//                    if (b) SetMyParam(b);
	//                });
	//            })();
	//`
	//		cef.MiscFunc.CefRegisterExtension("v8/test", jsCode, cef.AsEngV8Handler(v8Handler.AsIntfV8Handler()))
	//	})

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
		api.WidgetSetInitialization()
		// LCL窗口
		lcl.Application.Initialize()
		lcl.Application.SetMainFormOnTaskBar(true)
		lcl.Application.NewForm(&window.BW)
		lcl.Application.Run()
	}
}
