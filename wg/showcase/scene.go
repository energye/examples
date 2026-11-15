package main

import (
	"fmt"
	"math"
	"time"
	"unsafe"

	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
)

type EventKind uint8

const (
	EvMouseDown EventKind = iota
	EvMouseMove
	EvMouseUp
	EvMouseLeave
	EvKeyDown
)

type UIEvent struct {
	Kind  EventKind
	X, Y  int32
	Key   uint16
	Shift types.TShiftState
}

// Widget: a pure Go object. No LCL instance, no HWND.
type Widget interface {
	Name() string
	Rect() types.TRect
	Hit(x, y int32) bool
	Paint(p *Painter, s *Scene)
	Handle(e *UIEvent, s *Scene) bool
}

type baseWidget struct {
	name    string
	rect    types.TRect
	hitAble bool
}

func (b *baseWidget) Name() string          { return b.name }
func (b *baseWidget) Rect() types.TRect     { return b.rect }
func (b *baseWidget) SetRect(r types.TRect) { b.rect = r }
func (b *baseWidget) Hit(x, y int32) bool {
	if !b.hitAble {
		return false
	}
	return x >= b.rect.Left && x <= b.rect.Right && y >= b.rect.Top && y <= b.rect.Bottom
}
func (b *baseWidget) Paint(p *Painter, s *Scene)       {}
func (b *baseWidget) Handle(e *UIEvent, s *Scene) bool { return false }

// A widget may opt into the static layer. Static widgets are rasterised ONCE
// into an opaque full-window bitmap; every frame starts by copying that bitmap
// and only re-paints dynamic widgets on top. This is the single biggest win:
// glass cards + glow blobs are per-pixel alpha blends costing ~24ms/frame.
type staticWidget interface {
	IsStatic() bool
}

func isStatic(w Widget) bool {
	sw, ok := w.(staticWidget)
	return ok && sw.IsStatic()
}

// Scene owns the single canvas, the frame scheduler and the render stats.
type Scene struct {
	form    *TShowcaseForm
	host    lcl.ICustomGraphicControl
	painter *Painter
	theme   Theme
	cache   *LayerCache

	// Offscreen back buffer: the whole frame is drawn into this bitmap and
	// blitted to the screen once. Drawing directly onto the form DC makes the
	// 18ms incremental paint visible as flicker.
	backBuf        lcl.IBitmap
	backW, backH   int32
	doubleBuffered bool

	// Static layer: an opaque Pf24bit snapshot of everything that never
	// changes between frames (background, glow blobs, glass cards, legend).
	// Rebuilt only on theme / opacity / size / cache invalidation.
	staticBuf    lcl.IBitmap
	staticW      int32
	staticH      int32
	staticDirty  bool
	staticBakeUs int64
	useStatic    bool

	// Antialiasing backend. frameImg owns the raw 24bpp pixel block that the
	// software rasteriser writes into; staticPx is a one-off Go copy of the
	// GDI-baked static snapshot used as every frame's background.
	aaMode      bool
	frameImg    lcl.ILazIntfImage
	framePix    *FramePixels
	staticPx    []byte
	staticGen   int
	capturedGen int
	aaW, aaH    int32
	aaBakeUs    int64

	widgets        []Widget
	designRects    []types.TRect // design-space rects, parallel to widgets
	layout         Layout
	staticWidgets  []Widget
	dynamicWidgets []Widget
	hover          Widget
	capture        Widget

	// Frame pump: posts WM_APP frames to the main thread at an absolute 60Hz
	// wall-clock cadence, or exactly one frame for an idle dirty region.
	sched       *FrameScheduler
	started     time.Time
	pausedAt    time.Time
	pausedTotal time.Duration
	lastPaint   time.Time
	Phase       float64
	Frame       int
	Paused      bool
	Opacity     float64

	FPS     float64
	PaintUs int64
	Calls   int
	// EraseSuppressed counts WM_ERASEBKGND messages skipped after the first
	// frame. Exposed in the perf panel as evidence that no erase can expose
	// an empty background between two blitted frames.
	EraseSuppressed int64
	fpsFrames       int
	fpsWindow       time.Time
	dirty           types.TRect
	hasDirty        bool
	forceFull       bool // static snapshot changed -> next idle frame must be full
	fullFrame       bool // the frame currently being painted is a full repaint
	lastKey         string

	// Resize debounce. Re-baking 9 RGBA glass cards costs ~100ms per size;
	// doing that for every intermediate WM_SIZE while the user drags a window
	// edge starves the message loop and looks like a hang. During the cooling
	// window we keep the old snapshot and draw a light background instead.
	resizeAt   time.Time
	holdStatic bool

	// Optional per-frame profiling (WG_PROFILE=1). Timings are accumulated in
	// microseconds and dumped by --bench, so the cost of a whole frame can be
	// attributed to a specific widget / phase instead of guessed.
	profile  bool
	skipDraw bool
	prof     frameProfile

	// Status panel text is rebuilt 4x/second, not 60x/second. Keeping the
	// strings and the slice alive removes ~7 allocations per frame.
	statusLines []string
	statusAt    time.Time
}

