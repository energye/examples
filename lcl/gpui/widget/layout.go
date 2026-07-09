package widget

import (
	coremath "github.com/energye/examples/lcl/gpui/core/math"
	"github.com/energye/examples/lcl/gpui/layout"
)

// LayoutContainer applies the shared layout engine to real widget children.
type LayoutContainer struct {
	Container
	Style        layout.Style
	childStyles  map[Widget]layout.Style
	scroll       coremath.Vec2
	contentSize  coremath.Vec2
	cachedResult *layout.Result
	cacheValid   bool
	onScroll     func(x, y float32) // Scroll position change callback

	// Virtual scrolling support
	virtualScroll    bool    // Enable virtual scrolling
	itemHeight       float32 // Fixed item height for virtual scrolling (0 = variable)
	visibleStart     int     // First visible item index
	visibleEnd       int     // Last visible item index
	totalItems       int     // Total number of items
	onVisibleChanged func(start, end int) // Callback when visible range changes
}

// NewLayoutContainer creates a generic layout-backed container.
func NewLayoutContainer(style layout.Style) *LayoutContainer {
	c := &LayoutContainer{
		Container:   *NewContainer(),
		Style:       style,
		childStyles: make(map[Widget]layout.Style),
	}
	c.SetOwner(c)
	return c
}

// NewFlex creates a flex layout container.
func NewFlex(direction layout.Direction, gap float32) *LayoutContainer {
	return NewLayoutContainer(layout.Style{Direction: direction, Gap: gap})
}

// NewSpace creates a flex container that only spaces its children.
func NewSpace(direction layout.Direction, gap float32) *LayoutContainer {
	return NewLayoutContainer(layout.Style{Direction: direction, Gap: gap})
}

// NewAlign creates a one-axis flex alignment container.
func NewAlign(align layout.Align, justify layout.Justify) *LayoutContainer {
	return NewLayoutContainer(layout.Style{Direction: layout.Row, Align: align, Justify: justify})
}

// NewWrap creates a row flex container with wrapping enabled.
func NewWrap(gap float32) *LayoutContainer {
	return NewLayoutContainer(layout.Style{Direction: layout.Row, Gap: gap, Wrap: true})
}

// NewGrid creates a grid layout container.
func NewGrid(columns []layout.Value, gap float32) *LayoutContainer {
	return NewLayoutContainer(layout.Style{GridColumns: columns, Gap: gap})
}

// NewScrollArea creates a clipped scroll layout container.
func NewScrollArea(style layout.Style) *LayoutContainer {
	style.OverflowX = layout.OverflowScroll
	style.OverflowY = layout.OverflowScroll
	return NewLayoutContainer(style)
}

// Add appends a child with the default layout style.
func (c *LayoutContainer) Add(child Widget) {
	c.AddLayout(child, layout.Style{})
}

// AddLayout appends a child with explicit layout style.
func (c *LayoutContainer) AddLayout(child Widget, style layout.Style) {
	if c == nil || child == nil {
		return
	}
	if owned, ok := child.(interface{ SetOwner(Widget) }); ok {
		owned.SetOwner(child)
	}
	child.SetParent(c)
	c.children = append(c.children, child)
	c.childStyles[child] = style
	c.registerFocusable(child)
	c.Invalidate()
}

// Remove detaches a child widget.
func (c *LayoutContainer) Remove(child Widget) {
	if c == nil || child == nil {
		return
	}
	for i, item := range c.children {
		if item != child {
			continue
		}
		child.SetParent(nil)
		delete(c.childStyles, child)
		c.children = append(c.children[:i], c.children[i+1:]...)
		c.unregisterFocusable(child)

		// Clean up pointer capture state
		if c.pointerCapture == child {
			c.pointerCapture = nil
			c.pointerCaptureHit = nil
			c.pointerDragging = false
		}
		// Clean up hover state
		if c.hoverChild == child {
			c.hoverChild = nil
		}

		c.Invalidate()
		return
	}
}

// SetChildStyle updates a child's layout style.
func (c *LayoutContainer) SetChildStyle(child Widget, style layout.Style) {
	if c == nil || child == nil {
		return
	}
	c.childStyles[child] = style
	c.Invalidate()
}

// ChildStyle returns a child's layout style.
func (c *LayoutContainer) ChildStyle(child Widget) layout.Style {
	if c == nil || child == nil {
		return layout.Style{}
	}
	return c.childStyles[child]
}

// ContentSize returns the laid out content size before viewport clipping.
func (c *LayoutContainer) ContentSize() coremath.Vec2 {
	if c == nil {
		return coremath.Vec2{}
	}
	return c.contentSize
}

// Scroll returns the current scroll offset.
func (c *LayoutContainer) Scroll() coremath.Vec2 {
	if c == nil {
		return coremath.Vec2{}
	}
	return c.scroll
}

