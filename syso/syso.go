package syso

import (
	"bytes"
	"github.com/energye/lcl/api/libname"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/tool"
	"go/build"
	"os"
	"path/filepath"
	"strings"
)

// Set the working directory to the root of Go package, so that its assets can be accessed.
func Chdir(module string) {
	wd, err := os.Getwd()
	wd = filepath.ToSlash(wd)
	wd = strings.Replace(wd, module, "", 1)
	err = os.Chdir(filepath.Join(wd, module))
	if err != nil {
		println("os.Chdir:", err)
	}
}

// importPathToDir resolves the absolute path from importPath.
// There doesn't need to be a valid Go package inside that import path,
// but the directory must exist.
func importPathToDir(importPath string) (string, error) {
	p, err := build.Import(importPath, "", build.FindOnly)
	if err != nil {
		return "", err
	}
	return p.Dir, nil
}

var (
	wd, _ = os.Getwd()
)

func TestLoadLibPath() {
	var name string
	if tool.IsWindows() {
		name = "liblcl.dll"
	} else if tool.IsLinux() {
		name = "liblcl.so"
	}
	if name != "" {
		// 当前目录
		liblcl := filepath.Join(wd, name)
		if tool.IsExist(liblcl) {
			libname.LibName = liblcl
			return
		}
		// 测试编译输出目录
		if tool.IsWindows() {
			liblcl = filepath.Join("E:\\SWT\\gopath\\src\\github.com\\energye\\workspace\\gen\\gout", name)
		} else if tool.IsLinux() {
			liblcl = filepath.Join("/home/yanghy/app/gopath/src/github.com/energye/workspace/gen/gout", name)
		} else if tool.IsDarwin() {
			liblcl = filepath.Join("/Users/yanghy/app/workspace/gen/gout", name)
		}
		if tool.IsExist(liblcl) {
			libname.LibName = liblcl
			return
		}
	}
}

// ScaleSelf : 这个方法主要是用于当不使用资源窗口创建时用，这个方法要用于设置了Width, Height或者ClientWidth、ClientHeight之后
func ScaleSelf(f lcl.IEngForm) {
	if lcl.Application.Scaled() {
		f.SetClientWidth(int32(float64(f.ClientWidth()) * (float64(lcl.Screen.PixelsPerInch()) / 96.0)))
		f.SetClientHeight(int32(float64(f.ClientHeight()) * (float64(lcl.Screen.PixelsPerInch()) / 96.0)))
	}
}

var buf = bytes.Buffer{}

func Concat(vs ...string) (r string) {
	for _, v := range vs {
		buf.WriteString(v)
	}
	r = buf.String()
	buf.Reset()
	return
}