func NewScene(form *TShowcaseForm) *Scene {
	s := &Scene{
		form:        form,
		theme:       darkTheme(),
		cache:       NewLayerCache(),
		started:     time.Now(),
		pausedAt:    time.Now(),
		Opacity:     0.82,
		staticDirty: true,
		useStatic:   true,
		forceFull:   true,
	}
	s.painter = &Painter{
		Theme:     &s.theme,
		textCmds:  make([]textCmd, 0, 64),
		bezierPts: make([]types.TPoint, 0, 64),
		Scale:     1,
	}

	// The ONLY LCL drawing object in the window. It has no HWND.
	host := lcl.NewCustomGraphicControl(form)
	host.SetAlign(types.AlClient)
	host.SetParent(form)
	host.SetParentBackground(false)
	host.SetControlStyle(host.ControlStyle().Include(types.CsOpaque))
	s.host = host
	s.bindCanvas()

	// No LCL TTimer any more. Frames are delivered by FrameScheduler as a
	// private WM_APP message on an absolute 60Hz wall-clock timeline; input
	// events only mark dirty state, they never schedule a paint themselves.
	s.sched = NewFrameScheduler(0) // hwnd is bound once the window exists
	// A 1ms system timer granularity keeps the Go-side scheduler accurate.
	setTimerResolution(1)
	s.fpsWindow = time.Now()
	return s
}

// PhaseNow derives the animation phase from wall-clock time, so a dropped
// timer tick can never freeze or rewind the animation.
func (s *Scene) PhaseNow() float64 {
	end := time.Now()
	if s.Paused {
		end = s.pausedAt
	}
	return end.Sub(s.started).Seconds() - s.pausedTotal.Seconds()
}

func (s *Scene) SetPaused(v bool) {
	if v == s.Paused {
		return
	}
	now := time.Now()
	if v {
		s.pausedAt = now
	} else {
		s.pausedTotal += now.Sub(s.pausedAt)
	}
	s.Paused = v
	// Restart the FPS window on every animation start/stop so idle time is
	// never averaged into a frame-rate number.
	s.fpsWindow = time.Now()
	s.fpsFrames = 0
	// Turning animation off STOPS the frame pump (zero CPU while idle).
	// Turning it on restarts the absolute 60Hz cadence. One extra frame is
	// requested so the "已暂停" label updates immediately.
	s.sched.SetActive(!v)
	s.RequestFrame()
}

// Input no longer paints. A WM_MOUSEMOVE / key / hover change only mutates
// widget state and (when idle) marks a dirty rectangle; FrameScheduler decides
// when the next frame boundary is. This is why moving the mouse can no longer
// push the FPS above 60 - the frame rate belongs to the display clock alone.
// Phase still comes from the wall clock (PhaseNow), so even a starved message
// queue can never freeze or rewind the animation.

