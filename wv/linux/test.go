package main

import (
	"fmt"
	"github.com/energye/examples/wv/tool"
	"github.com/energye/lcl/api/imports"
)

func main() {
	dll, err := imports.NewDLL("libwebkit2gtk-4.0.so")
	fmt.Println(dll, err)
	dll, err = imports.NewDLL("libjavascriptcoregtk-4.0.so")
	fmt.Println(dll, err)
	dll, err = imports.NewDLL("libsoup-2.4.so.1")
	fmt.Println(dll, err)
	println(tool.FindLib("libwebkit2gtk-4.0.so"))
	println(tool.FindLib("libjavascriptcoregtk-4.0.so"))
	println(tool.FindLib("libsoup-2.4.so"))
}
