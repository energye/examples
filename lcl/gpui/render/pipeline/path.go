package pipeline

import "github.com/energye/examples/lcl/gpui/core/math"

// PathCommandKind identifies a vector path command.
type PathCommandKind int

const (
	PathMoveTo PathCommandKind = iota
	PathLineTo
	PathClose
)

// PathCommand stores one vector path command.
type PathCommand struct {
	Kind PathCommandKind
	Pos  math.Vec2
}

// Path stores a simple vector path.
type Path struct {
	commands []PathCommand
}

// NewPath creates an empty path.
func NewPath() *Path {
	return &Path{commands: make([]PathCommand, 0, 16)}
}

// MoveTo starts a new subpath.
func (p *Path) MoveTo(x, y float32) {
	p.commands = append(p.commands, PathCommand{Kind: PathMoveTo, Pos: math.NewVec2(x, y)})
}

// LineTo adds a line segment.
func (p *Path) LineTo(x, y float32) {
	p.commands = append(p.commands, PathCommand{Kind: PathLineTo, Pos: math.NewVec2(x, y)})
}

// Close closes the current subpath.
func (p *Path) Close() {
	p.commands = append(p.commands, PathCommand{Kind: PathClose})
}

// Commands returns the raw path command list.
func (p *Path) Commands() []PathCommand {
	return p.commands
}

// StrokePath strokes line segments in a path.
func (r *Renderer) StrokePath(path *Path, width float32, color math.Color) {
	if path == nil || width <= 0 {
		return
	}

	var start math.Vec2
	var current math.Vec2
	hasCurrent := false

	for _, cmd := range path.commands {
		switch cmd.Kind {
		case PathMoveTo:
			start = cmd.Pos
			current = cmd.Pos
			hasCurrent = true
		case PathLineTo:
			if hasCurrent {
				r.DrawLine(current.X, current.Y, cmd.Pos.X, cmd.Pos.Y, width, color)
			}
			current = cmd.Pos
			hasCurrent = true
		case PathClose:
			if hasCurrent {
				r.DrawLine(current.X, current.Y, start.X, start.Y, width, color)
				current = start
			}
		}
	}
}

// FillConvexPath fills a convex path using a triangle fan.
func (r *Renderer) FillConvexPath(path *Path, color math.Color) {
	points := pathPoints(path)
	if len(points) < 3 {
		return
	}

	shaderProg := r.shaderMgr.GetShader("color")
	center := polygonCenter(points)
	for i := 0; i < len(points); i++ {
		next := (i + 1) % len(points)
		verts := [3]Vertex{
			colorVertex(center, color),
			colorVertex(points[i], color),
			colorVertex(points[next], color),
		}
		r.addTriangle(shaderProg, 0, nil, verts)
	}
}

func pathPoints(path *Path) []math.Vec2 {
	if path == nil {
		return nil
	}

	points := make([]math.Vec2, 0, len(path.commands))
	for _, cmd := range path.commands {
		switch cmd.Kind {
		case PathMoveTo, PathLineTo:
			points = append(points, cmd.Pos)
		}
	}
	return points
}

func polygonCenter(points []math.Vec2) math.Vec2 {
	var center math.Vec2
	for _, point := range points {
		center.X += point.X
		center.Y += point.Y
	}
	count := float32(len(points))
	return math.NewVec2(center.X/count, center.Y/count)
}

func colorVertex(pos math.Vec2, color math.Color) Vertex {
	return Vertex{
		X: pos.X, Y: pos.Y,
		U: 0, V: 0,
		R: color.R, G: color.G, B: color.B, A: color.A,
	}
}
