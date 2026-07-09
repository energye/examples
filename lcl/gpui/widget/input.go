package widget

import (
	"strconv"

	"github.com/energye/examples/lcl/gpui/core/math"
	"github.com/energye/examples/lcl/gpui/render/font"
	"github.com/energye/examples/lcl/gpui/render/pipeline"
)

// InputType identifies the input type.
type InputType int

const (
	InputText     InputType = iota // Normal text input
	InputPassword                  // Password input (masked)
	InputSearch                    // Search input
	InputEmail                     // Email input
	InputNumber                    // Number input
)

// Input is a text input control.
type Input struct {
	ControlSurface
	text          string
	placeholder   string
	inputType     InputType
	maxLength     int
	readonly      bool
	allowClear    bool
	showCount     bool
	prefix        string
	suffix        string
	addonBefore   string
	addonAfter    string
	f             *font.Font
	font          *font.Font // Alias for f

	// Cursor state
	cursorPos     int
	selectionStart int
	selectionEnd   int

	// Callbacks
	onChange      func(text string)
	onSubmit      func(text string)
	onFocus       func()
	onBlur        func()
	onClear       func()

	// Internal state
	focused       bool
	dragging      bool
	dragStartPos  int
}

// InputProps stores mutable input properties.
type InputProps struct {
	Text        string
	Placeholder string
	InputType   InputType
	MaxLength   int
	Readonly    bool
	AllowClear  bool
	ShowCount   bool
	Prefix      string
	Suffix      string
	Font        *font.Font
	OnChange    func(text string)
	OnSubmit    func(text string)
	OnFocus     func()
	OnBlur      func()
	OnClear     func()
}

// NewInput creates a new input control.
func NewInput(placeholder string) *Input {
	i := &Input{
		ControlSurface: *NewControlSurface(),
		placeholder:    placeholder,
		inputType:      InputText,
	}
	i.SetOwner(i)
	i.interaction.SetTarget(i)
	i.SetFocusable(true)
	return i
}

// Text returns the current text.
func (i *Input) Text() string {
	if i == nil {
		return ""
	}
	return i.text
}

// SetText sets the text content.
func (i *Input) SetText(text string) {
	if i == nil {
		return
	}
	if i.maxLength > 0 && len(text) > i.maxLength {
		text = text[:i.maxLength]
	}
	if i.text == text {
		return
	}
	i.text = text
	i.cursorPos = len(i.text)
	i.clearSelection()
	i.Invalidate()
	if i.onChange != nil {
		i.onChange(i.text)
	}
}

// Placeholder returns the placeholder text.
func (i *Input) Placeholder() string {
	if i == nil {
		return ""
	}
	return i.placeholder
}

// SetPlaceholder sets the placeholder text.
func (i *Input) SetPlaceholder(placeholder string) {
	if i == nil || i.placeholder == placeholder {
		return
	}
	i.placeholder = placeholder
	i.Invalidate()
}

// InputType returns the input type.
func (i *Input) InputType() InputType {
	if i == nil {
		return InputText
	}
	return i.inputType
}

// SetInputType sets the input type.
func (i *Input) SetInputType(inputType InputType) {
	if i == nil || i.inputType == inputType {
		return
	}
	i.inputType = inputType
	i.Invalidate()
}

// MaxLength returns the max length.
func (i *Input) MaxLength() int {
	if i == nil {
		return 0
	}
	return i.maxLength
}

// SetMaxLength sets the maximum text length. 0 means unlimited.
func (i *Input) SetMaxLength(maxLength int) {
	if i == nil || i.maxLength == maxLength {
		return
	}
	i.maxLength = maxLength
	if i.maxLength > 0 && len(i.text) > i.maxLength {
		i.text = i.text[:i.maxLength]
		if i.cursorPos > len(i.text) {
			i.cursorPos = len(i.text)
		}
	}
	i.Invalidate()
}

// Readonly returns whether the input is readonly.
func (i *Input) Readonly() bool {
	return i != nil && i.readonly
}

// SetReadonly sets the readonly state.
func (i *Input) SetReadonly(readonly bool) {
	if i == nil || i.readonly == readonly {
		return
	}
	i.readonly = readonly
	i.Invalidate()
}

