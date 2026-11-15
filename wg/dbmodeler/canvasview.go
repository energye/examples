package main

// ER 图画布：网格 / 表卡片 / 字段行 / 连接端口 / 正交关系连线 / 基数标记。
// 全部由 Painter 画在中间那块区域里，没有任何原生控件。
// 交互：拖表头移动、拖端口创建关系、空白拖拽平移、滚轮缩放、框选、
// 双击就地改名、右键菜单、折叠表、一键加字段。

import (
	"math"
	"time"

	"github.com/energye/lcl/types"
)

type dragMode uint8

const (
	dragNone dragMode = iota
	dragPan
	dragMove
	dragLink
	dragMarquee
)

type CanvasView struct {
	scene *Scene
	rect  types.TRect
	zoom  float64
	panX  int32
	panY  int32

	drag        dragMode
	startX      int32
	startY      int32
	lastX       int32
	lastY       int32
	dragTableID string
	lastMoveX   int32
	lastMoveY   int32
	moveDirty   bool
	linkTable   string
	linkColumn  string
	linkSide    int32 // 起始端口所在侧：+1 右 / -1 左
	marquee     types.TRect

	hoverTable string
	hoverCol   string
	hoverRel   string
	hoverPort  bool

	clickAt time.Time
	clickX  int32
	clickY  int32
}

func NewCanvasView(s *Scene) *CanvasView {
	return &CanvasView{scene: s, zoom: 1, panX: 30, panY: 30}
}

func (v *CanvasView) SetRect(r types.TRect) { v.rect = r }
func (v *CanvasView) Zoom() float64         { return v.zoom }
func (v *CanvasView) Pan() (int32, int32)   { return v.panX, v.panY }

func (v *CanvasView) sz(world float64) int32 {
	return int32(math.Round(world * v.zoom))
}

func (v *CanvasView) W2S(wx, wy float64) (int32, int32) {
	return v.rect.Left + v.panX + int32(math.Round(wx*v.zoom)),
		v.rect.Top + v.panY + int32(math.Round(wy*v.zoom))
}

func (v *CanvasView) S2W(sx, sy int32) (float64, float64) {
	return float64(sx-v.rect.Left-v.panX) / v.zoom,
		float64(sy-v.rect.Top-v.panY) / v.zoom
}

func (v *CanvasView) tableRect(t *Table) types.TRect {
	x, y := v.W2S(t.X, t.Y)
	return r4(x, y, x+v.sz(tableWidth), y+v.sz(TableHeight(t)))
}

func (v *CanvasView) headerRect(t *Table) types.TRect {
	r := v.tableRect(t)
	r.Bottom = r.Top + v.sz(38)
	return r
}

func (v *CanvasView) rowRect(t *Table, idx int) types.TRect {
	r := v.tableRect(t)
	top := r.Top + v.sz(38) + int32(idx)*v.sz(26)
	return r4(r.Left, top, r.Right, top+v.sz(26))
}

// portPos 返回某字段行左右两侧的连接圆点坐标。
func (v *CanvasView) portPos(t *Table, idx, side int) (int32, int32) {
	r := v.rowRect(t, idx)
	y := r.Top + r.Height()/2
	if side > 0 {
		return r.Right, y
	}
	return r.Left, y
}

func (v *CanvasView) columnIndex(t *Table, id string) int {
	for i := range t.Columns {
		if t.Columns[i].ID == id {
			return i
		}
	}
	return -1
}

// ---- 绘制 ----

func (v *CanvasView) Paint(p *Painter) {
	p.SetClip(v.rect)
	p.FillRect(v.rect, p.Theme.Bg)
	v.paintGrid(p)
	v.paintRelations(p)
	for i := range v.scene.model.Tables {
		v.paintTable(p, &v.scene.model.Tables[i])
	}
	if v.drag == dragLink {
		x1, y1 := v.linkStartPos()
		p.Line(x1, y1, v.lastX, v.lastY, p.Theme.Warn, 2)
		p.Circle(v.lastX, v.lastY, 4, p.Theme.Warn, p.Theme.Warn)
	}
	if v.drag == dragMarquee {
		r := v.marquee
		p.DashRect(r, p.Theme.Accent)
		p.FillRect(r4(r.Left, r.Top, r.Right, r.Top+1), p.Theme.AccentSoft)
	}
	v.paintHUD(p)
	p.ClearClip()
}

