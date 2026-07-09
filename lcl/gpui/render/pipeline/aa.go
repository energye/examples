package pipeline

import "github.com/energye/examples/lcl/gpui/core/math"

const primitiveAAWidth = float32(1.0)

func expandTriangleForAA(p0, p1, p2 math.Vec2, amount float32) ([3]math.Vec2, bool) {
	var out [3]math.Vec2
	points := []math.Vec2{p0, p1, p2}
	expanded := offsetPolygonForAA(points, amount)
	if len(expanded) != 3 {
		return out, false
	}
	copy(out[:], expanded)
	return out, true
}

func offsetPolygonForAA(points []math.Vec2, amount float32) []math.Vec2 {
	if len(points) < 3 || amount <= 0 {
		return nil
	}

	ccw := polygonArea(points) >= 0
	expanded := make([]math.Vec2, len(points))
	maxMiter := amount * 4

	for i, curr := range points {
		prev := points[(i+len(points)-1)%len(points)]
		next := points[(i+1)%len(points)]

		prevNormal, okPrev := outwardEdgeNormal(prev, curr, ccw)
		nextNormal, okNext := outwardEdgeNormal(curr, next, ccw)
		if !okPrev || !okNext {
			expanded[i] = curr
			continue
		}

		prevA := prev.Add(prevNormal.Scale(amount))
		prevB := curr.Add(prevNormal.Scale(amount))
		nextA := curr.Add(nextNormal.Scale(amount))
		nextB := next.Add(nextNormal.Scale(amount))

		if p, ok := intersectLines(prevA, prevB, nextA, nextB); ok && p.Sub(curr).Length() <= maxMiter {
			expanded[i] = p
			continue
		}

		fallback := prevNormal.Add(nextNormal).Normalize()
		if fallback.Length() == 0 {
			fallback = prevNormal
		}
		expanded[i] = curr.Add(fallback.Scale(amount))
	}

	return expanded
}

func outwardEdgeNormal(a, b math.Vec2, ccw bool) (math.Vec2, bool) {
	edge := b.Sub(a)
	length := edge.Length()
	if length < 0.001 {
		return math.Vec2{}, false
	}
	if ccw {
		return math.NewVec2(edge.Y/length, -edge.X/length), true
	}
	return math.NewVec2(-edge.Y/length, edge.X/length), true
}

func intersectLines(a0, a1, b0, b1 math.Vec2) (math.Vec2, bool) {
	da := a1.Sub(a0)
	db := b1.Sub(b0)
	denom := da.X*db.Y - da.Y*db.X
	if denom > -0.00001 && denom < 0.00001 {
		return math.Vec2{}, false
	}
	delta := b0.Sub(a0)
	t := (delta.X*db.Y - delta.Y*db.X) / denom
	return a0.Add(da.Scale(t)), true
}

func (r *Renderer) transformedShaderPoint(p math.Vec2) math.Vec2 {
	if r == nil || len(r.transformStack) == 0 {
		return p
	}
	x, y := transformPoint(r.transformStack[len(r.transformStack)-1], p.X, p.Y)
	return math.NewVec2(x, y)
}