// AllowClear returns whether clear button is shown.
func (i *Input) AllowClear() bool {
	return i != nil && i.allowClear
}

// SetAllowClear toggles the clear button.
func (i *Input) SetAllowClear(allowClear bool) {
	if i == nil || i.allowClear == allowClear {
		return
	}
	i.allowClear = allowClear
	i.Invalidate()
}

// ShowCount returns whether character count is shown.
func (i *Input) ShowCount() bool {
	return i != nil && i.showCount
}

// SetShowCount toggles the character count display.
func (i *Input) SetShowCount(showCount bool) {
	if i == nil || i.showCount == showCount {
		return
	}
	i.showCount = showCount
	i.Invalidate()
}

// Prefix returns the prefix text.
func (i *Input) Prefix() string {
	if i == nil {
		return ""
	}
	return i.prefix
}

// SetPrefix sets the prefix text.
func (i *Input) SetPrefix(prefix string) {
	if i == nil || i.prefix == prefix {
		return
	}
	i.prefix = prefix
	i.Invalidate()
}

// Suffix returns the suffix text.
func (i *Input) Suffix() string {
	if i == nil {
		return ""
	}
	return i.suffix
}

// SetSuffix sets the suffix text.
func (i *Input) SetSuffix(suffix string) {
	if i == nil || i.suffix == suffix {
		return
	}
	i.suffix = suffix
	i.Invalidate()
}

// AddonBefore returns the addon before text.
func (i *Input) AddonBefore() string {
	if i == nil {
		return ""
	}
	return i.addonBefore
}

// SetAddonBefore sets the addon before text.
func (i *Input) SetAddonBefore(addon string) {
	if i == nil || i.addonBefore == addon {
		return
	}
	i.addonBefore = addon
	i.Invalidate()
}

// AddonAfter returns the addon after text.
func (i *Input) AddonAfter() string {
	if i == nil {
		return ""
	}
	return i.addonAfter
}

// SetAddonAfter sets the addon after text.
func (i *Input) SetAddonAfter(addon string) {
	if i == nil || i.addonAfter == addon {
		return
	}
	i.addonAfter = addon
	i.Invalidate()
}

// SetFont sets the font.
func (i *Input) SetFont(f *font.Font) {
	if i == nil || i.font == f {
		return
	}
	i.font = f
	i.Invalidate()
}

// SetOnChange sets the change callback.
func (i *Input) SetOnChange(handler func(text string)) {
	if i == nil {
		return
	}
	i.onChange = handler
}

// SetOnSubmit sets the submit callback (Enter key).
func (i *Input) SetOnSubmit(handler func(text string)) {
	if i == nil {
		return
	}
	i.onSubmit = handler
}

// SetOnFocus sets the focus callback.
func (i *Input) SetOnFocus(handler func()) {
	if i == nil {
		return
	}
	i.onFocus = handler
}

// SetOnBlur sets the blur callback.
func (i *Input) SetOnBlur(handler func()) {
	if i == nil {
		return
	}
	i.onBlur = handler
}

// SetOnClear sets the clear callback.
func (i *Input) SetOnClear(handler func()) {
	if i == nil {
		return
	}
	i.onClear = handler
}

// CursorPos returns the cursor position.
func (i *Input) CursorPos() int {
	if i == nil {
		return 0
	}
	return i.cursorPos
}

// SetCursorPos sets the cursor position.
func (i *Input) SetCursorPos(pos int) {
	if i == nil {
		return
	}
	if pos < 0 {
		pos = 0
	}
	if pos > len(i.text) {
		pos = len(i.text)
	}
	if i.cursorPos == pos {
		return
	}
	i.cursorPos = pos
	i.clearSelection()
	i.Invalidate()
}

// Selection returns the selection range (start, end).
func (i *Input) Selection() (start, end int) {
	if i == nil {
		return 0, 0
	}
	if i.selectionStart == i.selectionEnd {
		return i.cursorPos, i.cursorPos
	}
	if i.selectionStart < i.selectionEnd {
		return i.selectionStart, i.selectionEnd
	}
	return i.selectionEnd, i.selectionStart
}

