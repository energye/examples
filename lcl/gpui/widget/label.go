// Package widget provides UI widgets
package widget

import (
	"github.com/energye/examples/lcl/gpui/core/math"
	"github.com/energye/examples/lcl/gpui/render/font"
	"github.com/energye/examples/lcl/gpui/render/pipeline"
	"github.com/energye/examples/lcl/gpui/style/color"
)

// Label displays static text
type Label struct {
	BaseWidget
	text   string
	font   *font.Font
	col    math.Color
}

// NewLabel creates a new label
func NewLabel(text string, f *font.Font) *Label {
	return &Label{
		BaseWidget: NewBaseWidget(),
		text:       text,
		font:       f,
		col:        color.TextPrimary,
	}
}

// Text returns the label text
func (l *Label) Text() string {
	return l.text
}

// SetText sets the label text
func (l *Label) SetText(text string) {
	l.text = text
}

// SetColor sets the text color
func (l *Label) SetColor(c math.Color) {
	l.col = c
}

// SetFont sets the font
func (l *Label) SetFont(f *font.Font) {
	l.font = f
}

// Render renders the label
func (l *Label) Render(renderer *pipeline.Renderer) {
	if !l.visible || l.text == "" {
		return
	}

	if l.font == nil {
		return
	}

	// Draw text
	renderer.DrawText(l.text, l.bounds.X, l.bounds.Y, l.font, l.col)
}

// MeasureText returns the size needed for the text
func (l *Label) MeasureText() (float32, float32) {
	if l.font == nil {
		return 0, 0
	}
	return l.font.MeasureText(l.text)
}
