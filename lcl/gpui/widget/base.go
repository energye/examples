// Package widget provides the base widget interface and implementation
package widget

import (
	"github.com/energye/examples/lcl/gpui/core/math"
	"github.com/energye/examples/lcl/gpui/render/pipeline"
	"github.com/energye/examples/lcl/gpui/style/theme"
)

// WidgetState stores common widget state flags.
type WidgetState uint32

const (
	StateNormal WidgetState = 0
	StateHover  WidgetState = 1 << iota
	StateActive
	StateFocus
	StateDisabled
	StateSelected
	StateChecked
	StateLoading
	StateError
	StateWarning
	StateSuccess
)

// EventType identifies a high-level UI event.
type EventType int

const (
	EventMouseDown EventType = iota
	EventMouseUp
	EventMouseMove
	EventKeyDown
	EventCharInput
)

// UIEvent is the common event payload used by the widget lifecycle.
type UIEvent struct {
	Type   EventType
	X, Y   float32
	Button int
	Key    int
	Mods   int
	Char   rune
}

// Widget is the base interface for all widgets
type Widget interface {
	// Bounds returns the widget bounds
	Bounds() math.Rect
	SetBounds(rect math.Rect)

	// Position
	X() float32
	Y() float32
	SetPos(x, y float32)

	// Size
	Width() float32
	Height() float32
	SetSize(w, h float32)

	// Visibility
	Visible() bool
	SetVisible(visible bool)

	// Enable state
	Enabled() bool
	SetEnabled(enabled bool)

	// Parent
	Parent() Widget
	SetParent(parent Widget)

	// Rendering
	Measure(available math.Vec2) math.Vec2
	Layout(rect math.Rect)
	HitTest(x, y float32) bool
	Render(renderer *pipeline.Renderer)
	HandleEvent(event UIEvent) bool

	// Invalidation
	Invalidate()
	Invalidated() bool

	// Mouse events
	MouseDown(x, y float32, button int) bool
	MouseUp(x, y float32, button int) bool
	MouseMove(x, y float32) bool

	// Keyboard events
	KeyDown(key int, mods int) bool
	CharInput(char rune) bool

	// Focus
	Focusable() bool
	Focused() bool
	Focus()
	Blur()

	// State
	State() WidgetState
	SetState(state WidgetState)
	HasState(state WidgetState) bool
	SetStateFlag(state WidgetState, enabled bool)

	// Theme
	GetTheme() *theme.Theme
}

// BaseWidget provides default implementations for Widget
type BaseWidget struct {
	bounds      math.Rect
	visible     bool
	enabled     bool
	parent      Widget
	focused     bool
	invalidated bool
	state       WidgetState
	theme       *theme.Theme
}

// NewBaseWidget creates a new base widget
func NewBaseWidget() BaseWidget {
	return BaseWidget{
		visible: true,
		enabled: true,
		theme:   theme.GetTheme(),
	}
}

// Bounds returns the widget bounds
func (w *BaseWidget) Bounds() math.Rect {
	return w.bounds
}

// SetBounds sets the widget bounds
func (w *BaseWidget) SetBounds(rect math.Rect) {
	w.bounds = rect
}

// X returns the X position
func (w *BaseWidget) X() float32 {
	return w.bounds.X
}

// Y returns the Y position
func (w *BaseWidget) Y() float32 {
	return w.bounds.Y
}

// SetPos sets the position
func (w *BaseWidget) SetPos(x, y float32) {
	w.bounds.X = x
	w.bounds.Y = y
}

// Width returns the width
func (w *BaseWidget) Width() float32 {
	return w.bounds.W
}

// Height returns the height
func (w *BaseWidget) Height() float32 {
	return w.bounds.H
}

// SetSize sets the size
func (w *BaseWidget) SetSize(width, height float32) {
	w.bounds.W = width
	w.bounds.H = height
}

// Visible returns visibility
func (w *BaseWidget) Visible() bool {
	return w.visible
}

// SetVisible sets visibility
func (w *BaseWidget) SetVisible(visible bool) {
	w.visible = visible
}

// Enabled returns enabled state
func (w *BaseWidget) Enabled() bool {
	return w.enabled
}

