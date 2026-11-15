//go:build windows

package main

import "golang.org/x/sys/windows"

var postMessageW = windows.NewLazySystemDLL("user32.dll").NewProc("PostMessageW")

// postFrameMessage hands a frame slot to the MAIN thread's message queue.
// Only the main thread ever touches GDI; the scheduler goroutine just posts.
func postFrameMessage(h uintptr) bool {
	if h == 0 {
		return false
	}
	r, _, _ := postMessageW.Call(h, wmRequestFrame, 0, 0)
	return r != 0
}
