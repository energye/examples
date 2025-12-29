package application

import (
	wv "github.com/energye/wv/windows"
)

var GlobalWVLoader wv.IWVLoader

func NewWVLoader() wv.IWVLoader {
	if GlobalWVLoader == nil {
		if GlobalWVLoader = wv.GetGlobalWebView2Loader(); GlobalWVLoader != nil {
			return GlobalWVLoader
		} else {
			GlobalWVLoader = wv.NewLoader(nil)
			wv.SetGlobalWebView2Loader(GlobalWVLoader)
		}
	}
	return GlobalWVLoader
}

func WVCachePath() string {
	return "E:\\app\\workspace\\examples\\wv\\windows\\WV2Cache"
}

func WV2LoaderDllPath() string {
	return "E:\\app\\workspace\\gen\\gout\\WebView2Loader.dll"
}
