package widget

// FocusManager owns focus order for a widget tree.
type FocusManager struct {
	widgets []Widget
	current Widget
}

// NewFocusManager creates an empty focus manager.
func NewFocusManager() *FocusManager {
	return &FocusManager{widgets: make([]Widget, 0)}
}

// Add registers a focusable widget.
func (fm *FocusManager) Add(widget Widget) {
	if fm == nil || widget == nil {
		return
	}
	for _, item := range fm.widgets {
		if item == widget {
			return
		}
	}
	fm.widgets = append(fm.widgets, widget)
}

// Remove unregisters a widget.
func (fm *FocusManager) Remove(widget Widget) {
	if fm == nil || widget == nil {
		return
	}
	for i, item := range fm.widgets {
		if item != widget {
			continue
		}
		fm.widgets = append(fm.widgets[:i], fm.widgets[i+1:]...)
		if fm.current == widget {
			widget.Blur()
			fm.current = nil
		}
		return
	}
}

// SetFocus focuses a widget.
func (fm *FocusManager) SetFocus(widget Widget) {
	if fm == nil || fm.current == widget {
		return
	}
	if widget != nil && !widget.Focusable() {
		return
	}
	if fm.current != nil {
		fm.current.Blur()
	}
	fm.current = widget
	if fm.current != nil {
		fm.current.Focus()
	}
}

// Current returns the focused widget.
func (fm *FocusManager) Current() Widget {
	if fm == nil {
		return nil
	}
	return fm.current
}

// Next focuses the next available widget.
func (fm *FocusManager) Next() {
	if fm == nil || len(fm.widgets) == 0 {
		return
	}
	start := -1
	for i, item := range fm.widgets {
		if item == fm.current {
			start = i
			break
		}
	}
	for offset := 1; offset <= len(fm.widgets); offset++ {
		next := fm.widgets[(start+offset+len(fm.widgets))%len(fm.widgets)]
		if next.Focusable() {
			fm.SetFocus(next)
			return
		}
	}
}

// Prev focuses the previous available widget.
func (fm *FocusManager) Prev() {
	if fm == nil || len(fm.widgets) == 0 {
		return
	}
	start := len(fm.widgets)
	for i, item := range fm.widgets {
		if item == fm.current {
			start = i
			break
		}
	}
	for offset := 1; offset <= len(fm.widgets); offset++ {
		next := fm.widgets[(start-offset+len(fm.widgets))%len(fm.widgets)]
		if next.Focusable() {
			fm.SetFocus(next)
			return
		}
	}
}

