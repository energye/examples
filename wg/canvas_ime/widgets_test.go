package main

import (
	"testing"

	"github.com/energye/lcl/types"
)

func newTestInput() *Input {
	return NewInput("test", types.TRect{Left: 0, Top: 0, Right: 200, Bottom: 40})
}

// Chinese / emoji must be handled by rune index, never by byte index.
func TestInputRuneCaret(t *testing.T) {
	in := newTestInput()
	in.SetText("a中b")
	if in.caret != 3 {
		t.Fatalf("caret after SetText = %d, want 3 runes", in.caret)
	}
	in.moveCaret(2, false)
	in.insert("😀")
	if in.text != "a中😀b" {
		t.Fatalf("insert produced %q, want %q", in.text, "a中😀b")
	}
	if in.caret != 3 {
		t.Fatalf("caret after insert = %d, want 3", in.caret)
	}
}

func TestInputDeleteAndSelection(t *testing.T) {
	in := newTestInput()
	in.SetText("abcd")
	in.moveCaret(2, false)

	in.backspace() // delete 'b'
	if in.text != "acd" || in.caret != 1 {
		t.Fatalf("backspace: text=%q caret=%d", in.text, in.caret)
	}
	in.deleteForward() // delete 'c'
	if in.text != "ad" {
		t.Fatalf("deleteForward: text=%q, want %q", in.text, "ad")
	}

	in.anchor, in.caret = 0, 2
	lo, hi := in.selectionRange()
	if lo != 0 || hi != 2 {
		t.Fatalf("selectionRange = (%d,%d), want (0,2)", lo, hi)
	}
	in.deleteSelection()
	if in.text != "" || in.caret != 0 {
		t.Fatalf("deleteSelection: text=%q caret=%d", in.text, in.caret)
	}
}

func TestInputShiftSelectionAndUndo(t *testing.T) {
	in := newTestInput()
	in.SetText("hello")
	in.moveCaret(2, false)
	in.moveCaret(4, true) // Shift+Right twice
	lo, hi := in.selectionRange()
	if lo != 2 || hi != 4 {
		t.Fatalf("shift selection = (%d,%d), want (2,4)", lo, hi)
	}
	in.insert("X") // replace selection
	if in.text != "heXo" {
		t.Fatalf("replace selection: %q, want %q", in.text, "heXo")
	}
	in.undo()
	if in.text != "hello" {
		t.Fatalf("undo: %q, want %q", in.text, "hello")
	}
}

func TestInputSubmitAndEscape(t *testing.T) {
	in := newTestInput()
	var submitted string
	in.SetOnSubmit(func(v string) { submitted = v })
	in.SetText("你好")
	in.onKey(13, types.NewSet()) // Enter
	if submitted != "你好" {
		t.Fatalf("Enter submitted %q", submitted)
	}
	in.anchor, in.caret = 0, 2
	in.ClearSelection()
	if in.hasSelection() {
		t.Fatal("Escape/ClearSelection should collapse the selection")
	}
}

// Regression for the real bug: LCL's UTF8KeyPress also delivers Backspace
// (0x08), Ctrl+A..Z (0x01..0x1A), Tab (0x09), Enter (0x0D) and DEL (0x7F).
// Inserting those bytes produced garbled text after Delete/Backspace and
// destroyed the result of every Ctrl shortcut.
func TestSanitizeText(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"\x08", ""},
		{"\x01", ""},
		{"\t\r\n", ""},
		{"\x7f", ""},
		{"a\x00b", "ab"},
		{"a\x08中", "a中"},
		{"你好😀abc", "你好😀abc"}, // committed IME text must survive
		{string([]byte{0xff, 0xfe}), ""},
		{"", ""},
	}
	for _, c := range cases {
		if got := sanitizeText(c.in); got != c.want {
			t.Errorf("sanitizeText(% X) = %q, want %q", []byte(c.in), got, c.want)
		}
	}
}

func TestControlCharsNeverEnterBuffer(t *testing.T) {
	in := newTestInput()
	in.SetText("abc")
	for _, s := range []string{"\x08", "\x01", "\x03", "\x09", "\x0d", "\x16", "\x1a", "\x7f"} {
		in.Handle(&InputEvent{Kind: EvText, Text: s})
	}
	if in.text != "abc" {
		t.Fatalf("control chars leaked into the buffer: %q", in.text)
	}
	// The KeyDown half still performs the real edit.
	in.Handle(&InputEvent{Kind: EvKeyDown, Key: vkBack, Shift: types.NewSet()})
	if in.text != "ab" {
		t.Fatalf("Backspace: %q, want %q", in.text, "ab")
	}
	in.Handle(&InputEvent{Kind: EvKeyDown, Key: vkHome, Shift: types.NewSet()})
	in.Handle(&InputEvent{Kind: EvKeyDown, Key: vkDelete, Shift: types.NewSet()})
	if in.text != "b" {
		t.Fatalf("Home+Delete: %q, want %q", in.text, "b")
	}
}

func TestCtrlShortcutsAndLowercaseVK(t *testing.T) {
	for _, vk := range []uint16{'A', 'a'} {
		in := newTestInput()
		in.SetText("hello")
		in.Handle(&InputEvent{Kind: EvKeyDown, Key: vk, Shift: types.NewSet(types.SsCtrl)})
		lo, hi := in.selectionRange()
		if lo != 0 || hi != 5 {
			t.Fatalf("Ctrl+%c select-all = (%d,%d), want (0,5)", vk, lo, hi)
		}
	}
}

func TestUndoRestoresCaretPosition(t *testing.T) {
	in := newTestInput()
	in.SetText("abc")
	in.moveCaret(1, false)
	in.insert("X")
	if in.text != "aXbc" || in.caret != 2 {
		t.Fatalf("insert: %q caret=%d", in.text, in.caret)
	}
	in.Handle(&InputEvent{Kind: EvKeyDown, Key: 'Z', Shift: types.NewSet(types.SsCtrl)})
	if in.text != "abc" || in.caret != 1 {
		t.Fatalf("Ctrl+Z: %q caret=%d, want %q caret=1", in.text, in.caret, "abc")
	}
}

func TestWordMovementAndCtrlBackspace(t *testing.T) {
	in := newTestInput()
	in.SetText("hello world foo")
	ctrl := types.NewSet(types.SsCtrl)
	in.onKey(vkLeft, ctrl)
	if in.caret != 12 {
		t.Fatalf("Ctrl+Left caret=%d, want 12", in.caret)
	}
	in.onKey(vkLeft, ctrl)
	if in.caret != 6 {
		t.Fatalf("Ctrl+Left caret=%d, want 6", in.caret)
	}
	in.onKey(vkEnd, types.NewSet())
	in.onKey(vkBack, ctrl)
	if in.text != "hello world " {
		t.Fatalf("Ctrl+Backspace: %q", in.text)
	}
}
