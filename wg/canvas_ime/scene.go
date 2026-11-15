package main

import (
	"fmt"
	"strings"

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
	EvText
)

type InputEvent struct {
	Kind  EventKind
	X, Y  int32
	Key   uint16
	Text  string
	Shift types.TShiftState
}

// Widget: pure Go object. No LCL instance, no HWND, no OnDestroy hook.
type Widget interface {
	Name() string
	Rect() types.TRect
	SetRect(r types.TRect)
	Hit(x, y int32) bool
	Focusable() bool
	Paint(p *Painter)
	Handle(e *InputEvent) bool
	OnFocusGain()
	OnFocusLost()
}

// Scene is the only place that knows about the widget list.
type Scene struct {
	form    *TMainForm
	host    lcl.ICustomGraphicControl
	painter *Painter
	theme   Theme

	widgets []Widget
	focus   Widget
	hover   Widget
	capture Widget

	dirty    types.TRect
	hasDirty bool

	// IME test state, rendered by StatusPanel.
	composition string
	compCursor  int32
	imeResult   string
	resultCount int
	imeLog      []string
	imeMsgCount int
	lastText    string
	imePosInfo  string
	notice      string
	lastKey     string
	filtered    int // count of control chars rejected at the event boundary
}

func NewScene(form *TMainForm) *Scene {
	s := &Scene{form: form, theme: darkTheme(), notice: "就绪 · 试试 Ctrl+A/C/V/X/Z、Delete、方向键"}
	s.painter = &Painter{Theme: &s.theme}

	// The only LCL drawing object in the window. It has no HWND.
	host := lcl.NewCustomGraphicControl(form)
	host.SetAlign(types.AlClient)
	host.SetParent(form)
	host.SetParentBackground(false)
	host.SetControlStyle(host.ControlStyle().Include(types.CsOpaque))
	s.host = host
	s.bindCanvasEvents()
	return s
}

func (s *Scene) bindCanvasEvents() {
	s.host.SetOnPaint(func(sender lcl.IObject) {
		c := s.host.Canvas()
		if c == nil || !c.IsValid() {
			return
		}
		s.painter.Canvas = c
		s.painter.Theme = &s.theme

		s.FillBackground(s.host.ClientRect())
		for _, w := range s.widgets {
			w.Paint(s.painter)
		}
		s.hasDirty = false
	})

	s.host.SetOnMouseDown(func(sender lcl.IObject, button types.TMouseButton, shift types.TShiftState, x, y int32) {
		s.dispatch(&InputEvent{Kind: EvMouseDown, X: x, Y: y, Shift: shift})
	})
	s.host.SetOnMouseMove(func(sender lcl.IObject, shift types.TShiftState, x, y int32) {
		s.dispatch(&InputEvent{Kind: EvMouseMove, X: x, Y: y, Shift: shift})
	})
	s.host.SetOnMouseUp(func(sender lcl.IObject, button types.TMouseButton, shift types.TShiftState, x, y int32) {
		s.dispatch(&InputEvent{Kind: EvMouseUp, X: x, Y: y, Shift: shift})
	})
	s.host.SetOnMouseLeave(func(sender lcl.IObject) {
		if s.hover != nil {
			s.hover.Handle(&InputEvent{Kind: EvMouseLeave})
			s.hover = nil
			s.Invalidate()
		}
	})
}

func (s *Scene) FillBackground(r types.TRect) {
	s.painter.FillRect(r, s.theme.Bg)
}

func (s *Scene) Add(w Widget) { s.widgets = append(s.widgets, w) }

// InvalidateRect merges dirty rectangles into one bounding box. A production
// renderer would clip painting to this region; here it is shown in the panel.
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
	s.host.Invalidate()
}

func (s *Scene) Invalidate() { s.host.Invalidate() }

func (s *Scene) hit(x, y int32) Widget {
	for i := len(s.widgets) - 1; i >= 0; i-- {
		if s.widgets[i].Hit(x, y) {
			return s.widgets[i]
		}
	}
	return nil
}

func (s *Scene) SetFocus(w Widget) {
	if w != nil && !w.Focusable() {
		return
	}
	if s.focus == w {
		s.form.SetFocus()
		s.PositionIME() // pre-position before the IME creates its windows
		return
	}
	if s.focus != nil {
		s.focus.OnFocusLost()
	}
	s.focus = w
	if w != nil {
		w.OnFocusGain()
	}
	s.form.SetFocus() // system keyboard focus stays on the form
	s.InvalidateRect(s.host.ClientRect())
	s.PositionIME() // pre-position before the IME creates its windows
}

func (s *Scene) focusNext() {
	var list []Widget
	for _, w := range s.widgets {
		if w.Focusable() {
			list = append(list, w)
		}
	}
	if len(list) == 0 {
		return
	}
	next := list[0]
	for i, w := range list {
		if w == s.focus {
			next = list[(i+1)%len(list)]
			break
		}
	}
	s.SetFocus(next)
}

func (s *Scene) dispatch(e *InputEvent) {
	switch e.Kind {
	case EvMouseDown:
		w := s.hit(e.X, e.Y)
		s.capture = w
		s.SetFocus(w)
		if w != nil {
			w.Handle(e)
		}
	case EvMouseMove:
		w := s.hit(e.X, e.Y)
		if s.capture != nil {
			w = s.capture
		}
		if s.hover != w {
			if s.hover != nil {
				s.hover.Handle(&InputEvent{Kind: EvMouseLeave})
			}
			s.hover = w
		}
		if w != nil {
			w.Handle(e)
		}
	case EvMouseUp:
		w := s.capture
		if w == nil {
			w = s.hit(e.X, e.Y)
		}
		s.capture = nil
		if w != nil {
			w.Handle(e)
		}
	case EvKeyDown:
		if e.Key == 9 { // VK_TAB
			s.focusNext()
		} else if e.Key == 27 { // VK_ESCAPE
			s.SetComposition("", 0)
			if in, ok := s.focus.(*Input); ok {
				in.ClearSelection()
			}
		} else if s.focus != nil {
			s.focus.Handle(e)
		}
	case EvText:
		s.lastText = e.Text
		if s.focus != nil {
			s.focus.Handle(e)
		}
	}
	s.Invalidate()
}