// WG_FAST=1 disables the offscreen buffer: higher FPS, but the incremental
// GDI paint becomes visible as flicker on some systems. Kept as a switch so
// both behaviours can be compared on the same machine.
var disableDoubleBuffer bool

// newBackBuffer allocates the offscreen frame buffer. Package-level so the
// headless --selftest can assert the exact pixel format.
func newBackBuffer(w, h int32) lcl.IBitmap {
	b := lcl.NewBitmap()
	// 关键：离屏帧缓冲必须是不带 alpha 通道的格式。GDI 的 FillRect/Polyline/
	// TextOut 只写 RGB，Pf32bit DIB 的第 4 字节恒为 0；LCL 把这种位图 Draw
	// 到屏幕 DC 时按 AC_SRC_ALPHA 混合，所有矢量图元（曲线、柱、饼、文字）
	// 就全部被当成全透明而消失，只有自带 alpha 的玻璃图层还在 —— 实测截图
	// 表现为「内容全是暗灰色、不彩色」。Pf24bit 无 alpha 通道，最后一次
	// Blit 是纯拷贝，整帧才完整可见。
	b.SetPixelFormat(types.Pf24bit)
	b.SetSize(w, h)
	return b
}

// ensureBackBuffer creates (or resizes) the offscreen frame buffer.
func (s *Scene) ensureBackBuffer(w, h int32) {
	if disableDoubleBuffer {
		if s.backBuf != nil {
			s.backBuf.Free()
			s.backBuf = nil
		}
		s.doubleBuffered = false
		return
	}
	if w < 1 || h < 1 {
		return
	}
	if s.backBuf != nil && s.backW == w && s.backH == h {
		s.doubleBuffered = true
		return
	}
	if s.backBuf != nil {
		// frameImg was created FROM this bitmap (CreateIntfImage). Free the
		// Go view first: releasing the owner while the view still holds the
		// raw DIB pointer leaves framePix.px dangling, and the next paint
		// segfaults on a resize.
		s.freeAASurface()
		s.backBuf.Free()
	}
	b := newBackBuffer(w, h)
	bc := b.Canvas()
	s.doubleBuffered = bc != nil && bc.IsValid()
	if !s.doubleBuffered {
		b.Free()
		s.backBuf = nil
		s.backW, s.backH = 0, 0
		return
	}
	s.backBuf = b
	s.backW, s.backH = w, h
}

// ensureAASurface binds a writable Go view over the frame buffer's DIB.
// Called once per window size, never per frame.
func (s *Scene) ensureAASurface(w, h int32) {
	if !s.aaMode || w < 1 || h < 1 {
		return
	}
	if s.frameImg != nil && s.aaW == w && s.aaH == h {
		return
	}
	s.freeAASurface()
	if s.backBuf == nil {
		s.aaMode = false
		return
	}
	img := s.backBuf.CreateIntfImage()
	if img == nil || !img.IsValid() {
		s.aaMode = false
		return
	}
	fp := &FramePixels{}
	if !fp.bind(img, w, h) {
		img.Free()
		s.aaMode = false
		return
	}
	s.frameImg, s.framePix = img, fp
	s.aaW, s.aaH = w, h
}

func (s *Scene) freeAASurface() {
	if s.framePix != nil {
		// Detach the raw slice from the DIB before the owner is freed. A stale
		// view must never be able to write into released memory.
		s.framePix.px = nil
		s.framePix.img = nil
	}
	if s.frameImg != nil {
		s.frameImg.Free()
	}
	s.frameImg, s.framePix = nil, nil
	s.aaW, s.aaH = 0, 0
	s.capturedGen = -1
}

