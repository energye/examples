//go:build windows

package main

import (
	"fmt"
	"time"

	"golang.org/x/sys/windows"
)

// startMouseTest reproduces the reported bug: a continuous stream of
// WM_MOUSEMOVE messages. WM_PAINT and WM_TIMER are low-priority messages on
// Windows, so without a synchronous repaint path this freezes the animation.
// Run: go run ./wg/showcase --mousetest
const wmMouseMove = 0x0200

var (
	user32dll = windows.NewLazyDLL("user32.dll")
	postMsg   = user32dll.NewProc("PostMessageW")
)

func startMouseTest(s *Scene) {
	go func() {
		time.Sleep(700 * time.Millisecond)
		hwnd := uintptr(mainForm.Handle())
		startPhase := s.Phase
		startFrame := s.Frame
		deadline := time.Now().Add(3 * time.Second)
		n := 0
		for time.Now().Before(deadline) {
			t := time.Now().UnixMilli()
			x := int32(200 + t%700)
			y := int32(150 + (t/3)%400)
			lp := (uintptr(uint32(y)) << 16) | uintptr(uint32(x))
			postMsg.Call(hwnd, wmMouseMove, 0, lp)
			n++
			time.Sleep(2 * time.Millisecond)
		}
		dPhase := s.Phase - startPhase
		dFrame := s.Frame - startFrame
		fmt.Printf("==== MOUSETEST (%d 条 WM_MOUSEMOVE / 3s) ====\n", n)
		fmt.Printf("相位推进 %.2fs (应≈3.0)   实际绘制 %d 帧   FPS %.1f\n", dPhase, dFrame, s.FPS)
		for _, line := range s.StatusLines() {
			fmt.Println(line)
		}
		if dPhase < 2.0 || dFrame < 30 {
			fmt.Println("RESULT: FAIL — 鼠标移动期间动画被饿死")
			requestExit(1)
			return
		}
		fmt.Println("RESULT: PASS — 鼠标移动期间动画持续运行")
		requestExit(0)
	}()
}
