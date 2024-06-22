//go:build windows
// +build windows

package main

import (
	_ "github.com/energye/examples/syso"
	"github.com/energye/lcl/inits"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/pkgs/win"
	"github.com/energye/lcl/rtl"
	"os"
)

func main() {
	inits.Init(nil, nil)
	rtl.CreateURLShortCut(win.GetDesktopPath(), "energy", "https://github.com/energye/energy")
	rtl.CreateShortCut(win.GetDesktopPath(), "shortcut", os.Args[0], "", "描述", "-b -c")
	lcl.ShowMessage("Hello!")
}
