package utils

import (
	"github.com/energye/lcl/tool/exec"
	"path/filepath"
)

func RootPath() string {
	return filepath.Join(exec.CurrentDir, "cef")
}
