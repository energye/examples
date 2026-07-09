package pipeline

import (
	"testing"

	coremath "github.com/energye/examples/lcl/gpui/core/math"
)

func TestExpandTriangleForAAIncreasesArea(t *testing.T) {
	p0 := coremath.NewVec2(10, 10)
	p1 := coremath.NewVec2(60, 10)
	p2 := coremath.NewVec2(20, 50)

	expanded, ok := expandTriangleForAA(p0, p1, p2, 1)
	if !ok {
		t.Fatal("expected triangle expansion to succeed")
	}

	originalArea := polygonArea([]coremath.Vec2{p0, p1, p2})
	expandedArea := polygonArea(expanded[:])
	if expandedArea <= originalArea {
		t.Fatalf("expanded area = %v, want greater than original %v", expandedArea, originalArea)
	}
}

func TestOffsetPolygonForAARetainsVertexCount(t *testing.T) {
	points := []coremath.Vec2{
		coremath.NewVec2(10, 10),
		coremath.NewVec2(60, 10),
		coremath.NewVec2(60, 50),
		coremath.NewVec2(10, 50),
	}

	expanded := offsetPolygonForAA(points, 1)
	if len(expanded) != len(points) {
		t.Fatalf("expanded vertex count = %d, want %d", len(expanded), len(points))
	}
	if expanded[0].X >= points[0].X || expanded[0].Y >= points[0].Y {
		t.Fatalf("top-left vertex expanded to %#v, want outside %#v", expanded[0], points[0])
	}
}
