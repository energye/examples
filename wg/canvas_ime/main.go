// Single-canvas rendering demo + real IME test.
//
// Architecture (per the user's wording):
//
//	"只有一个画布，然后在上面画各种渲染基础能力和组件的功能实现。
//	 输入具有焦点的也都使用它实现。不使用有句柄的也就是纯自绘的任意组件。"
//
// The window (Form) owns the only HWND; one handle-less ICustomGraphicControl
// is the canvas; every widget below is a pure Go object drawn on that canvas.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/energye/lcl/api/libname"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
	"github.com/energye/lcl/types/messages"
)

type TMainForm struct {
	lcl.TEngForm
	scene *Scene
}

var (
	mainForm  TMainForm
	demoInput *Input // kept for the automated --keyselftest assertions
)

func main() {
	setupLib()

	// Non-GUI self test: verifies DLL exports and IME API bindings.
	if len(os.Args) > 1 && os.Args[1] == "--selftest" {
		fmt.Println("[SELFTEST] runtime library:", libname.LibName)
		fmt.Println("[SELFTEST] library exports:", libSelfTest(libname.LibName))
		fmt.Println("[SELFTEST] IME bindings:", imeSelfTest())
		fmt.Printf("[SELFTEST] GCS_RESULTSTR=0x%X\n", gcsResultStr)
		return
	}

	// Automated end-to-end key test: injects real Win32 messages and asserts
	// the self-drawn editor state. No user interaction needed.
	if len(os.Args) > 1 && os.Args[1] == "--keyselftest" {
		runKeySelfTest()
		return
	}

	lcl.Init()
	lcl.Application.Initialize()
	lcl.Application.SetMainFormOnTaskBar(true)
	lcl.Application.NewForm(&mainForm)
	lcl.Application.Run()
}

