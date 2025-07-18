package syso

import (
	"github.com/energye/lcl/api/libname"
	"github.com/energye/lcl/lcl"
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

func TestLoadLibPath() {
	libname.LibName = "E:\\SWT\\gopath\\src\\github.com\\energye\\workspace\\gen\\gout\\liblcl.dll"
}

// ScaleSelf : 这个方法主要是用于当不使用资源窗口创建时用，这个方法要用于设置了Width, Height或者ClientWidth、ClientHeight之后
func ScaleSelf(f lcl.IEngForm) {
	if lcl.Application.Scaled() {
		f.SetClientWidth(int32(float64(f.ClientWidth()) * (float64(lcl.Screen.PixelsPerInch()) / 96.0)))
		f.SetClientHeight(int32(float64(f.ClientHeight()) * (float64(lcl.Screen.PixelsPerInch()) / 96.0)))
	}
}
