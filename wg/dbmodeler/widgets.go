package main

// 通用纯自绘控件：Button / Label / Input / CheckBox / Combo / List。
// 它们都是纯 Go 对象，没有 LCL 实例、没有 HWND，全部画在同一张画布上。
// Input 从 canvas_ime 示例移植了完整的键盘编辑（rune 光标、选区、剪贴板、
// 自身撤销、按词移动）与 IME 预编辑渲染，保证中文表名/注释可以直接输入。

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
)

// ---- 事件 ----

type EventKind uint8

const (
	EvDown EventKind = iota
	EvMove
	EvUp
	EvWheel
	EvKey
	EvText
	EvDbl
	EvLeave
)

type Event struct {
	Kind   EventKind
	X, Y   int32
	Key    uint16
	Text   string
	Shift  types.TShiftState
	Button types.TMouseButton
	Wheel  int32
}

const (
	vkBack   = 8
	vkTab    = 9
	vkReturn = 13
	vkShift  = 16
	vkCtrl   = 17
	vkEscape = 27
	vkEnd    = 35
	vkHome   = 36
	vkLeft   = 37
	vkUp     = 38
	vkRight  = 39
	vkDown   = 40
	vkInsert = 45
	vkDelete = 46
	vkF1     = 112
	vkF2     = 113
)

// sanitizeText 在事件边界拦掉 Windows TranslateMessage 送来的控制字符：
// Backspace=0x08、Ctrl+A=0x01、Enter=0x0D、Tab=0x09 等都会进 UTF8KeyPress。
// 中文/emoji 上屏字符 rune >= 0xA0，不受影响。
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

type Widget interface {
	Name() string
	Rect() types.TRect
	SetRect(r types.TRect)
	Hit(x, y int32) bool
	Focusable() bool
	Paint(p *Painter)
	Handle(e *Event) bool
	OnFocusGain()
	OnFocusLost()
}

type baseWidget struct {
	name      string
	rect      types.TRect
	focusable bool
	enabled   bool
	visible   bool
}

func newBase(name string, r types.TRect, focusable bool) baseWidget {
	return baseWidget{name: name, rect: r, focusable: focusable, enabled: true, visible: true}
}
func (b *baseWidget) Name() string          { return b.name }
func (b *baseWidget) Rect() types.TRect     { return b.rect }
func (b *baseWidget) SetRect(r types.TRect) { b.rect = r }
func (b *baseWidget) Focusable() bool       { return b.focusable && b.enabled && b.visible }
func (b *baseWidget) OnFocusGain()          {}
func (b *baseWidget) OnFocusLost()          {}
func (b *baseWidget) Paint(p *Painter)      {}
func (b *baseWidget) Handle(e *Event) bool  { return false }
func (b *baseWidget) Enabled() bool         { return b.enabled && b.visible }
func (b *baseWidget) SetEnabled(v bool)     { b.enabled = v }
func (b *baseWidget) SetVisible(v bool)     { b.visible = v }
func (b *baseWidget) Visible() bool         { return b.visible }
func (b *baseWidget) Hit(x, y int32) bool {
	return b.visible && x >= b.rect.Left && x < b.rect.Right && y >= b.rect.Top && y < b.rect.Bottom
}

func r4(l, t, r, b int32) types.TRect {
	return types.TRect{Left: l, Top: t, Right: r, Bottom: b}
}

func insetR(r types.TRect, d int32) types.TRect {
	return r4(r.Left+d, r.Top+d, r.Right-d, r.Bottom-d)
}

// ---- Label ----

type Label struct {
	baseWidget
	Text     string
	Muted    bool
	Bold     bool
	Height   int32
	Align    int32
	Color    types.TColor
	HasColor bool
}

func NewLabel(name string, r types.TRect, text string) *Label {
	return &Label{baseWidget: newBase(name, r, false), Text: text, Height: 13, Align: AlignLeft}
}

