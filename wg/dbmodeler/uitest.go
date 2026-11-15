package main

// --uitest：无窗口的「事件链路」回归。
// 直接把鼠标 / 滚轮 / 键盘事件投给 Scene.Dispatch，走与真实窗口完全相同的
// 命中测试 → 焦点 → 组件 → 画布 → 命令链路，断言建模操作确实生效。
// 不依赖 HWND，也不依赖截图，可在 CI 里跑。

import (
	"fmt"
	"os"

	"github.com/energye/lcl/types"
)

type uiCheck struct {
	name string
	fn   func(*Scene) (bool, string)
}

func uiChecks() []uiCheck {
	return []uiCheck{
		{"点击表头选中表", uiSelectTable},
		{"拖动表头移动 + 撤销", uiDragMove},
		{"双击就地改名（含中文上屏）", uiInlineRename},
		{"拖端口创建外键关系", uiLinkRelation},
		{"滚轮缩放 / 右键菜单 / Esc", uiWheelMenu},
		{"属性面板输入框提交", uiInspectorCommit},
		{"删除表与撤销恢复", uiDeleteUndo},
		{"弹窗与快捷键", uiModal},
		{"工具栏全部按钮", uiToolbarButtons},
		{"右键菜单全部条目", uiCanvasMenus},
		{"快捷键全覆盖", uiShortcuts},
		{"输入框真实编辑", uiInputRealEdit},
		{"列表/弹窗/属性滚动条", uiScrollbars},
		{"基数影响 DDL", uiCardinalityControl},
		{"字段属性面板全项", uiColumnProps},
		{"多选与对齐", uiMultiSelectAlign},
		{"保存/打开/导出落盘", uiSaveOpenExport},
	}
}

func runUIChecks(verbose bool) (int, int) {
	if verbose {
		fmt.Println("==== DBMODELER UITEST ====")
	}
	dir, err := os.MkdirTemp("", "dbmodeler-uitest")
	if err == nil {
		defer os.RemoveAll(dir)
	}
	pass, fail := 0, 0
	for _, c := range uiChecks() {
		// 每项都用全新场景，避免前一项的状态干扰后一项。
		sc := NewScene(nil)
		sc.dataDir = dir // 保存 / 导出写临时目录，不污染包目录
		sc.Layout(1440, 880)
		ok, detail := c.fn(sc)
		if ok {
			pass++
		} else {
			fail++
		}
		if verbose {
			state := "PASS"
			if !ok {
				state = "FAIL"
			}
			fmt.Printf("  %-30s %s  %s\n", c.name, state, detail)
		}
	}
	if verbose {
		fmt.Printf("==== RESULT: %d/%d PASS ====\n", pass, pass+fail)
	}
	return pass, fail
}

func runUITest() int {
	_, fail := runUIChecks(true)
	if fail > 0 {
		return 1
	}
	return 0
}

func down(s *Scene, x, y int32, button types.TMouseButton, shift types.TShiftState) {
	s.Dispatch(&Event{Kind: EvDown, X: x, Y: y, Button: button, Shift: shift})
}
func move(s *Scene, x, y int32) {
	s.Dispatch(&Event{Kind: EvMove, X: x, Y: y})
}
func up(s *Scene, x, y int32) {
	s.Dispatch(&Event{Kind: EvUp, X: x, Y: y, Button: types.MbLeft})
}
func click(s *Scene, x, y int32) {
	down(s, x, y, types.MbLeft, 0)
	up(s, x, y)
}

func uiSelectTable(s *Scene) (bool, string) {
	t := s.model.TableByName("users")
	hdr := s.view.headerRect(t)
	click(s, hdr.Left+40, hdr.Top+10)
	if s.sel.Kind != "table" || s.sel.TableID != t.ID {
		return false, fmt.Sprintf("sel=%s/%s", s.sel.Kind, s.sel.TableID)
	}
	if len(s.inspWidgets) == 0 {
		return false, "属性面板未随选择刷新"
	}
	return true, "选中 " + t.Name + "，属性面板已重建"
}

func uiDragMove(s *Scene) (bool, string) {
	t := s.model.TableByName("users")
	tid := t.ID
	hdr := s.view.headerRect(t)
	oldX, oldY := t.X, t.Y
	down(s, hdr.Left+40, hdr.Top+10, types.MbLeft, 0)
	move(s, hdr.Left+72, hdr.Top+34)
	up(s, hdr.Left+72, hdr.Top+34)
	if t.X <= oldX || t.Y <= oldY {
		return false, fmt.Sprintf("坐标未变化 %.0f,%.0f -> %.0f,%.0f", oldX, oldY, t.X, t.Y)
	}
	if !s.history.CanUndo() {
		return false, "拖动未进入撤销栈"
	}
	s.Undo()
	// Undo 会整体替换 Model 的切片（Restore），旧元素指针随即失效，按 ID 重查。
	cur := s.model.TableByID(tid)
	if cur == nil || cur.X != oldX || cur.Y != oldY {
		return false, "撤销未恢复表位置"
	}
	return true, "移动 32×24 像素并成功撤销"
}

