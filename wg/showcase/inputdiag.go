//go:build windows

package main

// --inputdiag / --inputtest: evidence collector for the "resize / maximize
// kills the title bar and all mouse input while the animation keeps running"
// report. Nothing here runs unless one of the two flags is given.
//
// It logs the native message stream (WM_SIZE / WM_SYSCOMMAND / modal size-move
// loop notifications / capture notifications / mouse buttons), the Win32 mouse
// capture, the LCL graphic control's MouseCapture flag and the Scene's own
// capture pointer, so a stuck modal loop / stale capture can be told apart
// from a layout / hit-test mismatch instead of guessed.

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
	"github.com/energye/lcl/types/messages"
)

var (
	inputDiagMode bool
	inputTestMode bool

	// --inputtest accumulates real assertions; the scripted sequence is both
	// a reproduction and the permanent regression guard for the resize hang.
	diagChecks []struct {
		name   string
		ok     bool
		detail string
	}

	diagMu        sync.Mutex
	diagFile      *os.File
	diagStart     = time.Now()
	diagCount     = map[uint32]int{}
	diagDownCount atomic.Int32
)

func initInputDiag() {
	if !inputDiagMode && !inputTestMode {
		return
	}
	inputDiagMode = true
	f, err := os.Create("inputdiag.log")
	if err == nil {
		diagFile = f
	}
	diagLog("==== input diagnostics start (pid=%d) ====", os.Getpid())
}

func diagLog(format string, args ...any) {
	if !inputDiagMode {
		return
	}
	line := fmt.Sprintf("[%9.2fms] ", float64(time.Since(diagStart).Microseconds())/1000) +
		fmt.Sprintf(format, args...)
	diagMu.Lock()
	fmt.Println(line)
	if diagFile != nil {
		fmt.Fprintln(diagFile, line)
		diagFile.Sync()
	}
	diagMu.Unlock()
}

func diagCheck(name string, ok bool, format string, args ...any) {
	detail := fmt.Sprintf(format, args...)
	diagMu.Lock()
	diagChecks = append(diagChecks, struct {
		name   string
		ok     bool
		detail string
	}{name, ok, detail})
	diagMu.Unlock()
	diagLog("CHECK %-38s %s  %s", name, diagPassLabel(ok), detail)
}

func diagResults() (bool, string) {
	diagMu.Lock()
	defer diagMu.Unlock()
	ok := true
	out := ""
	for _, c := range diagChecks {
		if !c.ok {
			ok = false
		}
		out += fmt.Sprintf("  %-38s %s  %s\n", c.name, diagPassLabel(c.ok), c.detail)
	}
	return ok, out
}

var diagMsgNames = map[uint32]string{
	messages.WM_SIZE:              "WM_SIZE",
	messages.WM_ACTIVATE:          "WM_ACTIVATE",
	messages.WM_ACTIVATEAPP:       "WM_ACTIVATEAPP",
	messages.WM_CANCELMODE:        "WM_CANCELMODE",
	messages.WM_NCHITTEST:         "WM_NCHITTEST",
	messages.WM_NCLBUTTONDOWN:     "WM_NCLBUTTONDOWN",
	messages.WM_NCLBUTTONUP:       "WM_NCLBUTTONUP",
	messages.WM_NCLBUTTONDBLCLK:   "WM_NCLBUTTONDBLCLK",
	messages.WM_SYSCOMMAND:        "WM_SYSCOMMAND",
	messages.WM_MOUSEMOVE:         "WM_MOUSEMOVE",
	messages.WM_LBUTTONDOWN:       "WM_LBUTTONDOWN",
	messages.WM_LBUTTONUP:         "WM_LBUTTONUP",
	messages.WM_LBUTTONDBLCLK:     "WM_LBUTTONDBLCLK",
	messages.WM_SIZING:            "WM_SIZING",
	messages.WM_CAPTURECHANGED:    "WM_CAPTURECHANGED",
	messages.WM_ENTERSIZEMOVE:     "WM_ENTERSIZEMOVE",
	messages.WM_EXITSIZEMOVE:      "WM_EXITSIZEMOVE",
	messages.WM_WINDOWPOSCHANGING: "WM_WINDOWPOSCHANGING",
	messages.WM_WINDOWPOSCHANGED:  "WM_WINDOWPOSCHANGED",
}

