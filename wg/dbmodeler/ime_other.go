//go:build !windows

package main

// 非 Windows 平台的编译占位：GTK/Cocoa 上应换成 GtkIMContext /
// NSTextInputClient，本示例的 IME 实测链路是 Win32 专有的。
func imeComposition(hwnd uintptr) (string, int32, bool) { return "", 0, false }
func imeResult(hwnd uintptr) string                     { return "" }
func imePosition(hwnd uintptr, x, y int32) string       { return "SKIP: Win32 only" }
func imeSelfTest() string                               { return "SKIP: Win32 only" }
func libSelfTest(libPath string) string                 { return "SKIP: Win32 only" }
