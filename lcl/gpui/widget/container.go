// Package widget provides UI widgets
package widget

import (
	"github.com/energye/examples/lcl/gpui/core/math"
	"github.com/energye/examples/lcl/gpui/render/pipeline"
)

// Container is a widget that contains other widgets
type Container struct {
	BaseWidget
	children []Widget
	focusMgr *FocusManager
}

// NewContainer creates a new container
func NewContainer() *Container {
	return &Container{
		BaseWidget: NewBaseWidget(),
		children:   make([]Widget, 0),
		focusMgr:   NewFocusManager(),
	}
}

// FocusManager returns the focus manager
func (c *Container) FocusManager() *FocusManager {
	return c.focusMgr
}

// Add adds a child widget
func (c *Container) Add(child Widget) {
	child.SetParent(c)
	c.children = append(c.children, child)
	if child.Focusable() && c.focusMgr != nil {
		c.focusMgr.Add(child)
	}
}

// Remove removes a child widget
func (c *Container) Remove(child Widget) {
	for i, ch := range c.children {
		if ch == child {
			child.SetParent(nil)
			c.children = append(c.children[:i], c.children[i+1:]...)
			if c.focusMgr != nil {
				c.focusMgr.Remove(child)
			}
			return
		}
	}
}

// Children returns the child widgets
func (c *Container) Children() []Widget {
	return c.children
}

// Render renders the container and its children
func (c *Container) Render(renderer *pipeline.Renderer) {
	if !c.visible {
		return
	}

	renderer.PushTransform(math.TranslationMatrix(c.bounds.X, c.bounds.Y, 0))
	defer renderer.PopTransform()

	// Render children
	for _, child := range c.children {
		child.Render(renderer)
	}
}

// HandleEvent routes events to children using container-local coordinates.
func (c *Container) HandleEvent(event UIEvent) bool {
	switch event.Type {
	case EventMouseDown:
		localEvent := event
		localEvent.X -= c.bounds.X
		localEvent.Y -= c.bounds.Y

		for i := len(c.children) - 1; i >= 0; i-- {
			child := c.children[i]
			if !child.Visible() || !child.Enabled() {
				continue
			}
			if !child.Bounds().Contains(localEvent.X, localEvent.Y) {
				continue
			}
			if child.Focusable() && c.focusMgr != nil {
				c.focusMgr.SetFocus(child)
			}
			if child.HandleEvent(localEvent) || dispatchLegacyEvent(child, localEvent) {
				return true
			}
		}
		return false

	case EventMouseUp:
		localEvent := event
		localEvent.X -= c.bounds.X
		localEvent.Y -= c.bounds.Y

		handled := false
		for i := len(c.children) - 1; i >= 0; i-- {
			child := c.children[i]
			if !child.Visible() || !child.Enabled() {
				continue
			}
			if child.HandleEvent(localEvent) || dispatchLegacyEvent(child, localEvent) {
				handled = true
			}
		}
		return handled

	case EventMouseMove:
		localEvent := event
		localEvent.X -= c.bounds.X
		localEvent.Y -= c.bounds.Y

		handled := false
		for i := len(c.children) - 1; i >= 0; i-- {
			child := c.children[i]
			if !child.Visible() || !child.Enabled() {
				continue
			}
			if child.HandleEvent(localEvent) || dispatchLegacyEvent(child, localEvent) {
				handled = true
			}
		}
		return handled

	case EventKeyDown, EventCharInput:
		for _, child := range c.children {
			if child.Focused() {
				return child.HandleEvent(event) || dispatchLegacyEvent(child, event)
			}
		}
	}

	return false
}

// MouseDown handles mouse down
func (c *Container) MouseDown(x, y float32, button int) bool {
	return c.HandleEvent(UIEvent{Type: EventMouseDown, X: x, Y: y, Button: button})
}

// MouseUp handles mouse up
func (c *Container) MouseUp(x, y float32, button int) bool {
	return c.HandleEvent(UIEvent{Type: EventMouseUp, X: x, Y: y, Button: button})
}

// MouseMove handles mouse move
func (c *Container) MouseMove(x, y float32) bool {
	return c.HandleEvent(UIEvent{Type: EventMouseMove, X: x, Y: y})
}