// SetScroll updates scroll offset and clamps it to the current content.
func (c *LayoutContainer) SetScroll(x, y float32) {
	if c == nil {
		return
	}
	oldX, oldY := c.scroll.X, c.scroll.Y
	c.scroll = coremath.NewVec2(x, y)
	c.clampScroll()
	c.updateVisibleRange()
	c.Invalidate()
	// Fire callback if scroll position changed
	if c.onScroll != nil && (c.scroll.X != oldX || c.scroll.Y != oldY) {
		c.onScroll(c.scroll.X, c.scroll.Y)
	}
}

// ScrollTo adjusts scroll offset so a content-local rect becomes visible.
func (c *LayoutContainer) ScrollTo(rect coremath.Rect) {
	if c == nil {
		return
	}
	oldX, oldY := c.scroll.X, c.scroll.Y
	viewport := c.Bounds()
	if rect.X < c.scroll.X {
		c.scroll.X = rect.X
	} else if rect.X+rect.W > c.scroll.X+viewport.W {
		c.scroll.X = rect.X + rect.W - viewport.W
	}
	if rect.Y < c.scroll.Y {
		c.scroll.Y = rect.Y
	} else if rect.Y+rect.H > c.scroll.Y+viewport.H {
		c.scroll.Y = rect.Y + rect.H - viewport.H
	}
	c.clampScroll()
	c.updateVisibleRange()
	c.Invalidate()
	// Fire callback if scroll position changed
	if c.onScroll != nil && (c.scroll.X != oldX || c.scroll.Y != oldY) {
		c.onScroll(c.scroll.X, c.scroll.Y)
	}
}

// SetOnScroll sets the scroll position change callback.
func (c *LayoutContainer) SetOnScroll(handler func(x, y float32)) {
	if c == nil {
		return
	}
	c.onScroll = handler
}

// SetVirtualScroll enables virtual scrolling for large lists.
// itemHeight is the fixed height of each item (0 = variable height).
// totalItems is the total number of items in the list.
func (c *LayoutContainer) SetVirtualScroll(itemHeight float32, totalItems int) {
	if c == nil {
		return
	}
	c.virtualScroll = true
	c.itemHeight = itemHeight
	c.totalItems = totalItems
	c.updateVisibleRange()
}

// SetOnVisibleChanged sets the callback when visible range changes.
func (c *LayoutContainer) SetOnVisibleChanged(handler func(start, end int)) {
	if c == nil {
		return
	}
	c.onVisibleChanged = handler
}

// VisibleRange returns the current visible item range.
func (c *LayoutContainer) VisibleRange() (start, end int) {
	if c == nil {
		return 0, 0
	}
	return c.visibleStart, c.visibleEnd
}

// TotalItems returns the total number of items.
func (c *LayoutContainer) TotalItems() int {
	if c == nil {
		return 0
	}
	return c.totalItems
}

// updateVisibleRange calculates which items are visible based on scroll position.
func (c *LayoutContainer) updateVisibleRange() {
	if c == nil || !c.virtualScroll {
		return
	}

	bounds := c.Bounds()
	if c.itemHeight <= 0 {
		// Variable height - use content size / total items as estimate
		if c.totalItems > 0 && c.contentSize.Y > 0 {
			c.itemHeight = c.contentSize.Y / float32(c.totalItems)
		} else {
			return
		}
	}

	// Calculate visible range with buffer
	bufferItems := 3
	start := int(c.scroll.Y / c.itemHeight) - bufferItems
	if start < 0 {
		start = 0
	}
	end := int((c.scroll.Y + bounds.H) / c.itemHeight) + bufferItems
	if end > c.totalItems {
		end = c.totalItems
	}

	// Fire callback if range changed
	if start != c.visibleStart || end != c.visibleEnd {
		c.visibleStart = start
		c.visibleEnd = end
		if c.onVisibleChanged != nil {
			c.onVisibleChanged(start, end)
		}
	}
}

// IsVirtualScrollEnabled reports whether virtual scrolling is enabled.
func (c *LayoutContainer) IsVirtualScrollEnabled() bool {
	return c != nil && c.virtualScroll
}

// Invalidate clears the layout cache.
func (c *LayoutContainer) Invalidate() {
	if c == nil {
		return
	}
	c.cacheValid = false
	c.Container.Invalidate()
}

// Measure computes the container size using the shared layout engine.
func (c *LayoutContainer) Measure(ctx *Context, constraints Constraints) coremath.Vec2 {
	if c == nil {
		return coremath.Vec2{}
	}
	available := constraints.Max
	if available.X <= 0 {
		available.X = c.Bounds().W
	}
	if available.Y <= 0 {
		available.Y = c.Bounds().H
	}
	node := c.layoutNode(ctx)
	result := layout.Compute(node, available)
	c.cachedResult = &result
	c.cacheValid = true
	return ClampSize(coremath.NewVec2(result.Bounds.W, result.Bounds.H), constraints)
}