func (v *CanvasView) paintGrid(p *Painter) {
	step := v.sz(24)
	if step < 8 {
		step = 8
	}
	offX := v.panX % step
	offY := v.panY % step
	if offX < 0 {
		offX += step
	}
	if offY < 0 {
		offY += step
	}
	idx := 0
	for x := v.rect.Left + offX; x < v.rect.Right; x += step {
		c := p.Theme.Grid
		if (int(v.panX/step)+idx)%5 == 0 {
			c = p.Theme.GridMajor
		}
		p.Line(x, v.rect.Top, x, v.rect.Bottom, c, 1)
		idx++
	}
	idx = 0
	for y := v.rect.Top + offY; y < v.rect.Bottom; y += step {
		c := p.Theme.Grid
		if (int(v.panY/step)+idx)%5 == 0 {
			c = p.Theme.GridMajor
		}
		p.Line(v.rect.Left, y, v.rect.Right, y, c, 1)
		idx++
	}
}

func (v *CanvasView) paintHUD(p *Painter) {
	label := "ER 画布 · 模拟数据（未连接数据库）"
	w := p.TextWidth(label, 11, false) + 20
	box := r4(v.rect.Left+12, v.rect.Top+12, v.rect.Left+12+w, v.rect.Top+36)
	p.RoundRect(box, 5, p.Theme.Surface, p.Theme.Border)
	p.Text(insetR(box, 4), label, p.Theme.Muted, 11, AlignLeft, false)

	tip := "拖表头移动 · 拖字段圆点建关系 · 滚轮缩放 · 双击改名 · 右键菜单"
	tw := p.TextWidth(tip, 11, false)
	tipBox := r4(v.rect.Right-tw-24, v.rect.Bottom-32, v.rect.Right-12, v.rect.Bottom-10)
	if tipBox.Left > v.rect.Left+20 {
		p.Text(tipBox, tip, p.Theme.Muted, 11, AlignLeft, false)
	}
}

