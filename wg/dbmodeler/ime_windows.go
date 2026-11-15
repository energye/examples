//go:build windows

package main

// Win32 IME 绑定（参考 wg/canvas_ime）：纯自绘输入框没有原生编辑框，
// 预编辑串要自己从 IMM32 读出来画，候选框位置也要自己设置。
// COMPOSITIONFORM / CANDIDATEFORM 的 ptCurrentPos 是窗口客户区坐标，
// 不是屏幕坐标；Scene 已把画布坐标用 ClientToParent 转好。

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	gcsCompStr       = 0x0008
	gcsCursorPos     = 0x0080
	cfsPoint         = 0x0002
	cfsForcePosition = 0x0020
	cfsCandidatePos  = 0x0040
)

type wPoint struct{ X, Y int32 }
type wRect struct{ Left, Top, Right, Bottom int32 }

type compositionForm struct {
	DwStyle      uint32
	PtCurrentPos wPoint
	RcArea       wRect
}

type candidateForm struct {
	DwIndex      uint32
	DwStyle      uint32
	PtCurrentPos wPoint
	RcArea       wRect
}

var (
	imm32 = windows.NewLazyDLL("imm32.dll")
	usr32 = windows.NewLazyDLL("user32.dll")

	procImmGetContext            = imm32.NewProc("ImmGetContext")
	procImmReleaseContext        = imm32.NewProc("ImmReleaseContext")
	procImmGetCompositionStringW = imm32.NewProc("ImmGetCompositionStringW")
	procImmSetCompositionWindow  = imm32.NewProc("ImmSetCompositionWindow")
	procImmSetCandidateWindow    = imm32.NewProc("ImmSetCandidateWindow")
	procClientToScreen           = usr32.NewProc("ClientToScreen")
)

func imeContext(hwnd uintptr) uintptr {
	if hwnd == 0 {
		return 0
	}
	himc, _, _ := procImmGetContext.Call(hwnd)
	return himc
}

func imeRelease(hwnd, himc uintptr) {
	if hwnd != 0 && himc != 0 {
		procImmReleaseContext.Call(hwnd, himc)
	}
}

func imeComposition(hwnd uintptr) (string, int32, bool) {
	himc := imeContext(hwnd)
	if himc == 0 {
		return "", 0, false
	}
	defer imeRelease(hwnd, himc)
	n, _, _ := procImmGetCompositionStringW.Call(himc, gcsCompStr, 0, 0)
	var text string
	if n > 0 {
		buf := make([]uint16, int(n/2)+1)
		got, _, _ := procImmGetCompositionStringW.Call(
			himc, gcsCompStr, uintptr(unsafe.Pointer(&buf[0])), n)
		if got > 0 {
			text = windows.UTF16ToString(buf)
		}
	}
	cur, _, _ := procImmGetCompositionStringW.Call(himc, gcsCursorPos, 0, 0)
	return text, int32(cur), true
}

func imeResult(hwnd uintptr) string {
	himc := imeContext(hwnd)
	if himc == 0 {
		return ""
	}
	defer imeRelease(hwnd, himc)
	n, _, _ := procImmGetCompositionStringW.Call(himc, gcsResultStr, 0, 0)
	if n <= 0 {
		return ""
	}
	buf := make([]uint16, int(n/2)+1)
	got, _, _ := procImmGetCompositionStringW.Call(
		himc, gcsResultStr, uintptr(unsafe.Pointer(&buf[0])), n)
	if got <= 0 {
		return ""
	}
	return windows.UTF16ToString(buf)
}

func imePosition(hwnd uintptr, x, y int32) string {
	if hwnd == 0 {
		return "WAIT: no HWND"
	}
	himc := imeContext(hwnd)
	if himc == 0 {
		return "WAIT: no IME context"
	}
	defer imeRelease(hwnd, himc)

	pt := wPoint{X: x, Y: y}
	cf := compositionForm{
		DwStyle:      cfsPoint | cfsForcePosition,
		PtCurrentPos: pt,
		RcArea:       wRect{Left: pt.X, Top: pt.Y, Right: pt.X + 200, Bottom: pt.Y + 28},
	}
	compOK, _, _ := procImmSetCompositionWindow.Call(himc, uintptr(unsafe.Pointer(&cf)))
	candPos := wPoint{X: pt.X, Y: pt.Y + 20}
	cand := candidateForm{
		DwIndex:      0,
		DwStyle:      cfsCandidatePos,
		PtCurrentPos: candPos,
		RcArea: wRect{Left: candPos.X, Top: candPos.Y,
			Right: candPos.X + 320, Bottom: candPos.Y + 160},
	}
	candOK, _, _ := procImmSetCandidateWindow.Call(himc, uintptr(unsafe.Pointer(&cand)))
	return fmt.Sprintf("client(%d,%d) comp=%d cand=%d", x, y, compOK, candOK)
}

func imeSelfTest() string {
	for _, e := range []struct {
		name string
		proc *windows.LazyProc
	}{
		{"ImmGetContext", procImmGetContext},
		{"ImmReleaseContext", procImmReleaseContext},
		{"ImmGetCompositionStringW", procImmGetCompositionStringW},
		{"ImmSetCompositionWindow", procImmSetCompositionWindow},
		{"ImmSetCandidateWindow", procImmSetCandidateWindow},
		{"ClientToScreen", procClientToScreen},
	} {
		if err := e.proc.Find(); err != nil {
			return "FAIL " + e.name + ": " + err.Error()
		}
	}
	return "OK: 6 imm32/user32 IME procedures resolved"
}

func libSelfTest(libPath string) string {
	if libPath == "" {
		return "SKIP: runtime library path not configured"
	}
	dll := windows.NewLazyDLL(libPath)
	for _, sym := range []string{
		"TCustomGraphicControl_Create",
		"TCustomGraphicControl_OnPaint",
		"TCustomGraphicControl_OnMouseDown",
		"TEngForm_InheritedWndProc",
		"TEngForm_OnWndProc",
		"TWinControl_SetFocus",
	} {
		if err := dll.NewProc(sym).Find(); err != nil {
			return "FAIL missing export " + sym + ": " + err.Error()
		}
	}
	return "OK: 6 runtime library exports resolved"
}
