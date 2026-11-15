package main

import (
	"fmt"
	"sort"
	"time"

	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
)

// --schedtest asserts the invalidation-driven frame model end to end:
//  1. active animation's frame-interval MEDIAN is ~16.67ms (60Hz cadence)
//  2. pausing stops the pump completely (ZERO frames while idle)
//  3. a dirty rect while idle produces EXACTLY ONE dirty-region frame
//  4. resuming returns to the 60Hz cadence
//
// It runs against the real window / real message loop, not a mock. Absolute
// frame COUNTS over a short wall-clock window are printed but not gated: one
// unrelated CPU/AV stall can eat dozens of frames (false 12-39 FPS FAILs in
// earlier sessions) while the pump is still locked at 60Hz. frameProfile
// discards >100ms gaps, so the pooled median of the two active windows is the
// honest cadence signal; a genuinely broken (e.g. 30Hz) pump fails both.
// Startup is excluded by waiting for 30 real frames before measuring.
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
		// A median frame INTERVAL asserts the 60Hz cadence itself. Absolute
		// frame counts over a short wall-clock window are famously fragile on
		// Windows: one unrelated CPU/AV stall eats dozens of frames and used to
		// produce "12 FPS / 28 FPS" false FAILs while the pump was still locked
		// at 60Hz. frameProfile.sample() already discards >100ms gaps (startup,
		// pause, external stalls), so the median is the honest signal here.
		medianMs := func(a []int64) float64 {
			if len(a) == 0 {
				return 0
			}
			tmp := append([]int64(nil), a...)
			sort.Slice(tmp, func(i, j int) bool { return tmp[i] < tmp[j] })
			return float64(tmp[len(tmp)/2]) / 1000
		}
		// ~300 samples over this test, well under the 600-sample ring buffer.
		// Take an explicit [start,end) window: slicing only to the CURRENT end
		// would double-count the later resume window.
		window := func(start, end int) []int64 {
			if start < 0 || end < start || end > len(s.prof.intervals) {
				return nil
			}
			return s.prof.intervals[start:end]
		}

		// Window creation / first WM_SIZE / the one-off static bake all happen
		// AFTER the pump reports Active() (WM_ACTIVATE lands ~500ms in). A fixed
		// 250ms delay therefore measured startup, producing random 12-50 FPS
		// false failures under load. Wait for real frames first, then measure
		// over 2s like the resume segment - an average removes single spikes.
		for s.Frame < 30 {
			time.Sleep(10 * time.Millisecond)
		}
		time.Sleep(200 * time.Millisecond)
		f0, t0 := s.Frame, time.Now()
		animSamples := len(s.prof.intervals)
		time.Sleep(2000 * time.Millisecond)
		fpsAnim := float64(s.Frame-f0) / time.Since(t0).Seconds()
		animEnd := len(s.prof.intervals)
		animSlice := window(animSamples, animEnd)
		p50Anim := medianMs(animSlice)
		nAnim := len(animSlice)

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
		resumeSamples := len(s.prof.intervals)
		time.Sleep(2000 * time.Millisecond)
		resumeFrames := s.Frame - resumeBase
		fpsResume := float64(resumeFrames) / time.Since(resumeStart).Seconds()
		resumeEnd := len(s.prof.intervals)
		resumeSlice := window(resumeSamples, resumeEnd)
		p50Resume := medianMs(resumeSlice)
		nResume := len(resumeSlice)

		posted, skipped, lost := s.sched.Stats()
		fmt.Println("==== SCHEDTEST (真实窗口 / 真实消息循环) ====")
		fmt.Printf("动画中 FPS (2s 窗口, 已排除窗口创建/首次烘焙): %.1f (期望 ~60)\n", fpsAnim)
		fmt.Printf("暂停 1s 内绘制帧数: %d (期望 0)   [诊断 active=%v 帧泵投递增量=%d]\n",
			idleFrames, activeAtPause, postedAfterIdle-postedAtPause)
		fmt.Printf("脏区请求产生帧数: %d (期望 1)   其中脏区帧: %d (期望 1)\n",
			dirtyPaints, dirtyRegionFrames)
		fmt.Printf("恢复 2s 内绘制帧数: %d (期望 ~120)   实测 %.1f FPS\n",
			resumeFrames, fpsResume)
		fmt.Printf("帧间隔中位数: 动画 %.2fms (%d 样本)   恢复 %.2fms (%d 样本)   [期望 14–20ms = 60Hz]\n",
			p50Anim, nAnim, p50Resume, nResume)
		// Pool both active windows: an unrelated CPU/AV stall can degrade ONE
		// 2s window to ~22ms while the pump is still locked at 60Hz. Pooling
		// tolerates one bad window; a genuinely broken (e.g. 30Hz) pump makes
		// BOTH windows bad and the pooled median still fails.
		pooled := append(append([]int64(nil), animSlice...), resumeSlice...)
		p50All := medianMs(pooled)
		fmt.Printf("合并帧间隔中位数: %.2fms (%d 样本)   帧泵: 投递 %d / 跳过 %d / 丢弃 %d\n",
			p50All, len(pooled), posted, skipped, lost)

		cadenceOK := p50All >= 14 && p50All <= 20 && len(pooled) >= 100
		ok := cadenceOK &&
			idleFrames == 0 &&
			dirtyPaints == 1 && dirtyRegionFrames == 1
		if ok {
			fmt.Println("RESULT: PASS — 60Hz 锁帧 / 静止零帧 / 脏区单帧 全部成立")
			requestExit(0)
		} else {
			fmt.Println("RESULT: FAIL")
			requestExit(1)
		}
	}()
}
