package pipeline

import "github.com/energye/examples/lcl/gpui/core/math"

// FillRule controls compound path filling.
type FillRule int

const (
	FillRuleNonZero FillRule = iota
	FillRuleEvenOdd
)

// SVGIcon stores parsed icon geometry and its source viewBox.
type SVGIcon struct {
	Path     *Path
	ViewBox  math.Rect
	FillRule FillRule
}

// NewSVGIcon parses SVG path data for later rendering.
func NewSVGIcon(pathData string, viewBox math.Rect, fillRule FillRule) (*SVGIcon, error) {
	path, err := ParseSVGPath(pathData)
	if err != nil {
		return nil, err
	}
	return &SVGIcon{
		Path:     path,
		ViewBox:  viewBox,
		FillRule: fillRule,
	}, nil
}

// Render draws the icon into dst using currentColor.
func (icon *SVGIcon) Render(renderer *Renderer, dst math.Rect, currentColor math.Color) {
	if icon == nil || icon.Path == nil || dst.W <= 0 || dst.H <= 0 || icon.ViewBox.W == 0 || icon.ViewBox.H == 0 {
		return
	}

	sx := dst.W / icon.ViewBox.W
	sy := dst.H / icon.ViewBox.H
	transform := math.Mat4{
		sx, 0, 0, 0,
		0, sy, 0, 0,
		0, 0, 1, 0,
		dst.X - icon.ViewBox.X*sx,
		dst.Y - icon.ViewBox.Y*sy,
		0, 1,
	}

	renderer.PushTransform(transform)
	if icon.FillRule == FillRuleEvenOdd {
		renderer.FillPathEvenOdd(icon.Path, currentColor)
	} else {
		renderer.FillPathNonZero(icon.Path, currentColor)
	}
	renderer.PopTransform()
}