func (v *CanvasView) paintTable(p *Painter, t *Table) {
	r := v.tableRect(t)
	if !p.Visible(r4(r.Left, r.Top, r.Right+6, r.Bottom+6)) {
		return
	}
	sel := v.scene.sel
	selected := sel.TableID == t.ID || sel.Tables[t.ID]
	border := p.Theme.Border
	if selected {
		border = p.Theme.SelBorder
	}

	if p.Theme.Dark {
		p.FillRect(r4(r.Left+3, r.Top+4, r.Right+3, r.Bottom+4), rgb(10, 12, 18))
	} else {
		p.FillRect(r4(r.Left+3, r.Top+4, r.Right+3, r.Bottom+4), rgb(206, 214, 228))
	}
	p.RoundRect(r, 7, p.Theme.Surface, border)

	// 标题栏
	hdr := v.headerRect(t)
	gc := groupColor(*p.Theme, t.Group)
	p.FillRect(r4(hdr.Left+1, hdr.Top+1, hdr.Right-1, hdr.Bottom-1), gc)
	// 标题栏分上下两行：表名 / 分组+字段数，避免文字互相压叠。
	p.Text(r4(hdr.Left+v.sz(9), hdr.Top+1, hdr.Right-v.sz(64), hdr.Top+v.sz(21)),
		t.Name, rgb(255, 255, 255), 13, AlignLeft, true)
	p.Text(r4(hdr.Left+v.sz(9), hdr.Top+v.sz(21), hdr.Right-v.sz(48), hdr.Bottom),
		groupLabel(t.Group)+" · "+itoa(len(t.Columns))+" 字段",
		rgb(232, 240, 255), 9, AlignLeft, false)
	// 折叠指示用几何三角形（字体对 ▾/▸ 的支持不一致，会显示成方框）。
	ax, ay := hdr.Right-v.sz(36), hdr.Top+hdr.Height()/2
	s := v.sz(4)
	white := rgb(255, 255, 255)
	if t.Collapsed {
		p.Polygon([]types.TPoint{
			{X: ax - s/2, Y: ay - s}, {X: ax + s/2, Y: ay}, {X: ax - s/2, Y: ay + s},
		}, white, white)
	} else {
		p.Polygon([]types.TPoint{
			{X: ax - s, Y: ay - s/2}, {X: ax + s, Y: ay - s/2}, {X: ax, Y: ay + s/2},
		}, white, white)
	}
	p.Text(r4(hdr.Right-v.sz(25), hdr.Top, hdr.Right-v.sz(4), hdr.Bottom),
		"+", white, 15, AlignCenter, true)
	if t.Collapsed {
		return
	}

	for idx := range t.Columns {
		c := &t.Columns[idx]
		row := v.rowRect(t, idx)
		if !p.Visible(row) {
			continue
		}
		bg := p.Theme.Surface
		if idx%2 == 1 {
			bg = p.Theme.Bg
		}
		if sel.Kind == "column" && sel.TableID == t.ID && sel.ColumnID == c.ID {
			bg = p.Theme.Sel
		}
		if v.hoverTable == t.ID && v.hoverCol == c.ID && v.hoverRel == "" {
			if bg == p.Theme.Surface {
				bg = p.Theme.SurfaceAlt
			}
		}
		p.FillRect(r4(row.Left+1, row.Top, row.Right-1, row.Bottom), bg)

		pad := v.sz(9)
		// 键标记
		mark, markColor, markText := "", p.Theme.Muted, ""
		if c.PK {
			mark, markColor, markText = "◆", p.Theme.PK, "PK"
		} else if v.scene.model.IsForeignKey(t.ID, c.ID) {
			mark, markColor, markText = "◆", p.Theme.FK, "FK"
		}
		if mark != "" {
			p.Text(r4(row.Left+pad, row.Top, row.Left+pad+v.sz(15), row.Bottom),
				mark, markColor, 10, AlignLeft, true)
		}
		nameX := row.Left + pad + v.sz(17)
		p.Text(r4(nameX, row.Top, row.Right-v.sz(96), row.Bottom),
			c.Name, p.Theme.Text, 12, AlignLeft, c.PK)
		p.Text(r4(row.Right-v.sz(104), row.Top, row.Right-pad-v.sz(30), row.Bottom),
			c.TypeText(), p.Theme.Muted, 10, AlignRight, false)
		if markText != "" {
			p.Text(r4(row.Right-pad-v.sz(28), row.Top, row.Right-pad, row.Bottom),
				markText, markColor, 9, AlignRight, true)
		}
		// 连接端口（左右各一个）
		for _, side := range []int{-1, 1} {
			px, py := v.portPos(t, idx, side)
			hover := v.hoverPort && v.hoverTable == t.ID && v.hoverCol == c.ID
			col := p.Theme.Border
			if hover || (v.drag == dragLink && v.linkTable == t.ID && v.linkColumn == c.ID) {
				col = p.Theme.Warn
			} else if markText == "FK" && side < 0 {
				col = p.Theme.FK
			} else if c.PK && side < 0 {
				col = p.Theme.PK
			}
			p.Circle(px, py, v.sz(4)+1, col, p.Theme.Surface)
		}
	}
	// 底部分隔线
	p.FillRect(r4(r.Left+1, r.Bottom-v.sz(8), r.Right-1, r.Bottom-v.sz(8)+1), p.Theme.Border)
}

