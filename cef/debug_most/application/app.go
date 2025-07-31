package application

import (
	"fmt"
	"github.com/energye/cef/cef"
	"github.com/energye/examples/cef/debug_most/scheme"
	"github.com/energye/examples/cef/debug_most/v8context"
	"github.com/energye/lcl/api/exception"
	"github.com/energye/lcl/tool"
	"github.com/energye/lcl/tool/exec"
	"path/filepath"
)

var (
	GlobalCEFApp cef.ICefApplication
)

func NewApplication() cef.ICefApplication {
	exception.SetOnException(func(idType int32, message string) {
		fmt.Println("ERROR method id:", idType, "message:", message)
	})
	if GlobalCEFApp == nil {
		GlobalCEFApp = cef.NewApplication()
		cef.SetGlobalCEFApplication(GlobalCEFApp)
	}
	GlobalCEFApp.SetEnableGPU(true)
	v8context.Context(GlobalCEFApp)
	GlobalCEFApp.SetOnRegCustomSchemes(func(registrar cef.ICefSchemeRegistrarRef) {
		scheme.ApplicationOnRegCustomSchemes(registrar)
	})
	if !tool.IsDarwin() {
		// 非MacOS需要指定CEF框架目录，执行文件在CEF目录不需要设置
		// 指定 CEF Framework
		// 默认 CEF Framework 目录
		cfg := Get() // config.Get()
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
					GlobalCEFApp.SetResourcesDirPath(v)
					GlobalCEFApp.SetLocalesDirPath(filepath.Join(v, "locales"))
				}
				if tool.IsExist(filepath.Join(exec.Dir, libCef)) {
					GlobalCEFApp.SetFrameworkDirPath(exec.Dir)
					setOtherDirPath(exec.Dir)
				} else if frameworkDir := cfg.FrameworkPath(); tool.IsExist(filepath.Join(frameworkDir, libCef)) {
					GlobalCEFApp.SetFrameworkDirPath(frameworkDir)
					setOtherDirPath(frameworkDir)
				}
			}
		}
	}
	return GlobalCEFApp
}

type Config struct {
}

// Get 返回 config 环境
// mode: 构建模式 dev 或 prod
// 当 mode 是 dev 时即开发环境 使用 $HOME/.energy配置, 是 prod 时即生产模式不再使用 $HOME/.energy配置, 使用自定义或当前执行目录
func Get() *Config {
	return &Config{}
}

func (m *Config) FrameworkPath() string {
	if tool.IsWindows() {
		return "E:\\app\\energy\\CEF-136_WINDOWS_64"
	} else if tool.IsLinux() {
		return "/home/yanghy/app/energy/CEF-136_LINUX_64"
	}
	return ""
}
