// Package pipeline provides the rendering pipeline
package pipeline

import (
	stdmath "math"

	"github.com/energye/examples/lcl/gpui/core/math"
	"github.com/energye/examples/lcl/gpui/render/font"
)

// DrawText draws text at the given position
func (r *Renderer) DrawText(text string, x, y float32, f *font.Font, color math.Color) {
	if f == nil || text == "" {
		return
	}

	shaderProg := r.shaderMgr.GetShader("texture")
	r.shaderMgr.UseShader(shaderProg)

	// Bind font texture
	fontTex := f.Texture()
	if fontTex > 0 {
		// Already bound in batch manager
	}

	cx := x
	for _, ch := range text {
		g, ok := f.GetGlyph(ch)
		if !ok {
			continue
		}

		if g.Width > 0 && g.Height > 0 {
			gx := cx
			gy := y

			src := math.NewRect(g.U0, g.V0, g.U1-g.U0, g.V1-g.V0)
			dst := math.NewRect(gx, gy, g.Width, g.Height)

			verts := QuadVertices(dst, src, color)
			r.addQuad(shaderProg, fontTex, nil, verts)
		}

		cx += g.Advance
	}
}

// StrokeRect draws a rectangle outline
func (r *Renderer) StrokeRect(rect math.Rect, width float32, color math.Color) {
	x, y, w, h := rect.X, rect.Y, rect.W, rect.H

	// Top
	r.FillRect(math.NewRect(x, y, w, width), color)
	// Bottom
	r.FillRect(math.NewRect(x, y+h-width, w, width), color)
	// Left
	r.FillRect(math.NewRect(x, y, width, h), color)
	// Right
	r.FillRect(math.NewRect(x+w-width, y, width, h), color)
}

// StrokeRoundRect draws a rounded rectangle outline using SDF
func (r *Renderer) StrokeRoundRect(rect math.Rect, radius, width float32, color math.Color) {
	// For now, use a simple rectangle outline approximation
	// TODO: Implement proper SDF-based outline

	// Draw outer rounded rect
	r.FillRoundRect(rect, radius, color)

	// Draw inner rounded rect to create outline effect
	innerRect := math.NewRect(
		rect.X+width,
		rect.Y+width,
		rect.W-2*width,
		rect.H-2*width,
	)
	innerRadius := radius - width
	if innerRadius < 0 {
		innerRadius = 0
	}

	// Use background color for inner rect (assumes white background)
	bgColor := math.NewColor(1, 1, 1, 1)
	r.FillRoundRect(innerRect, innerRadius, bgColor)
}

// FillRectWithBorder draws a filled rectangle with border
func (r *Renderer) FillRectWithBorder(rect math.Rect, borderWidth float32, bgColor, borderColor math.Color) {
	// Draw border (larger rect)
	r.FillRect(rect, borderColor)

	// Draw background (smaller rect)
	innerRect := math.NewRect(
		rect.X+borderWidth,
		rect.Y+borderWidth,
		rect.W-2*borderWidth,
		rect.H-2*borderWidth,
	)
	r.FillRect(innerRect, bgColor)
}

// FillRoundRectWithBorder draws a filled rounded rectangle with border
func (r *Renderer) FillRoundRectWithBorder(rect math.Rect, radius, borderWidth float32, bgColor, borderColor math.Color) {
	// Draw border (larger rounded rect)
	r.FillRoundRect(rect, radius, borderColor)

	// Draw background (smaller rounded rect)
	innerRect := math.NewRect(
		rect.X+borderWidth,
		rect.Y+borderWidth,
		rect.W-2*borderWidth,
		rect.H-2*borderWidth,
	)
	innerRadius := radius - borderWidth
	if innerRadius < 0 {
		innerRadius = 0
	}
	r.FillRoundRect(innerRect, innerRadius, bgColor)
}

// FillCircle draws a filled circle using SDF
func (r *Renderer) FillCircle(center math.Vec2, radius float32, color math.Color) {
	// Create a rect that contains the circle
	rect := math.NewRect(
		center.X-radius,
		center.Y-radius,
		radius*2,
		radius*2,
	)

	// Use rounded rect with radius = half width (makes a circle)
	r.FillRoundRect(rect, radius, color)
}

// StrokeCircle draws a circle outline
func (r *Renderer) StrokeCircle(center math.Vec2, radius, width float32, color math.Color) {
	// Outer circle
	r.FillCircle(center, radius, color)

	// Inner circle (cutout)
	innerRadius := radius - width
	if innerRadius > 0 {
		r.FillCircle(center, innerRadius, math.NewColor(1, 1, 1, 1)) // Background color
	}
}

// DrawLine draws a line between two points
func (r *Renderer) DrawLine(x1, y1, x2, y2, width float32, color math.Color) {
	// Calculate direction
	dx := x2 - x1
	dy := y2 - y1
	length := float32(stdmath.Sqrt(float64(dx*dx + dy*dy)))

	if length < 0.001 {
		return
	}

	// Normalize
	nx := dx / length
	ny := dy / length

	// Perpendicular
	px := -ny * width * 0.5
	py := nx * width * 0.5

	// Create quad
	verts := [4]Vertex{
		{X: x1 + px, Y: y1 + py, U: 0, V: 0, R: color.R, G: color.G, B: color.B, A: color.A},
		{X: x1 - px, Y: y1 - py, U: 1, V: 0, R: color.R, G: color.G, B: color.B, A: color.A},
		{X: x2 - px, Y: y2 - py, U: 1, V: 1, R: color.R, G: color.G, B: color.B, A: color.A},
		{X: x2 + px, Y: y2 + py, U: 0, V: 1, R: color.R, G: color.G, B: color.B, A: color.A},
	}

	shaderProg := r.shaderMgr.GetShader("color")
	r.addQuad(shaderProg, 0, nil, verts)
}

