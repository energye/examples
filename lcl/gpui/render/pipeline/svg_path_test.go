package pipeline

import (
	"testing"

	coremath "github.com/energye/examples/lcl/gpui/core/math"
)

func TestParseSVGPathLinesAndRelativeCommands(t *testing.T) {
	path, err := ParseSVGPath("M10 10 h20 v10 l-20 0 z")
	if err != nil {
		t.Fatalf("ParseSVGPath returned error: %v", err)
	}

	subpaths := pathSubpaths(path)
	if len(subpaths) != 1 {
		t.Fatalf("expected 1 subpath, got %d", len(subpaths))
	}
	points := subpaths[0]
	if len(points) != 4 {
		t.Fatalf("expected 4 points, got %d", len(points))
	}

	want := [][2]float32{{10, 10}, {30, 10}, {30, 20}, {10, 20}}
	for i, point := range points {
		if point.X != want[i][0] || point.Y != want[i][1] {
			t.Fatalf("point %d = (%v,%v), want (%v,%v)", i, point.X, point.Y, want[i][0], want[i][1])
		}
	}
}

func TestParseSVGPathFlattensCubic(t *testing.T) {
	path, err := ParseSVGPath("M0 0 C10 0 10 10 20 10")
	if err != nil {
		t.Fatalf("ParseSVGPath returned error: %v", err)
	}

	points := pathPoints(path)
	if len(points) != curveSegments+1 {
		t.Fatalf("expected %d points, got %d", curveSegments+1, len(points))
	}
	last := points[len(points)-1]
	if last.X != 20 || last.Y != 10 {
		t.Fatalf("last point = (%v,%v), want (20,10)", last.X, last.Y)
	}
}

func TestBezierSegmentCountScalesWithCurveSize(t *testing.T) {
	short := cubicBezierSegmentCount(
		coremath.NewVec2(0, 0),
		coremath.NewVec2(10, 0),
		coremath.NewVec2(10, 10),
		coremath.NewVec2(20, 10),
	)
	long := cubicBezierSegmentCount(
		coremath.NewVec2(0, 0),
		coremath.NewVec2(200, -120),
		coremath.NewVec2(360, 260),
		coremath.NewVec2(520, 40),
	)
	if short < 12 {
		t.Fatalf("short curve segments = %d, want at least 12", short)
	}
	if long <= 16 {
		t.Fatalf("long curve segments = %d, want more than the old fixed 16", long)
	}
	if long <= short {
		t.Fatalf("long curve segments = %d, want greater than short curve %d", long, short)
	}
}

func TestParseSVGPathSmoothCubicAndQuadratic(t *testing.T) {
	path, err := ParseSVGPath("M0 0 C10 0 10 10 20 10 S30 20 40 10 Q50 0 60 10 T80 10")
	if err != nil {
		t.Fatalf("ParseSVGPath returned error: %v", err)
	}

	points := pathPoints(path)
	wantMin := 1 + curveSegments*4
	if len(points) < wantMin {
		t.Fatalf("expected at least %d points, got %d", wantMin, len(points))
	}
	last := points[len(points)-1]
	if last.X != 80 || last.Y != 10 {
		t.Fatalf("last point = (%v,%v), want (80,10)", last.X, last.Y)
	}
}

func TestParseSVGPathArc(t *testing.T) {
	path, err := ParseSVGPath("M10 10 A20 20 0 0 1 50 10")
	if err != nil {
		t.Fatalf("ParseSVGPath returned error: %v", err)
	}

	points := pathPoints(path)
	if len(points) <= 2 {
		t.Fatalf("expected arc to flatten into more than 2 points, got %d", len(points))
	}
	last := points[len(points)-1]
	if last.X < 49.99 || last.X > 50.01 || last.Y < 9.99 || last.Y > 10.01 {
		t.Fatalf("last point = (%v,%v), want approximately (50,10)", last.X, last.Y)
	}
}

func TestParseSVGPathRejectsIncompleteCommand(t *testing.T) {
	if _, err := ParseSVGPath("M 10"); err == nil {
		t.Fatal("expected incomplete move command to return an error")
	}
}

func TestNewSVGIconCompoundPath(t *testing.T) {
	icon, err := NewSVGIcon("M0 0 H100 V100 H0 Z M25 25 H75 V75 H25 Z", coremath.NewRect(0, 0, 100, 100), FillRuleEvenOdd)
	if err != nil {
		t.Fatalf("NewSVGIcon returned error: %v", err)
	}
	if icon.FillRule != FillRuleEvenOdd {
		t.Fatalf("fill rule = %v, want even-odd", icon.FillRule)
	}
	subpaths := pathSubpaths(icon.Path)
	if len(subpaths) != 2 {
		t.Fatalf("expected 2 subpaths, got %d", len(subpaths))
	}
}

func TestPathWindingDirection(t *testing.T) {
	ccw := NewPath()
	ccw.MoveTo(0, 0)
	ccw.LineTo(10, 0)
	ccw.LineTo(10, 10)
	ccw.LineTo(0, 10)
	ccw.Close()

	cw := NewPath()
	cw.MoveTo(0, 0)
	cw.LineTo(0, 10)
	cw.LineTo(10, 10)
	cw.LineTo(10, 0)
	cw.Close()

	if area := polygonArea(pathSubpaths(ccw)[0]); area <= 0 {
		t.Fatalf("ccw area = %v, want positive", area)
	}
	if area := polygonArea(pathSubpaths(cw)[0]); area >= 0 {
		t.Fatalf("cw area = %v, want negative", area)
	}
}

func TestTriangulateConcavePolygon(t *testing.T) {
	path := NewPath()
	path.MoveTo(0, 0)
	path.LineTo(4, 0)
	path.LineTo(4, 4)
	path.LineTo(2, 2)
	path.LineTo(0, 4)
	path.Close()

	points := pathSubpaths(path)[0]
	triangles := triangulateSimplePolygon(points)
	if len(triangles) != len(points)-2 {
		t.Fatalf("expected %d triangles, got %d", len(points)-2, len(triangles))
	}
}
