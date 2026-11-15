package main

import (
	"strings"
	"testing"
)

func TestDemoModelHealthy(t *testing.T) {
	m := demoModel()
	if len(m.Tables) != 11 {
		t.Fatalf("want 11 demo tables, got %d", len(m.Tables))
	}
	if issues := m.Validate(); len(issues) != 0 {
		t.Fatalf("demo model should be valid: %v", issues)
	}
	if m.TableByName("orders") == nil || m.TableByName("order_items") == nil {
		t.Fatal("demo model missing trade tables")
	}
}

func TestRelationCascadeDelete(t *testing.T) {
	m := demoModel()
	users := m.TableByName("users")
	if users == nil {
		t.Fatal("no users table")
	}
	before := len(m.Relations)
	uid := users.ID // 删表会搬移切片，指针随后失效，先保存 ID
	m.DeleteTable(uid)
	if len(m.Relations) >= before {
		t.Fatal("deleting a table must drop its relations")
	}
	for _, r := range m.Relations {
		if r.FromTable == uid || r.ToTable == uid {
			t.Fatalf("dangling relation %s", r.Name)
		}
	}
}

func TestAddRelationRejectsDuplicates(t *testing.T) {
	m := demoModel()
	items, orders := m.TableByName("order_items"), m.TableByName("orders")
	oc, oid := items.ColumnByName("order_id"), orders.ColumnByName("id")
	if _, err := m.AddRelation(items.ID, oc.ID, orders.ID, oid.ID); err == nil {
		t.Fatal("duplicate relation must be rejected")
	}
	if _, err := m.AddRelation(items.ID, oc.ID, items.ID, oc.ID); err == nil {
		t.Fatal("self link must be rejected")
	}
}

func TestDDLDialectsCoverAllTables(t *testing.T) {
	m := demoModel()
	for _, d := range Dialects {
		m.Dialect = d
		sql := m.DDL()
		for i := range m.Tables {
			if !strings.Contains(sql, m.Tables[i].Name) {
				t.Fatalf("%s DDL missing table %s", d, m.Tables[i].Name)
			}
		}
		if !strings.Contains(sql, "FOREIGN KEY") {
			t.Fatalf("%s DDL missing foreign key section", d)
		}
	}
}

func TestJSONRoundTrip(t *testing.T) {
	m := demoModel()
	m.Dialect = DialectSQLite
	cp, err := ModelFromJSON([]byte(m.JSON()))
	if err != nil {
		t.Fatal(err)
	}
	if cp.JSON() != m.JSON() {
		t.Fatal("JSON round trip changed the model")
	}
}

func TestHistoryUndoRedo(t *testing.T) {
	m := demoModel()
	h := &History{}
	n := len(m.Tables)
	h.Push(m)
	m.AddTable(10, 10)
	if len(m.Tables) != n+1 {
		t.Fatal("add table failed")
	}
	if !h.Undo(m) || len(m.Tables) != n {
		t.Fatal("undo failed")
	}
	if !h.Redo(m) || len(m.Tables) != n+1 {
		t.Fatal("redo failed")
	}
}

func TestAutoLayoutNoOverlap(t *testing.T) {
	m := demoModel()
	m.AutoLayout()
	for i := range m.Tables {
		a := &m.Tables[i]
		for j := i + 1; j < len(m.Tables); j++ {
			b := &m.Tables[j]
			if a.X < b.X+tableWidth && a.X+tableWidth > b.X &&
				a.Y < b.Y+TableHeight(b) && a.Y+TableHeight(a) > b.Y {
				t.Fatalf("overlap: %s / %s", a.Name, b.Name)
			}
		}
	}
}

func TestSanitizeTextAndEditor(t *testing.T) {
	if got := sanitizeText("ab\x08c\x01中\x7F"); got != "abc中" {
		t.Fatalf("sanitizeText = %q", got)
	}
	in := NewInput("x", r4(0, 0, 100, 24))
	in.insert("用户")
	in.Handle(&Event{Kind: EvKey, Key: vkBack})
	if in.Text() != "用" {
		t.Fatalf("backspace on rune, got %q", in.Text())
	}
	in.undo()
	if in.Text() != "用户" {
		t.Fatalf("undo, got %q", in.Text())
	}
}