// captureStatic reads the GDI-painted snapshot into Go memory ONCE per bake.
// CreateIntfImage().PixelData() is a bulk read (1.58ms for 1440x880); the
// per-pixel Colors() FFI path would take seconds.
func (s *Scene) captureStatic() {
	if !s.aaMode || s.staticBuf == nil || s.capturedGen == s.staticGen {
		return
	}
	// During the resize debounce the snapshot still has the OLD size while
	// the client rect is already the NEW one. Reading `stride * r.Height()`
	// bytes from that old DIB overruns it - --resizetest caught a real
	// `fatal error: fault` in memmove here. Use the SNAPSHOT's own size and
	// refuse to capture unless it matches the current frame buffer.
	if s.staticW != s.backW || s.staticH != s.backH || s.staticW < 1 || s.staticH < 1 {
		return
	}
	img := s.staticBuf.CreateIntfImage()
	if img == nil || !img.IsValid() {
		return
	}
	defer img.Free()
	ptr := uintptr(img.PixelData())
	if ptr == 0 {
		return
	}
	desc := img.DataDescription()
	bpp := int(desc.BitsPerPixel()) / 8
	if bpp < 3 {
		bpp = 3
	}
	s0 := uintptr(img.GetDataLineStart(0))
	s1 := uintptr(img.GetDataLineStart(1))
	stride := int(s.staticW) * bpp
	if s1 > s0 {
		stride = int(s1 - s0)
	}
	need := stride * int(s.staticH)
	if need <= 0 || need > 64<<20 {
		return
	}
	px := unsafe.Slice((*byte)(unsafe.Pointer(ptr)), need)
	t := time.Now()
	if len(s.staticPx) != need {
		s.staticPx = make([]byte, need)
	}
	copy(s.staticPx, px)
	s.aaBakeUs = time.Since(t).Microseconds()
	s.capturedGen = s.staticGen
}

// RenderFrame is invoked ONLY by the main thread, ONLY when FrameScheduler
// posts a 60Hz frame slot (or a single idle dirty frame). Repaint() runs
// OnPaint synchronously: PREPARE offscreen, then PRESENT atomically.
func (s *Scene) RenderFrame() { s.host.Repaint() }