// Layout computes and applies child bounds in container-local coordinates.
func (c *LayoutContainer) Layout(ctx *Context, rect coremath.Rect) {
	if c == nil {
		return
	}
	c.BaseWidget.Layout(ctx, rect)

	// Virtual scrolling: only layout visible children
	if c.virtualScroll && c.itemHeight > 0 && c.totalItems > 0 {
		c.layoutVirtual(ctx, rect)
		return
	}

	var result layout.Result
	if c.cacheValid && c.cachedResult != nil {
		result = *c.cachedResult
	} else {
		node := c.layoutNode(ctx)
		node.Style.Width = layout.Px(rect.W)
		node.Style.Height = layout.Px(rect.H)
		result = layout.Compute(node, coremath.NewVec2(rect.W, rect.H))
	}
	c.contentSize = result.ContentSize
	c.clampScroll()
	c.updateVisibleRange()

	for i, child := range c.children {
		if child == nil || i >= len(result.Children) {
			continue
		}
		child.Layout(ctx, result.Children[i].Bounds)
	}
}

// layoutVirtual performs layout for virtual scrolling mode.
func (c *LayoutContainer) layoutVirtual(ctx *Context, rect coremath.Rect) {
	if c == nil || !c.virtualScroll || c.itemHeight <= 0 || c.totalItems <= 0 {
		return
	}

	// Calculate total content size
	c.contentSize = coremath.NewVec2(rect.W, float32(c.totalItems)*c.itemHeight)
	c.clampScroll()
	c.updateVisibleRange()

	// Only layout visible children
	visibleCount := c.visibleEnd - c.visibleStart
	if visibleCount <= 0 {
		return
	}

	// Create layout only for visible children
	node := layout.NewNode(c.Style)
	node.Style.Width = layout.Px(rect.W)
	node.Style.Height = layout.Px(rect.H)
	node.Children = make([]*layout.Node, 0, visibleCount)

	for i := c.visibleStart; i < c.visibleEnd && i < len(c.children); i++ {
		child := c.children[i]
		if child == nil || !child.Visible() {
			node.Children = append(node.Children, nil)
			continue
		}
		style := c.childStyles[child]
		childRef := child
		node.Children = append(node.Children, &layout.Node{
			Style: style,
			Measure: func(available coremath.Vec2) coremath.Vec2 {
				return childRef.Measure(ctx, Constraints{Max: available})
			},
		})
	}

	result := layout.Compute(node, coremath.NewVec2(rect.W, rect.H))

	// Apply layout results to visible children
	for i, childResult := range result.Children {
		childIdx := c.visibleStart + i
		if childIdx >= len(c.children) || c.children[childIdx] == nil {
			continue
		}
		// Offset by scroll position
		bounds := childResult.Bounds
		bounds.Y += float32(c.visibleStart) * c.itemHeight
		c.children[childIdx].Layout(ctx, bounds)
	}
}

// Render renders children with optional clipping and scroll offset.
func (c *LayoutContainer) Render(ctx *Context) {
	if c == nil || ctx == nil || ctx.Renderer == nil || !c.Visible() {
		return
	}
	bounds := c.Bounds()
	ctx.Renderer.PushTransform(coremath.TranslationMatrix(bounds.X, bounds.Y, 0))
	clip := c.clip || c.Style.OverflowX != layout.OverflowVisible || c.Style.OverflowY != layout.OverflowVisible
	if clip {
		ctx.Renderer.PushClip(coremath.NewRect(0, 0, bounds.W, bounds.H))
	}
	if c.scroll.X != 0 || c.scroll.Y != 0 {
		ctx.Renderer.PushTransform(coremath.TranslationMatrix(-c.scroll.X, -c.scroll.Y, 0))
	}

	// Render children - skip invisible ones when virtual scrolling is enabled
	for i, child := range c.children {
		if child == nil || !child.Visible() {
			continue
		}
		// Virtual scrolling: skip children outside visible range
		if c.virtualScroll && (i < c.visibleStart || i >= c.visibleEnd) {
			continue
		}
		child.Render(ctx)
	}

	if c.scroll.X != 0 || c.scroll.Y != 0 {
		ctx.Renderer.PopTransform()
	}
	if clip {
		ctx.Renderer.PopClip()
	}
	ctx.Renderer.PopTransform()
}

