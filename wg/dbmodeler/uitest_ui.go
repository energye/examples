package main

// 控件级 / 菜单级 / 快捷键级回归：每个按钮都真的点下去，每个菜单项都真的
// 通过命中测试点中，每个快捷键都真的走 Scene.OnKey，然后断言模型确实变了。

import (
	"fmt"
	"os"
	"strings"

	"github.com/energye/lcl/types"
)

func ctrlSet() types.TShiftState  { return types.NewSet(types.SsCtrl) }
func shiftSet() types.TShiftState { return types.NewSet(types.SsShift) }

func findButton(s *Scene, caption string) *Button {
	for _, w := range s.inspWidgets {
		if b, ok := w.(*Button); ok && b.Caption == caption {
			return b
		}
	}
	return nil
}

func findInput(s *Scene, label string) *Input {
	for _, w := range s.inspWidgets {
		if in, ok := w.(*Input); ok && in.Name() == "insp_"+label {
			return in
		}
	}
	return nil
}

func findCombo(s *Scene, item string) *Combo {
	for _, w := range s.inspWidgets {
		if c, ok := w.(*Combo); ok {
			for _, it := range c.Items {
				if it == item {
					return c
				}
			}
		}
	}
	return nil
}

func findCheck(s *Scene, caption string) *CheckBox {
	for _, w := range s.inspWidgets {
		if c, ok := w.(*CheckBox); ok && c.Caption == caption {
			return c
		}
	}
	return nil
}

func clickWidget(s *Scene, w Widget) {
	r := w.Rect()
	click(s, r.Left+r.Width()/2, r.Top+r.Height()/2)
}

func pressToolbar(s *Scene, key string) {
	b := s.buttons[key]
	if b == nil {
		return
	}
	clickWidget(s, b)
}

// clickPopupItem 通过真实的菜单命中测试点击一行（而不是直接调 OnClick）。
func clickPopupItem(s *Scene, label string) bool {
	if s.popup == nil {
		return false
	}
	for i, it := range s.popup.Items {
		if it.Sep || it.Text != label {
			continue
		}
		y := s.popup.itemTop(i) + s.popup.itemHeight(it)/2
		click(s, s.popup.rect.Left+40, y)
		return true
	}
	return false
}

func openContextMenu(s *Scene, x, y int32) {
	down(s, x, y, types.MbRight, 0)
	up(s, x, y)
}

func emptyCanvasPoint(s *Scene) (int32, int32) {
	for _, p := range []types.TPoint{
		{X: s.viewRect.Left + 40, Y: s.viewRect.Bottom - 50},
		{X: s.viewRect.Left + 40, Y: s.viewRect.Top + 80},
		{X: s.viewRect.Right - 60, Y: s.viewRect.Bottom - 50},
	} {
		if s.view.tableAt(p.X, p.Y) == nil && s.view.relationAt(p.X, p.Y) == "" {
			return p.X, p.Y
		}
	}
	return s.viewRect.Left + 30, s.viewRect.Bottom - 40
}

