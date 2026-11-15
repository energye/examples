package main

import (
	"math"
	"runtime"
	"unsafe"

	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
)

// Painter is the ONLY type that touches lcl.ICanvas. Every widget draws
// through it, so DPI, clipping, antialiasing and a future GPU backend can be
// added in one place. Calls is an estimate of Go->DLL round trips per frame.
//
// Two backends, selected by Painter.frame:
//
//	frame == nil : every primitive goes straight to GDI (hard edges, the
//	               WG_AA=0 / WG_FAST=1 / WG_STATIC=0 comparison mode)
//	frame != nil : geometry is rasterised in Go with sub-pixel coverage into
//	               the opaque frame buffer; TEXT IS DEFERRED into textCmds
//	               and replayed with GDI after LoadFromIntfImage, so Chinese
//	               text keeps ClearType while shapes become antialiased.
type Painter struct {
	Canvas lcl.ICanvas
	Theme  *Theme
	Calls  int
	frame  *FramePixels

	textCmds  []textCmd
	bezierPts []types.TPoint
	two       [2]types.TPoint

	// Per-frame state cache: LCL setters are FFI calls, so skip redundant ones.
	// Reset() clears these flags on every paint to stay correct even if the
	// canvas gets re-bound to a different DC.
	fColor  types.TColor
	fHeight int32
	fOK     bool
	bColor  types.TColor
	bStyle  types.TFPBrushStyle
	bOK     bool
	pColor  types.TColor
	pWidth  int32
	pOK     bool

	twCache  map[metricKey]int32
	thCache  map[metricKey]int32
	fitCache map[fitKey]string

	// Scale keeps text size and metrics in step with the adaptive layout.
	// Without it, shrinking the window shrinks the cards but leaves the
	// fixed-pixel fonts/line spacing overflowing their cards.
	Scale float64
}

func (p *Painter) scaleHeight(h int32) int32 {
	if h == 0 || p.Scale <= 0 || p.Scale == 1 {
		return h
	}
	v := int32(math.Round(float64(h) * p.Scale))
	if h < 0 && v > -8 {
		v = -8 // never shrink a CJK glyph below legibility
	}
	if h > 0 && v < 8 {
		v = 8
	}
	return v
}

// ClearMetrics drops cached text widths after a layout scale change.
func (p *Painter) ClearMetrics() {
	p.twCache, p.thCache, p.fitCache = nil, nil, nil
}

type fitKey struct {
	height int32
	avail  int32
	text   string
}

type textCmd struct {
	r         types.TRect
	text      string
	c         types.TColor
	height    int32
	atX, atY  int32
	alignLeft bool
	isAt      bool
}

type metricKey struct {
	height int32
	text   string
}

func (p *Painter) ok() bool {
	return p != nil && p.Canvas != nil && p.Canvas.IsValid()
}

func (p *Painter) Reset() {
	p.Calls = 0
	p.fOK, p.bOK, p.pOK = false, false, false
	p.textCmds = p.textCmds[:0]
}

func (p *Painter) brush(c types.TColor, style types.TFPBrushStyle) {
	if p.bOK && p.bColor == c && p.bStyle == style {
		return
	}
	b := p.Canvas.BrushToBrush()
	// LCL 的 TBrush.SetColor 会把 Style 重置回 bsSolid（实测：先 SetStyle 再
	// SetColor 后读回 Style=0，TextOut 会画出一整块不透明色底）。所以颜色必须
	// 先写、样式最后写，顺序不能交换。
	b.SetColor(c)
	b.SetStyle(style)
	p.Calls += 2
	p.bOK, p.bColor, p.bStyle = true, c, style
}

func (p *Painter) pen(c types.TColor, width int32) {
	if p.pOK && p.pColor == c && p.pWidth == width {
		return
	}
	q := p.Canvas.PenToPen()
	q.SetStyle(types.PsSolid)
	q.SetColor(c)
	q.SetWidth(width)
	p.Calls += 3
	p.pOK, p.pColor, p.pWidth = true, c, width
}

