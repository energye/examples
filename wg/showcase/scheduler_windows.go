//go:build windows

package main

import "golang.org/x/sys/windows"

var (
	winmmDLL        = windows.NewLazySystemDLL("winmm.dll")
	procBeginPeriod = winmmDLL.NewProc("timeBeginPeriod")
	procEndPeriod   = winmmDLL.NewProc("timeEndPeriod")
	timerResolution uint32
)

// setTimerResolution raises the Windows multimedia timer granularity.
// The default ~15.6ms granularity quantises a 16ms TTimer to ~31ms ticks,
// which caps a 9ms paint at ~40 FPS. With timeBeginPeriod(1) the same 16ms
// interval actually fires at 16ms and the renderer reaches 60 FPS.
func setTimerResolution(ms uint32) {
	if r, _, _ := procBeginPeriod.Call(uintptr(ms)); r == 0 {
		timerResolution = ms
	}
}

func restoreTimerResolution() {
	if timerResolution != 0 {
		procEndPeriod.Call(uintptr(timerResolution))
		timerResolution = 0
	}
}