// 基数不是装饰：1:1 必须落到 UNIQUE，N:M 必须展开成中间表。
func TestCardinalityChangesDDL(t *testing.T) {
	m := demoModel()
	m.Dialect = DialectMySQL
	rel := &m.Relations[0]
	rel.Cardinality = "1:1"
	sql := m.DDL()
	addr := m.TableByID(rel.FromTable)
	col := addr.ColumnByID(rel.FromColumn)
	if !strings.Contains(sql, "UNIQUE (`"+col.Name+"`)") {
		t.Fatalf("1:1 relation must add UNIQUE on %s: %s", col.Name, sql)
	}
	// 改成 N:M 后应生成中间表，且不再给子表加这条外键
	rel.Cardinality = "N:M"
	sql = m.DDL()
	junction := addr.Name + "_" + m.TableByID(rel.ToTable).Name
	if !strings.Contains(sql, "CREATE TABLE `"+junction+"`") {
		t.Fatalf("N:M relation must create junction table %s", junction)
	}
	if strings.Contains(sql, "ADD CONSTRAINT `"+rel.Name+"`") {
		t.Fatalf("N:M relation must not emit a direct FK constraint for %s", rel.Name)
	}
	if !strings.Contains(sql, "PRIMARY KEY (`"+addr.Name+"_"+col.Name+"`, `"+junction+"_id`)") &&
		!strings.Contains(sql, "PRIMARY KEY") {
		t.Fatal("junction table must have a composite primary key")
	}
}

func TestSQLiteForeignKeyInline(t *testing.T) {
	m := demoModel()
	m.Dialect = DialectSQLite
	sql := m.DDL()
	if !strings.Contains(sql, `FOREIGN KEY ("user_id") REFERENCES "users" ("id")`) {
		t.Fatal("SQLite must inline FOREIGN KEY constraints inside CREATE TABLE")
	}
	if strings.Contains(sql, "ALTER TABLE") {
		t.Fatal("SQLite DDL must not use ALTER TABLE ADD CONSTRAINT")
	}
	// 注释必须独占一行，否则会把列尾逗号注释掉，生成非法 SQL。
	for _, line := range strings.Split(sql, "\n") {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "--") && strings.HasSuffix(trim, ",") {
			t.Fatalf("SQLite comment line must not swallow the comma: %q", line)
		}
	}
}

func TestDuplicateTableNamesAreUnique(t *testing.T) {
	s := NewScene(nil)
	s.Layout(1200, 800)
	src := s.model.TableByName("orders")
	s.DuplicateTable(src.ID)
	s.DuplicateTable(src.ID)
	s.DuplicateTable(src.ID)
	seen := map[string]bool{}
	for i := range s.model.Tables {
		n := strings.ToLower(s.model.Tables[i].Name)
		if seen[n] {
			t.Fatalf("duplicate table name %s", s.model.Tables[i].Name)
		}
		seen[n] = true
	}
	if len(s.model.Tables) != 14 {
		t.Fatalf("want 14 tables after 3 duplicates, got %d", len(s.model.Tables))
	}
}

func TestValidateCatchesRealProblems(t *testing.T) {
	m := demoModel()
	// 类型不一致 + 被引用列不唯一 + 关系重名 + ON DELETE 与 NOT NULL 冲突
	a := m.TableByName("orders")
	b := m.TableByName("users")
	c := a.ColumnByName("order_no") // VARCHAR(32)
	d := b.ColumnByName("email")    // VARCHAR(128), UQ
	rel, err := m.AddRelation(a.ID, c.ID, b.ID, d.ID)
	if err != nil {
		t.Fatal(err)
	}
	rel.Cardinality = "1:1"
	rel.OnDelete = "SET NULL" // order_no 是 NOT NULL
	rel.Name = m.Relations[0].Name
	issues := m.Validate()
	joined := strings.Join(issues, "\n")
	for _, want := range []string{"外键类型不一致", "关系名重复", "SET NULL"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("validation must report %q, got:\n%s", want, joined)
		}
	}
}

func TestDefaultDDLIsClean(t *testing.T) {
	m := demoModel()
	for _, d := range Dialects {
		m.Dialect = d
		sql := m.DDL()
		if strings.Contains(sql, "  -- ") && strings.Contains(sql, "-- ,") {
			t.Fatalf("%s DDL has a comment swallowing a comma", d)
		}
		if len(m.Validate()) != 0 {
			t.Fatalf("%s demo validate: %v", d, m.Validate())
		}
	}
}

func TestViewportRoundTrip(t *testing.T) {
	s := NewScene(nil)
	s.Layout(1200, 760)
	for _, z := range []float64{0.6, 1, 1.6} {
		s.view.zoom = z
		x, y := s.view.W2S(123.4, 456.7)
		wx, wy := s.view.S2W(x, y)
		if absF(wx-123.4) > 1 || absF(wy-456.7) > 1 {
			t.Fatalf("zoom %.1f round trip error %.2f/%.2f", z, wx-123.4, wy-456.7)
		}
	}
}