func (p *Painter) font(c types.TColor, height int32) {
	sh := p.scaleHeight(height)
	if p.fOK && p.fColor == c && (sh == 0 || p.fHeight == sh) {
		return
	}
	f := p.Canvas.FontToFont()
	f.SetColor(c)
	if sh != 0 {
		f.SetHeight(sh)
	}
	p.Calls += 2
	p.fOK, p.fColor = true, c
	if sh != 0 {
		p.fHeight = sh
	}
}

func (p *Painter) pointPtr(pts []types.TPoint) types.PPoint {
	if len(pts) == 0 {
		return 0
	}
	return types.PPoint(uintptr(unsafe.Pointer(&pts[0])))
}

func (p *Painter) FillRect(r types.TRect, c types.TColor) {
	if !p.ok() {
		return
	}
	if p.frame != nil {
		p.frame.fillRect(r, c)
		return
	}
	p.brush(c, types.BsSolid)
	p.Canvas.FillRectWithRect(r)
	p.Calls++
}

func (p *Painter) RoundRect(r types.TRect, radius int32, bg, border types.TColor) {
	if !p.ok() {
		return
	}
	if p.frame != nil {
		p.frame.roundRect(r, radius, bg, border)
		return
	}
	p.brush(bg, types.BsSolid)
	p.pen(border, 1)
	p.Canvas.RoundRectWithRectIntX2(r, radius, radius)
	p.Calls++
}

func (p *Painter) Gradient(r types.TRect, from, to types.TColor, dir types.TGradientDirection) {
	if !p.ok() {
		return
	}
	p.Canvas.GradientFill(r, from, to, dir)
	p.Calls++
}

// Text is deferred in AA mode: it is recorded, then replayed with GDI after
// the software-rasterised geometry has been pushed into the bitmap. Text
// metrics stay identical; the return value is the text advance (unused by
// the widgets today, but kept for API compatibility).
func (p *Painter) Text(r types.TRect, text string, c types.TColor, alignLeft bool, height int32) int32 {
	if !p.ok() {
		return 0
	}
	if p.frame != nil {
		p.textCmds = append(p.textCmds, textCmd{
			r: r, text: text, c: c, height: height, alignLeft: alignLeft,
		})
		return r.Left
	}
	p.font(c, height)
	p.brush(c, types.BsClear)
	text = p.fitText(text, r.Width())
	tw := p.TextWidth(text)
	th := p.TextHeight(text)
	x := r.Left + (r.Width()-tw)/2
	if alignLeft {
		x = r.Left
	}
	p.Canvas.TextOutWithIntX2Str(x, r.Top+(r.Height()-th)/2, text)
	p.Calls++
	return x + tw
}

func (p *Painter) TextAt(x, y int32, text string, c types.TColor, height int32) int32 {
	if !p.ok() {
		return x
	}
	if p.frame != nil {
		p.textCmds = append(p.textCmds, textCmd{
			text: text, c: c, height: height, atX: x, atY: y, isAt: true,
		})
		return x
	}
	p.font(c, height)
	p.brush(c, types.BsClear)
	p.Canvas.TextOutWithIntX2Str(x, y, text)
	p.Calls++
	return x + p.TextWidth(text)
}

// FlushText replays deferred text with GDI (ClearType) onto the frame bitmap.
func (p *Painter) FlushText() {
	if p.frame == nil || len(p.textCmds) == 0 || !p.ok() {
		p.textCmds = p.textCmds[:0]
		return
	}
	for i := range p.textCmds {
		q := &p.textCmds[i]
		p.font(q.c, q.height)
		p.brush(q.c, types.BsClear)
		if q.isAt {
			p.Canvas.TextOutWithIntX2Str(q.atX, q.atY, q.text)
		} else {
			txt := p.fitText(q.text, q.r.Width())
			tw := p.TextWidth(txt)
			th := p.TextHeight(txt)
			x := q.r.Left + (q.r.Width()-tw)/2
			if q.alignLeft {
				x = q.r.Left
			}
			p.Canvas.TextOutWithIntX2Str(x, q.r.Top+(q.r.Height()-th)/2, txt)
		}
		p.Calls++
	}
	p.textCmds = p.textCmds[:0]
}

