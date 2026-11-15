// Single-canvas showcase: a large, dense window with animated graphics,
// semi-transparent layers, irregular shapes and rounded shapes.
//
// Everything (including the glass cards, gauges, charts and controls) is a
// pure Go object drawn on ONE handle-less canvas. The only HWND is the form.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/energye/lcl/api/libname"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
	"github.com/energye/lcl/types/colors"
	"github.com/energye/lcl/types/messages"
)

var (
	benchMode      bool
	mouseTestMode  bool
	schedTestMode  bool
	resizeTestMode bool
	screenshotPath string
	exitCode       = -1 // -1 = 正常窗口运行，>=0 = 自检结束后请求退出
	// WG_ERASE=1 恢复系统擦背景，用来复现修复前的闪烁（A/B 对照用）。
	allowErase = os.Getenv("WG_ERASE") != ""
)

// requestExit ends the message loop from the MAIN thread. Calling os.Exit()
// from a worker goroutine while a Go<->DLL WndProc callback is in flight trips
// the runtime's "exitsyscall: syscall frame is no longer valid" assertion.
func requestExit(code int) {
	if mainForm.scene != nil {
		mainForm.scene.sched.Stop()
	}
	lcl.RunOnMainThreadAsync(func(id uint32) {
		exitCode = code
		lcl.Application.Terminate()
	})
}

type TShowcaseForm struct {
	lcl.TEngForm
	scene *Scene
}

var mainForm TShowcaseForm

func main() {
	setupLib()
	disableDoubleBuffer = os.Getenv("WG_FAST") != ""
	if len(os.Args) > 1 && os.Args[1] == "--selftest" {
		os.Exit(runSelfTest())
	}
	if len(os.Args) > 1 && os.Args[1] == "--aaprobe" {
		os.Exit(runAAProbe())
	}
	if len(os.Args) > 1 && os.Args[1] == "--bench" {
		benchMode = true
	}
	if len(os.Args) > 1 && os.Args[1] == "--mousetest" {
		mouseTestMode = true
	}
	if len(os.Args) > 1 && os.Args[1] == "--schedtest" {
		schedTestMode = true
	}
	if len(os.Args) > 1 && os.Args[1] == "--resizetest" {
		resizeTestMode = true
	}
	if len(os.Args) > 1 && os.Args[1] == "--screenshot" {
		screenshotPath = "screenshot.png"
		if len(os.Args) > 2 {
			screenshotPath = os.Args[2]
		}
	}

	lcl.Init()
	lcl.Application.Initialize()
	lcl.Application.SetMainFormOnTaskBar(true)
	lcl.Application.NewForm(&mainForm)
	lcl.Application.Run()
	restoreTimerResolution()
	if exitCode >= 0 {
		os.Exit(exitCode)
	}
}

func setupLib() {
	if p := os.Getenv("ENERGY_LIB"); fileExists(p) {
		libname.LibName = p
		return
	}
	home, _ := os.UserHomeDir()
	for _, p := range []string{
		"libenergy-amd64.dll",
		filepath.Join(home, ".energy", "runtime", "libenergy-amd64.dll"),
		"E:/app/workspace/lib/libenergy-x86_64-win64.dll",
	} {
		if fileExists(p) {
			libname.LibName = p
			return
		}
	}
}

func fileExists(p string) bool {
	if p == "" {
		return false
	}
	_, err := os.Stat(p)
	return err == nil
}