// DrawCheckmark draws a checkmark icon
func (r *Renderer) DrawCheckmark(rect math.Rect, size float32, color math.Color) {
	center := rect.Center()

	// Calculate checkmark points (simplified)
	x1 := center.X - size*0.4
	y1 := center.Y
	x2 := center.X - size*0.1
	y2 := center.Y + size*0.3
	x3 := center.X + size*0.4
	y3 := center.Y - size*0.3

	// Draw lines
	r.DrawLine(x1, y1, x2, y2, 2, color)
	r.DrawLine(x2, y2, x3, y3, 2, color)
}

// DrawShadow draws a shadow effect (simplified)
func (r *Renderer) DrawShadow(rect math.Rect, offset math.Vec2, blur float32, color math.Color) {
	// Simplified shadow: draw multiple offset rectangles with decreasing alpha
	steps := 3
	for i := 0; i < steps; i++ {
		alpha := color.A * (1 - float32(i)/float32(steps))
		shadowColor := math.NewColor(color.R, color.G, color.B, alpha)

		offsetX := offset.X * float32(i+1) / float32(steps)
		offsetY := offset.Y * float32(i+1) / float32(steps)

		shadowRect := math.NewRect(
			rect.X+offsetX,
			rect.Y+offsetY,
			rect.W,
			rect.H,
		)

		r.FillRoundRect(shadowRect, 4, shadowColor)
	}
}

// FillLinearGradient fills a rectangle with a linear gradient
func (r *Renderer) FillLinearGradient(rect math.Rect, start, end math.Vec2, startColor, endColor math.Color) {
	shaderProg := r.shaderMgr.GetShader("gradient")
	uniforms := UniformSet{
		"uColorStart": Vec4Uniform(startColor.R, startColor.G, startColor.B, startColor.A),
		"uColorEnd":   Vec4Uniform(endColor.R, endColor.G, endColor.B, endColor.A),
		"uStart":      Vec2Uniform(start.X, start.Y),
		"uEnd":        Vec2Uniform(end.X, end.Y),
	}

	// Draw quad
	uv := math.NewRect(0, 0, 1, 1)
	verts := QuadVertices(rect, uv, math.NewColor(1, 1, 1, 1))
	r.addQuad(shaderProg, 0, uniforms, verts)
}

// FillCircleFilled draws a filled circle using SDF
func (r *Renderer) FillCircleFilled(center math.Vec2, radius float32, color math.Color) {
	// Use round rect with radius = half width (makes a circle)
	r.FillCircle(center, radius, color)
}

// StrokeCircleOutline draws a circle outline using SDF
func (r *Renderer) StrokeCircleOutline(center math.Vec2, radius, width float32, color math.Color) {
	// Draw outer circle
	r.FillCircle(center, radius, color)
	// Cut out inner circle
	innerRadius := radius - width
	if innerRadius > 0 {
		r.FillCircle(center, innerRadius, math.NewColor(1, 1, 1, 1))
	}
}

// DrawArc draws an arc (portion of a circle)
func (r *Renderer) DrawArc(center math.Vec2, radius, width, startAngle, endAngle float32, color math.Color) {
	// Convert angles to radians
	startRad := startAngle * stdmath.Pi / 180
	endRad := endAngle * stdmath.Pi / 180

	// Number of segments
	segments := 32
	angleStep := (endRad - startRad) / float32(segments)

	// Draw arc using line segments
	for i := 0; i < segments; i++ {
		a1 := startRad + float32(i)*angleStep
		a2 := startRad + float32(i+1)*angleStep

		x1 := center.X + radius*float32(stdmath.Cos(float64(a1)))
		y1 := center.Y + radius*float32(stdmath.Sin(float64(a1)))
		x2 := center.X + radius*float32(stdmath.Cos(float64(a2)))
		y2 := center.Y + radius*float32(stdmath.Sin(float64(a2)))

		r.DrawLine(x1, y1, x2, y2, width, color)
	}
}

// DrawCircleArc draws a filled arc (pie shape)
func (r *Renderer) DrawCircleArc(center math.Vec2, radius, startAngle, endAngle float32, color math.Color) {
	// Convert angles to radians
	startRad := startAngle * stdmath.Pi / 180
	endRad := endAngle * stdmath.Pi / 180

	// Number of segments
	segments := 32
	angleStep := (endRad - startRad) / float32(segments)

	// Create vertices for triangle fan
	vertices := make([]math.Vec2, segments+2)
	vertices[0] = center // Center vertex

	for i := 0; i <= segments; i++ {
		angle := startRad + float32(i)*angleStep
		x := center.X + radius*float32(stdmath.Cos(float64(angle)))
		y := center.Y + radius*float32(stdmath.Sin(float64(angle)))
		vertices[i+1] = math.NewVec2(x, y)
	}

	// Draw triangles
	for i := 1; i <= segments; i++ {
		v0 := vertices[0]
		v1 := vertices[i]
		v2 := vertices[i+1]

		// Create triangle using three lines
		r.DrawLine(v0.X, v0.Y, v1.X, v1.Y, 1, color)
		r.DrawLine(v1.X, v1.Y, v2.X, v2.Y, 1, color)
		r.DrawLine(v2.X, v2.Y, v0.X, v0.Y, 1, color)
	}
}
