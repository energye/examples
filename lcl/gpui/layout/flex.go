package layout

import "github.com/energye/examples/lcl/gpui/core/math"

type childLayoutInput struct {
	node        *Node
	baseSize    math.Vec2
	margin      Edges
	flexGrow    float32
	flexShrink  float32
	finalSize   math.Vec2
	mainOffset  float32
	crossOffset float32
}

func layoutLinear(node *Node, size math.Vec2) []Result {
	content := contentSize(node.Style.Padding, size)
	children := measureChildren(node, content)
	mainSize := axisMain(node.Style.Direction, content)
	crossSize := axisCross(node.Style.Direction, content)

	effectiveGap := node.Style.Gap
	totalMain := totalChildrenMain(node.Style.Direction, children, effectiveGap)
	totalGrow := totalFlexGrow(children)
	remaining := mainSize - totalMain

	// Flex-grow: distribute extra space
	if remaining > 0 && totalGrow > 0 {
		for i := range children {
			if children[i].flexGrow <= 0 {
				continue
			}
			extra := remaining * children[i].flexGrow / totalGrow
			if node.Style.Direction == Row {
				children[i].finalSize.X += extra
				children[i].finalSize.X = clampToConstraints(children[i].finalSize.X, children[i].node.Style.MinWidth, children[i].node.Style.MaxWidth, mainSize)
			} else {
				children[i].finalSize.Y += extra
				children[i].finalSize.Y = clampToConstraints(children[i].finalSize.Y, children[i].node.Style.MinHeight, children[i].node.Style.MaxHeight, mainSize)
			}
		}
		totalMain = totalChildrenMain(node.Style.Direction, children, effectiveGap)
	}

	// Flex-shrink: reduce space when overflow
	totalShrink := totalFlexShrink(children)
	if remaining < 0 && totalShrink > 0 {
		deficit := -remaining
		for i := range children {
			if children[i].flexShrink <= 0 {
				continue
			}
			reduction := deficit * children[i].flexShrink / totalShrink
			if node.Style.Direction == Row {
				children[i].finalSize.X -= reduction
				children[i].finalSize.X = clampToConstraints(children[i].finalSize.X, children[i].node.Style.MinWidth, children[i].node.Style.MaxWidth, mainSize)
			} else {
				children[i].finalSize.Y -= reduction
				children[i].finalSize.Y = clampToConstraints(children[i].finalSize.Y, children[i].node.Style.MinHeight, children[i].node.Style.MaxHeight, mainSize)
			}
		}
		totalMain = totalChildrenMain(node.Style.Direction, children, effectiveGap)
	}
	if node.Style.Justify == JustifySpaceBetween && len(children) > 1 {
		childMainTotal := totalChildrenMain(node.Style.Direction, children, 0)
		free := mainSize - childMainTotal
		if free > 0 {
			effectiveGap = free / float32(len(children)-1)
			totalMain = mainSize
		}
	}

	cursor := mainStart(node.Style.Justify, mainSize, totalMain) + leadingPadding(node.Style.Direction, node.Style.Padding)
	for i := range children {
		mainExtent := childMain(node.Style.Direction, children[i])
		crossExtent := childCross(node.Style.Direction, children[i])
		children[i].mainOffset = cursor + leadingMargin(node.Style.Direction, children[i].margin)
		children[i].crossOffset = crossStart(node.Style.Align, crossSize, crossExtent) + leadingCrossPadding(node.Style.Direction, node.Style.Padding) + leadingCrossMargin(node.Style.Direction, children[i].margin)
		if node.Style.Align == AlignStretch {
			if node.Style.Direction == Row {
				children[i].finalSize.Y = crossSize - children[i].margin.Top - children[i].margin.Bottom
			} else {
				children[i].finalSize.X = crossSize - children[i].margin.Left - children[i].margin.Right
			}
		}
		cursor += mainExtent
		if i < len(children)-1 {
			cursor += effectiveGap
		}
	}

	return buildResults(node.Style.Direction, children)
}

