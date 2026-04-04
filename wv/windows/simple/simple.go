// ==============================================================================
// 📚 应用启动入口文件
// 📌 该文件在创建项目时创建
// ✏️ 可在此文件中添加业务逻辑
// ==============================================================================

package main

import (
	"github.com/energye/energy/v3/application"
	"github.com/energye/energy/v3/wv"
	"github.com/energye/examples/wv/windows/simple/app"
	_ "github.com/energye/examples/wv/windows/simple/app"
)

func main() {
	// 全局初始化
	wvApp := wv.Init(nil, nil)
	wvApp.SetOptions(application.Options{
		DefaultURL: "https://haokan.baidu.com/v?vid=16283970396480871799",
		Linux:      application.Linux{HardwareGPU: application.HGPUDisable}, // VM WARE
		//Linux: application.Linux{HardwareGPU: application.HGPUEnable}, // GPU Device
	})
	// 启动应用程序消息循环
	wv.Run(app.Forms...)
}
