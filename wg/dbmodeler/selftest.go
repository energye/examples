package main

// --selftest：不创建窗口、不连接数据库的纯逻辑回归。
// 覆盖模型 CRUD / 关系校验 / DDL 三方言 / JSON / 撤销重做 / 布局 /
// 坐标变换 / 自绘编辑器 / 场景命令 / 运行时导出。

import (
	"fmt"
	"strings"

	"github.com/energye/lcl/api/libname"
	"github.com/energye/lcl/types"
)

type check struct {
	name string
	fn   func() (bool, string)
}

func runSelfTest() int {
	fmt.Println("==== DBMODELER SELFTEST ====")
	fmt.Println("[INFO] runtime library:", libname.LibName)
	fmt.Println("[INFO] IME bindings:", imeSelfTest())
	fmt.Println("[INFO] library exports:", libSelfTest(libname.LibName))

	checks := []check{
		{"模拟数据规模与完整性", testDemoData},
		{"建表 / 删表级联关系", testTableCRUD},
		{"字段增删与关系级联", testColumnCRUD},
		{"关系创建与重复校验", testRelationRules},
		{"DDL 三方言生成", testDDLDialects},
		{"JSON 往返", testJSONRoundTrip},
		{"撤销 / 重做", testHistory},
		{"自动布局无重叠", testAutoLayout},
		{"世界坐标与屏幕坐标互转", testViewport},
		{"自绘输入框编辑与控制字符过滤", testEditor},
		{"场景命令与撤销", testSceneCommands},
		{"全自绘场景离屏布局", testHeadlessLayout},
		{"交互事件链路", func() (bool, string) {
			pass, fail := runUIChecks(false)
			return fail == 0, fmt.Sprintf("%d/%d 项通过", pass, pass+fail)
		}},
	}

	failed := 0
	for _, c := range checks {
		ok, detail := c.fn()
		state := "PASS"
		if !ok {
			state = "FAIL"
			failed++
		}
		fmt.Printf("  %-28s %s  %s\n", c.name, state, detail)
	}
	fmt.Printf("==== RESULT: %d/%d PASS ====\n", len(checks)-failed, len(checks))
	if failed > 0 {
		return 1
	}
	return 0
}

func testDemoData() (bool, string) {
	m := demoModel()
	cols := 0
	for i := range m.Tables {
		cols += len(m.Tables[i].Columns)
	}
	issues := m.Validate()
	if len(m.Tables) != 11 || len(m.Relations) != 14 || cols < 60 || len(issues) > 0 {
		return false, fmt.Sprintf("tables=%d relations=%d columns=%d issues=%v",
			len(m.Tables), len(m.Relations), cols, issues)
	}
	if m.TableByName("orders") == nil || m.TableByName("users") == nil {
		return false, "缺少关键模拟表"
	}
	return true, fmt.Sprintf("%d 表 / %d 字段 / %d 关系，校验通过",
		len(m.Tables), cols, len(m.Relations))
}

func testTableCRUD() (bool, string) {
	m := demoModel()
	before := len(m.Relations)
	t := m.AddTable(1000, 1000)
	name1 := t.Name
	colCount := len(t.Columns)
	if m.TableByName(name1) == nil || colCount != 3 {
		return false, "新表结构不正确"
	}
	t2 := m.AddTable(1000, 1000)
	if name1 == t2.Name {
		return false, "新表名未去重"
	}
	users := m.TableByName("users")
	if users == nil {
		return false, "缺少 users"
	}
	// TableBy* 返回切片内元素的指针：删表会就地搬移切片，先把 ID 取出来。
	uid := users.ID
	m.DeleteTable(uid)
	if m.TableByName("users") != nil {
		return false, "删表未生效"
	}
	for _, r := range m.Relations {
		if r.FromTable == uid || r.ToTable == uid {
			return false, "删表未级联删除关系"
		}
	}
	if len(m.Relations) >= before {
		return false, "关系数量未减少"
	}
	return true, "新表去重 / 默认 3 字段 / 删表级联删除外键"
}

func testColumnCRUD() (bool, string) {
	m := demoModel()
	t := m.TableByName("orders")
	relBefore := len(m.Relations)
	c := m.AddColumn(t.ID)
	if c == nil {
		return false, "添加字段失败"
	}
	cid := c.ID // 同样先取 ID：MoveColumn 会交换切片元素，指针内容会变
	if t.ColumnByID(cid) == nil {
		return false, "添加字段失败"
	}
	n := len(t.Columns)
	m.MoveColumn(t.ID, cid, -1)
	if t.Columns[n-2].ID != cid {
		return false, "字段上移失败"
	}
	// 删掉被关系引用的 user_id，应同时删除相关关系
	uid := t.ColumnByName("user_id")
	if uid == nil {
		return false, "缺少 user_id"
	}
	ucID := uid.ID
	m.DeleteColumn(t.ID, ucID)
	if t.ColumnByID(ucID) != nil {
		return false, "删除字段失败"
	}
	if len(m.Relations) >= relBefore {
		return false, "删除被引用字段未级联删除关系"
	}
	for _, r := range m.Relations {
		if r.FromColumn == ucID || r.ToColumn == ucID {
			return false, "存在悬空关系"
		}
	}
	return true, "字段增删 / 排序 / 级联清理关系"
}