func (m *TShowcaseForm) FormCreate(sender lcl.IObject) {
	m.SetCaption("单画布 · 复杂动态图形 / 半透明 / 异形 / 圆角 渲染示例")
	m.SetPosition(types.PoScreenCenter)
	m.SetWidth(1440)
	m.SetHeight(880)
	// The scene owns its offscreen frame buffer; a second LCL-level double
	// buffer on the form only adds another full-window copy per frame.
	m.SetDoubleBuffered(false)
	m.SetColor(colors.RGBToColor(14, 16, 22))
	m.SetKeyPreview(true)

	s := NewScene(m)
	m.scene = s
	// WG_PROFILE=1 dumps a per-widget timing breakdown at the end of --bench.
	// WG_BLANK=1 skips all widgets, measuring the irreducible clear+blit cost.
	s.profile = os.Getenv("WG_PROFILE") != "" || benchMode || schedTestMode
	s.skipDraw = os.Getenv("WG_BLANK") != ""
	// WG_STATIC=0 disables the baked static layer (A/B comparison, and it
	// restores the original per-frame drifting glow blobs).
	s.useStatic = os.Getenv("WG_STATIC") != "0"
	// WG_AA=0 restores the original all-GDI (hard-edged) renderer for A/B
	// comparison. AA needs the double buffer + baked static layer.
	s.aaMode = os.Getenv("WG_AA") != "0" && !disableDoubleBuffer && s.useStatic

	// 抑制 WM_ERASEBKGND：整帧离屏渲染要 30ms+，若每帧先由 LCL 把窗口擦成
	// 窗体底色，这 30ms 里屏幕上是空底，成品只在最后一次 Blit 后存在一瞬
	// —— 这是「一直闪」和大部分时间只看到深色空底的根因（A/B 实测 40 帧
	// 里 26 帧全空底）。「不彩色」的另一个根因是帧缓冲格式，见 scene.go。
	// 注意：TEngForm 挂了 OnWndProc 后完全接管消息，初始化阶段的
	// WM_ERASEBKGND 必须透传（此前无条件吞掉会让窗口创建卡死）。所以只在
	// 首帧成功绘制之后才接管；其余所有消息一律透传。
	m.SetOnWndProc(func(msg *types.TLMessage) {
		if msg.Msg == wmRequestFrame {
			// A frame slot arrived on the display's 60Hz boundary. Only the
			// main thread renders; Release inflight afterwards so the pump
			// never queues catch-up frames.
			m.InheritedWndProc(msg)
			s.RenderFrame()
			s.sched.FrameDone()
			return
		}
		if !allowErase && msg.Msg == messages.WM_ERASEBKGND && s.Frame > 0 {
			msg.Result = 1 // 已处理：不擦背景
			s.EraseSuppressed++
			return
		}
		if (msg.Msg == messages.WM_SIZE || msg.Msg == messages.WM_SIZING) && s.Frame > 0 {
			// Still transparent to LCL; only stamp the debounce timer.
			s.resizeAt = time.Now()
		}
		m.InheritedWndProc(msg)
	})

	// The HWND only exists once the form has been created, so bind and start
	// the frame pump from the first message-loop turn.
	lcl.RunOnMainThreadAsync(func(id uint32) {
		s.sched.hwnd = uintptr(m.Handle())
		// Pre-bake the static snapshot BEFORE starting the 60Hz pump. Its
		// one-off ~90ms cost would otherwise steal ~5 frame slots at startup
		// and show up as a visible hitch / dropped frames.
		if cr := s.host.ClientRect(); cr.Width() > 1 && cr.Height() > 1 {
			s.relayout(cr.Width(), cr.Height())
			s.ensureStaticBuffer(cr)
		}
		s.sched.Start()
		s.sched.SetActive(!s.Paused)
		// The frame-rate window starts HERE, not at form creation: otherwise
		// window creation + the one-off static bake drag the first sample down.
		s.fpsWindow = time.Now()
		s.fpsFrames = 0
		s.FPS = 0
		s.RequestFrame()
	})

	// Decorative irregular translucent blobs (behind everything).
	s.Add(NewGlowBlob("光斑A", rect(50, 90, 450, 430), 5, 0.3))
	s.Add(NewGlowBlob("光斑B", rect(920, 400, 1360, 820), 7, 1.9))

	s.Add(NewHeader(rect(24, 16, 1416, 82)))

	// Glass cards: same size/opacity/theme share ONE baked RGBA layer.
	s.Add(NewGlassCard("卡片-波形", "动态波形 · 180 点", rect(24, 96, 470, 340), 16))
	s.Add(NewGlassCard("卡片-仪表", "环形仪表 · 旋转星形", rect(486, 96, 950, 340), 16))
	s.Add(NewGlassCard("卡片-性能", "渲染性能实测", rect(966, 96, 1416, 300), 16))
	s.Add(NewGlassCard("卡片-折线", "实时折线 · 3 条曲线", rect(24, 356, 470, 604), 16))
	s.Add(NewGlassCard("卡片-柱状", "动态柱状图", rect(486, 356, 950, 604), 16))
	s.Add(NewGlassCard("卡片-控件", "", rect(966, 316, 1416, 604), 16))
	s.Add(NewGlassCard("卡片-花瓣", "异形花瓣 · 贝塞尔", rect(24, 620, 470, 828), 16))
	s.Add(NewGlassCard("卡片-饼图", "环形饼图", rect(486, 620, 950, 828), 16))
	s.Add(NewGlassCard("卡片-说明", "架构与操作", rect(966, 620, 1416, 828), 16))

	// Content widgets (drawn on top of the cards).
	s.Add(NewWaveChart("波形", rect(38, 140, 456, 328)))
	s.Add(NewRingGauge("仪表", rect(500, 140, 936, 328)))
	s.Add(NewPerfPanel("性能", rect(980, 140, 1402, 288)))
	s.Add(NewLineChart("折线", rect(38, 400, 456, 592)))
	s.Add(NewBarChart("柱状", rect(500, 400, 936, 592)))
	s.Add(NewPetalBlob("花瓣", rect(38, 664, 456, 816)))
	s.Add(NewDonutChart("饼图", rect(500, 664, 936, 816)))
	s.Add(NewLegend("说明", rect(980, 664, 1402, 816)))
	s.Add(NewControls("控件", rect(966, 316, 1416, 604)))

	m.SetOnKeyDown(func(sender lcl.IObject, key *uint16, shift types.TShiftState) {
		if key != nil {
			s.OnKey(*key)
		}
	})
	s.Invalidate()

	if benchMode {
		go func() {
			// Start the 3s window when the frame pump actually starts, so the
			// one-off window creation / static bake is not counted as "frames".
			for !s.sched.Active() {
				time.Sleep(5 * time.Millisecond)
			}
			time.Sleep(3 * time.Second)
			fmt.Println("==== BENCH (3s 真实窗口, 帧泵启动后计时) ====")
			for _, line := range s.StatusLines() {
				fmt.Println(line)
			}
			fmt.Print(s.ProfileReport())
			requestExit(0)
		}()
	}
	if mouseTestMode {
		startMouseTest(s)
	}
	if schedTestMode {
		startSchedTest(s)
	}
	if resizeTestMode {
		startResizeTest(s)
	}
	if screenshotPath != "" {
		startScreenshot(s, screenshotPath)
	}
}

