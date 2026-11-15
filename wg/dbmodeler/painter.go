package main

// Painter 是这个示例里唯一接触 lcl.ICanvas 的类型。所有组件、画布、菜单、
// 弹窗都只通过它下笔，因此裁剪、缩放、主题、文本度量可以在一处收口。
//
// 与 showcase 的 AA 光栅器不同：建模器是静态 UI，没有逐帧动画，用
// canvas_ime 的 OnPaint 直绘路径即可（唯一 HWND 仍是 Form，画布无句柄）。
// 这里额外做了一层「软裁剪」：每个绘制原语都与当前 Clip 求交，面板在不透明
// 覆盖顺序里最后绘制，等价于把 ER 画布的溢出内容裁掉，且不需要 GDI clip region。

import (
	"math"
	"runtime"
	"unsafe"

	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
)

const (
	AlignLeft = iota
	AlignCenter
	AlignRight
)

type fitKey struct {
	height int32
	bold   bool
	avail  int32
	scale  int32
	text   string
}

type Painter struct {
	Canvas   lcl.ICanvas
	Theme    *Theme
	FontName string
	Scale    float64 // 画布缩放：字体高度同步缩放

	clip      types.TRect
	clipOn    bool
	clipStack []clipState

	fColor types.TColor
	fH     int32
	fBold  bool
	fName  string
	fOK    bool
	bColor types.TColor
	bStyle types.TFPBrushStyle
	bOK    bool
	pColor types.TColor
	pWidth int32
	pOK    bool

	fitCache map[fitKey]string
}

func (p *Painter) ok() bool {
	return p != nil && p.Canvas != nil && p.Canvas.IsValid()
}

func (p *Painter) Reset() {
	p.fOK, p.bOK, p.pOK = false, false, false
	p.fitCache = nil
	p.clipStack = p.clipStack[:0]
}

type clipState struct {
	r  types.TRect
	on bool
}

// PushClip / PopClip 供输入框这类需要「局部裁剪 + 画完还原」的组件使用，
// 避免文本滚动时溢出到相邻面板。
func (p *Painter) PushClip(r types.TRect) {
	p.clipStack = append(p.clipStack, clipState{r: p.clip, on: p.clipOn})
	p.SetClip(r)
}

func (p *Painter) PopClip() {
	n := len(p.clipStack)
	if n == 0 {
		p.ClearClip()
		return
	}
	st := p.clipStack[n-1]
	p.clipStack = p.clipStack[:n-1]
	p.clip, p.clipOn = st.r, st.on
}

func (p *Painter) SetClip(r types.TRect) {
	p.clip, p.clipOn = r, true
}
func (p *Painter) ClearClip() { p.clipOn = false }
func (p *Painter) ClipRect() types.TRect {
	return p.clip
}

func (p *Painter) clipped(r types.TRect) (types.TRect, bool) {
	if !p.clipOn {
		return r, true
	}
	c := p.clip
	if r.Right <= c.Left || r.Left >= c.Right || r.Bottom <= c.Top || r.Top >= c.Bottom {
		return r, false
	}
	if r.Left < c.Left {
		r.Left = c.Left
	}
	if r.Top < c.Top {
		r.Top = c.Top
	}
	if r.Right > c.Right {
		r.Right = c.Right
	}
	if r.Bottom > c.Bottom {
		r.Bottom = c.Bottom
	}
	return r, r.Width() > 0 && r.Height() > 0
}

// Visible 判断矩形当前是否落在激活的裁剪区内，供列表/弹窗跳过不可见行。
func (p *Painter) Visible(r types.TRect) bool {
	if !p.clipOn {
		return true
	}
	_, ok := p.clipped(r)
	return ok
}

func (p *Painter) fullInside(r types.TRect) bool {
	if !p.clipOn {
		return true
	}
	c := p.clip
	return r.Left >= c.Left && r.Top >= c.Top && r.Right <= c.Right && r.Bottom <= c.Bottom
}

// ---- 状态设置：画刷必须先 SetColor 再 SetStyle（LCL 的 SetColor 会重置 Style）----

func (p *Painter) brush(c types.TColor, style types.TFPBrushStyle) {
	if p.bOK && p.bColor == c && p.bStyle == style {
		return
	}
	b := p.Canvas.BrushToBrush()
	b.SetColor(c)
	b.SetStyle(style)
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
	p.pOK, p.pColor, p.pWidth = true, c, width
}