func (s *Scene) bindCanvas() {
	s.host.SetOnPaint(func(sender lcl.IObject) {
		c := s.host.Canvas()
		if c == nil || !c.IsValid() {
			return
		}
		t0 := time.Now()
		// ==== PREPARE (inside the 16.67ms frame budget) ====
		// All geometry, text and layers are computed and rasterised into ONE
		// offscreen surface. Nothing is incrementally drawn on the visible DC,
		// so the user never sees a half-built frame.
		s.Phase = s.PhaseNow()
		r := s.host.ClientRect()
		// A WM_SIZE changes the client rect BEFORE this paint. Re-map all
		// design rects first, then resize/rebuild the buffers, then re-bake
		// the static snapshot. Order matters: baking with stale rects was one
		// of the resize failure modes.
		s.relayout(r.Width(), r.Height())
		s.ensureBackBuffer(r.Width(), r.Height())
		if s.aaMode {
			s.ensureAASurface(r.Width(), r.Height())
		}
		// While the window is being resized, keep the previous snapshot and
		// skip the ~100ms RGBA re-bake for every intermediate size.
		const resizeCooldown = 180 * time.Millisecond
		s.holdStatic = !s.resizeAt.IsZero() && time.Since(s.resizeAt) < resizeCooldown
		if s.useStatic {
			if !s.holdStatic {
				s.ensureStaticBuffer(r)
			}
			// captureStatic is cheap (~1-2ms) and self-gates on the bake
			// generation; keep it so AA never falls back just because a
			// resize happened right after a bake.
			if s.aaMode {
				s.captureStatic()
			}
		}
		staticSized := s.staticBuf != nil && s.staticW == r.Width() && s.staticH == r.Height()
		// The AA backend needs a snapshot that MATCHES the frame buffer. A
		// stale / differently sized one would make copyFrom read out of bounds.
		aa := s.aaMode && s.framePix != nil && s.doubleBuffered && s.useStatic &&
			staticSized && len(s.staticPx) == len(s.framePix.px) && s.backBuf != nil
		s.painter.frame = nil
		if aa {
			s.painter.frame = s.framePix
		}

		// Two frame sources: an active animation repaints the whole frame at
		// every 60Hz boundary; an idle dirty region repaints ONLY that region.
		full := s.sched.Active() || s.forceFull || !s.hasDirty ||
			!s.doubleBuffered || !s.useStatic || !staticSized
		s.fullFrame = full
		presentRect := r
		if !full {
			presentRect = clampRect(s.dirty, r)
		}

		s.painter.Theme = &s.theme
		s.painter.Reset()
		target := c
		if s.doubleBuffered {
			target = s.backBuf.Canvas()
		}
		s.painter.Canvas = target
		prof := s.profile
		if prof {
			s.prof.setupUs += measure(t0)
			s.prof.ensure(len(s.frameWidgets()))
		}

		var drawCalls int
		if aa {
			// ==== PREPARE: AA geometry into opaque Go pixels ====
			tbg := time.Now()
			fp := s.framePix
			if full {
				fp.clearClip()
				fp.copyFrom(s.staticPx, nil)
			} else {
				fp.setClip(presentRect)
				fp.copyFrom(s.staticPx, &presentRect)
			}
			if prof {
				s.prof.clearUs += measure(tbg)
			}

			frame := s.frameWidgets()
			tg := time.Now()
			if !s.skipDraw {
				for i, w := range frame {
					// A dirty frame still visits EVERY widget: the clip rejects
					// out-of-region geometry, but all text commands must be
					// replayed because LoadFromIntfImage overwrites the whole
					// HDC (including the previous frame's text).
					tw := time.Now()
					w.Paint(s.painter, s)
					if prof {
						s.prof.widgetUs[i] += measure(tw)
					}
				}
			}
			if prof {
				s.prof.rasterUs += measure(tg)
			}

			// Bulk Go pixels -> bitmap HDC (1.09ms for a full 1440x880 window).
			tl := time.Now()
			s.frameImg.BeginUpdate()
			s.frameImg.EndUpdate()
			s.backBuf.LoadFromIntfImage(s.frameImg)
			if prof {
				s.prof.loadUs += measure(tl)
			}

			// Text stays GDI (ClearType), replayed AFTER the pixel upload.
			s.painter.FlushText()

			// ==== PRESENT: opaque BitBlt, never alpha compositing ====
			tp := time.Now()
			if full {
				c.CopyRectWithRectX2Canvas(r, target, r)
			} else {
				c.CopyRectWithRectX2Canvas(presentRect, target, presentRect)
			}
			drawCalls = s.painter.Calls + 1
			if prof {
				s.prof.blitUs += measure(tp)
			}
		} else {
			// ==== legacy GDI path: WG_AA=0 / WG_FAST=1 / WG_STATIC=0 ====
			tbg := time.Now()
			if full {
				s.beginFrame(r)
			} else {
				target.CopyRectWithRectX2Canvas(presentRect, s.staticBuf.Canvas(), presentRect)
				s.painter.Calls++
			}
			if prof {
				s.prof.clearUs += measure(tbg)
			}
			frame := s.frameWidgets()
			if !s.skipDraw {
				for i, w := range frame {
					if !full && !rectIntersects(w.Rect(), presentRect) {
						continue
					}
					tw := time.Now()
					w.Paint(s.painter, s)
					if prof {
						s.prof.widgetUs[i] += measure(tw)
					}
				}
			}
			drawCalls = s.painter.Calls
			tp := time.Now()
			if s.doubleBuffered {
				s.painter.Canvas = c
				if full {
					c.DrawWithIntX2Graphic(r.Left, r.Top, s.backBuf)
				} else {
					c.CopyRectWithRectX2Canvas(presentRect, s.backBuf.Canvas(), presentRect)
				}
				drawCalls++
			}
			if prof {
				s.prof.blitUs += measure(tp)
			}
		}

		s.PaintUs = time.Since(t0).Microseconds()
		if prof {
			s.prof.frames++
			s.prof.sample(time.Now(), s.PaintUs)
			if full {
				s.prof.fullFrames++
			} else {
				s.prof.dirtyFrames++
			}
		}
		s.Calls = drawCalls
		s.hasDirty = false
		s.forceFull = false
		s.lastPaint = time.Now()

		// FPS is measured on ACTUAL ANIMATION paints with a 1s window +
		// light smoothing. Idle one-shot dirty frames are intentionally NOT
		// counted: a frame rate is meaningless when no frames are scheduled.
		s.Frame++
		if s.sched.Active() {
			s.fpsFrames++
			now := time.Now()
			if now.Sub(s.fpsWindow) >= time.Second {
				inst := float64(s.fpsFrames) / now.Sub(s.fpsWindow).Seconds()
				if s.FPS <= 0 {
					s.FPS = inst
				} else {
					s.FPS = s.FPS*0.35 + inst*0.65
				}
				s.fpsFrames = 0
				s.fpsWindow = now
			}
		}
	})

	s.host.SetOnMouseDown(func(sender lcl.IObject, button types.TMouseButton, shift types.TShiftState, x, y int32) {
		s.dispatch(&UIEvent{Kind: EvMouseDown, X: x, Y: y, Shift: shift})
	})
	s.host.SetOnMouseMove(func(sender lcl.IObject, shift types.TShiftState, x, y int32) {
		s.dispatch(&UIEvent{Kind: EvMouseMove, X: x, Y: y, Shift: shift})
	})
	s.host.SetOnMouseUp(func(sender lcl.IObject, button types.TMouseButton, shift types.TShiftState, x, y int32) {
		s.dispatch(&UIEvent{Kind: EvMouseUp, X: x, Y: y, Shift: shift})
	})
	s.host.SetOnMouseLeave(func(sender lcl.IObject) {
		if s.hover != nil {
			w := s.hover
			w.Handle(&UIEvent{Kind: EvMouseLeave}, s)
			s.hover = nil
			if !s.sched.Active() {
				s.InvalidateRect(w.Rect())
			}
		}
	})
}