// HitTest returns the deepest child hit by a point in parent-local coordinates.
func (c *LayoutContainer) HitTest(point coremath.Vec2) Widget {
	if c == nil || !c.Visible() || !c.Enabled() {
		return nil
	}
	bounds := c.Bounds()
	if !bounds.Contains(point.X, point.Y) {
		return nil
	}
	local := coremath.NewVec2(point.X-bounds.X+c.scroll.X, point.Y-bounds.Y+c.scroll.Y)
	for i := len(c.children) - 1; i >= 0; i-- {
		child := c.children[i]
		if child == nil {
			continue
		}
		// Virtual scrolling: skip children outside visible range
		if c.virtualScroll && (i < c.visibleStart || i >= c.visibleEnd) {
			continue
		}
		if hit := child.HitTest(local); hit != nil {
			return hit
		}
	}
	return c
}

// HandleEvent routes input using the same coordinates as rendering.
func (c *LayoutContainer) HandleEvent(ctx *Context, event Event) bool {
	if c == nil || !c.Visible() || !c.Enabled() {
		return false
	}
	local := coremath.NewVec2(event.X-c.bounds.X+c.scroll.X, event.Y-c.bounds.Y+c.scroll.Y)
	switch event.Type {
	case EventMouseDown:
		return c.dispatchLayoutPointer(ctx, event, local, true)
	case EventMouseUp:
		return c.dispatchLayoutPointer(ctx, event, local, false)
	case EventMouseMove:
		return c.dispatchLayoutPointer(ctx, event, local, false)
	case EventMouseWheel:
		// Handle scroll for the container itself
		if c.Style.OverflowX == layout.OverflowScroll || c.Style.OverflowY == layout.OverflowScroll {
			oldX, oldY := c.scroll.X, c.scroll.Y
			scrollSpeed := float32(30)
			if c.Style.OverflowY == layout.OverflowScroll {
				c.scroll.Y -= event.DeltaY * scrollSpeed
			}
			if c.Style.OverflowX == layout.OverflowScroll {
				c.scroll.X -= event.DeltaX * scrollSpeed
			}
			c.clampScroll()
			c.updateVisibleRange()
			c.Invalidate()
			if c.onScroll != nil && (c.scroll.X != oldX || c.scroll.Y != oldY) {
				c.onScroll(c.scroll.X, c.scroll.Y)
			}
			return true
		}
		return c.dispatchWheel(ctx, event, local)
	case EventDoubleClick:
		return c.dispatchLayoutPointer(ctx, event, local, true)
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

func (c *LayoutContainer) layoutNode(ctx *Context) *layout.Node {
	node := layout.NewNode(c.Style)
	node.Children = make([]*layout.Node, 0, len(c.children))
	for _, child := range c.children {
		if child == nil || !child.Visible() {
			node.Children = append(node.Children, nil)
			continue
		}
		style := c.childStyles[child]
		childRef := child
		node.Children = append(node.Children, &layout.Node{
			Style: style,
			Measure: func(available coremath.Vec2) coremath.Vec2 {
				return childRef.Measure(ctx, Constraints{Max: available})
			},
		})
	}
	return node
}

func (c *LayoutContainer) dispatchLayoutPointer(ctx *Context, event Event, point coremath.Vec2, focusOnHit bool) bool {
	if c.pointerCapture != nil && (event.Type == EventMouseMove || event.Type == EventMouseUp) {
		return c.dispatchCapturedPointer(ctx, event, point)
	}
	for i := len(c.children) - 1; i >= 0; i-- {
		child := c.children[i]
		if child == nil || !child.Visible() || !child.Enabled() {
			continue
		}
		// Virtual scrolling: skip children outside visible range
		if c.virtualScroll && (i < c.visibleStart || i >= c.visibleEnd) {
			continue
		}
		hit := child.HitTest(point)
		if hit == nil {
			continue
		}
		if event.Type == EventMouseMove {
			c.updateHover(ctx, child, event, point)
		}
		child.SetStateFlag(StateHover, hit != nil && event.Type == EventMouseMove)
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

func (c *LayoutContainer) clampScroll() {
	if c == nil {
		return
	}
	bounds := c.Bounds()
	maxX := c.contentSize.X - bounds.W
	maxY := c.contentSize.Y - bounds.H
	if maxX < 0 {
		maxX = 0
	}
	if maxY < 0 {
		maxY = 0
	}
	if c.Style.OverflowX != layout.OverflowScroll {
		maxX = 0
	}
	if c.Style.OverflowY != layout.OverflowScroll {
		maxY = 0
	}
	if c.scroll.X < 0 {
		c.scroll.X = 0
	}
	if c.scroll.Y < 0 {
		c.scroll.Y = 0
	}
	if c.scroll.X > maxX {
		c.scroll.X = maxX
	}
	if c.scroll.Y > maxY {
		c.scroll.Y = maxY
	}
}

func (c *LayoutContainer) addFocusable(widget Widget) {
	c.Container.addFocusable(widget)
}

func (c *LayoutContainer) removeFocusable(widget Widget) {
	c.Container.removeFocusable(widget)
}