// SetSelection sets the selection range.
func (i *Input) SetSelection(start, end int) {
	if i == nil {
		return
	}
	if start < 0 {
		start = 0
	}
	if end > len(i.text) {
		end = len(i.text)
	}
	i.selectionStart = start
	i.selectionEnd = end
	if start == end {
		i.cursorPos = start
	}
	i.Invalidate()
}

// SelectAll selects all text.
func (i *Input) SelectAll() {
	if i == nil {
		return
	}
	i.selectionStart = 0
	i.selectionEnd = len(i.text)
	i.cursorPos = len(i.text)
	i.Invalidate()
}

// HasSelection reports whether there is a selection.
func (i *Input) HasSelection() bool {
	return i != nil && i.selectionStart != i.selectionEnd
}

// DeleteSelection deletes the selected text.
func (i *Input) DeleteSelection() {
	if i == nil || !i.HasSelection() {
		return
	}
	start, end := i.Selection()
	i.text = i.text[:start] + i.text[end:]
	i.cursorPos = start
	i.clearSelection()
	i.Invalidate()
	if i.onChange != nil {
		i.onChange(i.text)
	}
}

// InsertText inserts text at the cursor position.
func (i *Input) InsertText(text string) {
	if i == nil || i.readonly {
		return
	}
	// Delete selection first
	if i.HasSelection() {
		i.DeleteSelection()
	}
	// Check max length
	if i.maxLength > 0 && len(i.text)+len(text) > i.maxLength {
		available := i.maxLength - len(i.text)
		if available <= 0 {
			return
		}
		text = text[:available]
	}
	// Insert at cursor position
	i.text = i.text[:i.cursorPos] + text + i.text[i.cursorPos:]
	i.cursorPos += len(text)
	i.Invalidate()
	if i.onChange != nil {
		i.onChange(i.text)
	}
}

// Clear clears the input text.
func (i *Input) Clear() {
	if i == nil || i.readonly {
		return
	}
	i.text = ""
	i.cursorPos = 0
	i.clearSelection()
	i.Invalidate()
	if i.onChange != nil {
		i.onChange(i.text)
	}
	if i.onClear != nil {
		i.onClear()
	}
}

// Measure returns the input size.
func (i *Input) Measure(ctx *Context, constraints Constraints) math.Vec2 {
	if i == nil {
		return math.Vec2{}
	}
	style := i.inputStyle(ctx)
	width := i.PreferredSize().X
	if width <= 0 {
		width = i.Bounds().W
	}
	if width <= 0 {
		width = style.Metrics.MinTouchSize.X * 4
	}
	height := i.PreferredSize().Y
	if height <= 0 {
		height = style.Metrics.Height
	}
	return ClampSize(math.NewVec2(width, height), constraints)
}

