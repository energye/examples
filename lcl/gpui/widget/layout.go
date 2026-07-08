package widget

import (
	coremath "github.com/energye/examples/lcl/gpui/core/math"
	"github.com/energye/examples/lcl/gpui/layout"
)

// LayoutContainer applies the shared layout engine to real widget children.
type LayoutContainer struct {
	Container
	Style       layout.Style
	childStyles map[Widget]layout.Style
	scroll      coremath.Vec2
	contentSize coremath.Vec2
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
	c.scroll = coremath.NewVec2(x, y)
	c.clampScroll()
	c.Invalidate()
}

// ScrollTo adjusts scroll offset so a content-local rect becomes visible.
func (c *LayoutContainer) ScrollTo(rect coremath.Rect) {
	if c == nil {
		return
	}
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
	c.Invalidate()
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
	result := layout.Compute(c.layoutNode(ctx), available)
	return ClampSize(coremath.NewVec2(result.Bounds.W, result.Bounds.H), constraints)
}

// Layout computes and applies child bounds in container-local coordinates.
func (c *LayoutContainer) Layout(ctx *Context, rect coremath.Rect) {
	if c == nil {
		return
	}
	c.BaseWidget.Layout(ctx, rect)
	node := c.layoutNode(ctx)
	node.Style.Width = layout.Px(rect.W)
	node.Style.Height = layout.Px(rect.H)
	result := layout.Compute(node, coremath.NewVec2(rect.W, rect.H))
	c.contentSize = result.ContentSize
	c.clampScroll()

	for i, child := range c.children {
		if child == nil || i >= len(result.Children) {
			continue
		}
		child.Layout(ctx, result.Children[i].Bounds)
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
	for _, child := range c.children {
		if child != nil && child.Visible() {
			child.Render(ctx)
		}
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