// runSelfTest verifies the RGBA bulk-write path without opening a window.
func runSelfTest() int {
	fmt.Println("[SELFTEST] runtime library:", libname.LibName)
	lcl.Init()
	l := NewAlphaLayer(64, 64)
	defer l.Free()
	if len(l.surf.Px) == 0 {
		fmt.Println("FAIL: PixelData() unavailable")
		return 1
	}
	ri, gi, bi, ai := l.surf.R, l.surf.G, l.surf.B, l.surf.A
	l.Bake(func(s *Surface) {
		for y := 0; y < s.H; y++ {
			for x := 0; x < s.W; x++ {
				o := s.offset(x, y)
				s.Px[o+s.R], s.Px[o+s.G], s.Px[o+s.B], s.Px[o+s.A] = 0x40, 0x80, 0xC0, 0xFF
			}
		}
	})
	got := l.img.Colors(10, 10)
	want := lcl.TFPColor{Red: 0x4040, Green: 0x8080, Blue: 0xC0C0, Alpha: 0xFFFF}
	ok := got.Red>>8 == want.Red>>8 && got.Green>>8 == want.Green>>8 &&
		got.Blue>>8 == want.Blue>>8 && got.Alpha>>8 == want.Alpha>>8
	fmt.Printf("[SELFTEST] channel order R=%d G=%d B=%d A=%d  stride=%d\n", ri, gi, bi, ai, l.surf.Stride)
	fmt.Printf("[SELFTEST] bulk-write readback %+v -> %v\n", got, passLabel(ok))
	fmt.Printf("[SELFTEST] bitmap valid: %v\n", l.bmp.IsValid())
	if corner := roundedAlpha(0, 0, 64, 64, 12); corner > 0.01 {
		fmt.Printf("FAIL: roundedAlpha corner = %.2f, want ~0\n", corner)
		return 1
	}
	if center := roundedAlpha(32, 32, 64, 64, 12); center != 1 {
		fmt.Printf("FAIL: roundedAlpha center = %.2f, want 1\n", center)
		return 1
	}
	fmt.Println("[SELFTEST] roundedAlpha corner/center: PASS")
	if !ok || !l.bmp.IsValid() {
		return 1
	}
	if !runRegressionChecks() {
		return 1
	}
	fmt.Println("RESULT: PASS")
	return 0
}

