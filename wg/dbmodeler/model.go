// Package main / model.go
//
// 纯逻辑数据模型层：不 import lcl、不碰画布、不连接任何数据库。
// 整个建模器只维护内存中的 Model（表 / 字段 / 关系），"保存"落到本地 JSON，
// "导出"生成 SQL 文本，二者都是模拟数据，不会打开任何数据库连接。
// 这一层可以直接被 go test 和 --selftest 覆盖。

package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// ---- 结构定义 ----

type Column struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Length  string `json:"length,omitempty"`
	PK      bool   `json:"pk"`
	NN      bool   `json:"nn"`
	UQ      bool   `json:"uq"`
	AI      bool   `json:"ai"`
	Default string `json:"default,omitempty"`
	Comment string `json:"comment,omitempty"`
}

type Table struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Comment   string   `json:"comment,omitempty"`
	X         float64  `json:"x"`
	Y         float64  `json:"y"`
	Collapsed bool     `json:"collapsed"`
	Group     string   `json:"group,omitempty"` // 配色分组：core / trade / catalog / promo
	Columns   []Column `json:"columns"`
}

type Relation struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	FromTable   string `json:"fromTable"`   // 子表：持有外键
	FromColumn  string `json:"fromColumn"`  // 外键列 ID
	ToTable     string `json:"toTable"`     // 父表：被引用
	ToColumn    string `json:"toColumn"`    // 被引用列 ID（通常主键）
	Cardinality string `json:"cardinality"` // 1:N / 1:1 / N:M
	OnDelete    string `json:"onDelete"`    // RESTRICT / CASCADE / SET NULL / NO ACTION
}

type Model struct {
	Name      string     `json:"name"`
	Dialect   string     `json:"dialect"` // MySQL / PostgreSQL / SQLite（仅影响 DDL 文本）
	Seq       int        `json:"seq"`     // ID 生成器游标
	Tables    []Table    `json:"tables"`
	Relations []Relation `json:"relations"`
}

const (
	DialectMySQL      = "MySQL"
	DialectPostgreSQL = "PostgreSQL"
	DialectSQLite     = "SQLite"
)

var Dialects = []string{DialectMySQL, DialectPostgreSQL, DialectSQLite}

// DataTypes 是属性面板下拉框与 DDL 生成器共用的类型清单。
var DataTypes = []string{
	"BIGINT", "INT", "TINYINT", "VARCHAR", "CHAR", "TEXT",
	"DECIMAL", "FLOAT", "BOOLEAN", "DATETIME", "DATE", "TIME",
	"TIMESTAMP", "JSON", "BLOB",
}

func defaultLength(typ string) string {
	switch typ {
	case "VARCHAR":
		return "64"
	case "CHAR":
		return "32"
	case "DECIMAL":
		return "10,2"
	}
	return ""
}

func (c Column) TypeText() string {
	if c.Length == "" {
		return c.Type
	}
	return c.Type + "(" + c.Length + ")"
}

// ---- 查找与 ID ----

func (m *Model) TableByID(id string) *Table {
	for i := range m.Tables {
		if m.Tables[i].ID == id {
			return &m.Tables[i]
		}
	}
	return nil
}

// uniqueTableName 在重名时追加 _2 / _3 …，用于新建表与复制表。
func (m *Model) uniqueTableName(base string) string {
	name := base
	for i := 2; m.TableByName(name) != nil; i++ {
		name = fmt.Sprintf("%s_%d", base, i)
	}
	return name
}

func (m *Model) TableByName(name string) *Table {
	for i := range m.Tables {
		if strings.EqualFold(m.Tables[i].Name, name) {
			return &m.Tables[i]
		}
	}
	return nil
}

func (t *Table) ColumnByID(id string) *Column {
	for i := range t.Columns {
		if t.Columns[i].ID == id {
			return &t.Columns[i]
		}
	}
	return nil
}

func (t *Table) ColumnByName(name string) *Column {
	for i := range t.Columns {
		if strings.EqualFold(t.Columns[i].Name, name) {
			return &t.Columns[i]
		}
	}
	return nil
}

func (m *Model) RelationByID(id string) *Relation {
	for i := range m.Relations {
		if m.Relations[i].ID == id {
			return &m.Relations[i]
		}
	}
	return nil
}

func (m *Model) nextID(prefix string) string {
	m.Seq++
	return fmt.Sprintf("%s%d", prefix, m.Seq)
}

// ---- 序列化 / 快照 ----

func (m *Model) JSON() string {
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(b)
}

func ModelFromJSON(data []byte) (*Model, error) {
	var m Model
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	if m.Name == "" {
		m.Name = "未命名模型"
	}
	if m.Dialect == "" {
		m.Dialect = DialectMySQL
	}
	// 老文件可能没有 seq，补一个不冲突的游标。
	maxSeq := m.Seq
	for _, t := range m.Tables {
		var n int
		if _, err := fmt.Sscanf(t.ID, "t%d", &n); err == nil && n > maxSeq {
			maxSeq = n
		}
		for _, c := range t.Columns {
			if _, err := fmt.Sscanf(c.ID, "c%d", &n); err == nil && n > maxSeq {
				maxSeq = n
			}
		}
	}
	for _, r := range m.Relations {
		var n int
		if _, err := fmt.Sscanf(r.ID, "r%d", &n); err == nil && n > maxSeq {
			maxSeq = n
		}
	}
	m.Seq = maxSeq
	return &m, nil
}

