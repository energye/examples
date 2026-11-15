package main

// 编辑器 / 滚动条 / 属性面板 / 文件命令的控件级回归。

import (
	"fmt"
	"os"
	"strings"

	"github.com/energye/lcl/types"
)

func paintOffscreen(s *Scene) {
	// 无画布也能算滚动与命中（字体宽度退化为估算），但走的是同一条 PaintAll 路径。
	s.painter.Canvas = nil
	s.PaintAll(s.painter)
}

func uiInputRealEdit(s *Scene) (bool, string) {
	t := s.model.TableByName("products")
	s.Select(Selection{Kind: "table", TableID: t.ID, Tables: map[string]bool{t.ID: true}})
	in := findInput(s, "表名")
	if in == nil {
		return false, "属性面板没有表名输入框"
	}
	in.SetText("products")
	s.SetFocus(in)
	paintOffscreen(s)

	r := in.Rect()
	click(s, r.Left+8, r.Top+r.Height()/2)
	if in.caret != 0 {
		return false, fmt.Sprintf("点击左端后光标在 %d（应为 0）", in.caret)
	}
	click(s, r.Right-6, r.Top+r.Height()/2)
	if in.caret != len([]rune("products")) {
		return false, fmt.Sprintf("点击右端后光标在 %d", in.caret)
	}

	// 拖动选区：从行首拖到第 3 个字符之后
	down(s, r.Left+8, r.Top+r.Height()/2, types.MbLeft, 0)
	move(s, r.Left+8+in.glyphX[3]+1, r.Top+r.Height()/2)
	up(s, r.Left+8+in.glyphX[3]+1, r.Top+r.Height()/2)
	lo, hi := in.selectionRange()
	if lo != 0 || hi != 3 {
		return false, fmt.Sprintf("拖动选区得到 %d..%d（应为 0..3）glyph=%v scroll=%d",
			lo, hi, in.glyphX, in.scrollX)
	}

	// Esc 取消：还原焦点文本，且不触发提交
	in.SelectAll()
	s.OnText("临时改名")
	if in.Text() != "临时改名" {
		return false, "输入未写入编辑器"
	}
	s.OnKey(vkEscape, 0)
	if in.Text() != "products" {
		return false, "Esc 未还原文本，当前 " + in.Text()
	}
	if s.focus != nil {
		return false, "Esc 后输入框仍持有焦点"
	}
	if s.model.TableByID(t.ID).Name != "products" {
		return false, "Esc 取消却提交了改名"
	}

	// Ctrl+Z / Ctrl+Y 属于输入框自身，不应撤销模型
	s.SetFocus(in)
	in.SelectAll()
	s.OnText("XYZ")
	s.OnKey('Z', ctrlSet())
	if in.Text() != "products" {
		return false, "Ctrl+Z 未撤销文本"
	}
	s.OnKey('Y', ctrlSet())
	if in.Text() != "XYZ" {
		return false, "Ctrl+Y 未重做文本"
	}
	if len(s.model.Tables) != 11 {
		return false, "文本框撤销误伤了模型"
	}
	s.OnKey(vkEscape, 0)

	// 超长文本：光标到末尾后必须横向滚动才能看见
	long := strings.Repeat("a", 90)
	in.SetText(long)
	s.SetFocus(in)
	s.OnKey(vkEnd, 0)
	paintOffscreen(s)
	if in.scrollX <= 0 {
		return false, "长文本未产生横向滚动"
	}
	if in.scrollX > in.textW {
		return false, "横向滚动超出文本宽度"
	}
	s.OnKey(vkEscape, 0)
	return true, "点选光标 / 拖动选区 / Esc 取消 / Ctrl+Z·Y / 长文本横向滚动"
}

