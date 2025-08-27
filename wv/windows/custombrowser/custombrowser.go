package main

import (
	"fmt"
	. "github.com/energye/examples/syso"
	"github.com/energye/examples/wv/windows/application"
	"github.com/energye/examples/wv/windows/custombrowser/window"
	"github.com/energye/lcl/api/exception"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/tool/exec"
	wv "github.com/energye/wv/windows"
	"path/filepath"
)

var (
	load wv.IWVLoader
)

func init() {
	TestLoadLibPath()
}

func main() {
	fmt.Println("Go ENERGY Run Main")
	wv.Init(nil, nil)
	exception.SetOnException(func(exception int32, message string) {
		fmt.Println("ERROR exception:", exception, "message:", message)
	})
	// GlobalWebView2Loader
	load = application.NewWVLoader()
	fmt.Println("当前目录:", exec.CurrentDir)
	fmt.Println("WebView2Loader.dll目录:", application.WV2LoaderDllPath())
	fmt.Println("用户缓存目录:", filepath.Join(application.WVCachePath(), "webview2Cache"))
	load.SetUserDataFolder(application.WVCachePath())
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
