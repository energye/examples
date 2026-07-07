package layout

import "github.com/energye/examples/lcl/gpui/core/math"

// Unit identifies how a layout value is resolved.
type Unit int

const (
	Auto Unit = iota
	Pixel
	Percent
)

// Value stores a scalar layout value.
type Value struct {
	Unit  Unit
	Value float32
}

// Px creates a pixel value.
func Px(value float32) Value {
	return Value{Unit: Pixel, Value: value}
}

// Pct creates a percent value where 1.0 means 100%.
func Pct(value float32) Value {
	return Value{Unit: Percent, Value: value}
}

// AutoValue creates an automatic value.
func AutoValue() Value {
	return Value{Unit: Auto}
}

// Edges stores box edge sizes.
type Edges struct {
	Top, Right, Bottom, Left float32
}

// EdgeAll creates equal edges.
func EdgeAll(value float32) Edges {
	return Edges{Top: value, Right: value, Bottom: value, Left: value}
}

// EdgeHV creates horizontal/vertical edges.
func EdgeHV(horizontal, vertical float32) Edges {
	return Edges{Top: vertical, Right: horizontal, Bottom: vertical, Left: horizontal}
}

// Direction controls flex layout direction.
type Direction int

const (
	Row Direction = iota
	Column
)

// Align controls cross-axis placement.
type Align int

const (
	AlignStart Align = iota
	AlignCenter
	AlignEnd
	AlignStretch
)

// Justify controls main-axis placement.
type Justify int

const (
	JustifyStart Justify = iota
	JustifyCenter
	JustifyEnd
	JustifySpaceBetween
)

// Style describes layout behavior.
type Style struct {
	Width     Value
	Height    Value
	MinWidth  Value
	MinHeight Value
	MaxWidth  Value
	MaxHeight Value

	Margin  Edges
	Padding Edges

	Direction Direction
	Wrap      bool
	Gap       float32
	Align     Align
	Justify   Justify
	FlexGrow  float32
}

// MeasureFunc returns a leaf node's desired size.
type MeasureFunc func(available math.Vec2) math.Vec2

// Node is a layout tree node.
type Node struct {
	Style    Style
	Measure  MeasureFunc
	Children []*Node
}

// NewNode creates a layout node.
func NewNode(style Style, children ...*Node) *Node {
	return &Node{Style: style, Children: children}
}

// NewSpace creates a gap-only flex container.
func NewSpace(direction Direction, gap float32, children ...*Node) *Node {
	return NewNode(Style{Direction: direction, Gap: gap}, children...)
}

// Result stores computed layout.
type Result struct {
	Bounds   math.Rect
	Children []Result
}

// Compute lays out a node within the available size.
func Compute(node *Node, available math.Vec2) Result {
	if node == nil {
		return Result{}
	}
	size := resolveNodeSize(node, available)
	result := Result{Bounds: math.NewRect(0, 0, size.X, size.Y)}
	if len(node.Children) == 0 {
		return result
	}

	if node.Style.Wrap && node.Style.Direction == Row {
		result.Children = layoutRowWrap(node, size)
	} else {
		result.Children = layoutLinear(node, size)
	}
	return result
}

func resolveNodeSize(node *Node, available math.Vec2) math.Vec2 {
	width := resolveValue(node.Style.Width, available.X)
	height := resolveValue(node.Style.Height, available.Y)

	if width <= 0 || height <= 0 {
		measured := math.Vec2{}
		if node.Measure != nil {
			measured = node.Measure(available)
		}
		if width <= 0 {
			width = measured.X
		}
		if height <= 0 {
			height = measured.Y
		}
	}

	if width <= 0 {
		width = available.X
	}
	if height <= 0 {
		height = available.Y
	}

	width = clampValue(width, node.Style.MinWidth, node.Style.MaxWidth, available.X)
	height = clampValue(height, node.Style.MinHeight, node.Style.MaxHeight, available.Y)
	return math.NewVec2(width, height)
}

func resolveValue(value Value, parent float32) float32 {
	switch value.Unit {
	case Pixel:
		return value.Value
	case Percent:
		return parent * value.Value
	default:
		return 0
	}
}

func clampValue(value float32, minValue, maxValue Value, parent float32) float32 {
	min := resolveValue(minValue, parent)
	max := resolveValue(maxValue, parent)
	if min > 0 && value < min {
		value = min
	}
	if max > 0 && value > max {
		value = max
	}
	return value
}