// RequestFrame is the ONLY entry point for "state changed, please repaint".
// While animating it is a no-op (the 60Hz pump owns the frame clock); while
// idle it asks for exactly ONE frame. This is the invalidation-driven model
// used by Flutter (scheduleFrame) / Chromium (BeginFrame) / WPF
// (InvalidateVisual) / Android (invalidate + Choreographer).
func (s *Scene) RequestFrame() { s.sched.RequestFrame() }

// Invalidate marks the WHOLE frame dirty. Kept for global state changes
// (theme, opacity, key state); it never posts a low-priority WM_PAINT.
func (s *Scene) Invalidate() {
	s.forceFull = true
	s.RequestFrame()
}

func (s *Scene) InvalidateRect(r types.TRect) {
	if !s.hasDirty {
		s.dirty, s.hasDirty = r, true
	} else {
		if r.Left < s.dirty.Left {
			s.dirty.Left = r.Left
		}
		if r.Top < s.dirty.Top {
			s.dirty.Top = r.Top
		}
		if r.Right > s.dirty.Right {
			s.dirty.Right = r.Right
		}
		if r.Bottom > s.dirty.Bottom {
			s.dirty.Bottom = r.Bottom
		}
	}
	s.RequestFrame()
}

func rectIntersects(a, b types.TRect) bool {
	return a.Left < b.Right && b.Left < a.Right && a.Top < b.Bottom && b.Top < a.Bottom
}

func clampRect(a, b types.TRect) types.TRect {
	if a.Left < b.Left {
		a.Left = b.Left
	}
	if a.Top < b.Top {
		a.Top = b.Top
	}
	if a.Right > b.Right {
		a.Right = b.Right
	}
	if a.Bottom > b.Bottom {
		a.Bottom = b.Bottom
	}
	if a.Right <= a.Left {
		a.Right = a.Left + 1
	}
	if a.Bottom <= a.Top {
		a.Bottom = a.Top + 1
	}
	return a
}

