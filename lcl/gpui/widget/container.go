package widget

import "github.com/energye/examples/lcl/gpui/core/math"

type focusRegistrar interface {
	addFocusable(widget Widget)
	removeFocusable(widget Widget)
}

// Container is a generic widget node that owns children.
type Container struct {
	BaseWidget
	children          []Widget
	focus             *FocusManager
	clip              bool
	hoverChild        Widget
	pointerCapture    Widget
	pointerCaptureHit Widget
	pointerStart      math.Vec2
	pointerDragging   bool
}

// NewContainer creates an empty container.
func NewContainer() *Container {
	c := &Container{
		BaseWidget: NewBaseWidget(),
		children:   make([]Widget, 0),
		focus:      NewFocusManager(),
	}
	c.SetOwner(c)
	return c
}

// SetClip toggles clipping children to this container's bounds.
func (c *Container) SetClip(clip bool) {
	if c == nil {
		return
	}
	c.clip = clip
	c.Invalidate()
}

// Add appends a child widget.
func (c *Container) Add(child Widget) {
	if c == nil || child == nil {
		return
	}
	if owned, ok := child.(interface{ SetOwner(Widget) }); ok {
		owned.SetOwner(child)
	}
	child.SetParent(c)
	c.children = append(c.children, child)
	c.registerFocusable(child)
	c.Invalidate()
}

// Remove detaches a child widget.
func (c *Container) Remove(child Widget) {
	if c == nil || child == nil {
		return
	}
	for i, item := range c.children {
		if item != child {
			continue
		}
		child.SetParent(nil)
		c.children = append(c.children[:i], c.children[i+1:]...)
		c.unregisterFocusable(child)
		c.Invalidate()
		return
	}
}

// Children returns a copy of the child list.
func (c *Container) Children() []Widget {
	if c == nil {
		return nil
	}
	out := make([]Widget, len(c.children))
	copy(out, c.children)
	return out
}

// FocusManager returns the container focus manager.
func (c *Container) FocusManager() *FocusManager {
	if c == nil {
		return nil
	}
	return c.focus
}

// RefreshFocus rebuilds focus order from the current child tree.
func (c *Container) RefreshFocus() {
	if c == nil {
		return
	}
	current := Widget(nil)
	if c.focus != nil {
		current = c.focus.Current()
	}
	c.focus = NewFocusManager()
	for _, child := range c.children {
		c.registerFocusable(child)
	}
	if current != nil && current.Focusable() {
		c.focus.SetFocus(current)
	}
}

// Measure returns the maximum child extent, constrained by available space.
func (c *Container) Measure(ctx *Context, constraints Constraints) math.Vec2 {
	if c == nil {
		return math.Vec2{}
	}
	size := c.BaseWidget.Measure(ctx, constraints)
	for _, child := range c.children {
		if child == nil || !child.Visible() {
			continue
		}
		childSize := child.Measure(ctx, constraints)
		bounds := child.Bounds()
		if bounds.W <= 0 {
			bounds.W = childSize.X
		}
		if bounds.H <= 0 {
			bounds.H = childSize.Y
		}
		if right := bounds.X + bounds.W; right > size.X {
			size.X = right
		}
		if bottom := bounds.Y + bounds.H; bottom > size.Y {
			size.Y = bottom
		}
	}
	return ClampSize(size, constraints)
}

// Layout assigns bounds to the container and preserves child local positions.
func (c *Container) Layout(ctx *Context, rect math.Rect) {
	if c == nil {
		return
	}
	c.BaseWidget.Layout(ctx, rect)
	for _, child := range c.children {
		if child == nil {
			continue
		}
		bounds := child.Bounds()
		if bounds.W <= 0 || bounds.H <= 0 {
			size := child.Measure(ctx, Constraints{Max: math.NewVec2(rect.W, rect.H)})
			if bounds.W <= 0 {
				bounds.W = size.X
			}
			if bounds.H <= 0 {
				bounds.H = size.Y
			}
		}
		child.Layout(ctx, bounds)
	}
}

