package layout

import "github.com/energye/examples/lcl/gpui/core/math"

func layoutGrid(node *Node, size math.Vec2) ([]Result, math.Vec2) {
	content := contentSize(node.Style.Padding, size)
	colGap := node.Style.ColumnGap
	if colGap == 0 {
		colGap = node.Style.Gap
	}
	rowGap := node.Style.RowGap
	if rowGap == 0 {
		rowGap = node.Style.Gap
	}

	cols := resolveTracks(node.Style.GridColumns, content.X, colGap)
	if len(cols) == 0 {
		return nil, content
	}
	rowCount := (len(node.Children) + len(cols) - 1) / len(cols)
	rows := resolveGridRows(node, content, rowCount, rowGap)

	results := make([]Result, len(node.Children))
	for i, child := range node.Children {
		if child == nil {
			continue
		}

		// Get span values (default to 1)
		colSpan := child.Style.GridColumnSpan
		if colSpan <= 0 {
			colSpan = 1
		}
		rowSpan := child.Style.GridRowSpan
		if rowSpan <= 0 {
			rowSpan = 1
		}

		// Calculate position
		col := i % len(cols)
		row := i / len(cols)

		// Calculate width spanning multiple columns
		w := float32(0)
		for s := 0; s < colSpan && col+s < len(cols); s++ {
			w += cols[col+s]
			if s > 0 {
				w += colGap
			}
		}

		// Calculate height spanning multiple rows
		h := float32(0)
		for s := 0; s < rowSpan && row+s < len(rows); s++ {
			h += rows[row+s]
			if s > 0 {
				h += rowGap
			}
		}

		x := node.Style.Padding.Left + trackOffset(cols, colGap, col)
		y := node.Style.Padding.Top + trackOffset(rows, rowGap, row)
		childResult := Compute(child, math.NewVec2(w, h))
		childResult.Bounds = math.NewRect(x, y, w, h)
		results[i] = childResult
	}

	contentW := sumTracks(cols, colGap) + node.Style.Padding.Left + node.Style.Padding.Right
	contentH := sumTracks(rows, rowGap) + node.Style.Padding.Top + node.Style.Padding.Bottom
	return results, math.NewVec2(contentW, contentH)
}

func resolveGridRows(node *Node, content math.Vec2, rowCount int, rowGap float32) []float32 {
	if len(node.Style.GridRows) > 0 {
		rows := resolveTracks(node.Style.GridRows, content.Y, rowGap)
		for len(rows) < rowCount {
			rows = append(rows, rows[len(rows)-1])
		}
		return rows[:rowCount]
	}

	rows := make([]float32, rowCount)
	cols := len(node.Style.GridColumns)
	for i, child := range node.Children {
		if child == nil {
			continue
		}
		row := i / cols
		measured := resolveNodeSize(child, content)
		if measured.Y > rows[row] {
			rows[row] = measured.Y
		}
	}
	return rows
}

func resolveTracks(values []Value, available, gap float32) []float32 {
	if len(values) == 0 {
		return nil
	}

	tracks := make([]float32, len(values))
	used := gap * float32(len(values)-1)
	autoCount := 0
	for i, value := range values {
		size := resolveValue(value, available)
		if value.Unit == Auto {
			autoCount++
			continue
		}
		tracks[i] = size
		used += size
	}

	autoSize := float32(0)
	if autoCount > 0 {
		remaining := available - used
		if remaining < 0 {
			remaining = 0
		}
		autoSize = remaining / float32(autoCount)
	}
	for i, value := range values {
		if value.Unit == Auto {
			tracks[i] = autoSize
		}
	}
	return tracks
}

func trackOffset(tracks []float32, gap float32, index int) float32 {
	var offset float32
	for i := 0; i < index; i++ {
		offset += tracks[i] + gap
	}
	return offset
}

func sumTracks(tracks []float32, gap float32) float32 {
	var total float32
	for i, track := range tracks {
		if i > 0 {
			total += gap
		}
		total += track
	}
	return total
}