func testRelationRules() (bool, string) {
	m := demoModel()
	items := m.TableByName("order_items")
	orders := m.TableByName("orders")
	if items == nil || orders == nil {
		return false, "缺少模拟表"
	}
	oc, ocid := items.ColumnByName("order_id"), orders.ColumnByName("id")
	if _, err := m.AddRelation(items.ID, oc.ID, orders.ID, ocid.ID); err == nil {
		return false, "重复关系未被拒绝"
	}
	if _, err := m.AddRelation(items.ID, oc.ID, items.ID, oc.ID); err == nil {
		return false, "字段自连未被拒绝"
	}
	if _, err := m.AddRelation(items.ID, "not-exist", orders.ID, ocid.ID); err == nil {
		return false, "悬空字段未被拒绝"
	}
	t := m.AddTable(2400, 100)
	rel, err := m.AddRelation(t.ID, t.Columns[1].ID, orders.ID, ocid.ID)
	if err != nil || rel == nil || rel.Cardinality != "1:N" {
		return false, "新建关系失败"
	}
	return true, "重复 / 自连 / 悬空均被拒绝，1:N 默认基数正确"
}

func testDDLDialects() (bool, string) {
	m := demoModel()
	want := []string{
		"CREATE TABLE", "users", "orders", "order_items", "PRIMARY KEY", "FOREIGN KEY",
	}
	for _, d := range Dialects {
		m.Dialect = d
		sql := m.DDL()
		for _, w := range want {
			if !strings.Contains(sql, w) {
				return false, d + " 缺少 " + w
			}
		}
	}
	m.Dialect = DialectMySQL
	if !strings.Contains(m.DDL(), "AUTO_INCREMENT") || !strings.Contains(m.DDL(), "ENGINE=InnoDB") {
		return false, "MySQL 方言缺少自增 / 引擎"
	}
	m.Dialect = DialectPostgreSQL
	if !strings.Contains(m.DDL(), "BIGSERIAL") || !strings.Contains(m.DDL(), "COMMENT ON COLUMN") {
		return false, "PostgreSQL 方言缺少 SERIAL / 注释语句"
	}
	m.Dialect = DialectSQLite
	if !strings.Contains(m.DDL(), "PRAGMA foreign_keys") ||
		!strings.Contains(m.DDL(), "AUTOINCREMENT") {
		return false, "SQLite 方言缺少 PRAGMA / AUTOINCREMENT"
	}
	return true, "MySQL / PostgreSQL / SQLite 关键字齐全"
}

func testJSONRoundTrip() (bool, string) {
	m := demoModel()
	m.Dialect = DialectPostgreSQL
	data := []byte(m.JSON())
	cp, err := ModelFromJSON(data)
	if err != nil {
		return false, err.Error()
	}
	if len(cp.Tables) != len(m.Tables) || len(cp.Relations) != len(m.Relations) ||
		cp.Dialect != m.Dialect || cp.Name != m.Name {
		return false, "往返后结构不一致"
	}
	if cp.TableByName("products").Columns[4].Name != m.TableByName("products").Columns[4].Name {
		return false, "字段顺序 / 内容不一致"
	}
	return true, fmt.Sprintf("%d 字节 JSON，结构完全一致", len(data))
}

func testHistory() (bool, string) {
	m := demoModel()
	h := &History{}
	n0 := len(m.Tables)
	h.Push(m)
	t := m.AddTable(500, 500)
	if !h.CanUndo() {
		return false, "撤销栈未记录"
	}
	h.Undo(m)
	if len(m.Tables) != n0 || m.TableByID(t.ID) != nil {
		return false, "撤销未恢复"
	}
	if !h.Redo(m) || len(m.Tables) != n0+1 {
		return false, "重做未恢复"
	}
	h.Push(m)
	m.DeleteTable(t.ID)
	h.DropLast()
	if len(h.undo) < 1 {
		return false, "DropLast 误删历史"
	}
	return true, "Push / Undo / Redo / DropLast 均正确"
}