func uiToolbarButtons(s *Scene) (bool, string) {
	m := s.model
	n0 := len(m.Tables)
	pressToolbar(s, "newtable")
	if len(m.Tables) != n0+1 {
		return false, fmt.Sprintf("新建表 %d → %d", n0, len(m.Tables))
	}
	tid := s.sel.TableID
	if tid == "" {
		return false, "新建表后未选中"
	}
	// 新建表必须整块落在当前视口内，否则用户点了「新建表」却看不到表。
	if t := m.TableByID(tid); t != nil {
		tl := s.view.tableRect(t)
		vr := s.viewRect
		if tl.Left < vr.Left || tl.Top < vr.Top || tl.Right > vr.Right || tl.Bottom > vr.Bottom {
			return false, fmt.Sprintf("新建表落在视口外 %v (视口 %v)", tl, vr)
		}
	}
	cols := len(m.TableByID(tid).Columns)
	pressToolbar(s, "addcol")
	if len(m.TableByID(tid).Columns) != cols+1 {
		return false, "加字段按钮未生效"
	}
	pressToolbar(s, "undo")
	if got := len(m.TableByID(tid).Columns); got != cols {
		return false, fmt.Sprintf("撤销加字段后 %d 列", got)
	}
	pressToolbar(s, "redo")
	if got := len(m.TableByID(tid).Columns); got != cols+1 {
		return false, "重做未恢复字段"
	}
	t := m.TableByID(tid)
	hdr := s.view.headerRect(t)
	click(s, hdr.Left+30, hdr.Top+10)
	pressToolbar(s, "delete")
	if len(m.Tables) != n0 {
		return false, fmt.Sprintf("删除按钮未删除选中表：sel=%s tables=%d enabled=%v hdr=%v view=%v hit=%v",
			s.sel.Kind, len(m.Tables), s.buttons["delete"].Enabled(),
			s.view.headerRect(t), s.viewRect, s.view.tableAt(hdr.Left+30, hdr.Top+10) != nil)
	}
	pressToolbar(s, "undo")
	if len(m.Tables) != n0+1 {
		return false, "撤销删除未恢复"
	}

	t = m.TableByID(tid)
	t.X += 57
	t.Y += 33
	moved := fmt.Sprintf("%.0f,%.0f", t.X, t.Y)
	pressToolbar(s, "layout")
	if got := fmt.Sprintf("%.0f,%.0f", m.TableByID(tid).X, m.TableByID(tid).Y); got == moved {
		return false, "自动布局未改变被移动的表"
	}

	s.view.zoom = 1.5
	pressToolbar(s, "fit")
	if absF(s.view.zoom-1.5) < 1e-9 {
		return false, "适应按钮未改变缩放"
	}
	z := s.view.zoom
	pressToolbar(s, "zin")
	if s.view.zoom <= z {
		return false, "放大按钮无效"
	}
	z = s.view.zoom
	pressToolbar(s, "zout")
	if s.view.zoom >= z {
		return false, "缩小按钮无效"
	}

	pressToolbar(s, "ddl")
	if s.modal == nil || !strings.Contains(s.modal.Title, "DDL") {
		return false, "导出 DDL 未打开预览"
	}
	s.CloseModal()

	pressToolbar(s, "save")
	raw, err := os.ReadFile(s.dataPath("model.json"))
	if err != nil {
		return false, "保存未落盘: " + err.Error()
	}
	if !strings.Contains(string(raw), "new_table") {
		return false, "保存的 JSON 缺少当前模型"
	}

	s.CommitName("table", tid, "", "toolbar_renamed")
	if m.TableByID(tid) == nil || m.TableByID(tid).Name != "toolbar_renamed" {
		return false, "改名前状态异常"
	}
	pressToolbar(s, "open")
	if m.TableByName("new_table") == nil {
		return false, "打开未恢复保存时的表名"
	}

	d0 := m.Dialect
	pressToolbar(s, "dialect")
	if m.Dialect == d0 {
		return false, "方言切换无效"
	}
	dark := s.theme.Dark
	pressToolbar(s, "theme")
	if s.theme.Dark == dark {
		return false, "主题切换无效"
	}
	pressToolbar(s, "help")
	if s.modal == nil || len(s.modal.Lines) < 10 {
		return false, "帮助未打开"
	}
	s.CloseModal()
	return true, "15 个工具栏按钮逐个点击并验证结果"
}