// diagWndProc is called at the very top of the form's WndProc hook.
func diagWndProc(msg *types.TLMessage) {
	if !inputDiagMode {
		return
	}
	id := uint32(msg.Msg)
	name, tracked := diagMsgNames[id]
	if !tracked {
		return
	}
	// Mouse moves / sizing notifications arrive in bursts of hundreds; log the
	// first few of each burst and then a periodic summary.
	if id == messages.WM_MOUSEMOVE || id == messages.WM_SIZING ||
		id == messages.WM_WINDOWPOSCHANGED || id == messages.WM_NCHITTEST {
		diagMu.Lock()
		diagCount[id]++
		n := diagCount[id]
		diagMu.Unlock()
		if n > 4 && n%20 != 0 {
			return
		}
	}
	extra := ""
	switch id {
	case messages.WM_SIZE:
		extra = fmt.Sprintf(" sizeType=%d client=%dx%d", uint32(msg.WParam),
			uint32(msg.LParam)&0xffff, (uint32(msg.LParam)>>16)&0xffff)
	case messages.WM_SYSCOMMAND:
		extra = fmt.Sprintf(" cmd=0x%04X", uint32(msg.WParam)&0xfff0)
	case messages.WM_ACTIVATE:
		extra = fmt.Sprintf(" active=%d", uint32(msg.WParam)&0xffff)
	case messages.WM_NCLBUTTONDOWN:
		extra = fmt.Sprintf(" hit=%d", uint32(msg.WParam))
	}
	diagLog("WND %-20s wp=0x%X lp=0x%X%s cap=0x%X", name,
		uintptr(msg.WParam), uintptr(msg.LParam), extra, diagGetCapture())
}

// diagWndProcAfter logs what the inherited WndProc wrote back for NC hit tests.
// WM_NCHITTEST reaching the window with HTCAPTION is the decisive evidence that
// the title bar is still alive.
func diagWndProcAfter(msg *types.TLMessage) {
	if !inputDiagMode || uint32(msg.Msg) != messages.WM_NCHITTEST {
		return
	}
	diagLog("WND WM_NCHITTEST -> result=%d cap=0x%X", int64(msg.Result), diagGetCapture())
}

func diagDispatch(e *UIEvent, s *Scene, hit Widget) {
	if !inputDiagMode {
		return
	}
	if e.Kind == EvMouseDown {
		diagDownCount.Add(1)
	}
	hitName := "nil"
	if hit != nil {
		hitName = hit.Name()
	}
	capName := "nil"
	if s.capture != nil {
		capName = s.capture.Name()
	}
	diagLog("DISPATCH kind=%d xy=(%d,%d) hit=%s cap=%s lclCapture=%v osCapture=0x%X",
		e.Kind, e.X, e.Y, hitName, capName, s.host.MouseCapture(), diagGetCapture())
}

func diagDispatchResult(e *UIEvent, s *Scene) {
	if !inputDiagMode {
		return
	}
	capName := "nil"
	if s.capture != nil {
		capName = s.capture.Name()
	}
	diagLog("DISPATCH-END kind=%d cap=%s lclCapture=%v osCapture=0x%X",
		e.Kind, capName, s.host.MouseCapture(), diagGetCapture())
}

// diagFrame is sampled every N frames from OnPaint: it shows whether frames
// are actually being produced at 60Hz or the main thread is saturated.
func diagFrame(s *Scene) {
	if !inputDiagMode {
		return
	}
	r := s.host.ClientRect()
	diagLog("FRAME #%-5d client=%dx%d paint=%6dus fps=%5.1f cap=0x%X hold=%v",
		s.Frame, r.Width(), r.Height(), s.PaintUs, s.FPS, diagGetCapture(), s.holdStatic)
}

