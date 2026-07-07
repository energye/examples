package layout

import (
	"testing"

	"github.com/energye/examples/lcl/gpui/core/math"
)

func TestRowLayoutGapPadding(t *testing.T) {
	root := &Node{
		Style: Style{
			Width:     Px(300),
			Height:    Px(100),
			Direction: Row,
			Padding:   EdgeAll(10),
			Gap:       5,
		},
		Children: []*Node{
			fixedNode(50, 20),
			fixedNode(60, 20),
		},
	}

	result := Compute(root, math.NewVec2(300, 100))
	assertRect(t, result.Children[0].Bounds, 10, 10, 50, 20)
	assertRect(t, result.Children[1].Bounds, 65, 10, 60, 20)
}

func TestColumnLayout(t *testing.T) {
	root := &Node{
		Style: Style{
			Width:     Px(100),
			Height:    Px(200),
			Direction: Column,
			Padding:   EdgeHV(8, 12),
			Gap:       4,
		},
		Children: []*Node{
			fixedNode(40, 20),
			fixedNode(50, 30),
		},
	}

	result := Compute(root, math.NewVec2(100, 200))
	assertRect(t, result.Children[0].Bounds, 8, 12, 40, 20)
	assertRect(t, result.Children[1].Bounds, 8, 36, 50, 30)
}

func TestFlexGrow(t *testing.T) {
	root := &Node{
		Style: Style{
			Width:     Px(300),
			Height:    Px(40),
			Direction: Row,
		},
		Children: []*Node{
			fixedNode(50, 20),
			&Node{Style: Style{Width: Px(50), Height: Px(20), FlexGrow: 1}},
		},
	}

	result := Compute(root, math.NewVec2(300, 40))
	assertRect(t, result.Children[0].Bounds, 0, 0, 50, 20)
	assertRect(t, result.Children[1].Bounds, 50, 0, 250, 20)
}

func TestJustifySpaceBetween(t *testing.T) {
	root := &Node{
		Style: Style{
			Width:     Px(200),
			Height:    Px(40),
			Direction: Row,
			Justify:   JustifySpaceBetween,
		},
		Children: []*Node{
			fixedNode(50, 20),
			fixedNode(50, 20),
			fixedNode(50, 20),
		},
	}

	result := Compute(root, math.NewVec2(200, 40))
	assertRect(t, result.Children[0].Bounds, 0, 0, 50, 20)
	assertRect(t, result.Children[1].Bounds, 75, 0, 50, 20)
	assertRect(t, result.Children[2].Bounds, 150, 0, 50, 20)
}

func TestRowWrap(t *testing.T) {
	root := &Node{
		Style: Style{
			Width:     Px(120),
			Height:    Px(100),
			Direction: Row,
			Wrap:      true,
			Gap:       10,
		},
		Children: []*Node{
			fixedNode(70, 20),
			fixedNode(70, 20),
			fixedNode(30, 20),
		},
	}

	result := Compute(root, math.NewVec2(120, 100))
	assertRect(t, result.Children[0].Bounds, 0, 0, 70, 20)
	assertRect(t, result.Children[1].Bounds, 0, 30, 70, 20)
	assertRect(t, result.Children[2].Bounds, 80, 30, 30, 20)
}

func TestMinMaxPercent(t *testing.T) {
	node := &Node{
		Style: Style{
			Width:    Pct(0.8),
			Height:   Px(20),
			MaxWidth: Px(120),
			MinWidth: Px(80),
		},
	}

	result := Compute(node, math.NewVec2(200, 100))
	assertRect(t, result.Bounds, 0, 0, 120, 20)
}

func TestGridLayout(t *testing.T) {
	root := &Node{
		Style: Style{
			Width:       Px(300),
			Height:      Px(200),
			Padding:     EdgeAll(10),
			GridColumns: []Value{Px(80), AutoValue()},
			GridRows:    []Value{Px(30), Px(40)},
			ColumnGap:   5,
			RowGap:      7,
		},
		Children: []*Node{
			fixedNode(10, 10),
			fixedNode(10, 10),
			fixedNode(10, 10),
		},
	}

	result := Compute(root, math.NewVec2(300, 200))
	assertRect(t, result.Children[0].Bounds, 10, 10, 80, 30)
	assertRect(t, result.Children[1].Bounds, 95, 10, 195, 30)
	assertRect(t, result.Children[2].Bounds, 10, 47, 80, 40)
}

func TestOverflowViewportAndContentSize(t *testing.T) {
	root := &Node{
		Style: Style{
			Width:     Px(100),
			Height:    Px(50),
			Direction: Row,
			OverflowX: OverflowScroll,
			OverflowY: OverflowHidden,
		},
		Children: []*Node{
			fixedNode(160, 20),
		},
	}

	result := Compute(root, math.NewVec2(100, 50))
	assertRect(t, result.Viewport, 0, 0, 100, 50)
	if result.ContentSize.X != 160 || result.ContentSize.Y != 50 {
		t.Fatalf("content size = (%v,%v), want (160,50)", result.ContentSize.X, result.ContentSize.Y)
	}
}

func fixedNode(w, h float32) *Node {
	return &Node{Style: Style{Width: Px(w), Height: Px(h)}}
}

func assertRect(t *testing.T, got math.Rect, x, y, w, h float32) {
	t.Helper()
	if got.X != x || got.Y != y || got.W != w || got.H != h {
		t.Fatalf("rect = (%v,%v,%v,%v), want (%v,%v,%v,%v)", got.X, got.Y, got.W, got.H, x, y, w, h)
	}
}