func (s *Scene) Add(w Widget) {
	s.widgets = append(s.widgets, w)
	s.designRects = append(s.designRects, w.Rect())
	if isStatic(w) {
		s.staticWidgets = append(s.staticWidgets, w)
	} else {
		s.dynamicWidgets = append(s.dynamicWidgets, w)
	}
}

// Scale returns the adaptive layout scale. Widgets use it for fixed-pixel
// line spacing and label offsets that cannot be expressed as a TRect.
func (s *Scene) Scale() float64 {
	if s.layout.scale <= 0 {
		return 1
	}
	return s.layout.scale
}

// S scales a fixed-pixel design constant (line spacing, label width, inset).
// Zero allocation: widgets call it directly in the 60Hz hot path.
func (s *Scene) S(v int32) int32 {
	if s.layout.scale == 1 || s.layout.scale <= 0 {
		return v
	}
	return int32(math.Round(float64(v) * s.layout.scale))
}

// relayout maps every design-space rect into the current client area. It is
// triggered once per size change, before any static re-bake, so resize cannot
// leave content clipped or pinned to the top-left corner.
func (s *Scene) relayout(cw, ch int32) {
	if !s.layout.update(cw, ch) {
		return
	}
	s.painter.Scale = s.layout.scale
	s.painter.ClearMetrics()
	for i, w := range s.widgets {
		if sr, ok := w.(interface{ SetRect(types.TRect) }); ok {
			sr.SetRect(s.layout.rect(s.designRects[i]))
		}
	}
	s.staticDirty = true
	s.forceFull = true
	s.Invalidate()
}

// frameWidgets are the widgets repainted every frame. With the static layer
// enabled this is only the dynamic subset; WG_STATIC=0 restores full repaint.
func (s *Scene) frameWidgets() []Widget {
	if s.useStatic {
		return s.dynamicWidgets
	}
	return s.widgets
}

func (s *Scene) hit(x, y int32) Widget {
	for i := len(s.widgets) - 1; i >= 0; i-- {
		if s.widgets[i].Hit(x, y) {
			return s.widgets[i]
		}
	}
	return nil
}

func (s *Scene) dispatch(e *UIEvent) {
	var affected []types.TRect
	changed := false
	switch e.Kind {
	case EvMouseDown:
		w := s.hit(e.X, e.Y)
		s.capture = w
		if w != nil {
			if w.Handle(e, s) {
				changed = true
			}
			affected = append(affected, w.Rect())
		}
	case EvMouseMove:
		w := s.hit(e.X, e.Y)
		if w != s.hover {
			if s.hover != nil {
				s.hover.Handle(&UIEvent{Kind: EvMouseLeave}, s)
				affected = append(affected, s.hover.Rect())
				changed = true
			}
			s.hover = w
			if w != nil {
				affected = append(affected, w.Rect())
				changed = true
			}
		}
		if s.capture != nil {
			if s.capture.Handle(e, s) {
				changed = true
			}
			affected = append(affected, s.capture.Rect())
		} else if w != nil && w.Handle(e, s) {
			changed = true
			affected = append(affected, w.Rect())
		}
	case EvMouseUp:
		w := s.capture
		if w == nil {
			w = s.hit(e.X, e.Y)
		}
		s.capture = nil
		if w != nil {
			if w.Handle(e, s) {
				changed = true
			}
			affected = append(affected, w.Rect())
		}
	}

	// Input NEVER schedules a frame while the 60Hz animation is live - this
	// is exactly what used to push FPS to 126 during mouse movement. State is
	// simply recorded and the next display boundary paints it.
	if s.sched.Active() {
		return
	}
	if !changed && len(affected) == 0 {
		return // pure pointer movement over static content: ZERO frames
	}
	for _, r := range affected {
		s.InvalidateRect(r)
	}
	if len(affected) == 0 {
		s.RequestFrame()
	}
}

