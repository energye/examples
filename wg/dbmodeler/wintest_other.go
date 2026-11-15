//go:build !windows

package main

// 非 Windows 平台没有 HWND 消息投递，--wintest 会明确报不支持。

import "fmt"

func winTestSupported() bool { return false }

func winPostMouse(hwnd uintptr, msg uint32, x, y int32, down bool) bool { return false }
func winPostKey(hwnd uintptr, vk uint16)                                {}
func winPostChar(hwnd uintptr, ch uint32)                               {}
func winPostWheel(hwnd uintptr, screenX, screenY int32, delta int16)    {}

func winCaptureClient(hwnd uintptr, path string) error {
	return fmt.Errorf("真实窗口截图仅支持 Windows")
}
