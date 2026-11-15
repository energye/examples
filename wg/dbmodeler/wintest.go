package main

// --wintest：真实窗口端到端测试。
// 与 --uitest 不同，这里不直接调 Scene.Dispatch，而是把鼠标 / 键盘 / 滚轮
// 消息 PostMessage 给真实的 Form HWND，由 Windows → LCL → 自绘画布 → Scene
// 的完整链路处理，再用模型状态与真实窗口像素验证结果。

import (
	"fmt"
	"os"

	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
)

// 真实窗口消息测试用到的 Win32 消息号。
const (
	winWMMouseMove   = 0x0200
	winWMLButtonDown = 0x0201
	winWMLButtonUp   = 0x0202
	winWMKeyDown     = 0x0100
	winWMKeyUp       = 0x0101
	winWMChar        = 0x0102
	winWMMouseWheel  = 0x020A
)

func typesPoint(x, y int32) types.TPoint { return types.TPoint{X: x, Y: y} }

type winTestState struct {
	step   int
	wait   int
	frames int
	done   bool
	pass   int
	fails  []string

	baseTables int
	tableID    string
	x0, y0     float64
	zoom0      float64
	shot       string
}

var winTest *winTestState

func (w *winTestState) logf(format string, a ...interface{}) {
	fmt.Printf("[WINTEST] "+format+"\n", a...)
}

func (w *winTestState) check(ok bool, msg string) {
	if ok {
		w.pass++
		w.logf("PASS  %s", msg)
		return
	}
	w.fail(msg)
}

func (w *winTestState) fail(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	w.fails = append(w.fails, msg)
	w.logf("FAIL  %s", msg)
}

func (w *winTestState) next(s *Scene, step int) {
	w.step = step
	w.wait = 2
	if s.host != nil {
		s.host.Invalidate()
	}
}

func (w *winTestState) finish(s *Scene) {
	if w.done {
		return
	}
	w.done = true
	total := w.pass + len(w.fails)
	if len(w.fails) == 0 {
		w.logf("==== RESULT: %d/%d PASS · 真实窗口消息链路 ====", w.pass, total)
	} else {
		w.logf("==== RESULT: %d/%d PASS · 失败 %d 项 ====", w.pass, total, len(w.fails))
		for _, f := range w.fails {
			w.logf("  - %s", f)
		}
	}
	if w.shot != "" {
		if _, err := os.Stat(w.shot); err == nil {
			w.logf("真实窗口截图: %s", w.shot)
		}
	}
	lcl.RunOnMainThreadAsync(func(id uint32) {
		lcl.Application.Terminate()
	})
}

func winTestFailed() bool {
	return winTest != nil && len(winTest.fails) > 0
}

