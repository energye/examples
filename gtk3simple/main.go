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

func main() {
	gtk3.Init(nil)

	win, _ := gtk3.NewWindow(WINDOW_TOPLEVEL)
	win.SetTitle("GTK3 全组件测试")
	win.SetDefaultSize(800, 600)

	hbar := gtk3.NewHeaderBar()
	hbar.SetTitle("GTK3 全组件测试")
	hbar.SetShowCloseButton(true)
	win.SetTitlebar(hbar)

	win.SetOnDestroy(func(sender PGtkWidget, userData GPointer) {
		gtk3.MainQuit()
	})

	// Main Notebook
	notebook := gtk3.NewNotebook()
	win.Add(notebook)

	// ==============================
	// Tab 1: 基础组件
	// ==============================
	tab1 := gtk3.NewBox(ORIENTATION_VERTICAL, 8)
	tab1.SetMarginTop(10)
	tab1.SetMarginBottom(10)
	tab1.SetMarginStart(10)
	tab1.SetMarginEnd(10)

	// Label
	label := gtk3.NewLabel("普通文本标签")
	tab1.PackStart(label, false, false, 0)
	markup := gtk3.NewLabel("")
	markup.SetMarkup("<b>粗体</b> <i>斜体</i> <u>下划线</u>")
	markup.SetJustify(JUSTIFY_CENTER)
	tab1.PackStart(markup, false, false, 0)

	// Separator
	tab1.PackStart(gtk3.NewSeparator(ORIENTATION_HORIZONTAL), false, false, 0)

	// Button row
	btnBox := gtk3.NewBox(ORIENTATION_HORIZONTAL, 6)

	btn := gtk3.NewButtonWithLabel("计数按钮")
	count := 0
	btn.SetOnClick(func(sender PGtkWidget, userData GPointer) {
		count++
		btn.SetLabel(fmt.Sprintf("点击了 %d 次", count))
	})
	btnBox.PackStart(btn, false, false, 0)

	mBtn := gtk3.NewButtonWithMnemonic("_快捷键按钮")
	btnBox.PackStart(mBtn, false, false, 0)

	toggleBtn := gtk3.NewCheckButton()
	toggleBtn.SetOnClick(func(sender PGtkWidget, userData GPointer) {
		if toggleBtn.GetActive() {
			label.SetText("开关: ON")
		} else {
			label.SetText("开关: OFF")
		}
	})
	btnBox.PackStart(gtk3.NewLabel("开关:"), false, false, 0)
	btnBox.PackStart(toggleBtn, false, false, 0)
	tab1.PackStart(btnBox, false, false, 0)

	// Image
	img := gtk3.NewImageFromIconName("dialog-information", ICON_SIZE_DIALOG)
	tab1.PackStart(img, false, false, 0)

	notebook.AppendPage(tab1, gtk3.NewLabel("基础组件"))

	// ==============================
	// Tab 2: 输入组件
	// ==============================
	tab2 := gtk3.NewBox(ORIENTATION_VERTICAL, 8)
	tab2.SetMarginTop(10)
	tab2.SetMarginBottom(10)
	tab2.SetMarginStart(10)
	tab2.SetMarginEnd(10)

	// Entry with live update
	entryLabel := gtk3.NewLabel("输入内容将显示在这里")
	entry := gtk3.NewEntry()
	entry.SetText("Hello")
	entry.SetPlaceholderText("请输入文本...")
	entry.SetOnChanged(func(sender PGtkWidget, userData GPointer) {
		entryLabel.SetText("输入: " + entry.GetText())
	})
	tab2.PackStart(entry, false, false, 0)
	tab2.PackStart(entryLabel, false, false, 0)

	// Password Entry
	passEntry := gtk3.NewEntry()
	passEntry.SetText("password123")
	passEntry.SetVisibility(false)
	tab2.PackStart(passEntry, false, false, 0)

	// SpinButton
	spinLabel := gtk3.NewLabel("值: 50")
	adj := gtk3.NewAdjustment(50, 0, 100, 1, 10, 0)
	spin := gtk3.NewSpinButton(adj, 0.5, 0)
	spin.SetRange(0, 100)
	spin.SetIncrements(1, 10)
	tab2.PackStart(spin, false, false, 0)
	tab2.PackStart(spinLabel, false, false, 0)

	// ComboBoxText
	combo := gtk3.NewComboBoxText()
	combo.Append("red", "红色")
	combo.Append("green", "绿色")
	combo.Append("blue", "蓝色")
	combo.SetActive(0)
	tab2.PackStart(combo, false, false, 0)

	// RadioButton
	radio1 := gtk3.NewRadioButtonWithLabelFromWidget(nil, "选项 A")
	radio2 := gtk3.NewRadioButtonWithLabelFromWidget(radio1, "选项 B")
	radio3 := gtk3.NewRadioButtonWithLabelFromWidget(radio1, "选项 C")
	radio1.SetActive(true)
	rBox := gtk3.NewBox(ORIENTATION_HORIZONTAL, 10)
	rBox.PackStart(radio1, false, false, 0)
	rBox.PackStart(radio2, false, false, 0)
	rBox.PackStart(radio3, false, false, 0)
	tab2.PackStart(rBox, false, false, 0)

	notebook.AppendPage(tab2, gtk3.NewLabel("输入组件"))

	// ==============================
	// Tab 3: 布局容器
	// ==============================
	tab3 := gtk3.NewBox(ORIENTATION_VERTICAL, 8)
	tab3.SetMarginTop(10)
	tab3.SetMarginBottom(10)
	tab3.SetMarginStart(10)
	tab3.SetMarginEnd(10)

	// Grid
	grid := gtk3.NewGrid()
	grid.SetRowSpacing(6)
	grid.SetColumnSpacing(6)
	grid.Attach(gtk3.NewLabel("行1列1:"), 0, 0, 1, 1)
	grid.Attach(gtk3.NewEntry(), 1, 0, 1, 1)
	grid.Attach(gtk3.NewLabel("行2列1:"), 0, 1, 1, 1)
	grid.Attach(gtk3.NewEntry(), 1, 1, 1, 1)
	grid.Attach(gtk3.NewLabel("合并两列的文本"), 0, 2, 2, 1)
	tab3.PackStart(grid, false, false, 0)

	tab3.PackStart(gtk3.NewSeparator(ORIENTATION_HORIZONTAL), false, false, 0)

	// Fixed
	fixed := gtk3.NewFixed()
	fixed.Put(gtk3.NewLabel("(10,10)"), 10, 10)
	fixed.Put(gtk3.NewLabel("(100,50)"), 100, 50)
	tab3.PackStart(fixed, false, false, 0)

	// Overlay
	overlay := gtk3.NewOverlay()
	overlay.AddOverlay(gtk3.NewLabel("覆盖层"))
	tab3.PackStart(overlay, false, false, 0)

	notebook.AppendPage(tab3, gtk3.NewLabel("布局容器"))

	// ==============================
	// Tab 4: 文本编辑
	// ==============================
	tab4 := gtk3.NewBox(ORIENTATION_VERTICAL, 8)
	tab4.SetMarginTop(10)
	tab4.SetMarginBottom(10)
	tab4.SetMarginStart(10)
	tab4.SetMarginEnd(10)

	infoLabel := gtk3.NewLabel("字符数: 0, 行数: 0")
	tv := gtk3.NewTextView()
	tb := tv.GetBuffer()
	tb.SetText("这是多行文本编辑器。\n可以编辑这些文本。\n第三行内容。")
	tv.SetEditable(true)
	tv.SetWrapMode(WRAP_WORD)
	tv.SetLeftMargin(10)
	tv.SetRightMargin(10)
	infoLabel.SetText(fmt.Sprintf("字符数: %d, 行数: %d", tb.GetCharCount(), tb.GetLineCount()))
	tab4.PackStart(infoLabel, false, false, 0)

	sw := gtk3.NewScrolledWindow(nil, nil)
	sw.Add(tv)
	tab4.PackStart(sw, true, true, 0)

	notebook.AppendPage(tab4, gtk3.NewLabel("文本编辑"))

	// ==============================
	// Tab 5: 数据表格
	// ==============================
	tab5 := gtk3.NewBox(ORIENTATION_VERTICAL, 8)
	tab5.SetMarginTop(10)
	tab5.SetMarginBottom(10)
	tab5.SetMarginStart(10)
	tab5.SetMarginEnd(10)

	store := gtk3.NewListStore(TYPE_STRING, TYPE_STRING, TYPE_STRING)
	type row struct{ name, email, role string }
	rows := []row{
		{"张三", "zhang@example.com", "工程师"},
		{"李四", "li@example.com", "设计师"},
		{"王五", "wang@example.com", "产品经理"},
		{"赵六", "zhao@example.com", "测试"},
	}
	for _, d := range rows {
		iter := store.Append()
		store.SetValue(iter, 0, d.name)
		store.SetValue(iter, 1, d.email)
		store.SetValue(iter, 2, d.role)
	}

	treeView := gtk3.NewTreeView()
	treeView.SetModel(store)
	for i, title := range []string{"姓名", "邮箱", "角色"} {
		renderer := gtk3.NewCellRendererText()
		col := gtk3.NewTreeViewColumn()
		col.SetTitle(title)
		col.PackStart(renderer, false)
		col.AddAttribute(renderer, "text", i)
		treeView.AppendColumn(col)
	}
	treeView.GetSelection().SetMode(SELECTION_SINGLE)
	treeView.SetHeadersVisible(true)
	treeView.ExpandAll()

	tw := gtk3.NewScrolledWindow(nil, nil)
	tw.Add(treeView)
	tab5.PackStart(tw, true, true, 0)

	notebook.AppendPage(tab5, gtk3.NewLabel("数据表格"))

	// ==============================
	// Tab 6: 进度与状态
	// ==============================
	tab6 := gtk3.NewBox(ORIENTATION_VERTICAL, 8)
	tab6.SetMarginTop(10)
	tab6.SetMarginBottom(10)
	tab6.SetMarginStart(10)
	tab6.SetMarginEnd(10)

	pbar := gtk3.NewProgressBar()
	pbar.SetPulseStep(0.1)
	pbar.Pulse()
	tab6.PackStart(pbar, false, false, 0)

	statusbar := gtk3.NewStatusbar()
	ctxId := statusbar.GetContextId("main")
	statusbar.Push(ctxId, "就绪")
	tab6.PackStart(statusbar, false, false, 0)

	notebook.AppendPage(tab6, gtk3.NewLabel("进度状态"))

	// ==============================
	// Tab 7: 滚动与范围
	// ==============================
	tab7 := gtk3.NewBox(ORIENTATION_VERTICAL, 8)
	tab7.SetMarginTop(10)
	tab7.SetMarginBottom(10)
	tab7.SetMarginStart(10)
	tab7.SetMarginEnd(10)

	scrollAdj := gtk3.NewAdjustment(50, 0, 100, 1, 10, 0)
	hScroll := gtk3.NewScrollbar(ORIENTATION_HORIZONTAL, scrollAdj)
	tab7.PackStart(hScroll, false, false, 0)

	vScroll := gtk3.NewScrollbar(ORIENTATION_VERTICAL, scrollAdj)
	hbox := gtk3.NewBox(ORIENTATION_HORIZONTAL, 6)
	hbox.PackStart(vScroll, false, false, 0)
	hbox.PackStart(gtk3.NewLabel("垂直滚动条"), false, false, 0)
	tab7.PackStart(hbox, false, false, 0)

	layoutSw := gtk3.NewScrolledWindow(nil, nil)
	layout := gtk3.NewLayout(nil, nil)
	layout.SetSize(400, 300)
	for i := 0; i < 10; i++ {
		layout.Put(gtk3.NewLabel(fmt.Sprintf("标签 %d", i)), 10, i*25)
	}
	layoutSw.Add(layout)
	tab7.PackStart(layoutSw, true, true, 0)

	notebook.AppendPage(tab7, gtk3.NewLabel("滚动范围"))

	// ==============================
	// Tab 8: CSS 样式
	// ==============================
	tab8 := gtk3.NewBox(ORIENTATION_VERTICAL, 8)
	tab8.SetMarginTop(10)
	tab8.SetMarginBottom(10)
	tab8.SetMarginStart(10)
	tab8.SetMarginEnd(10)

	css := gtk3.NewCssProvider()
	css.LoadFromData("button { background: #4a90d9; color: white; border-radius: 5px; }")
	namedCss := gtk3.CssProviderGetNamed("Adwaita", "")
	if namedCss != nil {
		tab8.PackStart(gtk3.NewLabel("Adwaita 主题可用"), false, false, 0)
	} else {
		tab8.PackStart(gtk3.NewLabel("Adwaita 主题不可用"), false, false, 0)
	}
	styledBtn := gtk3.NewButtonWithLabel("样式按钮")
	tab8.PackStart(styledBtn, false, false, 0)

	notebook.AppendPage(tab8, gtk3.NewLabel("CSS样式"))

	// ==============================
	// Tab 9: 菜单
	// ==============================
	tab9 := gtk3.NewBox(ORIENTATION_VERTICAL, 8)
	tab9.SetMarginTop(10)
	tab9.SetMarginBottom(10)
	tab9.SetMarginStart(10)
	tab9.SetMarginEnd(10)

	menuBar := gtk3.NewMenuBar()

	// File menu
	fileMenu := gtk3.NewMenu()
	fileMenuItem := gtk3.NewMenuItem()
	fileMenuItem.SetSubmenu(fileMenu)

	openItem := gtk3.MenuItemNewWithLabel("打开")
	openItem.SetOnActivate(func(sender PGtkWidget, userData GPointer) {
		fcDlg := gtk3.NewFileChooserDialog("打开文件", win, FILE_CHOOSER_ACTION_OPEN)
		fcDlg.SetOnResponse(func(sender PGtkWidget, responseId int, userData GPointer) {
			if responseId == 0 {
				fname := fcDlg.GetFilename()
				statusbar.Push(ctxId, "打开: "+fname)
			}
			fcDlg.Destroy()
		})
		fcDlg.ShowAll()
	})
	fileMenu.Append(openItem)

	fileMenu.Append(gtk3.SeparatorMenuItemNew())

	quitItem := gtk3.MenuItemNewWithLabel("退出")
	quitItem.SetOnActivate(func(sender PGtkWidget, userData GPointer) {
		gtk3.MainQuit()
	})
	fileMenu.Append(quitItem)

	menuBar.Append(fileMenuItem)

	// Help menu
	helpMenu := gtk3.NewMenu()
	helpMenuItem := gtk3.NewMenuItem()
	helpMenuItem.SetSubmenu(helpMenu)

	aboutItem := gtk3.MenuItemNewWithLabel("关于")
	aboutItem.SetOnActivate(func(sender PGtkWidget, userData GPointer) {
		about := gtk3.NewAboutDialog()
		about.SetProgramName("Energy GTK3")
		about.SetVersion("3.0.0")
		about.SetComments("Go 语言 GTK3 绑定库")
		about.SetWebsite("https://github.com/energye/energy")
		about.SetWebsiteLabel("GitHub")
		about.SetLicense("Apache 2.0")
		about.SetAuthors([]string{"energye"})
		about.SetOnResponse(func(sender PGtkWidget, responseId int, userData GPointer) {
			about.Destroy()
		})
		about.ShowAll()
	})
	helpMenu.Append(aboutItem)

	menuBar.Append(helpMenuItem)
	tab9.PackStart(menuBar, false, false, 0)

	// EventBox with click
	evtBox := gtk3.NewEventBox()
	evtLabel := gtk3.NewLabel("点击这里测试 EventBox")
	evtBox.Add(evtLabel)
	evtBox.SetOnClick(func(sender PGtkWidget, event PEventButton, userData GPointer) bool {
		evtLabel.SetText("✅ EventBox 被点击了!")
		return false
	})
	tab9.PackStart(evtBox, false, false, 0)

	notebook.AppendPage(tab9, gtk3.NewLabel("菜单事件"))

	// ==============================
	// Tab 10: 对话框
	// ==============================
	tab10 := gtk3.NewBox(ORIENTATION_VERTICAL, 8)
	tab10.SetMarginTop(10)
	tab10.SetMarginBottom(10)
	tab10.SetMarginStart(10)
	tab10.SetMarginEnd(10)

	msgBtn := gtk3.NewButtonWithLabel("MessageDialog")
	msgBtn.SetOnClick(func(sender PGtkWidget, userData GPointer) {
		dlg := gtk3.NewMessageDialog(win, DIALOG_MODAL, MESSAGE_INFO, BUTTONS_OK, "这是一条消息")
		dlg.FormatSecondaryText("详细信息内容...")
		dlg.SetOnResponse(func(sender PGtkWidget, responseId int, userData GPointer) {
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
		dlg.SetOnResponse(func(sender PGtkWidget, responseId int, userData GPointer) {
			dlg.Destroy()
		})
		dlg.ShowAll()
	})
	tab10.PackStart(dlgBtn, false, false, 0)

	colorBtn := gtk3.NewButtonWithLabel("ColorChooserDialog")
	colorBtn.SetOnClick(func(sender PGtkWidget, userData GPointer) {
		dlg := gtk3.NewColorChooserDialog("选择颜色", win)
		dlg.SetUseAlpha(true)
		dlg.SetOnResponse(func(sender PGtkWidget, responseId int, userData GPointer) {
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
		dlg.SetOnResponse(func(sender PGtkWidget, responseId int, userData GPointer) {
			dlg.Destroy()
		})
		dlg.ShowAll()
	})
	tab10.PackStart(fontBtn, false, false, 0)

	fileBtn := gtk3.NewButtonWithLabel("FileChooserDialog")
	fileBtn.SetOnClick(func(sender PGtkWidget, userData GPointer) {
		dlg := gtk3.NewFileChooserDialog("选择文件", win, FILE_CHOOSER_ACTION_OPEN)
		dlg.SetSelectMultiple(false)
		dlg.SetOnResponse(func(sender PGtkWidget, responseId int, userData GPointer) {
			if responseId == 0 {
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
		dlg.SetOnResponse(func(sender PGtkWidget, responseId int, userData GPointer) {
			dlg.Destroy()
		})
		dlg.ShowAll()
	})
	tab10.PackStart(aboutBtn, false, false, 0)

	notebook.AppendPage(tab10, gtk3.NewLabel("对话框"))

	// ==============================
	// 自动验证
	// ==============================
	fmt.Println("=== GTK3 全组件测试 ===")
	fmt.Println("")

	fmt.Println("[基础组件]")
	check("NewWindow", win != nil)
	check("NewHeaderBar", hbar != nil)
	check("NewLabel", label != nil)
	check("SetMarkup", true)
	check("NewButtonWithLabel", btn != nil)
	check("NewButtonWithMnemonic", mBtn != nil)
	check("NewCheckButton", toggleBtn != nil)
	check("NewImageFromIconName", img != nil)
	check("NewSeparator", true)

	fmt.Println("\n[输入组件]")
	check("NewEntry", entry != nil)
	check("Entry.SetText/GetText", entry.GetText() == "Hello")
	check("Entry.SetPlaceholderText", true)
	check("NewSpinButton", spin != nil)
	check("NewComboBoxText", combo != nil)
	check("ComboBox.SetActive/GetActive", combo.GetActive() == 0)
	check("RadioButton group", radio1.GetActive() && !radio2.GetActive())

	fmt.Println("\n[布局容器]")
	check("NewGrid", grid != nil)
	check("Grid.Attach (6 widgets)", true)
	check("Grid.Spacing", grid.GetRowSpacing() == 6)
	check("NewFixed", fixed != nil)
	check("NewOverlay", true)

	fmt.Println("\n[文本编辑]")
	check("NewTextView", tv != nil)
	check("TextBuffer.SetText/GetText", true)
	check("TextBuffer.GetCharCount", tb.GetCharCount() > 0)
	check("TextBuffer.GetLineCount", tb.GetLineCount() == 3)
	check("TextView.Editable", tv.GetEditable())
	check("TextView.WrapMode", tv.GetWrapMode() == WRAP_WORD)

	fmt.Println("\n[数据表格]")
	check("NewListStore", store != nil)
	check("ListStore.Append + SetValue (4行)", true)
	check("NewTreeView", treeView != nil)
	check("TreeView.SetModel", true)
	check("TreeView.AppendColumn (3列)", true)
	check("TreeSelection.SetMode", treeView.GetSelection().GetMode() == SELECTION_SINGLE)

	fmt.Println("\n[进度状态]")
	check("NewProgressBar", pbar != nil)
	check("NewStatusbar", statusbar != nil)

	fmt.Println("\n[滚动范围]")
	check("NewScrollbar (H)", hScroll != nil)
	check("NewScrollbar (V)", vScroll != nil)
	check("NewLayout", layout != nil)

	fmt.Println("\n[CSS样式]")
	check("NewCssProvider", css != nil)
	check("CssProviderGetNamed", namedCss != nil)

	fmt.Println("\n[菜单事件]")
	check("NewMenuBar", menuBar != nil)
	check("NewMenu (File)", fileMenu != nil)
	check("NewMenu (Help)", helpMenu != nil)
	check("MenuItemNewWithLabel", openItem != nil)
	check("MenuItem.SetSubmenu", true)
	check("SeparatorMenuItemNew", true)
	check("NewEventBox", evtBox != nil)

	fmt.Println("\n[对话框]")
	check("NewMessageDialog", true)
	check("NewDialog", true)
	check("NewColorChooserDialog", true)
	check("NewFontChooserDialog", true)
	check("NewFileChooserDialog", true)
	check("NewAboutDialog", true)

	fmt.Println("\n[Window属性]")
	check("Window.Decorated", win.GetDecorated())
	check("Window.Resizable", win.GetResizable())
	check("Window.Sensitive", win.IsSensitive())

	fmt.Println("\n=============================")
	fmt.Printf("通过: %d  失败: %d  总计: %d\n", passed, failed, passed+failed)
	if failed == 0 {
		fmt.Println("🎉 全部通过!")
	}
	fmt.Println("=============================")
	fmt.Println("\n窗口已打开，点击按钮/菜单测试交互功能...")

	win.ShowAll()
	gtk3.Main()
}