// setupLib makes the demo runnable out of the box: libname.LibName overrides
// the auto-detected runtime library path before lcl.Init().
func setupLib() {
	if p := os.Getenv("ENERGY_LIB"); fileExists(p) {
		libname.LibName = p
		return
	}
	home, _ := os.UserHomeDir()
	candidates := []string{
		"libenergy-amd64.dll",
		filepath.Join(home, ".energy", "runtime", "libenergy-amd64.dll"),
		"E:/app/workspace/lib/libenergy-x86_64-win64.dll",
	}
	for _, p := range candidates {
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

func (m *TMainForm) FormCreate(sender lcl.IObject) {
	m.SetCaption("单画布渲染示例 · 输入法(IME)测试")
	m.SetPosition(types.PoScreenCenter)
	m.SetWidth(960)
	m.SetHeight(640)
	m.SetDoubleBuffered(true)
	m.SetKeyPreview(true) // form receives keys even with other win controls around

	s := NewScene(m)
	m.scene = s

	input := NewInput("自绘输入框", types.TRect{Left: 40, Top: 118, Right: 620, Bottom: 166})
	demoInput = input
	input.SetOnSubmit(func(text string) {
		s.SetNotice("提交: " + text)
		fmt.Println("[UI] 提交:", text)
	})
	progress := NewProgressBar("渐变进度条", types.TRect{Left: 360, Top: 194, Right: 920, Bottom: 230})

	s.Add(NewLabel("标题", types.TRect{Left: 40, Top: 22, Right: 920, Bottom: 52},
		"只有一个画布，然后在上面画各种渲染基础能力和组件的功能实现", false))
	s.Add(NewLabel("副标题", types.TRect{Left: 40, Top: 58, Right: 920, Bottom: 88},
		"输入具有焦点的也都使用它实现；不使用有句柄的、纯自绘的任意组件", true))

	s.Add(input)
	s.Add(NewButton("模拟组合串", types.TRect{Left: 640, Top: 118, Right: 780, Bottom: 166},
		"模拟组合串", func() { s.SimulateComposition() }))
	s.Add(NewButton("切换主题", types.TRect{Left: 800, Top: 118, Right: 920, Bottom: 166},
		"切换主题", func() { s.ToggleTheme() }))
	s.Add(NewButton("清空输入", types.TRect{Left: 40, Top: 186, Right: 180, Bottom: 232},
		"清空输入", func() {
			input.SetText("")
			s.SetComposition("", 0)
			s.SetResult("")
			s.Invalidate()
		}))
	s.Add(NewButton("自绘按钮", types.TRect{Left: 200, Top: 186, Right: 340, Bottom: 232},
		"自绘按钮", func() {
			progress.value = 75
			fmt.Println("[UI] 自绘按钮点击 -> 进度条 75%")
			s.Invalidate()
		}))
	s.Add(progress)
	s.Add(NewStatusPanel("IME 状态面板", types.TRect{Left: 40, Top: 258, Right: 920, Bottom: 574}, s))
	s.Add(NewLabel("操作提示", types.TRect{Left: 40, Top: 588, Right: 920, Bottom: 616},
		"点击输入框 → 切换中文输入法 → 输入拼音 → 观察预编辑串 / 候选框 / 上屏字符；也可点「模拟组合串」自检渲染", true))

	m.bindInput(s)
	s.SetFocus(input)
	s.Invalidate()
}

// bindInput attaches all window-level input: WndProc (IME messages), UTF-8
// text and key events. Mouse and painting are bound by Scene on the canvas.
func (m *TMainForm) bindInput(s *Scene) {
	m.SetOnWndProc(func(msg *types.TLMessage) {
		hwnd := uintptr(m.Handle())
		reposition := false
		switch msg.Msg {
		case messages.WM_IME_STARTCOMPOSITION:
			s.LogIME("STARTCOMPOSITION")
			reposition = true
		case messages.WM_IME_COMPOSITION:
			text, cursor, ok := imeComposition(hwnd)
			if ok {
				s.SetComposition(text, cursor)
			}
			if uintptr(msg.LParam)&gcsResultStr != 0 {
				if r := imeResult(hwnd); r != "" {
					s.SetResult("GCS_RESULTSTR: " + r)
				}
			}
			s.LogIME(fmt.Sprintf("COMPOSITION lParam=0x%X comp=%q cursor=%d",
				uintptr(msg.LParam), text, cursor))
			reposition = true
		case messages.WM_IME_ENDCOMPOSITION:
			s.SetComposition("", 0)
			s.LogIME("ENDCOMPOSITION")
		case messages.WM_IME_NOTIFY:
			// High-frequency notification (candidate window / IME state).
			// Count it, but do not print or repaint on every single message.
			s.CountIMEMessage()
		case messages.WM_IME_SETCONTEXT:
			s.LogIME(fmt.Sprintf("IME msg 0x%X", msg.Msg))
			reposition = true
		case messages.WM_IME_CHAR:
			s.LogIME(fmt.Sprintf("IME msg 0x%X", msg.Msg))
		}
		// MANDATORY: TEngForm.WndProc hands every message to this callback, so
		// without the inherited call the window would stop painting/responding.
		m.InheritedWndProc(msg)

		// Position AFTER default processing: DefWindowProc / LCL can reset the
		// composition window position while handling STARTCOMPOSITION.
		if reposition {
			s.PositionIME()
		}
	})

	// UTF8KeyPress is NOT a "printable text" event. Windows TranslateMessage
	// also turns VK_BACK into 0x08, Ctrl+A into 0x01, Enter into 0x0D and Tab
	// into 0x09, and LCL forwards ALL of them here. Inserting those bytes was
	// the cause of the garbled text after Delete/Backspace and of broken
	// Ctrl+A/C/V/X/Z. Filter at the event boundary, before anything reaches
	// the editor. Verified with: go run ./wg/canvas_ime --keydiag
	m.SetOnUTF8KeyPress(func(sender lcl.IObject, text *string) {
		if text == nil {
			return
		}
		raw := *text
		clean := sanitizeText(raw)
		if clean != raw {
			*text = clean // also stop LCL from forwarding the control char
		}
		if clean == "" {
			s.LogKey(fmt.Sprintf("过滤控制字符 % X", []byte(raw)))
			return
		}
		s.LogKey(fmt.Sprintf("输入 %q", clean))
		s.OnText(clean)
		s.SetResult("UTF8KeyPress: " + clean)
	})

	m.SetOnKeyDown(func(sender lcl.IObject, key *uint16, shift types.TShiftState) {
		if key == nil {
			return
		}
		s.LogKey(fmt.Sprintf("按键 key=%d 0x%02X shift=0x%X", *key, *key, uint32(shift)))
		s.OnKey(*key, shift)
	})
}
