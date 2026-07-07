package pipeline

import "github.com/energye/examples/lcl/gpui/core/math"

// LinearGradient describes a rectangle-local linear gradient.
type LinearGradient struct {
	Start      math.Vec2
	End        math.Vec2
	StartColor math.Color
	EndColor   math.Color
}

// Shadow describes a simple box shadow.
type Shadow struct {
	Offset math.Vec2
	Blur   float32
	Color  math.Color
}

// BoxStyle describes the common visual model used by UI components.
type BoxStyle struct {
	Background     math.Color
	UseGradient    bool
	Gradient       LinearGradient
	BorderColor    math.Color
	BorderWidth    float32
	Radius         float32
	Shadows        []Shadow
	SkipBackground bool
	SkipBorder     bool
}

// DrawBox draws a rounded rectangle with optional shadows, background, gradient, and border.
func (r *Renderer) DrawBox(rect math.Rect, style BoxStyle) {
	if rect.W <= 0 || rect.H <= 0 {
		return
	}

	for _, shadow := range style.Shadows {
		if shadow.Color.A > 0 {
			r.DrawShadow(rect, shadow.Offset, shadow.Blur, shadow.Color)
		}
	}

	if !style.SkipBackground {
		if style.UseGradient {
			r.FillLinearGradient(rect, style.Gradient.Start, style.Gradient.End, style.Gradient.StartColor, style.Gradient.EndColor)
		} else if style.Radius > 0 {
			r.FillRoundRect(rect, style.Radius, style.Background)
		} else {
			r.FillRect(rect, style.Background)
		}
	}

	if !style.SkipBorder && style.BorderWidth > 0 && style.BorderColor.A > 0 {
		if style.Radius > 0 {
			r.StrokeRoundRect(rect, style.Radius, style.BorderWidth, style.BorderColor)
		} else {
			r.StrokeRect(rect, style.BorderWidth, style.BorderColor)
		}
	}
}