func (p *Painter) font(c types.TColor, height int32, bold bool) {
	scale := p.Scale
	if scale <= 0 {
		scale = 1
	}
	h := -int32(math.Round(float64(height) * scale))
	if h > -8 {
		h = -8
	}
	name := p.FontName
	if name == "" {
		name = "Microsoft YaHei UI"
	}
	if p.fOK && p.fColor == c && p.fH == h && p.fBold == bold && p.fName == name {
		return
	}
	f := p.Canvas.FontToFont()
	f.SetName(name)
	f.SetColor(c)
	f.SetHeight(h)
	if bold {
		f.SetStyle(types.NewSet(types.FsBold))
	} else {
		f.SetStyle(types.NewSet())
	}
	p.fOK, p.fColor, p.fH, p.fBold, p.fName = true, c, h, bold, name
}

// ---- 基础图元 ----

func (p *Painter) FillRect(r types.TRect, c types.TColor) {
	if !p.ok() {
		return
	}
	r, ok := p.clipped(r)
	if !ok {
		return
	}
	p.brush(c, types.BsSolid)
	p.Canvas.FillRectWithRect(r)
}

// StrokeRect 用四个实心细条画边框，对软裁剪友好（GDI Rectangle 无法裁剪）。
func (p *Painter) StrokeRect(r types.TRect, c types.TColor, w int32) {
	if !p.ok() {
		return
	}
	r, ok := p.clipped(r)
	if !ok {
		return
	}
	if w < 1 {
		w = 1
	}
	p.FillRect(types.TRect{Left: r.Left, Top: r.Top, Right: r.Right, Bottom: r.Top + w}, c)
	p.FillRect(types.TRect{Left: r.Left, Top: r.Bottom - w, Right: r.Right, Bottom: r.Bottom}, c)
	p.FillRect(types.TRect{Left: r.Left, Top: r.Top, Right: r.Left + w, Bottom: r.Bottom}, c)
	p.FillRect(types.TRect{Left: r.Right - w, Top: r.Top, Right: r.Right, Bottom: r.Bottom}, c)
}

func (p *Painter) RoundRect(r types.TRect, radius int32, bg, border types.TColor) {
	if !p.ok() {
		return
	}
	if _, ok := p.clipped(r); !ok {
		return
	}
	if !p.fullInside(r) {
		// 部分越过裁剪边界时用矩形降级绘制，避免圆角图元溢出到相邻面板。
		vis, ok := p.clipped(r)
		if !ok {
			return
		}
		p.FillRect(vis, bg)
		p.StrokeRect(vis, border, 1)
		return
	}
	p.brush(bg, types.BsSolid)
	p.pen(border, 1)
	p.Canvas.RoundRectWithRectIntX2(r, radius, radius)
}

func (p *Painter) Gradient(r types.TRect, from, to types.TColor) {
	if !p.ok() {
		return
	}
	r, ok := p.clipped(r)
	if !ok {
		return
	}
	p.Canvas.GradientFill(r, from, to, types.GdVertical)
	// GradientFill 会改写画刷状态，后续 FillRect 必须重新设置。
	p.bOK = false
}

func (p *Painter) pointPtr(pts []types.TPoint) types.PPoint {
	return types.PPoint(uintptr(unsafe.Pointer(&pts[0])))
}

func (p *Painter) Polyline(pts []types.TPoint, c types.TColor, width int32) {
	if !p.ok() || len(pts) < 2 {
		return
	}
	p.pen(c, width)
	p.Canvas.Polyline(p.pointPtr(pts), int32(len(pts)))
	runtime.KeepAlive(pts)
}

func (p *Painter) Line(x1, y1, x2, y2 int32, c types.TColor, width int32) {
	p.Polyline([]types.TPoint{{X: x1, Y: y1}, {X: x2, Y: y2}}, c, width)
}

func (p *Painter) Polygon(pts []types.TPoint, fill, border types.TColor) {
	if !p.ok() || len(pts) < 3 {
		return
	}
	p.brush(fill, types.BsSolid)
	p.pen(border, 1)
	p.Canvas.Polygon(p.pointPtr(pts), int32(len(pts)), false)
	runtime.KeepAlive(pts)
}

func (p *Painter) Ellipse(r types.TRect, bg, border types.TColor) {
	if !p.ok() {
		return
	}
	if _, ok := p.clipped(r); !ok {
		return
	}
	p.brush(bg, types.BsSolid)
	p.pen(border, 1)
	p.Canvas.EllipseWithRect(r)
}

func (p *Painter) Circle(cx, cy, radius int32, bg, border types.TColor) {
	p.Ellipse(types.TRect{Left: cx - radius, Top: cy - radius,
		Right: cx + radius, Bottom: cy + radius}, bg, border)
}

