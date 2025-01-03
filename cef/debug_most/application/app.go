package application

import (
	"fmt"
	"github.com/energye/cef/cef"
	"github.com/energye/examples/cef/debug_most/scheme"
	"github.com/energye/examples/cef/debug_most/v8context"
	"github.com/energye/lcl/api/exception"
	"github.com/energye/lcl/inits/config"
	"github.com/energye/lcl/tools"
	"github.com/energye/lcl/tools/exec"
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
		// 默认 CEF Framework 目录
		cfg := config.Get()
		if cfg != nil {
			libCef := func() string {
				if tools.IsWindows() {
					return "libcef.dll"
				} else if tools.IsLinux() {
					return "libcef.so"
				}
				return ""
			}()
			if libCef != "" {
				setOtherDirPath := func(v string) {
					app.SetResourcesDirPath(v)
					app.SetLocalesDirPath(filepath.Join(v, "locales"))
				}
				if tools.IsExist(filepath.Join(exec.Dir, libCef)) {
					app.SetFrameworkDirPath(exec.Dir)
					setOtherDirPath(exec.Dir)
				} else if frameworkDir := cfg.FrameworkPath(); tools.IsExist(filepath.Join(frameworkDir, libCef)) {
					app.SetFrameworkDirPath(frameworkDir)
					setOtherDirPath(frameworkDir)
				}
			}
		}
	}
	return app
}