func uiScrollbars(s *Scene) (bool, string) {
	// 列表滚动条
	l := s.navList
	l.SetRect(r4(s.navRect.Left+6, s.navRect.Top+90, s.navRect.Right-10, s.navRect.Top+290))
	if l.maxScroll() <= 0 {
		return false, "列表没有产生可滚动内容"
	}
	track := l.barTrack()
	thumb := scrollThumb(track, l.Scroll, l.maxScroll())
	down(s, (track.Left+track.Right)/2, (thumb.Top+thumb.Bottom)/2, types.MbLeft, 0)
	move(s, (track.Left+track.Right)/2, track.Bottom-2)
	up(s, (track.Left+track.Right)/2, track.Bottom-2)
	if l.Scroll != l.maxScroll() {
		return false, fmt.Sprintf("列表滑块拖到底得到 %d/%d", l.Scroll, l.maxScroll())
	}
	// 键盘导航也应改变选中行
	rows := len(l.Items)
	s.SetFocus(l)
	s.OnKey(vkDown, 0)
	s.OnKey(vkDown, 0)
	if l.cur != 1 {
		return false, fmt.Sprintf("列表方向键选中 %d（应为 1，共 %d 行）", l.cur, rows)
	}

	// 弹窗滚动条
	s.Help()
	d := s.modal
	if d == nil || d.maxScroll() <= 0 {
		return false, "帮助弹窗没有可滚动内容"
	}
	mtrack := d.barTrack()
	mthumb := scrollThumb(mtrack, d.Scroll, d.maxScroll())
	down(s, (mtrack.Left+mtrack.Right)/2, (mthumb.Top+mthumb.Bottom)/2, types.MbLeft, 0)
	move(s, (mtrack.Left+mtrack.Right)/2, mtrack.Bottom-2)
	up(s, (mtrack.Left+mtrack.Right)/2, mtrack.Bottom-2)
	if d.Scroll != d.maxScroll() {
		return false, fmt.Sprintf("弹窗滑块拖到底得到 %d/%d", d.Scroll, d.maxScroll())
	}
	s.OnKey(vkEscape, 0)

	// 属性面板滚动条（窄窗口 + 字段属性项）
	s.Layout(1000, 460)
	t := s.model.TableByName("users")
	c := t.ColumnByName("username")
	s.Select(Selection{Kind: "column", TableID: t.ID, ColumnID: c.ID,
		Tables: map[string]bool{t.ID: true}})
	if s.inspMaxScroll() <= 0 {
		return false, "字段属性面板没有溢出，无法测试滚动条"
	}
	itrack := s.inspBarTrack()
	ithumb := scrollThumb(itrack, s.inspScroll, s.inspMaxScroll())
	down(s, (itrack.Left+itrack.Right)/2, (ithumb.Top+ithumb.Bottom)/2, types.MbLeft, 0)
	move(s, (itrack.Left+itrack.Right)/2, itrack.Bottom-2)
	up(s, (itrack.Left+itrack.Right)/2, itrack.Bottom-2)
	if s.inspScroll != s.inspMaxScroll() {
		return false, fmt.Sprintf("属性面板滑块拖到底得到 %d/%d", s.inspScroll, s.inspMaxScroll())
	}
	return true, "列表 / 弹窗 / 属性面板三处滚动条均可拖动，列表支持键盘选行"
}

func uiCardinalityControl(s *Scene) (bool, string) {
	m := s.model
	m.Dialect = DialectMySQL
	rel := m.RelationByID(m.Relations[0].ID)
	if rel == nil {
		return false, "没有关系可测试"
	}
	s.Select(Selection{Kind: "relation", RelationID: rel.ID})
	c := findCombo(s, "1:1")
	if c == nil {
		return false, "关系属性里没有基数下拉"
	}
	clickWidget(s, c)
	if s.popup == nil {
		return false, "基数下拉未展开"
	}
	if !clickPopupItem(s, "1:1") {
		return false, "下拉里没有 1:1"
	}
	if m.RelationByID(rel.ID).Cardinality != "1:1" {
		return false, "选择 1:1 未写回模型"
	}
	child := m.TableByID(m.RelationByID(rel.ID).FromTable)
	fc := child.ColumnByID(m.RelationByID(rel.ID).FromColumn)
	if !strings.Contains(m.DDL(), "UNIQUE (`"+fc.Name+"`)") {
		return false, "1:1 未在 DDL 里生成 UNIQUE"
	}

	c = findCombo(s, "N:M")
	if c == nil {
		return false, "关系属性里没有 N:M"
	}
	clickWidget(s, c)
	if !clickPopupItem(s, "N:M") {
		return false, "下拉里没有 N:M"
	}
	rel = m.RelationByID(rel.ID)
	jun := child.Name + "_" + m.TableByID(rel.ToTable).Name
	if !strings.Contains(m.DDL(), "CREATE TABLE `"+jun+"`") {
		return false, "N:M 未生成中间表 " + jun
	}

	c = findCombo(s, "CASCADE")
	if c == nil {
		return false, "关系属性里没有 ON DELETE 下拉"
	}
	clickWidget(s, c)
	if !clickPopupItem(s, "CASCADE") {
		return false, "下拉里没有 CASCADE"
	}
	if m.RelationByID(rel.ID).OnDelete != "CASCADE" {
		return false, "ON DELETE 未写回模型"
	}
	return true, "基数 1:1→UNIQUE、N:M→中间表、ON DELETE→CASCADE 都在 DDL 中生效"
}