// KeyDown handles key down
func (c *Container) KeyDown(key int, mods int) bool {
	return c.HandleEvent(UIEvent{Type: EventKeyDown, Key: key, Mods: mods})
}

// CharInput handles character input
func (c *Container) CharInput(char rune) bool {
	return c.HandleEvent(UIEvent{Type: EventCharInput, Char: char})
}

// Focusable returns false (container itself is not focusable)
func (c *Container) Focusable() bool {
	return false
}

// FocusManager manages focus across widgets
type FocusManager struct {
	widgets []Widget
	current Widget
}

// NewFocusManager creates a new focus manager
func NewFocusManager() *FocusManager {
	return &FocusManager{
		widgets: make([]Widget, 0),
	}
}

// Add adds a widget to the focus manager
func (fm *FocusManager) Add(widget Widget) {
	fm.widgets = append(fm.widgets, widget)
}

// Remove removes a widget from the focus manager
func (fm *FocusManager) Remove(widget Widget) {
	for i, w := range fm.widgets {
		if w == widget {
			fm.widgets = append(fm.widgets[:i], fm.widgets[i+1:]...)
			if fm.current == widget {
				fm.current = nil
			}
			return
		}
	}
}

// SetFocus sets focus to a widget
func (fm *FocusManager) SetFocus(widget Widget) {
	if fm.current == widget {
		return
	}

	if fm.current != nil {
		fm.current.Blur()
	}

	fm.current = widget
	if widget != nil {
		widget.Focus()
	}
}

// Current returns the currently focused widget
func (fm *FocusManager) Current() Widget {
	return fm.current
}

// Next focuses the next focusable widget
func (fm *FocusManager) Next() {
	if len(fm.widgets) == 0 {
		return
	}

	// Find current index
	currentIdx := -1
	for i, w := range fm.widgets {
		if w == fm.current {
			currentIdx = i
			break
		}
	}

	// Find next focusable widget
	for i := 1; i <= len(fm.widgets); i++ {
		idx := (currentIdx + i) % len(fm.widgets)
		if fm.widgets[idx].Focusable() && fm.widgets[idx].Enabled() {
			fm.SetFocus(fm.widgets[idx])
			return
		}
	}
}

// Prev focuses the previous focusable widget
func (fm *FocusManager) Prev() {
	if len(fm.widgets) == 0 {
		return
	}

	// Find current index
	currentIdx := len(fm.widgets)
	for i, w := range fm.widgets {
		if w == fm.current {
			currentIdx = i
			break
		}
	}

	// Find previous focusable widget
	for i := 1; i <= len(fm.widgets); i++ {
		idx := (currentIdx - i + len(fm.widgets)) % len(fm.widgets)
		if fm.widgets[idx].Focusable() && fm.widgets[idx].Enabled() {
			fm.SetFocus(fm.widgets[idx])
			return
		}
	}
}

// Anchor represents widget anchor behavior
type Anchor int

const (
	AnchorLeftTop     Anchor = iota // Fixed position (default)
	AnchorRightTop                  // Follow right edge
	AnchorLeftBottom                // Follow bottom edge
	AnchorRightBottom               // Follow right and bottom edges
	AnchorAll                       // Stretch to fill
)

// AnchoredContainer is a container with anchor support
type AnchoredContainer struct {
	Container
}

// NewAnchoredContainer creates a new anchored container
func NewAnchoredContainer() *AnchoredContainer {
	return &AnchoredContainer{
		Container: *NewContainer(),
	}
}

// UpdateLayout updates child layouts based on size change
func (ac *AnchoredContainer) UpdateLayout(oldW, oldH, newW, newH float32) {
	for _, child := range ac.children {
		// Get anchor (default is left-top)
		anchor := AnchorLeftTop
		if anchorWidget, ok := child.(interface{ Anchor() Anchor }); ok {
			anchor = anchorWidget.Anchor()
		}

		bounds := child.Bounds()
		newBounds := bounds

		dx := newW - oldW
		dy := newH - oldH

		switch anchor {
		case AnchorLeftTop:
			// No change

		case AnchorRightTop:
			newBounds.X += dx

		case AnchorLeftBottom:
			newBounds.Y += dy

		case AnchorRightBottom:
			newBounds.X += dx
			newBounds.Y += dy

		case AnchorAll:
			newBounds.W += dx
			newBounds.H += dy
		}

		child.SetBounds(newBounds)
	}
}
