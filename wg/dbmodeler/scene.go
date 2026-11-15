package main

// Scene 是单画布宿主：Form 是唯一 HWND，ICustomGraphicControl 是唯一画布，
// 工具栏、左侧导航、右侧属性面板、状态栏、菜单、弹窗、就地编辑框全都是
// 画在这张画布上的纯 Go 对象。这里负责布局、事件分派与所有建模命令。

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
)

type Selection struct {
	Kind       string // none / table / column / relation / multi
	TableID    string
	ColumnID   string
	RelationID string
	Tables     map[string]bool
}

type inspItem struct {
	w    Widget
	h    int32
	gap  int32
	full bool // 是否占满整行（否则和下一个并列）
}

type Scene struct {
	form    *TMainForm
	host    lcl.ICustomGraphicControl
	painter *Painter
	theme   Theme

	model   *Model
	history *History
	view    *CanvasView
	sel     Selection
	status  string
	client  types.TRect
	frame   int64

	toolbarRect types.TRect
	navRect     types.TRect
	viewRect    types.TRect
	inspRect    types.TRect
	statusRect  types.TRect

	toolbar     []Widget
	navWidgets  []Widget
	inspItems   []inspItem
	inspWidgets []Widget
	inspScroll  int32
	inspContent int32

	buttons map[string]*Button
	search  *Input
	navList *List

	focus       Widget
	hover       Widget
	capture     Widget
	captureView bool
	popup       *PopupMenu
	modal       *Modal
	inline      *Input
	inlineKind  string
	inlineTable string
	inlineCol   string
	inlineOrig  string
	inlineDone  bool

	filter     string
	dataDir    string
	inspDrag   bool
	inspDragDY int32
}

func NewScene(form *TMainForm) *Scene {
	s := &Scene{
		form:    form,
		theme:   lightTheme(),
		model:   demoModel(),
		history: &History{},
		sel:     Selection{Kind: "none", Tables: map[string]bool{}},
		buttons: map[string]*Button{},
		status:  "就绪 · 已载入电商系统模拟数据（未连接数据库）",
		dataDir: defaultDataDir(),
	}
	s.painter = &Painter{Theme: &s.theme, Scale: 1}
	s.view = NewCanvasView(s)
	s.buildToolbar()
	s.buildNav()
	s.SyncInspector()
	if form != nil {
		s.createHost()
	}
	return s
}

func (s *Scene) createHost() {
	host := lcl.NewCustomGraphicControl(s.form)
	host.SetAlign(types.AlClient)
	host.SetParent(s.form)
	host.SetParentBackground(false)
	host.SetControlStyle(host.ControlStyle().Include(types.CsOpaque))
	s.host = host

	host.SetOnPaint(func(sender lcl.IObject) {
		c := host.Canvas()
		if c == nil || !c.IsValid() {
			return
		}
		s.painter.Canvas = c
		s.painter.Theme = &s.theme
		cr := host.ClientRect()
		s.Layout(cr.Width(), cr.Height())
		s.PaintAll(s.painter)
		s.frame++
		if winTest != nil {
			winTest.tick(s)
		}
		if smokeMode {
			if s.frame < 3 {
				host.Invalidate() // 烟测：确保连续 3 帧真实 OnPaint
			} else {
				lcl.RunOnMainThreadAsync(func(id uint32) {
					lcl.Application.Terminate()
				})
			}
		}
	})
	host.SetOnMouseDown(func(sender lcl.IObject, button types.TMouseButton,
		shift types.TShiftState, x, y int32) {
		s.Dispatch(&Event{Kind: EvDown, X: x, Y: y, Button: button, Shift: shift})
	})
	host.SetOnMouseMove(func(sender lcl.IObject, shift types.TShiftState, x, y int32) {
		s.Dispatch(&Event{Kind: EvMove, X: x, Y: y, Shift: shift})
	})
	host.SetOnMouseUp(func(sender lcl.IObject, button types.TMouseButton,
		shift types.TShiftState, x, y int32) {
		s.Dispatch(&Event{Kind: EvUp, X: x, Y: y, Button: button, Shift: shift})
	})
	host.SetOnMouseLeave(func(sender lcl.IObject) {
		s.Dispatch(&Event{Kind: EvLeave})
	})
	// 滚轮消息由 LCL 派发给光标下的控件；直接绑在画布上比自己解析
	// WM_MOUSEWHEEL 更可靠（坐标已经是控件客户区坐标）。
	host.SetOnMouseWheel(func(sender lcl.IObject, shift types.TShiftState,
		delta int32, pos types.TPoint, handled *bool) {
		s.Dispatch(&Event{Kind: EvWheel, X: pos.X, Y: pos.Y, Wheel: delta, Shift: shift})
		if handled != nil {
			*handled = true
		}
	})
}

func (s *Scene) Invalidate() {
	if s.host != nil {
		s.host.Invalidate()
	}
}

func (s *Scene) SetStatus(text string) {
	s.status = text
	s.Invalidate()
}

func (s *Scene) Model() *Model { return s.model }

// ---- 布局 ----

func (s *Scene) Layout(w, h int32) {
	if w < 1 || h < 1 {
		return
	}
	s.client = r4(0, 0, w, h)
	navW := int32(232)
	if w < 900 {
		navW = 180
	}
	inspW := int32(302)
	if w < 1100 {
		inspW = 250
	}
	s.toolbarRect = r4(0, 0, w, 54)
	s.statusRect = r4(0, h-28, w, h)
	s.navRect = r4(0, 54, navW, h-28)
	s.inspRect = r4(w-inspW, 54, w, h-28)
	s.viewRect = r4(navW, 54, w-inspW, h-28)
	s.view.SetRect(s.viewRect)

	// 工具栏：左侧标题区 + 等宽按钮
	start := int32(196)
	if w < 1000 {
		start = 150
	}
	count := int32(len(s.toolbar))
	gap := int32(5)
	bw := int32(76)
	if count > 0 && start+count*(bw+gap) > w-8 {
		bw = (w - start - 8 - count*gap) / count
		if bw < 44 {
			bw = 44
		}
	}
	for i, wgt := range s.toolbar {
		x := start + int32(i)*(bw+gap)
		wgt.SetRect(r4(x, 11, x+bw, 43))
	}

	// 左侧导航
	nav := s.navRect
	if len(s.navWidgets) >= 5 {
		s.navWidgets[0].SetRect(r4(nav.Left+12, nav.Top+8, nav.Right-12, nav.Top+30))
		s.navWidgets[1].SetRect(r4(nav.Left+12, nav.Top+34, nav.Right-12, nav.Top+62))
		s.navWidgets[2].SetRect(r4(nav.Left+12, nav.Top+66, nav.Right-12, nav.Top+86))
		s.navWidgets[3].SetRect(r4(nav.Left+6, nav.Top+90, nav.Right-6, nav.Bottom-52))
		s.navWidgets[4].SetRect(r4(nav.Left+12, nav.Bottom-44, nav.Right-12, nav.Bottom-12))
	}
	s.layoutInspector()
	if s.modal != nil {
		s.modal.Layout(s.client)
	}
}

func (s *Scene) buildToolbar() {
	type def struct {
		key, caption, hint string
		fn                 func()
	}
	defs := []def{
		{"newtable", "新建表", "Ctrl+N · 在画布中央新建一张表", s.NewTable},
		{"addcol", "加字段", "给选中表添加一个字段（也可点表头 +）", s.AddColumnCmd},
		{"delete", "删除", "Delete · 删除选中的表 / 字段 / 关系", s.DeleteSelection},
		{"undo", "撤销", "Ctrl+Z · 撤销上一步建模操作", s.Undo},
		{"redo", "重做", "Ctrl+Y · 重做", s.Redo},
		{"layout", "自动布局", "按外键关系分层排列所有表", s.AutoLayout},
		{"fit", "适应", "Ctrl+0 · 缩放并居中整个模型", s.Fit},
		{"zin", "放大", "+ / 滚轮向上", func() { s.view.ZoomBy(1.15) }},
		{"zout", "缩小", "- / 滚轮向下", func() { s.view.ZoomBy(1 / 1.15) }},
		{"ddl", "导出DDL", "Ctrl+E · 预览并导出建表 SQL（仅文本）", s.ExportDDL},
		{"save", "保存", "Ctrl+S · 保存为本地 model.json（模拟持久化）", s.Save},
		{"open", "打开", "Ctrl+O · 读取本地 model.json", s.Open},
		{"dialect", "MySQL", "切换 DDL 方言：MySQL / PostgreSQL / SQLite", s.ToggleDialect},
		{"theme", "主题", "切换明暗主题", s.ToggleTheme},
		{"help", "帮助", "F1 · 快捷键与操作说明", s.Help},
	}
	for _, d := range defs {
		b := NewButton("tb_"+d.key, d.caption, d.fn)
		b.Hint = d.hint
		b.OnHover = func(hint string) { s.SetStatus(hint) }
		s.buttons[d.key] = b
		s.toolbar = append(s.toolbar, b)
	}
	s.SyncToolbar()
}