func (m *Model) Clone() *Model {
	cp, err := ModelFromJSON([]byte(m.JSON()))
	if err != nil {
		// JSON 往返对纯数据结构不会失败；兜底返回浅拷贝，避免调用方拿到 nil。
		shallow := *m
		return &shallow
	}
	return cp
}

func (m *Model) Restore(src *Model) {
	if src == nil {
		return
	}
	*m = *src.Clone()
}

// ---- 增删改 ----

func (m *Model) AddTable(x, y float64) *Table {
	name := m.uniqueTableName("new_table")
	id := m.nextID("t")
	m.Tables = append(m.Tables, Table{
		ID: id, Name: name, Comment: "新建表", X: x, Y: y, Group: "core",
		Columns: []Column{
			{ID: m.nextID("c"), Name: "id", Type: "BIGINT", PK: true, NN: true, AI: true, Comment: "主键"},
			{ID: m.nextID("c"), Name: "name", Type: "VARCHAR", Length: "64", NN: true, Comment: "名称"},
			{ID: m.nextID("c"), Name: "created_at", Type: "DATETIME", NN: true, Default: "CURRENT_TIMESTAMP", Comment: "创建时间"},
		},
	})
	return &m.Tables[len(m.Tables)-1]
}

func (m *Model) DeleteTable(id string) {
	for i := range m.Tables {
		if m.Tables[i].ID != id {
			continue
		}
		m.Tables = append(m.Tables[:i], m.Tables[i+1:]...)
		break
	}
	out := m.Relations[:0]
	for _, r := range m.Relations {
		if r.FromTable != id && r.ToTable != id {
			out = append(out, r)
		}
	}
	m.Relations = out
}

func (m *Model) AddColumn(tableID string) *Column {
	t := m.TableByID(tableID)
	if t == nil {
		return nil
	}
	name := "new_column"
	for i := 1; t.ColumnByName(name) != nil; i++ {
		name = fmt.Sprintf("new_column_%d", i)
	}
	t.Columns = append(t.Columns, Column{
		ID: m.nextID("c"), Name: name, Type: "VARCHAR", Length: "64", Comment: "新字段",
	})
	return &t.Columns[len(t.Columns)-1]
}

func (m *Model) DeleteColumn(tableID, colID string) {
	t := m.TableByID(tableID)
	if t == nil {
		return
	}
	for i := range t.Columns {
		if t.Columns[i].ID == colID {
			t.Columns = append(t.Columns[:i], t.Columns[i+1:]...)
			break
		}
	}
	out := m.Relations[:0]
	for _, r := range m.Relations {
		if (r.FromTable == tableID && r.FromColumn == colID) ||
			(r.ToTable == tableID && r.ToColumn == colID) {
			continue
		}
		out = append(out, r)
	}
	m.Relations = out
}

func (m *Model) MoveColumn(tableID, colID string, delta int) {
	t := m.TableByID(tableID)
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
	if idx < 0 {
		return
	}
	dst := idx + delta
	if dst < 0 || dst >= len(t.Columns) {
		return
	}
	t.Columns[idx], t.Columns[dst] = t.Columns[dst], t.Columns[idx]
}

// AddRelation 从「外键列」拖到「被引用列」时创建关系。返回可读的错误提示。
func (m *Model) AddRelation(fromTable, fromCol, toTable, toCol string) (*Relation, error) {
	ft, tt := m.TableByID(fromTable), m.TableByID(toTable)
	if ft == nil || tt == nil {
		return nil, fmt.Errorf("目标表不存在")
	}
	fc, tc := ft.ColumnByID(fromCol), tt.ColumnByID(toCol)
	if fc == nil || tc == nil {
		return nil, fmt.Errorf("目标字段不存在")
	}
	if fromTable == toTable && fromCol == toCol {
		return nil, fmt.Errorf("字段不能连接到自己")
	}
	for _, r := range m.Relations {
		if r.FromTable == fromTable && r.FromColumn == fromCol &&
			r.ToTable == toTable && r.ToColumn == toCol {
			return nil, fmt.Errorf("该关系已存在")
		}
	}
	base := "fk_" + ft.Name + "_" + fc.Name
	name := base
	for i := 1; ; i++ {
		used := false
		for _, r := range m.Relations {
			if r.Name == name {
				used = true
				break
			}
		}
		if !used {
			break
		}
		name = fmt.Sprintf("%s_%d", base, i)
	}
	m.Relations = append(m.Relations, Relation{
		ID:          m.nextID("r"),
		Name:        name,
		FromTable:   fromTable,
		FromColumn:  fromCol,
		ToTable:     toTable,
		ToColumn:    toCol,
		Cardinality: "1:N",
		OnDelete:    "RESTRICT",
	})
	return &m.Relations[len(m.Relations)-1], nil
}

func (m *Model) DeleteRelation(id string) {
	for i := range m.Relations {
		if m.Relations[i].ID == id {
			m.Relations = append(m.Relations[:i], m.Relations[i+1:]...)
			return
		}
	}
}