// Render renders all visible children in container-local coordinates.
func (c *Container) Render(ctx *Context) {
	if c == nil || ctx == nil || ctx.Renderer == nil || !c.Visible() {
		return
	}
	bounds := c.Bounds()
	ctx.Renderer.PushTransform(math.TranslationMatrix(bounds.X, bounds.Y, 0))
	if c.clip {
		ctx.Renderer.PushClip(math.NewRect(0, 0, bounds.W, bounds.H))
	}
	for _, child := range c.children {
		if child != nil && child.Visible() {
			child.Render(ctx)
		}
	}
	if c.clip {
		ctx.Renderer.PopClip()
	}
	ctx.Renderer.PopTransform()
}

// HitTest returns the deepest child hit by a point in parent-local coordinates.
func (c *Container) HitTest(point math.Vec2) Widget {
	if c == nil || !c.Visible() || !c.Enabled() {
		return nil
	}
	bounds := c.Bounds()
	if !bounds.Contains(point.X, point.Y) {
		return nil
	}
	local := math.NewVec2(point.X-bounds.X, point.Y-bounds.Y)
	for i := len(c.children) - 1; i >= 0; i-- {
		child := c.children[i]
		if child == nil {
			continue
		}
		if hit := child.HitTest(local); hit != nil {
			return hit
		}
	}
	return c
}

// HandleEvent routes input to topmost children and updates common states.
func (c *Container) HandleEvent(ctx *Context, event Event) bool {
	if c == nil || !c.Visible() || !c.Enabled() {
		return false
	}
	local := math.NewVec2(event.X-c.bounds.X, event.Y-c.bounds.Y)
	switch event.Type {
	case EventMouseDown:
		return c.dispatchPointer(ctx, event, local, true)
	case EventMouseUp:
		return c.dispatchPointer(ctx, event, local, false)
	case EventMouseMove:
		return c.dispatchPointer(ctx, event, local, false)
	case EventMouseWheel:
		return c.dispatchWheel(ctx, event, local)
	case EventDoubleClick:
		return c.dispatchPointer(ctx, event, local, true)
	case EventKeyDown, EventCharInput:
		if c.focus == nil {
			return false
		}
		focused := c.focus.Current()
		return focused != nil && focused.HandleEvent(ctx, event)
	default:
		return false
	}
}

func (c *Container) dispatchPointer(ctx *Context, event Event, point math.Vec2, focusOnHit bool) bool {
	if c.pointerCapture != nil && (event.Type == EventMouseMove || event.Type == EventMouseUp) {
		return c.dispatchCapturedPointer(ctx, event, point)
	}
	for i := len(c.children) - 1; i >= 0; i-- {
		child := c.children[i]
		if child == nil || !child.Visible() || !child.Enabled() {
			continue
		}
		hit := child.HitTest(point)
		child.SetStateFlag(StateHover, hit != nil && event.Type == EventMouseMove)
		if hit == nil {
			continue
		}
		if event.Type == EventMouseMove {
			c.updateHover(ctx, child, event, point)
		}
		if focusOnHit && hit.Focusable() && c.focus != nil {
			c.focus.SetFocus(hit)
		}
		childBounds := child.Bounds()
		childEvent := event
		childEvent.X = point.X
		childEvent.Y = point.Y
		childEvent.LocalX = point.X - childBounds.X
		childEvent.LocalY = point.Y - childBounds.Y
		if event.Type == EventMouseDown {
			hit.SetStateFlag(StateActive, true)
			c.pointerCapture = child
			c.pointerCaptureHit = hit
			c.pointerStart = point
			c.pointerDragging = false
		}
		if event.Type == EventMouseUp {
			hit.SetStateFlag(StateActive, false)
		}
		return child.HandleEvent(ctx, childEvent)
	}
	if event.Type == EventMouseMove {
		c.updateHover(ctx, nil, event, point)
	}
	return false
}

func (c *Container) dispatchCapturedPointer(ctx *Context, event Event, point math.Vec2) bool {
	child := c.pointerCapture
	if child == nil {
		return false
	}
	childBounds := child.Bounds()
	childEvent := event
	childEvent.X = point.X
	childEvent.Y = point.Y
	childEvent.LocalX = point.X - childBounds.X
	childEvent.LocalY = point.Y - childBounds.Y
	handled := child.HandleEvent(ctx, childEvent)
	if event.Type == EventMouseMove {
		handled = c.dispatchCapturedDrag(ctx, child, event, point) || handled
	}
	if event.Type == EventMouseUp {
		if c.pointerDragging {
			dragEnd := childEvent
			dragEnd.Type = EventDragEnd
			dragEnd.DeltaX = point.X - c.pointerStart.X
			dragEnd.DeltaY = point.Y - c.pointerStart.Y
			handled = child.HandleEvent(ctx, dragEnd) || handled
		}
		if c.pointerCaptureHit != nil {
			c.pointerCaptureHit.SetStateFlag(StateActive, false)
		}
		c.pointerCapture = nil
		c.pointerCaptureHit = nil
		c.pointerDragging = false
		return true
	}
	return handled
}