func (s *Scene) buildNav() {
	title := NewLabel("nav_title", r4(0, 0, 10, 10), "表 / TABLES")
	title.Bold = true
	title.Height = 12
	s.search = NewInput("nav_search", r4(0, 0, 10, 10))
	s.search.SetPlaceholder("搜索表名 / 字段名…")
	s.search.OnChange = func(v string) {
		s.filter = strings.TrimSpace(v)
		s.RefreshNav()
	}
	count := NewLabel("nav_count", r4(0, 0, 10, 10), "")
	count.Muted = true
	count.Height = 10
	s.navList = NewList("nav_list")
	s.navList.OnSelect = func(idx int) {
		if idx < 0 || idx >= len(s.navList.Items) {
			return
		}
		id := s.navList.Items[idx].ID
		t := s.model.TableByID(id)
		if t == nil {
			return
		}
		s.Select(Selection{Kind: "table", TableID: id,
			Tables: map[string]bool{id: true}})
		s.view.CenterOn(t)
		s.SetStatus("定位到表 " + t.Name)
	}
	s.navList.OnHover = func(name string) { s.SetStatus("表 " + name + " · 单击定位，右键画布可查看更多操作") }
	newBtn := NewButton("nav_new", "＋ 新建表", s.NewTable)
	newBtn.Active = true
	newBtn.Hint = "Ctrl+N · 新建表"
	newBtn.OnHover = func(h string) { s.SetStatus(h) }
	s.navWidgets = []Widget{title, s.search, count, s.navList, newBtn}
	s.RefreshNav()
}

func (s *Scene) RefreshNav() {
	if s.navList == nil {
		return
	}
	filter := strings.ToLower(s.filter)
	items := make([]ListItem, 0, len(s.model.Tables))
	for i := range s.model.Tables {
		t := &s.model.Tables[i]
		if filter != "" {
			hit := strings.Contains(strings.ToLower(t.Name), filter) ||
				strings.Contains(strings.ToLower(t.Comment), filter)
			if !hit {
				for _, c := range t.Columns {
					if strings.Contains(strings.ToLower(c.Name), filter) {
						hit = true
						break
					}
				}
			}
			if !hit {
				continue
			}
		}
		items = append(items, ListItem{
			ID: t.ID, Title: t.Name, Sub: t.Comment, Group: t.Group,
			Count: len(t.Columns), Selected: s.sel.TableID == t.ID || s.sel.Tables[t.ID],
		})
	}
	s.navList.SetItems(items)
	if len(s.navWidgets) >= 3 {
		if lb, ok := s.navWidgets[2].(*Label); ok {
			lb.Text = fmt.Sprintf("共 %d 张表 · %d 个字段 · %d 条关系",
				len(s.model.Tables), s.countColumns(), len(s.model.Relations))
		}
	}
	s.Invalidate()
}

func (s *Scene) countColumns() int {
	n := 0
	for i := range s.model.Tables {
		n += len(s.model.Tables[i].Columns)
	}
	return n
}

// ---- 属性检查器 ----

func (s *Scene) addInspLabel(text string, muted, bold bool) {
	l := NewLabel("lbl", r4(0, 0, 10, 10), text)
	l.Muted, l.Bold = muted, bold
	l.Height = 11
	if bold {
		l.Height = 12
	}
	s.inspItems = append(s.inspItems, inspItem{w: l, h: 22, gap: 3})
}

func (s *Scene) addInspField(label string, value string, commit func(string)) {
	s.addInspLabel(label, true, false)
	in := NewInput("insp_"+label, r4(0, 0, 10, 10))
	in.SetText(value)
	in.OnSubmit = commit
	s.inspItems = append(s.inspItems, inspItem{w: in, h: 28, gap: 8})
}

func (s *Scene) addInspCombo(label string, items []string, value string, choose func(string)) {
	s.addInspLabel(label, true, false)
	c := NewCombo("combo_"+label, items)
	c.Value = value
	c.OnSelect = choose
	c.OnOpen = func(cb *Combo) { s.OpenComboPopup(cb) }
	s.inspItems = append(s.inspItems, inspItem{w: c, h: 28, gap: 8})
}

func (s *Scene) addInspChecks(items []struct {
	label   string
	checked bool
	toggle  func(bool)
}) {
	for _, it := range items {
		cb := NewCheckBox("chk_"+it.label, it.label)
		cb.Checked = it.checked
		cb.OnToggle = it.toggle
		s.inspItems = append(s.inspItems, inspItem{w: cb, h: 22, gap: 4})
	}
}

func (s *Scene) addInspButton(caption string, danger bool, fn func()) {
	b := NewButton("ib_"+caption, caption, fn)
	b.Danger = danger
	b.Hint = caption
	b.OnHover = func(h string) { s.SetStatus("属性面板 · " + h) }
	s.inspItems = append(s.inspItems, inspItem{w: b, h: 28, gap: 6})
}

func (s *Scene) addInspSeparator() {
	l := NewLabel("sep", r4(0, 0, 10, 10), "")
	s.inspItems = append(s.inspItems, inspItem{w: l, h: 8, gap: 2})
}

func (s *Scene) SyncInspector() {
	// 重建控件前先释放焦点，避免旧输入框失焦提交时写回已被替换的对象。
	if s.focus != nil && s.inspectorOwns(s.focus) {
		s.SetFocus(nil)
	}
	old := s.inspScroll
	s.inspItems = nil
	m := s.model

	switch s.sel.Kind {
	case "table":
		t := m.TableByID(s.sel.TableID)
		if t == nil {
			break
		}
		s.addInspLabel("表属性  ·  "+t.Name, false, true)
		s.addInspField("表名", t.Name, func(v string) { s.CommitName("table", t.ID, "", v) })
		s.addInspField("注释", t.Comment, func(v string) { s.CommitComment(t.ID, v) })
		s.addInspCombo("分组", []string{"核心域", "商品域", "交易域", "营销域"},
			groupLabel(t.Group), func(v string) { s.CommitGroup(t.ID, v) })
		s.addInspLabel("字段数 "+itoa(len(t.Columns))+" · 主键 "+s.pkNames(t), true, false)
		s.addInspSeparator()
		s.addInspButton("＋ 添加字段", false, func() { s.AddColumn(t.ID) })
		s.addInspButton("复制表", false, func() { s.DuplicateTable(t.ID) })
		s.addInspButton("删除表", true, func() { s.DeleteSelection() })
	case "column":
		t := m.TableByID(s.sel.TableID)
		if t == nil {
			break
		}
		c := t.ColumnByID(s.sel.ColumnID)
		if c == nil {
			break
		}
		s.addInspLabel("字段属性  ·  "+t.Name+"."+c.Name, false, true)
		s.addInspField("字段名", c.Name, func(v string) { s.CommitName("column", t.ID, c.ID, v) })
		s.addInspCombo("数据类型", DataTypes, c.Type, func(v string) { s.CommitColumn(c, "type", v) })
		s.addInspField("长度 / 精度", c.Length, func(v string) { s.CommitColumn(c, "length", v) })
		s.addInspField("默认值", c.Default, func(v string) { s.CommitColumn(c, "default", v) })
		s.addInspField("注释", c.Comment, func(v string) { s.CommitColumn(c, "comment", v) })
		type checkDef = struct {
			label   string
			checked bool
			toggle  func(bool)
		}
		checks := []checkDef{
			{"主键 PK", c.PK, func(v bool) { s.CommitColumn(c, "pk", boolText(v)) }},
			{"非空 NN", c.NN, func(v bool) { s.CommitColumn(c, "nn", boolText(v)) }},
			{"唯一 UQ", c.UQ, func(v bool) { s.CommitColumn(c, "uq", boolText(v)) }},
			{"自增 AI", c.AI, func(v bool) { s.CommitColumn(c, "ai", boolText(v)) }},
		}
		s.addInspChecks(checks)
		if m.IsForeignKey(t.ID, c.ID) {
			s.addInspLabel("该字段是外键（关系见画布连线）", true, false)
		}
		s.addInspSeparator()
		s.addInspButton("↑ 上移字段", false, func() { s.MoveColumn(t.ID, c.ID, -1) })
		s.addInspButton("↓ 下移字段", false, func() { s.MoveColumn(t.ID, c.ID, 1) })
		s.addInspButton("删除字段", true, func() { s.DeleteSelection() })
	case "relation":
		r := m.RelationByID(s.sel.RelationID)
		if r == nil {
			break
		}
		ft, tt := m.TableByID(r.FromTable), m.TableByID(r.ToTable)
		from, to := "?", "?"
		if ft != nil {
			from = ft.Name
			if fc := ft.ColumnByID(r.FromColumn); fc != nil {
				from += "." + fc.Name
			}
		}
		if tt != nil {
			to = tt.Name
			if tc := tt.ColumnByID(r.ToColumn); tc != nil {
				to += "." + tc.Name
			}
		}
		s.addInspLabel("外键关系", false, true)
		s.addInspField("约束名", r.Name, func(v string) { s.CommitRelation(r.ID, "name", v) })
		s.addInspLabel(from+"  →  "+to, true, false)
		s.addInspCombo("基数", []string{"1:N", "1:1", "N:M"}, r.Cardinality,
			func(v string) { s.CommitRelation(r.ID, "cardinality", v) })
		s.addInspCombo("ON DELETE", []string{"RESTRICT", "CASCADE", "SET NULL", "NO ACTION"},
			r.OnDelete, func(v string) { s.CommitRelation(r.ID, "onDelete", v) })
		s.addInspSeparator()
		s.addInspButton("删除关系", true, func() { s.DeleteSelection() })
	case "multi":
		s.addInspLabel("已选中 "+itoa(len(s.sel.Tables))+" 张表", false, true)
		s.addInspLabel("拖动任一表头可整体移动；Delete 批量删除", true, false)
		s.addInspSeparator()
		s.addInspButton("左对齐", false, func() { s.AlignTables(true, false) })
		s.addInspButton("顶部对齐", false, func() { s.AlignTables(false, true) })
		s.addInspButton("批量删除", true, func() { s.DeleteSelection() })
	default:
		s.addInspLabel("属性  ·  未选中对象", false, true)
		s.addInspLabel("在画布中单击表 / 字段 / 连线，或从左侧列表选择", true, false)
		s.addInspSeparator()
		s.addInspLabel("常用操作", false, true)
		s.addInspLabel("Ctrl+N 新建表    Delete 删除", true, false)
		s.addInspLabel("拖字段圆点 → 目标字段 = 建外键", true, false)
		s.addInspLabel("双击表名 / 字段名 = 就地改名", true, false)
		s.addInspLabel("Shift+拖拽 = 框选多张表", true, false)
		s.addInspLabel("滚轮缩放    空白拖拽平移", true, false)
		s.addInspLabel("Ctrl+Z/Y 撤销重做  Ctrl+S 保存", true, false)
		s.addInspLabel("Ctrl+E 导出 DDL   F1 帮助", true, false)
		s.addInspSeparator()
		s.addInspButton("校验模型", false, s.ValidateModel)
		s.addInspButton("查看 DDL", false, s.ExportDDL)
	}
	s.inspWidgets = nil
	for _, it := range s.inspItems {
		s.inspWidgets = append(s.inspWidgets, it.w)
	}
	s.inspScroll = old
	s.layoutInspector()
	s.SyncToolbar()
	s.Invalidate()
}