func diagState(s *Scene, tag string) {
	if !inputDiagMode {
		return
	}
	cr := s.host.ClientRect()
	capName := "nil"
	if s.capture != nil {
		capName = s.capture.Name()
	}
	posted, skipped, lost := s.sched.Stats()
	diagLog("STATE %-16s frame=%d client=%dx%d zoomed=%v osCapture=0x%X lclCapture=%v "+
		"sceneCapture=%s theme=%s layout=(%dx%d scale=%.3f) hold=%v bake=%dus "+
		"healed=%d sched(posted=%d skipped=%d lost=%d inflight=%d)",
		tag, s.Frame, cr.Width(), cr.Height(), diagIsZoomed(), diagGetCapture(),
		s.host.MouseCapture(), capName, s.ThemeName(), s.layout.cw, s.layout.ch,
		s.layout.scale, s.holdStatic, s.staticBakeUs, s.staleCaptureFixes,
		posted, skipped, lost, atomic.LoadInt32(&s.sched.inflight))
}

// ---------------------------------------------------------------------------
// Win32 helpers
// ---------------------------------------------------------------------------

var (
	diagUser32         = windows.NewLazySystemDLL("user32.dll")
	procGetCapture     = diagUser32.NewProc("GetCapture")
	procReleaseCapture = diagUser32.NewProc("ReleaseCapture")
	procIsZoomed       = diagUser32.NewProc("IsZoomed")
	procShowWindow     = diagUser32.NewProc("ShowWindow")
	procGetWindowRect  = diagUser32.NewProc("GetWindowRect")
	procGetForeground  = diagUser32.NewProc("GetForegroundWindow")
	procSetForeground  = diagUser32.NewProc("SetForegroundWindow")
	procSendMessageW   = diagUser32.NewProc("SendMessageW")
	procPostMessageW   = diagUser32.NewProc("PostMessageW")
	procGetAsyncKey    = diagUser32.NewProc("GetAsyncKeyState")
	procSendInput      = diagUser32.NewProc("SendInput")
	procGetSystemMet   = diagUser32.NewProc("GetSystemMetrics")
)

// INPUT / MOUSEINPUT for SendInput (x64 layout: 4 + 4 pad + 32 = 40 bytes).
type diagMouseInput struct {
	Dx, Dy      int32
	MouseData   uint32
	Flags       uint32
	Time        uint32
	ExtraInfo   uintptr
}

type diagNativeInput struct {
	Type uint32
	_    uint32
	Mi   diagMouseInput
}

const (
	diagMOUSEEVENTF_MOVE     = 0x0001
	diagMOUSEEVENTF_LEFTDOWN = 0x0002
	diagMOUSEEVENTF_LEFTUP   = 0x0004
	diagMOUSEEVENTF_ABSOLUTE = 0x8000
)

func diagSendMouse(flags uint32, x, y int32, absolute bool) bool {
	var in diagNativeInput
	in.Type = 0 // INPUT_MOUSE
	in.Mi.Flags = flags
	if absolute {
		in.Mi.Flags |= diagMOUSEEVENTF_ABSOLUTE
		sw, _, _ := procGetSystemMet.Call(0)
		sh, _, _ := procGetSystemMet.Call(1)
		if sw > 1 && sh > 1 {
			in.Mi.Dx = int32(int64(x) * 65535 / (int64(sw) - 1))
			in.Mi.Dy = int32(int64(y) * 65535 / (int64(sh) - 1))
		}
	} else {
		in.Mi.Dx, in.Mi.Dy = x, y
	}
	r, _, _ := procSendInput.Call(1, uintptr(unsafe.Pointer(&in)),
		uintptr(unsafe.Sizeof(in)))
	return r == 1
}

func diagMoveCursor(x, y int32) {
	diagSendMouse(diagMOUSEEVENTF_MOVE, x, y, true)
}

func diagClickScreen(x, y int32) {
	diagMoveCursor(x, y)
	time.Sleep(20 * time.Millisecond)
	diagSendMouse(diagMOUSEEVENTF_LEFTDOWN, 0, 0, false)
	time.Sleep(30 * time.Millisecond)
	diagSendMouse(diagMOUSEEVENTF_LEFTUP, 0, 0, false)
}

func diagGetCapture() uintptr {
	r, _, _ := procGetCapture.Call()
	return r
}

func diagReleaseCapture() bool {
	r, _, _ := procReleaseCapture.Call()
	return r != 0
}

func diagIsZoomed() bool {
	h := uintptr(mainForm.Handle())
	if h == 0 {
		return false
	}
	r, _, _ := procIsZoomed.Call(h)
	return r != 0
}

