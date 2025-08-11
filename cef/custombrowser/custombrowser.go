package main

import (
	"embed"
	"fmt"
	"github.com/energye/assetserve"
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

//go:embed resources
var resources embed.FS

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
	fmt.Println("ProcessType:", app.ProcessType())
	app.SetWindowlessRenderingEnabled(true)
	app.SetEnableGPU(true)
	app.SetLocale("zh-CN")

	if tool.IsWindows() {
		// win32 使用 lcl 窗口
		app.SetExternalMessagePump(false)
		app.SetMultiThreadedMessageLoop(true)
		app.SetRootCache(cacheRoot)
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
	}

	app.SetOnAlreadyRunningAppRelaunch(func(commandLine cef.ICefCommandLine, currentDirectory string, result *bool) {
		*result = true
	})
	// 主进程启动
	mainStart := app.StartMainProcess()
	if mainStart {
		httpServer()
		CEFINfo(app)
		// 结束应用后释放资源
		api.SetReleaseCallback(func() {
			fmt.Println("Run END. Release")
		})
		// LCL窗口
		lcl.Application.Initialize()
		lcl.Application.SetMainFormOnTaskBar(true)
		lcl.Application.NewForms(&window.BW)
		//lcl.Application.NewForms(&window.BW, &window.CW)
		lcl.Application.Run()
	}
}

func CEFINfo(app cef.ICefApplication) {
	// 输出版本信息
	var (
		chromiumVersion = &cef.TChromiumVersionInfo{}
		cefVersion      = &cef.TCefVersionInfo{}
	)
	app.GetChromiumVersionInfo(chromiumVersion)
	app.GetCEFVersionInfo(cefVersion)
	fmt.Println("ChromeVersion:", app.ChromeVersion())
	fmt.Println("ChromiumVersionInfo:", fmt.Sprintf("\n  Major: %v\n  Minor: %v\n  Build: %v\n  Patch: %v",
		chromiumVersion.VersionMajor,
		chromiumVersion.VersionMinor,
		chromiumVersion.VersionBuild,
		chromiumVersion.VersionPatch))
	fmt.Println("CefVersion:", app.LibCefVersion())
	fmt.Println("CefVersionInfo:", fmt.Sprintf("\n  Major: %v\n  Minor: %v\n  Build: %v\n  CommitNumber: %v",
		cefVersion.VersionMajor,
		cefVersion.VersionMinor,
		cefVersion.VersionPatch,
		cefVersion.CommitNumber))
}

func httpServer() {
	server := assetserve.NewAssetsHttpServer()
	server.PORT = 22022
	server.AssetsFSName = "resources" //必须设置目录名
	server.Assets = resources
	go server.StartHttpServer()
}