func layoutRowWrap(node *Node, size math.Vec2) []Result {
	content := contentSize(node.Style.Padding, size)
	children := measureChildren(node, content)
	lines := make([][]int, 0, 4)
	current := make([]int, 0, len(children))
	lineMain := float32(0)

	for i, child := range children {
		childMainSize := childMain(Row, child)
		nextMain := childMainSize
		if len(current) > 0 {
			nextMain += lineMain + node.Style.Gap
		}
		if len(current) > 0 && nextMain > content.X {
			lines = append(lines, current)
			current = make([]int, 0, len(children)-i)
			lineMain = 0
		}
		current = append(current, i)
		if lineMain > 0 {
			lineMain += node.Style.Gap
		}
		lineMain += childMainSize
	}
	if len(current) > 0 {
		lines = append(lines, current)
	}

	y := node.Style.Padding.Top
	for _, line := range lines {
		lineCross := float32(0)
		lineMain = 0
		for pos, idx := range line {
			child := children[idx]
			if pos > 0 {
				lineMain += node.Style.Gap
			}
			lineMain += childMain(Row, child)
			if cross := childCross(Row, child); cross > lineCross {
				lineCross = cross
			}
		}

		x := node.Style.Padding.Left + mainStart(node.Style.Justify, content.X, lineMain)
		for pos, idx := range line {
			child := &children[idx]
			child.mainOffset = x + child.margin.Left
			child.crossOffset = y + crossStart(node.Style.Align, lineCross, childCross(Row, *child)) + child.margin.Top
			x += childMain(Row, *child)
			if pos < len(line)-1 {
				x += node.Style.Gap
			}
		}
		y += lineCross + node.Style.Gap
	}

	return buildResults(Row, children)
}

func measureChildren(node *Node, available math.Vec2) []childLayoutInput {
	children := make([]childLayoutInput, len(node.Children))
	for i, child := range node.Children {
		if child == nil {
			continue
		}
		childSize := resolveNodeSize(child, available)
		children[i] = childLayoutInput{
			node:       child,
			baseSize:   childSize,
			margin:     child.Style.Margin,
			flexGrow:   child.Style.FlexGrow,
			flexShrink: child.Style.FlexShrink,
			finalSize:  childSize,
		}
	}
	return children
}

func buildResults(direction Direction, children []childLayoutInput) []Result {
	results := make([]Result, len(children))
	for i, child := range children {
		x, y := child.mainOffset, child.crossOffset
		if direction == Column {
			x, y = child.crossOffset, child.mainOffset
		}
		results[i] = Compute(child.node, child.finalSize)
		results[i].Bounds = math.NewRect(x, y, child.finalSize.X, child.finalSize.Y)
	}
	return results
}

func contentSize(padding Edges, size math.Vec2) math.Vec2 {
	w := size.X - padding.Left - padding.Right
	h := size.Y - padding.Top - padding.Bottom
	if w < 0 {
		w = 0
	}
	if h < 0 {
		h = 0
	}
	return math.NewVec2(w, h)
}

func axisMain(direction Direction, size math.Vec2) float32 {
	if direction == Row {
		return size.X
	}
	return size.Y
}

func axisCross(direction Direction, size math.Vec2) float32 {
	if direction == Row {
		return size.Y
	}
	return size.X
}

func childMain(direction Direction, child childLayoutInput) float32 {
	if direction == Row {
		return child.finalSize.X + child.margin.Left + child.margin.Right
	}
	return child.finalSize.Y + child.margin.Top + child.margin.Bottom
}

func childCross(direction Direction, child childLayoutInput) float32 {
	if direction == Row {
		return child.finalSize.Y + child.margin.Top + child.margin.Bottom
	}
	return child.finalSize.X + child.margin.Left + child.margin.Right
}

func totalChildrenMain(direction Direction, children []childLayoutInput, gap float32) float32 {
	var total float32
	for i, child := range children {
		if i > 0 {
			total += gap
		}
		total += childMain(direction, child)
	}
	return total
}

func totalFlexGrow(children []childLayoutInput) float32 {
	var total float32
	for _, child := range children {
		total += child.flexGrow
	}
	return total
}

func totalFlexShrink(children []childLayoutInput) float32 {
	var total float32
	for _, child := range children {
		total += child.flexShrink
	}
	return total
}

func mainStart(justify Justify, available, used float32) float32 {
	free := available - used
	if free < 0 {
		free = 0
	}
	switch justify {
	case JustifyCenter:
		return free / 2
	case JustifyEnd:
		return free
	default:
		return 0
	}
}

func crossStart(align Align, available, used float32) float32 {
	free := available - used
	if free < 0 {
		free = 0
	}
	switch align {
	case AlignCenter:
		return free / 2
	case AlignEnd:
		return free
	default:
		return 0
	}
}

func leadingPadding(direction Direction, padding Edges) float32 {
	if direction == Row {
		return padding.Left
	}
	return padding.Top
}

func leadingCrossPadding(direction Direction, padding Edges) float32 {
	if direction == Row {
		return padding.Top
	}
	return padding.Left
}

func leadingMargin(direction Direction, margin Edges) float32 {
	if direction == Row {
		return margin.Left
	}
	return margin.Top
}

func leadingCrossMargin(direction Direction, margin Edges) float32 {
	if direction == Row {
		return margin.Top
	}
	return margin.Left
}
