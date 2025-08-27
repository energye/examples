package main

import (
	"fmt"
	. "github.com/energye/examples/syso"
	"github.com/energye/examples/wv/windows/application"
	"github.com/energye/examples/wv/windows/custombrowser/window"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/tool/exec"
	wv "github.com/energye/wv/windows"
	"os"
	"path/filepath"
)

var (
	load             wv.IWVLoader
	wd, _            = os.Getwd()
	cacheRoot        = filepath.Join(wd, "ENERGY_WebView2_Cache") // 浏览器缓存目录
	siteResourceRoot = filepath.Join(cacheRoot, "SiteResource")   // 网站资源缓存目录
)

func init() {
	TestLoadLibPath()
}

func main() {
	window.CacheRoot = cacheRoot
	window.SiteResource = siteResourceRoot
	fmt.Println("Go ENERGY Run Main")
	wv.Init(nil, nil)
	// GlobalWebView2Loader
	load = application.NewWVLoader()
	fmt.Println("当前目录:", exec.CurrentDir)
	fmt.Println("WebView2Loader.dll目录:", application.WV2LoaderDllPath())
	fmt.Println("用户缓存目录:", filepath.Join(cacheRoot, "webview2Cache"))
	load.SetUserDataFolder(cacheRoot)
	load.SetLoaderDllPath(application.WV2LoaderDllPath())
	r := load.StartWebView2()
	fmt.Println("StartWebView2", r)
	window.Load = load
	lcl.Application.Initialize()
	lcl.Application.SetMainFormOnTaskBar(true)
	lcl.Application.NewForm(&window.Window)
	lcl.Application.Run()
	wv.DestroyGlobalWebView2Loader()
}
