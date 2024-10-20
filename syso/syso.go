package syso

import (
	"go/build"
	"log"
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
		log.Panicln("os.Chdir:", err)
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
