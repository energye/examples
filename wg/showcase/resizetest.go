package main

// --resizetest drives real WM_SIZE cycles through the message loop to catch
// the resize crash/hang. It also reports the exact client geometry so a
// layout offset (content sitting too low) can be measured instead of guessed.

import (
	"fmt"
	"time"

	"github.com/energye/lcl/lcl"
)

type sizeStep struct {
	w, h int32
}

func startResizeTest(s *Scene) {
	go func() {
		for !s.sched.Active() {
			time.Sleep(5 * time.Millisecond)
		}
		time.Sleep(250 * time.Millisecond)

		runMain := func(fn func()) bool {
			done := make(chan struct{})
			lcl.RunOnMainThreadAsync(func(id uint32) {
				fn()
				close(done)
			})
			select {
			case <-done:
				return true
			case <-time.After(4 * time.Second):
				fmt.Println("[RESIZETEST] TIMEOUT: main thread did not return")
				return false
			}
		}

		report := func(tag string) {
			runMain(func() {
				fc := mainForm.ClientRect()
				hc := s.host.ClientRect()
				fmt.Printf("[RESIZETEST] %-12s form client=(%d,%d,%d,%d) host client=(%d,%d,%d,%d) frame=%d\n",
					tag, fc.Left, fc.Top, fc.Right, fc.Bottom,
					hc.Left, hc.Top, hc.Right, hc.Bottom, s.Frame)
			})
		}
		report("startup")

		steps := []sizeStep{
			{1200, 760}, {1000, 640}, {1600, 960}, {900, 600},
			{1440, 880}, {1280, 720}, {1440, 880},
		}
		ok := true
		for i, st := range steps {
			f0 := s.Frame
			t0 := time.Now()
			if !runMain(func() {
				mainForm.SetWidth(st.w)
				mainForm.SetHeight(st.h)
				s.Invalidate()
			}) {
				ok = false
				break
			}
			time.Sleep(350 * time.Millisecond)
			f1 := s.Frame
			fmt.Printf("[RESIZETEST] step %d -> %dx%d  frames=%d  elapsed=%v\n",
				i+1, st.w, st.h, f1-f0, time.Since(t0).Round(time.Millisecond))
			report(fmt.Sprintf("step%d", i+1))
			if f1-f0 == 0 {
				fmt.Println("[RESIZETEST] WARN: no frames painted after resize (message loop stalled)")
				ok = false
			}
		}

		fmt.Printf("[RESIZETEST] final frame=%d  AA=%v  backBuf=%dx%d  AA surface=%dx%d\n",
			s.Frame, s.aaMode, s.backW, s.backH, s.aaW, s.aaH)
		if ok {
			fmt.Println("RESULT: PASS — 全部尺寸切换未卡死")
			requestExit(0)
		} else {
			fmt.Println("RESULT: FAIL")
			requestExit(1)
		}
	}()
}
