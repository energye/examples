package main

import (
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
	"github.com/energye/lcl/types/colors"
)

// Theme is the single source of truth for all colors: widgets never pick
// colors themselves, which is the point of a unified render layer.
type Theme struct {
	Bg, Panel, PanelAlt, Edge, Accent, AccentDown, Text, Muted, Compose, Caret types.TColor
}

func darkTheme() Theme {
	return Theme{
		Bg:         colors.RGBToColor(28, 30, 34),
		Panel:      colors.RGBToColor(42, 45, 51),
		PanelAlt:   colors.RGBToColor(52, 56, 64),
		Edge:       colors.RGBToColor(86, 92, 102),
		Accent:     colors.RGBToColor(66, 133, 244),
		AccentDown: colors.RGBToColor(38, 96, 190),
		Text:       colors.RGBToColor(235, 237, 240),
		Muted:      colors.RGBToColor(150, 156, 166),
		Compose:    colors.RGBToColor(255, 198, 92),
		Caret:      colors.RGBToColor(255, 255, 255),
	}
}

func lightTheme() Theme {
	return Theme{
		Bg:         colors.RGBToColor(242, 244, 248),
		Panel:      colors.RGBToColor(255, 255, 255),
		PanelAlt:   colors.RGBToColor(232, 236, 243),
		Edge:       colors.RGBToColor(176, 184, 196),
		Accent:     colors.RGBToColor(52, 116, 224),
		AccentDown: colors.RGBToColor(32, 82, 172),
		Text:       colors.RGBToColor(32, 36, 42),
		Muted:      colors.RGBToColor(96, 104, 116),
		Compose:    colors.RGBToColor(196, 104, 0),
		Caret:      colors.RGBToColor(24, 28, 34),
	}
}

// Painter is the ONLY type that talks to lcl.ICanvas. Every widget draws
// through it, so DPI scaling, clipping, layer caching and new primitives
// can be added in one place.
type Painter struct {
	Canvas lcl.ICanvas
	Theme  *Theme
}

func (p *Painter) ready() bool { return p != nil && p.Canvas != nil && p.Canvas.IsValid() }

func (p *Painter) FillRect(r types.TRect, c types.TColor) {
	if !p.ready() {
		return
	}
	b := p.Canvas.BrushToBrush()
	b.SetStyle(types.BsSolid)
	b.SetColor(c)
	p.Canvas.FillRectWithRect(r)
}

func (p *Painter) RoundRect(r types.TRect, radius int32, bg, border types.TColor) {
	if !p.ready() {
		return
	}
	b := p.Canvas.BrushToBrush()
	b.SetStyle(types.BsSolid)
	b.SetColor(bg)
	p.Canvas.PenToPen().SetColor(border)
	p.Canvas.RoundRectWithRectIntX2(r, radius, radius)
}

func (p *Painter) Gradient(r types.TRect, from, to types.TColor) {
	if !p.ready() {
		return
	}
	p.Canvas.GradientFill(r, from, to, types.GdVertical)
}

// Shadow is a cheap fake drop shadow: a translucent-looking offset rectangle
// built from a darker solid color. A production renderer would use an
// ILazIntfImage layer with per-pixel alpha instead.
func (p *Painter) Shadow(r types.TRect, radius int32) {
	if !p.ready() {
		return
	}
	shift := types.TRect{Left: r.Left + 3, Top: r.Top + 4, Right: r.Right + 3, Bottom: r.Bottom + 4}
	p.RoundRect(shift, radius, p.Theme.Bg, p.Theme.Bg)
}

func (p *Painter) Text(r types.TRect, text string, c types.TColor, alignLeft bool) int32 {
	if !p.ready() {
		return 0
	}
	b := p.Canvas.BrushToBrush()
	b.SetStyle(types.BsClear)
	p.Canvas.FontToFont().SetColor(c)
	tw := p.Canvas.TextWidthWithStr(text)
	th := p.Canvas.TextHeightWithStr(text)
	x := r.Left + (r.Width()-tw)/2
	if alignLeft {
		x = r.Left + 10
	}
	p.Canvas.TextOutWithIntX2Str(x, r.Top+(r.Height()-th)/2, text)
	b.SetStyle(types.BsSolid)
	return x + tw
}

// TextAt draws text at an absolute x and returns the x just after it.
// Used by the editor to render text run-by-run (selection + caret layout).
func (p *Painter) TextAt(x, y int32, text string, c types.TColor) int32 {
	if !p.ready() {
		return x
	}
	b := p.Canvas.BrushToBrush()
	b.SetStyle(types.BsClear)
	p.Canvas.FontToFont().SetColor(c)
	p.Canvas.TextOutWithIntX2Str(x, y, text)
	b.SetStyle(types.BsSolid)
	return x + p.Canvas.TextWidthWithStr(text)
}

func (p *Painter) TextWidth(text string) int32 {
	if !p.ready() {
		return 0
	}
	return p.Canvas.TextWidthWithStr(text)
}

func (p *Painter) TextHeight(text string) int32 {
	if !p.ready() {
		return 0
	}
	return p.Canvas.TextHeightWithStr(text)
}

func (p *Painter) Caret(x, y, h int32, c types.TColor) {
	if !p.ready() {
		return
	}
	b := p.Canvas.BrushToBrush()
	b.SetStyle(types.BsSolid)
	b.SetColor(c)
	p.Canvas.FillRectWithIntX4(x, y, x+2, y+h)
}