func testAutoLayout() (bool, string) {
	m := demoModel()
	m.AutoLayout()
	for i := range m.Tables {
		a := &m.Tables[i]
		for j := i + 1; j < len(m.Tables); j++ {
			b := &m.Tables[j]
			if a.X < b.X+tableWidth && a.X+tableWidth > b.X &&
				a.Y < b.Y+TableHeight(b) && a.Y+TableHeight(a) > b.Y {
				return false, "表矩形重叠: " + a.Name + " / " + b.Name
			}
		}
	}
	minX, minY, maxX, maxY := m.Bounds()
	if minX < 0 || minY < 0 || maxX <= minX || maxY <= minY {
		return false, "包围盒异常"
	}
	return true, fmt.Sprintf("%d 张表按外键分层，无重叠，包围盒 %.0f×%.0f",
		len(m.Tables), maxX-minX, maxY-minY)
}

func testViewport() (bool, string) {
	s := NewScene(nil)
	s.Layout(1440, 880)
	v := s.view
	t := s.model.TableByName("users")
	if t == nil {
		return false, "缺少 users"
	}
	for _, z := range []float64{0.5, 1.0, 1.5} {
		v.zoom = z
		x, y := v.W2S(t.X, t.Y)
		wx, wy := v.S2W(x, y)
		if absF(wx-t.X) > 1.1 || absF(wy-t.Y) > 1.1 {
			return false, fmt.Sprintf("zoom %.1f 坐标往返误差 %.2f/%.2f", z, wx-t.X, wy-t.Y)
		}
	}
	v.zoom = 1
	r := v.tableRect(t)
	if !r.PtInRect(types.TPoint{X: r.Left + 10, Y: r.Top + 10}) {
		return false, "表矩形命中异常"
	}
	v.Fit()
	if v.zoom < 0.45 || v.zoom > 1.3 {
		return false, fmt.Sprintf("Fit 缩放越界 %.2f", v.zoom)
	}
	return true, "多缩放级别坐标往返 / Fit 缩放范围正确"
}

func testEditor() (bool, string) {
	if sanitizeText("a\x08b\x01c\x0D中") != "abc中" {
		return false, "控制字符过滤失败"
	}
	in := NewInput("t", r4(0, 0, 200, 28))
	in.insert("abc")
	in.insert("中")
	if in.Text() != "abc中" {
		return false, "中文插入失败"
	}
	in.Handle(&Event{Kind: EvKey, Key: vkLeft})
	in.insert("X")
	if in.Text() != "abcX中" {
		return false, "光标插入位置错误"
	}
	in.Handle(&Event{Kind: EvKey, Key: vkReturn})
	in.undo()
	if in.Text() != "abc中" {
		return false, "编辑器撤销失败"
	}
	in.SelectAll()
	lo, hi := in.selectionRange()
	if lo != 0 || hi != len([]rune(in.Text())) {
		return false, "全选范围错误"
	}
	return true, "控制字符 / 中文 / 光标 / 撤销 / 全选"
}

func testSceneCommands() (bool, string) {
	s := NewScene(nil)
	s.Layout(1280, 800)
	n0 := len(s.model.Tables)
	s.NewTable()
	if len(s.model.Tables) != n0+1 || s.sel.Kind != "table" {
		return false, "新建表命令失败"
	}
	id := s.sel.TableID
	c0 := len(s.model.TableByID(id).Columns)
	s.AddColumn(id)
	if len(s.model.TableByID(id).Columns) != c0+1 {
		return false, "添加字段命令失败"
	}
	s.DeleteSelection() // 删除刚添加的字段
	if s.sel.Kind != "table" {
		return false, "删除字段后选择状态错误"
	}
	s.Undo()
	s.Undo()
	s.Undo()
	if len(s.model.Tables) != n0 {
		return false, "三次撤销后未回到初始表数"
	}
	if s.inspWidgets == nil || len(s.toolbar) != 15 || len(s.navWidgets) != 5 {
		return false, "自绘控件集合数量异常"
	}
	return true, "新建表 / 加字段 / 删除 / 连续撤销 / 控件集合"
}

func testHeadlessLayout() (bool, string) {
	s := NewScene(nil)
	s.Layout(1440, 880)
	if s.viewRect.Width() <= 0 || s.viewRect.Height() <= 0 ||
		s.inspRect.Left <= s.viewRect.Left || s.navRect.Right != s.viewRect.Left {
		return false, "区域划分错误"
	}
	if s.inspContent <= 0 {
		return false, "属性面板内容高度未计算"
	}
	// 离屏渲染路径不依赖窗口：给 Painter 一个 nil Canvas 也不能 panic。
	p := &Painter{Theme: &s.theme, Scale: 1}
	s.PaintAll(p)
	s.ToggleTheme()
	s.ToggleDialect()
	if s.theme.Dark != true || s.model.Dialect != DialectPostgreSQL {
		return false, "主题 / 方言切换失败"
	}
	return true, fmt.Sprintf("区域 %d×%d，无画布渲染不崩溃，主题/方言切换正常",
		s.viewRect.Width(), s.viewRect.Height())
}