// relationPoints 生成正交折线，并把端点坐标记在返回值里。
func (v *CanvasView) relationPoints(rel *Relation) []types.TPoint {
	m := v.scene.model
	ft, tt := m.TableByID(rel.FromTable), m.TableByID(rel.ToTable)
	if ft == nil || tt == nil {
		return nil
	}
	fi := v.columnIndex(ft, rel.FromColumn)
	ti := v.columnIndex(tt, rel.ToColumn)
	if fi < 0 || ti < 0 || ft.Collapsed || tt.Collapsed {
		// 折叠时落到表头中点，避免关系线消失
		fr, tr := v.tableRect(ft), v.tableRect(tt)
		fy, ty := fr.Top+fr.Height()/2, tr.Top+tr.Height()/2
		fx, tx := fr.Right, tr.Left
		if tr.Left < fr.Left {
			fx, tx = fr.Left, tr.Right
		}
		return v.ortho(fx, fy, tx, ty)
	}
	fr, tr := v.tableRect(ft), v.tableRect(tt)
	fromRight := tr.Left >= fr.Right || (tr.Left+tr.Width()/2) >= (fr.Left+fr.Width()/2)
	sideFrom, sideTo := 1, -1
	if !fromRight {
		sideFrom, sideTo = -1, 1
	}
	x1, y1 := v.portPos(ft, fi, sideFrom)
	x2, y2 := v.portPos(tt, ti, sideTo)
	return v.ortho(x1, y1, x2, y2)
}

func (v *CanvasView) ortho(x1, y1, x2, y2 int32) []types.TPoint {
	out := v.sz(20)
	if out < 8 {
		out = 8
	}
	dir := int32(1)
	if x2 < x1 {
		dir = -1
	}
	if (x2-x1)*dir > out*2+12 {
		mid := x1 + (x2-x1)/2
		return []types.TPoint{
			{X: x1, Y: y1}, {X: x1 + dir*out, Y: y1},
			{X: mid, Y: y1}, {X: mid, Y: y2},
			{X: x2 - dir*out, Y: y2}, {X: x2, Y: y2},
		}
	}
	// 目标位于出发侧后方：绕一个 U 形
	bend := x1 + dir*out
	if (x2-x1)*dir < out {
		bend = x1 + dir*out*2
	}
	midY := y1 + (y2-y1)/2
	return []types.TPoint{
		{X: x1, Y: y1}, {X: bend, Y: y1},
		{X: bend, Y: midY}, {X: x2 - dir*out, Y: midY},
		{X: x2 - dir*out, Y: y2}, {X: x2, Y: y2},
	}
}

func (v *CanvasView) paintRelations(p *Painter) {
	for i := range v.scene.model.Relations {
		rel := &v.scene.model.Relations[i]
		pts := v.relationPoints(rel)
		if len(pts) < 2 {
			continue
		}
		selected := v.scene.sel.Kind == "relation" && v.scene.sel.RelationID == rel.ID
		c := p.Theme.Line
		w := int32(2)
		if selected {
			c, w = p.Theme.Accent, 3
		} else if v.hoverRel == rel.ID {
			c, w = p.Theme.AccentDown, 3
		}
		p.Polyline(pts, c, w)
		first, last := pts[0], pts[len(pts)-1]
		p.Circle(first.X, first.Y, 3, c, p.Theme.Surface)
		p.Circle(last.X, last.Y, 3, c, p.Theme.Surface)
		// 基数：父端 1，子端 N / 1
		childText := "N"
		if rel.Cardinality == "1:1" {
			childText = "1"
		}
		dx := int32(12)
		if last.X < first.X {
			dx = -12
		}
		v.paintCardinality(p, last.X-dx, last.Y, "1", c)
		v.paintCardinality(p, first.X+dx, first.Y, childText, c)
		if selected {
			label := rel.Name + "  " + rel.Cardinality
			mid := pts[len(pts)/2]
			w := p.TextWidth(label, 10, false) + 12
			box := r4(mid.X-w/2, mid.Y-11, mid.X+w/2, mid.Y+9)
			p.RoundRect(box, 4, p.Theme.AccentSoft, p.Theme.Accent)
			p.Text(box, label, p.Theme.AccentDown, 10, AlignCenter, true)
		}
	}
}