func uiColumnProps(s *Scene) (bool, string) {
	m := s.model
	t := m.TableByName("users")
	c := t.ColumnByName("username")
	cid, tid := c.ID, t.ID
	s.Select(Selection{Kind: "column", TableID: tid, ColumnID: cid,
		Tables: map[string]bool{tid: true}})

	commit := func(label, value string) bool {
		in := findInput(s, label)
		if in == nil {
			return false
		}
		s.SetFocus(in)
		in.SelectAll()
		s.OnText(value)
		s.OnKey(vkReturn, 0)
		return true
	}
	if !commit("字段名", "login_name") {
		return false, "没有字段名输入框"
	}
	if t.ColumnByID(cid) == nil || t.ColumnByID(cid).Name != "login_name" {
		return false, "字段名未提交"
	}
	if !commit("注释", "登录名") || t.ColumnByID(cid).Comment != "登录名" {
		return false, "注释未提交"
	}
	if !commit("默认值", "''") || t.ColumnByID(cid).Default != "''" {
		return false, "默认值未提交"
	}

	// 类型下拉：改成 TEXT 后长度应被清空
	combo := findCombo(s, "TEXT")
	if combo == nil {
		return false, "没有数据类型下拉"
	}
	clickWidget(s, combo)
	if !clickPopupItem(s, "TEXT") {
		return false, "类型下拉里没有 TEXT"
	}
	if cur := t.ColumnByID(cid); cur.Type != "TEXT" || cur.Length != "" {
		return false, fmt.Sprintf("类型切换得到 %s(%s)", cur.Type, cur.Length)
	}
	if !commit("长度 / 精度", "512") || t.ColumnByID(cid).Length != "512" {
		return false, "长度未提交"
	}

	// 勾选框：按当前状态取反 → 再点回来（demo 的 username 本来就是 UQ）
	chk := findCheck(s, "唯一 UQ")
	if chk == nil {
		return false, "没有唯一勾选框"
	}
	want := !t.ColumnByID(cid).UQ
	clickWidget(s, chk)
	if t.ColumnByID(cid).UQ != want {
		return false, fmt.Sprintf("唯一勾选框未生效：want=%v got=%v checked=%v",
			want, t.ColumnByID(cid).UQ, chk.Checked)
	}
	chk = findCheck(s, "唯一 UQ")
	if chk.Checked != want {
		return false, "勾选框显示状态与模型不一致"
	}
	clickWidget(s, chk)
	if t.ColumnByID(cid).UQ == want {
		return false, "唯一勾选框第二次点击未还原"
	}

	// 字段上移：从下标 1 移到 0；再点一次应提示边界且不产生新快照
	idx := 0
	for i := range t.Columns {
		if t.Columns[i].ID == cid {
			idx = i
		}
	}
	if idx != 1 {
		return false, "username 不在下标 1，无法测试上移"
	}
	up := findButton(s, "↑ 上移字段")
	if up == nil {
		return false, "没有上移按钮"
	}
	clickWidget(s, up)
	if t.Columns[0].ID != cid {
		return false, "上移未生效"
	}
	undoDepth := len(s.history.undo)
	up = findButton(s, "↑ 上移字段")
	clickWidget(s, up)
	if len(s.history.undo) != undoDepth {
		return false, "边界上移不应写入撤销栈"
	}
	if !strings.Contains(s.status, "边界") {
		return false, "边界上移没有提示"
	}

	// 删除字段
	del := findButton(s, "删除字段")
	if del == nil {
		return false, "没有删除字段按钮"
	}
	n := len(t.Columns)
	clickWidget(s, del)
	if len(m.TableByID(tid).Columns) != n-1 {
		return false, "删除字段未生效"
	}
	return true, "字段名/注释/默认值/类型/长度/勾选框/上移边界/删除全部真实提交"
}