// fitText clamps a label to the widget's rect. TextOut has no clipping, so a
// long line would otherwise paint straight across the neighbouring card.
// Results are memoised by (font height, available width, text): static labels
// are measured once, and the per-frame status lines are a small fixed set, so
// the hot path stays allocation-free.
func (p *Painter) fitText(text string, avail int32) string {
	if avail <= 0 || text == "" {
		return text
	}
	k := fitKey{height: p.fHeight, avail: avail, text: text}
	if v, ok := p.fitCache[k]; ok {
		return v
	}
	out := text
	if p.TextWidth(text) > avail {
		rs := []rune(text)
		ell := "…"
		ew := p.TextWidth(ell)
		lo, hi := 0, len(rs)
		for lo < hi {
			mid := (lo + hi + 1) / 2
			if p.TextWidth(string(rs[:mid]))+ew <= avail {
				lo = mid
			} else {
				hi = mid - 1
			}
		}
		out = string(rs[:lo]) + ell
	}
	if p.fitCache == nil {
		p.fitCache = make(map[fitKey]string)
	}
	p.fitCache[k] = out
	return out
}

func (p *Painter) TextWidth(text string) int32 {
	if !p.ok() {
		return 0
	}
	k := metricKey{height: p.fHeight, text: text}
	if v, ok := p.twCache[k]; ok {
		return v
	}
	v := p.Canvas.TextWidthWithStr(text)
	p.Calls++
	if p.twCache == nil {
		p.twCache = make(map[metricKey]int32)
	}
	p.twCache[k] = v
	return v
}

func (p *Painter) TextHeight(text string) int32 {
	if !p.ok() {
		return 0
	}
	k := metricKey{height: p.fHeight, text: text}
	if v, ok := p.thCache[k]; ok {
		return v
	}
	v := p.Canvas.TextHeightWithStr(text)
	p.Calls++
	if p.thCache == nil {
		p.thCache = make(map[metricKey]int32)
	}
	p.thCache[k] = v
	return v
}

func (p *Painter) Polygon(pts []types.TPoint, fill, border types.TColor, winding bool) {
	if !p.ok() || len(pts) < 3 {
		return
	}
	if p.frame != nil {
		p.frame.polygon(pts, fill, border, 1)
		return
	}
	p.brush(fill, types.BsSolid)
	p.pen(border, 1)
	p.Canvas.Polygon(p.pointPtr(pts), int32(len(pts)), winding)
	p.Calls++
	runtime.KeepAlive(pts)
}

// GradientPolygon fills an irregular shape with a vertical two-stop gradient.
func (p *Painter) GradientPolygon(pts []types.TPoint, top, bottom, border types.TColor, winding bool) {
	if !p.ok() || len(pts) < 3 {
		return
	}
	if p.frame != nil {
		p.frame.polygonGradient(pts, top, bottom, border, 1)
		return
	}
	p.brush(top, types.BsSolid)
	p.pen(border, 1)
	p.Canvas.Polygon(p.pointPtr(pts), int32(len(pts)), winding)
	p.Calls++
	runtime.KeepAlive(pts)
}

// GlowPolyline lays down a wide, low-intensity halo plus the crisp core in
// ONE software-rasterizer pass (see FramePixels.lineCov). No extra FFI, no
// extra alpha layer, and no second distance field over the whole curve.
func (p *Painter) GlowPolyline(pts []types.TPoint, c types.TColor, width int32) {
	if !p.ok() || len(pts) < 2 {
		return
	}
	if p.frame == nil {
		p.Polyline(pts, c, width)
		return
	}
	p.frame.polylineGlow(pts, c, width, width+4, 0.24)
}