func uiCanvasMenus(s *Scene) (bool, string) {
	m := s.model
	tableMenu := func(name string) *Table {
		t := m.TableByName(name)
		if t == nil {
			return nil
		}
		hdr := s.view.headerRect(t)
		openContextMenu(s, hdr.Left+40, hdr.Top+10)
		return t
	}

	t := tableMenu("products")
	if t == nil || s.popup == nil {
		return false, "表菜单未打开"
	}
	cols := len(t.Columns)
	if !clickPopupItem(s, "新增字段") || len(m.TableByName("products").Columns) != cols+1 {
		return false, "菜单 · 新增字段无效"
	}
	s.Undo()

	tableMenu("products")
	if !clickPopupItem(s, "折叠") {
		return false, "菜单里没有折叠项"
	}
	if !m.TableByName("products").Collapsed {
		return false, "菜单 · 折叠无效"
	}
	s.Undo()

	tableMenu("products")
	n := len(m.Tables)
	if !clickPopupItem(s, "复制表") || len(m.Tables) != n+1 {
		return false, "菜单 · 复制表无效"
	}
	s.Undo()

	tableMenu("products")
	if !clickPopupItem(s, "重命名表") || s.inline == nil {
		return false, "菜单 · 重命名未打开就地编辑"
	}
	s.OnKey(vkEscape, 0)
	if s.inline != nil {
		return false, "Esc 未收起就地编辑"
	}

	tableMenu("products")
	n = len(m.Tables)
	if !clickPopupItem(s, "删除表") || len(m.Tables) != n-1 {
		return false, "菜单 · 删除表无效"
	}
	s.Undo()

	// 关系菜单：在连线中段右键
	rel := m.RelationByID(m.Relations[0].ID)
	relID := rel.ID // 删除会就地搬移切片，指针里的 ID 会被后一条关系覆盖
	pts := s.view.relationPoints(rel)
	hitX, hitY := int32(0), int32(0)
	found := false
	for i := 0; i+1 < len(pts); i++ {
		mx := (pts[i].X + pts[i+1].X) / 2
		my := (pts[i].Y + pts[i+1].Y) / 2
		if s.view.tableAt(mx, my) == nil && s.view.relationAt(mx, my) == rel.ID {
			hitX, hitY, found = mx, my, true
			break
		}
	}
	if !found {
		return false, "找不到可点击的关系连线中段"
	}
	openContextMenu(s, hitX, hitY)
	if s.popup == nil || !strings.Contains(s.popup.Title, "关系") {
		return false, "关系菜单未打开"
	}
	if !clickPopupItem(s, "基数 1:1") || m.RelationByID(relID).Cardinality != "1:1" {
		return false, "菜单 · 基数 1:1 无效"
	}
	openContextMenu(s, hitX, hitY)
	if !clickPopupItem(s, "ON DELETE CASCADE") || m.RelationByID(relID).OnDelete != "CASCADE" {
		return false, "菜单 · ON DELETE CASCADE 无效"
	}
	openContextMenu(s, hitX, hitY)
	title := ""
	if s.popup != nil {
		title = s.popup.Title
	}
	selBefore := s.sel.Kind + "/" + s.sel.RelationID
	if !clickPopupItem(s, "删除关系") {
		return false, fmt.Sprintf("菜单 · 删除关系无效：popup=%q sel=%s", title, selBefore)
	}
	if m.RelationByID(relID) != nil {
		return false, fmt.Sprintf("删除关系后仍然存在：popup=%q relID=%s selBefore=%s selAfter=%s status=%q",
			title, relID, selBefore, s.sel.Kind+"/"+s.sel.RelationID, s.status)
	}

	// 画布空白菜单
	x, y := emptyCanvasPoint(s)
	openContextMenu(s, x, y)
	if s.popup == nil || !strings.Contains(s.popup.Title, "画布") {
		return false, "画布菜单未打开"
	}
	n = len(m.Tables)
	if !clickPopupItem(s, "新建表") || len(m.Tables) != n+1 {
		return false, "画布菜单 · 新建表无效"
	}
	s.Undo()

	openContextMenu(s, x, y)
	s.view.zoom = 1.4
	if !clickPopupItem(s, "适应窗口") || absF(s.view.zoom-1.4) < 1e-9 {
		return false, "画布菜单 · 适应窗口无效"
	}

	openContextMenu(s, x, y)
	if !clickPopupItem(s, "校验模型") || s.modal == nil || s.modal.Title != "模型校验" {
		return false, "画布菜单 · 校验模型无效"
	}
	s.CloseModal()

	openContextMenu(s, x, y)
	if !clickPopupItem(s, "导出 DDL") || s.modal == nil {
		return false, "画布菜单 · 导出 DDL 无效"
	}
	s.CloseModal()

	openContextMenu(s, x, y)
	if !clickPopupItem(s, "保存模型") {
		return false, "画布菜单 · 保存模型无效"
	}
	if _, err := os.Stat(s.dataPath("model.json")); err != nil {
		return false, "画布菜单保存未落盘"
	}

	s.CommitName("table", m.TableByName("products").ID, "", "products_tmp")
	openContextMenu(s, x, y)
	if !clickPopupItem(s, "打开模型") || m.TableByName("products") == nil {
		return false, "画布菜单 · 打开模型无效"
	}

	openContextMenu(s, x, y)
	if !clickPopupItem(s, "帮助 / 快捷键") || s.modal == nil {
		return false, "画布菜单 · 帮助无效"
	}
	s.CloseModal()

	// 自动布局：先挪一张表
	if t := m.TableByName("users"); t != nil {
		t.X += 90
	}
	openContextMenu(s, x, y)
	before := fmt.Sprintf("%.0f", m.TableByName("users").X)
	if !clickPopupItem(s, "自动布局") {
		return false, "画布菜单 · 自动布局无效"
	}
	if fmt.Sprintf("%.0f", m.TableByName("users").X) == before && before != "" {
		return false, "自动布局未改变位置"
	}
	return true, "表菜单 5 项 / 关系菜单 3 项 / 画布菜单 8 项全部点中并生效"
}