func (w *winTestState) tick(s *Scene) {
	if w == nil || w.done {
		return
	}
	w.frames++
	if w.frames > 900 {
		w.fail("超时：%d 帧仍未完成，step=%d", w.frames, w.step)
		w.finish(s)
		return
	}
	if w.wait > 0 {
		w.wait--
		s.host.Invalidate()
		return
	}

	hwnd := uintptr(s.form.Handle())
	switch w.step {
	case 0:
		if s.frame < 3 {
			w.wait = 1
			s.host.Invalidate()
			return
		}
		w.check(hwnd != 0, "拿到真实窗口句柄")
		w.check(s.client.Right >= 1200 && s.client.Bottom >= 700,
			fmt.Sprintf("真实客户区 %dx%d", s.client.Right, s.client.Bottom))
		w.baseTables = len(s.model.Tables)
		if b := s.buttons["newtable"]; b != nil {
			c := b.Rect().CenterPoint()
			w.logf("向 HWND 投递鼠标点击：工具栏「新建表」@%d,%d", c.X, c.Y)
			w.mouseClick(s, c.X, c.Y)
		} else {
			w.fail("找不到新建表按钮")
		}
		w.next(s, 1)

	case 1:
		w.check(len(s.model.Tables) == w.baseTables+1,
			fmt.Sprintf("真实点击后表数 %d → %d", w.baseTables, len(s.model.Tables)))
		t := s.model.TableByID(s.sel.TableID)
		if t == nil {
			w.fail("新建表后没有选中，后续拖动无法进行")
			w.finish(s)
			return
		}
		w.tableID = t.ID
		w.x0, w.y0 = t.X, t.Y
		hdr := s.view.headerRect(t)
		x, y := hdr.Left+40, hdr.Top+14
		w.logf("向 HWND 投递鼠标拖动：表头 %s @%d,%d → +64,+48", t.Name, x, y)
		w.mouseDown(s, x, y)
		for i := 1; i <= 4; i++ {
			w.mouseMove(s, x+int32(i)*16, y+int32(i)*12, true)
		}
		w.mouseUp(s, x+64, y+48)
		w.next(s, 2)

	case 2:
		t := s.model.TableByID(w.tableID)
		if t == nil {
			w.fail("拖动后表消失了")
			w.finish(s)
			return
		}
		w.check(t.X != w.x0 || t.Y != w.y0,
			fmt.Sprintf("真实拖动生效：%.0f,%.0f → %.0f,%.0f", w.x0, w.y0, t.X, t.Y))
		w.zoom0 = s.view.zoom
		c := s.viewRect.CenterPoint()
		w.logf("向 HWND 投递滚轮：画布中心 @%d,%d delta=+120", c.X, c.Y)
		w.wheel(s, c.X, c.Y, 120)
		w.next(s, 3)

	case 3:
		w.check(s.view.zoom > w.zoom0,
			fmt.Sprintf("真实滚轮缩放 %.2f → %.2f", w.zoom0, s.view.zoom))
		w.logf("向 HWND 投递按键：F1")
		w.key(s, 0x70) // VK_F1
		w.next(s, 4)

	case 4:
		w.check(s.modal != nil, "真实键盘事件打开了帮助弹窗")
		w.logf("向 HWND 投递按键：Esc")
		w.key(s, 0x1B) // VK_ESCAPE
		w.next(s, 5)

	case 5:
		w.check(s.modal == nil, "真实 Esc 关闭了弹窗")
		// 点搜索框，再投递真实 WM_CHAR，验证「系统字符 → 自绘输入框 → 过滤」。
		r := s.search.Rect()
		w.logf("点击搜索框 @%d,%d 并投递 WM_CHAR \"users\"", r.Left+20, r.Top+10)
		w.mouseClick(s, r.Left+20, r.Top+10)
		for _, ch := range "users" {
			winPostChar(hwnd, uint32(ch))
		}
		w.next(s, 6)

	case 6:
		w.check(s.filter == "users", fmt.Sprintf("真实字符输入进入搜索框：filter=%q", s.filter))
		w.check(len(s.navList.Items) == 1, fmt.Sprintf("搜索过滤后剩 %d 张表", len(s.navList.Items)))
		w.key(s, 0x1B) // Esc 还原
		w.next(s, 7)

	case 7:
		w.check(s.filter == "" && len(s.navList.Items) == len(s.model.Tables),
			fmt.Sprintf("Esc 还原过滤：filter=%q tables=%d", s.filter, len(s.navList.Items)))
		if w.shot != "" {
			if err := winCaptureClient(hwnd, w.shot); err != nil {
				w.fail("真实窗口截图失败: %v", err)
			} else {
				w.logf("已抓取真实窗口客户区像素 → %s", w.shot)
			}
		}
		w.finish(s)
	}
}

// ---- 消息投递封装（平台实现见 wintest_windows.go）----

func (w *winTestState) mouseDown(s *Scene, x, y int32) {
	winPostMouse(uintptr(s.form.Handle()), winWMMouseMove, x, y, false)
	winPostMouse(uintptr(s.form.Handle()), winWMLButtonDown, x, y, true)
}

func (w *winTestState) mouseMove(s *Scene, x, y int32, down bool) {
	winPostMouse(uintptr(s.form.Handle()), winWMMouseMove, x, y, down)
}

func (w *winTestState) mouseUp(s *Scene, x, y int32) {
	winPostMouse(uintptr(s.form.Handle()), winWMLButtonUp, x, y, false)
}

func (w *winTestState) mouseClick(s *Scene, x, y int32) {
	w.mouseDown(s, x, y)
	w.mouseUp(s, x, y)
}

func (w *winTestState) key(s *Scene, vk uint16) {
	winPostKey(uintptr(s.form.Handle()), vk)
}

func (w *winTestState) wheel(s *Scene, x, y int32, delta int16) {
	pt := s.form.ClientToScreenWithPoint(typesPoint(x, y))
	winPostWheel(uintptr(s.form.Handle()), pt.X, pt.Y, delta)
}
