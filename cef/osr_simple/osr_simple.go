package main

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"

	vcef "github.com/energye/cef/147/cef"
	"github.com/energye/cef/base"
	"github.com/energye/cef/cef"
	cefTypes "github.com/energye/cef/cef/types"
	"github.com/energye/lcl/api"
	"github.com/energye/lcl/api/libname"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/rtl"
	"github.com/energye/lcl/tool"
	"github.com/energye/lcl/types"
)

type TForm struct {
	lcl.TEngForm
	chromium     cef.IChromium
	panel        vcef.IBufferPanel
	statusBar    lcl.IStatusBar
	canClose     bool
	creating     bool
	resizing     bool
	pendingResize bool
	retryCount   int
}

var log = func(m string, a ...interface{}) { fmt.Printf("[OSR] "+m+"\n", a...) }

func init() { libname.UseWS = "gtk3" }

func main() {
	runtime.LockOSThread()
	d := "/home/yanghy/.energy/chromium/linux_amd64_147.0.14"
	libname.LibName = filepath.Join(d, "libenergy-amd64-gtk3.so")
	lcl.Init()
	base.Init()
	a := cef.NewApplication()
	base.SetGlobalCEFApplication(a.Instance())
	if tool.IsExist(filepath.Join(d, "libcef.so")) {
		a.SetFrameworkDirPath(d)
		a.SetResourcesDirPath(d)
		a.SetLocalesDirPath(filepath.Join(d, "locales"))
	}
	a.SetWindowlessRenderingEnabled(true)
	a.SetExternalMessagePump(true)
	a.SetMultiThreadedMessageLoop(false)
	a.SetDisableZygote(true)
	a.SetRootCache(filepath.Join(os.TempDir(), "EC"))
	if a.ProcessType() == cefTypes.PtBrowser {
		s := cef.NewWorkScheduler(nil)
		base.SetGlobalCEFWorkSchedule(s.Instance())
		a.SetOnScheduleMessagePumpWork(func(d int64) { s.ScheduleMessagePumpWork(d) })
	} else {
		a.StartSubProcess()
		return
	}
	if !a.StartMainProcess() {
		return
	}
	api.SetOnReleaseCallback(func() {
		if tool.IsLinux() {
			api.WidgetSetFinalization()
		}
	})
	lcl.Application.Initialize()
	lcl.Application.SetMainFormOnTaskBar(true)
	f := &TForm{}
	lcl.Application.NewForms(f)
	lcl.Application.Run()
}

func mod(s types.TShiftState) cefTypes.TCefEventFlags {
	var f cefTypes.TCefEventFlags
	if s.In(types.SsShift) {
		f |= cefTypes.EVENTFLAG_SHIFT_DOWN
	}
	if s.In(types.SsAlt) {
		f |= cefTypes.EVENTFLAG_ALT_DOWN
	}
	if s.In(types.SsCtrl) {
		f |= cefTypes.EVENTFLAG_CONTROL_DOWN
	}
	return f
}
func cefBtn(b types.TMouseButton) cefTypes.TCefMouseButtonType {
	switch b {
	case types.MbRight:
		return cefTypes.MBT_RIGHT
	case types.MbMiddle:
		return cefTypes.MBT_MIDDLE
	default:
		return cefTypes.MBT_LEFT
	}
}