func (s *Scene) OnKey(key uint16) {
	switch key {
	case 32: // Space
		s.SetPaused(!s.Paused)
		s.lastKey = "Space 暂停/继续"
	case 84, 116: // T / t
		s.ToggleTheme()
		s.lastKey = "T 切换主题"
	case 80, 112: // P / p
		s.SetPaused(!s.Paused)
		s.lastKey = "P 暂停/继续"
	case 187, 107: // +
		s.SetOpacity(s.Opacity + 0.08)
		s.lastKey = "+ 提高不透明度"
	case 189, 109: // -
		s.SetOpacity(s.Opacity - 0.08)
		s.lastKey = "- 降低不透明度"
	default:
		s.lastKey = fmt.Sprintf("Key %d", key)
	}
	s.Invalidate()
}

func (s *Scene) SetOpacity(v float64) {
	if v < 0.12 {
		v = 0.12
	}
	if v > 1 {
		v = 1
	}
	if v == s.Opacity {
		return
	}
	s.Opacity = v
	// Opacity is baked into the layer alpha, so all layers AND the static
	// snapshot must be rebuilt.
	s.cache.InvalidateAll()
	s.InvalidateStatic()
}

func (s *Scene) ToggleTheme() {
	if s.theme.Bg == darkTheme().Bg {
		s.theme = lightTheme()
	} else {
		s.theme = darkTheme()
	}
	s.painter.Theme = &s.theme
	// Keep the native window color in sync so a forced erase never flashes a
	// different color (e.g. dark form behind a light theme).
	s.form.SetColor(s.theme.Bg)
	s.cache.InvalidateAll()
	s.InvalidateStatic()
	s.InvalidateRect(s.host.ClientRect())
}

func (s *Scene) ThemeName() string {
	if s.theme.Bg == darkTheme().Bg {
		return "dark"
	}
	return "light"
}

// StatusLines feeds the performance panel with REAL measured numbers.
func (s *Scene) StatusLines() []string {
	if len(s.statusLines) > 0 && time.Since(s.statusAt) < 250*time.Millisecond {
		return s.statusLines
	}
	s.statusAt = time.Now()
	hits, misses, layers, bakeUs := s.cache.Stats()
	state := "运行中 60Hz"
	if s.Paused {
		state = "已暂停 零帧"
	}
	staticOn, staticBake, dyn := s.staticInfo()
	posted, skipped, lost := s.sched.Stats()
	fps := s.FPS
	if !s.sched.Active() {
		fps = 0
	}
	mode := "脏区"
	if s.fullFrame {
		mode = "整帧"
	}
	s.statusLines = append(s.statusLines[:0],
		fmt.Sprintf("FPS %.1f   状态 %s   帧序 %d   相位 %.1fs", fps, state, s.Frame, s.Phase),
		fmt.Sprintf("PREPARE %d us   PRESENT 原子提交   画布调用 %d 次", s.PaintUs, s.Calls),
		fmt.Sprintf("帧泵 60Hz 投递 %d / 跳过 %d / 丢弃 %d   本帧 %s", posted, skipped, lost, mode),
		fmt.Sprintf("静态层烘焙 %s (一次 %d us)   每帧动态组件 %d   AA %s", boolLabel(staticOn), staticBake, dyn, boolLabel(s.aaMode && s.framePix != nil)),
		fmt.Sprintf("图层缓存 命中 %d / 未命中 %d   常驻 %d 层", hits, misses, layers),
		fmt.Sprintf("RGBA 烘焙 %d us (逐像素 SetColors 约 100x)", bakeUs),
		fmt.Sprintf("全局不透明度 %.2f   主题 %s   最近按键 %s", s.Opacity, s.ThemeName(), s.lastKey),
	)
	return s.statusLines
}

func boolLabel(v bool) string {
	if v {
		return "开"
	}
	return "关"
}

func (s *Scene) Free() {
	s.sched.Stop()
	if s.backBuf != nil {
		s.backBuf.Free()
		s.backBuf = nil
	}
	s.freeAASurface()
	s.freeStaticBuffer()
	s.cache.Free()
}
