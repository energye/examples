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
	cacheRoot        = filepath.Join(wd, "ENERGY_WebKit2_Cache") // 浏览器缓存目录
	siteResourceRoot = filepath.Join(cacheRoot, "SiteResource")
)

func init() {
	TestLoadLibPath()
}

/*
Now requires GTK >= 3.24.24 and Glib2.0 >= 2.66
GTK3: dpkg -l | grep libgtk-3-0
Glib: dpkg -l | grep libglib2.0

Webkit2: >= 2.28

	sudo apt install -y libwebkit2gtk-4.0-37 libjavascriptcoregtk-4.0-18 libsoup2.4-1

	Web2: dpkg -l | grep webkit2
	Web2: pkg-config --modversion webkit2gtk-4.0

ldd --version

# 查找库文件位置
sudo find / -name "libwebkit2gtk-4.0.so"
sudo find / -name "libjavascriptcoregtk-4.0.so"
sudo find / -name "libsoup-2.4.so.1"
*/
func main() {
	window.CacheRoot = cacheRoot
	window.SiteResource = siteResourceRoot
	wv.Init(nil, nil)

	load := wv.NewLoader(nil)
	load.SetLoaderWebKit2DllPath("/usr/lib/x86_64-linux-gnu/libwebkit2gtk-4.0.so.37")
	load.SetLoaderJavascriptCoreDllPath("/usr/lib/x86_64-linux-gnu/libjavascriptcoregtk-4.0.so.18")
	load.SetLoaderSoupDllPath("/usr/lib/x86_64-linux-gnu/libsoup-2.4.so.1")
	if load.StartWebKit2() {
		lcl.Application.Initialize()
		lcl.Application.SetScaled(true)
		lcl.Application.NewForm(&window.Window)
		lcl.Application.Run()
	}
}