func uiShortcuts(s *Scene) (bool, string) {
	m := s.model
	// Ctrl+N / Ctrl+Z / Ctrl+Y
	n := len(m.Tables)
	s.OnKey('N', ctrlSet())
	if len(m.Tables) != n+1 {
		return false, "Ctrl+N 未新建表"
	}
	s.OnKey('Z', ctrlSet())
	if len(m.Tables) != n {
		return false, "Ctrl+Z 未撤销"
	}
	s.OnKey('Y', ctrlSet())
	if len(m.Tables) != n+1 {
		return false, "Ctrl+Y 未重做"
	}
	// Ctrl+S / Ctrl+O
	s.OnKey('S', ctrlSet())
	if _, err := os.Stat(s.dataPath("model.json")); err != nil {
		return false, "Ctrl+S 未保存"
	}
	tid := s.sel.TableID
	if t := m.TableByID(tid); t != nil {
		s.CommitName("table", tid, "", "shortcut_tmp")
	}
	s.OnKey('O', ctrlSet())
	if m.TableByName("new_table") == nil {
		return false, "Ctrl+O 未载回模型"
	}
	// Ctrl+E
	s.OnKey('E', ctrlSet())
	if s.modal == nil {
		return false, "Ctrl+E 未打开 DDL"
	}
	s.CloseModal()
	// Ctrl+F
	s.OnKey('F', ctrlSet())
	if s.focus != s.search {
		return false, "Ctrl+F 未聚焦搜索框"
	}
	// Ctrl+0
	s.view.zoom = 1.6
	s.SetFocus(nil)
	s.OnKey('0', ctrlSet())
	if absF(s.view.zoom-1.6) < 1e-9 {
		return false, "Ctrl+0 未适应窗口"
	}
	// +/- 缩放（不带 Ctrl）
	z := s.view.zoom
	s.OnKey(187, 0)
	if s.view.zoom <= z {
		return false, "+ 未放大"
	}
	z = s.view.zoom
	s.OnKey(189, 0)
	if s.view.zoom >= z {
		return false, "- 未缩小"
	}
	// 方向键微调
	t := m.TableByName("users")
	s.Select(Selection{Kind: "table", TableID: t.ID, Tables: map[string]bool{t.ID: true}})
	x0 := t.X
	s.OnKey(vkRight, 0)
	if m.TableByID(t.ID).X != x0+8 {
		return false, "方向键未微调 8px"
	}
	s.OnKey(vkRight, shiftSet())
	if m.TableByID(t.ID).X != x0+48 {
		return false, "Shift+方向键未微调 40px"
	}
	// F2 聚焦名称输入框
	s.OnKey(vkF2, 0)
	if _, ok := s.focus.(*Input); !ok {
		return false, "F2 未聚焦名称输入框"
	}
	// Tab 切换焦点
	before := s.focus
	s.OnKey(vkTab, 0)
	if s.focus == before {
		return false, "Tab 未切换焦点"
	}
	// Delete
	s.SetFocus(nil)
	uid := t.ID
	s.Select(Selection{Kind: "table", TableID: uid, Tables: map[string]bool{uid: true}})
	n = len(m.Tables)
	s.OnKey(vkDelete, 0)
	if len(m.Tables) != n-1 {
		return false, "Delete 未删除选中表"
	}
	s.Undo()
	// Esc 取消选择
	s.Select(Selection{Kind: "table", TableID: uid, Tables: map[string]bool{uid: true}})
	s.OnKey(vkEscape, 0)
	if s.sel.Kind != "none" {
		return false, "Esc 未取消选择"
	}
	// F1
	s.OnKey(vkF1, 0)
	if s.modal == nil {
		return false, "F1 未打开帮助"
	}
	s.CloseModal()
	return true, "Ctrl+N/Z/Y/S/O/E/F/0、+/-、方向键、F1/F2/Tab/Delete/Esc 全部生效"
}