// ===== FormCreate =====
func (m *TForm) FormCreate(s lcl.IObject) {
	m.SetWidth(1024)
	m.SetHeight(768)
	m.SetCaption("CEF OSR")
	m.ScreenCenter()

	// 工具栏
	tp := lcl.NewPanel(m)
	tp.SetParent(m)
	tp.SetAlign(types.AlTop)
	tp.SetHeight(40)
	tp.SetBevelOuter(types.BvNone)
	ue := lcl.NewEdit(m)
	ue.SetParent(tp)
	ue.SetBounds(10, 8, 460, 24)
	ue.SetText("https://www.baidu.com")
	nb := lcl.NewButton(m)
	nb.SetParent(tp)
	nb.SetBounds(480, 6, 80, 28)
	nb.SetCaption("导航")

	// ===== TBufferPanel (Pascal 示例一致) =====
	m.panel = vcef.NewBufferPanel(m)
	m.panel.SetParent(m)
	m.panel.SetAlign(types.AlClient)
	m.panel.SetCopyOriginalBuffer(true) // ← Pascal 示例: CopyOriginalBuffer := True
	m.panel.SetTabStop(true)

	// 鼠标
	m.panel.SetOnMouseDown(func(_ lcl.IObject, btn types.TMouseButton, s types.TShiftState, x, y int32) {
		if m.chromium == nil || !m.chromium.Initialized() { return }
		if b := m.chromium.Browser(); b != nil { b.GetHost().SendMouseClickEvent(cef.TCefMouseEvent{X: x, Y: y, Modifiers: mod(s)}, cefBtn(btn), false, 1) }
	})
	m.panel.SetOnMouseUp(func(_ lcl.IObject, btn types.TMouseButton, s types.TShiftState, x, y int32) {
		if m.chromium == nil || !m.chromium.Initialized() { return }
		if b := m.chromium.Browser(); b != nil { b.GetHost().SendMouseClickEvent(cef.TCefMouseEvent{X: x, Y: y, Modifiers: mod(s)}, cefBtn(btn), true, 1) }
	})
	m.panel.SetOnMouseMove(func(_ lcl.IObject, s types.TShiftState, x, y int32) {
		if m.chromium != nil && m.chromium.Initialized() {
			if b := m.chromium.Browser(); b != nil {
				b.GetHost().SendMouseMoveEvent(cef.TCefMouseEvent{X: x, Y: y, Modifiers: mod(s)}, false)
			}
		}
	})
	m.panel.SetOnMouseWheel(func(_ lcl.IObject, s types.TShiftState, d int32, p types.TPoint, h *bool) {
		if m.chromium != nil && m.chromium.Initialized() {
			if b := m.chromium.Browser(); b != nil {
				b.GetHost().SendMouseWheelEvent(cef.TCefMouseEvent{X: p.X, Y: p.Y, Modifiers: mod(s)}, 0, d)
				*h = true
			}
		}
	})
	// 键盘: Panel 本身 + 窗口 KeyPreview 双保险
	m.panel.SetOnKeyDown(func(_ lcl.IObject, k *types.Char, s types.TShiftState) {
		if m.chromium != nil && m.chromium.Initialized() {
			if b := m.chromium.Browser(); b != nil {
				b.GetHost().SendKeyEvent(cef.TCefKeyEvent{Kind: cefTypes.KEYEVENT_RAWKEYDOWN, Modifiers: mod(s), WindowsKeyCode: int32(*k), NativeKeyCode: int32(*k)})
			}
		}
	})
	m.panel.SetOnKeyUp(func(_ lcl.IObject, k *types.Char, s types.TShiftState) {
		if m.chromium != nil && m.chromium.Initialized() {
			if b := m.chromium.Browser(); b != nil {
				b.GetHost().SendKeyEvent(cef.TCefKeyEvent{Kind: cefTypes.KEYEVENT_KEYUP, Modifiers: mod(s), WindowsKeyCode: int32(*k), NativeKeyCode: int32(*k)})
			}
		}
	})
	m.panel.SetOnKeyPress(func(_ lcl.IObject, k *types.Char) {
		if m.chromium != nil && m.chromium.Initialized() {
			if b := m.chromium.Browser(); b != nil {
				b.GetHost().SendKeyEvent(cef.TCefKeyEvent{Kind: cefTypes.KEYEVENT_CHAR, WindowsKeyCode: int32(*k), NativeKeyCode: int32(*k), Character: uint16(*k), UnmodifiedCharacter: uint16(*k)})
			}
		}
	})
	m.SetKeyPreview(true)
	m.SetOnKeyDown(func(_ lcl.IObject, k *types.Char, s types.TShiftState) {
		if m.chromium != nil && m.chromium.Initialized() {
			if b := m.chromium.Browser(); b != nil {
				b.GetHost().SendKeyEvent(cef.TCefKeyEvent{Kind: cefTypes.KEYEVENT_RAWKEYDOWN, Modifiers: mod(s), WindowsKeyCode: int32(*k)})
			}
		}
	})

	// 焦点
	m.panel.SetOnEnter(func(_ lcl.IObject) {
		if m.chromium != nil && m.chromium.Initialized() {
			if b := m.chromium.Browser(); b != nil {
				b.GetHost().SetFocus(true)
			}
		}
	})
	m.panel.SetOnExit(func(_ lcl.IObject) {
		if m.chromium != nil && m.chromium.Initialized() {
			if b := m.chromium.Browser(); b != nil {
				b.GetHost().SetFocus(false)
			}
		}
	})

	m.statusBar = lcl.NewStatusBar(m)
	m.statusBar.SetParent(m)
	m.statusBar.SetSimpleText("初始化...")

	// ===== Chromium =====
	m.chromium = cef.NewChromium(m)
	m.chromium.SetDefaultUrl("https://www.baidu.com")
	m.chromium.Options().SetWindowlessFrameRate(30)

	m.chromium.SetOnGetViewRect(func(_ lcl.IObject, _ cef.ICefBrowser, r *cef.TCefRect) {
		r.X, r.Y = 0, 0
		r.Width = m.panel.Width()
		r.Height = m.panel.Height()
	})
	m.chromium.SetOnGetScreenInfo(func(_ lcl.IObject, _ cef.ICefBrowser, i *cef.TCefScreenInfo, ok *bool) {
		w, h := m.panel.Width(), m.panel.Height()
		i.DeviceScaleFactor = 1.0
		i.Depth = 24
		i.Rect = cef.TCefRect{Width: w, Height: h}
		i.AvailableRect = i.Rect
		*ok = true
	})
	m.chromium.SetOnGetScreenPoint(func(_ lcl.IObject, _ cef.ICefBrowser, vx, vy int32, sx, sy *int32, ok *bool) {
		p := lcl.AsControl(m.panel.Instance()).ClientToScreenWithPoint(types.TPoint{X: vx, Y: vy})
		*sx, *sy = p.X, p.Y
		*ok = true
	})

	// ===== OnPaint (TBufferPanel 方式) =====
	m.chromium.SetOnPaint(func(_ lcl.IObject, _ cef.ICefBrowser, _ cefTypes.TCefPaintElementType, _ cefTypes.NativeUInt, dr cef.ICefRectArray, buf uintptr, w, h int32) {
		if !m.panel.BeginBufferDraw() {
			return
		}
		m.panel.UpdateBufferDimensions(w, h)
		m.panel.BufferIsResized(false)
		bmp := m.panel.Buffer()
		if bmp != nil && bmp.IsValid() {
			bmp.BeginUpdate(false)
			bw, bh := m.panel.BufferWidth(), m.panel.BufferHeight()
			ps, ss := 4, int(w)*4
			for i := 0; i < int(dr.Count()); i++ {
				r := dr.Get(i)
				if r.X < 0 || r.Y < 0 {
					continue
				}
				lb := int(math.Min(float64(r.Width), float64(bw-r.X))) * ps
				if lb <= 0 {
					continue
				}
				so := int((r.Y*w)+r.X) * ps
				do := int(r.X) * ps
				sp := uintptr(buf) + uintptr(so)
				ro := int(math.Min(float64(r.Height), float64(bh-r.Y)))
				for ri := 0; ri < ro; ri++ {
					rtl.Move(sp, uintptr(bmp.ScanLine(r.Y+int32(ri)))+uintptr(do), lb)
					sp += uintptr(ss)
				}
			}
			bmp.EndUpdate(false)
		}
		m.panel.EndBufferDraw()
		m.resizing = false
		if m.pendingResize { m.pendingResize = false; if b := m.chromium.Browser(); b != nil { b.GetHost().WasResized() } }
		if m.HandleAllocated() { m.panel.InvalidatePanel() }
	})

	// ===== 事件 =====
	m.chromium.SetOnAfterCreated(func(_ lcl.IObject, b cef.ICefBrowser) {
		log("[After] ID:%d", b.GetIdentifier())
		lcl.RunOnMainThreadAsync(func(uint32) {
			h := b.GetHost()
			h.SetFocus(true)
			h.WasResized()
			h.Invalidate(cefTypes.PET_VIEW)
			m.statusBar.SetSimpleText("浏览器已创建")
		})
	})
	m.chromium.SetOnTitleChange(func(_ lcl.IObject, _ cef.ICefBrowser, t string) {
		lcl.RunOnMainThreadAsync(func(uint32) { m.SetCaption("OSR - " + t) })
	})
	m.chromium.SetOnLoadStart(func(_ lcl.IObject, _ cef.ICefBrowser, _ cef.ICefFrame, _ cefTypes.TCefTransitionType) {
		m.statusBar.SetSimpleText("加载中...")
	})
	m.chromium.SetOnLoadEnd(func(_ lcl.IObject, _ cef.ICefBrowser, _ cef.ICefFrame, _ int32) {
		m.statusBar.SetSimpleText("加载完成")
	})
	m.chromium.SetOnCursorChange(func(_ lcl.IObject, _ cef.ICefBrowser, cur cefTypes.TCefCursorHandle, ct cefTypes.TCefCursorType, _ cef.TCefCursorInfo, r *bool) {
		m.panel.SetCursor(cef.MiscFunc.CefCursorToWindowsCursor(ct))
		*r = true
	})
	m.panel.SetOnResize(func(_ lcl.IObject) {
	 if m.chromium == nil || !m.chromium.Initialized() { return }
	 if m.resizing { m.pendingResize = true; return }
	 m.resizing = true
	 if m.panel.BufferIsResized(false) {
	  if b := m.chromium.Browser(); b != nil { b.GetHost().Invalidate(cefTypes.PET_VIEW) }
	 } else {
	  if b := m.chromium.Browser(); b != nil { b.GetHost().WasResized() }
	 }
	})

	nb.SetOnClick(func(lcl.IObject) {
		if u := ue.Text(); u != "" {
			m.chromium.LoadURLWithStrFrame(u, m.chromium.Browser().GetMainFrame())
		}
	})

	// 创建浏览器（延迟确保窗口就绪）
	lcl.RunOnMainThreadAsync(func(uint32) { m.createBrowser() })
}

func (m *TForm) createBrowser() {
	if m.chromium == nil || m.chromium.Initialized() || m.creating {
		return
	}
	m.creating = true
	m.retryCount = 0
	log("[Create] ...")
	t := lcl.NewTimer(m)
	t.SetInterval(100) // 100ms 间隔，减轻负担
	t.SetOnTimer(func(lcl.IObject) {
		if m.chromium.Initialized() {
			t.SetEnabled(false); t.Free(); m.creating = false
			return
		}
		m.retryCount++
		if m.retryCount > 50 { // 5秒超时
			t.SetEnabled(false); t.Free(); m.creating = false
			log("[Create] timeout")
			return
		}
		if m.chromium.CreateBrowserWithWControlStrRContextDValue(nil, "", nil, nil) || m.chromium.Initialized() {
			t.SetEnabled(false); t.Free(); m.creating = false
		}
	})
	t.SetEnabled(true)
}
