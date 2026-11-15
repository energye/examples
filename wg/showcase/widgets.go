package main

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/energye/lcl/types"
	"github.com/energye/lcl/types/colors"
)

func rect(l, t, r, b int32) types.TRect {
	return types.TRect{Left: l, Top: t, Right: r, Bottom: b}
}

func inset(r types.TRect, d int32) types.TRect {
	return rect(r.Left+d, r.Top+d, r.Right-d, r.Bottom-d)
}

func clampF(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// ---- Header ----

type Header struct {
	baseWidget
	badgeText string
	badgeAt   time.Time
	buf       []byte
}

func NewHeader(r types.TRect) *Header {
	return &Header{baseWidget: baseWidget{name: "标题栏", rect: r}}
}

func (w *Header) Paint(p *Painter, s *Scene) {
	p.Text(rect(w.rect.Left, w.rect.Top, w.rect.Left+s.S(760), w.rect.Top+s.S(34)),
		"单画布 · 复杂动态图形 / 半透明 / 异形 / 圆角 渲染示例", s.theme.Text, true, -19)
	p.Text(rect(w.rect.Left, w.rect.Top+s.S(34), w.rect.Left+s.S(900), w.rect.Top+s.S(58)),
		"一张画布、一个帧时钟、纯 Go 组件；RGBA 图层用 PixelData 批量写入并按样式共享缓存",
		s.theme.Muted, true, -13)
	badge := rect(w.rect.Right-s.S(238), w.rect.Top+s.S(6), w.rect.Right, w.rect.Top+s.S(40))
	p.RoundRect(badge, s.S(12), s.theme.Panel, s.theme.Edge)
	// Refresh the FPS badge 4x/second instead of allocating a new string on
	// every one of 60 frames. The visible value is already a 1s average.
	if w.badgeText == "" || time.Since(w.badgeAt) >= 250*time.Millisecond {
		w.buf = append(w.buf[:0], "  FPS "...)
		w.buf = strconv.AppendFloat(w.buf, s.FPS, 'f', 0, 64)
		w.buf = append(w.buf, "  ·  "...)
		w.buf = append(w.buf, s.ThemeName()...)
		w.buf = append(w.buf, "  "...)
		w.badgeText = string(w.buf)
		w.badgeAt = time.Now()
	}
	p.Text(inset(badge, s.S(6)), w.badgeText, s.theme.Accent3, false, -14)
}

// ---- GlassCard: anti-aliased rounded, semi-transparent RGBA layer ----

type GlassCard struct {
	baseWidget
	title  string
	radius int
}

func NewGlassCard(name, title string, r types.TRect, radius int) *GlassCard {
	return &GlassCard{baseWidget: baseWidget{name: name, rect: r}, title: title, radius: radius}
}

// IsStatic: the glass layer, title and divider never depend on the phase.
func (w *GlassCard) IsStatic() bool { return true }

func (w *GlassCard) Paint(p *Painter, s *Scene) {
	r := w.rect
	rad := int(s.S(int32(w.radius)))
	key := fmt.Sprintf("glass|%s|%d|%d|%d|%.2f", s.ThemeName(), r.Width(), r.Height(), rad, s.Opacity)
	l := s.cache.Get(key, r.Width(), r.Height(), func(sf *Surface) {
		bakeGlass(sf, s.theme, rad, s.Opacity, 0.25)
	})
	p.Layer(l, r.Left, r.Top)
	if w.title != "" {
		p.Text(rect(r.Left+s.S(16), r.Top+s.S(10), r.Left+s.S(360), r.Top+s.S(36)),
			w.title, s.theme.Text, true, -14)
		p.Line(r.Left+s.S(16), r.Top+s.S(40), r.Right-s.S(16), r.Top+s.S(40), s.theme.Edge, 1)
	}
}

// ---- GlowBlob: irregular translucent shape, slowly drifting ----

type GlowBlob struct {
	baseWidget
	lobes int
	seed  float64
}

func NewGlowBlob(name string, r types.TRect, lobes int, seed float64) *GlowBlob {
	return &GlowBlob{baseWidget: baseWidget{name: name, rect: r}, lobes: lobes, seed: seed}
}

// IsStatic: in the baked-static-layer mode the glow is a fixed background
// decoration. WG_STATIC=0 restores the original per-frame drift (see below).
func (w *GlowBlob) IsStatic() bool { return true }

func (w *GlowBlob) Paint(p *Painter, s *Scene) {
	r := w.rect
	key := fmt.Sprintf("glow|%s|%d|%d|%d|%.2f", s.ThemeName(), r.Width(), r.Height(), w.lobes, s.Opacity)
	l := s.cache.Get(key, r.Width(), r.Height(), func(sf *Surface) {
		bakeGlow(sf, s.theme, s.Opacity*0.55, w.lobes)
	})
	var dx, dy int32
	if !s.useStatic {
		dx = int32(28 * math.Sin(s.Phase*0.35+w.seed))
		dy = int32(18 * math.Cos(s.Phase*0.27+w.seed*1.7))
	}
	p.Layer(l, r.Left+dx, r.Top+dy)
}

// ---- WaveChart: dynamic curve + irregular filled area ----

type WaveChart struct {
	baseWidget
	points int
	pts    []types.TPoint // reused every frame: zero allocations in the hot path
	fill   []types.TPoint
}

func NewWaveChart(name string, r types.TRect) *WaveChart {
	return &WaveChart{baseWidget: baseWidget{name: name, rect: r}, points: 180}
}

func (w *WaveChart) Paint(p *Painter, s *Scene) {
	r := inset(w.rect, s.S(14))
	mid := r.Top + r.Height()/2
	amp := float64(r.Height()) * 0.30
	if cap(w.pts) < w.points {
		w.pts = make([]types.TPoint, w.points)
	} else {
		w.pts = w.pts[:w.points]
	}
	pts := w.pts
	for i := 0; i < w.points; i++ {
		f := float64(i) / float64(w.points-1)
		x := r.Left + int32(f*float64(r.Width()))
		v := math.Sin(f*13+s.Phase*2.1)*0.62 +
			math.Sin(f*29-s.Phase*1.4)*0.26 +
			math.Sin(f*5+s.Phase*0.7)*0.30
		y := mid + int32(v*amp)
		pts[i] = types.TPoint{X: x, Y: y}
	}
	// irregular filled area under the curve (one Polygon, not N rectangles)
	fill := w.fill[:0]
	fill = append(fill, pts...)
	fill = append(fill, types.TPoint{X: r.Right, Y: r.Bottom}, types.TPoint{X: r.Left, Y: r.Bottom})
	w.fill = fill
	p.GradientPolygon(fill,
		colors.RGBToColor(54, 108, 184), colors.RGBToColor(24, 46, 88),
		s.theme.Edge, false)
	for i := 0; i < 4; i++ {
		p.Line(r.Left, r.Top+int32(i)*r.Height()/3, r.Right, r.Top+int32(i)*r.Height()/3, s.theme.Grid, 1)
	}
	p.GlowPolyline(pts, s.theme.Accent, 2)
	// moving marker
	idx := int(math.Mod(s.Phase*40, float64(w.points)))
	mp := pts[idx]
	p.Ellipse(rect(mp.X-5, mp.Y-5, mp.X+5, mp.Y+5), s.theme.Warm, s.theme.Warm)
	p.Text(rect(r.Left, r.Top-s.S(2), r.Left+s.S(240), r.Top+s.S(20)),
		"180 点 / 1 次 Polyline", s.theme.Muted, true, -12)
}

// ---- LineChart: grid + 3 animated series ----

type LineChart struct {
	baseWidget
	series [3][]types.TPoint
}

func NewLineChart(name string, r types.TRect) *LineChart {
	return &LineChart{baseWidget: baseWidget{name: name, rect: r}}
}

func (w *LineChart) Paint(p *Painter, s *Scene) {
	r := inset(w.rect, s.S(18))
	for i := 0; i <= 4; i++ {
		y := r.Top + int32(i)*r.Height()/4
		p.Line(r.Left, y, r.Right, y, s.theme.Grid, 1)
	}
	for i := 0; i <= 6; i++ {
		x := r.Left + int32(i)*r.Width()/6
		p.Line(x, r.Top, x, r.Bottom, s.theme.Grid, 1)
	}
	series := []types.TColor{s.theme.Accent, s.theme.Accent3, s.theme.Warm}
	const n = 120
	for k, col := range series {
		if cap(w.series[k]) < n {
			w.series[k] = make([]types.TPoint, n)
		} else {
			w.series[k] = w.series[k][:n]
		}
		pts := w.series[k]
		for i := 0; i < n; i++ {
			f := float64(i) / float64(n-1)
			v := math.Sin(f*7+s.Phase*0.9+float64(k)*1.7)*0.30 +
				math.Sin(f*19-s.Phase*0.6+float64(k))*0.12 + float64(k)*0.12
			pts[i] = types.TPoint{
				X: r.Left + int32(f*float64(r.Width())),
				Y: r.Top + r.Height()/2 - int32(v*float64(r.Height())*0.8),
			}
		}
		p.GlowPolyline(pts, col, 2)
	}
	p.Text(rect(r.Left, r.Top-s.S(2), r.Left+s.S(260), r.Top+s.S(20)),
		"3 条实时曲线 / 360 点", s.theme.Muted, true, -12)
}

// ---- PetalBlob: non-rectangular polar shape + bezier ribbon ----

type PetalBlob struct {
	baseWidget
	poly   []types.TPoint
	star   []types.TPoint
	bezier []types.TPoint
}

func NewPetalBlob(name string, r types.TRect) *PetalBlob {
	return &PetalBlob{baseWidget: baseWidget{name: name, rect: r}}
}

// starPointsInto appends into a reusable buffer; the zero-allocation hot path
// never calls make(). starPoints stays as a convenience wrapper for tests.
func starPointsInto(dst []types.TPoint, cx, cy, rOut, rIn int32, n int, rot float64) []types.TPoint {
	pts := dst[:0]
	for i := 0; i < n*2; i++ {
		rad := rIn
		if i%2 == 0 {
			rad = rOut
		}
		a := rot + float64(i)*math.Pi/float64(n)
		pts = append(pts, types.TPoint{
			X: cx + int32(math.Cos(a)*float64(rad)),
			Y: cy + int32(math.Sin(a)*float64(rad)),
		})
	}
	return pts
}

func starPoints(cx, cy, rOut, rIn int32, n int, rot float64) []types.TPoint {
	return starPointsInto(nil, cx, cy, rOut, rIn, n, rot)
}

func (w *PetalBlob) Paint(p *Painter, s *Scene) {
	r := inset(w.rect, s.S(12))
	cx, cy := (r.Left+r.Right)/2, (r.Top+r.Bottom)/2
	rad := minInt(int(r.Width()), int(r.Height()))/2 - 8
	// irregular polar boundary (a "flower", not a rectangle or circle)
	const n = 96
	if cap(w.poly) < n {
		w.poly = make([]types.TPoint, n)
	} else {
		w.poly = w.poly[:n]
	}
	poly := w.poly
	for i := 0; i < n; i++ {
		a := float64(i) * 2 * math.Pi / n
		k := 0.74 + 0.22*math.Cos(6*a+s.Phase*0.8) + 0.06*math.Sin(11*a-s.Phase)
		poly[i] = types.TPoint{
			X: cx + int32(math.Cos(a)*float64(rad)*k),
			Y: cy + int32(math.Sin(a)*float64(rad)*k),
		}
	}
	p.Polygon(poly, s.theme.Accent2, s.theme.Edge, false)
	// rotating star on top
	w.star = starPointsInto(w.star, cx, cy, int32(rad)/2, int32(rad)/5, 8, -s.Phase*0.6)
	p.Polygon(w.star, s.theme.Warm, s.theme.Text, false)
	p.Ellipse(rect(cx-9, cy-9, cx+9, cy+9), s.theme.Text, s.theme.Text)
	// Two continuous cubic segments. p2/p3/p4 are collinear so the join is
	// C1-smooth instead of the sharp corner of the original control points.
	if len(w.bezier) == 0 {
		w.bezier = []types.TPoint{
			{X: r.Left, Y: r.Bottom - 10},
			{X: r.Left + r.Width()/4, Y: r.Top + 10},
			{X: r.Left + r.Width()/2 - 30, Y: r.Bottom - 10},
			{X: r.Left + r.Width()/2, Y: r.Bottom - 10},
			{X: r.Left + r.Width()/2 + 30, Y: r.Bottom - 10},
			{X: r.Left + 3*r.Width()/4, Y: r.Top + 10},
			{X: r.Right, Y: r.Bottom - 10},
		}
	}
	p.PolyBezier(w.bezier, s.theme.Bg, s.theme.Accent3, false, true)
	p.Text(rect(r.Left, r.Top-s.S(2), r.Left+s.S(300), r.Top+s.S(20)),
		"极坐标花瓣 + 旋转星形 + 贝塞尔", s.theme.Muted, true, -12)
}

// ---- RingGauge: arcs, sectors, center text ----

type RingGauge struct {
	baseWidget
	star    []types.TPoint
	pctBuf  []byte
	pctText string
	pctVal  int
}

func NewRingGauge(name string, r types.TRect) *RingGauge {
	return &RingGauge{baseWidget: baseWidget{name: name, rect: r}}
}

func (w *RingGauge) Paint(p *Painter, s *Scene) {
	r := inset(w.rect, s.S(14))
	size := minInt(int(r.Width()), int(r.Height())) - int(s.S(20))
	cx, cy := (r.Left+r.Right)/2, (r.Top+r.Bottom)/2
	outer := rect(cx-int32(size)/2, cy-int32(size)/2, cx+int32(size)/2, cy+int32(size)/2)
	p.Ellipse(outer, s.theme.Panel, s.theme.Edge)

	prog := 0.55 + 0.28*math.Sin(s.Phase*0.55)
	p.Arc(inset(outer, s.S(8)), -90, 360, s.theme.Grid, 14)
	p.Arc(inset(outer, s.S(8)), -90, prog*360, s.theme.Accent, 14)
	p.Arc(inset(outer, s.S(20)), -90+prog*360, 42, s.theme.Warm, 6)

	inner := inset(outer, s.S(34))
	p.Ellipse(inner, s.theme.Panel, s.theme.Edge)
	// Only re-allocate the label when the integer percentage actually changes.
	if pct := int(prog * 100); w.pctText == "" || pct != w.pctVal {
		w.pctBuf = strconv.AppendInt(w.pctBuf[:0], int64(pct), 10)
		w.pctBuf = append(w.pctBuf, '%')
		w.pctText = string(w.pctBuf)
		w.pctVal = pct
	}
	// Keep the big percentage in the UPPER part of the inner disc; the small
	// caption used to be centred on top of it (both rects overlapped).
	p.Text(rect(inner.Left, inner.Top, inner.Right, inner.Bottom-s.S(40)),
		w.pctText, s.theme.Text, false, -24)
	p.Text(rect(inner.Left, inner.Bottom-s.S(38), inner.Right, inner.Bottom-s.S(14)),
		"实时负载", s.theme.Muted, false, -12)

	// decorative rotating star to the right
	sx := r.Right - s.S(62)
	sy := r.Top + s.S(58)
	w.star = starPointsInto(w.star, sx, sy, 26, 11, 7, s.Phase*0.8)
	p.Polygon(w.star, s.theme.Accent3, s.theme.Text, false)
}

// ---- BarChart: rounded animated bars ----

type BarChart struct {
	baseWidget
}

func NewBarChart(name string, r types.TRect) *BarChart {
	return &BarChart{baseWidget: baseWidget{name: name, rect: r}}
}

func (w *BarChart) Paint(p *Painter, s *Scene) {
	r := inset(w.rect, s.S(18))
	baseY := r.Bottom - 24
	const n = 12
	gap := int32(7)
	bw := (r.Width() - gap*int32(n-1)) / n
	for i := 0; i < n; i++ {
		v := 0.5 + 0.5*math.Sin(s.Phase*1.15+float64(i)*0.55)
		h := int32(v * float64(r.Height()-50))
		x := r.Left + int32(i)*(bw+gap)
		col := s.theme.Accent
		if i%3 == 1 {
			col = s.theme.Accent2
		} else if i%3 == 2 {
			col = s.theme.Accent3
		}
		p.RoundRect(rect(x, baseY-h, x+bw, baseY), 5, col, col)
		p.RoundRect(rect(x, baseY-h, x+bw, baseY-h+6), 3, s.theme.Text, s.theme.Text)
	}
	p.Line(r.Left, baseY+2, r.Right, baseY+2, s.theme.Edge, 1)
	p.Text(rect(r.Left, r.Top-s.S(2), r.Left+s.S(300), r.Top+s.S(20)),
		"12 根圆角柱 / 动态高度", s.theme.Muted, true, -12)
}

// ---- DonutChart: pie sectors with a hollow center ----

var donutLabels = [4]string{"系列 1", "系列 2", "系列 3", "系列 4"}

type DonutChart struct {
	baseWidget
}

func NewDonutChart(name string, r types.TRect) *DonutChart {
	return &DonutChart{baseWidget: baseWidget{name: name, rect: r}}
}

func (w *DonutChart) Paint(p *Painter, s *Scene) {
	r := inset(w.rect, s.S(14))
	size := minInt(int(r.Width()), int(r.Height())) - 16
	cx, cy := r.Left+int32(size)/2+40, (r.Top+r.Bottom)/2
	outer := rect(cx-int32(size)/2, cy-int32(size)/2, cx+int32(size)/2, cy+int32(size)/2)
	segs := []struct {
		v float64
		c types.TColor
	}{
		{0.34 + 0.05*math.Sin(s.Phase*0.6), s.theme.Accent},
		{0.26, s.theme.Accent2},
		{0.22, s.theme.Accent3},
		{0.18 - 0.05*math.Sin(s.Phase*0.6), s.theme.Warm},
	}
	start := -90.0
	for _, sg := range segs {
		sweep := sg.v * 360
		p.Pie(outer, start, sweep, sg.c, s.theme.Panel)
		start += sweep
	}
	inner := inset(outer, s.S(32))
	p.Ellipse(inner, s.theme.Panel, s.theme.Panel)
	p.Text(inner, "4 段", s.theme.Text, false, -16)
	// legend
	lx := outer.Right + s.S(24)
	step := s.S(26)
	for i, sg := range segs {
		y := r.Top + s.S(26) + int32(i)*step
		q := s.S(14)
		p.RoundRect(rect(lx, y, lx+q, y+q), s.S(3), sg.c, sg.c)
		p.Text(rect(lx+s.S(20), y-s.S(2), lx+s.S(140), y+s.S(18)),
			donutLabels[i], s.theme.Muted, true, -12)
	}
}

// ---- PerfPanel: real measured numbers ----

type PerfPanel struct {
	baseWidget
}

func NewPerfPanel(name string, r types.TRect) *PerfPanel {
	return &PerfPanel{baseWidget: baseWidget{name: name, rect: r}}
}

func (w *PerfPanel) Paint(p *Painter, s *Scene) {
	r := inset(w.rect, s.S(16))
	step := s.S(22)
	for i, line := range s.StatusLines() {
		c := s.theme.Text
		if i > 1 {
			c = s.theme.Muted
		}
		y := r.Top + int32(i)*step
		p.Text(rect(r.Left, y, r.Right, y+s.S(21)), line, c, true, -13)
	}
}

// ---- Controls: rounded buttons, slider, toggle, progress ----

type miniCtl struct {
	label string
	r     types.TRect
	kind  string
}

type Controls struct {
	baseWidget
	items       []miniCtl
	designItems []miniCtl // design-space child rects, kept for relayout
	designRect  types.TRect
	hoverIdx    int
	dragging    bool
	toggleOn    bool
}

func NewControls(name string, r types.TRect) *Controls {
	c := &Controls{
		baseWidget: baseWidget{name: name, rect: r, hitAble: true},
		hoverIdx:   -1, toggleOn: true, designRect: r,
	}
	left, top := r.Left+16, r.Top+50
	c.items = append(c.items,
		miniCtl{"暂停 / 继续", rect(left, top, left+130, top+34), "pause"},
		miniCtl{"切换主题", rect(left+140, top, left+270, top+34), "theme"},
		miniCtl{"重置缓存", rect(left+280, top, left+410, top+34), "reset"},
		miniCtl{"", rect(left, top+58, left+410, top+78), "slider"},
		miniCtl{"动画开关", rect(left, top+96, left+150, top+128), "toggle"},
		miniCtl{"", rect(left+166, top+104, left+410, top+124), "progress"},
	)
	c.designItems = append(c.designItems, c.items...)
	return c
}

// SetRect re-maps the child controls together with the card, so hit testing
// keeps working after a window resize.
func (c *Controls) SetRect(r types.TRect) {
	d := c.designRect
	dw, dh := float64(d.Width()), float64(d.Height())
	if dw <= 0 || dh <= 0 {
		c.baseWidget.rect = r
		return
	}
	sx, sy := float64(r.Width())/dw, float64(r.Height())/dh
	c.baseWidget.rect = r
	for i := range c.items {
		q := c.designItems[i].r
		c.items[i].r = types.TRect{
			Left:   r.Left + int32(math.Round(float64(q.Left-d.Left)*sx)),
			Top:    r.Top + int32(math.Round(float64(q.Top-d.Top)*sy)),
			Right:  r.Left + int32(math.Round(float64(q.Right-d.Left)*sx)),
			Bottom: r.Top + int32(math.Round(float64(q.Bottom-d.Top)*sy)),
		}
	}
}

func (c *Controls) itemAt(x, y int32) int {
	for i, it := range c.items {
		if x >= it.r.Left && x <= it.r.Right && y >= it.r.Top && y <= it.r.Bottom {
			return i
		}
	}
	return -1
}

func (c *Controls) Paint(p *Painter, s *Scene) {
	p.Text(rect(c.rect.Left+s.S(16), c.rect.Top+s.S(10),
		c.rect.Right-s.S(16), c.rect.Top+s.S(38)),
		"自绘控件（全部无句柄）", s.theme.Text, true, -14)
	for i, it := range c.items {
		switch it.kind {
		case "pause", "theme", "reset":
			bg := s.theme.Panel
			if i == c.hoverIdx {
				bg = s.theme.Accent
			}
			p.RoundRect(it.r, s.S(8), bg, s.theme.Edge)
			p.Text(it.r, it.label, s.theme.Text, false, -13)
		case "slider":
			p.RoundRect(it.r, s.S(8), s.theme.Edge, s.theme.Edge)
			w := int32(float64(it.r.Width()) * (s.Opacity - 0.12) / 0.88)
			p.RoundRect(rect(it.r.Left, it.r.Top, it.r.Left+w, it.r.Bottom), s.S(8), s.theme.Accent, s.theme.Accent)
			kx := it.r.Left + w - s.S(7)
			p.Ellipse(rect(kx, it.r.Top-s.S(5), kx+s.S(14), it.r.Bottom+s.S(5)), s.theme.Text, s.theme.Text)
			p.Text(rect(it.r.Left, it.r.Top-s.S(24), it.r.Right, it.r.Top-s.S(4)),
				"拖动调整全局不透明度（触发 RGBA 图层重烘焙）", s.theme.Muted, true, -12)
		case "toggle":
			bg := s.theme.Edge
			col := s.theme.Muted
			if c.toggleOn {
				bg, col = s.theme.Accent3, s.theme.Text
			}
			p.RoundRect(rect(it.r.Left, it.r.Top, it.r.Left+s.S(52), it.r.Bottom), s.S(14), bg, bg)
			kx := it.r.Left + s.S(4)
			if c.toggleOn {
				kx = it.r.Left + s.S(30)
			}
			p.Ellipse(rect(kx, it.r.Top+s.S(2), kx+s.S(18), it.r.Bottom-s.S(2)), s.theme.Text, s.theme.Text)
			p.Text(rect(it.r.Left+s.S(60), it.r.Top, it.r.Right, it.r.Bottom), it.label, col, true, -12)
		case "progress":
			p.RoundRect(it.r, s.S(8), s.theme.Edge, s.theme.Edge)
			v := 0.5 + 0.5*math.Sin(s.Phase*0.9)
			w := int32(v * float64(it.r.Width()))
			p.RoundRect(rect(it.r.Left, it.r.Top, it.r.Left+w, it.r.Bottom), s.S(8), s.theme.Accent2, s.theme.Accent2)
		}
	}
}

func (c *Controls) Handle(e *UIEvent, s *Scene) bool {
	switch e.Kind {
	case EvMouseMove:
		if c.dragging {
			sl := c.items[3].r
			s.SetOpacity(0.12 + 0.88*clampF(float64(e.X-sl.Left)/float64(sl.Width()), 0, 1))
			return true
		}
		idx := c.itemAt(e.X, e.Y)
		if idx != c.hoverIdx {
			c.hoverIdx = idx
			return true
		}
	case EvMouseLeave:
		c.hoverIdx = -1
		return true
	case EvMouseDown:
		idx := c.itemAt(e.X, e.Y)
		if idx < 0 {
			return false
		}
		switch c.items[idx].kind {
		case "pause":
			s.SetPaused(!s.Paused)
		case "theme":
			s.ToggleTheme()
		case "reset":
			s.cache.InvalidateAll()
			s.InvalidateStatic()
		case "toggle":
			c.toggleOn = !c.toggleOn
			s.SetPaused(!c.toggleOn)
		case "slider":
			c.dragging = true
			sl := c.items[idx].r
			s.SetOpacity(0.12 + 0.88*clampF(float64(e.X-sl.Left)/float64(sl.Width()), 0, 1))
		}
		return true
	case EvMouseUp:
		c.dragging = false
		return true
	}
	return false
}

// ---- Legend ----

type Legend struct {
	baseWidget
}

func NewLegend(name string, r types.TRect) *Legend {
	return &Legend{baseWidget: baseWidget{name: name, rect: r}}
}

func (w *Legend) IsStatic() bool { return true }

func (w *Legend) Paint(p *Painter, s *Scene) {
	lines := []string{
		"渲染管线：Form(唯一 HWND) → 无句柄画布 → 纯 Go 组件",
		"抗锯齿：几何按覆盖率就地混合 Pf24bit，文字仍用 GDI ClearType",
		"帧调度：失效驱动 60Hz 被动投递，静止零帧 / 脏区单帧",
		"静态层：卡片与光斑一次烘焙，逐帧只重画 9 个动态组件",
		"异形：Polygon / Polyline / PolyBezier / Pie / Arc + 渐变与发光",
		"交互：按钮 / 滑块 / 空格暂停 / T 切主题 / + - 调不透明度",
	}
	r := inset(w.rect, s.S(16))
	step := s.S(24)
	for i, line := range lines {
		c := s.theme.Muted
		if i == 0 {
			c = s.theme.Text
		}
		y := r.Top + int32(i)*step
		p.Text(rect(r.Left, y, r.Right, y+s.S(22)), line, c, true, -12)
	}
}
