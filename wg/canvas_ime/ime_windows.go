//go:build windows

package main

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Windows IME constants (imm32.h / winuser.h).
const (
	gcsCompStr       = 0x0008
	gcsCursorPos     = 0x0080
	cfsPoint         = 0x0002
	cfsForcePosition = 0x0020
	cfsCandidatePos  = 0x0040
)

type wPoint struct{ X, Y int32 }
type wRect struct{ Left, Top, Right, Bottom int32 }

// COMPOSITIONFORM: DWORD dwStyle; POINT ptCurrentPos; RECT rcArea;
type compositionForm struct {
	DwStyle      uint32
	PtCurrentPos wPoint
	RcArea       wRect
}

// CANDIDATEFORM: DWORD dwIndex; DWORD dwStyle; POINT ptCurrentPos; RECT rcArea;
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

// imeComposition reads the pre-edit (composition) string and cursor position.
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

// imeResult reads the committed string (GCS_RESULTSTR). Only used for the
// debug panel: actual insertion goes through UTF8KeyPress to avoid duplicates.
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

// imeSelfTest verifies that every Win32 IME entry point used by this demo
// resolves at runtime, without opening a window or requiring typed input.
func imeSelfTest() string {
	type entry struct {
		name string
		proc *windows.LazyProc
	}
	entries := []entry{
		{"ImmGetContext", procImmGetContext},
		{"ImmReleaseContext", procImmReleaseContext},
		{"ImmGetCompositionStringW", procImmGetCompositionStringW},
		{"ImmSetCompositionWindow", procImmSetCompositionWindow},
		{"ImmSetCandidateWindow", procImmSetCandidateWindow},
		{"ClientToScreen", procClientToScreen},
	}
	for _, e := range entries {
		if err := e.proc.Find(); err != nil {
			return "FAIL " + e.name + ": " + err.Error()
		}
	}
	return "OK: 6 imm32/user32 IME procedures resolved"
}

// libSelfTest verifies that the loaded runtime library exports the exact
// symbols this architecture depends on (canvas + WndProc passthrough + focus).
func libSelfTest(libPath string) string {
	if libPath == "" {
		return "FAIL: no runtime library configured"
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

// imePosition puts the composition window and candidate window at the caret.
//
// IMPORTANT (Microsoft docs, COMPOSITIONFORM / CANDIDATEFORM):
//
//	ptCurrentPos is relative to the upper-left corner of the WINDOW that
//	contains the composition/candidate window, i.e. CLIENT coordinates of
//	hwnd -- NOT screen coordinates. Converting with ClientToScreen was the
//	original bug: the IME windows were shifted by the form's screen origin.
//
// x/y must already be hwnd client coordinates (Scene converts canvas
// coordinates with TControl.ClientToParent before calling this).
// ClientToScreen here is used ONLY for the diagnostic string.
func imePosition(hwnd uintptr, x, y int32) string {
	if hwnd == 0 {
		return "WAIT: no HWND yet"
	}
	himc := imeContext(hwnd)
	if himc == 0 {
		return "WAIT: no IME context (window not focused)"
	}
	defer imeRelease(hwnd, himc)

	pt := wPoint{X: x, Y: y}
	cf := compositionForm{
		DwStyle:      cfsPoint | cfsForcePosition,
		PtCurrentPos: pt,
		RcArea:       wRect{Left: pt.X, Top: pt.Y, Right: pt.X + 200, Bottom: pt.Y + 28},
	}
	compOK, _, _ := procImmSetCompositionWindow.Call(himc, uintptr(unsafe.Pointer(&cf)))

	// Candidate list goes just below the caret line.
	candPos := wPoint{X: pt.X, Y: pt.Y + 20}
	cand := candidateForm{
		DwIndex:      0,
		DwStyle:      cfsCandidatePos,
		PtCurrentPos: candPos,
		RcArea:       wRect{Left: candPos.X, Top: candPos.Y, Right: candPos.X + 320, Bottom: candPos.Y + 160},
	}
	candOK, _, _ := procImmSetCandidateWindow.Call(himc, uintptr(unsafe.Pointer(&cand)))

	sp := wPoint{X: x, Y: y}
	procClientToScreen.Call(hwnd, uintptr(unsafe.Pointer(&sp)))
	return fmt.Sprintf("client(%d,%d) screen(%d,%d) comp=%d cand=%d",
		x, y, sp.X, sp.Y, compOK, candOK)
}