// Render draws the input.
func (i *Input) Render(ctx *Context) {
	if i == nil || ctx == nil || ctx.Renderer == nil || !i.Visible() {
		return
	}
	style := i.ResolveAnimatedControlStyle(ctx, i.inputStyle(ctx))
	bounds := i.Bounds()

	f := i.effectiveFont(ctx)
	if f == nil {
		return
	}

	// Draw addonBefore if present
	addonBeforeWidth := float32(0)
	if i.addonBefore != "" {
		addonBeforeWidth = f.TextWidth(i.addonBefore) + style.Metrics.PaddingH*2
		addonRect := math.NewRect(bounds.X, bounds.Y, addonBeforeWidth, bounds.H)
		ctx.Renderer.FillRoundRect(addonRect, style.Metrics.Radius, style.Palette.Background)
		ctx.Renderer.DrawText(i.addonBefore, addonRect.X+style.Metrics.PaddingH, addonRect.Y+(addonRect.H-f.LineHeight())/2, f, style.Palette.Text)
	}

	// Draw addonAfter if present
	addonAfterWidth := float32(0)
	if i.addonAfter != "" {
		addonAfterWidth = f.TextWidth(i.addonAfter) + style.Metrics.PaddingH*2
		addonRect := math.NewRect(bounds.X+bounds.W-addonAfterWidth, bounds.Y, addonAfterWidth, bounds.H)
		ctx.Renderer.FillRoundRect(addonRect, style.Metrics.Radius, style.Palette.Background)
		ctx.Renderer.DrawText(i.addonAfter, addonRect.X+style.Metrics.PaddingH, addonRect.Y+(addonRect.H-f.LineHeight())/2, f, style.Palette.Text)
	}

	// Adjust main input bounds for addons
	mainBounds := math.NewRect(bounds.X+addonBeforeWidth, bounds.Y, bounds.W-addonBeforeWidth-addonAfterWidth, bounds.H)
	ctx.Renderer.DrawBox(mainBounds, style.BoxStyle())
	i.RenderMotionOverlay(ctx, mainBounds)
	i.RenderFocusRing(ctx, mainBounds, style.Metrics.Radius)

	textRect := mainBounds.Shrink(style.Metrics.PaddingH, style.Metrics.PaddingV)

	// Draw prefix if present
	prefixWidth := float32(0)
	if i.prefix != "" {
		prefixWidth = f.TextWidth(i.prefix) + style.Metrics.IconGap
		ctx.Renderer.DrawText(i.prefix, textRect.X, textRect.Y+(textRect.H-f.LineHeight())/2, f, style.Palette.Placeholder)
	}

	// Draw suffix if present
	suffixWidth := float32(0)
	if i.suffix != "" {
		suffixWidth = f.TextWidth(i.suffix) + style.Metrics.IconGap
		suffixX := textRect.X + textRect.W - suffixWidth
		ctx.Renderer.DrawText(i.suffix, suffixX, textRect.Y+(textRect.H-f.LineHeight())/2, f, style.Palette.Placeholder)
	}

	// Adjust text rect for prefix/suffix
	textRect.X += prefixWidth
	textRect.W -= prefixWidth + suffixWidth

	// Draw selection highlight
	if i.HasSelection() {
		i.renderSelection(ctx, textRect, f)
	}

	// Draw text or placeholder
	if i.text == "" && i.placeholder != "" {
		// Draw placeholder
		ctx.Renderer.DrawTextInRect(i.placeholder, textRect, pipeline.TextOptions{
			Font:       f,
			Color:      style.Palette.Placeholder,
			Align:      pipeline.TextAlignLeft,
			MaxLines:   1,
			Ellipsis:   true,
			LineHeight: f.LineHeight(),
		})
	} else {
		// Draw text
		displayText := i.text
		if i.inputType == InputPassword {
			displayText = maskPassword(i.text)
		}
		ctx.Renderer.DrawTextInRect(displayText, textRect, pipeline.TextOptions{
			Font:       f,
			Color:      style.Palette.Text,
			Align:      pipeline.TextAlignLeft,
			MaxLines:   1,
			Ellipsis:   true,
			LineHeight: f.LineHeight(),
		})
	}

	// Draw cursor if focused
	if i.focused && !i.readonly {
		i.renderCursor(ctx, textRect, f, style)
	}

	// Draw clear button
	if i.allowClear && i.text != "" && i.HasState(StateHover) {
		i.renderClearButton(ctx, bounds, style)
	}

	// Draw character count
	if i.showCount && i.maxLength > 0 {
		i.renderCount(ctx, bounds, style, f)
	}
}

// HandleEvent handles input interaction.
func (i *Input) HandleEvent(ctx *Context, event Event) bool {
	if i == nil || !i.Enabled() {
		return false
	}
	if i.interaction == nil {
		i.interaction = NewInteractionController(i)
	}

	switch event.Type {
	case EventMouseDown:
		if !i.Focused() {
			i.Focus()
			if i.onFocus != nil {
				i.onFocus()
			}
		}
		// Set cursor position based on click
		i.setCursorFromClick(event)
		i.dragging = true
		i.dragStartPos = i.cursorPos
		return true
	case EventMouseUp:
		i.dragging = false
		return true
	case EventMouseMove:
		if i.dragging {
			i.setCursorFromClick(event)
			start := i.dragStartPos
			end := i.cursorPos
			if start > end {
				start, end = end, start
			}
			i.selectionStart = start
			i.selectionEnd = end
			i.Invalidate()
		}
		return true
	case EventDoubleClick:
		// Select word on double click
		i.selectWord()
		return true
	case EventKeyDown:
		return i.handleKeyEvent(event)
	case EventCharInput:
		if !i.readonly && event.Char > 0 {
			i.InsertText(string(event.Char))
			return true
		}
		return false
	default:
		return i.interaction.HandleEvent(ctx, event)
	}
}

