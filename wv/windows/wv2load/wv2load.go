package wv2load

import (
	"github.com/energye/lcl/api/libname"
	"path/filepath"
)

func Wv2Load() (string, string) {
	wv2Path, _ := filepath.Split(libname.LibName)
	return wv2Path, filepath.Join(wv2Path, "WebView2Loader.dll")
}
