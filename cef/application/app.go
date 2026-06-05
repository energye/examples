package application

import (
	"fmt"
	"github.com/energye/cef/cef"
	"github.com/energye/cef/config"
	"github.com/energye/lcl/api"
	"github.com/energye/lcl/tool"
	"github.com/energye/lcl/tool/exec"
	"path/filepath"
)

var (
	GlobalCEFApp cef.ICefApplication
)

func NewApplication() cef.ICefApplication {
	api.SetOnException(func(exceptionID, message string) {
		fmt.Println("[ERROR] exception:", exceptionID, "message:", message)
	})
	if GlobalCEFApp == nil {
		GlobalCEFApp = cef.NewApplication()
		cef.SetGlobalCEFApplication(GlobalCEFApp)
	}
	if !tool.IsDarwin() {
		// 非MacOS需要指定CEF框架目录，执行文件在CEF目录不需要设置
		// 默认 CEF Framework 目录
		cfg := config.GConfig
		if cfg != nil {
			libCef := func() string {
				if tool.IsWindows() {
					return "libcef.dll"
				} else if tool.IsLinux() {
					return "libcef.so"
				}
				return ""
			}()
			if libCef != "" {
				setOtherDirPath := func(v string) {
					GlobalCEFApp.SetFrameworkDirPath(v)
					GlobalCEFApp.SetResourcesDirPath(v)
					GlobalCEFApp.SetLocalesDirPath(filepath.Join(v, "locales"))
				}
				if tool.IsExist(filepath.Join(exec.Dir, libCef)) {
					setOtherDirPath(exec.Dir)
				} else if frameworkDir := cfg.ChromiumPath(); tool.IsExist(filepath.Join(frameworkDir, libCef)) {
					setOtherDirPath(frameworkDir)
				}
			}
		}
	}
	return GlobalCEFApp
}
