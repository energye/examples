package wv2load

import (
	"github.com/energye/lcl/api/libname"
	"path"
)

func Wv2Load() (string, string) {
	wv2Path := libname.GetLibPath("WebView2Loader.dll")
	home, _ := path.Split(wv2Path)
	return home, wv2Path
}
