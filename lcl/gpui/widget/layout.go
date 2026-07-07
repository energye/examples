package widget

import "github.com/energye/examples/lcl/gpui/core/math"

// FlexDirection controls the main axis of a FlexContainer.
type FlexDirection int

const (
	FlexRow FlexDirection = iota
	FlexColumn
)

// FlexContainer lays out children along one axis.
type FlexContainer struct {
	Container
	direction FlexDirection
	gap       float32
	padding   float32
}

// NewFlexContainer creates a flex layout container.
func NewFlexContainer(direction FlexDirection) *FlexContainer {
	return &FlexContainer{
		Container: *NewContainer(),
		direction: direction,
	}
}

// Add adds a child and refreshes layout.
func (fc *FlexContainer) Add(child Widget) {
	fc.Container.Add(child)
	fc.layoutChildren()
}

// Remove removes a child and refreshes layout.
func (fc *FlexContainer) Remove(child Widget) {
	fc.Container.Remove(child)
	fc.layoutChildren()
}

// SetGap sets the spacing between children.
func (fc *FlexContainer) SetGap(gap float32) {
	fc.gap = gap
	fc.layoutChildren()
}

// SetPadding sets equal padding on all sides.
func (fc *FlexContainer) SetPadding(padding float32) {
	fc.padding = padding
	fc.layoutChildren()
}

// SetDirection changes the flex direction.
func (fc *FlexContainer) SetDirection(direction FlexDirection) {
	fc.direction = direction
	fc.layoutChildren()
}

// SetBounds sets bounds and updates child layout.
func (fc *FlexContainer) SetBounds(rect math.Rect) {
	fc.Container.SetBounds(rect)
	fc.layoutChildren()
}

// Layout assigns final bounds and updates child layout.
func (fc *FlexContainer) Layout(rect math.Rect) {
	fc.SetBounds(rect)
}

// Measure returns the desired size for all children.
func (fc *FlexContainer) Measure(available math.Vec2) math.Vec2 {
	var main, cross float32
	for i, child := range fc.children {
		size := child.Measure(available)
		if i > 0 {
			main += fc.gap
		}
		if fc.direction == FlexRow {
			main += size.X
			if size.Y > cross {
				cross = size.Y
			}
		} else {
			main += size.Y
			if size.X > cross {
				cross = size.X
			}
		}
	}

	if fc.direction == FlexRow {
		return math.NewVec2(main+fc.padding*2, cross+fc.padding*2)
	}
	return math.NewVec2(cross+fc.padding*2, main+fc.padding*2)
}

func (fc *FlexContainer) layoutChildren() {
	cursor := fc.padding
	available := math.NewVec2(fc.bounds.W-fc.padding*2, fc.bounds.H-fc.padding*2)
	if available.X < 0 {
		available.X = 0
	}
	if available.Y < 0 {
		available.Y = 0
	}

	for _, child := range fc.children {
		size := child.Measure(available)
		if fc.direction == FlexRow {
			child.Layout(math.NewRect(cursor, fc.padding, size.X, size.Y))
			cursor += size.X + fc.gap
		} else {
			child.Layout(math.NewRect(fc.padding, cursor, size.X, size.Y))
			cursor += size.Y + fc.gap
		}
	}
}
