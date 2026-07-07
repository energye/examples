// Package math provides mathematical utilities for GPU UI
package math

import (
	"math"
)

// Vec2 represents a 2D vector
type Vec2 struct {
	X, Y float32
}

// NewVec2 creates a new 2D vector
func NewVec2(x, y float32) Vec2 {
	return Vec2{X: x, Y: y}
}

// Add adds two vectors
func (v Vec2) Add(other Vec2) Vec2 {
	return Vec2{X: v.X + other.X, Y: v.Y + other.Y}
}

// Sub subtracts two vectors
func (v Vec2) Sub(other Vec2) Vec2 {
	return Vec2{X: v.X - other.X, Y: v.Y - other.Y}
}

// Scale scales the vector by a scalar
func (v Vec2) Scale(s float32) Vec2 {
	return Vec2{X: v.X * s, Y: v.Y * s}
}

// Length returns the length of the vector
func (v Vec2) Length() float32 {
	return float32(math.Sqrt(float64(v.X*v.X + v.Y*v.Y)))
}

// Normalize returns the normalized vector
func (v Vec2) Normalize() Vec2 {
	l := v.Length()
	if l == 0 {
		return Vec2{}
	}
	return Vec2{X: v.X / l, Y: v.Y / l}
}

// Dot returns the dot product
func (v Vec2) Dot(other Vec2) float32 {
	return v.X*other.X + v.Y*other.Y
}

// Rect represents a 2D rectangle
type Rect struct {
	X, Y, W, H float32
}

// NewRect creates a new rectangle
func NewRect(x, y, w, h float32) Rect {
	return Rect{X: x, Y: y, W: w, H: h}
}

// Contains checks if a point is inside the rectangle
func (r Rect) Contains(x, y float32) bool {
	return x >= r.X && x <= r.X+r.W && y >= r.Y && y <= r.Y+r.H
}

// Expand expands the rectangle by the given amount
func (r Rect) Expand(amount float32) Rect {
	return Rect{
		X: r.X - amount,
		Y: r.Y - amount,
		W: r.W + 2*amount,
		H: r.H + 2*amount,
	}
}

// Shrink shrinks the rectangle by the given amount
func (r Rect) Shrink(horizontal, vertical float32) Rect {
	return Rect{
		X: r.X + horizontal,
		Y: r.Y + vertical,
		W: r.W - 2*horizontal,
		H: r.H - 2*vertical,
	}
}

// Center returns the center point of the rectangle
func (r Rect) Center() Vec2 {
	return Vec2{
		X: r.X + r.W/2,
		Y: r.Y + r.H/2,
	}
}

// Intersect returns the intersection of two rectangles
func (r Rect) Intersect(other Rect) Rect {
	x := max(r.X, other.X)
	y := max(r.Y, other.Y)
	w := min(r.X+r.W, other.X+other.W) - x
	h := min(r.Y+r.H, other.Y+other.H) - y

	if w < 0 || h < 0 {
		return Rect{}
	}

	return Rect{X: x, Y: y, W: w, H: h}
}

// Union returns the union of two rectangles
func (r Rect) Union(other Rect) Rect {
	x := min(r.X, other.X)
	y := min(r.Y, other.Y)
	w := max(r.X+r.W, other.X+other.W) - x
	h := max(r.Y+r.H, other.Y+other.H) - y

	return Rect{X: x, Y: y, W: w, H: h}
}

// Color represents an RGBA color
type Color struct {
	R, G, B, A float32
}

// NewColor creates a new color
func NewColor(r, g, b, a float32) Color {
	return Color{R: r, G: g, B: b, A: a}
}

// NewColorFromHex creates a color from hex value (0xRRGGBBAA)
func NewColorFromHex(hex uint32) Color {
	return Color{
		R: float32((hex>>24)&0xFF) / 255.0,
		G: float32((hex>>16)&0xFF) / 255.0,
		B: float32((hex>>8)&0xFF) / 255.0,
		A: float32(hex&0xFF) / 255.0,
	}
}

// Lighten lightens the color by the given amount (0-1)
func (c Color) Lighten(amount float32) Color {
	return Color{
		R: min(c.R+amount, 1.0),
		G: min(c.G+amount, 1.0),
		B: min(c.B+amount, 1.0),
		A: c.A,
	}
}

// Darken darkens the color by the given amount (0-1)
func (c Color) Darken(amount float32) Color {
	return Color{
		R: max(c.R-amount, 0.0),
		G: max(c.G-amount, 0.0),
		B: max(c.B-amount, 0.0),
		A: c.A,
	}
}

// WithAlpha returns a new color with the given alpha
func (c Color) WithAlpha(alpha float32) Color {
	return Color{
		R: c.R,
		G: c.G,
		B: c.B,
		A: alpha,
	}
}

// Lerp performs linear interpolation between two colors
func (c Color) Lerp(other Color, t float32) Color {
	return Color{
		R: c.R + (other.R-c.R)*t,
		G: c.G + (other.G-c.G)*t,
		B: c.B + (other.B-c.B)*t,
		A: c.A + (other.A-c.A)*t,
	}
}

// Mat4 represents a 4x4 matrix
type Mat4 [16]float32

// IdentityMatrix returns an identity matrix
func IdentityMatrix() Mat4 {
	return Mat4{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
}

// OrthoMatrix creates an orthographic projection matrix
func OrthoMatrix(left, right, bottom, top, near, far float32) Mat4 {
	return Mat4{
		2 / (right - left), 0, 0, 0,
		0, 2 / (top - bottom), 0, 0,
		0, 0, -2 / (far - near), 0,
		-(right + left) / (right - left), -(top + bottom) / (top - bottom), -(far + near) / (far - near), 1,
	}
}

// TranslationMatrix creates a translation matrix
func TranslationMatrix(x, y, z float32) Mat4 {
	return Mat4{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		x, y, z, 1,
	}
}

// ScaleMatrix creates a scale matrix
func ScaleMatrix(x, y, z float32) Mat4 {
	return Mat4{
		x, 0, 0, 0,
		0, y, 0, 0,
		0, 0, z, 0,
		0, 0, 0, 1,
	}
}

// RotationMatrix creates a rotation matrix around Z axis
func RotationMatrix(angle float32) Mat4 {
	c := float32(math.Cos(float64(angle)))
	s := float32(math.Sin(float64(angle)))
	return Mat4{
		c, s, 0, 0,
		-s, c, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
}

// Multiply multiplies two matrices
func (m Mat4) Multiply(other Mat4) Mat4 {
	var result Mat4
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			for k := 0; k < 4; k++ {
				result[i*4+j] += m[i*4+k] * other[k*4+j]
			}
		}
	}
	return result
}

// Helper functions
func min(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

func max(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}
