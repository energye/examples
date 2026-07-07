// Package widget provides UI widgets
package widget

import (
	"time"

	"github.com/energye/examples/lcl/gpui/core/math"
	"github.com/energye/examples/lcl/gpui/render/font"
	"github.com/energye/examples/lcl/gpui/render/pipeline"
	"github.com/energye/examples/lcl/gpui/style/animation"
	"github.com/energye/examples/lcl/gpui/style/color"
	"github.com/energye/examples/lcl/gpui/style/theme"
)

// TextBox is a text input widget
type TextBox struct {
	BaseWidget

	// Content
	text        []rune
	placeholder string
	cursorPos   int
	selStart    int
	selEnd      int

	// Style
	font    *font.Font
	scrollX float32

	// State
	hovered bool

	// Animation
	focusAnim  *animation.Animation
	cursorAnim *animation.Animation

	// Event handlers
	onChange func(text string)
	onSubmit func(text string)
}

// NewTextBox creates a new text box
func NewTextBox(placeholder string, f *font.Font) *TextBox {
	th := theme.GetTheme()

	tb := &TextBox{
		BaseWidget:  NewBaseWidget(),
		text:        make([]rune, 0),
		placeholder: placeholder,
		font:        f,
		selStart:    -1,
		selEnd:      -1,
	}

	// Create animations
	tb.focusAnim = animation.NewAnimation(0, 1, th.DurationNormal, animation.EaseOut)
	// Create looping animation for cursor blink (500ms per cycle)
	tb.cursorAnim = animation.NewLoopAnimation(0, 1, 500*time.Millisecond, animation.Linear)

	// Set size
	tb.SetSize(200, th.Input.Height)

	return tb
}

// Text returns the text content
func (tb *TextBox) Text() string {
	return string(tb.text)
}

// SetText sets the text content
func (tb *TextBox) SetText(text string) {
	tb.text = []rune(text)
	tb.cursorPos = len(tb.text)
}

// SetPlaceholder sets the placeholder text
func (tb *TextBox) SetPlaceholder(placeholder string) {
	tb.placeholder = placeholder
}

// SetOnChange sets the change handler
func (tb *TextBox) SetOnChange(handler func(text string)) {
	tb.onChange = handler
}

// SetOnSubmit sets the submit handler
func (tb *TextBox) SetOnSubmit(handler func(text string)) {
	tb.onSubmit = handler
}

// Focusable returns true
func (tb *TextBox) Focusable() bool {
	return true
}

// Render renders the text box
func (tb *TextBox) Render(renderer *pipeline.Renderer) {
	if !tb.visible {
		return
	}

	th := theme.GetTheme()

	// Update animations (MUST call every frame)
	focusT := tb.focusAnim.Value()
	cursorT := tb.cursorAnim.Value()

	// Calculate colors
	bg, border, textCol, placeCol := tb.calculateColors(focusT)

	// Draw background
	renderer.FillRoundRect(tb.bounds, th.Input.Radius, bg)

	// Draw border
	borderW := th.Input.BorderW
	if tb.focused {
		borderW = 2
	}
	renderer.StrokeRoundRect(tb.bounds, th.Input.Radius, borderW, border)

	// Draw text area with smaller padding
	paddingH := float32(8) // Reduced from 12
	paddingV := float32(4)
	textRect := tb.bounds.Shrink(paddingH, paddingV)
	tb.ensureCursorVisible(textRect)
	renderer.PushClip(textRect)

	// Draw selection
	if tb.selStart >= 0 && tb.selEnd != tb.selStart {
		selRect := tb.calculateSelectionRect(textRect)
		renderer.FillRect(selRect, color.PrimaryBg)
	}

	// Draw text or placeholder
	if len(tb.text) == 0 && !tb.focused {
		// Draw placeholder
		if tb.placeholder != "" && tb.font != nil {
			renderer.DrawText(tb.placeholder, textRect.X, textRect.Y, tb.font, placeCol)
		}
	} else if len(tb.text) > 0 {
		// Draw text
		if tb.font != nil {
			// Apply scroll offset
			textX := textRect.X - tb.scrollX
			renderer.DrawText(string(tb.text), textX, textRect.Y, tb.font, textCol)
		}
	}

	// Draw cursor (blinking)
	if tb.focused {
		// Blink: show for 0.5s, hide for 0.5s
		if cursorT > 0.5 {
			cursorX := tb.calculateCursorX(textRect)
			cursorRect := math.NewRect(cursorX, textRect.Y, 2, textRect.H)
			renderer.FillRect(cursorRect, textCol)
		}
	}
	renderer.PopClip()
}

// calculateColors calculates the current colors
func (tb *TextBox) calculateColors(focusT float32) (bg, border, text, placeholder math.Color) {
	th := theme.GetTheme()

	bg = th.Input.Background
	border = th.Input.Border
	text = th.Input.Text
	placeholder = th.Input.Placeholder

	// Apply disabled state
	if !tb.enabled {
		bg = color.BgDisabled
		border = color.BorderDisabled
		text = color.TextDisabled
	}

	// Apply focus effect
	if focusT > 0 {
		border = border.Lerp(th.Input.Focus, focusT)
	}

	// Apply hover effect
	if tb.hovered && !tb.focused {
		border = color.BorderHover
	}

	return bg, border, text, placeholder
}

// calculateSelectionRect calculates the selection rectangle
func (tb *TextBox) calculateSelectionRect(textRect math.Rect) math.Rect {
	if tb.selStart < 0 || tb.font == nil {
		return math.Rect{}
	}

	start := tb.selStart
	end := tb.selEnd
	if start > end {
		start, end = end, start
	}

	startX := textRect.X - tb.scrollX + tb.font.TextWidth(string(tb.text[:start]))
	endX := textRect.X - tb.scrollX + tb.font.TextWidth(string(tb.text[:end]))

	return math.NewRect(startX, textRect.Y, endX-startX, textRect.H)
}