func boolText(v bool) string {
	if v {
		return "1"
	}
	return "0"
}

func (s *Scene) pkNames(t *Table) string {
	var names []string
	for _, c := range t.Columns {
		if c.PK {
			names = append(names, c.Name)
		}
	}
	if len(names) == 0 {
		return "（无）"
	}
	return strings.Join(names, ", ")
}

func (s *Scene) inspectorOwns(w Widget) bool {
	for _, x := range s.inspWidgets {
		if x == w {
			return true
		}
	}
	return false
}

func (s *Scene) layoutInspector() {
	r := s.inspRect
	x, right := r.Left+14, r.Right-14
	y := r.Top + 48 - s.inspScroll
	total := int32(0)
	for _, it := range s.inspItems {
		it.w.SetRect(r4(x, y, right, y+it.h))
		y += it.h + it.gap
		total += it.h + it.gap
	}
	s.inspContent = total
	s.clampInspScroll()
}

func (s *Scene) inspViewportH() int32 {
	h := s.inspRect.Height() - 48
	if h < 1 {
		h = 1
	}
	return h
}

func (s *Scene) inspMaxScroll() int32 {
	max := s.inspContent - s.inspViewportH()
	if max < 0 {
		max = 0
	}
	return max
}

func (s *Scene) inspBarTrack() types.TRect {
	return r4(s.inspRect.Right-9, s.inspRect.Top+46, s.inspRect.Right-3, s.inspRect.Bottom-6)
}

func (s *Scene) clampInspScroll() {
	if max := s.inspMaxScroll(); s.inspScroll > max {
		s.inspScroll = max
	}
	if s.inspScroll < 0 {
		s.inspScroll = 0
	}
}

func (s *Scene) SyncToolbar() {
	canDelete := s.sel.Kind != "none"
	s.buttons["undo"].SetEnabled(s.history.CanUndo())
	s.buttons["redo"].SetEnabled(s.history.CanRedo())
	s.buttons["addcol"].SetEnabled(s.sel.Kind == "table" || s.sel.Kind == "column")
	s.buttons["delete"].SetEnabled(canDelete)
	s.buttons["dialect"].Caption = s.model.Dialect
}

// ---- 绘制 ----

func (s *Scene) PaintAll(p *Painter) {
	p.Theme = &s.theme
	p.Reset()
	p.ClearClip()
	p.Scale = 1
	p.FillRect(s.client, s.theme.Bg)

	// 1) 中央 ER 画布先画；面板随后以不透明底色覆盖，等价于裁剪溢出内容。
	p.Scale = s.view.zoom
	s.view.Paint(p)
	p.Scale = 1
	p.Reset()

	// 2) 工具栏
	p.SetClip(s.toolbarRect)
	p.FillRect(s.toolbarRect, s.theme.Surface)
	p.FillRect(r4(0, s.toolbarRect.Bottom-1, s.client.Right, s.toolbarRect.Bottom), s.theme.Border)
	p.Text(r4(12, 5, 192, 27), "DB Modeler", s.theme.Text, 17, AlignLeft, true)
	p.Text(r4(12, 28, 192, 48), "单画布 · 纯自绘 ER 建模", s.theme.Muted, 10, AlignLeft, false)
	for _, w := range s.toolbar {
		w.Paint(p)
	}

	// 3) 左侧导航
	p.SetClip(s.navRect)
	p.FillRect(s.navRect, s.theme.Surface)
	p.FillRect(r4(s.navRect.Right-1, s.navRect.Top, s.navRect.Right, s.navRect.Bottom), s.theme.Border)
	for _, w := range s.navWidgets {
		w.Paint(p)
	}

	// 4) 右侧属性面板
	p.SetClip(s.inspRect)
	p.FillRect(s.inspRect, s.theme.Surface)
	p.FillRect(r4(s.inspRect.Left, s.inspRect.Top, s.inspRect.Left+1, s.inspRect.Bottom), s.theme.Border)
	p.Text(r4(s.inspRect.Left+14, s.inspRect.Top+10, s.inspRect.Right-14, s.inspRect.Top+34),
		"属性  /  PROPERTIES", s.theme.Text, 13, AlignLeft, true)
	p.FillRect(r4(s.inspRect.Left+14, s.inspRect.Top+38, s.inspRect.Right-14, s.inspRect.Top+39), s.theme.Border)
	for _, w := range s.inspWidgets {
		w.Paint(p)
	}
	// 属性面板内容可能超出高度（字段属性项最多），给一条可拖动的真实滚动条。
	if max := s.inspMaxScroll(); max > 0 {
		track := s.inspBarTrack()
		p.FillRect(track, s.theme.SurfaceAlt)
		thumb := scrollThumb(track, s.inspScroll, max)
		p.RoundRect(thumb, 3, s.theme.Border, s.theme.Border)
	}

	// 5) 状态栏
	p.SetClip(s.statusRect)
	p.FillRect(s.statusRect, s.theme.Surface)
	p.FillRect(r4(0, s.statusRect.Top, s.client.Right, s.statusRect.Top+1), s.theme.Border)
	rightText := fmt.Sprintf("%d 表 · %d 字段 · %d 关系   |   %s   |   缩放 %s   |   %s",
		len(s.model.Tables), s.countColumns(), len(s.model.Relations),
		s.model.Dialect, percent(s.view.zoom), "DEMO · 未连接数据库")
	rw := p.TextWidth(rightText, 11, false)
	p.Text(r4(s.client.Right-rw-16, s.statusRect.Top+2, s.client.Right-12, s.statusRect.Bottom),
		rightText, s.theme.Muted, 11, AlignRight, false)
	p.Text(r4(12, s.statusRect.Top+2, s.client.Right-rw-28, s.statusRect.Bottom),
		s.status, s.theme.Text, 11, AlignLeft, false)
	p.ClearClip()

	// 6) 就地编辑框 / 菜单 / 弹窗（最上层）
	if s.inline != nil && s.inline.Visible() {
		s.inline.Paint(p)
	}
	if s.popup != nil {
		s.popup.Paint(p)
	}
	if s.modal != nil {
		s.modal.Paint(p)
	}
}