// handleKeyEvent handles keyboard events.
func (i *Input) handleKeyEvent(event Event) bool {
	if i == nil {
		return false
	}

	// Handle Ctrl+A (Select All)
	if event.Mods != 0 && event.Char == 'a' {
		i.SelectAll()
		return true
	}

	// Handle Ctrl+C (Copy) - placeholder, needs platform support
	// Handle Ctrl+V (Paste) - placeholder, needs platform support
	// Handle Ctrl+X (Cut) - placeholder, needs platform support

	switch event.Key {
	case keyBackspace:
		if i.readonly {
			return false
		}
		if i.HasSelection() {
			i.DeleteSelection()
		} else if i.cursorPos > 0 {
			i.text = i.text[:i.cursorPos-1] + i.text[i.cursorPos:]
			i.cursorPos--
			i.Invalidate()
			if i.onChange != nil {
				i.onChange(i.text)
			}
		}
		return true
	case keyDelete:
		if i.readonly {
			return false
		}
		if i.HasSelection() {
			i.DeleteSelection()
		} else if i.cursorPos < len(i.text) {
			i.text = i.text[:i.cursorPos] + i.text[i.cursorPos+1:]
			i.Invalidate()
			if i.onChange != nil {
				i.onChange(i.text)
			}
		}
		return true
	case keyArrowLeft:
		if i.cursorPos > 0 {
			i.cursorPos--
			i.clearSelection()
			i.Invalidate()
		}
		return true
	case keyArrowRight:
		if i.cursorPos < len(i.text) {
			i.cursorPos++
			i.clearSelection()
			i.Invalidate()
		}
		return true
	case keyHome:
		i.cursorPos = 0
		i.clearSelection()
		i.Invalidate()
		return true
	case keyEnd:
		i.cursorPos = len(i.text)
		i.clearSelection()
		i.Invalidate()
		return true
	case keyEnter:
		if i.onSubmit != nil {
			i.onSubmit(i.text)
		}
		return true
	case keyEscape:
		i.Blur()
		if i.onBlur != nil {
			i.onBlur()
		}
		return true
	}
	return false
}

// Focus marks the input focused.
func (i *Input) Focus() {
	if i == nil {
		return
	}
	i.ControlSurface.Focus()
	i.focused = true
}

// Blur removes focus from the input.
func (i *Input) Blur() {
	if i == nil {
		return
	}
	i.ControlSurface.Blur()
	i.focused = false
	i.clearSelection()
}

// inputStyle resolves the input-specific style.
func (i *Input) inputStyle(ctx *Context) ControlStyle {
	base := i.ComponentBase
	base.variant = VariantOutlined
	// Use status from ComponentBase (supports StatusError, StatusWarning, StatusSuccess)
	style := base.ResolveControlStyle(ctx)
	return style
}

// effectiveFont returns the font to use.
func (i *Input) effectiveFont(ctx *Context) *font.Font {
	if i != nil && i.font != nil {
		return i.font
	}
	if ctx != nil {
		return ctx.Font
	}
	return nil
}

// clearSelection clears the selection.
func (i *Input) clearSelection() {
	if i == nil {
		return
	}
	i.selectionStart = 0
	i.selectionEnd = 0
}

// setCursorFromClick sets cursor position from mouse click.
func (i *Input) setCursorFromClick(event Event) {
	if i == nil {
		return
	}
	// Simple implementation: set cursor to end
	// A full implementation would calculate character position from click coordinates
	i.cursorPos = len(i.text)
	i.clearSelection()
	i.Invalidate()
}

// selectWord selects the word at the cursor position.
func (i *Input) selectWord() {
	if i == nil || len(i.text) == 0 {
		return
	}
	// Find word boundaries
	start := i.cursorPos
	end := i.cursorPos

	// Move start backwards to find word start
	for start > 0 && isWordChar(i.text[start-1]) {
		start--
	}
	// Move end forwards to find word end
	for end < len(i.text) && isWordChar(i.text[end]) {
		end++
	}

	i.selectionStart = start
	i.selectionEnd = end
	i.cursorPos = end
	i.Invalidate()
}

