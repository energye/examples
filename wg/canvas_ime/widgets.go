package main

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
)

// Windows virtual-key codes used by the editor. LCL's KeyDown event passes
// the VK code (verified: --keydiag reports Delete=46, Back=8, A=65).
const (
	vkBack   = 8
	vkTab    = 9
	vkReturn = 13
	vkEscape = 27
	vkEnd    = 35
	vkHome   = 36
	vkLeft   = 37
	vkRight  = 39
	vkInsert = 45
	vkDelete = 46
)

// sanitizeText removes the control characters that LCL's UTF8KeyPress also
// delivers: Backspace 0x08, Ctrl+A..Z 0x01..0x1A, Tab 0x09, Enter 0x0D,
// DEL 0x7F, C1 0x80..0x9F, invalid UTF-8 and shortstring NUL padding.
// Real committed IME text (Chinese, emoji) is untouched because those runes
// are >= 0xA0. This is the single gate every text input must pass.
func sanitizeText(s string) string {
	if s == "" {
		return ""
	}
	var b strings.Builder
	for _, r := range s {
		if r == utf8.RuneError || !utf8.ValidRune(r) {
			continue
		}
		if r < 0x20 || r == 0x7F || (r >= 0x80 && r <= 0x9F) {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

type baseWidget struct {
	name      string
	rect      types.TRect
	focusable bool
}

func (b *baseWidget) Name() string              { return b.name }
func (b *baseWidget) Rect() types.TRect         { return b.rect }
func (b *baseWidget) SetRect(r types.TRect)     { b.rect = r }
func (b *baseWidget) Focusable() bool           { return b.focusable }
func (b *baseWidget) OnFocusGain()              {}
func (b *baseWidget) OnFocusLost()              {}
func (b *baseWidget) Paint(p *Painter)          {}
func (b *baseWidget) Handle(e *InputEvent) bool { return false }
func (b *baseWidget) Hit(x, y int32) bool {
	return x >= b.rect.Left && x <= b.rect.Right && y >= b.rect.Top && y <= b.rect.Bottom
}

func inset(r types.TRect, d int32) types.TRect {
	return types.TRect{Left: r.Left + d, Top: r.Top + d, Right: r.Right - d, Bottom: r.Bottom - d}
}

// ---- Label ----

type Label struct {
	baseWidget
	text  string
	muted bool
}

func NewLabel(name string, r types.TRect, text string, muted bool) *Label {
	return &Label{baseWidget: baseWidget{name: name, rect: r}, text: text, muted: muted}
}

func (l *Label) Paint(p *Painter) {
	c := p.Theme.Text
	if l.muted {
		c = p.Theme.Muted
	}
	p.Text(l.rect, l.text, c, true)
}

// ---- Button ----

type Button struct {
	baseWidget
	text    string
	hover   bool
	down    bool
	focused bool
	onClick func()
}

func NewButton(name string, r types.TRect, text string, onClick func()) *Button {
	return &Button{baseWidget: baseWidget{name: name, rect: r, focusable: true}, text: text, onClick: onClick}
}

func (b *Button) OnFocusGain() { b.focused = true }
func (b *Button) OnFocusLost() { b.focused = false }

func (b *Button) Paint(p *Painter) {
	bg := p.Theme.Accent
	if b.down {
		bg = p.Theme.AccentDown
	} else if b.hover {
		bg = p.Theme.PanelAlt
	}
	border := p.Theme.Edge
	if b.focused {
		border = p.Theme.Text
	}
	p.Shadow(b.rect, 8)
	p.RoundRect(b.rect, 8, bg, border)
	p.Text(b.rect, b.text, p.Theme.Text, false)
}

func (b *Button) Handle(e *InputEvent) bool {
	switch e.Kind {
	case EvMouseDown:
		b.down = true
	case EvMouseMove:
		b.hover = true
	case EvMouseLeave:
		b.hover = false
	case EvMouseUp:
		wasDown := b.down
		b.down = false
		if wasDown && b.Hit(e.X, e.Y) && b.onClick != nil {
			b.onClick()
		}
	case EvKeyDown:
		if (e.Key == 13 || e.Key == 32) && b.onClick != nil { // Enter / Space
			b.onClick()
			return true
		}
	}
	return true
}

// ---- Input: composition + caret + selection + standard edit keys, all self drawn ----

// editSnapshot restores both the text and the editing position on undo.
type editSnapshot struct {
	text          string
	caret, anchor int
}

type Input struct {
	baseWidget
	text        string
	composition string
	compCursor  int32
	focused     bool
	caretX      int32
	caretY      int32

	caret    int // caret position as a rune index into text
	anchor   int // selection anchor; anchor != caret means a selection
	history  []editSnapshot
	onSubmit func(string)
}

func NewInput(name string, r types.TRect) *Input {
	return &Input{baseWidget: baseWidget{name: name, rect: r, focusable: true}}
}

func (i *Input) Text() string                { return i.text }
func (i *Input) SetOnSubmit(fn func(string)) { i.onSubmit = fn }
func (i *Input) SetText(v string) {
	i.text = v
	n := len([]rune(v))
	i.caret, i.anchor = n, n
}
func (i *Input) OnFocusGain() { i.focused = true }
func (i *Input) OnFocusLost() { i.focused = false }

func (i *Input) runes() []rune { return []rune(i.text) }

func (i *Input) clamp(n int) {
	if i.caret < 0 {
		i.caret = 0
	}
	if i.caret > n {
		i.caret = n
	}
	if i.anchor < 0 {
		i.anchor = 0
	}
	if i.anchor > n {
		i.anchor = n
	}
}

func (i *Input) hasSelection() bool { return i.anchor != i.caret }

func (i *Input) selectionRange() (int, int) {
	if i.anchor < i.caret {
		return i.anchor, i.caret
	}
	return i.caret, i.anchor
}

func (i *Input) ClearSelection() { i.anchor = i.caret }

func (i *Input) pushUndo() {
	i.history = append(i.history, editSnapshot{text: i.text, caret: i.caret, anchor: i.anchor})
	if len(i.history) > 30 {
		i.history = i.history[1:]
	}
}

func (i *Input) deleteSelection() {
	lo, hi := i.selectionRange()
	if lo == hi {
		return
	}
	rs := i.runes()
	out := make([]rune, 0, len(rs)-(hi-lo))
	out = append(out, rs[:lo]...)
	out = append(out, rs[hi:]...)
	i.text = string(out)
	i.caret, i.anchor = lo, lo
}

func (i *Input) insert(s string) {
	s = sanitizeText(s) // defense in depth: never let control bytes reach the buffer
	if s == "" {
		return
	}
	i.clamp(len(i.runes()))
	i.pushUndo()
	i.deleteSelection()
	rs := i.runes()
	ins := []rune(s)
	out := make([]rune, 0, len(rs)+len(ins))
	out = append(out, rs[:i.caret]...)
	out = append(out, ins...)
	out = append(out, rs[i.caret:]...)
	i.text = string(out)
	i.caret += len(ins)
	i.anchor = i.caret
}

func (i *Input) undo() {
	if len(i.history) == 0 {
		return
	}
	snap := i.history[len(i.history)-1]
	i.history = i.history[:len(i.history)-1]
	i.text = snap.text
	n := len(i.runes())
	i.caret, i.anchor = snap.caret, snap.anchor
	i.clamp(n)
}

// Clipboard access is guarded: in unit tests lcl.Init() is never called, so
// lcl.Clipboard is nil, and a real desktop session can also fail to open the
// clipboard while another app holds it. Neither case may crash the editor.
func clipboardText() (t string) {
	if lcl.Clipboard == nil {
		return ""
	}
	defer func() { _ = recover() }()
	return lcl.Clipboard.AsText()
}

func setClipboardText(s string) (ok bool) {
	if lcl.Clipboard == nil {
		return false
	}
	defer func() {
		if recover() != nil {
			ok = false
		}
	}()
	lcl.Clipboard.SetAsText(s)
	return true
}

func (i *Input) copySelection(cut bool) {
	lo, hi := i.selectionRange()
	if lo == hi {
		return
	}
	sel := string(i.runes()[lo:hi])
	if !setClipboardText(sel) {
		return // never destroy a selection when the clipboard is unavailable
	}
	if cut {
		i.pushUndo()
		i.deleteSelection()
	}
}

func (i *Input) paste() {
	t := clipboardText()
	t = strings.ReplaceAll(t, "\r\n", " ")
	t = strings.ReplaceAll(t, "\n", " ")
	t = strings.ReplaceAll(t, "\r", " ")
	if t != "" {
		i.insert(t)
	}
}

func (i *Input) moveCaret(pos int, extend bool) {
	n := len(i.runes())
	if pos < 0 {
		pos = 0
	}
	if pos > n {
		pos = n
	}
	if !extend && i.hasSelection() {
		lo, hi := i.selectionRange()
		if pos < i.caret {
			pos = lo
		} else {
			pos = hi
		}
	}
	i.caret = pos
	if !extend {
		i.anchor = pos
	}
}

func (i *Input) backspace() {
	if i.hasSelection() {
		i.pushUndo()
		i.deleteSelection()
		return
	}
	if i.caret == 0 {
		return
	}
	i.pushUndo()
	rs := i.runes()
	out := make([]rune, 0, len(rs)-1)
	out = append(out, rs[:i.caret-1]...)
	out = append(out, rs[i.caret:]...)
	i.text = string(out)
	i.caret--
	i.anchor = i.caret
}

func (i *Input) deleteForward() {
	if i.hasSelection() {
		i.pushUndo()
		i.deleteSelection()
		return
	}
	rs := i.runes()
	if i.caret >= len(rs) {
		return
	}
	i.pushUndo()
	out := make([]rune, 0, len(rs)-1)
	out = append(out, rs[:i.caret]...)
	out = append(out, rs[i.caret+1:]...)
	i.text = string(out)
	i.anchor = i.caret
}

func isWordSep(r rune) bool {
	switch r {
	case ' ', '\t', '/', '\\', '.', ',', ';', ':', '-', '(', ')', '[', ']':
		return true
	}
	return false
}

// wordBoundary returns the previous/next word start when dir is -1/+1.
func (i *Input) wordBoundary(dir int) int {
	rs := i.runes()
	p := i.caret
	if dir < 0 {
		for p > 0 && isWordSep(rs[p-1]) {
			p--
		}
		for p > 0 && !isWordSep(rs[p-1]) {
			p--
		}
	} else {
		for p < len(rs) && isWordSep(rs[p]) {
			p++
		}
		for p < len(rs) && !isWordSep(rs[p]) {
			p++
		}
	}
	return p
}

func (i *Input) deleteWordBackward() {
	if i.hasSelection() {
		i.pushUndo()
		i.deleteSelection()
		return
	}
	from := i.wordBoundary(-1)
	if from == i.caret {
		return
	}
	i.pushUndo()
	rs := i.runes()
	out := make([]rune, 0, len(rs)-(i.caret-from))
	out = append(out, rs[:from]...)
	out = append(out, rs[i.caret:]...)
	i.text = string(out)
	i.caret, i.anchor = from, from
}

func (i *Input) onKey(key uint16, shift types.TShiftState) {
	n := len(i.runes())
	i.clamp(n)
	ctrl := shift.In(types.SsCtrl)
	ext := shift.In(types.SsShift)
	// LCL passes an uppercase VK code for letter keys, but accept lowercase too
	// so shortcut matching never depends on the keyboard layout / Shift state.
	letter := key
	if letter >= 'a' && letter <= 'z' {
		letter -= 'a' - 'A'
	}
	switch {
	case ctrl && letter == 'A': // Ctrl+A
		i.anchor, i.caret = 0, n
	case ctrl && letter == 'C', ctrl && key == vkInsert: // Ctrl+C / Ctrl+Ins
		i.copySelection(false)
	case ctrl && letter == 'X', ext && key == vkDelete: // Ctrl+X / Shift+Del
		i.copySelection(true)
	case ctrl && letter == 'V', ext && key == vkInsert: // Ctrl+V / Shift+Ins
		i.paste()
	case ctrl && letter == 'Z': // Ctrl+Z
		i.undo()
	case key == vkBack:
		if ctrl {
			i.deleteWordBackward()
		} else {
			i.backspace()
		}
	case key == vkDelete:
		i.deleteForward()
	case key == vkLeft:
		if ctrl {
			i.moveCaret(i.wordBoundary(-1), ext)
		} else {
			i.moveCaret(i.caret-1, ext)
		}
	case key == vkRight:
		if ctrl {
			i.moveCaret(i.wordBoundary(1), ext)
		} else {
			i.moveCaret(i.caret+1, ext)
		}
	case key == vkHome:
		i.moveCaret(0, ext)
	case key == vkEnd:
		i.moveCaret(n, ext)
	case key == vkReturn: // single-line input -> submit
		if i.onSubmit != nil {
			i.onSubmit(i.text)
		}
	}
}

func (i *Input) Paint(p *Painter) {
	edge := p.Theme.Edge
	if i.focused {
		edge = p.Theme.Accent
	}
	p.Shadow(i.rect, 6)
	p.RoundRect(i.rect, 6, p.Theme.Panel, edge)

	rs := i.runes()
	n := len(rs)
	i.clamp(n)
	th := p.TextHeight("Hg")
	y := i.rect.Top + (i.rect.Height()-th)/2
	x := i.rect.Left + 10
	lo, hi := i.selectionRange()

	// Pre-edit string is inserted exactly at the caret, with an underline.
	drawComp := func(px int32) int32 {
		if i.composition == "" || p.Canvas == nil {
			return px
		}
		brush := p.Canvas.BrushToBrush()
		brush.SetStyle(types.BsClear)
		p.Canvas.FontToFont().SetColor(p.Theme.Compose)
		p.Canvas.TextOutWithIntX2Str(px+2, y, i.composition)
		brush.SetStyle(types.BsSolid)
		cw := p.TextWidth(i.composition)
		p.FillRect(types.TRect{
			Left: px + 2, Top: y + th - 1,
			Right: px + 2 + cw, Bottom: y + th,
		}, p.Theme.Compose)
		return px + cw + 2
	}

	for idx, r := range rs {
		if idx == i.caret {
			x = drawComp(x)
			i.caretX, i.caretY = x+2, y
		}
		ch := string(r)
		cw := p.TextWidth(ch)
		if idx >= lo && idx < hi {
			p.FillRect(types.TRect{
				Left: x, Top: y - 2,
				Right: x + cw, Bottom: y + th + 2,
			}, p.Theme.Accent)
		}
		x = p.TextAt(x, y, ch, p.Theme.Text)
	}
	if i.caret == n {
		x = drawComp(x)
		i.caretX, i.caretY = x+2, y
	}
	if i.focused {
		p.Caret(i.caretX, i.caretY, th, p.Theme.Caret)
	}
}

func (i *Input) Handle(e *InputEvent) bool {
	switch e.Kind {
	case EvText:
		i.insert(e.Text)
	case EvKeyDown:
		i.onKey(e.Key, e.Shift)
	}
	return true
}

// ---- ProgressBar: gradient fill, click to change value ----

type ProgressBar struct {
	baseWidget
	value float64 // 0..100
}

func NewProgressBar(name string, r types.TRect) *ProgressBar {
	return &ProgressBar{baseWidget: baseWidget{name: name, rect: r}, value: 40}
}

func (b *ProgressBar) Paint(p *Painter) {
	p.Shadow(b.rect, 8)
	p.RoundRect(b.rect, 8, p.Theme.PanelAlt, p.Theme.Edge)
	inner := inset(b.rect, 4)
	w := int32(inner.Width() * int32(b.value) / 100)
	if w > 0 {
		p.Gradient(types.TRect{
			Left: inner.Left, Top: inner.Top,
			Right: inner.Left + w, Bottom: inner.Bottom,
		}, p.Theme.Accent, p.Theme.AccentDown)
	}
	p.Text(b.rect, fmt.Sprintf("%d%%  点击调整", int32(b.value)), p.Theme.Text, false)
}

func (b *ProgressBar) Handle(e *InputEvent) bool {
	if e.Kind == EvMouseDown || (e.Kind == EvMouseMove && b.Hit(e.X, e.Y)) {
		v := float64(e.X-b.rect.Left) * 100 / float64(b.rect.Width())
		if v < 0 {
			v = 0
		}
		if v > 100 {
			v = 100
		}
		b.value = v
		return true
	}
	return false
}

// ---- StatusPanel: layered card, gradient strip, multi-line IME report ----

type StatusPanel struct {
	baseWidget
	scene *Scene
}

func NewStatusPanel(name string, r types.TRect, s *Scene) *StatusPanel {
	return &StatusPanel{baseWidget: baseWidget{name: name, rect: r}, scene: s}
}

func (sp *StatusPanel) Paint(p *Painter) {
	r := sp.rect
	p.Shadow(r, 12)
	p.RoundRect(r, 12, p.Theme.Panel, p.Theme.Edge)
	p.Gradient(types.TRect{
		Left: r.Left + 14, Top: r.Top + 10,
		Right: r.Right - 14, Bottom: r.Top + 16,
	}, p.Theme.Accent, p.Theme.AccentDown)

	title := types.TRect{Left: r.Left + 16, Top: r.Top + 24, Right: r.Right - 16, Bottom: r.Top + 48}
	p.Text(title, "IME 测试面板 · 单画布渲染状态", p.Theme.Text, true)

	top := r.Top + 56
	for idx, line := range sp.scene.StatusLines() {
		c := p.Theme.Text
		if idx >= 5 {
			c = p.Theme.Muted
		}
		lr := types.TRect{
			Left: r.Left + 16, Top: top, Right: r.Right - 16, Bottom: top + 22,
		}
		p.Text(lr, line, c, true)
		top += 22
	}
}