// ---- 焦点 / IME ----

func (s *Scene) SetFocus(w Widget) {
	if w != nil && !w.Focusable() {
		w = nil
	}
	if s.focus == w {
		if s.form != nil {
			s.form.SetFocus()
			s.PositionIME()
		}
		return
	}
	if s.focus != nil {
		s.focus.OnFocusLost()
	}
	s.focus = w
	if w != nil {
		w.OnFocusGain()
	}
	if s.form != nil {
		s.form.SetFocus() // 系统键盘焦点始终在 Form 上
		s.PositionIME()
	}
	s.Invalidate()
}

func (s *Scene) Focused() Widget { return s.focus }

func (s *Scene) PositionIME() {
	if s.form == nil || s.host == nil {
		return
	}
	x, y := int32(40), int32(140)
	if in, ok := s.focus.(*Input); ok {
		x, y = in.CaretX, in.CaretY
	}
	pt := s.host.ClientToParent(types.TPoint{X: x, Y: y}, s.form)
	imePosition(uintptr(s.form.Handle()), pt.X, pt.Y)
}

func (s *Scene) SetComposition(text string, cursor int32) {
	if in, ok := s.focus.(*Input); ok {
		in.SetComposition(text, cursor)
	}
	s.Invalidate()
}

// ---- 事件分派 ----

func (s *Scene) inRect(r types.TRect, x, y int32) bool {
	return x >= r.Left && x < r.Right && y >= r.Top && y < r.Bottom
}

func (s *Scene) hitWidget(x, y int32) Widget {
	if s.modal != nil {
		return s.modal
	}
	if s.popup != nil && s.popup.Hit(x, y) {
		return s.popup
	}
	if s.inline != nil && s.inline.Visible() && s.inline.Hit(x, y) {
		return s.inline
	}
	if s.inRect(s.inspRect, x, y) {
		for i := len(s.inspWidgets) - 1; i >= 0; i-- {
			if s.inspWidgets[i].Hit(x, y) {
				return s.inspWidgets[i]
			}
		}
		return nil
	}
	if s.inRect(s.navRect, x, y) {
		for i := len(s.navWidgets) - 1; i >= 0; i-- {
			if s.navWidgets[i].Hit(x, y) {
				return s.navWidgets[i]
			}
		}
		return nil
	}
	if s.inRect(s.toolbarRect, x, y) {
		for i := len(s.toolbar) - 1; i >= 0; i-- {
			if s.toolbar[i].Hit(x, y) {
				return s.toolbar[i]
			}
		}
	}
	return nil
}

func (s *Scene) inspBarHit(x, y int32) bool {
	if s.inspMaxScroll() <= 0 {
		return false
	}
	track := s.inspBarTrack()
	return x >= track.Left-2 && x <= s.inspRect.Right && y >= track.Top && y <= track.Bottom
}

func (s *Scene) Dispatch(e *Event) {
	switch e.Kind {
	case EvDown:
		if s.modal != nil {
			s.modal.Handle(e)
			s.Invalidate()
			return
		}
		if s.popup != nil && !s.popup.Hit(e.X, e.Y) {
			s.ClosePopup()
			return
		}
		if s.inspBarHit(e.X, e.Y) {
			track := s.inspBarTrack()
			thumb := scrollThumb(track, s.inspScroll, s.inspMaxScroll())
			s.inspDrag = true
			if e.Y >= thumb.Top && e.Y < thumb.Bottom {
				s.inspDragDY = e.Y - thumb.Top
			} else {
				if e.Y < thumb.Top {
					s.inspScroll -= s.inspViewportH()
				} else {
					s.inspScroll += s.inspViewportH()
				}
				s.clampInspScroll()
				s.layoutInspector()
				s.inspDragDY = thumb.Height() / 2
			}
			s.SetFocus(nil)
			s.capture = nil
			s.Invalidate()
			return
		}
		w := s.hitWidget(e.X, e.Y)
		if s.inline != nil && s.inline.Visible() && w != s.inline {
			// 点击就地编辑框以外：提交并收起
			s.inline.Commit()
			s.finishInline()
			w = s.hitWidget(e.X, e.Y) // 编辑框已消失，重新命中下面的对象
		}
		if w == nil {
			if s.inRect(s.viewRect, e.X, e.Y) {
				s.captureView = true
				s.SetFocus(nil)
				s.view.Handle(e)
			} else {
				s.SetFocus(nil)
			}
		} else {
			s.capture = w
			if w.Focusable() {
				s.SetFocus(w)
			} else {
				s.SetFocus(nil)
			}
			w.Handle(e)
		}
	case EvMove:
		if s.inspDrag {
			track := s.inspBarTrack()
			thumb := scrollThumb(track, s.inspScroll, s.inspMaxScroll())
			s.inspScroll = scrollFromThumb(track, thumb.Height(), e.Y, s.inspDragDY, s.inspMaxScroll())
			s.clampInspScroll()
			s.layoutInspector()
			s.Invalidate()
			return
		}
		w := s.capture
		if w == nil {
			w = s.hitWidget(e.X, e.Y)
			if s.hover != w {
				if s.hover != nil {
					s.hover.Handle(&Event{Kind: EvLeave})
				}
				s.hover = w
				if s.hover == nil && !s.inRect(s.viewRect, e.X, e.Y) {
					s.SetStatus("就绪 · 纯自绘 ER 建模（模拟数据）")
				}
			}
		}
		if s.captureView {
			s.view.Handle(e)
		} else if w != nil {
			w.Handle(e)
		}
	case EvUp:
		if s.modal != nil {
			s.modal.Handle(e)
		} else if s.popup != nil {
			s.popup.Handle(e)
		} else if s.captureView {
			s.view.Handle(e)
		} else if s.capture != nil {
			s.capture.Handle(e)
		}
		s.capture = nil
		s.captureView = false
		s.inspDrag = false
	case EvWheel:
		s.onWheel(e)
	case EvLeave:
		s.inspDrag = false
		if s.hover != nil {
			s.hover.Handle(e)
			s.hover = nil
		}
		if s.captureView {
			s.view.Handle(e)
		}
	}
	s.Invalidate()
}

func (s *Scene) onWheel(e *Event) {
	if s.modal != nil {
		s.modal.Handle(e)
		return
	}
	if s.popup != nil {
		return
	}
	switch {
	case s.inRect(s.navRect, e.X, e.Y):
		s.navList.Handle(e)
	case s.inRect(s.inspRect, e.X, e.Y):
		s.inspScroll -= e.Wheel / 3
		s.clampInspScroll()
		s.layoutInspector()
	case s.inRect(s.viewRect, e.X, e.Y):
		s.view.Handle(e)
	}
}

func (s *Scene) OnText(text string) {
	if in, ok := s.focus.(*Input); ok {
		in.Handle(&Event{Kind: EvText, Text: text})
		s.PositionIME()
		s.Invalidate()
	}
}

