package application

import (
	"fmt"
	"github.com/energye/cef/cef"
	"github.com/energye/examples/cef/debug_most/scheme"
	"github.com/energye/examples/cef/debug_most/v8context"
	"github.com/energye/lcl/api/exception"
	"github.com/energye/lcl/tools"
	"os"
	"path/filepath"
)

func NewApplication() cef.ICefApplication {
	exception.SetOnException(func(funcName, message string) {
		fmt.Println("ERROR funcName:", funcName, "message:", message)
	})
	app := cef.NewCefApplication()
	app.SetEnableGPU(true)
	v8context.Context(app)
	cef.SetGlobalCEFApp(app)
	app.SetOnRegCustomSchemes(func(registrar cef.ICefSchemeRegistrarRef) {
		scheme.ApplicationOnRegCustomSchemes(registrar)
	})
	if !tools.IsDarwin() {
		// 非MacOS需要指定CEF框架目录，执行文件在CEF目录不需要设置
		// 指定 CEF Framework
		frameworkDir := os.Getenv("ENERGY_HOME")
		app.SetFrameworkDirPath(frameworkDir)
		app.SetResourcesDirPath(frameworkDir)
		app.SetLocalesDirPath(filepath.Join(frameworkDir, "locales"))
	}
	return app
}