// Arrow 在 (x,y) 画一个朝向 dx/dy 的小三角，用于关系连线端点。
func (p *Painter) Arrow(x, y, dx int32, c types.TColor) {
	size := int32(5)
	if dx >= 0 {
		p.Polygon([]types.TPoint{
			{X: x, Y: y}, {X: x - size, Y: y - size}, {X: x - size, Y: y + size},
		}, c, c)
	} else {
		p.Polygon([]types.TPoint{
			{X: x, Y: y}, {X: x + size, Y: y - size}, {X: x + size, Y: y + size},
		}, c, c)
	}
}

// DashRect 画虚线选框（自绘而不是依赖 Pen 样式，缩放时线宽稳定）。
func (p *Painter) DashRect(r types.TRect, c types.TColor) {
	seg, gap := int32(6), int32(4)
	horiz := func(y int32) {
		for x := r.Left; x < r.Right; x += seg + gap {
			end := x + seg
			if end > r.Right {
				end = r.Right
			}
			p.Line(x, y, end, y, c, 1)
		}
	}
	vert := func(x int32) {
		for y := r.Top; y < r.Bottom; y += seg + gap {
			end := y + seg
			if end > r.Bottom {
				end = r.Bottom
			}
			p.Line(x, y, x, end, c, 1)
		}
	}
	horiz(r.Top)
	horiz(r.Bottom - 1)
	vert(r.Left)
	vert(r.Right - 1)
}

// ---- 文本 ----

func (p *Painter) TextWidth(text string, height int32, bold bool) int32 {
	if !p.ok() {
		return int32(len([]rune(text)) * int(height) / 2)
	}
	p.font(p.Theme.Text, height, bold)
	return p.Canvas.TextWidthWithStr(text)
}

func (p *Painter) TextHeight(text string, height int32, bold bool) int32 {
	if !p.ok() {
		return height
	}
	p.font(p.Theme.Text, height, bold)
	return p.Canvas.TextHeightWithStr(text)
}

func (p *Painter) fitText(text string, avail int32, height int32, bold bool) string {
	if avail <= 0 || text == "" {
		return text
	}
	k := fitKey{height: height, bold: bold, avail: avail,
		scale: int32(math.Round(p.Scale * 100)), text: text}
	if p.fitCache != nil {
		if v, ok := p.fitCache[k]; ok {
			return v
		}
	}
	// 不要在这里重设字体：调用方（Text）已经按指定颜色/字号设置好了，
	// fitText 若再设成 Theme.Text 会把标题栏白字、PK 金色等全部染成默认色。
	// 文本宽度只取决于字体名/字号/粗细，与颜色无关。
	out := text
	if p.Canvas.TextWidthWithStr(text) > avail {
		rs := []rune(text)
		ell := "…"
		ew := p.Canvas.TextWidthWithStr(ell)
		lo, hi := 0, len(rs)
		for lo < hi {
			mid := (lo + hi + 1) / 2
			if p.Canvas.TextWidthWithStr(string(rs[:mid]))+ew <= avail {
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

// Text 在一个 rect 里画单行文本；超出裁剪右边界时自动加省略号。
func (p *Painter) Text(r types.TRect, text string, c types.TColor, height, align int32, bold bool) {
	if !p.ok() || text == "" {
		return
	}
	p.font(c, height, bold)
	p.brush(c, types.BsClear)
	avail := r.Width()
	if p.clipOn && p.clip.Right > r.Left {
		avail = p.clip.Right - r.Left - 2
	}
	if avail < 4 {
		return
	}
	txt := p.fitText(text, avail, height, bold)
	tw := p.Canvas.TextWidthWithStr(txt)
	th := p.Canvas.TextHeightWithStr(txt)
	x := r.Left + 2
	if align == AlignCenter {
		x = r.Left + (r.Width()-tw)/2
	} else if align == AlignRight {
		x = r.Right - tw - 2
	}
	y := r.Top + (r.Height()-th)/2
	if p.clipOn && (y+th <= p.clip.Top || y >= p.clip.Bottom) {
		return
	}
	p.Canvas.TextOutWithIntX2Str(x, y, txt)
}

// TextAt 在绝对坐标画文本并返回右端 x，供自绘输入框逐 rune 排版。
func (p *Painter) TextAt(x, y int32, text string, c types.TColor, height int32, bold bool) int32 {
	if !p.ok() || text == "" {
		return x
	}
	p.font(c, height, bold)
	p.brush(c, types.BsClear)
	p.Canvas.TextOutWithIntX2Str(x, y, text)
	return x + p.Canvas.TextWidthWithStr(text)
}

func (p *Painter) Caret(x, y, h int32, c types.TColor) {
	p.FillRect(types.TRect{Left: x, Top: y, Right: x + 2, Bottom: y + h}, c)
}
