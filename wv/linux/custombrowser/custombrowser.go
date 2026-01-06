package main

import (
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

/*
   Now requires GTK >= 3.24.24 and Glib2.0 >= 2.66
   GTK3: dpkg -l | grep libgtk-3-0
   Glib: dpkg -l | grep libglib2.0

   Webkit2: >= 2.28

		安装指定版本，需要指定软链接, 或制定 dlopen 小版本号的库
		sudo apt-get install -y libwebkit2gtk-4.0-37 libjavascriptcoregtk-4.0-18 libsoup2.4-1
		安装完整版本. 带有完整的开发包等等
   	    sudo apt-get install libwebkit2gtk-4.0
   	    sudo apt-get install libjavascriptcoregtk-4.0
   	    sudo apt-get install libsoup2.4

   		Web2: dpkg -l | grep webkit2
   		Web2: pkg-config --modversion webkit2gtk-4.0

   ldd --version

   查找库文件位置
   sudo find / -name "libwebkit2gtk-4.0.so"
   sudo find / -name "libjavascriptcoregtk-4.0.so"
   sudo find / -name "libsoup-2.4.so.1"

   缺少 GStreamer FDK AAC 插件，这会导致 AAC 格式的音频播放可能无法正常工作
   sudo apt-get update
   sudo apt-get install gstreamer1.0-libav gstreamer1.0-plugins-bad gstreamer1.0-plugins-ugly gstreamer1.0-fdk
*/

func main() {
	window.CacheRoot = cacheRoot
	window.SiteResource = siteResourceRoot
	os.Setenv("--ws", "gtk3")
	lcl.Init(nil, nil)
	wv.Init()

	load := wv.NewLoader(nil)
	// 直接通过指定小版本号： sudo apt-get install -y libwebkit2gtk-4.0-37 libjavascriptcoregtk-4.0-18 libsoup2.4-1
	load.SetLoaderWebKit2DllPath("libwebkit2gtk-4.0.so.37")
	load.SetLoaderJavascriptCoreDllPath("libjavascriptcoregtk-4.0.so.18")
	load.SetLoaderSoupDllPath("libsoup-2.4.so.1")
	if load.StartWebKit2() {
		lcl.Application.Initialize()
		lcl.Application.SetScaled(true)
		lcl.Application.NewForm(&window.Window)
		lcl.Application.Run()
	}
}