func (s *Scene) OnKey(key uint16, shift types.TShiftState) {
	s.dispatch(&InputEvent{Kind: EvKeyDown, Key: key, Shift: shift})
}

func (s *Scene) OnText(text string) {
	s.dispatch(&InputEvent{Kind: EvText, Text: text})
	s.PositionIME() // caret moved after inserting committed text
}

// ---- IME test plumbing ----

func (s *Scene) SetComposition(text string, cursor int32) {
	s.composition, s.compCursor = text, cursor
	if in, ok := s.focus.(*Input); ok {
		in.composition = text
		in.compCursor = cursor
	}
	s.Invalidate()
}

func (s *Scene) SetResult(text string) {
	if text != "" {
		s.imeResult = text
		s.resultCount++
	}
	s.Invalidate()
}

func (s *Scene) LogIME(line string) {
	s.imeMsgCount++
	fmt.Println("[IME]", line)
	s.imeLog = append(s.imeLog, line)
	if len(s.imeLog) > 3 {
		s.imeLog = s.imeLog[len(s.imeLog)-3:]
	}
	s.Invalidate()
}

// LogKey records what the window actually delivered. It makes it obvious when
// a key was swallowed as a control char instead of silently corrupting text.
func (s *Scene) LogKey(line string) {
	s.lastKey = line
	if strings.HasPrefix(line, "过滤") {
		s.filtered++
	}
	fmt.Println("[KEY]", line)
	s.Invalidate()
}

// CountIMEMessage handles WM_IME_NOTIFY, which arrives far more often than
// real composition events. Counting without logging/repainting keeps the
// console readable and avoids forcing a full redraw on every notification.
func (s *Scene) CountIMEMessage() {
	s.imeMsgCount++
	if s.imeMsgCount%30 == 0 {
		s.Invalidate()
	}
}

func (s *Scene) CaretPos() (int32, int32) {
	if in, ok := s.focus.(*Input); ok {
		return in.caretX, in.caretY
	}
	return 40, 140
}

// PositionIME moves the composition/candidate windows to the self-drawn caret.
// CaretPos is in CANVAS coordinates; COMPOSITIONFORM/CANDIDATEFORM need
// coordinates relative to the FORM's client area, so convert with
// TControl.ClientToParent. Never use screen coordinates here.
func (s *Scene) PositionIME() {
	x, y := s.CaretPos()
	pt := s.host.ClientToParent(types.TPoint{X: x, Y: y}, s.form)
	s.imePosInfo = imePosition(uintptr(s.form.Handle()), pt.X, pt.Y)
	s.Invalidate()
}

// SimulateComposition exercises the pre-edit rendering path without an IME.
func (s *Scene) SimulateComposition() {
	s.LogIME("SIMULATE composition ni'hao")
	s.SetComposition("ni'hao", 6)
	s.PositionIME()
}

func (s *Scene) ToggleTheme() {
	if s.theme.Bg == darkTheme().Bg {
		s.theme = lightTheme()
	} else {
		s.theme = darkTheme()
	}
	s.painter.Theme = &s.theme
	s.InvalidateRect(s.host.ClientRect())
}

func (s *Scene) StatusLines() []string {
	focusName := "<none>"
	if s.focus != nil {
		focusName = s.focus.Name()
	}
	lines := []string{
		"焦点组件: " + focusName + "    主题: " + themeName(s.theme),
		"IME 组合串: " + quote(s.composition) + "    光标位置: " + fmt.Sprint(s.compCursor),
		"上屏结果: " + quote(s.imeResult) + "    UTF8 次数: " + fmt.Sprint(s.resultCount),
		"IME 消息数: " + fmt.Sprint(s.imeMsgCount) + "    最近输入: " + quote(s.lastText),
		"IME 定位: " + s.imePosInfo + "   (comp=1/cand=1 表示设置成功)",
		"编辑: " + s.EditStatus() + "   |   " + s.notice,
		"最近按键: " + s.lastKey + "    已过滤控制字符: " + fmt.Sprint(s.filtered),
		"合并脏区: " + dirtyString(s.dirty, s.hasDirty),
	}
	for _, l := range s.imeLog {
		lines = append(lines, "> "+l)
	}
	return lines
}

func (s *Scene) SetNotice(v string) {
	s.notice = v
	s.Invalidate()
}

func (s *Scene) EditStatus() string {
	if in, ok := s.focus.(*Input); ok {
		lo, hi := in.selectionRange()
		return fmt.Sprintf("caret=%d 选区=[%d,%d) 长度=%d 撤销=%d",
			in.caret, lo, hi, len(in.runes()), len(in.history))
	}
	return "焦点不在输入框"
}

func themeName(t Theme) string {
	if t.Bg == darkTheme().Bg {
		return "dark"
	}
	return "light"
}

func quote(v string) string {
	if v == "" {
		return "<empty>"
	}
	return "\"" + v + "\""
}

func dirtyString(r types.TRect, ok bool) string {
	if !ok {
		return "<clean>"
	}
	return fmt.Sprintf("(%d,%d)-(%d,%d)", r.Left, r.Top, r.Right, r.Bottom)
}