func diagShowWindow(cmd uintptr) uintptr {
	r, _, _ := procShowWindow.Call(uintptr(mainForm.Handle()), cmd)
	return r
}

func diagPost(msg, wp, lp uintptr) uintptr {
	r, _, _ := procPostMessageW.Call(uintptr(mainForm.Handle()), msg, wp, lp)
	return r
}

func diagSend(msg, wp, lp uintptr) uintptr {
	r, _, _ := procSendMessageW.Call(uintptr(mainForm.Handle()), msg, wp, lp)
	return r
}

type diagRect struct{ Left, Top, Right, Bottom int32 }

func diagWindowRectSafe() (diagRect, bool) {
	var r diagRect
	ok, _, _ := procGetWindowRect.Call(uintptr(mainForm.Handle()),
		uintptr(unsafe.Pointer(&r)))
	return r, ok != 0
}

func diagGetForegroundWindow() uintptr {
	r, _, _ := procGetForeground.Call()
	return r
}

func diagSetForegroundWindow() {
	procSetForeground.Call(uintptr(mainForm.Handle()))
}

func diagLeftButtonDown() bool { return mouseLeftDown() }

func diagLParam(x, y int32) uintptr {
	return (uintptr(uint32(y)) << 16) | uintptr(uint32(x))
}

// ---------------------------------------------------------------------------
// --inputtest scripted reproduction
// ---------------------------------------------------------------------------