func (s *Scene) OnKey(key uint16, shift types.TShiftState) {
	ctrl := shift.In(types.SsCtrl)
	upper := key
	if upper >= 'a' && upper <= 'z' {
		upper -= 'a' - 'A'
	}
	if key == vkEscape {
		if s.modal != nil {
			s.CloseModal()
		} else if s.popup != nil {
			s.ClosePopup()
		} else if s.inline != nil && s.inline.Visible() {
			s.inline.Handle(&Event{Kind: EvKey, Key: vkEscape})
			s.finishInline()
		} else if in, ok := s.focus.(*Input); ok {
			// Esc = 取消本次编辑：输入框自己还原文本并置 canceled，
			// 随后 SetFocus(nil) 触发的失焦提交会被跳过。
			in.Handle(&Event{Kind: EvKey, Key: vkEscape})
			s.SetFocus(nil)
		} else {
			s.view.drag = dragNone
			s.Select(Selection{Kind: "none", Tables: map[string]bool{}})
		}
		s.Invalidate()
		return
	}
	if s.modal != nil {
		s.modal.Handle(&Event{Kind: EvKey, Key: key, Shift: shift})
		s.Invalidate()
		return
	}
	if s.popup != nil {
		if key == vkReturn && s.popup.hover >= 0 {
			it := s.popup.Items[s.popup.hover]
			if it.Enabled && it.OnClick != nil {
				it.OnClick()
			}
			s.ClosePopup()
		} else if key == vkDown || key == vkUp {
			dir := 1
			if key == vkUp {
				dir = -1
			}
			for i := 0; i < len(s.popup.Items); i++ {
				s.popup.hover += dir
				if s.popup.hover < 0 {
					s.popup.hover = len(s.popup.Items) - 1
				}
				if s.popup.hover >= len(s.popup.Items) {
					s.popup.hover = 0
				}
				if !s.popup.Items[s.popup.hover].Sep {
					break
				}
			}
		}
		s.Invalidate()
		return
	}
	// 全局快捷键
	if ctrl {
		// 正在编辑文本时，Ctrl+Z/Y 属于输入框自己的撤销，不能撤销整个模型。
		if upper == 'Z' || upper == 'Y' {
			if in, ok := s.focus.(*Input); ok {
				in.Handle(&Event{Kind: EvKey, Key: key, Shift: shift})
				s.Invalidate()
				return
			}
		}
		switch upper {
		case 'N':
			s.NewTable()
			return
		case 'S':
			s.Save()
			return
		case 'O':
			s.Open()
			return
		case 'E':
			s.ExportDDL()
			return
		case 'F':
			s.SetFocus(s.search)
			return
		case 'Z':
			s.Undo()
			return
		case 'Y':
			s.Redo()
			return
		case '0':
			s.Fit()
			return
		}
		if key == 187 || key == 107 { // +/= 与数字键盘 +
			s.view.ZoomBy(1.15)
			return
		}
		if key == 189 || key == 109 { // - 与数字键盘 -
			s.view.ZoomBy(1 / 1.15)
			return
		}
	}
	if key == vkF1 {
		s.Help()
		return
	}
	if key == vkF2 {
		s.RenameSelected()
		return
	}
	if key == vkTab {
		s.focusNext()
		return
	}
	// 焦点在输入框时，编辑类按键交给输入框
	if in, ok := s.focus.(*Input); ok {
		editKey := key == vkBack || key == vkDelete || key == vkLeft ||
			key == vkRight || key == vkHome || key == vkEnd || key == vkReturn ||
			(ctrl && (upper == 'A' || upper == 'C' || upper == 'V' || upper == 'X'))
		if editKey {
			in.Handle(&Event{Kind: EvKey, Key: key, Shift: shift})
			s.Invalidate()
			return
		}
	}
	// + / - 缩放：文本输入时让位给字符输入，其它情况按帮助里写的行为生效。
	if _, editing := s.focus.(*Input); !editing {
		if key == 187 || key == 107 {
			s.view.ZoomBy(1.15)
			s.SetStatus("缩放 " + percent(s.view.zoom))
			return
		}
		if key == 189 || key == 109 {
			s.view.ZoomBy(1 / 1.15)
			s.SetStatus("缩放 " + percent(s.view.zoom))
			return
		}
	}
	if key == vkDelete {
		s.DeleteSelection()
		return
	}
	if key == vkLeft || key == vkRight || key == vkUp || key == vkDown {
		// 列表 / 下拉框获得焦点时，方向键属于控件自身（选行、展开下拉），
		// 不能拿去微调画布上的表位置。
		switch s.focus.(type) {
		case *List, *Combo:
			s.focus.Handle(&Event{Kind: EvKey, Key: key, Shift: shift})
			s.Invalidate()
			return
		}
		s.nudgeSelection(key, shift)
		return
	}
	if key == vkReturn || key == 32 {
		if s.focus != nil {
			s.focus.Handle(&Event{Kind: EvKey, Key: key, Shift: shift})
			s.Invalidate()
		}
		return
	}
	if s.focus != nil {
		s.focus.Handle(&Event{Kind: EvKey, Key: key, Shift: shift})
		s.Invalidate()
	}
}

func (s *Scene) focusNext() {
	var list []Widget
	list = append(list, s.toolbar...)
	list = append(list, s.navWidgets...)
	list = append(list, s.inspWidgets...)
	var focusables []Widget
	for _, w := range list {
		if w.Focusable() {
			focusables = append(focusables, w)
		}
	}
	if len(focusables) == 0 {
		return
	}
	next := focusables[0]
	for i, w := range focusables {
		if w == s.focus {
			next = focusables[(i+1)%len(focusables)]
			break
		}
	}
	s.SetFocus(next)
}

func (s *Scene) nudgeSelection(key uint16, shift types.TShiftState) {
	if len(s.sel.Tables) == 0 {
		return
	}
	step := 8.0
	if shift.In(types.SsShift) {
		step = 40
	}
	dx, dy := 0.0, 0.0
	switch key {
	case vkLeft:
		dx = -step
	case vkRight:
		dx = step
	case vkUp:
		dy = -step
	case vkDown:
		dy = step
	}
	s.PushHistory()
	for id, on := range s.sel.Tables {
		if !on {
			continue
		}
		if t := s.model.TableByID(id); t != nil {
			t.X += dx
			t.Y += dy
		}
	}
	s.Invalidate()
}

// ---- 选择 / 菜单 / 就地编辑 ----

func (s *Scene) Select(sel Selection) {
	if sel.Tables == nil {
		sel.Tables = map[string]bool{}
	}
	s.sel = sel
	s.SyncFromSelection()
}

func (s *Scene) SyncFromSelection() {
	s.RefreshNav()
	s.SyncInspector()
	s.Invalidate()
}

func (s *Scene) BeginInlineEdit(kind, tableID, colID string, r types.TRect) {
	if s.inline != nil {
		s.inline.SetVisible(false)
	}
	in := NewInput("inline", insetR(r, 2))
	s.inlineOrig = ""
	switch kind {
	case "table":
		if t := s.model.TableByID(tableID); t != nil {
			in.SetText(t.Name)
			s.inlineOrig = t.Name
		}
	default:
		if t := s.model.TableByID(tableID); t != nil {
			if c := t.ColumnByID(colID); c != nil {
				in.SetText(c.Name)
				s.inlineOrig = c.Name
			}
		}
	}
	in.SelectAll()
	in.OnSubmit = func(v string) {
		if s.inlineDone {
			return // 防止 SetFocus(nil) 触发 OnFocusLost 时重复提交
		}
		s.inlineDone = true
		s.CommitName(kind, tableID, colID, v)
		s.finishInline()
		s.inlineDone = false
	}
	in.OnCancel = func() { in.SetText(s.inlineOrig) }
	s.inline = in
	s.inlineKind, s.inlineTable, s.inlineCol = kind, tableID, colID
	s.SetFocus(in)
}

// finishInline 收起就地编辑框。提交与取消两条路径共用。
func (s *Scene) finishInline() {
	if s.inline == nil {
		return
	}
	in := s.inline
	in.SetVisible(false)
	s.inline = nil
	if s.focus == in {
		s.SetFocus(nil)
	}
}

func (s *Scene) RenameSelected() {
	if s.sel.Kind == "table" || s.sel.Kind == "column" {
		t := s.model.TableByID(s.sel.TableID)
		if t == nil {
			return
		}
		// 聚焦属性面板的名称输入框（同样可编辑，且支持 IME）
		for _, w := range s.inspWidgets {
			if in, ok := w.(*Input); ok {
				in.SelectAll()
				s.SetFocus(in)
				s.SetStatus("F2 · 正在编辑名称，Enter 提交 / Esc 取消")
				return
			}
		}
	}
}

func (s *Scene) ClosePopup() {
	if s.popup == nil {
		return
	}
	s.popup = nil
	s.capture = nil
	s.captureView = false
	s.view.drag = dragNone
	s.Invalidate()
}

func (s *Scene) showPopup(title string, items []MenuItem, x, y int32) {
	s.ClosePopup()
	m := NewPopupMenu(title, items)
	m.PlaceAt(x, y, s.client.Right, s.client.Bottom)
	m.OnHover = func(h string) {
		if h != "" {
			s.SetStatus(h)
		}
	}
	m.OnClose = s.ClosePopup
	s.popup = m
	s.Invalidate()
}

func (s *Scene) OpenComboPopup(c *Combo) {
	items := make([]MenuItem, 0, len(c.Items))
	for _, text := range c.Items {
		value := text
		items = append(items, MenuItem{
			Text: value, Enabled: true, Checked: value == c.Value,
			OnClick: func() {
				c.Value = value
				if c.OnSelect != nil {
					c.OnSelect(value)
				}
			},
		})
	}
	s.showPopup("选择 "+c.Name(), items, c.rect.Left, c.rect.Bottom)
	// 用键盘打开下拉时，默认高亮当前值，Enter 即可确认。
	if s.popup != nil {
		for i, it := range s.popup.Items {
			if it.Checked {
				s.popup.hover = i
				break
			}
		}
	}
}