func (v *CanvasView) paintCardinality(p *Painter, x, y int32, text string, c types.TColor) {
	box := r4(x-8, y-9, x+8, y+9)
	p.RoundRect(box, 4, p.Theme.Surface, c)
	p.Text(box, text, c, 10, AlignCenter, true)
}

// ---- 命中测试 ----

func (v *CanvasView) tableAt(x, y int32) *Table {
	m := v.scene.model
	for i := len(m.Tables) - 1; i >= 0; i-- {
		t := &m.Tables[i]
		r := v.tableRect(t)
		// 端口在表边缘外 6px 内也算命中
		if x >= r.Left-6 && x <= r.Right+6 && y >= r.Top && y <= r.Bottom {
			return t
		}
	}
	return nil
}

func (v *CanvasView) columnIndexAt(t *Table, x, y int32) int {
	if t.Collapsed {
		return -1
	}
	r := v.tableRect(t)
	if y < r.Top+v.sz(38) {
		return -1
	}
	idx := int((y - r.Top - v.sz(38)) / maxInt32(v.sz(26), 1))
	if idx < 0 || idx >= len(t.Columns) {
		return -1
	}
	if x < r.Left-6 || x > r.Right+6 {
		return -1
	}
	return idx
}

func maxInt32(a, b int32) int32 {
	if a > b {
		return a
	}
	return b
}

func (v *CanvasView) portAt(x, y int32) (*Table, string) {
	m := v.scene.model
	for i := len(m.Tables) - 1; i >= 0; i-- {
		t := &m.Tables[i]
		if t.Collapsed {
			continue
		}
		r := v.tableRect(t)
		if x < r.Left-8 || x > r.Right+8 {
			continue
		}
		for j := range t.Columns {
			for _, side := range []int{-1, 1} {
				px, py := v.portPos(t, j, side)
				if abs32(x-px) <= 8 && abs32(y-py) <= 8 {
					return t, t.Columns[j].ID
				}
			}
		}
	}
	return nil, ""
}

func abs32(v int32) int32 {
	if v < 0 {
		return -v
	}
	return v
}

func (v *CanvasView) relationAt(x, y int32) string {
	for i := range v.scene.model.Relations {
		rel := &v.scene.model.Relations[i]
		pts := v.relationPoints(rel)
		for k := 0; k+1 < len(pts); k++ {
			if distToSegment(x, y, pts[k], pts[k+1]) < 7 {
				return rel.ID
			}
		}
	}
	return ""
}

func distToSegment(px, py int32, a, b types.TPoint) float64 {
	dx, dy := float64(b.X-a.X), float64(b.Y-a.Y)
	if dx == 0 && dy == 0 {
		return math.Hypot(float64(px-a.X), float64(py-a.Y))
	}
	t := ((float64(px-a.X)*dx + float64(py-a.Y)*dy) / (dx*dx + dy*dy))
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	return math.Hypot(float64(px)-(float64(a.X)+t*dx), float64(py)-(float64(a.Y)+t*dy))
}

func (v *CanvasView) linkStartPos() (int32, int32) {
	m := v.scene.model
	t := m.TableByID(v.linkTable)
	if t == nil {
		return v.startX, v.startY
	}
	idx := v.columnIndex(t, v.linkColumn)
	if idx < 0 {
		r := v.tableRect(t)
		return r.Left, r.Top + r.Height()/2
	}
	return v.portPos(t, idx, int(v.linkSide))
}

// ---- 视图操作 ----