func (l *Label) Paint(p *Painter) {
	if !l.visible {
		return
	}
	c := p.Theme.Text
	if l.Muted {
		c = p.Theme.Muted
	}
	if l.HasColor {
		c = l.Color
	}
	p.Text(l.rect, l.Text, c, l.Height, l.Align, l.Bold)
}

// ---- Button ----

type Button struct {
	baseWidget
	Caption string
	Hint    string
	Active  bool
	Danger  bool
	Compact bool
	hover   bool
	down    bool
	focused bool
	OnClick func()
	OnHover func(string)
}

func NewButton(name, caption string, onClick func()) *Button {
	return &Button{baseWidget: newBase(name, r4(0, 0, 100, 30), true), Caption: caption, OnClick: onClick}
}

func (b *Button) OnFocusGain() { b.focused = true }
func (b *Button) OnFocusLost() { b.focused = false }

func (b *Button) Paint(p *Painter) {
	if !b.visible {
		return
	}
	r := b.rect
	bg, fg, border := p.Theme.Surface, p.Theme.Text, p.Theme.Border
	if b.Active {
		bg, fg, border = p.Theme.Accent, p.Theme.Surface, p.Theme.AccentDown
	}
	if b.hover && b.enabled {
		if !b.Active {
			bg = p.Theme.SurfaceAlt
		}
		border = p.Theme.Accent
	}
	if b.down && b.enabled {
		bg = p.Theme.AccentDown
		fg = p.Theme.Surface
	}
	if !b.enabled {
		bg, fg, border = p.Theme.Surface, p.Theme.Muted, p.Theme.Border
	}
	if b.Danger && b.enabled {
		fg = p.Theme.Danger
	}
	if b.focused {
		border = p.Theme.Accent
	}
	p.RoundRect(r, 5, bg, border)
	h := int32(12)
	if b.Compact {
		h = 11
	}
	p.Text(insetR(r, 4), b.Caption, fg, h, AlignCenter, b.Active)
}

func (b *Button) Handle(e *Event) bool {
	switch e.Kind {
	case EvDown:
		if b.enabled && e.Button == types.MbLeft {
			b.down = true
		}
	case EvMove:
		b.hover = true
		if b.OnHover != nil {
			b.OnHover(b.Hint)
		}
	case EvLeave:
		b.hover = false
		b.down = false
	case EvUp:
		wasDown := b.down
		b.down = false
		if wasDown && b.enabled && b.Hit(e.X, e.Y) && b.OnClick != nil {
			b.OnClick()
		}
	case EvKey:
		if (e.Key == vkReturn || e.Key == 32) && b.enabled && b.OnClick != nil {
			b.OnClick()
		}
	}
	return true
}

// ---- Input：全自绘单行编辑器 ----

type editSnapshot struct {
	text          string
	caret, anchor int
}

type Input struct {
	baseWidget
	text        string
	composition string
	compCursor  int32
	placeholder string
	focused     bool
	readOnly    bool
	canceled    bool
	CaretX      int32
	CaretY      int32

	caret   int // rune 下标
	anchor  int
	history []editSnapshot
	redo    []editSnapshot

	// 真实鼠标编辑所需的状态：焦点快照（Esc 还原）、逐 rune 像素坐标
	// （点击定位 / 选区 / 横向滚动），以及拖拽与双击检测。
	focusText string
	glyphX    []int32
	textW     int32
	scrollX   int32
	dragging  bool
	clickAt   time.Time
	clickX    int32

	OnSubmit func(string)
	OnChange func(string)
	OnCancel func()
}

func NewInput(name string, r types.TRect) *Input {
	return &Input{baseWidget: newBase(name, r, true)}
}

func (i *Input) Text() string { return i.text }
func (i *Input) SetText(v string) {
	i.text = v
	n := len([]rune(v))
	i.caret, i.anchor = n, n
	i.glyphX, i.textW, i.scrollX = nil, 0, 0
}
func (i *Input) SetPlaceholder(v string) { i.placeholder = v }
func (i *Input) SetReadOnly(v bool)      { i.readOnly = v }
func (i *Input) SetComposition(text string, cursor int32) {
	i.composition, i.compCursor = text, cursor
}