func (s *Scene) OpenCanvasMenu(x, y int32) {
	t, colID := s.view.portAt(x, y)
	if t == nil {
		t = s.view.tableAt(x, y)
	}
	if t != nil {
		if colID == "" {
			if idx := s.view.columnIndexAt(t, x, y); idx >= 0 {
				colID = t.Columns[idx].ID
			}
		}
		// TableBy*/tableAt 返回切片内元素指针，菜单闭包可能延迟执行，
		// 因此一律保存 ID，执行时再查表，避免切片搬移后指针失效。
		tid, cid := t.ID, colID
		tableName := t.Name
		s.Select(Selection{Kind: "table", TableID: tid,
			Tables: map[string]bool{tid: true}})
		label := "展开"
		if !t.Collapsed {
			label = "折叠"
		}
		items := []MenuItem{
			{Text: "重命名表", Hint: "双击表头也可就地改名", Enabled: true, Shortcut: "F2",
				OnClick: func() {
					if cur := s.model.TableByID(tid); cur != nil {
						s.BeginInlineEdit("table", tid, "", s.view.headerRect(cur))
					}
				}},
			{Text: "新增字段", Hint: "在表末尾追加一个字段", Enabled: true,
				OnClick: func() { s.AddColumn(tid) }},
			{Text: label, Hint: "折叠 / 展开字段列表", Enabled: true,
				OnClick: func() { s.ToggleCollapse(tid) }},
			{Text: "复制表", Hint: "复制表结构与字段", Enabled: true,
				OnClick: func() { s.DuplicateTable(tid) }},
			miSep(),
			{Text: "删除表", Hint: "同时删除与它相连的关系", Enabled: true, Danger: true,
				Shortcut: "Del", OnClick: s.DeleteSelection},
		}
		if cid != "" {
			items = append(items, miSep(),
				MenuItem{Text: "重命名选中字段", Enabled: true,
					OnClick: func() {
						cur := s.model.TableByID(tid)
						if cur == nil {
							return
						}
						idx := s.view.columnIndex(cur, cid)
						if idx >= 0 {
							s.BeginInlineEdit("column", tid, cid, s.view.rowRect(cur, idx))
						}
					}},
				MenuItem{Text: "删除选中字段", Enabled: true, Danger: true,
					OnClick: func() {
						s.Select(Selection{Kind: "column", TableID: tid, ColumnID: cid,
							Tables: map[string]bool{tid: true}})
						s.DeleteSelection()
					}})
		}
		s.showPopup("表 · "+tableName, items, x, y)
		return
	}
	if rid := s.view.relationAt(x, y); rid != "" {
		r := s.model.RelationByID(rid)
		if r != nil {
			s.Select(Selection{Kind: "relation", RelationID: rid})
			items := []MenuItem{
				{Text: "基数 1:N", Enabled: true, Checked: r.Cardinality == "1:N",
					OnClick: func() { s.CommitRelation(r.ID, "cardinality", "1:N") }},
				{Text: "基数 1:1", Enabled: true, Checked: r.Cardinality == "1:1",
					OnClick: func() { s.CommitRelation(r.ID, "cardinality", "1:1") }},
				{Text: "ON DELETE CASCADE", Enabled: true, Checked: r.OnDelete == "CASCADE",
					OnClick: func() { s.CommitRelation(r.ID, "onDelete", "CASCADE") }},
				miSep(),
				{Text: "删除关系", Enabled: true, Danger: true, Shortcut: "Del",
					OnClick: s.DeleteSelection},
			}
			s.showPopup("关系 · "+r.Name, items, x, y)
			return
		}
	}
	items := []MenuItem{
		{Text: "新建表", Enabled: true, Shortcut: "Ctrl+N", OnClick: s.NewTable},
		{Text: "自动布局", Enabled: true, OnClick: s.AutoLayout},
		{Text: "适应窗口", Enabled: true, Shortcut: "Ctrl+0", OnClick: s.Fit},
		miSep(),
		{Text: "导出 DDL", Enabled: true, Shortcut: "Ctrl+E", OnClick: s.ExportDDL},
		{Text: "保存模型", Enabled: true, Shortcut: "Ctrl+S", OnClick: s.Save},
		{Text: "打开模型", Enabled: true, Shortcut: "Ctrl+O", OnClick: s.Open},
		{Text: "校验模型", Enabled: true, OnClick: s.ValidateModel},
		miSep(),
		{Text: "帮助 / 快捷键", Enabled: true, Shortcut: "F1", OnClick: s.Help},
	}
	s.showPopup("画布", items, x, y)
}

func (s *Scene) CloseModal() {
	if s.modal == nil {
		return
	}
	s.modal = nil
	s.Invalidate()
}

func (s *Scene) showModal(title, subtitle string, lines []string, buttons []ModalButton) {
	s.ClosePopup()
	m := NewModal(title, lines)
	m.Subtitle = subtitle
	m.Buttons = buttons
	m.OnClose = s.CloseModal
	m.Layout(s.client)
	s.modal = m
	s.SetFocus(nil)
	s.Invalidate()
}

// ---- 建模命令 ----

func (s *Scene) PushHistory() {
	s.history.Push(s.model)
	s.SyncToolbar()
}

func (s *Scene) afterChange(msg string) {
	s.RefreshNav()
	s.SyncInspector()
	s.SetStatus(msg)
}

// freeTableSpot 在**当前可见区域**里找一个不和已有表重叠的落点。
// 旧实现从中心一路 +40/+36 避让，表多时会走到视口外，新建的表直接看不见。
func (s *Scene) freeTableSpot() (float64, float64) {
	newH := 38.0 + 3*26 + 8 // 新表固定 3 个字段
	x0, y0 := s.view.S2W(s.viewRect.Left+18, s.viewRect.Top+18)
	x1, y1 := s.view.S2W(s.viewRect.Right-18, s.viewRect.Bottom-18)
	x1 -= tableWidth
	y1 -= newH
	if x1 < x0 {
		x1 = x0
	}
	if y1 < y0 {
		y1 = y0
	}
	stepX, stepY := tableWidth+24, newH+24
	// 逐个格子算重叠面积：找到 0 就直接用；一格都没有时退回「压得最轻」的那格，
	// 而不是盲退到视口中心（那样新表会整块压在已有表上，看着像渲染错位）。
	bestX, bestY, bestOverlap := (x0+x1)/2, (y0+y1)/2, math.MaxFloat64
	for y := y0; y <= y1; y += stepY {
		for x := x0; x <= x1; x += stepX {
			var overlap float64
			for i := range s.model.Tables {
				t := &s.model.Tables[i]
				overlap += overlapArea(x, y, tableWidth, newH, t.X, t.Y, tableWidth, TableHeight(t))
			}
			if overlap <= 0 {
				return x, y
			}
			if overlap < bestOverlap {
				bestOverlap, bestX, bestY = overlap, x, y
			}
		}
	}
	return bestX, bestY
}

// overlapArea 返回两个矩形的相交面积（不相交为 0）。
func overlapArea(ax, ay, aw, ah, bx, by, bw, bh float64) float64 {
	ox := math.Min(ax+aw, bx+bw) - math.Max(ax, bx)
	oy := math.Min(ay+ah, by+bh) - math.Max(ay, by)
	if ox <= 0 || oy <= 0 {
		return 0
	}
	return ox * oy
}

func (s *Scene) NewTable() {
	s.PushHistory()
	wx, wy := s.freeTableSpot()
	t := s.model.AddTable(wx, wy)
	s.Select(Selection{Kind: "table", TableID: t.ID,
		Tables: map[string]bool{t.ID: true}})
	s.afterChange("已新建表 " + t.Name + "（双击表头改名）")
}

