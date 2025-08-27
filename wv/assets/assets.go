package assets

import (
	"embed"
	"github.com/energye/lcl/tool"
	"os"
	"path/filepath"
)

//go:embed resources
var Assets embed.FS
var wd, _ = os.Getwd()

func GetResourcePath(name string) string {
	var sourcePath string
	sourcePath = filepath.Join(wd, "resources", name)
	if tool.IsExist(sourcePath) {
		return sourcePath
	}
	sourcePath = filepath.Join("./", "resources", name)
	if tool.IsExist(sourcePath) {
		return sourcePath
	}
	if tool.IsWindows() {
		sourcePath = filepath.Join("E:\\SWT\\gopath\\src\\github.com\\energye\\workspace\\examples\\wv\\assets\\resources", name)
	} else if tool.IsLinux() {
		sourcePath = filepath.Join("/home/yanghy/app/gopath/src/github.com/energye/workspace/examples/cef/custombrowser/resources", name)
	} else if tool.IsDarwin() {
		sourcePath = filepath.Join("/Users/yanghy/app/workspace/examples/cef/custombrowser/resources", name)
	}
	if tool.IsExist(sourcePath) {
		return sourcePath
	}
	return ""
}