func startInputTest(s *Scene) {
	go func() {
		for !s.sched.Active() {
			time.Sleep(5 * time.Millisecond)
		}
		time.Sleep(400 * time.Millisecond)
		hwnd := uintptr(mainForm.Handle())
		diagLog("==== INPUTTEST hwnd=0x%X foreground=0x%X ====", hwnd, diagGetForegroundWindow())
		diagSetForegroundWindow()
		time.Sleep(150 * time.Millisecond)
		diagLog("==== foreground after SetForegroundWindow=0x%X ====", diagGetForegroundWindow())

		diagState(s, "baseline")
		diagCheck("基线标题栏可命中", diagProbeTitleBar(s, "baseline") >= 2, "NCHITTEST 应返回非客户区命中码")

		// Step 1: maximize through the real system command path.
		diagLog("---- STEP 1 SendMessage(WM_SYSCOMMAND, SC_MAXIMIZE) ----")
		diagSend(messages.WM_SYSCOMMAND, 0xF030, 0)
		time.Sleep(1200 * time.Millisecond) // 让 2880x1557 的静态层重烘焙完成
		diagState(s, "after-maximize")
		diagCheck("最大化生效", diagIsZoomed(), "IsZoomed=true")
		diagCheck("最大化后标题栏可命中", diagProbeTitleBar(s, "after-maximize") >= 2,
			"NCHITTEST 应返回 HTCAPTION/HTTOP")
		diagCheck("最大化后无残留鼠标捕获", diagGetCapture() == 0 && !s.host.MouseCapture() && s.capture == nil,
			"osCapture=0x%X lclCapture=%v sceneCapture=%v", diagGetCapture(), s.host.MouseCapture(), s.capture != nil)
		maxClick, _ := diagProbeWidgetClick(s, "after-maximize")
		diagCheck("最大化后点击生效", maxClick, "主题按钮应切换主题")
		maxLatency := diagProbeClickLatency(s, "pump-ON@max")
		diagCheck("最大化下输入延迟 <120ms", maxLatency < 120*time.Millisecond,
			"实测 %v（修复前 382ms）", maxLatency.Round(time.Millisecond))
		diagCheck("帧消息队列最多 1 个 outstanding", atomic.LoadInt32(&s.sched.inflight) <= 1,
			"inflight=%d", atomic.LoadInt32(&s.sched.inflight))

		// Step 2: restore through the real system command path.
		diagLog("---- STEP 2 SendMessage(WM_SYSCOMMAND, SC_RESTORE) ----")
		diagSend(messages.WM_SYSCOMMAND, 0xF120, 0)
		time.Sleep(900 * time.Millisecond)
		diagState(s, "after-restore")
		diagCheck("还原生效", !diagIsZoomed(), "IsZoomed=false")
		diagCheck("还原后标题栏可命中", diagProbeTitleBar(s, "after-restore") >= 2, "NCHITTEST 正常")
		restoreClick, _ := diagProbeWidgetClick(s, "after-restore")
		diagCheck("还原后点击生效", restoreClick, "主题按钮应切换主题")

		// Step 3: a client mouse-down whose WM_LBUTTONUP is swallowed by the
		// modal size/move loop (the exact sequence the handoff suspects).
		diagLog("---- STEP 3 mouse-down on widget, no button-up (lost-up simulation) ----")
		healedBefore := s.staleCaptureFixes
		p := diagThemeButtonCenter(s)
		lp := diagLParam(p.X, p.Y)
		diagPost(messages.WM_LBUTTONDOWN, 0x0001, lp)
		time.Sleep(150 * time.Millisecond)
		diagState(s, "after-down")
		diagLog("STEP3 note: sceneCapture=%v healed=%d（可能被系统鼠标移动提前自愈）",
			s.capture != nil, s.staleCaptureFixes)
		diagLog("---- STEP 3b post WM_NCLBUTTONDOWN(HTCAPTION) -> modal loop ----")
		sr, _ := diagWindowRectSafe()
		nclp := diagLParam(sr.Left+(sr.Right-sr.Left)/2, sr.Top+8)
		diagPost(messages.WM_NCLBUTTONDOWN, 2 /*HTCAPTION*/, nclp)
		time.Sleep(250 * time.Millisecond)
		diagLog("---- STEP 3c post WM_MOUSEMOVE + WM_LBUTTONUP to end the loop ----")
		diagPost(messages.WM_MOUSEMOVE, 0x0001, diagLParam(p.X+40, p.Y+20))
		diagPost(messages.WM_LBUTTONUP, 0, diagLParam(p.X+40, p.Y+20))
		time.Sleep(600 * time.Millisecond)
		diagState(s, "after-modalloop")
		diagProbeTitleBar(s, "after-modalloop")
		diagProbeWidgetClick(s, "after-modalloop")
		diagCheck("丢失 MouseUp 后 capture 不残留", s.capture == nil,
			"healed=%d (before=%d) sceneCapture=%v", s.staleCaptureFixes, healedBefore, s.capture != nil)

		// Step 4: hold a maximized (much larger) window for 2.5s and measure
		// real frame throughput + input latency. A saturated main thread would
		// starve input while the frame pump keeps producing frames.
		diagLog("---- STEP 4 ShowWindow(SW_MAXIMIZE), hold 2.5s ----")
		diagShowWindow(3)
		f0, t0 := s.Frame, time.Now()
		time.Sleep(2500 * time.Millisecond)
		diagLog("STEP4 maximized: frames=%d / %v = %.1f FPS  paint=%dus fpsPanel=%.1f",
			s.Frame-f0, time.Since(t0).Round(time.Millisecond),
			float64(s.Frame-f0)/time.Since(t0).Seconds(), s.PaintUs, s.FPS)
		diagState(s, "max-held")
		diagProbeTitleBar(s, "max-held")
		heldClick, _ := diagProbeWidgetClick(s, "max-held")
		diagCheck("持续最大化下点击生效", heldClick, "修复前此处为 FAIL（输入被帧消息饿死）")
		diagCheck("最大化下动画仍在渲染", s.Frame-f0 >= 20, "2.5s 内绘制 %d 帧", s.Frame-f0)
		_, _, lostNow := s.sched.Stats()
		diagCheck("慢帧后帧队列仍不堆积", atomic.LoadInt32(&s.sched.inflight) <= 1,
			"inflight=%d lost=%d", atomic.LoadInt32(&s.sched.inflight), lostNow)
		diagLog("---- STEP 4b ShowWindow(SW_RESTORE), hold 1s ----")
		diagShowWindow(9)
		f0, t0 = s.Frame, time.Now()
		time.Sleep(1000 * time.Millisecond)
		diagLog("STEP4b restored: frames=%d / %v = %.1f FPS  paint=%dus",
			s.Frame-f0, time.Since(t0).Round(time.Millisecond),
			float64(s.Frame-f0)/time.Since(t0).Seconds(), s.PaintUs)
		diagState(s, "restored")
		diagCheck("还原后标题栏仍可命中", diagProbeTitleBar(s, "restored") >= 2, "NCHITTEST 正常")
		restoredClick, _ := diagProbeWidgetClick(s, "restored")
		diagCheck("还原后点击仍生效", restoredClick, "主题按钮应切换主题")

		// Step 5: REAL mouse drag of the window border (SendInput). This is the
		// only way to enter Windows' true modal size/move loop.
		diagLog("---- STEP 5 real SendInput drag on right border ----")
		if r, ok := diagWindowRectSafe(); ok {
			x, y := r.Right-2, r.Top+(r.Bottom-r.Top)/2
			diagMoveCursor(x, y)
			time.Sleep(60 * time.Millisecond)
			diagSendMouse(diagMOUSEEVENTF_LEFTDOWN, 0, 0, false)
			for i := 0; i < 8; i++ {
				diagSendMouse(diagMOUSEEVENTF_MOVE, 12, 0, false)
				time.Sleep(25 * time.Millisecond)
			}
			time.Sleep(60 * time.Millisecond)
			diagSendMouse(diagMOUSEEVENTF_LEFTUP, 0, 0, false)
		}
		time.Sleep(900 * time.Millisecond)
		diagState(s, "after-real-drag")
		diagCheck("真实拖拽后标题栏可命中", diagProbeTitleBar(s, "after-real-drag") >= 2, "NCHITTEST 正常")
		dragClick, _ := diagProbeWidgetClick(s, "after-real-drag")
		diagCheck("真实拖拽后点击生效", dragClick, "主题按钮应切换主题")
		diagCheck("真实拖拽后无残留鼠标捕获", diagGetCapture() == 0 && s.capture == nil,
			"osCapture=0x%X sceneCapture=%v", diagGetCapture(), s.capture != nil)

		// Step 6: real double-click on the caption -> maximize, again -> restore.
		diagLog("---- STEP 6 real double-click caption (maximize) ----")
		if r, ok := diagWindowRectSafe(); ok {
			cx, cy := r.Left+(r.Right-r.Left)/2, r.Top+15
			diagClickScreen(cx, cy)
			time.Sleep(70 * time.Millisecond)
			diagClickScreen(cx, cy)
		}
		time.Sleep(700 * time.Millisecond)
		diagState(s, "after-real-dblclk")
		diagCheck("标题栏双击后仍可命中", diagProbeTitleBar(s, "after-real-dblclk") >= 2, "NCHITTEST 正常")
		dblClick, _ := diagProbeWidgetClick(s, "after-real-dblclk")
		diagCheck("标题栏双击后点击生效", dblClick, "主题按钮应切换主题")

		// Step 7: decisive A/B on the frame pump. At the large (maximized)
		// size, a click must be dispatched promptly with the pump running; if
		// it is not, but becomes immediate when the pump is stopped, the frame
		// post queue is starving input rather than a stale mouse capture.
		diagLog("---- STEP 7 A/B: input latency with frame pump ON / OFF ----")
		diagShowWindow(3) // SW_MAXIMIZE
		time.Sleep(1500 * time.Millisecond)
		diagState(s, "ab-pump-on")
		onLatency := diagProbeClickLatency(s, "pump-ON")
		s.sched.SetActive(false)
		time.Sleep(400 * time.Millisecond)
		diagState(s, "ab-pump-off")
		offLatency := diagProbeClickLatency(s, "pump-OFF")
		diagLog("STEP7 RESULT pumpON=%v pumpOFF=%v ratio=%.1fx  (capture=0x%X)",
			onLatency.Round(time.Millisecond), offLatency.Round(time.Millisecond),
			float64(onLatency)/float64(offLatency+time.Millisecond), diagGetCapture())
		diagCheck("帧泵开启时输入延迟 <120ms", onLatency < 120*time.Millisecond,
			"pumpON=%v pumpOFF=%v（修复前 382ms / 14ms）",
			onLatency.Round(time.Millisecond), offLatency.Round(time.Millisecond))
		s.sched.SetActive(true)
		diagShowWindow(9) // SW_RESTORE
		time.Sleep(300 * time.Millisecond)

		ok, report := diagResults()
		fmt.Println("==== INPUTTEST (最大化/还原/真实拖拽 + 输入延迟) ====")
		fmt.Print(report)
		if ok {
			fmt.Println("RESULT: PASS — 调整大小/最大化后标题栏、还原、鼠标操作均正常")
			requestExit(0)
		} else {
			fmt.Println("RESULT: FAIL — 输入被帧消息队列饿死或 capture 残留")
			requestExit(1)
		}
	}()
}