func absF(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

func (s *Scene) AddColumnCmd() {
	id := s.sel.TableID
	if id == "" {
		s.SetStatus("请先选中一张表，再添加字段")
		return
	}
	s.AddColumn(id)
}

func (s *Scene) AddColumn(tableID string) {
	if t := s.model.TableByID(tableID); t == nil {
		s.SetStatus("添加字段失败：表已不存在")
		return
	} else if t.Collapsed {
		t.Collapsed = false // 折叠状态下加字段会看不见，先展开
	}
	s.PushHistory()
	c := s.model.AddColumn(tableID)
	if c == nil {
		s.SetStatus("添加字段失败：表已不存在")
		return
	}
	s.Select(Selection{Kind: "column", TableID: tableID, ColumnID: c.ID,
		Tables: map[string]bool{tableID: true}})
	s.afterChange("已添加字段 " + c.Name)
}

func (s *Scene) DeleteSelection() {
	switch s.sel.Kind {
	case "none":
		s.SetStatus("没有选中的对象")
		return
	case "multi":
		s.PushHistory()
		n := 0
		for id, on := range s.sel.Tables {
			if on {
				s.model.DeleteTable(id)
				n++
			}
		}
		s.Select(Selection{Kind: "none", Tables: map[string]bool{}})
		s.afterChange("已批量删除 " + itoa(n) + " 张表及其关联关系")
		return
	}
	s.PushHistory()
	switch s.sel.Kind {
	case "table":
		t := s.model.TableByID(s.sel.TableID)
		name := "表"
		if t != nil {
			name = t.Name
		}
		s.model.DeleteTable(s.sel.TableID)
		s.Select(Selection{Kind: "none", Tables: map[string]bool{}})
		s.afterChange("已删除表 " + name + " 及其关联关系")
	case "column":
		t := s.model.TableByID(s.sel.TableID)
		name := "字段"
		if t != nil {
			if c := t.ColumnByID(s.sel.ColumnID); c != nil {
				name = t.Name + "." + c.Name
			}
		}
		s.model.DeleteColumn(s.sel.TableID, s.sel.ColumnID)
		s.Select(Selection{Kind: "table", TableID: s.sel.TableID,
			Tables: map[string]bool{s.sel.TableID: true}})
		s.afterChange("已删除字段 " + name)
	case "relation":
		name := "关系"
		if r := s.model.RelationByID(s.sel.RelationID); r != nil {
			name = r.Name
		}
		s.model.DeleteRelation(s.sel.RelationID)
		s.Select(Selection{Kind: "none", Tables: map[string]bool{}})
		s.afterChange("已删除关系 " + name)
	}
}

func (s *Scene) ToggleCollapse(tableID string) {
	t := s.model.TableByID(tableID)
	if t == nil {
		return
	}
	// 折叠状态存在模型里、也会写进 model.json，因此同样进入撤销栈。
	s.PushHistory()
	t.Collapsed = !t.Collapsed
	s.SetStatus("已" + map[bool]string{true: "折叠", false: "展开"}[t.Collapsed] + " " + t.Name)
	s.Invalidate()
}

func (s *Scene) DuplicateTable(id string) {
	src := s.model.TableByID(id)
	if src == nil {
		return
	}
	s.PushHistory()
	cp := *src
	cp.ID = s.model.nextID("t")
	// 连续复制同一张表时必须生成不冲突的名字，否则模型里会出现两张同名表。
	cp.Name = s.model.uniqueTableName(src.Name + "_copy")
	cp.X += 40
	cp.Y += 40
	cp.Columns = append([]Column(nil), src.Columns...)
	for i := range cp.Columns {
		cp.Columns[i].ID = s.model.nextID("c")
	}
	s.model.Tables = append(s.model.Tables, cp)
	s.Select(Selection{Kind: "table", TableID: cp.ID,
		Tables: map[string]bool{cp.ID: true}})
	s.afterChange("已复制表结构：" + cp.Name)
}

func (s *Scene) MoveColumn(tableID, colID string, delta int) {
	t := s.model.TableByID(tableID)
	if t == nil {
		return
	}
	idx := -1
	for i := range t.Columns {
		if t.Columns[i].ID == colID {
			idx = i
			break
		}
	}
	if idx < 0 || idx+delta < 0 || idx+delta >= len(t.Columns) {
		s.SetStatus("字段已在列表边界，无法继续移动")
		return
	}
	s.PushHistory()
	s.model.MoveColumn(tableID, colID, delta)
	s.SyncInspector()
	s.SetStatus("已调整字段顺序（Ctrl+Z 可撤销）")
}

func (s *Scene) AddRelation(fromTable, fromCol, toTable, toCol string) {
	if toCol == "" {
		s.SetStatus("目标表没有可用字段")
		return
	}
	s.PushHistory()
	rel, err := s.model.AddRelation(fromTable, fromCol, toTable, toCol)
	if err != nil {
		// 失败时直接丢弃刚压入的快照，避免污染撤销 / 重做栈
		s.history.DropLast()
		s.SyncToolbar()
		s.SetStatus("创建关系失败：" + err.Error())
		return
	}
	s.Select(Selection{Kind: "relation", RelationID: rel.ID})
	s.afterChange("已创建外键关系 " + rel.Name + "（1:N，可在属性面板修改基数）")
}

func (s *Scene) CommitName(kind, tableID, colID, value string) {
	value = strings.TrimSpace(sanitizeText(value))
	if value == "" {
		s.SetStatus("名称不能为空，已保留原名")
		return
	}
	if kind == "table" {
		t := s.model.TableByID(tableID)
		if t == nil || t.Name == value {
			return
		}
		if other := s.model.TableByName(value); other != nil && other.ID != tableID {
			s.SetStatus("表名重复：" + value + "（已保留原名）")
			s.RefreshNav()
			return
		}
		s.PushHistory()
		old := t.Name
		t.Name = value
		s.afterChange("表 " + old + " 已改名为 " + value)
		return
	}
	t := s.model.TableByID(tableID)
	if t == nil {
		return
	}
	c := t.ColumnByID(colID)
	if c == nil || c.Name == value {
		return
	}
	if other := t.ColumnByName(value); other != nil && other.ID != colID {
		s.SetStatus("字段名重复：" + t.Name + "." + value + "（已保留原名）")
		return
	}
	s.PushHistory()
	c.Name = value
	s.afterChange("字段已改名为 " + value)
}

func (s *Scene) CommitComment(tableID, value string) {
	t := s.model.TableByID(tableID)
	if t == nil || t.Comment == value {
		return
	}
	s.PushHistory()
	t.Comment = value
	s.afterChange("已更新表注释")
}

func (s *Scene) CommitGroup(tableID, label string) {
	t := s.model.TableByID(tableID)
	if t == nil {
		return
	}
	key := "core"
	for _, g := range groupKeys {
		if groupLabel(g) == label {
			key = g
			break
		}
	}
	if t.Group == key {
		return
	}
	s.PushHistory()
	t.Group = key
	s.afterChange("已把 " + t.Name + " 归入 " + label)
}

func (s *Scene) CommitColumn(c *Column, field, value string) {
	// c 是模型切片里的指针；SyncInspector 重建控件后旧指针可能失效，
	// 因此用 ID 重新查找，保证写回的一定是当前模型对象。
	var cur *Column
	for i := range s.model.Tables {
		t := &s.model.Tables[i]
		if col := t.ColumnByID(c.ID); col != nil {
			cur = col
			break
		}
	}
	if cur == nil {
		return
	}
	var apply func()
	switch field {
	case "type":
		if cur.Type != value {
			apply = func() { cur.Type = value; cur.Length = defaultLength(value) }
		}
	case "length":
		if cur.Length != value {
			apply = func() { cur.Length = value }
		}
	case "default":
		if cur.Default != value {
			apply = func() { cur.Default = value }
		}
	case "comment":
		if cur.Comment != value {
			apply = func() { cur.Comment = value }
		}
	case "pk", "nn", "uq", "ai":
		v := value == "1"
		var old *bool
		switch field {
		case "pk":
			old = &cur.PK
		case "nn":
			old = &cur.NN
		case "uq":
			old = &cur.UQ
		case "ai":
			old = &cur.AI
		}
		if *old != v {
			apply = func() {
				*old = v
				if field == "pk" && v {
					cur.NN = true
				}
			}
		}
	}
	if apply == nil {
		return
	}
	s.PushHistory() // 快照必须压在修改之前
	apply()
	// 勾选框 / 下拉类变更不重建控件（避免焦点抖动），只刷新显示。
	s.RefreshNav()
	s.SyncToolbar()
	s.SetStatus("已更新字段 " + cur.Name + " 的 " + field)
	s.Invalidate()
}

func (s *Scene) CommitRelation(id, field, value string) {
	r := s.model.RelationByID(id)
	if r == nil {
		return
	}
	var apply func()
	switch field {
	case "name":
		value = strings.TrimSpace(value)
		if value != "" && r.Name != value {
			apply = func() { r.Name = value }
		}
	case "cardinality":
		if r.Cardinality != value {
			apply = func() { r.Cardinality = value }
		}
	case "onDelete":
		if r.OnDelete != value {
			apply = func() { r.OnDelete = value }
		}
	}
	if apply == nil {
		return
	}
	s.PushHistory()
	apply()
	s.SyncInspector()
	s.SetStatus("已更新关系 " + r.Name)
}

func (s *Scene) AlignTables(alignLeft, alignTop bool) {
	if len(s.sel.Tables) == 0 {
		return
	}
	minX, minY := 1e9, 1e9
	for id, on := range s.sel.Tables {
		if !on {
			continue
		}
		if t := s.model.TableByID(id); t != nil {
			minX = minFloat(minX, t.X)
			minY = minFloat(minY, t.Y)
		}
	}
	// 已经对齐时不要压空快照，否则撤销栈里会出现「撤销了但界面没变」。
	changed := false
	for id, on := range s.sel.Tables {
		if !on {
			continue
		}
		if t := s.model.TableByID(id); t != nil {
			if alignLeft && t.X != minX {
				changed = true
			}
			if alignTop && t.Y != minY {
				changed = true
			}
		}
	}
	if !changed {
		s.SetStatus("选中的表已经对齐")
		return
	}
	s.PushHistory()
	for id, on := range s.sel.Tables {
		if !on {
			continue
		}
		if t := s.model.TableByID(id); t != nil {
			if alignLeft {
				t.X = minX
			}
			if alignTop {
				t.Y = minY
			}
		}
	}
	s.SetStatus("已对齐选中的表")
	s.Invalidate()
}

func (s *Scene) Undo() {
	if !s.history.Undo(s.model) {
		s.SetStatus("没有可撤销的操作")
		return
	}
	s.sel = Selection{Kind: "none", Tables: map[string]bool{}}
	s.afterChange("已撤销（Ctrl+Y 可重做）")
}

func (s *Scene) Redo() {
	if !s.history.Redo(s.model) {
		s.SetStatus("没有可重做的操作")
		return
	}
	s.sel = Selection{Kind: "none", Tables: map[string]bool{}}
	s.afterChange("已重做")
}

func (s *Scene) AutoLayout() {
	s.PushHistory()
	s.model.AutoLayout()
	s.view.Fit()
	s.SetStatus("已按外键关系分层自动布局（Ctrl+Z 可撤销）")
	s.Invalidate()
}

func (s *Scene) Fit() {
	s.view.Fit()
	s.SetStatus("已适应窗口 · 缩放 " + percent(s.view.zoom))
}

func (s *Scene) ToggleDialect() {
	for i, d := range Dialects {
		if d == s.model.Dialect {
			s.model.Dialect = Dialects[(i+1)%len(Dialects)]
			break
		}
	}
	s.SyncToolbar()
	s.SetStatus("DDL 方言切换为 " + s.model.Dialect + "（仅影响生成文本，不连接数据库）")
	s.Invalidate()
}

func (s *Scene) ToggleTheme() {
	if s.theme.Dark {
		s.theme = lightTheme()
	} else {
		s.theme = darkTheme()
	}
	s.painter.Theme = &s.theme
	s.SetStatus("已切换为" + map[bool]string{true: "深色", false: "浅色"}[s.theme.Dark] + "主题")
	s.Invalidate()
}

// ---- 文件 / DDL / 帮助（全部是本地文本，不连接数据库）----

func defaultDataDir() string {
	if exe, err := os.Executable(); err == nil {
		return filepath.Dir(exe)
	}
	return "."
}

// dataPath 把文件名解析到当前数据目录。目录可注入，测试用临时目录，
// 正常运行时用可执行文件所在目录。
func (s *Scene) dataPath(name string) string {
	dir := s.dataDir
	if dir == "" {
		dir = defaultDataDir()
	}
	return filepath.Join(dir, name)
}

func (s *Scene) Save() {
	path := s.dataPath("model.json")
	if err := os.WriteFile(path, []byte(s.model.JSON()), 0644); err != nil {
		s.SetStatus("保存失败：" + err.Error())
		return
	}
	s.SetStatus("已保存到 " + path + "（" + itoa(len(s.model.Tables)) + " 表 / " +
		itoa(s.countColumns()) + " 字段 / " + itoa(len(s.model.Relations)) + " 关系）")
}

func (s *Scene) Open() {
	path := s.dataPath("model.json")
	data, err := os.ReadFile(path)
	if err != nil {
		s.SetStatus("打开失败：" + err.Error() + "；可先 Ctrl+S 保存一份")
		return
	}
	m, err := ModelFromJSON(data)
	if err != nil {
		s.SetStatus("模型文件解析失败：" + err.Error())
		return
	}
	s.PushHistory()
	s.model.Restore(m)
	s.sel = Selection{Kind: "none", Tables: map[string]bool{}}
	s.afterChange("已载入 " + path + "（真实数据库连接：无）")
	s.view.Fit()
}

func (s *Scene) ExportDDL() {
	ddl := s.model.DDL()
	path := s.dataPath("schema.sql")
	subtitle := "共 %d 行 · 纯文本生成，未执行任何 SQL"
	if err := os.WriteFile(path, []byte(ddl), 0644); err != nil {
		s.SetStatus("导出失败：" + err.Error())
		subtitle = "共 %d 行 · 写入失败：" + err.Error() + " · 纯文本生成，未执行任何 SQL"
	} else {
		s.SetStatus("DDL 已导出到 " + path)
		subtitle = "共 %d 行 · 已写入 " + path + " · 纯文本生成，未执行任何 SQL"
	}
	lines := strings.Split(strings.ReplaceAll(ddl, "\r\n", "\n"), "\n")
	buttons := []ModalButton{
		{Text: "复制 SQL", OnClick: func() {
			if setClipboardText(ddl) {
				s.SetStatus("SQL 已复制到剪贴板")
			} else {
				s.SetStatus("剪贴板不可用")
			}
		}},
		{Text: "关闭", Primary: true, OnClick: s.CloseModal},
	}
	s.showModal("DDL 预览 · "+s.model.Dialect,
		fmt.Sprintf(subtitle, len(lines)), lines, buttons)
}

func (s *Scene) ValidateModel() {
	issues := s.model.Validate()
	lines := []string{
		"模型：" + s.model.Name,
		"方言：" + s.model.Dialect,
		"表 / 字段 / 关系：" + itoa(len(s.model.Tables)) + " / " +
			itoa(s.countColumns()) + " / " + itoa(len(s.model.Relations)),
		"",
	}
	if len(issues) == 0 {
		lines = append(lines, "✓ 校验通过：表名 / 字段名无重复，主键齐全，关系引用有效。")
	} else {
		lines = append(lines, "发现 "+itoa(len(issues))+" 个问题：")
		for i, issue := range issues {
			lines = append(lines, fmt.Sprintf("  %d. %s", i+1, issue))
		}
	}
	lines = append(lines, "", "说明：本窗口只维护内存模型与本地 JSON，不连接任何数据库。")
	s.showModal("模型校验", "静态检查（不执行 SQL）", lines,
		[]ModalButton{{Text: "关闭", Primary: true, OnClick: s.CloseModal}})
}

func (s *Scene) Help() {
	lines := []string{
		"DB Modeler · 单画布纯自绘数据库建模示例",
		"",
		"【架构】Form 是唯一 HWND；ICustomGraphicControl 是唯一画布；",
		"工具栏 / 导航 / 属性面板 / 状态栏 / 菜单 / 弹窗 / 输入框全部纯 Go 自绘。",
		"",
		"【建模】",
		"  Ctrl+N            新建表（预置 id / name / created_at 三个模拟字段）",
		"  表头 +            给该表追加字段；表头 ▾ 折叠 / 展开",
		"  拖表头            移动表（Shift+点击可多选，一起移动）",
		"  拖字段圆点        拖到目标表字段 / 表头释放 = 创建外键关系",
		"  双击表头 / 字段   就地改名（支持中文输入法）",
		"  Shift+空白拖拽    框选多张表",
		"  右键              上下文菜单（表 / 关系 / 画布）",
		"",
		"【视图】",
		"  空白拖拽 / 中键   平移画布        滚轮 / + / -   缩放（以鼠标为锚点）",
		"  Ctrl+0            适应窗口        方向键          微调选中表位置",
		"",
		"【编辑与文件】",
		"  Ctrl+Z / Ctrl+Y   撤销 / 重做      Delete         删除选中对象",
		"  Ctrl+S / Ctrl+O   保存 / 打开 model.json（本地 JSON 持久化）",
		"  Ctrl+E            预览并导出 schema.sql（MySQL / PostgreSQL / SQLite 方言）",
		"  Ctrl+F            聚焦左侧搜索框   F1 本帮助      F2 重命名",
		"  Esc               取消连线 / 关闭菜单 / 取消选择 / 取消本次输入",
		"",
		"【输入】自绘输入框支持：鼠标点选光标、拖动选区、双击选词、",
		"Ctrl+A/C/V/X、Ctrl+Z/Y、Home/End、按词移动、长文本横向滚动、",
		"剪贴板、IME 预编辑与候选框定位；控制字符在事件边界被过滤。",
		"输入到一半按 Esc 会还原成进入编辑前的文本，不会误提交。",
		"",
		"【DDL 规则】1:1 在外键列生成 UNIQUE；N:M 生成真实的中间表",
		"（复合主键 + 两条外键）；SQLite 用建表内联 FOREIGN KEY。",
		"",
		"【说明】表/字段/关系数据是内存里的模拟电商数据，不建立任何数据库",
		"连接；生成的 SQL 只是文本，不会被执行。",
	}
	s.showModal("帮助 · 快捷键", "wg 示例集合 · 全自绘组件", lines,
		[]ModalButton{{Text: "知道了", Primary: true, OnClick: s.CloseModal}})
}