// SetEnabled sets enabled state
func (w *BaseWidget) SetEnabled(enabled bool) {
	w.enabled = enabled
	w.SetStateFlag(StateDisabled, !enabled)
}

// Parent returns the parent widget
func (w *BaseWidget) Parent() Widget {
	return w.parent
}

// SetParent sets the parent widget
func (w *BaseWidget) SetParent(parent Widget) {
	w.parent = parent
}

// Invalidate marks the widget as needing redraw
func (w *BaseWidget) Invalidate() {
	w.invalidated = true
}

// Invalidated returns whether the widget needs redraw
func (w *BaseWidget) Invalidated() bool {
	return w.invalidated
}

// Focusable returns whether the widget can receive focus
func (w *BaseWidget) Focusable() bool {
	return false
}

// Focused returns whether the widget has focus
func (w *BaseWidget) Focused() bool {
	return w.focused
}

// Focus gives focus to the widget
func (w *BaseWidget) Focus() {
	w.focused = true
	w.SetStateFlag(StateFocus, true)
}

// Blur removes focus from the widget
func (w *BaseWidget) Blur() {
	w.focused = false
	w.SetStateFlag(StateFocus, false)
}

// GetTheme returns the current theme
func (w *BaseWidget) GetTheme() *theme.Theme {
	return w.theme
}

// Measure returns the widget's desired size.
func (w *BaseWidget) Measure(available math.Vec2) math.Vec2 {
	return math.NewVec2(w.bounds.W, w.bounds.H)
}

// Layout assigns final bounds to the widget.
func (w *BaseWidget) Layout(rect math.Rect) {
	w.SetBounds(rect)
}

// HitTest checks whether a point is inside this widget.
func (w *BaseWidget) HitTest(x, y float32) bool {
	return w.visible && w.enabled && w.bounds.Contains(x, y)
}

// Render renders the widget (default implementation does nothing)
func (w *BaseWidget) Render(renderer *pipeline.Renderer) {
	// Override in subclasses
}

// HandleEvent handles a generic UI event.
func (w *BaseWidget) HandleEvent(event UIEvent) bool {
	return false
}

func dispatchLegacyEvent(widget Widget, event UIEvent) bool {
	switch event.Type {
	case EventMouseDown:
		return widget.MouseDown(event.X, event.Y, event.Button)
	case EventMouseUp:
		return widget.MouseUp(event.X, event.Y, event.Button)
	case EventMouseMove:
		return widget.MouseMove(event.X, event.Y)
	case EventKeyDown:
		return widget.KeyDown(event.Key, event.Mods)
	case EventCharInput:
		return widget.CharInput(event.Char)
	default:
		return false
	}
}

// State returns the current widget state flags.
func (w *BaseWidget) State() WidgetState {
	return w.state
}

// SetState replaces the current widget state flags.
func (w *BaseWidget) SetState(state WidgetState) {
	w.state = state
	w.focused = state&StateFocus != 0
	w.enabled = state&StateDisabled == 0
}

// HasState reports whether all requested state flags are set.
func (w *BaseWidget) HasState(state WidgetState) bool {
	return w.state&state == state
}

// SetStateFlag toggles a state flag.
func (w *BaseWidget) SetStateFlag(state WidgetState, enabled bool) {
	if enabled {
		w.state |= state
	} else {
		w.state &^= state
	}
	if state&StateDisabled != 0 {
		w.enabled = !enabled
	}
	if state&StateFocus != 0 {
		w.focused = enabled
	}
}

// MouseDown handles mouse down (default returns false)
func (w *BaseWidget) MouseDown(x, y float32, button int) bool {
	return false
}

// MouseUp handles mouse up (default returns false)
func (w *BaseWidget) MouseUp(x, y float32, button int) bool {
	return false
}

// MouseMove handles mouse move (default returns false)
func (w *BaseWidget) MouseMove(x, y float32) bool {
	return false
}

// KeyDown handles key down (default returns false)
func (w *BaseWidget) KeyDown(key int, mods int) bool {
	return false
}

// CharInput handles character input (default returns false)
func (w *BaseWidget) CharInput(char rune) bool {
	return false
}