// renderSelection renders the selection highlight.
func (i *Input) renderSelection(ctx *Context, textRect math.Rect, f *font.Font) {
	if i == nil || !i.HasSelection() || f == nil {
		return
	}
	start, end := i.Selection()
	if start >= end {
		return
	}

	// Get display text (masked for password)
	displayText := i.text
	if i.inputType == InputPassword {
		displayText = maskPassword(i.text)
	}

	// Clamp to display text length
	if start > len(displayText) {
		start = len(displayText)
	}
	if end > len(displayText) {
		end = len(displayText)
	}

	// Calculate selection rectangle based on character positions
	startX := textRect.X
	if start > 0 {
		startX += f.TextWidth(displayText[:start])
	}
	endX := textRect.X
	if end > 0 {
		endX += f.TextWidth(displayText[:end])
	}

	selectionRect := math.NewRect(startX, textRect.Y, endX-startX, textRect.H)
	selectionColor := ctx.Tokens.Global.ColorPrimary.WithAlpha(0.3)
	ctx.Renderer.DrawSelectionHighlight(selectionRect, selectionColor)
}

// renderCursor renders the text cursor.
func (i *Input) renderCursor(ctx *Context, textRect math.Rect, f *font.Font, style ControlStyle) {
	if i == nil || f == nil {
		return
	}
	// Get display text (masked for password)
	displayText := i.text
	if i.inputType == InputPassword {
		displayText = maskPassword(i.text)
	}

	// Calculate cursor position
	cursorX := textRect.X
	if i.cursorPos > 0 && len(displayText) > 0 {
		// Clamp cursor position to display text length
		pos := i.cursorPos
		if pos > len(displayText) {
			pos = len(displayText)
		}
		cursorX += f.TextWidth(displayText[:pos])
	}
	cursorColor := style.Palette.Text
	ctx.Renderer.DrawTextCursor(cursorX, textRect.Y, textRect.H, 1.5, cursorColor)
}

// renderClearButton renders the clear button.
func (i *Input) renderClearButton(ctx *Context, bounds math.Rect, style ControlStyle) {
	if i == nil {
		return
	}
	buttonSize := style.Metrics.Height * 0.4
	buttonX := bounds.X + bounds.W - style.Metrics.PaddingH - buttonSize
	buttonY := bounds.Y + (bounds.H-buttonSize)/2
	buttonRect := math.NewRect(buttonX, buttonY, buttonSize, buttonSize)

	// Draw X icon
	iconSize := buttonSize * 0.4
	cx := buttonRect.X + buttonRect.W/2
	cy := buttonRect.Y + buttonRect.H/2
	ctx.Renderer.DrawLine(cx-iconSize, cy-iconSize, cx+iconSize, cy+iconSize, 1.5, style.Palette.Placeholder)
	ctx.Renderer.DrawLine(cx+iconSize, cy-iconSize, cx-iconSize, cy+iconSize, 1.5, style.Palette.Placeholder)
}

// renderCount renders the character count.
func (i *Input) renderCount(ctx *Context, bounds math.Rect, style ControlStyle, f *font.Font) {
	if i == nil || f == nil {
		return
	}
	countText := formatCount(len(i.text), i.maxLength)
	countWidth := f.TextWidth(countText)
	countX := bounds.X + bounds.W - style.Metrics.PaddingH - countWidth
	countY := bounds.Y + bounds.H - f.LineHeight() - style.Metrics.PaddingV/2
	ctx.Renderer.DrawText(countText, countX, countY, f, style.Palette.Placeholder)
}

// maskPassword masks a password string with asterisks.
func maskPassword(text string) string {
	result := make([]byte, len(text))
	for i := range result {
		result[i] = '*'
	}
	return string(result)
}

// isWordChar reports whether a byte is a word character.
func isWordChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
}

// formatCount formats the character count string.
func formatCount(current, max int) string {
	if max > 0 {
		return strconv.Itoa(current) + "/" + strconv.Itoa(max)
	}
	return strconv.Itoa(current)
}
