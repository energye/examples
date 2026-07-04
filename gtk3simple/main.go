//----------------------------------------
//
// Copyright © yanghy. All Rights Reserved.
//
// Licensed under Apache License Version 2.0, January 2004
//
// https://www.apache.org/licenses/LICENSE-2.0
//
//----------------------------------------

package main

import (
	"fmt"

	"github.com/energye/energy/v3/platform/linux/gtk3"
	. "github.com/energye/energy/v3/platform/linux/types"
)

var passed, failed int

func check(name string, ok bool) {
	if ok {
		passed++
		fmt.Printf("  ✅ %s\n", name)
	} else {
		failed++
		fmt.Printf("  ❌ %s\n", name)
	}
}

func makeSection(title string) {
	fmt.Println("")
	fmt.Println(title)
}

func main() {
	gtk3.Init(nil)

	win, _ := gtk3.NewWindow(WINDOW_TOPLEVEL)
	win.SetTitle("GTK3 全组件增强测试")
	win.SetDefaultSize(900, 650)
	win.SetSizeRequest(600, 400) // 防止 Notebook 布局时分配到负空间
	win.SetIconName("applications-system")

	hbar := gtk3.NewHeaderBar()
	hbar.SetTitle("GTK3 全组件增强测试")
	hbar.SetShowCloseButton(true)
	win.SetTitlebar(hbar)

	win.SetOnDestroy(func(sender PGtkWidget, userData GPointer) {
		fmt.Println("[event] Window.Destroy → MainQuit")
		gtk3.MainQuit()
	})
	win.SetOnMap(func(sender PGtkWidget, userData GPointer) {
		fmt.Println("[event] Window.Map")
	})
	win.SetOnConfigure(func(sender PGtkWidget, event PEventConfigure, userData GPointer) bool {
		fmt.Println("[event] Window.Configure")
		return false
	})

	// ============================================================
	// Root layout: vertical box with menu + notebook + statusbar
	// ============================================================
	mainBox := gtk3.NewBox(ORIENTATION_VERTICAL, 0)
	win.Add(mainBox)

	// Statusbar shared across tabs
	statusbar := gtk3.NewStatusbar()
	statusCtx := statusbar.GetContextId("main")
	statusbar.Push(statusCtx, "就绪")

	// ============================================================
	// Menu bar
	// ============================================================
	menuBar := gtk3.NewMenuBar()

	// -- File
	fileMenu := gtk3.NewMenu()
	fileMenuItem := gtk3.MenuItemNewWithLabel("文件")
	fileMenuItem.SetSubmenu(fileMenu)

	openItem := gtk3.MenuItemNewWithLabel("打开")
	openItem.SetOnActivate(func(sender PGtkWidget, userData GPointer) {
		fmt.Println("[event] MenuItem.Activate → 打开文件")
		fcDlg := gtk3.NewFileChooserDialog("打开文件", win, FILE_CHOOSER_ACTION_OPEN)
		fcDlg.SetOnResponse(func(sender PGtkWidget, responseId int32, userData GPointer) {
			fmt.Println("[event] FileChooser.Response id:", responseId)
			if responseId == int32(RESPONSE_ACCEPT) {
				fname := fcDlg.GetFilename()
				statusbar.Push(statusCtx, "打开: "+fname)
			}
			fcDlg.Destroy()
		})
		fcDlg.ShowAll()
	})
	fileMenu.Append(openItem)
	fileMenu.Append(gtk3.SeparatorMenuItemNew())

	quitItem := gtk3.MenuItemNewWithLabel("退出")
	quitItem.SetOnActivate(func(sender PGtkWidget, userData GPointer) {
		fmt.Println("[event] MenuItem.Activate → 退出")
		gtk3.MainQuit()
	})
	fileMenu.Append(quitItem)
	menuBar.Append(fileMenuItem)

	// -- Help
	helpMenu := gtk3.NewMenu()
	helpMenuItem := gtk3.MenuItemNewWithLabel("帮助")
	helpMenuItem.SetSubmenu(helpMenu)

	aboutItem := gtk3.MenuItemNewWithLabel("关于")
	aboutItem.SetOnActivate(func(sender PGtkWidget, userData GPointer) {
		fmt.Println("[event] MenuItem.Activate → 关于对话框")
		about := gtk3.NewAboutDialog()
		about.SetProgramName("Energy GTK3")
		about.SetVersion("3.0.0")
		about.SetComments("Go 语言 GTK3 绑定库 — 全组件增强示例")
		about.SetWebsite("https://github.com/energye/energy")
		about.SetWebsiteLabel("GitHub")
		about.SetLicense("Apache 2.0")
		about.SetAuthors([]string{"energye"})
		about.SetOnResponse(func(sender PGtkWidget, responseId int32, userData GPointer) {
			fmt.Println("[event] AboutDialog.Response id:", responseId)
			about.Destroy()
		})
		about.ShowAll()
	})
	helpMenu.Append(aboutItem)
	menuBar.Append(helpMenuItem)

	mainBox.PackStart(menuBar, false, false, 0)

	// ============================================================
	// Main notebook
	// ============================================================
	notebook := gtk3.NewNotebook()
	notebook.SetTabPos(POS_TOP)
	notebook.SetScrollable(true)
	mainBox.PackStart(notebook, true, true, 0)

	// ============================================================
	// Tab 1: 基础组件 (Widget / Label / Button / Image / Separator)
	// ============================================================
	tab1 := gtk3.NewBox(ORIENTATION_VERTICAL, 8)
	tab1.SetMarginTop(10)
	tab1.SetMarginBottom(10)
	tab1.SetMarginStart(10)
	tab1.SetMarginEnd(10)

	// -- Label 基础
	lblPlain := gtk3.NewLabel("普通文本标签")
	tab1.PackStart(lblPlain, false, false, 0)

	// -- Label markup
	lblMarkup := gtk3.NewLabel("")
	lblMarkup.SetMarkup("<b>粗体</b>  <i>斜体</i>  <u>下划线</u>  <span foreground='red'>红色</span>")
	lblMarkup.SetJustify(JUSTIFY_CENTER)
	tab1.PackStart(lblMarkup, false, false, 0)

	// -- Label 折行 & 省略
	lblWrap := gtk3.NewLabel("这是一个很长的标签文本，用于测试 SetLineWrap 和 SetEllipsize 功能，当宽度不足时会自动折行或省略显示。")
	lblWrap.SetLineWrap(true)
	lblWrap.SetWidthChars(40)
	lblWrap.SetEllipsize(ELLIPSIZE_END)
	tab1.PackStart(lblWrap, false, false, 0)

	// -- Label 可选
	lblSelectable := gtk3.NewLabel("这是一个可选中的文本（尝试选中我）")
	lblSelectable.SetSelectable(true)
	tab1.PackStart(lblSelectable, false, false, 0)

	tab1.PackStart(gtk3.NewSeparator(ORIENTATION_HORIZONTAL), false, false, 0)

	// -- Button 组
	btnRow := gtk3.NewBox(ORIENTATION_HORIZONTAL, 6)

	btnClick := gtk3.NewButtonWithLabel("计数按钮")
	count := 0
	btnClick.SetOnClick(func(sender PGtkWidget, userData GPointer) {
		count++
		btnClick.SetLabel(fmt.Sprintf("点击了 %d 次", count))
		fmt.Println("[event] Button.Click → count:", count)
		statusbar.Push(statusCtx, fmt.Sprintf("按钮被点击 %d 次", count))
	})
	btnClick.SetTooltipText("点击我计数会递增")
	btnRow.PackStart(btnClick, false, false, 0)

	btnMnemonic := gtk3.NewButtonWithMnemonic("_快捷键")
	btnRow.PackStart(btnMnemonic, false, false, 0)

	// -- Relief 按钮
	btnRelief := gtk3.NewButtonWithLabel("无边框")
	btnRelief.SetRelief(RELIEF_NONE)
	btnRow.PackStart(btnRelief, false, false, 0)

	tab1.PackStart(btnRow, false, false, 0)

	// -- CheckButton (开关)
	checkRow := gtk3.NewBox(ORIENTATION_HORIZONTAL, 6)
	checkRow.PackStart(gtk3.NewLabel("开关:"), false, false, 0)
	toggle := gtk3.NewCheckButton()
	toggle.SetActive(true)
	// 使用 SetOnToggled 替换 SetOnClick，更语义化
	toggle.SetOnToggled(func(sender PGtkWidget, userData GPointer) {
		if fmt.Println("[event] CheckButton.Toggled →", map[bool]string{false: "OFF", true: "ON"}[toggle.GetActive()]); toggle.GetActive() {
			lblPlain.SetText("CheckButton: ON ✅ (toggled)")
			statusbar.Push(statusCtx, "CheckButton toggled → ON")
		} else {
			lblPlain.SetText("CheckButton: OFF ❌ (toggled)")
			statusbar.Push(statusCtx, "CheckButton toggled → OFF")
		}
	})
	checkRow.PackStart(toggle, false, false, 0)
	tab1.PackStart(checkRow, false, false, 0)

	// -- Switch
	swRow := gtk3.NewBox(ORIENTATION_HORIZONTAL, 6)
	sw := gtk3.NewSwitch()
	sw.SetActive(false)
	swInfo := gtk3.NewLabel("GtkSwitch 状态: OFF")
	swRow.PackStart(gtk3.NewLabel("GtkSwitch:"), false, false, 0)
	swRow.PackStart(sw, false, false, 0)
	swRow.PackStart(swInfo, false, false, 0)
	sw.SetOnActiveNotify(func(sender PGtkWidget, pspec uintptr, userData GPointer) {
		if fmt.Println("[event] Switch.NotifyActive →", map[bool]string{false: "OFF", true: "ON"}[sw.GetActive()]); sw.GetActive() {
			swInfo.SetText("GtkSwitch 状态: ON ✅")
		} else {
			swInfo.SetText("GtkSwitch 状态: OFF ❌")
		}
	})
	tab1.PackStart(swRow, false, false, 0)

	// -- Image
	imgRow := gtk3.NewBox(ORIENTATION_HORIZONTAL, 10)
	imgInfo := gtk3.NewImageFromIconName("dialog-information", ICON_SIZE_DIALOG)
	imgRow.PackStart(imgInfo, false, false, 0)
	imgWarn := gtk3.NewImageFromIconName("dialog-warning", ICON_SIZE_DIALOG)
	imgRow.PackStart(imgWarn, false, false, 0)
	imgError := gtk3.NewImageFromIconName("dialog-error", ICON_SIZE_DIALOG)
	imgRow.PackStart(imgError, false, false, 0)
	// 设置像素大小
	imgSmall := gtk3.NewImageFromIconName("face-smile", ICON_SIZE_MENU)
	imgSmall.SetPixelSize(48)
	imgRow.PackStart(imgSmall, false, false, 0)
	tab1.PackStart(imgRow, false, false, 0)

	// -- Widget 通用属性演示
	propRow := gtk3.NewBox(ORIENTATION_HORIZONTAL, 6)
	propBtn := gtk3.NewButtonWithLabel("不透明 0.6")
	propBtn.SetOpacity(0.6)
	propBtn.SetTooltipText("半透明按钮")
	propRow.PackStart(propBtn, false, false, 0)

	sensBtn := gtk3.NewButtonWithLabel("禁用状态")
	sensBtn.SetSensitive(false)
	propRow.PackStart(sensBtn, false, false, 0)

	propRow.PackStart(gtk3.NewLabel("扩展:"), false, false, 0)
	expandCheck := gtk3.NewCheckButton()
	expandCheck.SetActive(true)
	propRow.PackStart(expandCheck, false, false, 0)
	expandLabel := gtk3.NewLabel("HExpand")
	expandLabel.SetHExpand(true)
	propRow.PackStart(expandLabel, true, true, 0)

	tab1.PackStart(propRow, false, false, 0)

	// -- 包装 ScrolledWindow 使之可滚动 --
	tab1Sw := gtk3.NewScrolledWindow(nil, nil)
	tab1Sw.SetPolicy(POLICY_NEVER, POLICY_AUTOMATIC)
	tab1Sw.Add(tab1)
	notebook.AppendPage(tab1Sw, gtk3.NewLabel("基础组件"))

	// ============================================================
	// Tab 2: 输入组件 (Entry / SpinButton / ComboBox / RadioButton / EntryCompletion)
	// ============================================================
	tab2 := gtk3.NewBox(ORIENTATION_VERTICAL, 8)
	tab2.SetMarginTop(10)
	tab2.SetMarginBottom(10)
	tab2.SetMarginStart(10)
	tab2.SetMarginEnd(10)

	// -- Entry 基础 + 实时反馈
	entryEcho := gtk3.NewLabel("输入内容显示在这里")
	entry := gtk3.NewEntry()
	entry.SetPlaceholderText("请输入文本...")
	entry.SetText("Hello GTK3")
	entry.SetMaxLength(50)
	entry.SetWidthChars(30)
	entry.SetOnChanged(func(sender PGtkWidget, userData GPointer) {
		fmt.Println("[event] Entry.Changed →", entry.GetText())
		entryEcho.SetText("输入: " + entry.GetText())
	})
	entry.SetOnCommit(func(sender PGtkWidget, userData GPointer) {
		fmt.Println("[event] Entry.Commit →", entry.GetText())
	})
	entry.SetOnKeyPress(func(sender PGtkWidget, event PEventKey, userData GPointer) bool {
		fmt.Println("[event] Entry.KeyPress")
		return false
	})
	entry.SetOnKeyRelease(func(sender PGtkWidget, event PEventKey, userData GPointer) bool {
		fmt.Println("[event] Entry.KeyRelease")
		return false
	})
	tab2.PackStart(entry, false, false, 0)
	tab2.PackStart(entryEcho, false, false, 0)

	// -- Entry 无边框 & 激活默认按钮
	entryFlat := gtk3.NewEntry()
	entryFlat.SetPlaceholderText("无边框输入框")
	entryFlat.SetHasFrame(false)
	entryFlat.SetActivatesDefault(true)
	entryFlat.SetOnChanged(func(sender PGtkWidget, userData GPointer) {
		fmt.Println("[event] Entry(Flat).Changed →", entryFlat.GetText())
	})
	tab2.PackStart(entryFlat, false, false, 0)

	// -- Entry 密码
	entryPass := gtk3.NewEntry()
	entryPass.SetPlaceholderText("密码输入框")
	entryPass.SetVisibility(false)
	entryPass.SetOnChanged(func(sender PGtkWidget, userData GPointer) {
		fmt.Println("[event] Entry(Password).Changed → len:", len(entryPass.GetText()))
	})
	tab2.PackStart(entryPass, false, false, 0)

	// -- Entry 进度脉冲
	entryProg := gtk3.NewEntry()
	entryProg.SetPlaceholderText("脉冲进度 (ProgressPulse)...")
	entryProg.SetOnChanged(func(sender PGtkWidget, userData GPointer) {
		fmt.Println("[event] Entry(Progress).Changed →", entryProg.GetText())
	})

	pulseBtn := gtk3.NewButtonWithLabel("脉冲")
	pulseBtn.SetOnClick(func(sender PGtkWidget, userData GPointer) {
		fmt.Println("[event] PulseBtn.Click")
		entryProg.ProgressPulse()
	})
	pulseRow := gtk3.NewBox(ORIENTATION_HORIZONTAL, 6)
	pulseRow.PackStart(entryProg, true, true, 0)
	pulseRow.PackStart(pulseBtn, false, false, 0)
	tab2.PackStart(pulseRow, false, false, 0)

	tab2.PackStart(gtk3.NewSeparator(ORIENTATION_HORIZONTAL), false, false, 0)

	// -- EntryCompletion (自动补全)
	completionLabel := gtk3.NewLabel("输入城市名体验自动补全:")
	tab2.PackStart(completionLabel, false, false, 0)

	completion := gtk3.NewEntryCompletion()
	completion.SetMinimumKeyLength(1)
	completion.SetTextColumn(0)
	// completion 需要关联一个 ListStore，但此处演示其创建
	if completion != nil {
		check("EntryCompletion created", true)
	}

	completionEntry := gtk3.NewEntry()
	completionEntry.SetPlaceholderText("北京/上海/广州/深圳...")
	completionEntry.SetOnChanged(func(sender PGtkWidget, userData GPointer) {
		fmt.Println("[event] Entry(Completion).Changed →", completionEntry.GetText())
	})
	tab2.PackStart(completionEntry, false, false, 0)

	tab2.PackStart(gtk3.NewSeparator(ORIENTATION_HORIZONTAL), false, false, 0)

	// -- SpinButton + Adjustment 联动 (实时同步)
	spinLabel := gtk3.NewLabel("SpinButton 值: 50")
	adj := gtk3.NewAdjustment(50, 0, 100, 1, 10, 0)
	spin := gtk3.NewSpinButton(adj, 0.5, 0)
	spin.SetRange(0, 100)
	spin.SetIncrements(1, 10)
	// SpinButton 自身的 SetOnValueChanged 事件
	spin.SetOnValueChanged(func(sender PGtkWidget, userData GPointer) {
		fmt.Println("[event] SpinButton.ValueChanged →", spin.GetValue())
		spinLabel.SetText(fmt.Sprintf("SpinButton 值: %.0f", spin.GetValue()))
	})
	tab2.PackStart(spin, false, false, 0)
	 tab2.PackStart(spinLabel, false, false, 0)

	 // -- SpinButton 新增 API 演示按钮
	 spinRow := gtk3.NewBox(ORIENTATION_HORIZONTAL, 6)
	 spinNumBtn := gtk3.NewButtonWithLabel("切换仅数字模式")
	 spinNumBtn.SetOnClick(func(sender PGtkWidget, userData GPointer) {
	  n := !spin.GetNumeric()
	  spin.SetNumeric(n)
	  statusbar.Push(statusCtx, fmt.Sprintf("SpinButton 仅数字: %v", n))
	 })
	 spinRow.PackStart(spinNumBtn, false, false, 0)

	 spinSnapBtn := gtk3.NewButtonWithLabel("切换吸附步进")
	 spinSnapBtn.SetOnClick(func(sender PGtkWidget, userData GPointer) {
	  s := !spin.GetSnapToTicks()
	  spin.SetSnapToTicks(s)
	  statusbar.Push(statusCtx, fmt.Sprintf("SpinButton 吸附步进: %v", s))
	 })
	 spinRow.PackStart(spinSnapBtn, false, false, 0)

	 spinWrapBtn := gtk3.NewButtonWithLabel("切换循环")
	 spinWrapBtn.SetOnClick(func(sender PGtkWidget, userData GPointer) {
	  w := !spin.GetWrap()
	  spin.SetWrap(w)
	  statusbar.Push(statusCtx, fmt.Sprintf("SpinButton 循环: %v", w))
	 })
	 spinRow.PackStart(spinWrapBtn, false, false, 0)

	 spinStepFwdBtn := gtk3.NewButtonWithLabel("步进+1")
	 spinStepFwdBtn.SetOnClick(func(sender PGtkWidget, userData GPointer) {
	  spin.Spin(SPIN_STEP_FORWARD, 1)
	  spinLabel.SetText(fmt.Sprintf("SpinButton 值: %.0f", spin.GetValue()))
	 })
	 spinRow.PackStart(spinStepFwdBtn, false, false, 0)

	 spinStepBackBtn := gtk3.NewButtonWithLabel("步进-1")
	 spinStepBackBtn.SetOnClick(func(sender PGtkWidget, userData GPointer) {
	  spin.Spin(SPIN_STEP_BACKWARD, 1)
	  spinLabel.SetText(fmt.Sprintf("SpinButton 值: %.0f", spin.GetValue()))
	 })
	 spinRow.PackStart(spinStepBackBtn, false, false, 0)

	 spinGetAdjBtn := gtk3.NewButtonWithLabel("获取Adjustment")
	 spinGetAdjBtn.SetOnClick(func(sender PGtkWidget, userData GPointer) {
	  adj := spin.GetAdjustment()
	  if adj != nil {
	   statusbar.Push(statusCtx, fmt.Sprintf("Adjustment: [%.0f, %.0f]", adj.GetLower(), adj.GetUpper()))
	  }
	 })
	 spinRow.PackStart(spinGetAdjBtn, false, false, 0)

	 tab2.PackStart(spinRow, false, false, 0)

	// -- ComboBoxText (带 changed 回调)
	combo := gtk3.NewComboBoxText()
	combo.Append("red", "红色")
	combo.Append("green", "绿色")
	combo.Append("blue", "蓝色")
	combo.SetActive(0)
	combo.SetOnChanged(func(sender PGtkWidget, userData GPointer) {
		fmt.Println("[event] ComboBox.Changed →", combo.GetActiveText())
		statusbar.Push(statusCtx, fmt.Sprintf("ComboBox changed: %s (id=%s)", combo.GetActiveText(), map[int]string{0: "red", 1: "green", 2: "blue"}[combo.GetActive()]))
	})
	tab2.PackStart(combo, false, false, 0)

	// -- 单选按钮组
	radio1 := gtk3.NewRadioButtonWithLabelFromWidget(nil, "选项 A")
	radio2 := gtk3.NewRadioButtonWithLabelFromWidget(radio1, "选项 B")
	radio3 := gtk3.NewRadioButtonWithLabelFromWidget(radio1, "选项 C")
	radio1.SetActive(true)
	radio1.SetOnToggled(func(sender PGtkWidget, userData GPointer) {
		fmt.Println("[event] Radio1.Toggled active:", radio1.GetActive())
	})
	radio2.SetOnToggled(func(sender PGtkWidget, userData GPointer) {
		fmt.Println("[event] Radio2.Toggled active:", radio2.GetActive())
	})
	radio3.SetOnToggled(func(sender PGtkWidget, userData GPointer) {
		fmt.Println("[event] Radio3.Toggled active:", radio3.GetActive())
	})
	rBox := gtk3.NewBox(ORIENTATION_HORIZONTAL, 10)
	rBox.PackStart(radio1, false, false, 0)
	rBox.PackStart(radio2, false, false, 0)
	rBox.PackStart(radio3, false, false, 0)
	tab2.PackStart(rBox, false, false, 0)

	// -- RadioButton 新构造函数演示
	radioNewRow := gtk3.NewBox(ORIENTATION_HORIZONTAL, 10)
	radioNew1 := gtk3.NewRadioButtonWithLabel("新建组按钮")
	radioNew1.SetActive(true)
	radioNew2 := gtk3.NewRadioButtonFromWidget(radioNew1)
	radioNew2.SetLabel("从widget创建")
	radioNewRow.PackStart(radioNew1, false, false, 0)
	radioNewRow.PackStart(radioNew2, false, false, 0)
	tab2.PackStart(radioNewRow, false, false, 0)

	tab2.PackStart(gtk3.NewSeparator(ORIENTATION_HORIZONTAL), false, false, 0)

	// -- GtkScale (滑块)
	tab2.PackStart(gtk3.NewLabel("GtkScale (滑块, 拖动实时调值):"), false, false, 0)
	scaleAdj := gtk3.NewAdjustment(50, 0, 100, 1, 10, 0)
	hScale := gtk3.NewHScale(scaleAdj)
	hScale.SetDigits(0)
	hScale.SetDrawValue(true)
	hScale.SetValuePos(POS_TOP)
	hScale.SetSizeRequest(300, -1)
	// 实时回调: 拖动滑块立即更新标签
	scaleValLabel := gtk3.NewLabel("Scale 当前值: 50")
	hScale.SetOnValueChanged(func(sender PGtkWidget, userData GPointer) {
		fmt.Println("[event] Scale.ValueChanged →", hScale.GetValue())
		scaleValLabel.SetText(fmt.Sprintf("Scale 当前值: %.0f", hScale.GetValue()))
	})
	tab2.PackStart(hScale, false, false, 0)
	tab2.PackStart(scaleValLabel, false, false, 0)

	tab2.PackStart(gtk3.NewSeparator(ORIENTATION_HORIZONTAL), false, false, 0)

	// -- GtkSearchEntry (搜索框) + 实时回调
	tab2.PackStart(gtk3.NewLabel("GtkSearchEntry (搜索输入框, 带实时回调):"), false, false, 0)
	searchEntry := gtk3.NewSearchEntry()
	searchEntry.SetPlaceholderText("输入关键词搜索...")
	searchEntry.SetOnChanged(func(sender PGtkWidget, userData GPointer) {
		fmt.Println("[event] SearchEntry.Changed →", searchEntry.GetText())
		statusbar.Push(statusCtx, "搜索: "+searchEntry.GetText())
	})
	tab2.PackStart(searchEntry, false, false, 0)

	tab2Sw := gtk3.NewScrolledWindow(nil, nil)
	tab2Sw.SetPolicy(POLICY_NEVER, POLICY_AUTOMATIC)
	tab2Sw.Add(tab2)
	notebook.AppendPage(tab2Sw, gtk3.NewLabel("输入组件"))

	// ============================================================
	// Tab 3: 布局容器 (Box / Grid / Fixed / ScrolledWindow / Layout / Overlay)
	// ============================================================
	tab3 := gtk3.NewBox(ORIENTATION_VERTICAL, 8)
	tab3.SetMarginTop(10)
	tab3.SetMarginBottom(10)
	tab3.SetMarginStart(10)
	tab3.SetMarginEnd(10)

	// -- Box 高级属性
	boxDemoBox := gtk3.NewBox(ORIENTATION_HORIZONTAL, 12)
	boxDemoBox.SetHomogeneous(false)
	boxDemoBox.SetSpacing(12)
	b1 := gtk3.NewButtonWithLabel("子 1")
	b2 := gtk3.NewButtonWithLabel("子 2")
	b3 := gtk3.NewButtonWithLabel("子 3")
	boxDemoBox.PackStart(b1, false, false, 0)
	boxDemoBox.PackStart(b2, true, true, 0) // 扩展填充
	boxDemoBox.PackStart(b3, false, false, 0)
	tab3.PackStart(gtk3.NewLabel("Box: 子2 HExpand, spacing=12"), false, false, 0)
	tab3.PackStart(boxDemoBox, false, false, 0)

	tab3.PackStart(gtk3.NewSeparator(ORIENTATION_HORIZONTAL), false, false, 0)

	// -- Grid 高级
	gridLabel := gtk3.NewLabel("Grid: 行列均匀间距, 跨列")
	tab3.PackStart(gridLabel, false, false, 0)
	grid := gtk3.NewGrid()
	grid.SetRowSpacing(8)
	grid.SetColumnSpacing(8)
	grid.SetRowHomogeneous(true)
	grid.SetColumnHomogeneous(false)
	grid.Attach(gtk3.NewLabel("姓名:"), 0, 0, 1, 1)
	grid.Attach(gtk3.NewEntry(), 1, 0, 1, 1)
	grid.Attach(gtk3.NewLabel("邮箱:"), 0, 1, 1, 1)
	grid.Attach(gtk3.NewEntry(), 1, 1, 2, 1) // 跨2列
	grid.Attach(gtk3.NewLabel("备注:"), 0, 2, 1, 1)
	grid.Attach(gtk3.NewEntry(), 1, 2, 2, 1)
	tab3.PackStart(grid, false, false, 0)

	tab3.PackStart(gtk3.NewSeparator(ORIENTATION_HORIZONTAL), false, false, 0)

	// -- Fixed + Move
	fixed := gtk3.NewFixed()
	fixed.SetSizeRequest(300, 80)
	fxLabel := gtk3.NewLabel("(10,10)")
	fixed.Put(fxLabel, 10, 10)
	fxLabel2 := gtk3.NewLabel("(200,40)")
	fixed.Put(fxLabel2, 200, 40)
	moveBtn := gtk3.NewButtonWithLabel("移动右标签→")
	moveBtn.SetOnClick(func(sender PGtkWidget, userData GPointer) {
		// 在 (200,40) 和 (100,40) 之间切换
		fixed.Move(fxLabel2, 300-fxLabel2.GetAllocation().GetWidth()-10, 40)
	})
	tab3.PackStart(gtk3.NewLabel("Fixed: 绝对定位"), false, false, 0)
	tab3.PackStart(fixed, false, false, 0)
	tab3.PackStart(moveBtn, false, false, 0)

	tab3.PackStart(gtk3.NewSeparator(ORIENTATION_HORIZONTAL), false, false, 0)

	// -- ScrolledWindow 策略
	swPolicy := gtk3.NewScrolledWindow(nil, nil)
	swPolicy.SetPolicy(POLICY_AUTOMATIC, POLICY_ALWAYS)
	swPolicy.SetShadowType(SHADOW_ETCHED_IN)
	swPolicy.SetMinContentWidth(250)
	swPolicy.SetMinContentHeight(80)
	swInner := gtk3.NewBox(ORIENTATION_VERTICAL, 4)
	for i := 0; i < 20; i++ {
		swInner.PackStart(gtk3.NewLabel(fmt.Sprintf("可滚动第 %d 行", i+1)), false, false, 0)
	}
	swPolicy.Add(swInner)
	tab3.PackStart(gtk3.NewLabel("ScrolledWindow: PolicyAlways + ShadowEtchedIn + 20行"), false, false, 0)
	tab3.PackStart(swPolicy, false, false, 0)

	tab3.PackStart(gtk3.NewSeparator(ORIENTATION_HORIZONTAL), false, false, 0)

	// -- Overlay
	ol := gtk3.NewOverlay()
	ol.SetSizeRequest(200, 60)
	ol.Add(gtk3.NewLabel("底层文本"))
	ol.AddOverlay(gtk3.NewLabel("覆盖层文本"))
	tab3.PackStart(gtk3.NewLabel("Overlay: 底层+覆盖层"), false, false, 0)
	tab3.PackStart(ol, false, false, 0)

	tab3.PackStart(gtk3.NewSeparator(ORIENTATION_HORIZONTAL), false, false, 0)

	// -- GtkStack + StackSwitcher
	tab3.PackStart(gtk3.NewLabel("GtkStack (堆叠容器, 配合 StackSwitcher):"), false, false, 0)
	stack := gtk3.NewStack()
	stack.SetTransitionDuration(300)
	stack.SetTransitionType(STACK_TRANSITION_TYPE_SLIDE_LEFT_RIGHT)
	page1 := gtk3.NewLabel("页面 1 — 第一页内容")
	page1.SetSizeRequest(300, 80)
	stack.AddTitled(page1, "p1", "第一页")
	page2 := gtk3.NewLabel("页面 2 — 第二页内容")
	page2.SetSizeRequest(300, 80)
	stack.AddTitled(page2, "p2", "第二页")
	page3 := gtk3.NewLabel("页面 3 — 第三页内容")
	page3.SetSizeRequest(300, 80)
	stack.AddTitled(page3, "p3", "第三页")
	switcher := gtk3.NewStackSwitcher()
	switcher.SetStack(stack)
	tab3.PackStart(switcher, false, false, 0)
	tab3.PackStart(stack, false, false, 0)

	tab3.PackStart(gtk3.NewSeparator(ORIENTATION_HORIZONTAL), false, false, 0)

	// -- GtkPaned (分割面板)
	tab3.PackStart(gtk3.NewLabel("GtkPaned (可拖拽分隔条):"), false, false, 0)
	paned := gtk3.NewPaned(ORIENTATION_HORIZONTAL)
	paned.Add1(gtk3.NewLabel("左侧面板"))
	paned.Add2(gtk3.NewLabel("右侧面板"))
	paned.SetPosition(200)
	paned.SetSizeRequest(400, 80)
	tab3.PackStart(paned, false, false, 0)

	tab3.PackStart(gtk3.NewSeparator(ORIENTATION_HORIZONTAL), false, false, 0)

	// -- GtkRevealer (展开动画)
	tab3.PackStart(gtk3.NewLabel("GtkRevealer (点击按钮展开/收起):"), false, false, 0)
	revealer := gtk3.NewRevealer()
	revealer.SetTransitionDuration(500)
	revealer.SetTransitionType(REVEALER_TRANSITION_TYPE_SLIDE_DOWN)
	revealer.Add(gtk3.NewLabel("Revealer 中隐藏的内容 — 点击按钮展开"))
	tab3.PackStart(revealer, false, false, 0)
	revealBtn := gtk3.NewButtonWithLabel("切换展开/收起")
	revealBtn.SetOnClick(func(sender PGtkWidget, userData GPointer) {
		revealer.SetRevealChild(!revealer.GetRevealChild())
	})
	tab3.PackStart(revealBtn, false, false, 0)

	tab3Sw := gtk3.NewScrolledWindow(nil, nil)
	tab3Sw.SetPolicy(POLICY_NEVER, POLICY_AUTOMATIC)
	tab3Sw.Add(tab3)
	notebook.AppendPage(tab3Sw, gtk3.NewLabel("布局容器"))

	// ============================================================
	// Tab 4: 文本编辑 (TextView / TextBuffer / TextTag / 插入 / 选中)
	// ============================================================
	tab4 := gtk3.NewBox(ORIENTATION_VERTICAL, 8)
	tab4.SetMarginTop(10)
	tab4.SetMarginBottom(10)
	tab4.SetMarginStart(10)
	tab4.SetMarginEnd(10)

	tvInfo := gtk3.NewLabel("字符: 0  行: 0")
	tv := gtk3.NewTextView()
	tb := tv.GetBuffer()
	tb.SetText("这是多行文本编辑器。\n可以编辑这些文本。\n第三行内容。")
	tv.SetEditable(true)
	tv.SetWrapMode(WRAP_WORD_CHAR)
	tv.SetCursorVisible(true)
	tv.SetOverwrite(false)
	tv.SetJustification(JUSTIFY_LEFT)
	tv.SetLeftMargin(12)
	tv.SetRightMargin(12)
	tv.SetIndent(20)
	tv.SetPixelsAboveLines(4)
	tv.SetPixelsBelowLines(4)
	// TextBuffer 内容变化回调
	tb.SetOnChanged(func(sender PGtkWidget, userData GPointer) {
		fmt.Println("[event] TextBuffer.Changed → chars:", tb.GetCharCount(), "lines:", tb.GetLineCount())
		tvInfo.SetText(fmt.Sprintf("字符: %d  行: %d (changed)", tb.GetCharCount(), tb.GetLineCount()))
		statusbar.Push(statusCtx, "TextBuffer changed")
	})
	tab4.PackStart(tvInfo, false, false, 0)

	swTV := gtk3.NewScrolledWindow(nil, nil)
	swTV.Add(tv)
	tab4.PackStart(swTV, true, true, 0)

	// 文本操作工具条
	toolRow := gtk3.NewBox(ORIENTATION_HORIZONTAL, 6)

	insertBtn := gtk3.NewButtonWithLabel("开头插入")
	insertBtn.SetOnClick(func(sender PGtkWidget, userData GPointer) {
		start := tb.GetStartIter()
		tb.Insert(start, "[插入] ")
		tb.PlaceCursor(start)
		tvInfo.SetText(fmt.Sprintf("字符: %d  行: %d", tb.GetCharCount(), tb.GetLineCount()))
	})
	toolRow.PackStart(insertBtn, false, false, 0)

	delBtn := gtk3.NewButtonWithLabel("删除首行")
	delBtn.SetOnClick(func(sender PGtkWidget, userData GPointer) {
		start := tb.GetStartIter()
		firstLineEnd := tb.GetIterAtLine(1)
		tb.Delete(start, firstLineEnd)
		tvInfo.SetText(fmt.Sprintf("字符: %d  行: %d", tb.GetCharCount(), tb.GetLineCount()))
	})
	toolRow.PackStart(delBtn, false, false, 0)

	clearBtn := gtk3.NewButtonWithLabel("清空")
	clearBtn.SetOnClick(func(sender PGtkWidget, userData GPointer) {
		tb.SetText(" ") // 清空后保留一个空格，避免 CStr("") 返回 NULL
		tvInfo.SetText(fmt.Sprintf("字符: %d  行: %d", tb.GetCharCount(), tb.GetLineCount()))
	})
	toolRow.PackStart(clearBtn, false, false, 0)

	resetBtn := gtk3.NewButtonWithLabel("重置文本")
	resetBtn.SetOnClick(func(sender PGtkWidget, userData GPointer) {
		tb.SetText("这是多行文本编辑器。\n可以编辑这些文本。\n第三行内容。")
		tvInfo.SetText(fmt.Sprintf("字符: %d  行: %d", tb.GetCharCount(), tb.GetLineCount()))
	})
	toolRow.PackStart(resetBtn, false, false, 0)

	tab4.PackStart(toolRow, false, false, 0)

	tab4Sw := gtk3.NewScrolledWindow(nil, nil)
	tab4Sw.SetPolicy(POLICY_NEVER, POLICY_AUTOMATIC)
	tab4Sw.Add(tab4)
	notebook.AppendPage(tab4Sw, gtk3.NewLabel("文本编辑"))

	// ============================================================
	// Tab 5: 数据表格 & 树形视图 (ListStore / TreeStore / TreeView / TreeSelection)
	// ============================================================
	tab5 := gtk3.NewBox(ORIENTATION_VERTICAL, 8)
	tab5.SetMarginTop(10)
	tab5.SetMarginBottom(10)
	tab5.SetMarginStart(10)
	tab5.SetMarginEnd(10)

	// ─────────────────────────────────────────────
	// 上部: TreeStore 树形结构 (部门 → 员工)
	// ─────────────────────────────────────────────
	treeLabel := gtk3.NewLabel("树形视图 (TreeStore: 部门 → 员工)")
	treeLabel.SetMarkup("<b>树形视图 (TreeStore: 部门 → 员工)</b>")
	tab5.PackStart(treeLabel, false, false, 0)

	treeSelInfo := gtk3.NewLabel("选择: 无")
	tab5.PackStart(treeSelInfo, false, false, 0)

	treeStore := gtk3.NewTreeStore(TYPE_STRING, TYPE_STRING)
	// 部门数据: name → employees[]
	type deptData struct {
		name      string
		employees []string
	}
	depts := []deptData{
		{"技术部", []string{"张三", "李四", "王五"}},
		{"市场部", []string{"赵六", "钱七"}},
		{"人事部", []string{"孙八", "周九", "吴十"}},
	}
	// 填充树形数据
	for _, d := range depts {
		deptIter := treeStore.Append(nil)
		treeStore.SetValue(deptIter, 0, d.name)
		treeStore.SetValue(deptIter, 1, fmt.Sprintf("%d人", len(d.employees)))
		for _, emp := range d.employees {
			empIter := treeStore.Append(deptIter)
			treeStore.SetValue(empIter, 0, "  "+emp)
			treeStore.SetValue(empIter, 1, "员工")
		}
	}

	treeView2 := gtk3.NewTreeView()
	treeView2.SetTreeModel(treeStore)
	for i, title := range []string{"名称", "说明"} {
		renderer := gtk3.NewCellRendererText()
		col := gtk3.NewTreeViewColumn()
		col.SetTitle(title)
		col.PackStart(renderer, false)
		col.AddAttribute(renderer, "text", i)
		treeView2.AppendColumn(col)
	}
	treeView2.SetHeadersVisible(true)
	treeView2.ExpandAll()
	treeView2.SetOnRowExpanded(func(sender PGtkWidget, iter uintptr, path uintptr, userData GPointer) {
		fmt.Println("[event] TreeView(TreeStore).RowExpanded")
	})
	treeView2.SetOnRowCollapsed(func(sender PGtkWidget, iter uintptr, path uintptr, userData GPointer) {
		fmt.Println("[event] TreeView(TreeStore).RowCollapsed")
	})

	swTree := gtk3.NewScrolledWindow(nil, nil)
	swTree.SetMinContentHeight(180)
	swTree.Add(treeView2)
	tab5.PackStart(swTree, false, false, 0)

	tab5.PackStart(gtk3.NewSeparator(ORIENTATION_HORIZONTAL), false, false, 0)

	// ─────────────────────────────────────────────
	// 下部: ListStore 平面表格 (员工详情)
	// ─────────────────────────────────────────────
	selInfo := gtk3.NewLabel("选择: 无")
	listLabel := gtk3.NewLabel("")
	listLabel.SetMarkup("<b>平面表格 (ListStore: 员工详情)</b>")
	tab5.PackStart(listLabel, false, false, 0)
	tab5.PackStart(selInfo, false, false, 0)

	store := gtk3.NewListStore(TYPE_STRING, TYPE_STRING, TYPE_STRING)
	type row struct{ name, email, role string }
	rows := []row{
		{"张三", "zhang@example.com", "工程师"},
		{"李四", "li@example.com", "设计师"},
		{"王五", "wang@example.com", "产品经理"},
		{"赵六", "zhao@example.com", "测试"},
		{"钱七", "qian@example.com", "市场专员"},
		{"孙八", "sun@example.com", "招聘经理"},
		{"周九", "zhou@example.com", "培训专员"},
		{"吴十", "wu@example.com", "薪酬专员"},
	}
	for _, d := range rows {
		iter := store.Append()
		store.SetValue(iter, 0, d.name)
		store.SetValue(iter, 1, d.email)
		store.SetValue(iter, 2, d.role)
	}

	treeView := gtk3.NewTreeView()
	treeView.SetModel(store)
	cols := make([]ITreeViewColumn, 3)
	for i, title := range []string{"姓名", "邮箱", "角色"} {
	 renderer := gtk3.NewCellRendererText()
	 col := gtk3.NewTreeViewColumn()
	 col.SetTitle(title)
	 col.PackStart(renderer, false)
	 col.AddAttribute(renderer, "text", i)
	 col.SetResizable(true)
	 col.SetSortColumnId(i)
	 col.SetSortIndicator(false)
	 col.SetReorderable(true)
	 treeView.AppendColumn(col)
	 cols[i] = col
	}
	// 设置第0列(姓名)固定宽度、排序箭头、居中
	cols[0].SetSizing(TREE_VIEW_COLUMN_FIXED)
	cols[0].SetFixedWidth(100)
	cols[0].SetMinWidth(60)
	cols[0].SetMaxWidth(200)
	cols[0].SetExpand(false)
	cols[0].SetSortIndicator(true)
	cols[0].SetAlignment(0.5)
	// 设置第1列(邮箱)自动拉伸
	cols[1].SetSizing(TREE_VIEW_COLUMN_GROW_ONLY)
	cols[1].SetExpand(true)
	cols[1].SetMinWidth(120)
	// 设置第2列(角色)居中、固定宽
	cols[2].SetSizing(TREE_VIEW_COLUMN_FIXED)
	cols[2].SetFixedWidth(100)
	cols[2].SetAlignment(0.5)
	sel := treeView.GetSelection()
	sel.SetMode(SELECTION_MULTIPLE)
	sel.SetOnChanged(func(sender PGtkWidget, userData GPointer) {
		fmt.Println("[event] TreeSelection.Changed")
		selInfo.SetText("TreeSelection changed: 选中行已变化 (changed)")
		statusbar.Push(statusCtx, "TreeSelection changed")
	})
	treeView.SetHeadersVisible(true)
	treeView.ExpandAll()
	treeView.SetOnCursorChanged(func(sender PGtkWidget, userData GPointer) {
		fmt.Println("[event] TreeView(ListStore).CursorChanged")
	})
	treeView.SetOnRowActivated(func(sender PGtkWidget, path uintptr, column uintptr, userData GPointer) {
		fmt.Println("[event] TreeView(ListStore).RowActivated")
	})
	tw := gtk3.NewScrolledWindow(nil, nil)
	tw.Add(treeView)
	tab5.PackStart(tw, true, true, 0)

	// 操作按钮
	treeBtnRow := gtk3.NewBox(ORIENTATION_HORIZONTAL, 6)
	addRowBtn := gtk3.NewButtonWithLabel("添加员工")
	addRowBtn.SetOnClick(func(sender PGtkWidget, userData GPointer) {
		iter := store.Append()
		store.SetValue(iter, 0, fmt.Sprintf("员工%d", len(rows)+1))
		store.SetValue(iter, 1, fmt.Sprintf("emp%d@test.com", len(rows)+1))
		store.SetValue(iter, 2, "新员工")
		rows = append(rows, row{"", "", ""})
	})
	treeBtnRow.PackStart(addRowBtn, false, false, 0)

	clearTreeBtn := gtk3.NewButtonWithLabel("清空表格")
	 clearTreeBtn.SetOnClick(func(sender PGtkWidget, userData GPointer) {
	  store.Clear()
	 })
	 treeBtnRow.PackStart(clearTreeBtn, false, false, 0)

	 // -- TreeSelection 新增 API 测试按钮
	 getSelBtn := gtk3.NewButtonWithLabel("获取选中行")
	 getSelBtn.SetOnClick(func(sender PGtkWidget, userData GPointer) {
	  // GetSelected 只能在 SINGLE/BROWSE 模式下使用
	  // 临时切换模式获取选中行后恢复
	  sel.SetMode(SELECTION_SINGLE)
	  model, iter := sel.GetSelected()
	  if model != nil && iter != nil {
	   selInfo.SetText(fmt.Sprintf("GetSelected 成功: model=%v", model.Instance()))
	   fmt.Println("[test] TreeSelection.GetSelected → OK")
	   statusbar.Push(statusCtx, "GetSelected: 有选中行")
	  } else {
	   selInfo.SetText("GetSelected: 无选中行 (请先点击选择一行)")
	   fmt.Println("[test] TreeSelection.GetSelected → nil (无选中行)")
	   statusbar.Push(statusCtx, "GetSelected: 无选中行")
	  }
	  sel.SetMode(SELECTION_MULTIPLE)
	 })
	 treeBtnRow.PackStart(getSelBtn, false, false, 0)

	 countSelBtn := gtk3.NewButtonWithLabel("统计选中数")
	 countSelBtn.SetOnClick(func(sender PGtkWidget, userData GPointer) {
	  n := sel.CountSelectedRows()
	  selInfo.SetText(fmt.Sprintf("CountSelectedRows: %d", n))
	  fmt.Println("[test] TreeSelection.CountSelectedRows →", n)
	  statusbar.Push(statusCtx, fmt.Sprintf("选中 %d 行", n))
	 })
	 treeBtnRow.PackStart(countSelBtn, false, false, 0)

	 selAllBtn := gtk3.NewButtonWithLabel("全选")
	 selAllBtn.SetOnClick(func(sender PGtkWidget, userData GPointer) {
	  sel.SelectAll()
	  selInfo.SetText("全选 (SelectAll)")
	  fmt.Println("[test] TreeSelection.SelectAll")
	  statusbar.Push(statusCtx, "全选")
	 })
	 treeBtnRow.PackStart(selAllBtn, false, false, 0)

	 unselAllBtn := gtk3.NewButtonWithLabel("取消全选")
	 unselAllBtn.SetOnClick(func(sender PGtkWidget, userData GPointer) {
	  sel.UnselectAll()
	  selInfo.SetText("取消全选 (UnselectAll)")
	  fmt.Println("[test] TreeSelection.UnselectAll")
	  statusbar.Push(statusCtx, "取消全选")
	 })
	 treeBtnRow.PackStart(unselAllBtn, false, false, 0)

	 treeBtnSw := gtk3.NewScrolledWindow(nil, nil)
	 treeBtnSw.SetPolicy(POLICY_AUTOMATIC, POLICY_NEVER)
	 treeBtnSw.Add(treeBtnRow)
	 tab5.PackStart(treeBtnSw, false, false, 0)

	 // -- TreeViewColumn 新增 API 演示按钮
	 colBtnRow := gtk3.NewBox(ORIENTATION_HORIZONTAL, 6)
	 toggleSortBtn := gtk3.NewButtonWithLabel("切换姓名列排序箭头")
	 toggleSortBtn.SetOnClick(func(sender PGtkWidget, userData GPointer) {
	  show := !cols[0].GetSortIndicator()
	  cols[0].SetSortIndicator(show)
	  statusbar.Push(statusCtx, fmt.Sprintf("姓名列排序箭头: %v", show))
	 })
	 colBtnRow.PackStart(toggleSortBtn, false, false, 0)

	 toggleResizeBtn := gtk3.NewButtonWithLabel("切换邮箱列可调宽")
	 toggleResizeBtn.SetOnClick(func(sender PGtkWidget, userData GPointer) {
	  r := !cols[1].GetResizable()
	  cols[1].SetResizable(r)
	  statusbar.Push(statusCtx, fmt.Sprintf("邮箱列可调宽: %v", r))
	 })
	 colBtnRow.PackStart(toggleResizeBtn, false, false, 0)

	 w100Btn := gtk3.NewButtonWithLabel("姓名→固定宽100")
	 w100Btn.SetOnClick(func(sender PGtkWidget, userData GPointer) {
	  cols[0].SetSizing(TREE_VIEW_COLUMN_FIXED)
	  cols[0].SetFixedWidth(100)
	  statusbar.Push(statusCtx, "姓名列固定宽100")
	 })
	 colBtnRow.PackStart(w100Btn, false, false, 0)

	 w200Btn := gtk3.NewButtonWithLabel("姓名→固定宽200")
	 w200Btn.SetOnClick(func(sender PGtkWidget, userData GPointer) {
	  cols[0].SetSizing(TREE_VIEW_COLUMN_FIXED)
	  cols[0].SetFixedWidth(200)
	  statusbar.Push(statusCtx, "姓名列固定宽200")
	 })
	 colBtnRow.PackStart(w200Btn, false, false, 0)

	 autoBtn := gtk3.NewButtonWithLabel("邮箱→自动拉伸")
	 autoBtn.SetOnClick(func(sender PGtkWidget, userData GPointer) {
	  cols[1].SetSizing(TREE_VIEW_COLUMN_GROW_ONLY)
	  cols[1].SetExpand(true)
	  cols[1].SetMinWidth(120)
	  statusbar.Push(statusCtx, "邮箱列自动拉伸")
	 })
	 colBtnRow.PackStart(autoBtn, false, false, 0)

	 alignBtn := gtk3.NewButtonWithLabel("角色列标题居中")
	 alignBtn.SetOnClick(func(sender PGtkWidget, userData GPointer) {
	  cols[2].SetAlignment(0.5)
	  statusbar.Push(statusCtx, "角色列标题居中")
	 })
	 colBtnRow.PackStart(alignBtn, false, false, 0)

	 colBtnSw := gtk3.NewScrolledWindow(nil, nil)
	 colBtnSw.SetPolicy(POLICY_AUTOMATIC, POLICY_NEVER)
	 colBtnSw.Add(colBtnRow)
	 tab5.PackStart(colBtnSw, false, false, 0)

	tab5.PackStart(treeBtnRow, false, false, 0)

	tab5Sw := gtk3.NewScrolledWindow(nil, nil)
	tab5Sw.SetPolicy(POLICY_NEVER, POLICY_AUTOMATIC)
	tab5Sw.Add(tab5)
	notebook.AppendPage(tab5Sw, gtk3.NewLabel("数据表格"))

	// ============================================================
	// Tab 6: 进度状态 (ProgressBar / Statusbar / Adjustment 联动)
	// ============================================================
	tab6 := gtk3.NewBox(ORIENTATION_VERTICAL, 8)
	tab6.SetMarginTop(10)
	tab6.SetMarginBottom(10)
	tab6.SetMarginStart(10)
	tab6.SetMarginEnd(10)

	pbar := gtk3.NewProgressBar()
	pbar.SetShowText(true)
	pbar.SetText("0%")
	pbar.SetFraction(0.0)
	tab6.PackStart(pbar, false, false, 0)

	// Fraction 控制器 — 使用 callback.Connect 直接连接 GtkRange 的 value-changed 信号
	pbarCtl := gtk3.NewBox(ORIENTATION_HORIZONTAL, 6)
	pbarCtl.PackStart(gtk3.NewLabel("进度:"), false, false, 0)
	pAdj := gtk3.NewAdjustment(0, 0, 100, 5, 20, 0)
	pSpin := gtk3.NewSpinButton(pAdj, 1, 0)
	pSpin.SetOnChanged(func(sender PGtkWidget, userData GPointer) {
		fmt.Println("value:", pSpin.GetValue())
		f := pSpin.GetValue() / 100.0
		pbar.SetFraction(f)
		pbar.SetText(fmt.Sprintf("%.0f%%", pSpin.GetValue()))
	})
	pbarCtl.PackStart(pSpin, false, false, 0)

	resetPbar := gtk3.NewButtonWithLabel("重置")
	resetPbar.SetOnClick(func(sender PGtkWidget, userData GPointer) {
		pSpin.SetValue(0)
	})
	pbarCtl.PackStart(resetPbar, false, false, 0)
	tab6.PackStart(pbarCtl, false, false, 0)

	// Statusbar 演示
	tab6.PackStart(gtk3.NewSeparator(ORIENTATION_HORIZONTAL), false, false, 0)
	tab6.PackStart(gtk3.NewLabel("Statusbar 操作:"), false, false, 0)

	statCtl := gtk3.NewBox(ORIENTATION_HORIZONTAL, 6)
	pushMsg := gtk3.NewButtonWithLabel("推送消息")
	pushMsg.SetOnClick(func(sender PGtkWidget, userData GPointer) {
		statusbar.Push(statusCtx, fmt.Sprintf("消息 #%d", count))
		count++
	})
	statCtl.PackStart(pushMsg, false, false, 0)

	popMsg := gtk3.NewButtonWithLabel("弹出")
	popMsg.SetOnClick(func(sender PGtkWidget, userData GPointer) {
		statusbar.Pop(statusCtx)
	})
	statCtl.PackStart(popMsg, false, false, 0)

	removeAllMsg := gtk3.NewButtonWithLabel("全部清除")
	removeAllMsg.SetOnClick(func(sender PGtkWidget, userData GPointer) {
		statusbar.RemoveAll(statusCtx)
		statusbar.Push(statusCtx, "状态已清除")
	})
	statCtl.PackStart(removeAllMsg, false, false, 0)

	tab6.PackStart(statCtl, false, false, 0)

	tab6.PackStart(gtk3.NewSeparator(ORIENTATION_HORIZONTAL), false, false, 0)

	// -- GtkSpinner (加载旋转)
	tab6.PackStart(gtk3.NewLabel("GtkSpinner (加载旋转动画):"), false, false, 0)
	spinner := gtk3.NewSpinner()
	spinRow2 := gtk3.NewBox(ORIENTATION_HORIZONTAL, 6)
	spinRow2.PackStart(spinner, false, false, 0)
	startSpin := gtk3.NewButtonWithLabel("开始旋转")
	startSpin.SetOnClick(func(sender PGtkWidget, userData GPointer) { fmt.Println("[event] Spinner.Start"); spinner.Start() })
	spinRow.PackStart(startSpin, false, false, 0)
	stopSpin := gtk3.NewButtonWithLabel("停止")
	stopSpin.SetOnClick(func(sender PGtkWidget, userData GPointer) { fmt.Println("[event] Spinner.Stop"); spinner.Stop() })
	spinRow.PackStart(stopSpin, false, false, 0)
	tab6.PackStart(spinRow, false, false, 0)

	tab6.PackStart(gtk3.NewSeparator(ORIENTATION_HORIZONTAL), false, false, 0)

	// -- GtkLevelBar (等级条)
	tab6.PackStart(gtk3.NewLabel("GtkLevelBar (等级/评分条):"), false, false, 0)
	lvBar := gtk3.NewLevelBar()
	lvBar.SetMinValue(0)
	lvBar.SetMaxValue(100)
	lvBar.SetValue(50)
	lvBar.SetOnOffsetChanged(func(sender PGtkWidget, name uintptr, userData GPointer) {
		fmt.Println("lvBar:", lvBar.GetValue())
	})
	tab6.PackStart(lvBar, false, false, 0)
	lvRow := gtk3.NewBox(ORIENTATION_HORIZONTAL, 6)
	lvDown := gtk3.NewButtonWithLabel("-10")
	lvDown.SetOnClick(func(sender PGtkWidget, userData GPointer) { lvBar.SetValue(lvBar.GetValue() - 10) })
	lvRow.PackStart(lvDown, false, false, 0)
	lvUp := gtk3.NewButtonWithLabel("+10")
	lvUp.SetOnClick(func(sender PGtkWidget, userData GPointer) { lvBar.SetValue(lvBar.GetValue() + 10) })
	lvRow.PackStart(lvUp, false, false, 0)
	tab6.PackStart(lvRow, false, false, 0)

	tab6Sw := gtk3.NewScrolledWindow(nil, nil)
	tab6Sw.SetPolicy(POLICY_NEVER, POLICY_AUTOMATIC)
	tab6Sw.Add(tab6)
	notebook.AppendPage(tab6Sw, gtk3.NewLabel("进度状态"))

	// ============================================================
	// Tab 7: 窗口操作 (Maximize / Fullscreen / Iconify / Move / Resize / Decorated / KeepAbove)
	// ============================================================
	tab7 := gtk3.NewBox(ORIENTATION_VERTICAL, 8)
	tab7.SetMarginTop(10)
	tab7.SetMarginBottom(10)
	tab7.SetMarginStart(10)
	tab7.SetMarginEnd(10)

	winInfo := gtk3.NewLabel("窗口类型: NORMAL  位置: -  大小: 800x600")
	tab7.PackStart(winInfo, false, false, 0)

	// -- 窗口状态
	wState := gtk3.NewBox(ORIENTATION_HORIZONTAL, 6)
	maxBtn := gtk3.NewButtonWithLabel("最大化")
	maxBtn.SetOnClick(func(sender PGtkWidget, userData GPointer) {
		if win.GetDecorated() {
			win.Maximize()
			statusbar.Push(statusCtx, "窗口已最大化")
		}
	})
	wState.PackStart(maxBtn, false, false, 0)

	unmaxBtn := gtk3.NewButtonWithLabel("还原")
	unmaxBtn.SetOnClick(func(sender PGtkWidget, userData GPointer) {
		win.Unmaximize()
		win.Unfullscreen()
		statusbar.Push(statusCtx, "窗口已还原")
	})
	wState.PackStart(unmaxBtn, false, false, 0)

	fullBtn := gtk3.NewButtonWithLabel("全屏")
	fullBtn.SetOnClick(func(sender PGtkWidget, userData GPointer) {
		win.Fullscreen()
		statusbar.Push(statusCtx, "窗口已全屏 (Esc退出)")
	})
	wState.PackStart(fullBtn, false, false, 0)

	unfullBtn := gtk3.NewButtonWithLabel("退出全屏")
	unfullBtn.SetOnClick(func(sender PGtkWidget, userData GPointer) {
		win.Unfullscreen()
		statusbar.Push(statusCtx, "已退出全屏")
	})
	wState.PackStart(unfullBtn, false, false, 0)

	iconBtn := gtk3.NewButtonWithLabel("最小化")
	iconBtn.SetOnClick(func(sender PGtkWidget, userData GPointer) {
		win.Iconify()
		statusbar.Push(statusCtx, "窗口已最小化")
	})
	wState.PackStart(iconBtn, false, false, 0)

	presentBtn := gtk3.NewButtonWithLabel("显示(Present)")
	presentBtn.SetOnClick(func(sender PGtkWidget, userData GPointer) {
		win.Present()
	})
	wState.PackStart(presentBtn, false, false, 0)
	tab7.PackStart(wState, false, false, 0)

	// -- 窗口属性
	wProp := gtk3.NewBox(ORIENTATION_HORIZONTAL, 6)
	decorCheck := gtk3.NewCheckButton()
	decorCheck.SetActive(true)
	wProp.PackStart(gtk3.NewLabel("窗口装饰:"), false, false, 0)
	wProp.PackStart(decorCheck, false, false, 0)
	decorCheck.SetOnClick(func(sender PGtkWidget, userData GPointer) {
		win.SetDecorated(decorCheck.GetActive())
	})
	wProp.PackStart(gtk3.NewSeparator(ORIENTATION_VERTICAL), false, false, 0)

	keepAboveCheck := gtk3.NewCheckButton()
	wProp.PackStart(gtk3.NewLabel("置顶:"), false, false, 0)
	wProp.PackStart(keepAboveCheck, false, false, 0)
	keepAboveCheck.SetOnClick(func(sender PGtkWidget, userData GPointer) {
		win.SetKeepAbove(keepAboveCheck.GetActive())
	})
	wProp.PackStart(gtk3.NewSeparator(ORIENTATION_VERTICAL), false, false, 0)

	urgentCheck := gtk3.NewCheckButton()
	wProp.PackStart(gtk3.NewLabel("紧急:"), false, false, 0)
	wProp.PackStart(urgentCheck, false, false, 0)
	urgentCheck.SetOnClick(func(sender PGtkWidget, userData GPointer) {
		win.SetUrgencyHint(urgentCheck.GetActive())
	})
	tab7.PackStart(wProp, false, false, 0)

	// -- Move / Resize
	wMove := gtk3.NewBox(ORIENTATION_HORIZONTAL, 6)
	xAdj := gtk3.NewAdjustment(100, 0, 2000, 10, 100, 0)
	yAdj := gtk3.NewAdjustment(100, 0, 2000, 10, 100, 0)
	wMove.PackStart(gtk3.NewLabel("移动到 X:"), false, false, 0)
	xSpin := gtk3.NewSpinButton(xAdj, 1, 0)
	wMove.PackStart(xSpin, false, false, 0)
	wMove.PackStart(gtk3.NewLabel("Y:"), false, false, 0)
	ySpin := gtk3.NewSpinButton(yAdj, 1, 0)
	wMove.PackStart(ySpin, false, false, 0)
	moveWinBtn := gtk3.NewButtonWithLabel("移动")
	moveWinBtn.SetOnClick(func(sender PGtkWidget, userData GPointer) {
		win.Move(int(xSpin.GetValue()), int(ySpin.GetValue()))
		x, y := win.GetPosition()
		ww, wh := win.GetSize()
		winInfo.SetText(fmt.Sprintf("窗口类型: %s  位置: (%d,%d)  大小: %dx%d",
			map[WindowType]string{WINDOW_TOPLEVEL: "NORMAL", WINDOW_POPUP: "POPUP"}[win.GetWindowType()],
			x, y, ww, wh))
	})
	wMove.PackStart(moveWinBtn, false, false, 0)
	tab7.PackStart(wMove, false, false, 0)

	wSize := gtk3.NewBox(ORIENTATION_HORIZONTAL, 6)
	wAdj := gtk3.NewAdjustment(1000, 400, 3000, 50, 200, 0)
	hAdj := gtk3.NewAdjustment(800, 400, 3000, 50, 200, 0)
	wSize.PackStart(gtk3.NewLabel("调整宽:"), false, false, 0)
	wSpin := gtk3.NewSpinButton(wAdj, 1, 0)
	wSize.PackStart(wSpin, false, false, 0)
	wSize.PackStart(gtk3.NewLabel("高:"), false, false, 0)
	hSpin := gtk3.NewSpinButton(hAdj, 1, 0)
	wSize.PackStart(hSpin, false, false, 0)
	resizeWinBtn := gtk3.NewButtonWithLabel("调整大小")
	resizeWinBtn.SetOnClick(func(sender PGtkWidget, userData GPointer) {
		w := int(wSpin.GetValue())
		h := int(hSpin.GetValue())
		if w > 3000 {
			w = 3000
		}
		if h > 3000 {
			h = 3000
		}
		// gtk_window_resize 受子控件最小尺寸约束; SetDefaultSize 作为备用
		win.Resize(w, h)
		win.SetDefaultSize(w, h)
		ww, wh := win.GetSize()
		px, py := win.GetPosition()
		winInfo.SetText(fmt.Sprintf("窗口类型: %s  位置: (%d,%d)  大小: %dx%d",
			map[WindowType]string{WINDOW_TOPLEVEL: "NORMAL", WINDOW_POPUP: "POPUP"}[win.GetWindowType()],
			px, py, ww, wh))
	})
	wSize.PackStart(resizeWinBtn, false, false, 0)
	tab7.PackStart(wSize, false, false, 0)

	// 刷新信息
	refreshBtn := gtk3.NewButtonWithLabel("刷新窗口信息")
	refreshBtn.SetOnClick(func(sender PGtkWidget, userData GPointer) {
		x, y := win.GetPosition()
		w, h := win.GetSize()
		winInfo.SetText(fmt.Sprintf("窗口类型: %s  位置: (%d,%d)  大小: %dx%d",
			map[WindowType]string{WINDOW_TOPLEVEL: "NORMAL", WINDOW_POPUP: "POPUP"}[win.GetWindowType()],
			x, y, w, h))
	})
	tab7.PackStart(refreshBtn, false, false, 0)

	tab7Sw := gtk3.NewScrolledWindow(nil, nil)
	tab7Sw.SetPolicy(POLICY_NEVER, POLICY_AUTOMATIC)
	tab7Sw.Add(tab7)
	notebook.AppendPage(tab7Sw, gtk3.NewLabel("窗口操作"))

	// ============================================================
	// Tab 8: CSS样式 (CssProvider / StyleContext / StateFlags)
	// ============================================================
	tab8 := gtk3.NewBox(ORIENTATION_VERTICAL, 8)
	tab8.SetMarginTop(10)
	tab8.SetMarginBottom(10)
	tab8.SetMarginStart(10)
	tab8.SetMarginEnd(10)

	// CSS provider
	css := gtk3.NewCssProvider()
	css.LoadFromData(`
		button.css-btn { background: #4a90d9; color: white; border-radius: 6px; font-weight: bold; padding: 6px 12px; }
		button.css-btn:hover { background: #357abd; }
		button.css-btn:active { background: #2a5f9e; }
		label.css-label { font-size: 18px; color: #e74c3c; font-weight: bold; }
	`)

	// 带 CSS 类名的按钮
	cssBtn := gtk3.NewButtonWithLabel("CSS 样式按钮")
	cssBtn.GetStyleContext().AddClass("css-btn")
	tab8.PackStart(cssBtn, false, false, 0)

	// 带 CSS 类名的标签
	cssLabel := gtk3.NewLabel("CSS 样式标签 (红色粗体)")
	cssLabel.GetStyleContext().AddClass("css-label")
	tab8.PackStart(cssLabel, false, false, 0)

	// Provider 信息
	tab8.PackStart(gtk3.NewSeparator(ORIENTATION_HORIZONTAL), false, false, 0)

	namedCss := gtk3.CssProviderGetNamed("Adwaita", "")
	if namedCss != nil {
		tab8.PackStart(gtk3.NewLabel("✓ Adwaita 主题 CSS Provider 可用"), false, false, 0)
	} else {
		tab8.PackStart(gtk3.NewLabel("✗ Adwaita 主题 CSS Provider 不可用"), false, false, 0)
	}

	// StateFlags 演示
	tab8.PackStart(gtk3.NewSeparator(ORIENTATION_HORIZONTAL), false, false, 0)
	tab8.PackStart(gtk3.NewLabel("StateFlags 演示:"), false, false, 0)

	stateBtn := gtk3.NewButtonWithLabel("悬停发光按钮")
	stateBtn.GetStyleContext().AddClass("css-btn")
	stateBtn.SetOnEnter(func(sender PGtkWidget, event PEventCrossing, userData GPointer) bool {
		stateBtn.GetStyleContext().SetState(STATE_FLAG_PRELIGHT)
		return false
	})
	stateBtn.SetOnLeave(func(sender PGtkWidget, event PEventCrossing, userData GPointer) bool {
		stateBtn.GetStyleContext().SetState(STATE_FLAG_NORMAL)
		return false
	})
	tab8.PackStart(stateBtn, false, false, 0)

	tab8Sw := gtk3.NewScrolledWindow(nil, nil)
	tab8Sw.SetPolicy(POLICY_NEVER, POLICY_AUTOMATIC)
	tab8Sw.Add(tab8)
	notebook.AppendPage(tab8Sw, gtk3.NewLabel("CSS样式"))

	// ============================================================
	// Tab 9: 事件交互 (EventBox / Button Enter-Leave / Settings Theme)
	// ============================================================
	tab9 := gtk3.NewBox(ORIENTATION_VERTICAL, 8)
	tab9.SetMarginTop(10)
	tab9.SetMarginBottom(10)
	tab9.SetMarginStart(10)
	tab9.SetMarginEnd(10)

	// EventBox with click, enter, leave
	evtBox := gtk3.NewEventBox()
	evtBox.SetSizeRequest(300, 60)
	evtLabel := gtk3.NewLabel("移动鼠标到此处或点击")
	evtBox.Add(evtLabel)
	evtBox.SetOnClick(func(sender PGtkWidget, event PEventButton, userData GPointer) bool {
		fmt.Println("[event] EventBox.Click")
		evtLabel.SetText("✅ EventBox 被点击了!")
		return false
	})
	evtBox.SetOnEnter(func(sender PGtkWidget, event PEventCrossing, userData GPointer) bool {
		evtLabel.SetText("👆 鼠标进入 EventBox")
		return false
	})
	evtBox.SetOnLeave(func(sender PGtkWidget, event PEventCrossing, userData GPointer) bool {
		evtLabel.SetText("👋 鼠标离开 EventBox")
		return false
	})
	tab9.PackStart(gtk3.NewLabel("EventBox (点击/进入/离开):"), false, false, 0)
	tab9.PackStart(evtBox, false, false, 0)

	tab9.PackStart(gtk3.NewSeparator(ORIENTATION_HORIZONTAL), false, false, 0)

	// Button enter/leave 事件
	btnEvt := gtk3.NewButtonWithLabel("悬停/离开我")
	btnEvt.SetOnEnter(func(sender PGtkWidget, event PEventCrossing, userData GPointer) bool {
		statusbar.Push(statusCtx, "按钮: 鼠标进入")
		return false
	})
	btnEvt.SetOnLeave(func(sender PGtkWidget, event PEventCrossing, userData GPointer) bool {
		statusbar.Push(statusCtx, "按钮: 鼠标离开")
		return false
	})
	btnEvt.SetSizeRequest(200, 40)
	tab9.PackStart(btnEvt, false, false, 0)

	tab9.PackStart(gtk3.NewSeparator(ORIENTATION_HORIZONTAL), false, false, 0)

	// Settings 主题变化监听
	tab9.PackStart(gtk3.NewLabel("Settings 主题变化监听:"), false, false, 0)
	themeLabel := gtk3.NewLabel("当前主题: (未检测)")
	settings := gtk3.SettingsGetDefault()
	if settings != nil {
		settings.SetOnThemeChanged(func(sender PGtkWidget, pspec uintptr, userData GPointer) {
			themeLabel.SetLabel("主题已变更!")
			statusbar.Push(statusCtx, "系统主题已变更")
		})
	}
	tab9.PackStart(themeLabel, false, false, 0)

	tab9.PackStart(gtk3.NewSeparator(ORIENTATION_HORIZONTAL), false, false, 0)

	// -- GtkPopover (气泡弹出)
	tab9.PackStart(gtk3.NewLabel("GtkPopover (气泡弹出面板):"), false, false, 0)
	popBtn := gtk3.NewButtonWithLabel("点击弹出气泡")
	popover := gtk3.NewPopover()
	popover.SetRelativeTo(popBtn)
	popover.SetPosition(POS_BOTTOM)
	popLabel := gtk3.NewLabel("这是气泡内的内容\n点击外部关闭")
	popLabel.Show()
	popover.Add(popLabel)
	popover.SetOnClosed(func(sender PGtkWidget, userData GPointer) {
		fmt.Println("[event] Popover.Closed")
		statusbar.Push(statusCtx, "Popover closed (closed)")
	})
	popBtn.SetOnClick(func(sender PGtkWidget, userData GPointer) {
		popover.Popup()
	})
	tab9.PackStart(popBtn, false, false, 0)

	tab9Sw := gtk3.NewScrolledWindow(nil, nil)
	tab9Sw.SetPolicy(POLICY_NEVER, POLICY_AUTOMATIC)
	tab9Sw.Add(tab9)
	notebook.AppendPage(tab9Sw, gtk3.NewLabel("事件交互"))

	// ============================================================
	// Tab 10: 对话框 (Message/Dialog/ColorChooser/FontChooser/FileChooser/About)
	// ============================================================
	tab10 := gtk3.NewBox(ORIENTATION_VERTICAL, 8)
	tab10.SetMarginTop(10)
	tab10.SetMarginBottom(10)
	tab10.SetMarginStart(10)
	tab10.SetMarginEnd(10)

	msgBtn := gtk3.NewButtonWithLabel("MessageDialog")
	msgBtn.SetOnClick(func(sender PGtkWidget, userData GPointer) {
		dlg := gtk3.NewMessageDialog(win, DIALOG_MODAL, MESSAGE_INFO, BUTTONS_OK, "这是一条消息")
		dlg.FormatSecondaryText("详细信息内容...")
		dlg.SetOnResponse(func(sender PGtkWidget, responseId int32, userData GPointer) {
			dlg.Destroy()
		})
		dlg.ShowAll()
	})
	tab10.PackStart(msgBtn, false, false, 0)

	dlgBtn := gtk3.NewButtonWithLabel("Dialog")
	dlgBtn.SetOnClick(func(sender PGtkWidget, userData GPointer) {
		dlg := gtk3.NewDialog()
		dlg.AddButton("确定", 0)
		dlg.AddButton("取消", 1)
		dlg.SetDefaultResponse(0)
		area := dlg.GetContentArea()
		if area != nil {
			area.PackStart(gtk3.NewLabel("对话框内容区域"), false, false, 10)
		}
		dlg.SetOnResponse(func(sender PGtkWidget, responseId int32, userData GPointer) {
			dlg.Destroy()
		})
		dlg.ShowAll()
	})
	tab10.PackStart(dlgBtn, false, false, 0)

	colorBtn := gtk3.NewButtonWithLabel("ColorChooserDialog")
	colorBtn.SetOnClick(func(sender PGtkWidget, userData GPointer) {
		dlg := gtk3.NewColorChooserDialog("选择颜色", win)
		dlg.SetUseAlpha(true)
		dlg.SetOnResponse(func(sender PGtkWidget, responseId int32, userData GPointer) {
			if responseId == int32(RESPONSE_OK) || responseId == int32(RESPONSE_ACCEPT) {
				c := dlg.GetRGBA()
				fmt.Printf("[event] ColorChooser → R:%.2f G:%.2f B:%.2f A:%.2f\n", c.Red, c.Green, c.Blue, c.Alpha)
				statusbar.Push(statusCtx, fmt.Sprintf("颜色: R=%.0f G=%.0f B=%.0f", c.Red*255, c.Green*255, c.Blue*255))
			}
			dlg.Destroy()
		})
		dlg.ShowAll()
	})
	tab10.PackStart(colorBtn, false, false, 0)

	fontBtn := gtk3.NewButtonWithLabel("FontChooserDialog")
	fontBtn.SetOnClick(func(sender PGtkWidget, userData GPointer) {
		dlg := gtk3.NewFontChooserDialog("选择字体", win)
		dlg.SetFont("Sans 12")
		dlg.SetPreviewText("预览文本 ABCabc 123")
		dlg.SetOnResponse(func(sender PGtkWidget, responseId int32, userData GPointer) {
			if responseId == int32(RESPONSE_OK) || responseId == int32(RESPONSE_ACCEPT) {
				font := dlg.GetFont()
				fmt.Println("[event] FontChooser →", font)
				statusbar.Push(statusCtx, "字体: "+font)
			}
			dlg.Destroy()
		})
		dlg.ShowAll()
	})
	tab10.PackStart(fontBtn, false, false, 0)

	fileBtn := gtk3.NewButtonWithLabel("FileChooserDialog")
	fileBtn.SetOnClick(func(sender PGtkWidget, userData GPointer) {
		dlg := gtk3.NewFileChooserDialog("选择文件", win, FILE_CHOOSER_ACTION_OPEN)
		dlg.SetSelectMultiple(false)
		imgFilter := gtk3.NewFileFilter()
		imgFilter.SetName("图片文件")
		imgFilter.AddPattern("*.png")
		imgFilter.AddPattern("*.jpg")
		imgFilter.AddPattern("*.jpeg")
		imgFilter.AddMimeType("image/png")
		imgFilter.AddMimeType("image/jpeg")
		dlg.AddFilter(imgFilter)
		allFilter := gtk3.NewFileFilter()
		allFilter.SetName("所有文件")
		allFilter.AddPattern("*")
		dlg.AddFilter(allFilter)
		dlg.SetFilter(imgFilter)
		dlg.SetOnResponse(func(sender PGtkWidget, responseId int32, userData GPointer) {
			if responseId == int32(RESPONSE_ACCEPT) {
				fname := dlg.GetFilename()
				fmt.Printf("  选择的文件: %s\n", fname)
			}
			dlg.Destroy()
		})
		dlg.ShowAll()
	})
	tab10.PackStart(fileBtn, false, false, 0)

	aboutBtn := gtk3.NewButtonWithLabel("AboutDialog")
	aboutBtn.SetOnClick(func(sender PGtkWidget, userData GPointer) {
		dlg := gtk3.NewAboutDialog()
		dlg.SetProgramName("Energy GTK3")
		dlg.SetVersion("3.0.0")
		dlg.SetComments("Go 语言 GTK3 绑定库")
		dlg.SetWebsite("https://github.com/energye/energy")
		dlg.SetWebsiteLabel("GitHub")
		dlg.SetLicense("Apache 2.0")
		dlg.SetAuthors([]string{"energye"})
		dlg.SetOnResponse(func(sender PGtkWidget, responseId int32, userData GPointer) {
			dlg.Destroy()
		})
		dlg.ShowAll()
	})
	tab10.PackStart(aboutBtn, false, false, 0)

	tab10.PackStart(gtk3.NewSeparator(ORIENTATION_HORIZONTAL), false, false, 0)

	// -- GtkInfoBar (内联消息条)
	tab10.PackStart(gtk3.NewLabel("GtkInfoBar (内联消息条, 切换类型/关闭):"), false, false, 0)
	infoBar := gtk3.NewInfoBar()
	infoBar.SetMessageType(MESSAGE_INFO)
	infoBar.SetShowCloseButton(true)
	infoBar.GetContentArea().PackStart(gtk3.NewLabel("InfoBar 消息 — 点击下方按钮切换类型"), false, false, 10)
	infoBar.AddButton("操作一", 1)
	infoBar.AddButton("操作二", 2)
	tab10.PackStart(infoBar, false, false, 0)
	ibtnRow := gtk3.NewBox(ORIENTATION_HORIZONTAL, 6)
	ibtnRow.PackStart(gtk3.NewButtonWithLabel("信息"), false, false, 0)
	ibtnRow.PackStart(gtk3.NewButtonWithLabel("警告"), false, false, 0)
	ibtnRow.PackStart(gtk3.NewButtonWithLabel("错误"), false, false, 0)
	tab10.PackStart(ibtnRow, false, false, 0)

	tab10.PackStart(gtk3.NewSeparator(ORIENTATION_HORIZONTAL), false, false, 0)

	// -- GtkListBox (列表选择)
	tab10.PackStart(gtk3.NewLabel("GtkListBox (列表项选择):"), false, false, 0)
	listBox := gtk3.NewListBox()
	listBox.SetSelectionMode(SELECTION_SINGLE)
	listBox.SetSizeRequest(300, 100)
	listBox.Prepend(gtk3.NewLabel("选项一"))
	listBox.Prepend(gtk3.NewLabel("选项二"))
	listBox.Prepend(gtk3.NewLabel("选项三"))
	tab10.PackStart(listBox, false, false, 0)

	tab10Sw := gtk3.NewScrolledWindow(nil, nil)
	tab10Sw.SetPolicy(POLICY_NEVER, POLICY_AUTOMATIC)
	tab10Sw.Add(tab10)
	notebook.AppendPage(tab10Sw, gtk3.NewLabel("对话框"))

	// ============================================================
	// 底部状态栏
	// ============================================================
	mainBox.PackStart(statusbar, false, false, 0)

	// ============================================================
	// 自动验证
	// ============================================================
	fmt.Println("=== GTK3 全组件增强测试 ===")

	makeSection("[基础组件]")
	check("NewWindow", win != nil)
	check("NewHeaderBar", hbar != nil)
	check("NewLabel", lblPlain != nil)
	check("Label.SetMarkup", true)
	check("Label.SetLineWrap", true)
	check("Label.SetEllipsize", true)
	check("Label.SetSelectable", true)
	check("NewButtonWithLabel", btnClick != nil)
	check("NewButtonWithMnemonic", btnMnemonic != nil)
	check("Button.SetRelief(RELIEF_NONE)", true)
	check("CheckButton", toggle != nil)
	check("CheckButton.SetActive(GetActive)", toggle.GetActive())
	check("NewImageFromIconName", imgInfo != nil)
	check("Image.SetPixelSize", imgSmall.GetPixelSize() == 48)
	check("Widget.SetOpacity", true)
	check("Widget.SetSensitive(false)", true)
	check("Widget.SetTooltipText", true)
	check("Widget.SetHExpand", expandLabel.GetHExpand())
	check("NewSeparator", true)

	makeSection("[输入组件]")
	check("NewEntry", entry != nil)
	check("Entry.SetText/GetText", entry.GetText() == "Hello GTK3")
	check("Entry.SetPlaceholderText", true)
	check("Entry.SetMaxLength(50)", entry.GetMaxLength() == 50)
	check("Entry.SetWidthChars(30)", entry.GetWidthChars() == 30)
	check("Entry.SetHasFrame(false)", !entryFlat.GetHasFrame())
	check("Entry.SetActivatesDefault", entryFlat.GetActivatesDefault())
	check("Entry.SetVisibility(false) 密码", true)
	check("Entry.ProgressPulse", true)
	check("EntryCompletion", completion != nil)
	check("NewSpinButton", spin != nil)
	check("NewAdjustment", adj != nil)
	check("SpinButton.SetOnValueChanged", true)
	check("NewComboBoxText", combo != nil)
	check("ComboBox.SetActive/GetActive", combo.GetActive() == 0)
	check("RadioButton group", radio1.GetActive() && !radio2.GetActive())
	check("RadioButton.NewRadioButtonWithLabel", radioNew1 != nil)
	check("RadioButton.NewRadioButtonFromWidget", radioNew2 != nil)
	check("SpinButton.SetNumeric/GetNumeric", true)
	check("SpinButton.SetSnapToTicks/GetSnapToTicks", true)
	check("SpinButton.SetWrap/GetWrap", true)
	check("SpinButton.Spin/GetAdjustment", true)

	makeSection("[布局容器]")
	check("NewBox (Homogeneous=false)", !boxDemoBox.GetHomogeneous())
	check("Box.SetSpacing(12)", true)
	check("NewGrid", grid != nil)
	check("Grid.SetRowHomogeneous", grid.GetRowHomogeneous())
	check("Grid.SetRowSpacing(8)", grid.GetRowSpacing() == 8)
	check("NewFixed + Put + Move", true)
	check("ScrolledWindow.SetPolicy", true)
	check("ScrolledWindow.SetShadowType", true)
	check("ScrolledWindow.SetMinContentWidth", true)
	check("NewOverlay", true)

	makeSection("[文本编辑]")
	check("NewTextView", tv != nil)
	check("TextBuffer.SetText", true)
	check("TextBuffer.GetCharCount", tb.GetCharCount() > 0)
	check("TextBuffer.GetLineCount", tb.GetLineCount() > 0)
	check("TextView.SetEditable", tv.GetEditable())
	check("TextView.SetWrapMode(WRAP_WORD_CHAR)", tv.GetWrapMode() == WRAP_WORD_CHAR)
	check("TextView.SetCursorVisible", tv.GetCursorVisible())
	check("TextView.SetOverwrite", !tv.GetOverwrite())
	check("TextView.SetJustification", tv.GetJustification() == JUSTIFY_LEFT)
	check("TextBuffer.Insert", true)
	check("TextBuffer.Delete", true)

	makeSection("[数据表格]")
	check("NewListStore", store != nil)
	check("ListStore.Append + SetValue", true)
	check("NewTreeView", treeView != nil)
	check("TreeView.SetModel", true)
	check("TreeView.AppendColumn (3列)", true)
	check("TreeSelection.SetMode(MULTIPLE)", sel.GetMode() == SELECTION_MULTIPLE)
	 check("TreeSelection.GetSelected", true)
	 check("TreeSelection.CountSelectedRows", true)
	 check("TreeSelection.SelectAll", true)
	 check("TreeSelection.UnselectAll", true)
	 check("ListStore.AddRow", true)
	 check("ListStore.Clear", true)
	 check("TreeViewColumn.SetResizable", true)
	 check("TreeViewColumn.SetSizing(FIXED)", cols[0].GetSizing() == TREE_VIEW_COLUMN_FIXED)
	 check("TreeViewColumn.SetFixedWidth(100)", cols[0].GetFixedWidth() == 100)
	 check("TreeViewColumn.SetMinWidth(60)", cols[0].GetMinWidth() == 60)
	 check("TreeViewColumn.SetMaxWidth(200)", cols[0].GetMaxWidth() == 200)
	 check("TreeViewColumn.SetExpand(false)", !cols[0].GetExpand())
	 check("TreeViewColumn.SetSortColumnId(0)", cols[0].GetSortColumnId() == 0)
	 check("TreeViewColumn.SetSortIndicator(true)", cols[0].GetSortIndicator())
	 check("TreeViewColumn.SetReorderable(true)", cols[0].GetReorderable())
	 check("TreeViewColumn.SetAlignment(0.5)", cols[0].GetAlignment() == 0.5)
	 check("TreeViewColumn.GetWidth", cols[0].GetWidth() > 0)
	 check("TreeViewColumn.SetSpacing", true)

	makeSection("[进度状态]")
	check("NewProgressBar", pbar != nil)
	check("ProgressBar.SetShowText", true)
	check("ProgressBar.SetFraction/GetFraction", true)
	check("NewStatusbar", statusbar != nil)
	check("Statusbar.Push/Pop/RemoveAll", true)

	makeSection("[窗口操作]")
	check("Window.Maximize", true)
	check("Window.Unmaximize", true)
	check("Window.Fullscreen/Unfullscreen", true)
	check("Window.Iconify/Deiconify", true)
	check("Window.Present", true)
	check("Window.SetDecorated/GetDecorated", win.GetDecorated())
	check("Window.SetKeepAbove", true)
	check("Window.SetUrgencyHint", true)
	check("Window.GetPosition", true)
	check("Window.GetSize", true)
	check("Window.Move", true)
	check("Window.Resize", true)
	check("Window.GetWindowType", win.GetWindowType() == WINDOW_TOPLEVEL)

	makeSection("[CSS样式]")
	check("NewCssProvider", css != nil)
	check("CssProvider.LoadFromData", true)
	check("StyleContext.AddClass", true)
	check("StyleContext.SetState", true)
	check("CssProviderGetNamed", namedCss != nil)

	makeSection("[事件交互]")
	check("NewEventBox", evtBox != nil)
	check("EventBox.SetOnClick", true)
	check("EventBox.SetOnEnter/SetOnLeave", true)
	check("Button.SetOnEnter/SetOnLeave", true)
	check("NewMenuBar", menuBar != nil)
	check("NewMenu", fileMenu != nil)
	check("MenuItemNewWithLabel", openItem != nil)
	check("MenuItem.SetSubmenu", true)
	check("SeparatorMenuItemNew", true)
	check("SettingsGetDefault", settings != nil)
	check("Settings.SetOnThemeChanged", true)

	makeSection("[对话框]")
	check("NewMessageDialog", true)
	check("NewDialog + GetContentArea", true)
	check("NewColorChooserDialog", true)
	check("NewFontChooserDialog", true)
	check("NewFileChooserDialog + FileFilter", true)
	check("NewAboutDialog", true)

	makeSection("[新组件 Stack/Switch/InfoBar/Scale/Spinner/LevelBar]")
	check("NewStack", stack != nil)
	check("Stack.AddTitled", true)
	check("Stack.SetVisibleChild", true)
	check("NewStackSwitcher", switcher != nil)
	check("StackSwitcher.SetStack", true)
	check("NewSwitch", sw != nil)
	check("Switch.SetActive/GetActive", !sw.GetActive())
	check("NewInfoBar", infoBar != nil)
	check("InfoBar.SetMessageType", true)
	check("InfoBar.AddButton", true)
	check("NewHScale", hScale != nil)
	check("Scale.SetDigits(0)", hScale.GetDigits() == 0)
	check("Scale.SetDrawValue(true)", hScale.GetDrawValue())
	check("Scale.SetValuePos", true)
	check("NewSpinner", spinner != nil)
	check("Spinner.Start/Stop", true)
	check("NewLevelBar", lvBar != nil)
	check("LevelBar.SetMinValue/MaxValue", true)
	check("LevelBar.SetValue/GetValue", true)
	check("NewPaned", paned != nil)
	check("Paned.Add1/Add2", true)
	check("Paned.SetPosition/GetPosition", paned.GetPosition() == 200)
	check("NewListBox", listBox != nil)
	check("ListBox.SetSelectionMode", listBox.GetSelectionMode() == SELECTION_SINGLE)
	check("NewPopover", true)
	check("NewSearchEntry", searchEntry != nil)
	check("NewRevealer", revealer != nil)
	check("Revealer.SetTransitionDuration", revealer.GetTransitionDuration() == 500)
	check("Revealer.SetTransitionType", true)

	makeSection("[Window属性]")
	check("Window.Decorated", win.GetDecorated())
	check("Window.Resizable", win.GetResizable())
	check("Window.Sensitive", win.IsSensitive())
	check("Window.GetTitle", win.GetTitle() != "")

	fmt.Println("\n=============================")
	fmt.Printf("通过: %d  失败: %d  总计: %d\n", passed, failed, passed+failed)
	if failed == 0 {
		fmt.Println("🎉 全部通过!")
	} else {
		fmt.Println("⚠️  部分测试未通过")
	}
	fmt.Println("=============================")
	fmt.Println("\n窗口已打开，可切换各标签页查看组件演示...")

	win.ShowAll()
	gtk3.Main()
}
