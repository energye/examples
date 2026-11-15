//go:build windows

package main

import (
	"fmt"
	"os"
	"time"

	"github.com/energye/lcl/lcl"
)

// runKeySelfTest is an end-to-end regression test for the keyboard layer.
// It opens the REAL demo form (same event wiring as the app), injects real
// Win32 key messages with PostMessage, and asserts the self-drawn editor's
// state. It specifically proves that control characters arriving through
// UTF8KeyPress (Backspace 0x08, Ctrl+A 0x01, Tab 0x09, Enter 0x0D, ...)
// no longer corrupt the text and that Backspace/Delete/Home actually edit.
//
// Run: go run ./wg/canvas_ime --keyselftest

const (
	wmKeyDown = 0x0100
	wmKeyUp   = 0x0101
	wmChar    = 0x0102
	wmClose   = 0x0010
)

var (
	postMessageW    = usr32.NewProc("PostMessageW")
	keyTestLines    []string
	keyTestFailed   bool
	keyTestFiltered int
)

func runKeySelfTest() {
	lcl.Init()
	lcl.Application.Initialize()
	lcl.Application.NewForm(&mainForm)
	go injectAndAssert()
	lcl.Application.Run()

	fmt.Println("==== KEYSELFTEST RESULT ====")
	for _, l := range keyTestLines {
		fmt.Println(l)
	}
	fmt.Printf("过滤控制字符数: %d (应 >= 6)\n", keyTestFiltered)
	if keyTestFiltered < 6 {
		keyTestFailed = true
		fmt.Println("FAIL: control chars were not filtered at the event boundary")
	}
	if keyTestFailed {
		fmt.Println("RESULT: FAIL")
		os.Exit(1)
	}
	fmt.Println("RESULT: PASS")
}

func injectAndAssert() {
	time.Sleep(700 * time.Millisecond)
	h := uintptr(mainForm.Handle())

	post := func(msg uint32, w uintptr) {
		postMessageW.Call(h, uintptr(msg), w, 0)
		time.Sleep(90 * time.Millisecond)
	}
	key := func(vk uint16) {
		post(wmKeyDown, uintptr(vk))
		post(wmKeyUp, uintptr(vk))
	}
	check := func(name, want string) {
		got := demoInput.Text()
		ok := got == want
		if !ok {
			keyTestFailed = true
		}
		keyTestLines = append(keyTestLines,
			fmt.Sprintf("%-22s got=%q want=%q  %s", name, got, want, passLabel(ok)))
	}

	// Start from a known buffer with the caret at the end.
	demoInput.SetText("abcd")
	check("初始值", "abcd")

	// Backspace: KeyDown deletes 'd'; WM_CHAR 0x08 must NOT insert a control byte.
	post(wmKeyDown, vkBack)
	post(wmChar, 0x08)
	post(wmKeyUp, vkBack)
	check("Backspace", "abc")

	// Home moves the caret to 0, Delete removes 'b'. Delete produces no WM_CHAR.
	key(vkHome)
	key(vkDelete)
	check("Home + Delete", "bc")

	// Every one of these is a real WM_CHAR that LCL forwards to UTF8KeyPress.
	// None may change the buffer.
	for _, c := range []uintptr{0x01 /*Ctrl+A*/, 0x03 /*Ctrl+C*/, 0x08, /*Back*/
		0x09 /*Tab*/, 0x0D /*Enter*/, 0x16 /*Ctrl+V*/, 0x1A /*Ctrl+Z*/, 0x7F /*DEL*/} {
		post(wmChar, c)
	}
	check("控制字符全部过滤", "bc")

	// Arrow keys must not corrupt text either (no WM_CHAR for them, but the
	// KeyDown handler must stay in bounds at both edges).
	key(vkEnd)
	key(vkRight)
	key(vkDelete)
	key(vkHome)
	key(vkLeft)
	key(vkBack)
	check("方向键边界安全", "bc")

	keyTestFiltered = mainForm.scene.filtered
	time.Sleep(250 * time.Millisecond)
	postMessageW.Call(h, wmClose, 0, 0)
}

func passLabel(ok bool) string {
	if ok {
		return "PASS"
	}
	return "FAIL"
}
