package main

import (
	"github.com/energye/energy/v3/cef"
)

func main() {
	//全局初始化 每个应用都必须调用的
	cef.Init(nil, nil)
	//创建应用
	//app := cef.NewApplication()
	//指定一个URL地址，或本地html文件目录
	//cef.BrowserWindow.Config.Url = "https://www.baidu.com"
	//运行应用
	//cef.Run(app)
}
