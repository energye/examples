// DB Modeler · 单画布纯自绘数据库建模示例
//
// 参考 wg/canvas_ime 与 wg/showcase 的架构：Form 是唯一持有 HWND 的对象，
// 一张无句柄 ICustomGraphicControl 是唯一画布，所有组件（含可输入中文的
// 输入框、菜单、弹窗）都是画在这张画布上的纯 Go 对象。
// 数据全部在内存里，预置电商系统模拟数据；不建立任何数据库连接。
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/energye/lcl/api/libname"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
	"github.com/energye/lcl/types/messages"
)

type TMainForm struct {
	lcl.TEngForm
	scene *Scene
}

var (
	mainForm   TMainForm
	smokeMode  bool
	exitCode   = -1
	screenshot string
)

func main() {
	setupLib()
	args := os.Args[1:]
	if len(args) > 0 {
		switch args[0] {
		case "--selftest":
			os.Exit(runSelfTest())
		case "--uitest":
			os.Exit(runUITest())
		case "--help", "-h":
			printCLIHelp()
			return
		case "--ddl":
			printDDL()
			return
		case "--smoke":
			smokeMode = true
		case "--wintest":
			if !winTestSupported() {
				fmt.Println("[WINTEST] FAIL: 当前平台不支持真实窗口消息测试")
				os.Exit(1)
			}
			winTest = &winTestState{shot: "wintest_real_window.png"}
			if len(args) > 1 {
				winTest.shot = args[1]
			}
		case "--screenshot":
			screenshot = "dbmodeler.png"
			if len(args) > 1 {
				screenshot = args[1]
			}
			lcl.Init()
			if err := headlessScreenshot(screenshot, 1440, 880); err != nil {
				fmt.Println("[SCREENSHOT] FAIL:", err)
				os.Exit(1)
			}
			fmt.Println("[SCREENSHOT] OK:", screenshot, "· 1440x880 离屏整帧渲染")
			return
		}
	}

	lcl.Init()
	lcl.Application.Initialize()
	lcl.Application.SetMainFormOnTaskBar(true)
	lcl.Application.NewForm(&mainForm)
	lcl.Application.Run()

	if winTest != nil {
		if winTest.done && len(winTest.fails) == 0 {
			exitCode = 0
		} else {
			exitCode = 1
		}
	}
	if smokeMode {
		s := mainForm.scene
		if s != nil && s.frame >= 3 {
			fmt.Printf("[SMOKE] PASS frames=%d client=%dx%d tables=%d columns=%d relations=%d\n",
				s.frame, s.client.Right, s.client.Bottom, len(s.model.Tables),
				s.countColumns(), len(s.model.Relations))
		} else {
			fmt.Println("[SMOKE] FAIL: 未达到 3 帧 OnPaint")
			exitCode = 1
		}
	}
	if exitCode >= 0 {
		os.Exit(exitCode)
	}
}

func (m *TMainForm) FormCreate(sender lcl.IObject) {
	m.SetCaption("DB Modeler · 单画布纯自绘数据库建模（模拟数据）")
	m.SetPosition(types.PoScreenCenter)
	m.SetWidth(1440)
	m.SetHeight(880)
	m.SetDoubleBuffered(true)
	m.SetKeyPreview(true)

	s := NewScene(m)
	m.scene = s
	m.bindInput(s)
}

// bindInput 绑定窗口级输入：WndProc（IME / 滚轮）、UTF-8 上屏字符、虚拟键。
// 鼠标与绘制由 Scene 绑在唯一画布上。
func (m *TMainForm) bindInput(s *Scene) {
	m.SetOnWndProc(func(msg *types.TLMessage) {
		hwnd := uintptr(m.Handle())
		reposition := false
		switch msg.Msg {
		case messages.WM_IME_STARTCOMPOSITION:
			reposition = true
		case messages.WM_IME_COMPOSITION:
			if text, cursor, ok := imeComposition(hwnd); ok {
				s.SetComposition(text, cursor)
			}
			reposition = true
		case messages.WM_IME_ENDCOMPOSITION:
			s.SetComposition("", 0)
		case messages.WM_IME_SETCONTEXT:
			reposition = true
		}
		// TEngForm 完全接管消息：未处理消息必须透传，否则窗口停止响应。
		m.InheritedWndProc(msg)
		if reposition {
			s.PositionIME()
		}
	})

	// UTF8KeyPress 不是「可打印字符」事件：Backspace / Ctrl+A / Enter / Tab
	// 都会被 TranslateMessage 变成控制字符送进来。必须在事件边界过滤。
	m.SetOnUTF8KeyPress(func(sender lcl.IObject, text *string) {
		if text == nil {
			return
		}
		raw := *text
		clean := sanitizeText(raw)
		if clean != raw {
			*text = clean // 同时阻止 LCL 继续转发控制字符
		}
		if clean == "" {
			return
		}
		s.OnText(clean)
	})

	m.SetOnKeyDown(func(sender lcl.IObject, key *uint16, shift types.TShiftState) {
		if key == nil {
			return
		}
		s.OnKey(*key, shift)
	})
}

func setupLib() {
	if p := os.Getenv("ENERGY_LIB"); fileExists(p) {
		libname.LibName = p
		return
	}
	home, _ := os.UserHomeDir()
	for _, p := range []string{
		"libenergy-amd64.dll",
		filepath.Join(home, ".energy", "runtime", "libenergy-amd64.dll"),
		"E:/app/workspace/lib/libenergy-x86_64-win64.dll",
	} {
		if fileExists(p) {
			libname.LibName = p
			return
		}
	}
}

func fileExists(p string) bool {
	if p == "" {
		return false
	}
	_, err := os.Stat(p)
	return err == nil
}

func printCLIHelp() {
	fmt.Println(`DB Modeler · 单画布纯自绘数据库建模示例（不连接数据库）

用法：
  go run ./wg/dbmodeler                     打开建模窗口
  go run ./wg/dbmodeler --selftest         无窗口纯逻辑自检（13 项）
  go run ./wg/dbmodeler --uitest           无窗口逐按钮/菜单/快捷键断言（17 项）
  go run ./wg/dbmodeler --smoke            真实窗口画 3 帧后自动退出
  go run ./wg/dbmodeler --wintest out.png  真实窗口注入鼠标/键盘/滚轮消息并断言（10 项）
  go run ./wg/dbmodeler --screenshot out.png  离屏渲染整帧并导出 PNG
  go run ./wg/dbmodeler --ddl              打印三种方言的模拟 DDL
  go test ./wg/dbmodeler                    Go 单测

窗口内快捷键：Ctrl+N 新建表 / Ctrl+Z,Y 撤销重做 / Ctrl+S,O 保存打开
  Ctrl+E 导出 DDL / Ctrl+0 适应 / F1 帮助 / Delete 删除 / F2 改名
  拖表头移动，拖字段圆点建关系，滚轮缩放，双击改名，右键菜单。`)
}

func printDDL() {
	m := demoModel()
	for _, d := range Dialects {
		m.Dialect = d
		fmt.Println("==== " + d + " ====")
		fmt.Println(m.DDL())
	}
}
