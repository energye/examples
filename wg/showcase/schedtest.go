package main

import (
	"fmt"
	"time"

	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
)

// --schedtest asserts the invalidation-driven frame model end to end:
//  1. active animation paints at ~60 FPS
//  2. pausing stops the pump completely (ZERO frames while idle)
//  3. a dirty rect while idle produces EXACTLY ONE dirty-region frame
//  4. resuming returns to the 60Hz cadence
//
// It runs against the real window / real message loop, not a mock.
func startSchedTest(s *Scene) {
	go func() {
		for !s.sched.Active() {
			time.Sleep(5 * time.Millisecond)
		}
		runMain := func(fn func()) {
			done := make(chan struct{})
			lcl.RunOnMainThreadAsync(func(id uint32) {
				fn()
				close(done)
			})
			<-done
		}

		// Let the one-off startup bake finish, then measure REAL paints over
		// a 1s interval (not the smoothed display value).
		time.Sleep(250 * time.Millisecond)
		f0, t0 := s.Frame, time.Now()
		time.Sleep(1000 * time.Millisecond)
		fpsAnim := float64(s.Frame-f0) / time.Since(t0).Seconds()

		// --- idle: no animation => no frames at all ---
		runMain(func() { s.SetPaused(true) })
		time.Sleep(120 * time.Millisecond) // allow the final pause frame
		base := s.Frame
		postedAtPause, _, _ := s.sched.Stats()
		activeAtPause := s.sched.Active()
		time.Sleep(1000 * time.Millisecond)
		idleFrames := s.Frame - base
		postedAfterIdle, _, _ := s.sched.Stats()

		// --- one dirty rect => exactly one dirty-region frame ---
		dirtyBefore := s.prof.dirtyFrames
		runMain(func() {
			s.InvalidateRect(types.TRect{Left: 966, Top: 316, Right: 1416, Bottom: 604})
		})
		time.Sleep(300 * time.Millisecond)
		dirtyPaints := s.Frame - base - idleFrames
		dirtyRegionFrames := s.prof.dirtyFrames - dirtyBefore

		// --- resume => back to 60Hz ---
		// Measure over 2s: a single 1s window can lose ~10 frames to an
		// unrelated CPU load spike on this software (no VSync) clock, which
		// used to produce a false FAIL while the pump was still locked 60Hz.
		runMain(func() { s.SetPaused(false) })
		resumeBase, resumeStart := s.Frame, time.Now()
		time.Sleep(2000 * time.Millisecond)
		resumeFrames := s.Frame - resumeBase
		fpsResume := float64(resumeFrames) / time.Since(resumeStart).Seconds()

		posted, skipped, lost := s.sched.Stats()
		fmt.Println("==== SCHEDTEST (真实窗口 / 真实消息循环) ====")
		fmt.Printf("动画中 FPS: %.1f (期望 ~60)\n", fpsAnim)
		fmt.Printf("暂停 1s 内绘制帧数: %d (期望 0)   [诊断 active=%v 帧泵投递增量=%d]\n",
			idleFrames, activeAtPause, postedAfterIdle-postedAtPause)
		fmt.Printf("脏区请求产生帧数: %d (期望 1)   其中脏区帧: %d (期望 1)\n",
			dirtyPaints, dirtyRegionFrames)
		fmt.Printf("恢复 2s 内绘制帧数: %d (期望 ~120)   实测 %.1f FPS\n",
			resumeFrames, fpsResume)
		fmt.Printf("帧泵: 投递 %d / 跳过 %d / 丢弃 %d\n", posted, skipped, lost)

		ok := fpsAnim >= 55 && fpsAnim <= 65 &&
			idleFrames == 0 &&
			dirtyPaints == 1 && dirtyRegionFrames == 1 &&
			fpsResume >= 55 && fpsResume <= 65
		if ok {
			fmt.Println("RESULT: PASS — 60Hz 锁帧 / 静止零帧 / 脏区单帧 全部成立")
			requestExit(0)
		} else {
			fmt.Println("RESULT: FAIL")
			requestExit(1)
		}
	}()
}
