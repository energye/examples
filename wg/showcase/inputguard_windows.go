//go:build windows

package main

import "golang.org/x/sys/windows"

var (
	guardUser32       = windows.NewLazySystemDLL("user32.dll")
	guardGetAsyncKey  = guardUser32.NewProc("GetAsyncKeyState")
)

// mouseLeftDown reads the REAL hardware button state, not the state encoded in
// a WM_MOUSEMOVE. A Windows modal move/size loop can swallow WM_LBUTTONUP, so
// the only reliable way to notice a stale drag is to ask the OS whether the
// button is still physically down.
func mouseLeftDown() bool {
	r, _, _ := guardGetAsyncKey.Call(0x01) // VK_LBUTTON
	return r&0x8000 != 0
}