// metrics 作废逐 rune 坐标缓存，下一次绘制 / 点击会重新测量。
func (i *Input) metrics() { i.glyphX, i.textW = nil, 0 }

func (i *Input) OnFocusGain() {
	i.focused = true
	i.canceled = false
	i.focusText = i.text
	i.clickAt = time.Time{}
}

func (i *Input) OnFocusLost() {
	i.focused = false
	i.composition = ""
	i.dragging = false
	if !i.canceled && i.OnSubmit != nil {
		i.OnSubmit(i.text)
	}
	i.canceled = false
}
func (i *Input) Commit() {
	if i.OnSubmit != nil {
		i.OnSubmit(i.text)
	}
}
func (i *Input) SelectAll() {
	n := len(i.runes())
	i.anchor, i.caret = 0, n
}

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
	i.redo = nil
}

func (i *Input) snapshot() editSnapshot {
	return editSnapshot{text: i.text, caret: i.caret, anchor: i.anchor}
}

func (i *Input) applySnapshot(snap editSnapshot) {
	i.text = snap.text
	i.caret, i.anchor = snap.caret, snap.anchor
	i.clamp(len(i.runes()))
	i.changed()
}

func (i *Input) changed() {
	i.metrics()
	if i.OnChange != nil {
		i.OnChange(i.text)
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
	i.changed()
}

func (i *Input) insert(s string) {
	if i.readOnly {
		return
	}
	s = sanitizeText(s)
	if s == "" {
		return
	}
	i.clamp(len(i.runes()))
	i.pushUndo()
	i.deleteSelection()
	rs, ins := i.runes(), []rune(s)
	out := make([]rune, 0, len(rs)+len(ins))
	out = append(out, rs[:i.caret]...)
	out = append(out, ins...)
	out = append(out, rs[i.caret:]...)
	i.text = string(out)
	i.caret += len(ins)
	i.anchor = i.caret
	i.changed()
}

func (i *Input) undo() {
	if len(i.history) == 0 {
		return
	}
	snap := i.history[len(i.history)-1]
	i.history = i.history[:len(i.history)-1]
	i.redo = append(i.redo, i.snapshot())
	i.applySnapshot(snap)
}

func (i *Input) redoEdit() {
	if len(i.redo) == 0 {
		return
	}
	snap := i.redo[len(i.redo)-1]
	i.redo = i.redo[:len(i.redo)-1]
	i.history = append(i.history, i.snapshot())
	i.applySnapshot(snap)
}

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
		return
	}
	if cut && !i.readOnly {
		i.pushUndo()
		i.deleteSelection()
	}
}