// Polyline draws a whole curve in ONE round trip (GDI) or as a chain of
// distance-field AA segments (software backend).
func (p *Painter) Polyline(pts []types.TPoint, c types.TColor, width int32) {
	if !p.ok() || len(pts) < 2 {
		return
	}
	if p.frame != nil {
		p.frame.polyline(pts, c, width)
		return
	}
	p.pen(c, width)
	p.Canvas.Polyline(p.pointPtr(pts), int32(len(pts)))
	p.Calls++
	runtime.KeepAlive(pts)
}

func (p *Painter) PolyBezier(pts []types.TPoint, fill, border types.TColor, filled, continuous bool) {
	if !p.ok() || len(pts) < 3 {
		return
	}
	if p.frame != nil {
		p.bezierPts = p.bezierPts[:0]
		const samples = 20
		for i := 0; i+3 < len(pts); i += 3 {
			x0, y0 := float64(pts[i].X), float64(pts[i].Y)
			x1, y1 := float64(pts[i+1].X), float64(pts[i+1].Y)
			x2, y2 := float64(pts[i+2].X), float64(pts[i+2].Y)
			x3, y3 := float64(pts[i+3].X), float64(pts[i+3].Y)
			for k := 0; k <= samples; k++ {
				t := float64(k) / samples
				u := 1 - t
				a, b, c, d := u*u*u, 3*u*u*t, 3*u*t*t, t*t*t
				p.bezierPts = append(p.bezierPts, types.TPoint{
					X: int32(a*x0 + b*x1 + c*x2 + d*x3 + 0.5),
					Y: int32(a*y0 + b*y1 + c*y2 + d*y3 + 0.5),
				})
			}
		}
		if filled {
			p.frame.polygon(p.bezierPts, fill, border, 1)
		} else {
			p.frame.polylineGlow(p.bezierPts, border, 2, 7, 0.26)
		}
		return
	}
	p.brush(fill, types.BsSolid)
	p.pen(border, 1)
	p.Canvas.PolyBezierWithPointIntBoolX2(p.pointPtr(pts), int32(len(pts)), filled, continuous)
	p.Calls++
	runtime.KeepAlive(pts)
}

func (p *Painter) Ellipse(r types.TRect, bg, border types.TColor) {
	if !p.ok() {
		return
	}
	if p.frame != nil {
		p.frame.ellipse(r, bg, border, 1)
		return
	}
	p.brush(bg, types.BsSolid)
	p.pen(border, 1)
	p.Canvas.EllipseWithRect(r)
	p.Calls++
}

func (p *Painter) Arc(r types.TRect, startDeg, sweepDeg float64, c types.TColor, width int32) {
	if !p.ok() {
		return
	}
	if p.frame != nil {
		p.frame.arc(r, startDeg, sweepDeg, c, width)
		return
	}
	p.pen(c, width)
	p.Canvas.ArcWithIntX6(r.Left, r.Top, r.Right, r.Bottom,
		int32(startDeg*16), int32(sweepDeg*16))
	p.Calls++
}

func (p *Painter) Pie(r types.TRect, startDeg, sweepDeg float64, fill, border types.TColor) {
	if !p.ok() {
		return
	}
	if p.frame != nil {
		p.frame.pie(r, startDeg, sweepDeg, fill, border)
		return
	}
	p.brush(fill, types.BsSolid)
	p.pen(border, 1)
	p.Canvas.RadialPieWithIntX6(r.Left, r.Top, r.Right, r.Bottom,
		int32(startDeg*16), int32(sweepDeg*16))
	p.Calls++
}

func (p *Painter) Line(x1, y1, x2, y2 int32, c types.TColor, width int32) {
	p.two[0] = types.TPoint{X: x1, Y: y1}
	p.two[1] = types.TPoint{X: x2, Y: y2}
	p.Polyline(p.two[:], c, width)
}

func (p *Painter) Layer(l *AlphaLayer, x, y int32) {
	if !p.ok() || l == nil {
		return
	}
	l.Draw(p.Canvas, x, y)
	p.Calls++
}

func (p *Painter) Caret(x, y, h int32, c types.TColor) {
	if !p.ok() {
		return
	}
	p.brush(c, types.BsSolid)
	p.Canvas.FillRectWithIntX4(x, y, x+2, y+h)
	p.Calls++
}
