package main

import (
	. "github.com/energye/examples/syso"
	"github.com/energye/examples/wv/linux/custombrowser/window"
	"github.com/energye/lcl/lcl"
	wv "github.com/energye/wv/linux"
	"os"
	"path/filepath"
)

var (
	wd, _            = os.Getwd()
	cacheRoot        = filepath.Join(wd, "ENERGY_WebView2_Cache") // 浏览器缓存目录
	siteResourceRoot = filepath.Join(cacheRoot, "SiteResource")
)

func init() {
	TestLoadLibPath()
}

/*
Now requires GTK >= 3.24.24 and Glib2.0 >= 2.66
GTK3: dpkg -l | grep libgtk-3-0
Glib: dpkg -l | grep libglib2.0
Web2: dpkg -l | grep webkit2
ldd --version
*/
func main() {
	window.CacheRoot = cacheRoot
	window.SiteResource = siteResourceRoot
	wv.Init(nil, nil)
	lcl.Application.Initialize()
	lcl.Application.SetScaled(true)
	lcl.Application.NewForm(&window.Window)
	lcl.Application.Run()
}