func (i *Input) paste() {
	t := clipboardText()
	t = strings.NewReplacer("\r\n", " ", "\n", " ", "\r", " ").Replace(t)
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

func isWordSep(r rune) bool {
	switch r {
	case ' ', '\t', '/', '\\', '.', ',', ';', ':', '-', '_', '(', ')', '[', ']':
		return true
	}
	return false
}

func (i *Input) wordBoundary(dir int) int {
	rs, p := i.runes(), i.caret
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

func (i *Input) backspace() {
	if i.readOnly {
		return
	}
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
	i.changed()
}

func (i *Input) deleteForward() {
	if i.readOnly {
		return
	}
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
	i.changed()
}

func deleteWordBackward(i *Input) {
	if i.readOnly {
		return
	}
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
	i.changed()
}

func (i *Input) onKey(key uint16, shift types.TShiftState) {
	n := len(i.runes())
	i.clamp(n)
	ctrl := shift.In(types.SsCtrl)
	ext := shift.In(types.SsShift)
	letter := key
	if letter >= 'a' && letter <= 'z' {
		letter -= 'a' - 'A'
	}
	switch {
	case ctrl && letter == 'A':
		i.SelectAll()
	case ctrl && letter == 'C', ctrl && key == vkInsert:
		i.copySelection(false)
	case ctrl && letter == 'X', ext && key == vkDelete:
		i.copySelection(true)
	case ctrl && letter == 'V', ext && key == vkInsert:
		i.paste()
	case ctrl && letter == 'Z':
		i.undo()
	case ctrl && letter == 'Y':
		i.redoEdit()
	case key == vkBack:
		if ctrl {
			deleteWordBackward(i)
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
	case key == vkReturn:
		i.Commit()
	case key == vkEscape:
		// Esc 是「取消本次编辑」：还原获得焦点时的文本，并标记为已取消，
		// 这样随后的失焦不会再把半成品 OnSubmit 出去。
		i.canceled = true
		i.composition = ""
		if i.text != i.focusText {
			i.text = i.focusText
			n := len(i.runes())
			i.caret, i.anchor = n, n
			i.changed()
		}
		i.ClearSelection()
		if i.OnCancel != nil {
			i.OnCancel()
		}
	}
}

// estimateRuneWidth 是没有画布时的字符宽度估算：中文/全角按字号 1:1，
// 半角按约 0.54 倍。只用于离屏逻辑测试里的点击命中，真实窗口用字体测量。
func estimateRuneWidth(r rune, h int32) int32 {
	if r >= 0x2E80 {
		return h
	}
	return h * 7 / 13
}

// ensureMetrics 建立「第 k 个 rune 的像素偏移」缓存，长度恒为 n+1。
func (i *Input) ensureMetrics(p *Painter) {
	rs := i.runes()
	n := len(rs)
	if len(i.glyphX) == n+1 {
		return
	}
	i.glyphX = make([]int32, n+1)
	acc := int32(0)
	for k, r := range rs {
		i.glyphX[k] = acc
		if p != nil && p.ok() {
			acc += p.TextWidth(string(r), 13, false)
		} else {
			acc += estimateRuneWidth(r, 13)
		}
	}
	i.glyphX[n] = acc
	i.textW = acc
}

func (i *Input) Paint(p *Painter) {
	if !i.visible {
		return
	}
	edge := p.Theme.Border
	if i.focused {
		edge = p.Theme.Accent
	}
	bg := p.Theme.Surface
	if !i.enabled {
		bg = p.Theme.SurfaceAlt
	}
	p.RoundRect(i.rect, 5, bg, edge)

	rs := i.runes()
	n := len(rs)
	i.clamp(n)
	th := p.TextHeight("Hg", 13, false)
	y := i.rect.Top + (i.rect.Height()-th)/2
	textLeft := i.rect.Left + 7
	right := i.rect.Right - 5

	if n == 0 && i.composition == "" && !i.focused && i.placeholder != "" {
		p.Text(insetR(i.rect, 7), i.placeholder, p.Theme.Muted, 12, AlignLeft, false)
	}

	i.ensureMetrics(p)
	viewW := right - textLeft
	if viewW < 8 {
		viewW = 8
	}
	// 让光标始终留在可视区内：超出右边界就向左滚，回到 0 左边就向右滚。
	caretPx := i.glyphX[i.caret]
	if caretPx-i.scrollX > viewW {
		i.scrollX = caretPx - viewW
	}
	if caretPx-i.scrollX < 0 {
		i.scrollX = caretPx
	}
	maxScroll := i.textW - viewW
	if maxScroll < 0 {
		maxScroll = 0
	}
	if i.scrollX > maxScroll {
		i.scrollX = maxScroll
	}
	if i.scrollX < 0 {
		i.scrollX = 0
	}

	base := textLeft - i.scrollX
	inner := r4(textLeft-3, i.rect.Top+2, right+2, i.rect.Bottom-2)
	p.PushClip(inner)
	lo, hi := i.selectionRange()
	for idx := 0; idx < n; idx++ {
		cx := base + i.glyphX[idx]
		cw := i.glyphX[idx+1] - i.glyphX[idx]
		if cx+cw < inner.Left || cx > inner.Right {
			continue
		}
		if idx >= lo && idx < hi {
			p.FillRect(r4(cx, y-1, cx+cw, y+th+1), p.Theme.AccentSoft)
		}
		p.TextAt(cx, y, string(rs[idx]), p.Theme.Text, 13, false)
	}
	caretX := base + i.glyphX[i.caret]
	compStart := caretX
	if i.composition != "" && p.ok() {
		p.font(p.Theme.Warn, 13, false)
		p.brush(p.Theme.Warn, types.BsClear)
		p.Canvas.TextOutWithIntX2Str(caretX+1, y, i.composition)
		cw := p.Canvas.TextWidthWithStr(i.composition)
		p.FillRect(r4(caretX+1, y+th-1, caretX+1+cw, y+th), p.Theme.Warn)
		caretX += cw + 2
	}
	if i.focused {
		p.Caret(caretX, y, th, p.Theme.Accent)
	}
	p.PopClip()

	// 候选框跟随预编辑文本起点，而不是越过预编辑文本。
	i.CaretX, i.CaretY = compStart, y
}

// caretFromX 把鼠标横坐标换算成 rune 下标（取最近的字间边界）。
func (i *Input) caretFromX(x int32) int {
	n := len(i.runes())
	if n == 0 {
		return 0
	}
	local := x - (i.rect.Left + 7 - i.scrollX)
	if len(i.glyphX) == n+1 {
		for k := 0; k < n; k++ {
			if local < (i.glyphX[k]+i.glyphX[k+1])/2 {
				return k
			}
		}
		return n
	}
	acc := int32(0)
	for k, r := range i.runes() {
		w := estimateRuneWidth(r, 13)
		if local < acc+w/2 {
			return k
		}
		acc += w
	}
	return n
}

// selectWordAt 双击选中一个词；中文没有分隔符时会选中连续的一段。
func (i *Input) selectWordAt(pos int) {
	rs := i.runes()
	if len(rs) == 0 {
		i.caret, i.anchor = 0, 0
		return
	}
	if pos >= len(rs) {
		pos = len(rs) - 1
	}
	if isWordSep(rs[pos]) {
		i.caret, i.anchor = pos, pos
		return
	}
	lo, hi := pos, pos+1
	for lo > 0 && !isWordSep(rs[lo-1]) {
		lo--
	}
	for hi < len(rs) && !isWordSep(rs[hi]) {
		hi++
	}
	i.anchor, i.caret = lo, hi
}

func (i *Input) Handle(e *Event) bool {
	switch e.Kind {
	case EvText:
		i.insert(e.Text)
	case EvKey:
		i.onKey(e.Key, e.Shift)
	case EvDown:
		if e.Button != types.MbLeft {
			return true
		}
		now := time.Now()
		if now.Sub(i.clickAt) < 400*time.Millisecond && abs32(e.X-i.clickX) < 4 {
			// 双击选词（与画布双击改名共用同一套自绘双击检测思路）
			i.clickAt = time.Time{}
			i.dragging = false
			i.selectWordAt(i.caretFromX(e.X))
			return true
		}
		i.clickAt, i.clickX = now, e.X
		i.caret = i.caretFromX(e.X)
		i.anchor = i.caret
		i.dragging = true
	case EvMove:
		if i.dragging {
			i.caret = i.caretFromX(e.X)
		}
	case EvUp:
		i.dragging = false
	case EvLeave:
		i.dragging = false
	}
	return true
}

// ---- CheckBox ----

type CheckBox struct {
	baseWidget
	Caption  string
	Checked  bool
	hover    bool
	OnToggle func(bool)
}

func NewCheckBox(name, caption string) *CheckBox {
	return &CheckBox{baseWidget: newBase(name, r4(0, 0, 100, 22), true), Caption: caption}
}

func (c *CheckBox) Paint(p *Painter) {
	if !c.visible {
		return
	}
	r := c.rect
	box := r4(r.Left, r.Top+(r.Height()-15)/2, r.Left+15, r.Top+(r.Height()-15)/2+15)
	bg := p.Theme.Surface
	if c.Checked {
		bg = p.Theme.Accent
	}
	border := p.Theme.Border
	if c.hover || c.Checked {
		border = p.Theme.Accent
	}
	p.RoundRect(box, 3, bg, border)
	if c.Checked {
		tick := p.Theme.Surface
		p.Line(box.Left+3, box.Top+7, box.Left+6, box.Top+11, tick, 2)
		p.Line(box.Left+6, box.Top+11, box.Left+12, box.Top+3, tick, 2)
	}
	p.Text(r4(box.Right+7, r.Top, r.Right, r.Bottom), c.Caption, p.Theme.Text, 12, AlignLeft, false)
}

func (c *CheckBox) Handle(e *Event) bool {
	switch e.Kind {
	case EvMove:
		c.hover = true
	case EvLeave:
		c.hover = false
	case EvDown, EvKey:
		if (e.Kind == EvKey && e.Key != vkReturn && e.Key != 32) ||
			(e.Kind == EvDown && e.Button != types.MbLeft) {
			return true
		}
		if c.enabled {
			c.Checked = !c.Checked
			if c.OnToggle != nil {
				c.OnToggle(c.Checked)
			}
		}
	}
	return true
}

// ---- Combo（点击弹出纯自绘下拉列表）----

type Combo struct {
	baseWidget
	Items    []string
	Value    string
	hover    bool
	OnSelect func(string)
	OnOpen   func(*Combo)
}

func NewCombo(name string, items []string) *Combo {
	return &Combo{baseWidget: newBase(name, r4(0, 0, 120, 28), true), Items: items}
}

func (c *Combo) Paint(p *Painter) {
	if !c.visible {
		return
	}
	edge := p.Theme.Border
	if c.hover {
		edge = p.Theme.Accent
	}
	p.RoundRect(c.rect, 5, p.Theme.Surface, edge)
	p.Text(r4(c.rect.Left+7, c.rect.Top, c.rect.Right-22, c.rect.Bottom),
		c.Value, p.Theme.Text, 12, AlignLeft, false)
	cx, cy := c.rect.Right-13, c.rect.Top+c.rect.Height()/2
	p.Polygon([]types.TPoint{
		{X: cx - 4, Y: cy - 2}, {X: cx + 4, Y: cy - 2}, {X: cx, Y: cy + 3},
	}, p.Theme.Muted, p.Theme.Muted)
}

func (c *Combo) Handle(e *Event) bool {
	switch e.Kind {
	case EvMove:
		c.hover = true
	case EvLeave:
		c.hover = false
	case EvDown:
		if e.Button == types.MbLeft && c.OnOpen != nil {
			c.OnOpen(c)
		}
	case EvKey:
		if (e.Key == vkReturn || e.Key == 32 || e.Key == vkDown) && c.OnOpen != nil {
			c.OnOpen(c)
		}
	}
	return true
}

// ---- List（左侧表导航）----

type ListItem struct {
	ID       string
	Title    string
	Sub      string
	Group    string
	Selected bool
	Count    int
}

type List struct {
	baseWidget
	Items    []ListItem
	Scroll   int32
	RowH     int32
	hover    int
	cur      int
	focused  bool
	dragBar  bool
	dragDY   int32
	OnSelect func(int)
	OnHover  func(string)
	OnWheel  func(int32)
}

func NewList(name string) *List {
	return &List{baseWidget: newBase(name, r4(0, 0, 200, 300), true), RowH: 44, hover: -1, cur: -1}
}

func (l *List) OnFocusGain() { l.focused = true }
func (l *List) OnFocusLost() { l.focused = false; l.dragBar = false }

func (l *List) SetItems(items []ListItem) {
	l.Items = items
	for i, it := range items {
		if it.Selected {
			l.cur = i
			break
		}
	}
	if l.cur >= len(items) {
		l.cur = len(items) - 1
	}
	l.clampScroll()
}

func (l *List) maxScroll() int32 {
	need := int32(len(l.Items)) * l.RowH
	avail := l.rect.Height()
	if need <= avail {
		return 0
	}
	return need - avail
}

func (l *List) clampScroll() {
	if l.Scroll < 0 {
		l.Scroll = 0
	}
	if m := l.maxScroll(); l.Scroll > m {
		l.Scroll = m
	}
}

func (l *List) barTrack() types.TRect {
	return r4(l.rect.Right-9, l.rect.Top+2, l.rect.Right-2, l.rect.Bottom-2)
}

func (l *List) Paint(p *Painter) {
	if !l.visible {
		return
	}
	r := l.rect
	if l.focused {
		p.StrokeRect(r, p.Theme.Accent, 1)
	}
	for idx, it := range l.Items {
		top := r.Top + int32(idx)*l.RowH - l.Scroll
		bot := top + l.RowH - 3
		if bot <= r.Top || top >= r.Bottom {
			continue
		}
		row := r4(r.Left+6, top, r.Right-6, bot)
		if !p.Visible(row) {
			continue
		}
		bg := p.Theme.Surface
		if it.Selected {
			bg = p.Theme.AccentSoft
		} else if idx == l.hover {
			bg = p.Theme.SurfaceAlt
		}
		border := p.Theme.Border
		if it.Selected {
			border = p.Theme.Accent
		}
		p.RoundRect(row, 5, bg, border)
		p.FillRect(r4(row.Left+1, row.Top+8, row.Left+4, row.Bottom-8),
			groupColor(*p.Theme, it.Group))
		p.Text(r4(row.Left+11, row.Top+4, row.Right-46, row.Top+22),
			it.Title, p.Theme.Text, 12, AlignLeft, it.Selected)
		p.Text(r4(row.Left+11, row.Top+21, row.Right-46, row.Bottom-2),
			it.Sub, p.Theme.Muted, 10, AlignLeft, false)
		p.Text(r4(row.Right-46, row.Top+4, row.Right-6, row.Bottom-2),
			countText(it.Count), p.Theme.Muted, 10, AlignRight, false)
	}
	// 真实滚动条：轨道 + 可拖动滑块，滚轮 / 键盘 / 拖动三套输入都改同一个 Scroll。
	if max := l.maxScroll(); max > 0 {
		track := l.barTrack()
		p.FillRect(track, p.Theme.SurfaceAlt)
		thumb := scrollThumb(track, l.Scroll, max)
		p.RoundRect(thumb, 3, p.Theme.Border, p.Theme.Border)
	}
}

// ---- 通用滚动条几何 ----

// scrollThumb 按 scroll/maxScroll 比例算出滑块矩形（最小高度 20）。
func scrollThumb(track types.TRect, scroll, maxScroll int32) types.TRect {
	h := int32(20)
	if h > track.Height() {
		h = track.Height()
	}
	if maxScroll < 0 {
		maxScroll = 0
	}
	span := track.Height() - h
	y := track.Top
	if span > 0 && maxScroll > 0 {
		y += int32(float64(scroll) / float64(maxScroll) * float64(span))
	}
	return r4(track.Left, y, track.Right, y+h)
}

// scrollFromThumb 把鼠标 y 反算成 scroll。grabDY 是按下时鼠标在滑块内的偏移，
// 这样拖动时滑块不会跳到鼠标中心。
func scrollFromThumb(track types.TRect, thumbH, y, grabDY, maxScroll int32) int32 {
	span := track.Height() - thumbH
	if span <= 0 || maxScroll <= 0 {
		return 0
	}
	rel := y - track.Top - grabDY
	if rel < 0 {
		rel = 0
	}
	if rel > span {
		rel = span
	}
	return int32(float64(rel) / float64(span) * float64(maxScroll))
}

func countText(n int) string {
	if n < 0 {
		n = 0
	}
	if n < 10 {
		return " " + string(rune('0'+n)) + " 列"
	}
	return fmt.Sprintf("%d 列", n)
}

func (l *List) indexAt(y int32) int {
	idx := int((y - l.rect.Top + l.Scroll) / l.RowH)
	if idx < 0 || idx >= len(l.Items) {
		return -1
	}
	return idx
}

func (l *List) visibleRows() int32 {
	n := l.rect.Height() / l.RowH
	if n < 1 {
		n = 1
	}
	return n
}

func (l *List) ensureVisible(idx int) {
	if idx < 0 {
		return
	}
	top := int32(idx) * l.RowH
	if top < l.Scroll {
		l.Scroll = top
	}
	if bottom := top + l.RowH; bottom > l.Scroll+l.rect.Height() {
		l.Scroll = bottom - l.rect.Height()
	}
	l.clampScroll()
}

func (l *List) selectIndex(idx int) {
	if idx < 0 || idx >= len(l.Items) {
		return
	}
	l.cur = idx
	l.ensureVisible(idx)
	if l.OnSelect != nil {
		l.OnSelect(idx)
	}
}

func (l *List) moveSel(delta int) {
	if len(l.Items) == 0 {
		return
	}
	idx := l.cur
	if idx < 0 {
		idx = 0
	} else {
		idx += delta
	}
	if idx < 0 {
		idx = 0
	}
	if idx >= len(l.Items) {
		idx = len(l.Items) - 1
	}
	l.selectIndex(idx)
}

func (l *List) Handle(e *Event) bool {
	switch e.Kind {
	case EvMove:
		if l.dragBar {
			track := l.barTrack()
			thumb := scrollThumb(track, l.Scroll, l.maxScroll())
			l.Scroll = scrollFromThumb(track, thumb.Height(), e.Y, l.dragDY, l.maxScroll())
			l.clampScroll()
			return true
		}
		l.hover = l.indexAt(e.Y)
		if l.hover >= 0 && l.OnHover != nil {
			l.OnHover(l.Items[l.hover].Title)
		}
	case EvLeave:
		l.hover = -1
		l.dragBar = false
	case EvDown:
		if e.Button != types.MbLeft {
			return true
		}
		if max := l.maxScroll(); max > 0 && e.X >= l.barTrack().Left {
			track := l.barTrack()
			thumb := scrollThumb(track, l.Scroll, max)
			if e.Y >= thumb.Top && e.Y < thumb.Bottom {
				l.dragBar = true
				l.dragDY = e.Y - thumb.Top
			} else {
				// 点击轨道：上/下翻一屏，然后可以继续拖动
				if e.Y < thumb.Top {
					l.Scroll -= l.rect.Height() - l.RowH
				} else {
					l.Scroll += l.rect.Height() - l.RowH
				}
				l.clampScroll()
				l.dragBar = true
				l.dragDY = thumb.Height() / 2
			}
			return true
		}
		l.selectIndex(l.indexAt(e.Y))
	case EvUp:
		l.dragBar = false
	case EvWheel:
		l.Scroll -= e.Wheel / 4
		l.clampScroll()
		if l.OnWheel != nil {
			l.OnWheel(e.Wheel)
		}
	case EvKey:
		switch e.Key {
		case vkUp:
			l.moveSel(-1)
		case vkDown:
			l.moveSel(1)
		case vkHome:
			l.selectIndex(0)
		case vkEnd:
			l.selectIndex(len(l.Items) - 1)
		case 33: // PageUp
			l.moveSel(-int(l.visibleRows()))
		case 34: // PageDown
			l.moveSel(int(l.visibleRows()))
		}
	}
	return true
}
