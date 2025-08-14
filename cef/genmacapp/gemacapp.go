package main

import "github.com/energye/lcl/pkgs/mac"

func main() {
	mainExe := "/Users/yanghy/app/workspace/examples/cef/debug_most/go_build_cef_debug_most_go"
	subExe := "/Users/yanghy/app/workspace/examples/cef/debug_most/helper/helper"
	mac.GenApp(mainExe, subExe)
}