func (v *CanvasView) ZoomAt(factor float64, cx, cy int32) {
	old := v.zoom
	next := old * factor
	if next < 0.45 {
		next = 0.45
	}
	if next > 1.8 {
		next = 1.8
	}
	if next == old {
		return
	}
	wx, wy := v.S2W(cx, cy)
	v.zoom = next
	v.panX = cx - v.rect.Left - int32(math.Round(wx*next))
	v.panY = cy - v.rect.Top - int32(math.Round(wy*next))
	v.scene.Invalidate()
}

func (v *CanvasView) ZoomBy(factor float64) {
	c := v.rect.CenterPoint()
	v.ZoomAt(factor, c.X, c.Y)
}

func (v *CanvasView) Fit() {
	minX, minY, maxX, maxY := v.scene.model.Bounds()
	bw, bh := maxFloat(maxX-minX, 1), maxFloat(maxY-minY, 1)
	availW := float64(v.rect.Width() - 96)
	availH := float64(v.rect.Height() - 96)
	s := math.Min(availW/bw, availH/bh)
	if s < 0.45 {
		s = 0.45
	}
	if s > 1.3 {
		s = 1.3
	}
	v.zoom = s
	v.panX = int32(float64(v.rect.Width())/2 - (minX+bw/2)*s)
	v.panY = int32(float64(v.rect.Height())/2 - (minY+bh/2)*s)
	v.scene.Invalidate()
}

func (v *CanvasView) CenterOn(t *Table) {
	if t == nil {
		return
	}
	cx := t.X + tableWidth/2
	cy := t.Y + TableHeight(t)/2
	v.panX = int32(float64(v.rect.Width())/2 - cx*v.zoom)
	v.panY = int32(float64(v.rect.Height())/2 - cy*v.zoom)
	v.scene.Invalidate()
}

// ---- 交互 ----

func (v *CanvasView) inView(x, y int32) bool {
	return x >= v.rect.Left && x < v.rect.Right && y >= v.rect.Top && y < v.rect.Bottom
}

func (v *CanvasView) Handle(e *Event) bool {
	switch e.Kind {
	case EvDown:
		return v.onDown(e)
	case EvMove:
		return v.onMove(e)
	case EvUp:
		return v.onUp(e)
	case EvWheel:
		if e.Shift.In(types.SsCtrl) {
			v.ZoomAt(1.08, e.X, e.Y)
		} else if e.Wheel > 0 {
			v.ZoomAt(1.10, e.X, e.Y)
		} else {
			v.ZoomAt(1/1.10, e.X, e.Y)
		}
		v.scene.SetStatus("缩放 " + percent(v.zoom))
		return true
	case EvLeave:
		v.hoverTable, v.hoverCol, v.hoverRel, v.hoverPort = "", "", "", false
		v.scene.Invalidate()
	}
	return true
}

func percent(z float64) string {
	return itoa(int(math.Round(z*100))) + "%"
}