// diagThemeButtonCenter returns the center of the "切换主题" mini control in
// CLIENT coordinates, so injected mouse messages hit a real interactive widget.
func diagThemeButtonCenter(s *Scene) struct{ X, Y int32 } {
	for _, w := range s.widgets {
		if c, ok := w.(*Controls); ok {
			for _, it := range c.items {
				if it.kind == "theme" {
					return struct{ X, Y int32 }{
						(it.r.Left + it.r.Right) / 2,
						(it.r.Top + it.r.Bottom) / 2,
					}
				}
			}
		}
	}
	return struct{ X, Y int32 }{100, 100}
}

// diagProbeTitleBar asks Windows for an NC hit test at the caption area.
func diagProbeTitleBar(s *Scene, tag string) int64 {
	r, ok := diagWindowRectSafe()
	if !ok {
		diagLog("PROBE %s: GetWindowRect failed", tag)
		return -1
	}
	x := r.Left + (r.Right-r.Left)/2
	y := r.Top + 8
	hit := diagSend(messages.WM_NCHITTEST, 0, diagLParam(x, y))
	diagLog("PROBE %-16s NCHITTEST(caption) -> %d (2=HTCAPTION) cap=0x%X leftDown=%v",
		tag, int64(hit), diagGetCapture(), diagLeftButtonDown())
	return int64(hit)
}

