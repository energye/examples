// Package widget provides scroll container
package widget

import (
	"github.com/energye/examples/lcl/gpui/core/math"
	"github.com/energye/examples/lcl/gpui/render/pipeline"
	"github.com/energye/examples/lcl/gpui/style/color"
)

// ScrollContainer is a container with scrolling support
type ScrollContainer struct {
	BaseWidget
	children    []Widget
	scrollX     float32
	scrollY     float32
	contentW    float32
	contentH    float32
	scrollBarW  float32
	showScrollX bool
	showScrollY bool
	dragging    bool
	dragStartX  float32
	dragStartY  float32
	dragScrollX float32
	dragScrollY float32
}

// NewScrollContainer creates a new scroll container
func NewScrollContainer() *ScrollContainer {
	return &ScrollContainer{
		BaseWidget:  NewBaseWidget(),
		children:    make([]Widget, 0),
		scrollBarW:  8,
		showScrollX: true,
		showScrollY: true,
	}
}

// Add adds a child widget
func (sc *ScrollContainer) Add(child Widget) {
	child.SetParent(sc)
	sc.children = append(sc.children, child)
	sc.updateContentSize()
}

// Remove removes a child widget
func (sc *ScrollContainer) Remove(child Widget) {
	for i, ch := range sc.children {
		if ch == child {
			child.SetParent(nil)
			sc.children = append(sc.children[:i], sc.children[i+1:]...)
			sc.updateContentSize()
			return
		}
	}
}

// Children returns the child widgets
func (sc *ScrollContainer) Children() []Widget {
	return sc.children
}

// ScrollX returns the horizontal scroll position
func (sc *ScrollContainer) ScrollX() float32 {
	return sc.scrollX
}

// ScrollY returns the vertical scroll position
func (sc *ScrollContainer) ScrollY() float32 {
	return sc.scrollY
}

// SetScroll sets the scroll position
func (sc *ScrollContainer) SetScroll(x, y float32) {
	sc.scrollX = x
	sc.scrollY = y
	sc.clampScroll()
}

// ScrollTo scrolls to make a rectangle visible
func (sc *ScrollContainer) ScrollTo(rect math.Rect) {
	// Horizontal
	if rect.X < sc.scrollX {
		sc.scrollX = rect.X
	} else if rect.X+rect.W > sc.scrollX+sc.bounds.W {
		sc.scrollX = rect.X + rect.W - sc.bounds.W
	}

	// Vertical
	if rect.Y < sc.scrollY {
		sc.scrollY = rect.Y
	} else if rect.Y+rect.H > sc.scrollY+sc.bounds.H {
		sc.scrollY = rect.Y + rect.H - sc.bounds.H
	}

	sc.clampScroll()
}

// updateContentSize updates the content size based on children
func (sc *ScrollContainer) updateContentSize() {
	sc.contentW = 0
	sc.contentH = 0

	for _, child := range sc.children {
		bounds := child.Bounds()
		right := bounds.X + bounds.W
		bottom := bounds.Y + bounds.H

		if right > sc.contentW {
			sc.contentW = right
		}
		if bottom > sc.contentH {
			sc.contentH = bottom
		}
	}
}

// clampScroll clamps scroll position to valid range
func (sc *ScrollContainer) clampScroll() {
	maxScrollX := sc.contentW - sc.bounds.W
	maxScrollY := sc.contentH - sc.bounds.H

	if maxScrollX < 0 {
		maxScrollX = 0
	}
	if maxScrollY < 0 {
		maxScrollY = 0
	}

	if sc.scrollX < 0 {
		sc.scrollX = 0
	}
	if sc.scrollX > maxScrollX {
		sc.scrollX = maxScrollX
	}
	if sc.scrollY < 0 {
		sc.scrollY = 0
	}
	if sc.scrollY > maxScrollY {
		sc.scrollY = maxScrollY
	}
}

// needsScrollbarX returns whether horizontal scrollbar is needed
func (sc *ScrollContainer) needsScrollbarX() bool {
	return sc.showScrollX && sc.contentW > sc.bounds.W
}

// needsScrollbarY returns whether vertical scrollbar is needed
func (sc *ScrollContainer) needsScrollbarY() bool {
	return sc.showScrollY && sc.contentH > sc.bounds.H
}

// Render renders the scroll container
func (sc *ScrollContainer) Render(renderer *pipeline.Renderer) {
	if !sc.visible {
		return
	}

	// Draw background
	renderer.FillRoundRect(sc.bounds, 0, color.BgBase)

	// Set clipping (simplified - just skip children outside bounds)
	// TODO: Implement proper clipping

	// Render children with scroll offset
	for _, child := range sc.children {
		childBounds := child.Bounds()

		// Check if child is visible
		if childBounds.X+childBounds.W < sc.scrollX ||
			childBounds.X > sc.scrollX+sc.bounds.W ||
			childBounds.Y+childBounds.H < sc.scrollY ||
			childBounds.Y > sc.scrollY+sc.bounds.H {
			continue // Skip invisible children
		}

		// TODO: Apply scroll offset to rendering
		child.Render(renderer)
	}

	// Draw scrollbars
	sc.drawScrollbars(renderer)
}