func uiMultiSelectAlign(s *Scene) (bool, string) {
	m := s.model
	a, b := m.TableByName("users"), m.TableByName("orders")
	if a == nil || b == nil {
		return false, "缺少多选测试表"
	}
	ha := s.view.headerRect(a)
	click(s, ha.Left+30, ha.Top+10)
	hb := s.view.headerRect(b)
	down(s, hb.Left+30, hb.Top+10, types.MbLeft, shiftSet())
	up(s, hb.Left+30, hb.Top+10)
	if s.sel.Kind != "multi" || len(s.sel.Tables) != 2 {
		return false, fmt.Sprintf("Shift 点击得到 %s / %d 张", s.sel.Kind, len(s.sel.Tables))
	}
	// 先错开位置，确保对齐真的做事
	m.TableByID(a.ID).X += 40
	if btn := findButton(s, "左对齐"); btn == nil {
		return false, "多选属性面板没有左对齐"
	} else {
		clickWidget(s, btn)
	}
	if m.TableByID(a.ID).X != m.TableByID(b.ID).X {
		return false, "左对齐未生效"
	}
	if btn := findButton(s, "顶部对齐"); btn == nil {
		return false, "多选属性面板没有顶部对齐"
	} else {
		clickWidget(s, btn)
	}
	if m.TableByID(a.ID).Y != m.TableByID(b.ID).Y {
		return false, "顶部对齐未生效"
	}
	n := len(m.Tables)
	if btn := findButton(s, "批量删除"); btn == nil {
		return false, "多选属性面板没有批量删除"
	} else {
		clickWidget(s, btn)
	}
	if len(m.Tables) != n-2 {
		return false, fmt.Sprintf("批量删除后 %d 张表", len(m.Tables))
	}
	s.Undo()
	if len(m.Tables) != n {
		return false, "批量删除无法撤销"
	}
	return true, "Shift 多选 / 左对齐 / 顶部对齐 / 批量删除 + 撤销"
}

func uiSaveOpenExport(s *Scene) (bool, string) {
	m := s.model
	s.Save()
	raw, err := os.ReadFile(s.dataPath("model.json"))
	if err != nil {
		return false, "保存失败: " + err.Error()
	}
	parsed, err := ModelFromJSON(raw)
	if err != nil {
		return false, "保存的 JSON 无法解析: " + err.Error()
	}
	if len(parsed.Tables) != len(m.Tables) || len(parsed.Relations) != len(m.Relations) {
		return false, "保存内容与内存模型不一致"
	}

	t := m.TableByName("products")
	s.CommitName("table", t.ID, "", "products_dirty")
	if m.TableByName("products") != nil {
		return false, "改名前状态异常"
	}
	s.Open()
	if m.TableByName("products") == nil {
		return false, "打开后没有恢复保存时的模型"
	}

	s.ExportDDL()
	if s.modal == nil {
		return false, "导出未打开预览"
	}
	ddl, err := os.ReadFile(s.dataPath("schema.sql"))
	if err != nil {
		return false, "schema.sql 未落盘: " + err.Error()
	}
	if string(ddl) != m.DDL() {
		return false, "落盘的 SQL 与模型 DDL 不一致"
	}
	if !strings.Contains(string(ddl), "CREATE TABLE") || !strings.Contains(string(ddl), "FOREIGN KEY") {
		return false, "导出的 SQL 内容不完整"
	}
	lines := len(strings.Split(string(ddl), "\n"))
	s.CloseModal()
	return true, fmt.Sprintf("JSON %d 字节 / SQL %d 字节 %d 行，落盘内容与模型一致", len(raw), len(ddl), lines)
}