// diagProbeWidgetClick injects a click on the theme button and checks the
// theme actually toggles, i.e. the native message -> LCL control -> scene
// dispatch chain is alive end to end.
func diagProbeWidgetClick(s *Scene, tag string) (bool, string) {
	before := s.ThemeName()
	p := diagThemeButtonCenter(s)
	lp := diagLParam(p.X, p.Y)
	diagPost(messages.WM_LBUTTONDOWN, 0x0001, lp)
	time.Sleep(30 * time.Millisecond)
	diagPost(messages.WM_LBUTTONUP, 0, lp)
	time.Sleep(200 * time.Millisecond)
	after := s.ThemeName()
	ok := before != after
	diagLog("PROBE %-16s click theme at (%d,%d): %s -> %s  %s  cap=0x%X",
		tag, p.X, p.Y, before, after, diagPassLabel(ok), diagGetCapture())
	return ok, after
}

// diagProbeClickLatency posts a side-effect-free click on empty canvas and
// measures how long the main thread takes to dispatch WM_LBUTTONDOWN. This
// isolates message-queue starvation: with the 60Hz post queue saturated,
// hardware/posted input is starved; with the pump stopped it is immediate.
func diagProbeClickLatency(s *Scene, tag string) time.Duration {
	start := diagDownCount.Load()
	t0 := time.Now()
	lp := diagLParam(30, 30) // outside every hitAble widget: no repaint, no bake
	diagPost(messages.WM_LBUTTONDOWN, 0x0001, lp)
	diagPost(messages.WM_LBUTTONUP, 0, lp)
	for diagDownCount.Load() == start && time.Since(t0) < 2*time.Second {
		time.Sleep(time.Millisecond)
	}
	d := time.Since(t0)
	diagLog("LATENCY %-16s dispatch=%v  frames=%d cap=0x%X",
		tag, d.Round(time.Millisecond), s.Frame, diagGetCapture())
	return d
}

func diagPassLabel(ok bool) string {
	if ok {
		return "PASS"
	}
	return "FAIL"
}

// diagRunOnMain runs fn on the LCL main thread with a timeout so a stuck modal
// loop shows up as TIMEOUT instead of hanging the whole test process.
func diagRunOnMain(fn func()) bool {
	done := make(chan struct{})
	lcl.RunOnMainThreadAsync(func(id uint32) {
		fn()
		close(done)
	})
	select {
	case <-done:
		return true
	case <-time.After(3 * time.Second):
		return false
	}
}