// drawScrollbars draws the scrollbars
func (sc *ScrollContainer) drawScrollbars(renderer *pipeline.Renderer) {
	scrollBarColor := math.NewColor(0, 0, 0, 0.3)
	scrollBarHoverColor := math.NewColor(0, 0, 0, 0.5)

	// Vertical scrollbar
	if sc.needsScrollbarY() {
		trackHeight := sc.bounds.H
		thumbHeight := (sc.bounds.H / sc.contentH) * trackHeight
		if thumbHeight < 20 {
			thumbHeight = 20
		}

		thumbY := sc.bounds.Y + (sc.scrollY/sc.contentH)*(trackHeight-thumbHeight)
		thumbRect := math.NewRect(
			sc.bounds.X+sc.bounds.W-sc.scrollBarW,
			thumbY,
			sc.scrollBarW,
			thumbHeight,
		)

		// Draw track
		trackRect := math.NewRect(
			sc.bounds.X+sc.bounds.W-sc.scrollBarW,
			sc.bounds.Y,
			sc.scrollBarW,
			sc.bounds.H,
		)
		renderer.FillRoundRect(trackRect, sc.scrollBarW/2, math.NewColor(0, 0, 0, 0.1))

		// Draw thumb
		renderer.FillRoundRect(thumbRect, sc.scrollBarW/2, scrollBarColor)
	}

	// Horizontal scrollbar
	if sc.needsScrollbarX() {
		trackWidth := sc.bounds.W
		thumbWidth := (sc.bounds.W / sc.contentW) * trackWidth
		if thumbWidth < 20 {
			thumbWidth = 20
		}

		thumbX := sc.bounds.X + (sc.scrollX/sc.contentW)*(trackWidth-thumbWidth)
		thumbRect := math.NewRect(
			thumbX,
			sc.bounds.Y+sc.bounds.H-sc.scrollBarW,
			thumbWidth,
			sc.scrollBarW,
		)

		// Draw track
		trackRect := math.NewRect(
			sc.bounds.X,
			sc.bounds.Y+sc.bounds.H-sc.scrollBarW,
			sc.bounds.W,
			sc.scrollBarW,
		)
		renderer.FillRoundRect(trackRect, sc.scrollBarW/2, math.NewColor(0, 0, 0, 0.1))

		// Draw thumb
		renderer.FillRoundRect(thumbRect, sc.scrollBarW/2, scrollBarColor)
	}

	_ = scrollBarHoverColor
}

// MouseDown handles mouse down
func (sc *ScrollContainer) MouseDown(x, y float32, button int) bool {
	if !sc.enabled || !sc.bounds.Contains(x, y) {
		return false
	}

	// Check if clicking on scrollbar
	if sc.needsScrollbarY() {
		scrollBarX := sc.bounds.X + sc.bounds.W - sc.scrollBarW
		if x >= scrollBarX {
			// Start dragging scrollbar
			sc.dragging = true
			sc.dragStartX = x
			sc.dragStartY = y
			sc.dragScrollX = sc.scrollX
			sc.dragScrollY = sc.scrollY
			return true
		}
	}

	if sc.needsScrollbarX() {
		scrollBarY := sc.bounds.Y + sc.bounds.H - sc.scrollBarW
		if y >= scrollBarY {
			// Start dragging scrollbar
			sc.dragging = true
			sc.dragStartX = x
			sc.dragStartY = y
			sc.dragScrollX = sc.scrollX
			sc.dragScrollY = sc.scrollY
			return true
		}
	}

	// Pass to children
	for i := len(sc.children) - 1; i >= 0; i-- {
		child := sc.children[i]
		childBounds := child.Bounds()

		// Adjust coordinates for scroll offset
		childX := x + sc.scrollX
		childY := y + sc.scrollY

		if childBounds.Contains(childX, childY) {
			return child.MouseDown(childX, childY, button)
		}
	}

	return false
}

// MouseUp handles mouse up
func (sc *ScrollContainer) MouseUp(x, y float32, button int) bool {
	sc.dragging = false
	return false
}

// MouseMove handles mouse move
func (sc *ScrollContainer) MouseMove(x, y float32) bool {
	if sc.dragging {
		dx := x - sc.dragStartX
		dy := y - sc.dragStartY

		sc.scrollX = sc.dragScrollX + dx
		sc.scrollY = sc.dragScrollY + dy
		sc.clampScroll()
		return true
	}

	// Pass to children
	for _, child := range sc.children {
		childBounds := child.Bounds()
		childX := x + sc.scrollX
		childY := y + sc.scrollY

		if childBounds.Contains(childX, childY) {
			child.MouseMove(childX, childY)
		}
	}

	return false
}

// MouseWheel handles mouse wheel for scrolling
func (sc *ScrollContainer) MouseWheel(x, y, delta float32) bool {
	if !sc.enabled || !sc.bounds.Contains(x, y) {
		return false
	}

	sc.scrollY -= delta * 20 // Scroll speed
	sc.clampScroll()
	return true
}
