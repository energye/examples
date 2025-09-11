package assets

import (
	"embed"
	"github.com/energye/lcl/tool"
	"os"
	"path/filepath"
)

//go:embed resources
var Assets embed.FS

var (
	wd, _ = os.Getwd()
)

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
	sourcePath = filepath.Join(wd, "wv", "assets", "resources", name)
	if tool.IsExist(sourcePath) {
		return sourcePath
	}
	return ""
}