// IsForeignKey 供画布画 FK 标记、属性面板显示只读外键信息。
func (m *Model) IsForeignKey(tableID, colID string) bool {
	for _, r := range m.Relations {
		if r.FromTable == tableID && r.FromColumn == colID {
			return true
		}
	}
	return false
}

// Validate 返回模型里的所有问题（重复名、悬空引用等）。空切片表示健康。
func (m *Model) Validate() []string {
	var issues []string
	seenTable := map[string]bool{}
	for i := range m.Tables {
		t := &m.Tables[i]
		if strings.TrimSpace(t.Name) == "" {
			issues = append(issues, "表名称为空: "+t.ID)
		} else if seenTable[strings.ToLower(t.Name)] {
			issues = append(issues, "表名重复: "+t.Name)
		}
		seenTable[strings.ToLower(t.Name)] = true
		if len(t.Columns) == 0 {
			issues = append(issues, "表没有任何字段: "+t.Name)
		}
		seenCol := map[string]bool{}
		pk := 0
		for _, c := range t.Columns {
			if strings.TrimSpace(c.Name) == "" {
				issues = append(issues, "字段名为空: "+t.Name)
			}
			k := strings.ToLower(c.Name)
			if seenCol[k] {
				issues = append(issues, "字段重复: "+t.Name+"."+c.Name)
			}
			seenCol[k] = true
			if c.PK {
				pk++
			}
			if c.AI && c.Type != "BIGINT" && c.Type != "INT" && c.Type != "TINYINT" {
				issues = append(issues, "自增字段类型不支持: "+t.Name+"."+c.Name+" ("+c.Type+")")
			}
			if c.PK && c.UQ {
				issues = append(issues, "主键字段重复声明唯一: "+t.Name+"."+c.Name)
			}
		}
		if pk == 0 {
			issues = append(issues, "缺少主键: "+t.Name)
		}
	}
	seenRel := map[string]bool{}
	for _, r := range m.Relations {
		if strings.TrimSpace(r.Name) == "" {
			issues = append(issues, "关系名为空（"+r.ID+"）")
		} else if seenRel[strings.ToLower(r.Name)] {
			issues = append(issues, "关系名重复: "+r.Name)
		}
		seenRel[strings.ToLower(r.Name)] = true
		ft, tt := m.TableByID(r.FromTable), m.TableByID(r.ToTable)
		if ft == nil || tt == nil {
			issues = append(issues, "关系引用了不存在的表: "+r.Name)
			continue
		}
		fc, tc := ft.ColumnByID(r.FromColumn), tt.ColumnByID(r.ToColumn)
		if fc == nil || tc == nil {
			issues = append(issues, "关系引用了不存在的字段: "+r.Name)
			continue
		}
		if !strings.EqualFold(fc.Type, tc.Type) || fc.Length != tc.Length {
			issues = append(issues, "外键类型不一致: "+ft.Name+"."+fc.Name+" ("+fc.TypeText()+
				") → "+tt.Name+"."+tc.Name+" ("+tc.TypeText()+")")
		}
		if !tc.PK && !tc.UQ {
			issues = append(issues, "被引用字段不是主键 / 唯一: "+tt.Name+"."+tc.Name)
		}
		switch r.Cardinality {
		case "1:N", "1:1", "N:M":
		default:
			issues = append(issues, "未知基数: "+r.Name+" ("+r.Cardinality+")")
		}
		switch r.OnDelete {
		case "RESTRICT", "CASCADE", "SET NULL", "NO ACTION":
		default:
			issues = append(issues, "未知 ON DELETE 动作: "+r.Name+" ("+r.OnDelete+")")
		}
		if r.OnDelete == "SET NULL" && fc.NN {
			issues = append(issues, "ON DELETE SET NULL 但外键列非空: "+ft.Name+"."+fc.Name)
		}
		if r.Cardinality == "1:1" && r.FromTable == r.ToTable {
			issues = append(issues, "1:1 自引用关系无法建唯一约束: "+r.Name)
		}
	}
	return issues
}

// ---- 布局 ----

// TableHeight 是画布命中测试与自动布局共用的世界坐标高度。
func TableHeight(t *Table) float64 {
	if t.Collapsed {
		return 38
	}
	return float64(38 + len(t.Columns)*26 + 8)
}

const tableWidth = 236.0

