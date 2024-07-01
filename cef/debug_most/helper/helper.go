package main

import (
	"github.com/energye/cef/cef"
	"github.com/energye/examples/cef/debug_most/application"
)

func main() {
	//全局配置初始化
	cef.Init(nil, nil)
	//创建Cef应用
	app := application.NewApplication()
	//启动子进程
	app.StartSubProcess()
	app.Free()
}
