// ==============================================================================
// 📚 应用启动入口文件
// 📌 该文件在创建项目时创建
// ✏️ 可在此文件中添加业务逻辑
// ==============================================================================

package main

import (
	"github.com/energye/energy/v3/application"
	"github.com/energye/energy/v3/wv"
	"github.com/energye/examples/wv/linux/simple_switch_parent/app"
	"github.com/energye/lcl/api"
	"github.com/energye/lcl/api/libname"
)

func main() {
	api.SetDebug(true)
	libname.LibName = "/home/yanghy/app/workspace/gen/gout/libenergy-gtk3.so"
	// 全局初始化
	wvApp := wv.Init()
	wvApp.SetOptions(application.Options{
		DefaultURL: "https://energye.gitee.io/",
		Linux:      application.Linux{HardwareGPU: application.HGPUDisable}, // VM WARE
		//Linux: application.Linux{HardwareGPU: application.HGPUEnable}, // GPU Device
	})
	// 启动应用程序消息循环
	wv.Run(app.Forms...)
}