func (m *Model) Bounds() (minX, minY, maxX, maxY float64) {
	if len(m.Tables) == 0 {
		return 0, 0, 800, 600
	}
	minX, minY = 1e9, 1e9
	maxX, maxY = -1e9, -1e9
	for i := range m.Tables {
		t := &m.Tables[i]
		minX = minFloat(minX, t.X)
		minY = minFloat(minY, t.Y)
		maxX = maxFloat(maxX, t.X+tableWidth)
		maxY = maxFloat(maxY, t.Y+TableHeight(t))
	}
	return
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// AutoLayout 按「父表在左、子表在右」分层排布，自引用关系不计层级。
func (m *Model) AutoLayout() {
	level := map[string]int{}
	for i := range m.Tables {
		level[m.Tables[i].ID] = 0
	}
	for pass := 0; pass < len(m.Tables)+1; pass++ {
		changed := false
		for _, r := range m.Relations {
			if r.FromTable == r.ToTable {
				continue
			}
			if _, ok := level[r.FromTable]; !ok {
				continue
			}
			if _, ok := level[r.ToTable]; !ok {
				continue
			}
			want := level[r.ToTable] + 1
			if want > 4 {
				want = 4 // 超过 4 层就折回，避免模型一路铺到屏幕外
			}
			if level[r.FromTable] < want {
				level[r.FromTable] = want
				changed = true
			}
		}
		if !changed {
			break
		}
	}
	byLevel := map[int][]*Table{}
	for i := range m.Tables {
		t := &m.Tables[i]
		byLevel[level[t.ID]] = append(byLevel[level[t.ID]], t)
	}
	for lvl := range byLevel {
		sort.Slice(byLevel[lvl], func(a, b int) bool {
			return byLevel[lvl][a].Name < byLevel[lvl][b].Name
		})
	}
	// 同一层内的表按名字排序竖向排布；层与层用固定列宽隔开，所以任意层、
	// 任意表都不会互相重叠。模型很大时可以继续向下铺，用画布平移查看。
	for lvl, list := range byLevel {
		y := 48.0
		for _, t := range list {
			t.X = 48 + float64(lvl)*296
			t.Y = y
			y += TableHeight(t) + 26
		}
	}
}

// ---- DDL（只生成文本，不执行）----

// oneLine 把用户输入的注释压成一行，避免换行破坏 -- / '...' 形式的 SQL 注释。
func oneLine(s string) string {
	return strings.TrimSpace(strings.NewReplacer("\r\n", " ", "\n", " ", "\r", " ").Replace(s))
}

func quoteIdent(s, dialect string) string {
	if dialect == DialectMySQL {
		return "`" + strings.ReplaceAll(s, "`", "``") + "`"
	}
	return "\"" + strings.ReplaceAll(s, "\"", "\"\"") + "\""
}

func sqlType(c Column, dialect string) string {
	t := strings.ToUpper(c.Type)
	lenPart := ""
	if c.Length != "" {
		lenPart = "(" + c.Length + ")"
	}
	switch dialect {
	case DialectPostgreSQL:
		switch t {
		case "INT", "TINYINT":
			return "INTEGER"
		case "DATETIME":
			return "TIMESTAMP"
		case "DECIMAL":
			return "NUMERIC" + lenPart
		case "JSON":
			return "JSONB"
		case "BLOB":
			return "BYTEA"
		case "FLOAT":
			return "DOUBLE PRECISION"
		default:
			return t + lenPart
		}
	case DialectSQLite:
		switch t {
		case "BIGINT", "INT", "TINYINT", "BOOLEAN":
			return "INTEGER"
		case "VARCHAR", "CHAR", "TEXT", "DATETIME", "DATE", "TIME", "TIMESTAMP", "JSON":
			return "TEXT"
		case "DECIMAL":
			return "NUMERIC" + lenPart
		case "BLOB":
			return "BLOB"
		default:
			return t + lenPart
		}
	default:
		return t + lenPart
	}
}

func (m *Model) DDL() string {
	d := m.Dialect
	if d == "" {
		d = DialectMySQL
	}
	plan := m.ddlPlan()
	var b strings.Builder
	columns := 0
	for i := range m.Tables {
		columns += len(m.Tables[i].Columns)
	}
	b.WriteString("-- " + m.Name + " · wg/dbmodeler 生成的模拟 DDL\n")
	b.WriteString("-- 方言: " + d + " | 表: " + itoa(len(m.Tables)) +
		" | 字段: " + itoa(columns) + " | 关系: " + itoa(len(m.Relations)) + "\n")
	if len(plan.junctions) > 0 {
		b.WriteString("-- N:M 关系已展开为中间表: " + itoa(len(plan.junctions)) + " 张\n")
	}
	b.WriteString("-- 说明: 纯内存模型生成，未连接任何数据库\n\n")
	if d == DialectSQLite {
		b.WriteString("PRAGMA foreign_keys = ON;\n\n")
	}
	for i := range m.Tables {
		t := &m.Tables[i]
		lines := m.tableLines(t, d, plan, true)
		b.WriteString("CREATE TABLE " + quoteIdent(t.Name, d) + " (\n")
		b.WriteString(strings.Join(lines, ",\n"))
		b.WriteString("\n)")
		if d == DialectMySQL {
			b.WriteString(" ENGINE=InnoDB DEFAULT CHARSET=utf8mb4")
			if t.Comment != "" {
				b.WriteString(" COMMENT='" + strings.ReplaceAll(oneLine(t.Comment), "'", "''") + "'")
			}
		}
		b.WriteString(";\n")
		if d == DialectPostgreSQL && t.Comment != "" {
			b.WriteString("COMMENT ON TABLE " + quoteIdent(t.Name, d) +
				" IS '" + strings.ReplaceAll(oneLine(t.Comment), "'", "''") + "';\n")
		}
		if d == DialectPostgreSQL {
			for _, c := range t.Columns {
				if c.Comment != "" {
					b.WriteString("COMMENT ON COLUMN " + quoteIdent(t.Name, d) + "." +
						quoteIdent(c.Name, d) + " IS '" +
						strings.ReplaceAll(oneLine(c.Comment), "'", "''") + "';\n")
				}
			}
		}
		b.WriteString("\n")
	}
	// N:M 关系的中间表：真实建表 + 复合主键 + 两条外键，而不是只写一行注释。
	for _, j := range plan.junctions {
		b.WriteString("-- N:M 中间表 " + j.relName + "（" + j.fromName + " ↔ " + j.toName + "）\n")
		b.WriteString("CREATE TABLE " + quoteIdent(j.name, d) + " (\n")
		lines := []string{
			"  " + quoteIdent(j.left.Name, d) + " " + sqlType(j.left, d) + " NOT NULL",
			"  " + quoteIdent(j.right.Name, d) + " " + sqlType(j.right, d) + " NOT NULL",
			"  PRIMARY KEY (" + quoteIdent(j.left.Name, d) + ", " + quoteIdent(j.right.Name, d) + ")",
		}
		if d == DialectSQLite {
			lines = append(lines,
				"  FOREIGN KEY ("+quoteIdent(j.left.Name, d)+") REFERENCES "+
					quoteIdent(j.fromName, d)+" ("+quoteIdent(j.fromCol, d)+") ON DELETE "+j.leftFK.onDelete,
				"  FOREIGN KEY ("+quoteIdent(j.right.Name, d)+") REFERENCES "+
					quoteIdent(j.toName, d)+" ("+quoteIdent(j.toCol, d)+") ON DELETE "+j.rightFK.onDelete)
		}
		b.WriteString(strings.Join(lines, ",\n"))
		b.WriteString("\n);\n\n")
	}
	// MySQL / PostgreSQL 用 ALTER TABLE 追加外键；SQLite 已在建表语句内联。
	if d != DialectSQLite {
		fks := plan.allFKs()
		if len(fks) > 0 {
			b.WriteString("-- 外键约束\n")
			for _, fk := range fks {
				b.WriteString("ALTER TABLE " + quoteIdent(fk.childTable, d) +
					" ADD CONSTRAINT " + quoteIdent(fk.name, d) +
					" FOREIGN KEY (" + quoteIdent(fk.childCol, d) + ")" +
					" REFERENCES " + quoteIdent(fk.parentTable, d) +
					" (" + quoteIdent(fk.parentCol, d) + ")" +
					" ON DELETE " + fk.onDelete + ";\n")
			}
		}
	}
	return b.String()
}

// fkSpec 是一条真正会写进 DDL 的外键约束。
type fkSpec struct {
	childTable  string
	childCol    string
	parentTable string
	parentCol   string
	name        string
	onDelete    string
}

// junctionSpec 是 N:M 关系展开出来的中间表。
type junctionSpec struct {
	name     string
	relName  string
	fromName string
	fromCol  string
	toName   string
	toCol    string
	left     Column
	right    Column
	leftFK   fkSpec
	rightFK  fkSpec
}

type ddlPlanT struct {
	fks       map[string][]fkSpec // 子表名 -> 外键（SQLite 内联用）
	alter     []fkSpec            // MySQL / PostgreSQL 的 ALTER 列表
	unique1x1 map[string]map[string]bool
	junctions []junctionSpec
}

func (p ddlPlanT) allFKs() []fkSpec {
	out := make([]fkSpec, 0, len(p.alter))
	out = append(out, p.alter...)
	for _, j := range p.junctions {
		out = append(out, j.leftFK, j.rightFK)
	}
	return out
}

// ddlPlan 把模型翻译成「外键 + 1:1 唯一约束 + N:M 中间表」。
// 名字都做去重，保证生成的多条 CREATE/ALTER 不会撞名。
func (m *Model) ddlPlan() ddlPlanT {
	plan := ddlPlanT{
		fks:       map[string][]fkSpec{},
		unique1x1: map[string]map[string]bool{},
	}
	usedTable := map[string]bool{}
	for i := range m.Tables {
		usedTable[strings.ToLower(m.Tables[i].Name)] = true
	}
	usedFK := map[string]bool{}
	uniqueName := func(base string, used map[string]bool) string {
		name := base
		for i := 2; used[strings.ToLower(name)]; i++ {
			name = fmt.Sprintf("%s_%d", base, i)
		}
		used[strings.ToLower(name)] = true
		return name
	}

	for _, r := range m.Relations {
		ft, tt := m.TableByID(r.FromTable), m.TableByID(r.ToTable)
		if ft == nil || tt == nil {
			continue
		}
		fc, tc := ft.ColumnByID(r.FromColumn), tt.ColumnByID(r.ToColumn)
		if fc == nil || tc == nil {
			continue
		}
		if r.Cardinality == "N:M" {
			name := uniqueName(ft.Name+"_"+tt.Name, usedTable)
			leftName := ft.Name + "_" + fc.Name
			rightName := tt.Name + "_" + tc.Name
			if leftName == rightName {
				rightName += "_2"
			}
			left := *fc
			left.Name, left.PK, left.NN, left.UQ, left.AI, left.Default, left.Comment = leftName, true, true, false, false, "", ""
			right := *tc
			right.Name, right.PK, right.NN, right.UQ, right.AI, right.Default, right.Comment = rightName, true, true, false, false, "", ""
			j := junctionSpec{
				name: name, relName: r.Name,
				fromName: ft.Name, fromCol: fc.Name,
				toName: tt.Name, toCol: tc.Name,
				left: left, right: right,
				leftFK: fkSpec{childTable: name, childCol: leftName,
					parentTable: ft.Name, parentCol: fc.Name,
					name: uniqueName(r.Name+"_from", usedFK), onDelete: r.OnDelete},
				rightFK: fkSpec{childTable: name, childCol: rightName,
					parentTable: tt.Name, parentCol: tc.Name,
					name: uniqueName(r.Name+"_to", usedFK), onDelete: r.OnDelete},
			}
			plan.junctions = append(plan.junctions, j)
			continue
		}
		fk := fkSpec{
			childTable: ft.Name, childCol: fc.Name,
			parentTable: tt.Name, parentCol: tc.Name,
			name: uniqueName(r.Name, usedFK), onDelete: r.OnDelete,
		}
		plan.fks[ft.Name] = append(plan.fks[ft.Name], fk)
		plan.alter = append(plan.alter, fk)
		if r.Cardinality == "1:1" {
			if plan.unique1x1[ft.Name] == nil {
				plan.unique1x1[ft.Name] = map[string]bool{}
			}
			plan.unique1x1[ft.Name][fc.Name] = true
		}
	}
	return plan
}

// tableLines 生成一张表的列定义 + 主键 / 唯一 / （SQLite）外键。
func (m *Model) tableLines(t *Table, d string, plan ddlPlanT, withFK bool) []string {
	var lines, pks, uqs []string
	sqliteRowID := false
	uqSeen := map[string]bool{}
	for _, c := range t.Columns {
		comment := oneLine(c.Comment)
		col := "  " + quoteIdent(c.Name, d) + " " + sqlType(c, d)
		if d == DialectMySQL {
			if c.AI {
				col += " AUTO_INCREMENT"
			}
			if c.NN || c.PK {
				col += " NOT NULL"
			}
			if c.Default != "" {
				col += " DEFAULT " + c.Default
			}
			if comment != "" {
				col += " COMMENT '" + strings.ReplaceAll(comment, "'", "''") + "'"
			}
		} else if d == DialectPostgreSQL {
			if c.AI && (c.Type == "BIGINT" || c.Type == "INT") {
				if strings.EqualFold(c.Type, "BIGINT") {
					col = "  " + quoteIdent(c.Name, d) + " BIGSERIAL"
				} else {
					col = "  " + quoteIdent(c.Name, d) + " SERIAL"
				}
			} else if c.Default != "" {
				col += " DEFAULT " + c.Default
			}
			if c.NN || c.PK {
				col += " NOT NULL"
			}
		} else { // SQLite
			if c.AI && c.PK && strings.EqualFold(sqlType(c, d), "INTEGER") {
				col = "  " + quoteIdent(c.Name, d) + " INTEGER PRIMARY KEY AUTOINCREMENT"
				sqliteRowID = true
			} else {
				if c.Default != "" {
					col += " DEFAULT " + c.Default
				}
				if c.NN || c.PK {
					col += " NOT NULL"
				}
				// SQLite 的注释只能单独占一行：写在列尾会把后面的逗号一起注释掉。
				// 这里把注释和列定义放在同一个元素里，join 的逗号仍落在列定义后。
				if comment != "" {
					col = "  -- " + comment + "\n" + col
				}
			}
		}
		lines = append(lines, col)
		if c.PK {
			pks = append(pks, quoteIdent(c.Name, d))
		}
		if c.UQ && !c.PK {
			uqs = append(uqs, quoteIdent(c.Name, d))
			uqSeen[c.Name] = true
		}
	}
	// 1:1 关系必须在外键列上加唯一约束，否则「一对一」在数据库里退化成一對多。
	for col := range plan.unique1x1[t.Name] {
		if uqSeen[col] {
			continue
		}
		c := t.ColumnByName(col)
		if c != nil && !c.PK {
			uqs = append(uqs, quoteIdent(c.Name, d))
			uqSeen[col] = true
		}
	}
	if len(pks) > 0 && !sqliteRowID {
		lines = append(lines, "  PRIMARY KEY ("+strings.Join(pks, ", ")+")")
	}
	for _, u := range uqs {
		lines = append(lines, "  UNIQUE ("+u+")")
	}
	if withFK && d == DialectSQLite {
		for _, fk := range plan.fks[t.Name] {
			lines = append(lines, "  FOREIGN KEY ("+quoteIdent(fk.childCol, d)+") REFERENCES "+
				quoteIdent(fk.parentTable, d)+" ("+quoteIdent(fk.parentCol, d)+") ON DELETE "+fk.onDelete)
		}
	}
	return lines
}

// ---- 撤销 / 重做（快照栈）----

type History struct {
	undo []string
	redo []string
}

func (h *History) Reset() {
	h.undo = nil
	h.redo = nil
}
func (h *History) Push(m *Model) {
	h.undo = append(h.undo, m.JSON())
	if len(h.undo) > 60 {
		h.undo = h.undo[1:]
	}
	h.redo = nil
}
func (h *History) Undo(m *Model) bool {
	if len(h.undo) == 0 {
		return false
	}
	last := h.undo[len(h.undo)-1]
	h.undo = h.undo[:len(h.undo)-1]
	h.redo = append(h.redo, m.JSON())
	snap, err := ModelFromJSON([]byte(last))
	if err != nil {
		return false
	}
	m.Restore(snap)
	return true
}
func (h *History) Redo(m *Model) bool {
	if len(h.redo) == 0 {
		return false
	}
	last := h.redo[len(h.redo)-1]
	h.redo = h.redo[:len(h.redo)-1]
	h.undo = append(h.undo, m.JSON())
	snap, err := ModelFromJSON([]byte(last))
	if err != nil {
		return false
	}
	m.Restore(snap)
	return true
}

// DropLast 丢弃最近一次 Push（用于「预压快照但操作被校验拒绝」的回滚）。
func (h *History) DropLast() {
	if len(h.undo) > 0 {
		h.undo = h.undo[:len(h.undo)-1]
	}
}
func (h *History) CanUndo() bool { return len(h.undo) > 0 }
func (h *History) CanRedo() bool { return len(h.redo) > 0 }

// ---- 模拟数据集 ----

// demoModel 预置一个电商系统模型：11 张表、15 条关系，全部是内存模拟数据。
func demoModel() *Model {
	m := &Model{Name: "电商系统 · 示例模型", Dialect: DialectMySQL}

	addTable := func(name, comment, group string, cols []Column) *Table {
		id := m.nextID("t")
		m.Tables = append(m.Tables, Table{ID: id, Name: name, Comment: comment, Group: group})
		t := &m.Tables[len(m.Tables)-1]
		for i := range cols {
			cols[i].ID = m.nextID("c")
			t.Columns = append(t.Columns, cols[i])
		}
		return t
	}
	pk := func(comment string) Column {
		return Column{Name: "id", Type: "BIGINT", PK: true, NN: true, AI: true, Comment: comment}
	}
	ts := func(comment string) Column {
		return Column{Name: "created_at", Type: "DATETIME", NN: true, Default: "CURRENT_TIMESTAMP", Comment: comment}
	}

	users := addTable("users", "用户主表", "core", []Column{
		pk("用户 ID"),
		{Name: "username", Type: "VARCHAR", Length: "64", NN: true, UQ: true, Comment: "登录名"},
		{Name: "email", Type: "VARCHAR", Length: "128", UQ: true, Comment: "邮箱"},
		{Name: "password_hash", Type: "VARCHAR", Length: "128", NN: true, Comment: "密码哈希"},
		{Name: "phone", Type: "VARCHAR", Length: "20", Comment: "手机号"},
		{Name: "status", Type: "TINYINT", NN: true, Default: "1", Comment: "1 正常 0 禁用"},
		ts("注册时间"),
		{Name: "updated_at", Type: "DATETIME", Comment: "更新时间"},
	})
	addr := addTable("addresses", "收货地址", "core", []Column{
		pk("地址 ID"),
		{Name: "user_id", Type: "BIGINT", NN: true, Comment: "所属用户"},
		{Name: "receiver", Type: "VARCHAR", Length: "64", NN: true, Comment: "收货人"},
		{Name: "phone", Type: "VARCHAR", Length: "20", NN: true, Comment: "联系电话"},
		{Name: "province", Type: "VARCHAR", Length: "32", Comment: "省"},
		{Name: "city", Type: "VARCHAR", Length: "32", Comment: "市"},
		{Name: "detail", Type: "VARCHAR", Length: "255", Comment: "详细地址"},
		{Name: "is_default", Type: "BOOLEAN", NN: true, Default: "0", Comment: "默认地址"},
	})
	cats := addTable("categories", "商品分类（自引用）", "catalog", []Column{
		pk("分类 ID"),
		{Name: "parent_id", Type: "BIGINT", Comment: "父分类"},
		{Name: "name", Type: "VARCHAR", Length: "64", NN: true, Comment: "分类名"},
		{Name: "sort_order", Type: "INT", NN: true, Default: "0", Comment: "排序"},
		ts("创建时间"),
	})
	products := addTable("products", "商品 SPU", "catalog", []Column{
		pk("商品 ID"),
		{Name: "category_id", Type: "BIGINT", NN: true, Comment: "所属分类"},
		{Name: "name", Type: "VARCHAR", Length: "128", NN: true, Comment: "商品名"},
		{Name: "sku", Type: "VARCHAR", Length: "64", NN: true, UQ: true, Comment: "SKU 编码"},
		{Name: "price", Type: "DECIMAL", Length: "10,2", NN: true, Default: "0.00", Comment: "售价"},
		{Name: "stock", Type: "INT", NN: true, Default: "0", Comment: "库存"},
		{Name: "status", Type: "TINYINT", NN: true, Default: "1", Comment: "1 上架 0 下架"},
		{Name: "description", Type: "TEXT", Comment: "详情"},
		ts("上架时间"),
	})
	images := addTable("product_images", "商品图片", "catalog", []Column{
		pk("图片 ID"),
		{Name: "product_id", Type: "BIGINT", NN: true, Comment: "所属商品"},
		{Name: "url", Type: "VARCHAR", Length: "255", NN: true, Comment: "图片地址"},
		{Name: "sort_order", Type: "INT", NN: true, Default: "0", Comment: "排序"},
	})
	carts := addTable("cart_items", "购物车明细", "trade", []Column{
		pk("明细 ID"),
		{Name: "user_id", Type: "BIGINT", NN: true, Comment: "所属用户"},
		{Name: "product_id", Type: "BIGINT", NN: true, Comment: "商品"},
		{Name: "quantity", Type: "INT", NN: true, Default: "1", Comment: "数量"},
		{Name: "added_at", Type: "DATETIME", NN: true, Default: "CURRENT_TIMESTAMP", Comment: "加入时间"},
	})
	orders := addTable("orders", "订单主表", "trade", []Column{
		pk("订单 ID"),
		{Name: "order_no", Type: "VARCHAR", Length: "32", NN: true, UQ: true, Comment: "订单号"},
		{Name: "user_id", Type: "BIGINT", NN: true, Comment: "下单用户"},
		{Name: "address_id", Type: "BIGINT", Comment: "收货地址"},
		{Name: "total_amount", Type: "DECIMAL", Length: "10,2", NN: true, Default: "0.00", Comment: "订单总额"},
		{Name: "status", Type: "TINYINT", NN: true, Default: "0", Comment: "0 待付 1 已付 2 发货 3 完成"},
		{Name: "paid_at", Type: "DATETIME", Comment: "支付时间"},
		ts("下单时间"),
	})
	items := addTable("order_items", "订单明细", "trade", []Column{
		pk("明细 ID"),
		{Name: "order_id", Type: "BIGINT", NN: true, Comment: "所属订单"},
		{Name: "product_id", Type: "BIGINT", NN: true, Comment: "商品快照"},
		{Name: "product_name", Type: "VARCHAR", Length: "128", NN: true, Comment: "商品名快照"},
		{Name: "unit_price", Type: "DECIMAL", Length: "10,2", NN: true, Comment: "成交单价"},
		{Name: "quantity", Type: "INT", NN: true, Default: "1", Comment: "数量"},
		{Name: "subtotal", Type: "DECIMAL", Length: "10,2", NN: true, Comment: "小计"},
	})
	payments := addTable("payments", "支付流水", "trade", []Column{
		pk("流水 ID"),
		{Name: "order_id", Type: "BIGINT", NN: true, Comment: "关联订单"},
		{Name: "channel", Type: "VARCHAR", Length: "32", NN: true, Comment: "支付渠道"},
		{Name: "trade_no", Type: "VARCHAR", Length: "64", UQ: true, Comment: "第三方交易号"},
		{Name: "amount", Type: "DECIMAL", Length: "10,2", NN: true, Comment: "支付金额"},
		{Name: "status", Type: "TINYINT", NN: true, Default: "0", Comment: "0 待支付 1 成功 2 失败"},
		{Name: "paid_at", Type: "DATETIME", Comment: "回调时间"},
	})
	coupons := addTable("coupons", "优惠券模板", "promo", []Column{
		pk("券 ID"),
		{Name: "code", Type: "VARCHAR", Length: "32", NN: true, UQ: true, Comment: "券码"},
		{Name: "discount", Type: "DECIMAL", Length: "10,2", NN: true, Comment: "面额"},
		{Name: "min_amount", Type: "DECIMAL", Length: "10,2", NN: true, Default: "0.00", Comment: "使用门槛"},
		{Name: "starts_at", Type: "DATETIME", Comment: "生效时间"},
		{Name: "ends_at", Type: "DATETIME", Comment: "失效时间"},
	})
	userCoupons := addTable("user_coupons", "用户领券记录", "promo", []Column{
		pk("记录 ID"),
		{Name: "user_id", Type: "BIGINT", NN: true, Comment: "用户"},
		{Name: "coupon_id", Type: "BIGINT", NN: true, Comment: "优惠券"},
		{Name: "order_id", Type: "BIGINT", Comment: "使用的订单"},
		{Name: "used_at", Type: "DATETIME", Comment: "使用时间"},
	})

	link := func(child *Table, childCol string, parent *Table, parentCol string) {
		cc, pc := child.ColumnByName(childCol), parent.ColumnByName(parentCol)
		if cc == nil || pc == nil {
			return
		}
		_, _ = m.AddRelation(child.ID, cc.ID, parent.ID, pc.ID)
	}
	link(addr, "user_id", users, "id")
	link(cats, "parent_id", cats, "id")
	link(products, "category_id", cats, "id")
	link(images, "product_id", products, "id")
	link(carts, "user_id", users, "id")
	link(carts, "product_id", products, "id")
	link(orders, "user_id", users, "id")
	link(orders, "address_id", addr, "id")
	link(items, "order_id", orders, "id")
	link(items, "product_id", products, "id")
	link(payments, "order_id", orders, "id")
	link(userCoupons, "user_id", users, "id")
	link(userCoupons, "coupon_id", coupons, "id")
	link(userCoupons, "order_id", orders, "id")

	m.AutoLayout()
	return m
}