// calculateCursorX calculates the cursor X position
func (tb *TextBox) calculateCursorX(textRect math.Rect) float32 {
	if tb.font == nil {
		return textRect.X
	}
	return textRect.X - tb.scrollX + tb.font.TextWidth(string(tb.text[:tb.cursorPos]))
}

func (tb *TextBox) ensureCursorVisible(textRect math.Rect) {
	if tb.font == nil || tb.cursorPos < 0 || tb.cursorPos > len(tb.text) {
		return
	}

	cursorLocalX := tb.font.TextWidth(string(tb.text[:tb.cursorPos]))
	if cursorLocalX-tb.scrollX > textRect.W {
		tb.scrollX = cursorLocalX - textRect.W + 2
	}
	if cursorLocalX-tb.scrollX < 0 {
		tb.scrollX = cursorLocalX
	}
	if tb.scrollX < 0 {
		tb.scrollX = 0
	}
}

// hitTestCursor calculates cursor position from mouse X
func (tb *TextBox) hitTestCursor(mouseX float32, textRect math.Rect) int {
	if tb.font == nil || len(tb.text) == 0 {
		return 0
	}

	relX := mouseX - textRect.X + tb.scrollX
	var width float32
	for i, ch := range tb.text {
		g, ok := tb.font.GetGlyph(ch)
		if !ok {
			continue
		}
		width += g.Advance
		if relX < width {
			return i
		}
	}
	return len(tb.text)
}

// MouseDown handles mouse down
func (tb *TextBox) MouseDown(x, y float32, button int) bool {
	if !tb.enabled {
		return false
	}

	if tb.bounds.Contains(x, y) {
		tb.Focus()

		// Calculate cursor position
		th := theme.GetTheme()
		textRect := tb.bounds.Shrink(th.Input.PaddingH, th.Input.PaddingV)
		tb.cursorPos = tb.hitTestCursor(x, textRect)
		tb.selStart = tb.cursorPos
		tb.selEnd = tb.cursorPos

		// Reset cursor animation
		tb.cursorAnim.Reset()
		tb.cursorAnim.PlayForward()

		return true
	}

	tb.Blur()
	return false
}

// MouseUp handles mouse up
func (tb *TextBox) MouseUp(x, y float32, button int) bool {
	return false
}

// MouseMove handles mouse move
func (tb *TextBox) MouseMove(x, y float32) bool {
	wasHovered := tb.hovered
	tb.hovered = tb.bounds.Contains(x, y) && tb.enabled
	tb.SetStateFlag(StateHover, tb.hovered)

	// Update selection if dragging
	if tb.selStart >= 0 && tb.focused {
		theme := tb.GetTheme()
		textRect := tb.bounds.Shrink(theme.Input.PaddingH, theme.Input.PaddingV)
		tb.cursorPos = tb.hitTestCursor(x, textRect)
		tb.selEnd = tb.cursorPos
	}

	return tb.hovered || wasHovered
}

// KeyDown handles key down
func (tb *TextBox) KeyDown(key int, mods int) bool {
	if !tb.focused {
		return false
	}

	switch key {
	case 8: // Backspace
		if tb.cursorPos > 0 {
			tb.text = append(tb.text[:tb.cursorPos-1], tb.text[tb.cursorPos:]...)
			tb.cursorPos--
			tb.notifyChange()
		}
		return true

	case 46: // Delete
		if tb.cursorPos < len(tb.text) {
			tb.text = append(tb.text[:tb.cursorPos], tb.text[tb.cursorPos+1:]...)
			tb.notifyChange()
		}
		return true

	case 37: // Left
		if tb.cursorPos > 0 {
			tb.cursorPos--
		}
		return true

	case 39: // Right
		if tb.cursorPos < len(tb.text) {
			tb.cursorPos++
		}
		return true

	case 36: // Home
		tb.cursorPos = 0
		return true

	case 35: // End
		tb.cursorPos = len(tb.text)
		return true

	case 13: // Enter
		if tb.onSubmit != nil {
			tb.onSubmit(string(tb.text))
		}
		return true

	case 65: // A (Ctrl+A)
		if mods&2 == 2 {
			tb.selStart = 0
			tb.selEnd = len(tb.text)
			return true
		}
	}

	return false
}

// CharInput handles character input
func (tb *TextBox) CharInput(char rune) bool {
	if !tb.focused {
		return false
	}

	// Insert character at cursor position
	newText := make([]rune, 0, len(tb.text)+1)
	newText = append(newText, tb.text[:tb.cursorPos]...)
	newText = append(newText, char)
	newText = append(newText, tb.text[tb.cursorPos:]...)
	tb.text = newText
	tb.cursorPos++

	// Clear selection
	tb.selStart = -1
	tb.selEnd = -1

	// Reset cursor animation
	tb.cursorAnim.Reset()
	tb.cursorAnim.PlayForward()

	tb.notifyChange()
	return true
}

// notifyChange notifies the change handler
func (tb *TextBox) notifyChange() {
	if tb.onChange != nil {
		tb.onChange(string(tb.text))
	}
}

// Focus gives focus to the text box
func (tb *TextBox) Focus() {
	tb.BaseWidget.Focus()
	tb.focusAnim.PlayForward()
	tb.cursorAnim.PlayForward()
}

// Blur removes focus from the text box
func (tb *TextBox) Blur() {
	tb.BaseWidget.Blur()
	tb.focusAnim.PlayReverse()
	tb.cursorAnim.Stop()
}