func uiInlineRename(s *Scene) (bool, string) {
	t := s.model.TableByName("users")
	hdr := s.view.headerRect(t)
	// 两次快速按下 = 自绘双击检测
	down(s, hdr.Left+40, hdr.Top+10, types.MbLeft, 0)
	up(s, hdr.Left+40, hdr.Top+10)
	down(s, hdr.Left+40, hdr.Top+10, types.MbLeft, 0)
	up(s, hdr.Left+40, hdr.Top+10)
	if s.inline == nil || !s.inline.Visible() {
		return false, "双击未打开就地编辑框"
	}
	if s.focus != s.inline {
		return false, "就地编辑框未获得键盘焦点"
	}
	s.OnText("会员") // 已全选，直接替换
	s.OnKey(vkReturn, 0)
	if s.inline != nil {
		return false, "提交后编辑框未收起"
	}
	if s.model.TableByName("会员") == nil {
		return false, "中文表名未写回模型"
	}
	if !s.history.CanUndo() {
		return false, "改名未进入撤销栈"
	}
	return true, "双击 → 中文上屏 → Enter 提交 → 模型更新"
}

func uiLinkRelation(s *Scene) (bool, string) {
	s.view.Fit() // 保证两端端口都落在画布可见区域
	child := s.model.TableByName("product_images")
	parent := s.model.TableByName("categories")
	ci := s.view.columnIndex(child, child.ColumnByName("url").ID)
	pi := s.view.columnIndex(parent, parent.ColumnByName("id").ID)
	px, py := s.view.portPos(child, ci, 1)
	qx, qy := s.view.portPos(parent, pi, -1)
	before := len(s.model.Relations)
	down(s, px, py, types.MbLeft, 0)
	move(s, qx, qy)
	up(s, qx, qy)
	if len(s.model.Relations) != before+1 {
		return false, fmt.Sprintf("关系数 %d -> %d", before, len(s.model.Relations))
	}
	rel := s.model.Relations[len(s.model.Relations)-1]
	if rel.FromTable != child.ID || rel.ToTable != parent.ID || rel.Cardinality != "1:N" {
		return false, "新关系端点或基数不正确"
	}
	if s.sel.Kind != "relation" {
		return false, "创建后未选中新关系"
	}
	return true, "从 " + child.Name + ".url 拖到 " + parent.Name + ".id 建立外键"
}

func uiWheelMenu(s *Scene) (bool, string) {
	c := s.viewRect.CenterPoint()
	z0 := s.view.zoom
	s.Dispatch(&Event{Kind: EvWheel, X: c.X, Y: c.Y, Wheel: 120})
	if s.view.zoom <= z0 {
		return false, "滚轮未放大"
	}
	down(s, c.X, c.Y, types.MbRight, 0)
	if s.popup == nil || len(s.popup.Items) == 0 {
		return false, "右键未打开自绘菜单"
	}
	n := len(s.popup.Items)
	s.OnKey(vkEscape, 0)
	if s.popup != nil {
		return false, "Esc 未关闭菜单"
	}
	return true, fmt.Sprintf("缩放 %.2f→%.2f，右键 %d 项菜单，Esc 关闭", z0, s.view.zoom, n)
}

func uiInspectorCommit(s *Scene) (bool, string) {
	t := s.model.TableByName("products")
	hdr := s.view.headerRect(t)
	click(s, hdr.Left+40, hdr.Top+10)
	var nameInput *Input
	for _, w := range s.inspWidgets {
		if in, ok := w.(*Input); ok {
			nameInput = in
			break
		}
	}
	if nameInput == nil {
		return false, "属性面板没有名称输入框"
	}
	s.SetFocus(nameInput)
	nameInput.SetText("商品主表_v2")
	s.SetFocus(nil) // 失焦提交
	if t.Name != "商品主表_v2" {
		return false, "输入未写回模型，当前值 " + t.Name
	}
	if s.model.TableByName("商品主表_v2") == nil {
		return false, "模型索引未更新"
	}
	return true, "属性面板输入 → 失焦提交 → 表名更新"
}

func uiDeleteUndo(s *Scene) (bool, string) {
	t := s.model.TableByName("orders")
	hdr := s.view.headerRect(t)
	click(s, hdr.Left+40, hdr.Top+10)
	n := len(s.model.Tables)
	relBefore := len(s.model.Relations)
	s.DeleteSelection()
	if len(s.model.Tables) != n-1 {
		return false, "删除表失败"
	}
	if len(s.model.Relations) >= relBefore {
		return false, "关系未级联删除"
	}
	s.Undo()
	if len(s.model.Tables) != n || s.model.TableByName("orders") == nil {
		return false, "撤销未恢复表"
	}
	return true, "删除表 + 级联删关系 + 撤销完整恢复"
}

func uiModal(s *Scene) (bool, string) {
	s.Help()
	if s.modal == nil || len(s.modal.Lines) < 10 {
		return false, "帮助弹窗未打开"
	}
	lines := len(s.modal.Lines)
	s.OnKey(vkEscape, 0)
	if s.modal != nil {
		return false, "Esc 未关闭弹窗"
	}
	s.ExportDDL()
	if s.modal == nil {
		return false, "DDL 预览弹窗未打开"
	}
	if len(s.modal.Buttons) != 2 {
		return false, "DDL 弹窗按钮数量异常"
	}
	s.CloseModal()
	return true, fmt.Sprintf("帮助 %d 行 / Esc 关闭 / DDL 弹窗按钮正常", lines)
}
