package main

import (
	"fmt"
	"github.com/energye/cef/cef"
	"github.com/energye/examples/cef/debug_most/helper/internal"
	"github.com/energye/lcl/api/exception"
)

func main() {
	cef.Init(nil, nil)
	exception.SetOnException(func(exception int32, message string) {
		fmt.Println("[ERROR] exception:", exception, "message:", message)
	})
	//创建Cef应用
	app := internal.InitApplication()
	ok := app.StartSubProcess()
	println("sub process start:", ok)
}