// runRegressionChecks guards the three rendering bugs found in the field.
// All of them are headless: no form, no window, no WM_PAINT needed.
func runRegressionChecks() bool {
	allOK := true
	check := func(name string, cond bool, detail string) {
		label := "PASS"
		if !cond {
			label = "FAIL"
			allOK = false
		}
		fmt.Printf("[SELFTEST] %-34s %s  %s\n", name, label, detail)
	}

	back := newBackBuffer(64, 64)
	defer back.Free()

	// (1) The frame buffer must have NO alpha channel. GDI writes RGB only, so
	// a Pf32bit buffer leaves A=0 and LCL alpha-blends the whole frame away.
	pf := back.PixelFormat()
	img := back.CreateIntfImage()
	bpp := img.DataDescription().BitsPerPixel()
	img.Free()
	check("back buffer pixel format", pf == types.Pf24bit,
		fmt.Sprintf("pf=%d bitsPerPixel=%d (want Pf24bit/24)", pf, bpp))

	// (2) LCL's TBrush.SetColor resets Style to bsSolid. Painter.brush MUST
	// write the style last, otherwise TextOut paints an opaque colour block.
	th := darkTheme()
	p := &Painter{Canvas: back.Canvas(), Theme: &th}
	p.Reset()
	white := colors.RGBToColor(240, 245, 255)
	p.brush(white, types.BsClear)
	br := back.Canvas().BrushToBrush()
	check("Painter.brush keeps BsClear", br.Style() == types.BsClear,
		fmt.Sprintf("style=%d (want BsClear=%d)", br.Style(), types.BsClear))

	// (3) A solid fill through Painter must actually land the requested colour
	// in the offscreen buffer (guards the SetColor/SetStyle ordering too).
	red := colors.RGBToColor(230, 40, 60)
	p.brush(red, types.BsSolid)
	br = back.Canvas().BrushToBrush()
	check("Painter.brush keeps BsSolid+colour",
		br.Style() == types.BsSolid && br.Color() == red,
		fmt.Sprintf("style=%d color=0x%06x", br.Style(), uint32(br.Color())))
	p.FillRect(types.TRect{Left: 0, Top: 0, Right: 64, Bottom: 64}, red)
	img = back.CreateIntfImage()
	got := img.Colors(32, 32)
	img.Free()
	wantOK := got.Red>>8 == 0xE6 && got.Green>>8 == 0x28 && got.Blue>>8 == 0x3C
	check("offscreen GDI fill colour", wantOK,
		fmt.Sprintf("R=%02x G=%02x B=%02x (want e6/28/3c)",
			got.Red>>8, got.Green>>8, got.Blue>>8))
	return allOK
}

func passLabel(ok bool) string {
	if ok {
		return "PASS"
	}
	return "FAIL"
}