func (v *CanvasView) onDown(e *Event) bool {
	if !v.inView(e.X, e.Y) {
		return false
	}
	v.startX, v.startY, v.lastX, v.lastY = e.X, e.Y, e.X, e.Y
	shift := e.Shift.In(types.SsShift)
	if e.Button == types.MbRight {
		v.drag = dragPan
		v.scene.OpenCanvasMenu(e.X, e.Y)
		return true
	}
	if e.Button == types.MbMiddle {
		v.drag = dragPan
		return true
	}
	if e.Button != types.MbLeft {
		return true
	}

	// 双击检测（自绘，不依赖原生 DblClick 事件，因为它不带坐标）
	now := time.Now()
	if now.Sub(v.clickAt) < 380*time.Millisecond &&
		abs32(e.X-v.clickX) < 5 && abs32(e.Y-v.clickY) < 5 {
		v.clickAt = time.Time{}
		v.drag = dragNone // 第一次按下残留的移动状态不能带进就地编辑
		v.onDblClick(e)
		return true
	}
	v.clickAt, v.clickX, v.clickY = now, e.X, e.Y

	if t, colID := v.portAt(e.X, e.Y); t != nil {
		v.drag = dragLink
		v.linkTable, v.linkColumn = t.ID, colID
		r := v.tableRect(t)
		if e.X-r.Left < r.Right-e.X {
			v.linkSide = -1
		} else {
			v.linkSide = 1
		}
		v.scene.SetStatus("拖到目标表的主键字段上释放，创建外键关系；Esc 取消")
		v.scene.Invalidate()
		return true
	}
	if t := v.tableAt(e.X, e.Y); t != nil {
		hdr := v.headerRect(t)
		plusX := hdr.Right - v.sz(25)
		arrowX := hdr.Right - v.sz(44)
		if e.Y >= hdr.Top && e.Y < hdr.Bottom {
			if e.X >= plusX {
				v.scene.AddColumn(t.ID)
				return true
			}
			if e.X >= arrowX && e.X < plusX {
				v.scene.ToggleCollapse(t.ID)
				return true
			}
			if shift {
				sel := v.scene.sel
				if sel.Tables == nil {
					sel.Tables = map[string]bool{}
				}
				if sel.Kind != "multi" {
					if sel.TableID != "" {
						sel.Tables[sel.TableID] = true
					}
				}
				sel.Tables[t.ID] = !sel.Tables[t.ID]
				sel.Kind = "multi"
				sel.TableID, sel.ColumnID, sel.RelationID = "", "", ""
				v.scene.sel = sel
				v.scene.SyncFromSelection()
			} else {
				v.scene.Select(Selection{Kind: "table", TableID: t.ID,
					Tables: map[string]bool{t.ID: true}})
			}
			v.drag, v.dragTableID = dragMove, t.ID
			v.lastMoveX, v.lastMoveY = e.X, e.Y
			v.moveDirty = false
			return true
		}
		if idx := v.columnIndexAt(t, e.X, e.Y); idx >= 0 {
			c := &t.Columns[idx]
			v.scene.Select(Selection{Kind: "column", TableID: t.ID, ColumnID: c.ID,
				Tables: map[string]bool{t.ID: true}})
			v.scene.SetStatus("字段 " + t.Name + "." + c.Name + " · " + c.TypeText())
			return true
		}
		return true
	}
	if rid := v.relationAt(e.X, e.Y); rid != "" {
		v.scene.Select(Selection{Kind: "relation", RelationID: rid})
		return true
	}
	if shift {
		v.drag = dragMarquee
		v.marquee = r4(e.X, e.Y, e.X, e.Y)
	} else {
		v.drag = dragPan
		v.scene.Select(Selection{Kind: "none", Tables: map[string]bool{}})
	}
	v.scene.Invalidate()
	return true
}

func (v *CanvasView) onMove(e *Event) bool {
	v.lastX, v.lastY = e.X, e.Y
	switch v.drag {
	case dragPan:
		v.panX += e.X - v.startX
		v.panY += e.Y - v.startY
		v.startX, v.startY = e.X, e.Y
		v.scene.Invalidate()
		return true
	case dragMove:
		dx := float64(e.X-v.lastMoveX) / v.zoom
		dy := float64(e.Y-v.lastMoveY) / v.zoom
		if dx != 0 || dy != 0 {
			if !v.moveDirty {
				v.scene.PushHistory() // 真正移动了才记录一次，点击选中不进撤销栈
				v.moveDirty = true
			}
			ids := v.scene.sel.Tables
			if len(ids) == 0 {
				ids = map[string]bool{v.dragTableID: true}
			}
			for id, on := range ids {
				if !on {
					continue
				}
				if t := v.scene.model.TableByID(id); t != nil {
					t.X += dx
					t.Y += dy
				}
			}
			v.lastMoveX, v.lastMoveY = e.X, e.Y
			v.scene.Invalidate()
		}
		return true
	case dragLink:
		v.scene.Invalidate()
		return true
	case dragMarquee:
		v.marquee = normRect(v.startX, v.startY, e.X, e.Y)
		v.scene.Invalidate()
		return true
	}

	// hover 反馈
	oldT, oldC, oldR, oldP := v.hoverTable, v.hoverCol, v.hoverRel, v.hoverPort
	v.hoverTable, v.hoverCol, v.hoverRel, v.hoverPort = "", "", "", false
	if t, col := v.portAt(e.X, e.Y); t != nil {
		v.hoverTable, v.hoverCol, v.hoverPort = t.ID, col, true
	} else if t := v.tableAt(e.X, e.Y); t != nil {
		v.hoverTable = t.ID
		if idx := v.columnIndexAt(t, e.X, e.Y); idx >= 0 {
			v.hoverCol = t.Columns[idx].ID
		}
	} else if rid := v.relationAt(e.X, e.Y); rid != "" {
		v.hoverRel = rid
	}
	if oldT != v.hoverTable || oldC != v.hoverCol || oldR != v.hoverRel || oldP != v.hoverPort {
		v.scene.Invalidate()
	}
	return true
}

