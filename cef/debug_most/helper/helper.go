package main

import "github.com/energye/energy/v3/cef"

func main() {
	//全局配置初始化
	cef.GlobalInit(nil, nil)
	//创建Cef应用
	application := app.GetApplication()
	// 渲染进程监听IPC事件
	helperLisIPC()
	//启动子进程
	application.StartSubProcess()
	application.Free()
}
