package application

import (
	wv "github.com/energye/wv/windows"
)

var GlobalWVLoader wv.IWVLoader

func NewWVLoader() wv.IWVLoader {
	if GlobalWVLoader == nil {
		GlobalWVLoader = wv.NewLoader(nil)
		wv.SetGlobalWebView2Loader(GlobalWVLoader)
		GlobalWVLoader.SetUserDataFolder(WVCachePath())
		GlobalWVLoader.SetLoaderDllPath(WV2LoaderDllPath())
	}
	return GlobalWVLoader
}

func WVCachePath() string {
	return "E:\\SWT\\gopath\\src\\github.com\\energye\\workspace\\examples\\wv\\windows\\WV2Cache"
}

func WV2LoaderDllPath() string {
	return "E:\\SWT\\gopath\\src\\github.com\\energye\\workspace\\gen\\gout\\WebView2Loader.dll"
}