func (v *CanvasView) onUp(e *Event) bool {
	mode := v.drag
	v.drag = dragNone
	switch mode {
	case dragLink:
		target, colID := v.portAt(e.X, e.Y)
		if target == nil {
			if t := v.tableAt(e.X, e.Y); t != nil {
				target = t
				colID = ""
			}
		}
		if target != nil {
			if colID == "" {
				// 落在表上但未指到字段：默认连到主键
				for i := range target.Columns {
					if target.Columns[i].PK {
						colID = target.Columns[i].ID
						break
					}
				}
				if colID == "" && len(target.Columns) > 0 {
					colID = target.Columns[0].ID
				}
			}
			v.scene.AddRelation(v.linkTable, v.linkColumn, target.ID, colID)
		} else {
			v.scene.SetStatus("已取消连线：请释放到目标表的字段或表头上")
		}
		v.scene.Invalidate()
	case dragMarquee:
		r := normRect(v.startX, v.startY, e.X, e.Y)
		sel := Selection{Kind: "none", Tables: map[string]bool{}}
		for i := range v.scene.model.Tables {
			t := &v.scene.model.Tables[i]
			tr := v.tableRect(t)
			if tr.Left < r.Right && tr.Right > r.Left && tr.Top < r.Bottom && tr.Bottom > r.Top {
				sel.Tables[t.ID] = true
			}
		}
		if len(sel.Tables) > 0 {
			sel.Kind = "multi"
			if len(sel.Tables) == 1 {
				sel.Kind = "table"
				for id := range sel.Tables {
					sel.TableID = id
				}
			}
			v.scene.SetStatus("框选了 " + itoa(len(sel.Tables)) + " 张表")
		}
		v.scene.Select(sel)
	case dragMove:
		if v.moveDirty {
			v.scene.SetStatus("已移动表位置（Ctrl+Z 可撤销）")
		}
	}
	return true
}

func normRect(x1, y1, x2, y2 int32) types.TRect {
	return r4(minI(x1, x2), minI(y1, y2), maxI(x1, x2), maxI(y1, y2))
}
func minI(a, b int32) int32 {
	if a < b {
		return a
	}
	return b
}
func maxI(a, b int32) int32 {
	if a > b {
		return a
	}
	return b
}

func (v *CanvasView) onDblClick(e *Event) bool {
	t := v.tableAt(e.X, e.Y)
	if t == nil {
		return true
	}
	hdr := v.headerRect(t)
	if e.Y >= hdr.Top && e.Y < hdr.Bottom && e.X < hdr.Right-v.sz(48) {
		v.scene.BeginInlineEdit("table", t.ID, "", hdr)
		return true
	}
	if idx := v.columnIndexAt(t, e.X, e.Y); idx >= 0 {
		v.scene.BeginInlineEdit("column", t.ID, t.Columns[idx].ID, v.rowRect(t, idx))
		return true
	}
	return true
}

// ---- 多选拖动用的上一帧鼠标位置 ----