func (c *Container) dispatchCapturedDrag(ctx *Context, child Widget, event Event, point math.Vec2) bool {
	if child == nil {
		return false
	}
	dx := point.X - c.pointerStart.X
	dy := point.Y - c.pointerStart.Y
	if !c.pointerDragging && dx*dx+dy*dy < 16 {
		return false
	}
	childBounds := child.Bounds()
	dragEvent := event
	dragEvent.X = point.X
	dragEvent.Y = point.Y
	dragEvent.LocalX = point.X - childBounds.X
	dragEvent.LocalY = point.Y - childBounds.Y
	dragEvent.DeltaX = dx
	dragEvent.DeltaY = dy
	if !c.pointerDragging {
		c.pointerDragging = true
		dragStart := dragEvent
		dragStart.Type = EventDragStart
		child.HandleEvent(ctx, dragStart)
	}
	dragEvent.Type = EventDragMove
	return child.HandleEvent(ctx, dragEvent)
}

func (c *Container) dispatchWheel(ctx *Context, event Event, point math.Vec2) bool {
	for i := len(c.children) - 1; i >= 0; i-- {
		child := c.children[i]
		if child == nil || !child.Visible() || !child.Enabled() {
			continue
		}
		if child.HitTest(point) == nil {
			continue
		}
		childBounds := child.Bounds()
		childEvent := event
		childEvent.X = point.X
		childEvent.Y = point.Y
		childEvent.LocalX = point.X - childBounds.X
		childEvent.LocalY = point.Y - childBounds.Y
		return child.HandleEvent(ctx, childEvent)
	}
	return false
}

func (c *Container) updateHover(ctx *Context, child Widget, event Event, point math.Vec2) {
	if c == nil || c.hoverChild == child {
		return
	}
	if c.hoverChild != nil {
		c.hoverChild.SetStateFlag(StateHover, false)
		c.dispatchHoverEvent(ctx, c.hoverChild, event, point, EventMouseLeave)
	}
	c.hoverChild = child
	if c.hoverChild != nil {
		c.hoverChild.SetStateFlag(StateHover, true)
		c.dispatchHoverEvent(ctx, c.hoverChild, event, point, EventMouseEnter)
	}
}

func (c *Container) dispatchHoverEvent(ctx *Context, child Widget, event Event, point math.Vec2, eventType EventType) {
	if child == nil {
		return
	}
	childBounds := child.Bounds()
	childEvent := event
	childEvent.Type = eventType
	childEvent.X = point.X
	childEvent.Y = point.Y
	childEvent.LocalX = point.X - childBounds.X
	childEvent.LocalY = point.Y - childBounds.Y
	child.HandleEvent(ctx, childEvent)
}

func (c *Container) registerFocusable(widget Widget) {
	if c == nil || widget == nil || c.focus == nil {
		return
	}
	if widget.Focusable() {
		c.addFocusable(widget)
	}
	switch nested := widget.(type) {
	case *Container:
		for _, child := range nested.children {
			c.registerFocusable(child)
		}
	case *LayoutContainer:
		for _, child := range nested.children {
			c.registerFocusable(child)
		}
	}
}

func (c *Container) unregisterFocusable(widget Widget) {
	if c == nil || widget == nil || c.focus == nil {
		return
	}
	c.removeFocusable(widget)
	switch nested := widget.(type) {
	case *Container:
		for _, child := range nested.children {
			c.unregisterFocusable(child)
		}
	case *LayoutContainer:
		for _, child := range nested.children {
			c.unregisterFocusable(child)
		}
	}
}

func (c *Container) addFocusable(widget Widget) {
	if c == nil || c.focus == nil || widget == nil {
		return
	}
	c.focus.Add(widget)
}

func (c *Container) removeFocusable(widget Widget) {
	if c == nil || c.focus == nil || widget == nil {
		return
	}
	c.focus.Remove(widget)
}
