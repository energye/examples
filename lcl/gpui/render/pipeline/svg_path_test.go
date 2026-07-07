package pipeline

import "testing"

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
