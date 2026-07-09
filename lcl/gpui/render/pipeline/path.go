package pipeline

import (
	"github.com/energye/examples/lcl/gpui/core/gl"
	"github.com/energye/examples/lcl/gpui/core/math"
)

// PathCommandKind identifies a vector path command.
type PathCommandKind int

const (
	PathMoveTo PathCommandKind = iota
	PathLineTo
	PathQuadTo  // Quadratic Bezier: control point + end point
	PathCubicTo // Cubic Bezier: two control points + end point
	PathClose
)

// PathCommand stores one vector path command.
type PathCommand struct {
	Kind   PathCommandKind
	Pos    math.Vec2
	Ctrl1  math.Vec2 // Control point 1 (for QuadTo/CubicTo)
	Ctrl2  math.Vec2 // Control point 2 (for CubicTo)
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

// QuadTo adds a quadratic Bezier curve segment.
func (p *Path) QuadTo(cx, cy, x, y float32) {
	p.commands = append(p.commands, PathCommand{
		Kind: PathQuadTo,
		Pos:  math.NewVec2(x, y),
		Ctrl1: math.NewVec2(cx, cy),
	})
}

// CubicTo adds a cubic Bezier curve segment.
func (p *Path) CubicTo(cx1, cy1, cx2, cy2, x, y float32) {
	p.commands = append(p.commands, PathCommand{
		Kind:  PathCubicTo,
		Pos:   math.NewVec2(x, y),
		Ctrl1: math.NewVec2(cx1, cy1),
		Ctrl2: math.NewVec2(cx2, cy2),
	})
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
		case PathQuadTo:
			if hasCurrent {
				// Flatten quadratic Bezier to line segments
				segments := 16
				for i := 1; i <= segments; i++ {
					t := float32(i) / float32(segments)
					mt := 1 - t
					x := mt*mt*current.X + 2*mt*t*cmd.Ctrl1.X + t*t*cmd.Pos.X
					y := mt*mt*current.Y + 2*mt*t*cmd.Ctrl1.Y + t*t*cmd.Pos.Y
					r.DrawLine(current.X, current.Y, x, y, width, color)
					current = math.NewVec2(x, y)
				}
			}
			current = cmd.Pos
			hasCurrent = true
		case PathCubicTo:
			if hasCurrent {
				// Flatten cubic Bezier to line segments
				segments := 16
				for i := 1; i <= segments; i++ {
					t := float32(i) / float32(segments)
					mt := 1 - t
					x := mt*mt*mt*current.X + 3*mt*mt*t*cmd.Ctrl1.X + 3*mt*t*t*cmd.Ctrl2.X + t*t*t*cmd.Pos.X
					y := mt*mt*mt*current.Y + 3*mt*mt*t*cmd.Ctrl1.Y + 3*mt*t*t*cmd.Ctrl2.Y + t*t*t*cmd.Pos.Y
					r.DrawLine(current.X, current.Y, x, y, width, color)
					current = math.NewVec2(x, y)
				}
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
	r.fillTriangleFan(points, color)
}

// FillPath fills simple subpaths. Non-convex paths are triangulated with ear clipping.
func (r *Renderer) FillPath(path *Path, color math.Color) {
	for _, points := range pathSubpaths(path) {
		if len(points) < 3 {
			continue
		}
		triangles := triangulateSimplePolygon(points)
		if len(triangles) == 0 {
			r.fillTriangleFan(points, color)
			continue
		}

		shaderProg := r.shaderMgr.GetShader("color")
		for _, tri := range triangles {
			verts := [3]Vertex{
				colorVertex(points[tri[0]], color),
				colorVertex(points[tri[1]], color),
				colorVertex(points[tri[2]], color),
			}
			r.addTriangle(shaderProg, 0, nil, verts)
		}
	}
}

// FillPathEvenOdd fills compound paths using the even-odd rule via the stencil buffer.
func (r *Renderer) FillPathEvenOdd(path *Path, color math.Color) {
	if r == nil || !stencilGLReady() {
		return
	}
	bounds, ok := pathBounds(path)
	if !ok {
		return
	}

	r.Flush()
	gl.Enable(gl.GL_STENCIL_TEST)
	defer func() {
		gl.ColorMask(true, true, true, true)
		gl.Disable(gl.GL_STENCIL_TEST)
	}()
	gl.ClearStencil(0)
	gl.Clear(gl.GL_STENCIL_BUFFER_BIT)
	gl.ColorMask(false, false, false, false)
	gl.StencilFunc(gl.GL_ALWAYS, 1, 0xFF)
	gl.StencilOp(gl.GL_KEEP, gl.GL_KEEP, gl.GL_INVERT)

	for _, points := range pathSubpaths(path) {
		r.fillSubpathTriangles(points, color)
	}
	r.Flush()

	gl.ColorMask(true, true, true, true)
	gl.StencilFunc(gl.GL_EQUAL, 1, 0xFF)
	gl.StencilOp(gl.GL_KEEP, gl.GL_KEEP, gl.GL_KEEP)
	r.FillRect(bounds, color)
	r.Flush()
}

// FillPathNonZero fills compound paths using the non-zero winding rule via the stencil buffer.
func (r *Renderer) FillPathNonZero(path *Path, color math.Color) {
	if r == nil || !stencilGLReady() {
		return
	}
	bounds, ok := pathBounds(path)
	if !ok {
		return
	}

	r.Flush()
	gl.Enable(gl.GL_STENCIL_TEST)
	defer func() {
		gl.ColorMask(true, true, true, true)
		gl.Disable(gl.GL_STENCIL_TEST)
	}()
	gl.ClearStencil(0)
	gl.Clear(gl.GL_STENCIL_BUFFER_BIT)
	gl.ColorMask(false, false, false, false)
	gl.StencilFunc(gl.GL_ALWAYS, 1, 0xFF)

	for _, points := range pathSubpaths(path) {
		if polygonArea(points) >= 0 {
			gl.StencilOp(gl.GL_KEEP, gl.GL_KEEP, gl.GL_INCR_WRAP)
		} else {
			gl.StencilOp(gl.GL_KEEP, gl.GL_KEEP, gl.GL_DECR_WRAP)
		}
		r.fillSubpathTriangles(points, color)
		r.Flush()
	}

	gl.ColorMask(true, true, true, true)
	gl.StencilFunc(gl.GL_NOTEQUAL, 0, 0xFF)
	gl.StencilOp(gl.GL_KEEP, gl.GL_KEEP, gl.GL_KEEP)
	r.FillRect(bounds, color)
	r.Flush()
}

func (r *Renderer) fillTriangleFan(points []math.Vec2, color math.Color) {
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

func (r *Renderer) fillSubpathTriangles(points []math.Vec2, color math.Color) {
	if len(points) < 3 {
		return
	}
	triangles := triangulateSimplePolygon(points)
	if len(triangles) == 0 {
		r.fillTriangleFan(points, color)
		return
	}

	shaderProg := r.shaderMgr.GetShader("color")
	for _, tri := range triangles {
		verts := [3]Vertex{
			colorVertex(points[tri[0]], color),
			colorVertex(points[tri[1]], color),
			colorVertex(points[tri[2]], color),
		}
		r.addTriangle(shaderProg, 0, nil, verts)
	}
}

func pathSubpaths(path *Path) [][]math.Vec2 {
	if path == nil {
		return nil
	}

	var subpaths [][]math.Vec2
	var current []math.Vec2
	var currentPos math.Vec2

	for _, cmd := range path.commands {
		switch cmd.Kind {
		case PathMoveTo:
			if len(current) > 0 {
				subpaths = append(subpaths, current)
			}
			current = []math.Vec2{cmd.Pos}
			currentPos = cmd.Pos
		case PathLineTo:
			current = append(current, cmd.Pos)
			currentPos = cmd.Pos
		case PathQuadTo:
			flattened := flattenQuadBezier(currentPos, cmd.Ctrl1, cmd.Pos, 16)
			current = append(current, flattened...)
			currentPos = cmd.Pos
		case PathCubicTo:
			flattened := flattenCubicBezier(currentPos, cmd.Ctrl1, cmd.Ctrl2, cmd.Pos, 16)
			current = append(current, flattened...)
			currentPos = cmd.Pos
		case PathClose:
			if len(current) > 0 {
				subpaths = append(subpaths, current)
				current = nil
			}
		}
	}
	if len(current) > 0 {
		subpaths = append(subpaths, current)
	}
	return subpaths
}

func pathBounds(path *Path) (math.Rect, bool) {
	points := pathPoints(path)
	if len(points) == 0 {
		return math.Rect{}, false
	}

	minX, maxX := points[0].X, points[0].X
	minY, maxY := points[0].Y, points[0].Y
	for _, point := range points[1:] {
		if point.X < minX {
			minX = point.X
		}
		if point.X > maxX {
			maxX = point.X
		}
		if point.Y < minY {
			minY = point.Y
		}
		if point.Y > maxY {
			maxY = point.Y
		}
	}
	return math.NewRect(minX, minY, maxX-minX, maxY-minY), true
}

func triangulateSimplePolygon(points []math.Vec2) [][3]int {
	if len(points) < 3 {
		return nil
	}

	indices := make([]int, len(points))
	if polygonArea(points) >= 0 {
		for i := range indices {
			indices[i] = i
		}
	} else {
		for i := range indices {
			indices[i] = len(points) - 1 - i
		}
	}

	triangles := make([][3]int, 0, len(points)-2)
	guard := 0
	for len(indices) > 3 && guard < len(points)*len(points) {
		guard++
		earFound := false

		for i := range indices {
			prev := indices[(i+len(indices)-1)%len(indices)]
			curr := indices[i]
			next := indices[(i+1)%len(indices)]

			if !isConvex(points[prev], points[curr], points[next]) {
				continue
			}
			if containsAnyPoint(points, indices, prev, curr, next) {
				continue
			}

			triangles = append(triangles, [3]int{prev, curr, next})
			indices = append(indices[:i], indices[i+1:]...)
			earFound = true
			break
		}

		if !earFound {
			return nil
		}
	}

	if len(indices) == 3 {
		triangles = append(triangles, [3]int{indices[0], indices[1], indices[2]})
	}
	return triangles
}

func polygonArea(points []math.Vec2) float32 {
	var area float32
	for i, p := range points {
		q := points[(i+1)%len(points)]
		area += p.X*q.Y - q.X*p.Y
	}
	return area * 0.5
}

func isConvex(a, b, c math.Vec2) bool {
	ab := b.Sub(a)
	bc := c.Sub(b)
	return ab.X*bc.Y-ab.Y*bc.X > 0
}

func containsAnyPoint(points []math.Vec2, indices []int, a, b, c int) bool {
	for _, idx := range indices {
		if idx == a || idx == b || idx == c {
			continue
		}
		if pointInTriangle(points[idx], points[a], points[b], points[c]) {
			return true
		}
	}
	return false
}

func pointInTriangle(p, a, b, c math.Vec2) bool {
	d1 := signedArea(p, a, b)
	d2 := signedArea(p, b, c)
	d3 := signedArea(p, c, a)

	hasNeg := d1 < 0 || d2 < 0 || d3 < 0
	hasPos := d1 > 0 || d2 > 0 || d3 > 0
	return !(hasNeg && hasPos)
}

func signedArea(a, b, c math.Vec2) float32 {
	return (a.X-c.X)*(b.Y-c.Y) - (b.X-c.X)*(a.Y-c.Y)
}

func pathPoints(path *Path) []math.Vec2 {
	if path == nil {
		return nil
	}

	points := make([]math.Vec2, 0, len(path.commands))
	var current math.Vec2
	for _, cmd := range path.commands {
		switch cmd.Kind {
		case PathMoveTo:
			current = cmd.Pos
			points = append(points, cmd.Pos)
		case PathLineTo:
			current = cmd.Pos
			points = append(points, cmd.Pos)
		case PathQuadTo:
			// Flatten quadratic Bezier to line segments
			flattened := flattenQuadBezier(current, cmd.Ctrl1, cmd.Pos, 16)
			points = append(points, flattened...)
			current = cmd.Pos
		case PathCubicTo:
			// Flatten cubic Bezier to line segments
			flattened := flattenCubicBezier(current, cmd.Ctrl1, cmd.Ctrl2, cmd.Pos, 16)
			points = append(points, flattened...)
			current = cmd.Pos
		}
	}
	return points
}

// flattenQuadBezier flattens a quadratic Bezier curve into line segments
func flattenQuadBezier(p0, p1, p2 math.Vec2, segments int) []math.Vec2 {
	points := make([]math.Vec2, segments)
	for i := 1; i <= segments; i++ {
		t := float32(i) / float32(segments)
		mt := 1 - t
		// Quadratic Bezier formula: B(t) = (1-t)^2*P0 + 2*(1-t)*t*P1 + t^2*P2
		x := mt*mt*p0.X + 2*mt*t*p1.X + t*t*p2.X
		y := mt*mt*p0.Y + 2*mt*t*p1.Y + t*t*p2.Y
		points[i-1] = math.NewVec2(x, y)
	}
	return points
}

// flattenCubicBezier flattens a cubic Bezier curve into line segments
func flattenCubicBezier(p0, p1, p2, p3 math.Vec2, segments int) []math.Vec2 {
	points := make([]math.Vec2, segments)
	for i := 1; i <= segments; i++ {
		t := float32(i) / float32(segments)
		mt := 1 - t
		// Cubic Bezier formula: B(t) = (1-t)^3*P0 + 3*(1-t)^2*t*P1 + 3*(1-t)*t^2*P2 + t^3*P3
		x := mt*mt*mt*p0.X + 3*mt*mt*t*p1.X + 3*mt*t*t*p2.X + t*t*t*p3.X
		y := mt*mt*mt*p0.Y + 3*mt*mt*t*p1.Y + 3*mt*t*t*p2.Y + t*t*t*p3.Y
		points[i-1] = math.NewVec2(x, y)
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

func stencilGLReady() bool {
	return gl.Enable != nil &&
		gl.Disable != nil &&
		gl.ClearStencil != nil &&
		gl.Clear != nil &&
		gl.ColorMask != nil &&
		gl.StencilFunc != nil &&
		gl.StencilOp != nil
}
