package widget

import "github.com/energye/examples/lcl/gpui/core/math"

// BaseWidget provides default widget lifecycle and state behavior.
type BaseWidget struct {
	owner       Widget
	bounds      math.Rect
	parent      Widget
	visible     bool
	enabled     bool
	focusable   bool
	focused     bool
	state       State
	preferred   math.Vec2
	invalidated bool
}

// NewBaseWidget creates a visible enabled base widget.
func NewBaseWidget() BaseWidget {
	return BaseWidget{
		visible: true,
		enabled: true,
	}
}

// SetPreferredSize stores the widget's default measured size.
func (w *BaseWidget) SetPreferredSize(size math.Vec2) {
	if w == nil {
		return
	}
	w.preferred = size
	w.Invalidate()
}

// PreferredSize returns the stored default measured size.
func (w *BaseWidget) PreferredSize() math.Vec2 {
	if w == nil {
		return math.Vec2{}
	}
	return w.preferred
}

// Bounds returns the widget bounds in parent-local coordinates.
func (w *BaseWidget) Bounds() math.Rect {
	if w == nil {
		return math.Rect{}
	}
	return w.bounds
}

// SetBounds assigns widget bounds in parent-local coordinates.
func (w *BaseWidget) SetBounds(rect math.Rect) {
	if w == nil {
		return
	}
	w.bounds = rect
	w.Invalidate()
}

// SetPos assigns the widget position in parent-local coordinates.
func (w *BaseWidget) SetPos(x, y float32) {
	if w == nil {
		return
	}
	w.bounds.X = x
	w.bounds.Y = y
	w.Invalidate()
}

// SetSize assigns the widget size.
func (w *BaseWidget) SetSize(width, height float32) {
	if w == nil {
		return
	}
	w.bounds.W = width
	w.bounds.H = height
	w.Invalidate()
}

// Parent returns the parent widget.
func (w *BaseWidget) Parent() Widget {
	if w == nil {
		return nil
	}
	return w.parent
}

// SetParent assigns the parent widget.
func (w *BaseWidget) SetParent(parent Widget) {
	if w == nil {
		return
	}
	w.parent = parent
}

// Visible reports whether the widget should render and receive pointer events.
func (w *BaseWidget) Visible() bool {
	return w != nil && w.visible
}

// SetVisible toggles visibility.
func (w *BaseWidget) SetVisible(visible bool) {
	if w == nil {
		return
	}
	w.visible = visible
	w.Invalidate()
}

// Enabled reports whether the widget can receive input.
func (w *BaseWidget) Enabled() bool {
	return w != nil && w.enabled
}

// SetEnabled toggles enabled state.
func (w *BaseWidget) SetEnabled(enabled bool) {
	if w == nil {
		return
	}
	w.enabled = enabled
	w.SetStateFlag(StateDisabled, !enabled)
}

// Focusable reports whether the widget can receive focus.
func (w *BaseWidget) Focusable() bool {
	return w != nil && w.focusable && w.enabled && w.visible
}

// SetFocusable toggles focusability.
func (w *BaseWidget) SetFocusable(focusable bool) {
	if w == nil {
		return
	}
	w.focusable = focusable
	owner := w.self()
	if owner == nil {
		return
	}
	parent, ok := w.parent.(focusRegistrar)
	if !ok {
		return
	}
	if focusable {
		parent.addFocusable(owner)
	} else {
		parent.removeFocusable(owner)
	}
}

// Focused reports whether this widget has focus.
func (w *BaseWidget) Focused() bool {
	return w != nil && w.focused
}

// Focus marks the widget focused.
func (w *BaseWidget) Focus() {
	if w == nil || !w.Focusable() {
		return
	}
	w.focused = true
	w.SetStateFlag(StateFocus, true)
}

// Blur removes focus from the widget.
func (w *BaseWidget) Blur() {
	if w == nil {
		return
	}
	w.focused = false
	w.SetStateFlag(StateFocus, false)
}

// State returns the current widget state flags.
func (w *BaseWidget) State() State {
	if w == nil {
		return StateNormal
	}
	return w.state
}

// SetState replaces the current widget state flags.
func (w *BaseWidget) SetState(state State) {
	if w == nil {
		return
	}
	w.state = state
	w.focused = state&StateFocus != 0
	w.enabled = state&StateDisabled == 0
	w.Invalidate()
}

// HasState reports whether all requested state flags are set.
func (w *BaseWidget) HasState(state State) bool {
	return w != nil && w.state&state == state
}

// SetStateFlag toggles a state flag.
func (w *BaseWidget) SetStateFlag(state State, enabled bool) {
	if w == nil {
		return
	}
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
	w.Invalidate()
}

// Invalidate marks the widget dirty.
func (w *BaseWidget) Invalidate() {
	if w == nil {
		return
	}
	w.invalidated = true
}

// Invalidated reports whether the widget is dirty.
func (w *BaseWidget) Invalidated() bool {
	return w != nil && w.invalidated
}

// ClearInvalidated clears the dirty flag.
func (w *BaseWidget) ClearInvalidated() {
	if w == nil {
		return
	}
	w.invalidated = false
}

// Measure returns the preferred size constrained by available space.
func (w *BaseWidget) Measure(ctx *Context, constraints Constraints) math.Vec2 {
	if w == nil {
		return math.Vec2{}
	}
	size := w.preferred
	if size.X <= 0 {
		size.X = w.bounds.W
	}
	if size.Y <= 0 {
		size.Y = w.bounds.H
	}
	return ClampSize(size, constraints)
}

// Layout assigns final bounds.
func (w *BaseWidget) Layout(ctx *Context, rect math.Rect) {
	w.SetBounds(rect)
}

// Render renders nothing by default.
func (w *BaseWidget) Render(ctx *Context) {}

// HitTest returns this widget if the point is inside parent-local bounds.
func (w *BaseWidget) HitTest(point math.Vec2) Widget {
	if w == nil || !w.visible || !w.enabled {
		return nil
	}
	if !w.bounds.Contains(point.X, point.Y) {
		return nil
	}
	return w.self()
}

// HandleEvent handles no events by default.
func (w *BaseWidget) HandleEvent(ctx *Context, event Event) bool {
	return false
}

// SetOwner binds this base node to the outer widget that embeds it.
func (w *BaseWidget) SetOwner(owner Widget) {
	if w == nil {
		return
	}
	w.owner = owner
}

func (w *BaseWidget) self() Widget {
	if w == nil {
		return nil
	}
	if w.owner != nil {
		return w.owner
	}
	return w
}
