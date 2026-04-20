package main

import (
	"embed"
	"fmt"
	"github.com/energye/energy/v3/application"
	"github.com/energye/energy/v3/ipc"
	"github.com/energye/energy/v3/window"
	"github.com/energye/energy/v3/wv"
	"github.com/energye/lcl/api"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
	"runtime"
	"time"
)

//go:embed resources/*
var resources embed.FS

type TSystemInfoForm struct {
	window.TWindow
	Webview1 wv.IWebview
}

var SystemInfoForm TSystemInfoForm

var startTime = time.Now()

func main() {
	api.SetDebug(true)
	wvApp := wv.Init(nil, nil)

	wvApp.SetOptions(application.Options{
		DefaultURL: "sysinfo://main/index.html",
		Caption:    "系统信息查看器",
		Width:      1000,
		Height:     700,
	})

	wvApp.SetLocalLoad(application.LocalLoad{
		Scheme:     "sysinfo",
		Domain:     "main",
		ResRootDir: "resources",
		FS:         resources,
	})

	ipc.On("get-system-info", func(context ipc.IContext) {
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)

		info := map[string]interface{}{
			"os":          runtime.GOOS,
			"arch":        runtime.GOARCH,
			"cpuCores":    runtime.NumCPU(),
			"goVersion":   runtime.Version(),
			"goroutines":  runtime.NumGoroutine(),
			"allocMemory": formatBytes(memStats.Alloc),
			"totalMemory": formatBytes(memStats.TotalAlloc),
			"gcCount":     memStats.NumGC,
			"uptime":      time.Since(startTime).String(),
			"browserId":   context.BrowserId(),
		}
		context.Result(info)
	})

	ipc.On("refresh-stats", func(context ipc.IContext) {
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)

		stats := map[string]interface{}{
			"goroutines":  runtime.NumGoroutine(),
			"allocMemory": formatBytes(memStats.Alloc),
			"gcCount":     memStats.NumGC,
			"uptime":      time.Since(startTime).String(),
		}
		context.Result(stats)
	})

	wv.Run(&SystemInfoForm)
}

func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func (m *TSystemInfoForm) FormCreate(sender lcl.IObject) {
	m.InternalBeforeFormCreate()

	m.SetCaption("系统信息查看器")
	m.SetPosition(types.PoScreenCenter)
	m.SetWidth(1000)
	m.SetHeight(700)

	m.Webview1 = wv.NewWebview(m)
	m.Webview1.SetParent(m)
	m.Webview1.SetAlign(types.AlClient)
	m.Webview1.SetWindow(m)

	m.Webview1.SetOnLoadChange(func(url, title string, load wv.TLoadChange) {
		if load == wv.LcFinish {
			fmt.Println("系统信息页面加载完成")
		}
	})

	m.TWindow.FormCreate(sender)
}

func (m *TSystemInfoForm) OnShow(sender lcl.IObject) {
	m.WorkAreaCenter()
	m.Webview1.CreateBrowser()
}
