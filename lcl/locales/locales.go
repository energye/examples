package main

import (
	"fmt"
	_ "github.com/energye/examples/syso/windows"
	"github.com/energye/lcl/api"
	"github.com/energye/lcl/api/libname"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/locales"
	"github.com/energye/lcl/types"
	"os"
	"path/filepath"
)

// 支持的语言列表
var supportedLangs = []struct {
	code string
	name string
}{
	{"zh-CN", "简体中文"},
	{"en-US", "English (US)"},
	{"ja", "日本語"},
	{"ko", "한국어"},
}

// ==================== 主窗体 ====================

type TMainForm struct {
	lcl.TEngForm
	// 语言切换
	langLabel    lcl.ILabel
	langComboBox lcl.IComboBox
	// 顶部面板
	topPanel lcl.IPanel
	// 主菜单
	mainMenu lcl.IMainMenu
	// 工具栏
	toolBar lcl.IToolBar
	// 状态栏
	statusBar lcl.IStatusBar
	// 页面控制
	pageControl lcl.IPageControl
	// === 基本控件页 ===
	basicTab       lcl.ITabSheet
	helloLabel     lcl.ILabel
	welcomeLabel   lcl.ILabel
	nameEdit       lcl.IEdit
	nameEditHint   lcl.ILabel
	passwordEdit   lcl.IEdit
	passwordLabel  lcl.ILabel
	greetBtn       lcl.IButton
	aboutBtn       lcl.IButton
	closeBtn       lcl.IButton
	enableCheckBox lcl.ICheckBox
	// === 选择控件页 ===
	choiceTab      lcl.ITabSheet
	genderGroup    lcl.IGroupBox
	maleRadio      lcl.IRadioButton
	femaleRadio    lcl.IRadioButton
	languageLabel  lcl.ILabel
	languageCombo  lcl.IComboBox
	fruitGroup     lcl.IGroupBox
	fruitCheckList lcl.ICheckListBox
	cityLabel      lcl.ILabel
	cityListBox    lcl.IListBox
	// === 文本控件页 ===
	textTab   lcl.ITabSheet
	memoLabel lcl.ILabel
	memo      lcl.IMemo
	clearBtn  lcl.IButton
	appendBtn lcl.IButton
	spinLabel lcl.ILabel
	spinEdit  lcl.ISpinEdit
	// === 列表控件页 ===
	listTab  lcl.ITabSheet
	listView lcl.IListView
	treeView lcl.ITreeView
	// === 布局控件页 ===
	layoutTab    lcl.ITabSheet
	editGroup    lcl.IGroupBox
	labeledEdit1 lcl.ILabeledEdit
	labeledEdit2 lcl.ILabeledEdit
	panelGroup   lcl.IGroupBox
	infoPanel    lcl.IPanel
	toolPanel    lcl.IPanel
	headerCtl    lcl.IHeaderControl
	btnGroup     lcl.IGroupBox
	okBtn        lcl.IButton
	cancelBtn    lcl.IButton
	radioGroup   lcl.IRadioGroup
	checkGroup   lcl.ICheckGroup
	// === 扩展控件页 ===
	extraTab      lcl.ITabSheet
	staticText    lcl.IStaticText
	maskEdit      lcl.IMaskEdit
	floatSpinEdit lcl.IFloatSpinEdit
	trackBar      lcl.ITrackBar
	progressBar   lcl.IProgressBar
	upDown        lcl.IUpDown
	toggleBox     lcl.IToggleBox
	colorBox      lcl.IColorBox
	colorButton   lcl.IColorButton
	// === 动作页 ===
	actionTab  lcl.ITabSheet
	actionList lcl.IActionList
	actionNew  lcl.IAction
	actionOpen lcl.IAction
	actionSave lcl.IAction
	actionCopy lcl.IAction
	actionMemo lcl.IMemo
	// === 多窗口测试 ===
	openSubFormBtn lcl.IButton
	// 弹出菜单
	popupMenu   lcl.IPopupMenu
	currentLang string
	localeDir   string
}

// ==================== 子窗体 ====================

type TSubForm struct {
	lcl.TEngForm
	// 标题
	titleLabel lcl.ILabel
	// 输入控件
	dateEdit      lcl.IDateEdit
	timeEdit      lcl.ITimeEdit
	fileNameEdit  lcl.IFileNameEdit
	directoryEdit lcl.IDirectoryEdit
	// 按钮
	okBtn     lcl.IButton
	cancelBtn lcl.IButton
	// 状态
	statusLabel lcl.ILabel
}

var mainForm TMainForm
var subForm TSubForm

func main() {
	libname.LibName = "C:\\app\\workspace\\gen\\gout\\libenergy-amd64.dll"
	lcl.Init()
	lcl.Application.Initialize()
	lcl.Application.SetMainFormOnTaskBar(true)
	lcl.Application.NewForm(&mainForm)
	lcl.Application.NewForm(&subForm)
	lcl.Application.Run()
}

// ==================== 主窗体方法 ====================

func (m *TMainForm) FormCreate(sender lcl.IObject) {
	m.currentLang = "zh-CN"
	m.SetName("MainForm")
	m.localeDir = "C:\\app\\workspace\\examples\\lcl\\locales"

	m.SetCaption("国际化示例")
	m.SetPosition(types.PoScreenCenter)
	m.SetWidth(800)
	m.SetHeight(650)

	m.createTopPanel()
	m.createMainMenu()
	m.createToolBar()
	m.createPageControl()
	m.createStatusBar()
	m.createPopupMenu()
}

// ---------- 顶部语言切换面板 ----------

func (m *TMainForm) createTopPanel() {
	m.topPanel = lcl.NewPanel(m)
	m.topPanel.SetParent(m)
	m.topPanel.SetAlign(types.AlTop)
	m.topPanel.SetHeight(40)

	m.langLabel = lcl.NewLabel(m.topPanel)
	m.langLabel.SetParent(m.topPanel)
	m.langLabel.SetName("LangLabel")
	m.langLabel.SetLeft(12)
	m.langLabel.SetTop(10)
	m.langLabel.SetCaption("选择语言：")

	m.langComboBox = lcl.NewComboBox(m.topPanel)
	m.langComboBox.SetParent(m.topPanel)
	m.langComboBox.SetName("LangComboBox")
	m.langComboBox.SetLeft(90)
	m.langComboBox.SetTop(7)
	m.langComboBox.SetWidth(150)
	m.langComboBox.SetStyle(types.CsDropDownList)
	for _, lang := range supportedLangs {
		m.langComboBox.Items().Add(lang.name)
	}
	m.langComboBox.SetItemIndex(0)
	m.langComboBox.SetOnChange(m.onLangChange)

	// 打开子窗口按钮
	m.openSubFormBtn = lcl.NewButton(m.topPanel)
	m.openSubFormBtn.SetParent(m.topPanel)
	m.openSubFormBtn.SetName("OpenSubFormBtn")
	m.openSubFormBtn.SetLeft(260)
	m.openSubFormBtn.SetTop(5)
	m.openSubFormBtn.SetWidth(120)
	m.openSubFormBtn.SetHeight(28)
	m.openSubFormBtn.SetCaption("打开子窗口")
	m.openSubFormBtn.SetHint("点击打开子窗口")
	m.openSubFormBtn.SetOnClick(m.onOpenSubForm)
}

// ---------- 主菜单 ----------

func (m *TMainForm) createMainMenu() {
	m.mainMenu = lcl.NewMainMenu(m)

	// 文件菜单
	fileItem := lcl.NewMenuItem(m)
	fileItem.SetName("MenuFile")
	fileItem.SetCaption("文件(&F)")

	subNew := lcl.NewMenuItem(m)
	subNew.SetName("MenuFileNew")
	subNew.SetCaption("新建(&N)")
	subNew.SetShortCut(api.TextToShortCut("Ctrl+N"))
	subNew.SetHint("创建新文件")
	fileItem.Add(subNew)

	subOpen := lcl.NewMenuItem(m)
	subOpen.SetName("MenuFileOpen")
	subOpen.SetCaption("打开(&O)")
	subOpen.SetShortCut(api.TextToShortCut("Ctrl+O"))
	subOpen.SetHint("打开已有文件")
	fileItem.Add(subOpen)

	subSave := lcl.NewMenuItem(m)
	subSave.SetName("MenuFileSave")
	subSave.SetCaption("保存(&S)")
	subSave.SetShortCut(api.TextToShortCut("Ctrl+S"))
	subSave.SetHint("保存当前文件")
	fileItem.Add(subSave)

	sep := lcl.NewMenuItem(m)
	sep.SetCaption("-")
	fileItem.Add(sep)

	subExit := lcl.NewMenuItem(m)
	subExit.SetName("MenuFileExit")
	subExit.SetCaption("退出(&Q)")
	subExit.SetShortCut(api.TextToShortCut("Ctrl+Q"))
	subExit.SetOnClick(func(lcl.IObject) { m.Close() })
	fileItem.Add(subExit)

	m.mainMenu.Items().Add(fileItem)

	// 编辑菜单
	editItem := lcl.NewMenuItem(m)
	editItem.SetName("MenuEdit")
	editItem.SetCaption("编辑(&E)")

	subCopy := lcl.NewMenuItem(m)
	subCopy.SetName("MenuEditCopy")
	subCopy.SetCaption("复制(&C)")
	subCopy.SetShortCut(api.TextToShortCut("Ctrl+C"))
	editItem.Add(subCopy)

	subPaste := lcl.NewMenuItem(m)
	subPaste.SetName("MenuEditPaste")
	subPaste.SetCaption("粘贴(&P)")
	subPaste.SetShortCut(api.TextToShortCut("Ctrl+V"))
	editItem.Add(subPaste)

	subFind := lcl.NewMenuItem(m)
	subFind.SetName("MenuEditFind")
	subFind.SetCaption("查找(&F)")
	subFind.SetShortCut(api.TextToShortCut("Ctrl+F"))
	editItem.Add(subFind)

	subReplace := lcl.NewMenuItem(m)
	subReplace.SetName("MenuEditReplace")
	subReplace.SetCaption("替换(&R)")
	subReplace.SetShortCut(api.TextToShortCut("Ctrl+H"))
	editItem.Add(subReplace)

	m.mainMenu.Items().Add(editItem)

	// 帮助菜单
	helpItem := lcl.NewMenuItem(m)
	helpItem.SetName("MenuHelp")
	helpItem.SetCaption("帮助(&H)")

	subDoc := lcl.NewMenuItem(m)
	subDoc.SetName("MenuHelpDoc")
	subDoc.SetCaption("文档(&D)")
	subDoc.SetShortCut(api.TextToShortCut("F1"))
	helpItem.Add(subDoc)

	sep2 := lcl.NewMenuItem(m)
	sep2.SetCaption("-")
	helpItem.Add(sep2)

	subAbout := lcl.NewMenuItem(m)
	subAbout.SetName("MenuHelpAbout")
	subAbout.SetCaption("关于(&A)")
	subAbout.SetOnClick(m.onAboutClick)
	helpItem.Add(subAbout)

	m.mainMenu.Items().Add(helpItem)
}

// ---------- 工具栏 ----------

func (m *TMainForm) createToolBar() {
	m.toolBar = lcl.NewToolBar(m)
	m.toolBar.SetParent(m)
	m.toolBar.SetName("ToolBar")
	m.toolBar.SetShowCaptions(true)

	tbNew := lcl.NewToolButton(m.toolBar)
	tbNew.SetParent(m.toolBar)
	tbNew.SetName("ToolBtnNew")
	tbNew.SetCaption("新建")
	tbNew.SetHint("创建新文件")

	tbOpen := lcl.NewToolButton(m.toolBar)
	tbOpen.SetParent(m.toolBar)
	tbOpen.SetName("ToolBtnOpen")
	tbOpen.SetCaption("打开")
	tbOpen.SetHint("打开已有文件")

	tbSave := lcl.NewToolButton(m.toolBar)
	tbSave.SetParent(m.toolBar)
	tbSave.SetName("ToolBtnSave")
	tbSave.SetCaption("保存")
	tbSave.SetHint("保存当前文件")

	tbSep := lcl.NewToolButton(m.toolBar)
	tbSep.SetParent(m.toolBar)
	tbSep.SetStyle(types.TbsSeparator)

	tbCut := lcl.NewToolButton(m.toolBar)
	tbCut.SetParent(m.toolBar)
	tbCut.SetName("ToolBtnCut")
	tbCut.SetCaption("剪切")
	tbCut.SetHint("剪切选中内容")

	tbCopy := lcl.NewToolButton(m.toolBar)
	tbCopy.SetParent(m.toolBar)
	tbCopy.SetName("ToolBtnCopy")
	tbCopy.SetCaption("复制")
	tbCopy.SetHint("复制选中内容")

	tbPaste := lcl.NewToolButton(m.toolBar)
	tbPaste.SetParent(m.toolBar)
	tbPaste.SetName("ToolBtnPaste")
	tbPaste.SetCaption("粘贴")
	tbPaste.SetHint("粘贴剪贴板内容")
}

// ---------- 页面控制 ----------

func (m *TMainForm) createPageControl() {
	m.pageControl = lcl.NewPageControl(m)
	m.pageControl.SetParent(m)
	m.pageControl.SetName("PageControl")
	m.pageControl.SetAlign(types.AlClient)

	m.createBasicTab()
	m.createChoiceTab()
	m.createTextTab()
	m.createListTab()
	m.createLayoutTab()
	m.createExtraTab()
	m.createActionTab()
}

// ---------- 基本控件页 ----------

func (m *TMainForm) createBasicTab() {
	m.basicTab = lcl.NewTabSheet(m)
	m.basicTab.SetName("BasicTab")
	m.basicTab.SetPageControl(m.pageControl)
	m.basicTab.SetCaption("基本控件")

	m.helloLabel = lcl.NewLabel(m.basicTab)
	m.helloLabel.SetParent(m.basicTab)
	m.helloLabel.SetName("HelloLabel")
	m.helloLabel.SetLeft(20)
	m.helloLabel.SetTop(15)
	m.helloLabel.SetCaption("你好，世界！")
	m.helloLabel.Font().SetSize(16)
	m.helloLabel.Font().SetBold(true)

	m.welcomeLabel = lcl.NewLabel(m.basicTab)
	m.welcomeLabel.SetParent(m.basicTab)
	m.welcomeLabel.SetName("WelcomeLabel")
	m.welcomeLabel.SetLeft(20)
	m.welcomeLabel.SetTop(50)
	m.welcomeLabel.SetCaption("欢迎使用 LCL 国际化示例程序。")

	m.nameEditHint = lcl.NewLabel(m.basicTab)
	m.nameEditHint.SetParent(m.basicTab)
	m.nameEditHint.SetName("NameEditHint")
	m.nameEditHint.SetLeft(20)
	m.nameEditHint.SetTop(85)
	m.nameEditHint.SetCaption("请输入姓名：")

	m.nameEdit = lcl.NewEdit(m.basicTab)
	m.nameEdit.SetParent(m.basicTab)
	m.nameEdit.SetName("NameEdit")
	m.nameEdit.SetLeft(110)
	m.nameEdit.SetTop(82)
	m.nameEdit.SetWidth(200)
	m.nameEdit.SetTextHint("在此输入姓名")
	m.nameEdit.SetHint("输入您的姓名")
	m.nameEdit.SetText("")

	m.passwordLabel = lcl.NewLabel(m.basicTab)
	m.passwordLabel.SetParent(m.basicTab)
	m.passwordLabel.SetName("PasswordLabel")
	m.passwordLabel.SetLeft(20)
	m.passwordLabel.SetTop(115)
	m.passwordLabel.SetCaption("密码：")

	m.passwordEdit = lcl.NewEdit(m.basicTab)
	m.passwordEdit.SetParent(m.basicTab)
	m.passwordEdit.SetName("PasswordEdit")
	m.passwordEdit.SetLeft(110)
	m.passwordEdit.SetTop(112)
	m.passwordEdit.SetWidth(200)
	m.passwordEdit.SetPasswordChar('*')
	m.passwordEdit.SetTextHint("请输入密码")
	m.passwordEdit.SetText("")

	m.enableCheckBox = lcl.NewCheckBox(m.basicTab)
	m.enableCheckBox.SetParent(m.basicTab)
	m.enableCheckBox.SetName("EnableCheckBox")
	m.enableCheckBox.SetLeft(20)
	m.enableCheckBox.SetTop(150)
	m.enableCheckBox.SetCaption("启用按钮")
	m.enableCheckBox.SetChecked(true)
	m.enableCheckBox.SetOnClick(func(sender lcl.IObject) {
		enabled := m.enableCheckBox.Checked()
		m.greetBtn.SetEnabled(enabled)
		m.aboutBtn.SetEnabled(enabled)
	})

	m.greetBtn = lcl.NewButton(m.basicTab)
	m.greetBtn.SetParent(m.basicTab)
	m.greetBtn.SetName("GreetBtn")
	m.greetBtn.SetLeft(20)
	m.greetBtn.SetTop(185)
	m.greetBtn.SetWidth(100)
	m.greetBtn.SetCaption("问候")
	m.greetBtn.SetOnClick(m.onGreetClick)

	m.aboutBtn = lcl.NewButton(m.basicTab)
	m.aboutBtn.SetParent(m.basicTab)
	m.aboutBtn.SetName("AboutBtn")
	m.aboutBtn.SetLeft(130)
	m.aboutBtn.SetTop(185)
	m.aboutBtn.SetWidth(100)
	m.aboutBtn.SetCaption("关于")
	m.aboutBtn.SetOnClick(m.onAboutClick)

	m.closeBtn = lcl.NewButton(m.basicTab)
	m.closeBtn.SetParent(m.basicTab)
	m.closeBtn.SetName("CloseBtn")
	m.closeBtn.SetLeft(240)
	m.closeBtn.SetTop(185)
	m.closeBtn.SetWidth(100)
	m.closeBtn.SetCaption("关闭")
	m.closeBtn.SetOnClick(func(sender lcl.IObject) { m.Close() })
}

// ---------- 选择控件页 ----------

func (m *TMainForm) createChoiceTab() {
	m.choiceTab = lcl.NewTabSheet(m)
	m.choiceTab.SetName("ChoiceTab")
	m.choiceTab.SetPageControl(m.pageControl)
	m.choiceTab.SetCaption("选择控件")

	// GroupBox + RadioButton
	m.genderGroup = lcl.NewGroupBox(m.choiceTab)
	m.genderGroup.SetParent(m.choiceTab)
	m.genderGroup.SetName("GenderGroup")
	m.genderGroup.SetBounds(20, 10, 200, 60)
	m.genderGroup.SetCaption("性别")

	m.maleRadio = lcl.NewRadioButton(m.genderGroup)
	m.maleRadio.SetParent(m.genderGroup)
	m.maleRadio.SetName("MaleRadio")
	m.maleRadio.SetLeft(15)
	m.maleRadio.SetTop(5)
	m.maleRadio.SetCaption("男")
	m.maleRadio.SetChecked(true)

	m.femaleRadio = lcl.NewRadioButton(m.genderGroup)
	m.femaleRadio.SetParent(m.genderGroup)
	m.femaleRadio.SetName("FemaleRadio")
	m.femaleRadio.SetLeft(80)
	m.femaleRadio.SetTop(5)
	m.femaleRadio.SetCaption("女")

	// 语言下拉框
	m.languageLabel = lcl.NewLabel(m.choiceTab)
	m.languageLabel.SetParent(m.choiceTab)
	m.languageLabel.SetName("LanguageLabel")
	m.languageLabel.SetLeft(20)
	m.languageLabel.SetTop(80)
	m.languageLabel.SetCaption("语言偏好：")

	m.languageCombo = lcl.NewComboBox(m.choiceTab)
	m.languageCombo.SetParent(m.choiceTab)
	m.languageCombo.SetName("LanguageCombo")
	m.languageCombo.SetLeft(80)
	m.languageCombo.SetTop(77)
	m.languageCombo.SetWidth(140)
	m.languageCombo.SetStyle(types.CsDropDownList)
	m.languageCombo.SetHint("选择您偏好的语言")
	m.languageCombo.Items().Add("中文")
	m.languageCombo.Items().Add("英文")
	m.languageCombo.Items().Add("日文")
	m.languageCombo.Items().Add("韩文")
	m.languageCombo.SetItemIndex(0)

	// GroupBox + CheckListBox
	m.fruitGroup = lcl.NewGroupBox(m.choiceTab)
	m.fruitGroup.SetParent(m.choiceTab)
	m.fruitGroup.SetName("FruitGroup")
	m.fruitGroup.SetBounds(20, 110, 200, 110)
	m.fruitGroup.SetCaption("喜欢的水果")

	m.fruitCheckList = lcl.NewCheckListBox(m.fruitGroup)
	m.fruitCheckList.SetParent(m.fruitGroup)
	m.fruitCheckList.SetName("FruitCheckList")
	m.fruitCheckList.SetBounds(10, 20, 180, 80)
	m.fruitCheckList.SetHint("勾选您喜欢的水果")
	m.fruitCheckList.Items().Add("苹果")
	m.fruitCheckList.Items().Add("香蕉")
	m.fruitCheckList.Items().Add("橙子")
	m.fruitCheckList.Items().Add("葡萄")
	m.fruitCheckList.SetChecked(0, true)

	// 城市列表
	m.cityLabel = lcl.NewLabel(m.choiceTab)
	m.cityLabel.SetParent(m.choiceTab)
	m.cityLabel.SetName("CityLabel")
	m.cityLabel.SetLeft(240)
	m.cityLabel.SetTop(10)
	m.cityLabel.SetCaption("选择城市：")

	m.cityListBox = lcl.NewListBox(m.choiceTab)
	m.cityListBox.SetParent(m.choiceTab)
	m.cityListBox.SetName("CityListBox")
	m.cityListBox.SetBounds(240, 30, 180, 110)
	m.cityListBox.SetHint("选择一个城市")
	m.cityListBox.Items().Add("北京")
	m.cityListBox.Items().Add("上海")
	m.cityListBox.Items().Add("广州")
	m.cityListBox.Items().Add("深圳")
	m.cityListBox.SetItemIndex(0)

	// RadioGroup
	m.radioGroup = lcl.NewRadioGroup(m.choiceTab)
	m.radioGroup.SetParent(m.choiceTab)
	m.radioGroup.SetName("RadioGroup")
	m.radioGroup.SetBounds(240, 150, 200, 80)
	m.radioGroup.SetCaption("字体大小")
	m.radioGroup.SetHint("选择字体大小")
	m.radioGroup.Items().Add("小")
	m.radioGroup.Items().Add("中")
	m.radioGroup.Items().Add("大")
	m.radioGroup.SetItemIndex(1)

	// CheckGroup
	m.checkGroup = lcl.NewCheckGroup(m.choiceTab)
	m.checkGroup.SetParent(m.choiceTab)
	m.checkGroup.SetName("CheckGroup")
	m.checkGroup.SetBounds(460, 10, 200, 100)
	m.checkGroup.SetCaption("功能选项")
	m.checkGroup.SetHint("勾选需要的功能")
	m.checkGroup.Items().Add("自动保存")
	m.checkGroup.Items().Add("显示行号")
	m.checkGroup.Items().Add("语法高亮")
	m.checkGroup.SetChecked(0, true)
}

// ---------- 文本控件页 ----------

func (m *TMainForm) createTextTab() {
	m.textTab = lcl.NewTabSheet(m)
	m.textTab.SetName("TextTab")
	m.textTab.SetPageControl(m.pageControl)
	m.textTab.SetCaption("文本控件")

	m.memoLabel = lcl.NewLabel(m.textTab)
	m.memoLabel.SetParent(m.textTab)
	m.memoLabel.SetName("MemoLabel")
	m.memoLabel.SetLeft(20)
	m.memoLabel.SetTop(10)
	m.memoLabel.SetCaption("备注信息：")

	m.memo = lcl.NewMemo(m.textTab)
	m.memo.SetParent(m.textTab)
	m.memo.SetName("Memo")
	m.memo.SetBounds(20, 30, 350, 120)
	m.memo.SetHint("在此输入备注内容")
	m.memo.Lines().Add("这是一段示例文本。")
	m.memo.Lines().Add("可以在这里输入多行内容。")
	m.memo.SetScrollBars(types.SsVertical)

	m.clearBtn = lcl.NewButton(m.textTab)
	m.clearBtn.SetParent(m.textTab)
	m.clearBtn.SetName("ClearBtn")
	m.clearBtn.SetLeft(20)
	m.clearBtn.SetTop(160)
	m.clearBtn.SetWidth(80)
	m.clearBtn.SetCaption("清空")
	m.clearBtn.SetOnClick(func(sender lcl.IObject) {
		m.memo.Lines().Clear()
	})

	m.appendBtn = lcl.NewButton(m.textTab)
	m.appendBtn.SetParent(m.textTab)
	m.appendBtn.SetName("AppendBtn")
	m.appendBtn.SetLeft(110)
	m.appendBtn.SetTop(160)
	m.appendBtn.SetWidth(80)
	m.appendBtn.SetCaption("追加")
	m.appendBtn.SetOnClick(func(sender lcl.IObject) {
		m.memo.Lines().Add("追加的新行内容。")
	})

	m.spinLabel = lcl.NewLabel(m.textTab)
	m.spinLabel.SetParent(m.textTab)
	m.spinLabel.SetName("SpinLabel")
	m.spinLabel.SetLeft(20)
	m.spinLabel.SetTop(200)
	m.spinLabel.SetCaption("数量：")

	m.spinEdit = lcl.NewSpinEdit(m.textTab)
	m.spinEdit.SetParent(m.textTab)
	m.spinEdit.SetName("SpinEdit")
	m.spinEdit.SetLeft(80)
	m.spinEdit.SetTop(197)
	m.spinEdit.SetWidth(100)
	m.spinEdit.SetMinValue(0)
	m.spinEdit.SetMaxValue(100)
	m.spinEdit.SetValue(10)
	m.spinEdit.SetHint("输入数量")
}

// ---------- 列表控件页 ----------

func (m *TMainForm) createListTab() {
	m.listTab = lcl.NewTabSheet(m)
	m.listTab.SetName("ListTab")
	m.listTab.SetPageControl(m.pageControl)
	m.listTab.SetCaption("列表控件")

	tvLabel := lcl.NewLabel(m.listTab)
	tvLabel.SetParent(m.listTab)
	tvLabel.SetName("TreeLabel")
	tvLabel.SetLeft(10)
	tvLabel.SetTop(5)
	tvLabel.SetCaption("目录树：")

	m.treeView = lcl.NewTreeView(m.listTab)
	m.treeView.SetParent(m.listTab)
	m.treeView.SetName("TreeView")
	m.treeView.SetBounds(10, 25, 220, 200)
	m.treeView.SetAutoExpand(true)
	m.treeView.SetHint("浏览目录结构")

	m.treeView.Items().BeginUpdate()
	root1 := m.treeView.Items().AddChild(nil, "文档")
	m.treeView.Items().AddChild(root1, "工作文档")
	m.treeView.Items().AddChild(root1, "个人文档")
	root2 := m.treeView.Items().AddChild(nil, "图片")
	m.treeView.Items().AddChild(root2, "截图")
	m.treeView.Items().AddChild(root2, "照片")
	root3 := m.treeView.Items().AddChild(nil, "音乐")
	m.treeView.Items().AddChild(root3, "流行")
	m.treeView.Items().AddChild(root3, "古典")
	m.treeView.Items().EndUpdate()

	lvLabel := lcl.NewLabel(m.listTab)
	lvLabel.SetParent(m.listTab)
	lvLabel.SetName("ListLabel")
	lvLabel.SetLeft(250)
	lvLabel.SetTop(5)
	lvLabel.SetCaption("文件列表：")

	m.listView = lcl.NewListView(m.listTab)
	m.listView.SetParent(m.listTab)
	m.listView.SetName("ListView")
	m.listView.SetBounds(250, 25, 300, 200)
	m.listView.SetViewStyle(types.VsReport)
	m.listView.SetRowSelect(true)
	m.listView.SetReadOnly(true)
	m.listView.SetGridLines(true)
	m.listView.SetHint("文件列表")

	col1 := m.listView.Columns().AddToListColumn()
	col1.SetCaption("名称")
	col1.SetWidth(120)

	col2 := m.listView.Columns().AddToListColumn()
	col2.SetCaption("大小")
	col2.SetWidth(80)

	col3 := m.listView.Columns().AddToListColumn()
	col3.SetCaption("类型")
	col3.SetWidth(80)

	m.listView.Items().BeginUpdate()
	for i := 1; i <= 5; i++ {
		item := m.listView.Items().Add()
		item.SetCaption(fmt.Sprintf("文件%d.txt", i))
		item.SubItems().Add(fmt.Sprintf("%d KB", i*10))
		item.SubItems().Add("文本文件")
	}
	m.listView.Items().EndUpdate()
}

// ---------- 布局控件页 ----------

func (m *TMainForm) createLayoutTab() {
	m.layoutTab = lcl.NewTabSheet(m)
	m.layoutTab.SetName("LayoutTab")
	m.layoutTab.SetPageControl(m.pageControl)
	m.layoutTab.SetCaption("布局控件")

	// GroupBox + LabeledEdit
	m.editGroup = lcl.NewGroupBox(m.layoutTab)
	m.editGroup.SetParent(m.layoutTab)
	m.editGroup.SetName("EditGroup")
	m.editGroup.SetBounds(20, 10, 340, 100)
	m.editGroup.SetCaption("用户信息")

	m.labeledEdit1 = lcl.NewLabeledEdit(m.editGroup)
	m.labeledEdit1.SetParent(m.editGroup)
	m.labeledEdit1.SetName("LabeledEditUser")
	m.labeledEdit1.SetLabelPosition(types.LpLeft)
	m.labeledEdit1.SetBounds(100, 5, 210, 25)
	m.labeledEdit1.EditLabel().SetCaption("用户名：")
	m.labeledEdit1.SetTextHint("请输入用户名")
	m.labeledEdit1.SetHint("输入您的用户名")
	m.labeledEdit1.SetText("")

	m.labeledEdit2 = lcl.NewLabeledEdit(m.editGroup)
	m.labeledEdit2.SetParent(m.editGroup)
	m.labeledEdit2.SetName("LabeledEditEmail")
	m.labeledEdit2.SetLabelPosition(types.LpLeft)
	m.labeledEdit2.SetBounds(100, 35, 210, 25)
	m.labeledEdit2.EditLabel().SetCaption("邮箱：")
	m.labeledEdit2.SetTextHint("请输入邮箱地址")
	m.labeledEdit2.SetHint("输入您的邮箱地址")
	m.labeledEdit2.SetText("")

	// GroupBox + Panel
	m.panelGroup = lcl.NewGroupBox(m.layoutTab)
	m.panelGroup.SetParent(m.layoutTab)
	m.panelGroup.SetName("PanelGroup")
	m.panelGroup.SetBounds(380, 10, 340, 100)
	m.panelGroup.SetCaption("面板容器")

	m.infoPanel = lcl.NewPanel(m.panelGroup)
	m.infoPanel.SetParent(m.panelGroup)
	m.infoPanel.SetName("InfoPanel")
	m.infoPanel.SetBounds(10, 5, 150, 60)
	m.infoPanel.SetCaption("信息面板")

	m.toolPanel = lcl.NewPanel(m.panelGroup)
	m.toolPanel.SetParent(m.panelGroup)
	m.toolPanel.SetName("ToolPanel")
	m.toolPanel.SetBounds(170, 5, 150, 60)
	m.toolPanel.SetCaption("工具面板")

	// HeaderControl — Section 无 Name 属性，不参与 i18n
	m.headerCtl = lcl.NewHeaderControl(m.layoutTab)
	m.headerCtl.SetParent(m.layoutTab)
	m.headerCtl.SetName("HeaderControl")
	m.headerCtl.SetBounds(20, 120, 700, 25)

	hs1 := m.headerCtl.Sections().AddToHeaderSection()
	hs1.SetText("名称")
	hs1.SetWidth(200)

	hs2 := m.headerCtl.Sections().AddToHeaderSection()
	hs2.SetText("描述")
	hs2.SetWidth(300)

	hs3 := m.headerCtl.Sections().AddToHeaderSection()
	hs3.SetText("日期")
	hs3.SetWidth(200)

	// GroupBox + 按钮
	m.btnGroup = lcl.NewGroupBox(m.layoutTab)
	m.btnGroup.SetParent(m.layoutTab)
	m.btnGroup.SetName("BtnGroup")
	m.btnGroup.SetBounds(20, 155, 340, 70)
	m.btnGroup.SetCaption("操作按钮")

	m.okBtn = lcl.NewButton(m.btnGroup)
	m.okBtn.SetParent(m.btnGroup)
	m.okBtn.SetName("BitBtnOK")
	m.okBtn.SetBounds(10, 5, 90, 30)
	m.okBtn.SetCaption("确定")
	m.okBtn.SetHint("确认操作")

	m.cancelBtn = lcl.NewButton(m.btnGroup)
	m.cancelBtn.SetParent(m.btnGroup)
	m.cancelBtn.SetName("BitBtnCancel")
	m.cancelBtn.SetBounds(110, 5, 90, 30)
	m.cancelBtn.SetCaption("取消")
	m.cancelBtn.SetHint("取消操作")
}

// ---------- 扩展控件页 ----------

func (m *TMainForm) createExtraTab() {
	m.extraTab = lcl.NewTabSheet(m)
	m.extraTab.SetName("ExtraTab")
	m.extraTab.SetPageControl(m.pageControl)
	m.extraTab.SetCaption("扩展控件")

	// StaticText
	m.staticText = lcl.NewStaticText(m.extraTab)
	m.staticText.SetParent(m.extraTab)
	m.staticText.SetName("StaticText")
	m.staticText.SetLeft(20)
	m.staticText.SetTop(15)
	m.staticText.SetCaption("这是静态文本（StaticText）")
	m.staticText.SetHint("StaticText 提示")

	// MaskEdit
	maskLabel := lcl.NewLabel(m.extraTab)
	maskLabel.SetParent(m.extraTab)
	maskLabel.SetName("MaskEditLabel")
	maskLabel.SetLeft(20)
	maskLabel.SetTop(45)
	maskLabel.SetCaption("电话号码：")

	m.maskEdit = lcl.NewMaskEdit(m.extraTab)
	m.maskEdit.SetParent(m.extraTab)
	m.maskEdit.SetName("MaskEdit")
	m.maskEdit.SetLeft(100)
	m.maskEdit.SetTop(42)
	m.maskEdit.SetWidth(200)
	m.maskEdit.SetTextHint("请输入电话号码")
	m.maskEdit.SetHint("输入您的电话号码")
	m.maskEdit.SetText("")

	// FloatSpinEdit
	floatLabel := lcl.NewLabel(m.extraTab)
	floatLabel.SetParent(m.extraTab)
	floatLabel.SetName("FloatSpinLabel")
	floatLabel.SetLeft(20)
	floatLabel.SetTop(80)
	floatLabel.SetCaption("浮点数：")

	m.floatSpinEdit = lcl.NewFloatSpinEdit(m.extraTab)
	m.floatSpinEdit.SetParent(m.extraTab)
	m.floatSpinEdit.SetName("FloatSpinEdit")
	m.floatSpinEdit.SetLeft(100)
	m.floatSpinEdit.SetTop(77)
	m.floatSpinEdit.SetWidth(120)
	m.floatSpinEdit.SetMinValue(0.0)
	m.floatSpinEdit.SetMaxValue(100.0)
	m.floatSpinEdit.SetValue(3.14)
	m.floatSpinEdit.SetHint("输入浮点数")

	// TrackBar
	trackLabel := lcl.NewLabel(m.extraTab)
	trackLabel.SetParent(m.extraTab)
	trackLabel.SetName("TrackBarLabel")
	trackLabel.SetLeft(20)
	trackLabel.SetTop(115)
	trackLabel.SetCaption("滑块：")

	m.trackBar = lcl.NewTrackBar(m.extraTab)
	m.trackBar.SetParent(m.extraTab)
	m.trackBar.SetName("TrackBar")
	m.trackBar.SetBounds(80, 110, 200, 30)
	m.trackBar.SetMin(0)
	m.trackBar.SetMax(100)
	m.trackBar.SetPosition(50)
	m.trackBar.SetHint("拖动滑块")

	// ProgressBar
	progressLabel := lcl.NewLabel(m.extraTab)
	progressLabel.SetParent(m.extraTab)
	progressLabel.SetName("ProgressLabel")
	progressLabel.SetLeft(20)
	progressLabel.SetTop(155)
	progressLabel.SetCaption("进度条：")

	m.progressBar = lcl.NewProgressBar(m.extraTab)
	m.progressBar.SetParent(m.extraTab)
	m.progressBar.SetName("ProgressBar")
	m.progressBar.SetBounds(80, 150, 200, 25)
	m.progressBar.SetPosition(60)
	m.progressBar.SetHint("进度显示")

	// UpDown
	upDownLabel := lcl.NewLabel(m.extraTab)
	upDownLabel.SetParent(m.extraTab)
	upDownLabel.SetName("UpDownLabel")
	upDownLabel.SetLeft(20)
	upDownLabel.SetTop(190)
	upDownLabel.SetCaption("微调框：")

	m.upDown = lcl.NewUpDown(m.extraTab)
	m.upDown.SetParent(m.extraTab)
	m.upDown.SetName("UpDown")
	m.upDown.SetBounds(100, 185, 50, 25)
	m.upDown.SetMin(0)
	m.upDown.SetMax(100)
	m.upDown.SetPosition(50)
	m.upDown.SetHint("点击调整数值")

	// ToggleBox
	m.toggleBox = lcl.NewToggleBox(m.extraTab)
	m.toggleBox.SetParent(m.extraTab)
	m.toggleBox.SetName("ToggleBox")
	m.toggleBox.SetBounds(20, 225, 120, 25)
	m.toggleBox.SetCaption("切换开关")
	m.toggleBox.SetHint("点击切换状态")
	m.toggleBox.SetChecked(false)

	// ColorBox
	colorBoxLabel := lcl.NewLabel(m.extraTab)
	colorBoxLabel.SetParent(m.extraTab)
	colorBoxLabel.SetName("ColorBoxLabel")
	colorBoxLabel.SetLeft(20)
	colorBoxLabel.SetTop(265)
	colorBoxLabel.SetCaption("颜色选择：")

	m.colorBox = lcl.NewColorBox(m.extraTab)
	m.colorBox.SetParent(m.extraTab)
	m.colorBox.SetName("ColorBox")
	m.colorBox.SetBounds(100, 260, 150, 25)
	m.colorBox.SetHint("选择颜色")

	// ColorButton
	m.colorButton = lcl.NewColorButton(m.extraTab)
	m.colorButton.SetParent(m.extraTab)
	m.colorButton.SetName("ColorButton")
	m.colorButton.SetBounds(270, 260, 80, 25)
	m.colorButton.SetCaption("颜色")
	m.colorButton.SetHint("点击选择颜色")
}

// ---------- 动作页 ----------

func (m *TMainForm) createActionTab() {
	m.actionTab = lcl.NewTabSheet(m)
	m.actionTab.SetName("ActionTab")
	m.actionTab.SetPageControl(m.pageControl)
	m.actionTab.SetCaption("动作系统")

	lbl := lcl.NewLabel(m.actionTab)
	lbl.SetParent(m.actionTab)
	lbl.SetName("ActionHintLabel")
	lbl.SetLeft(20)
	lbl.SetTop(10)
	lbl.SetCaption("以下按钮和菜单共享同一组动作，修改动作 Caption 会同步更新：")

	m.actionList = lcl.NewActionList(m.actionTab)

	m.actionNew = lcl.NewAction(m.actionList)
	m.actionNew.SetName("ActionNew")
	m.actionNew.SetCaption("新建文档")
	m.actionNew.SetHint("创建一个新文档")

	m.actionOpen = lcl.NewAction(m.actionList)
	m.actionOpen.SetName("ActionOpen")
	m.actionOpen.SetCaption("打开文档")
	m.actionOpen.SetHint("打开一个已有文档")

	m.actionSave = lcl.NewAction(m.actionList)
	m.actionSave.SetName("ActionSave")
	m.actionSave.SetCaption("保存文档")
	m.actionSave.SetHint("保存当前文档")

	m.actionCopy = lcl.NewAction(m.actionList)
	m.actionCopy.SetName("ActionCopy")
	m.actionCopy.SetCaption("复制内容")
	m.actionCopy.SetHint("复制选中的内容")

	// 绑定到按钮
	btnNew := lcl.NewButton(m.actionTab)
	btnNew.SetParent(m.actionTab)
	btnNew.SetName("ActionBtnNew")
	btnNew.SetBounds(20, 40, 100, 30)
	btnNew.SetAction(m.actionNew)

	btnOpen := lcl.NewButton(m.actionTab)
	btnOpen.SetParent(m.actionTab)
	btnOpen.SetName("ActionBtnOpen")
	btnOpen.SetBounds(130, 40, 100, 30)
	btnOpen.SetAction(m.actionOpen)

	btnSave := lcl.NewButton(m.actionTab)
	btnSave.SetParent(m.actionTab)
	btnSave.SetName("ActionBtnSave")
	btnSave.SetBounds(240, 40, 100, 30)
	btnSave.SetAction(m.actionSave)

	btnCopy := lcl.NewButton(m.actionTab)
	btnCopy.SetParent(m.actionTab)
	btnCopy.SetName("ActionBtnCopy")
	btnCopy.SetBounds(350, 40, 100, 30)
	btnCopy.SetAction(m.actionCopy)

	// 日志 Memo
	actionMemoLabel := lcl.NewLabel(m.actionTab)
	actionMemoLabel.SetParent(m.actionTab)
	actionMemoLabel.SetName("ActionMemoLabel")
	actionMemoLabel.SetLeft(20)
	actionMemoLabel.SetTop(80)
	actionMemoLabel.SetCaption("动作日志：")

	m.actionMemo = lcl.NewMemo(m.actionTab)
	m.actionMemo.SetParent(m.actionTab)
	m.actionMemo.SetName("ActionMemo")
	m.actionMemo.SetBounds(20, 100, 500, 120)
	m.actionMemo.SetHint("动作执行日志")
	m.actionMemo.SetScrollBars(types.SsVertical)

	m.actionNew.SetOnExecute(func(lcl.IObject) {
		m.actionMemo.Lines().Add("[新建] 动作已执行")
	})
	m.actionOpen.SetOnExecute(func(lcl.IObject) {
		m.actionMemo.Lines().Add("[打开] 动作已执行")
	})
	m.actionSave.SetOnExecute(func(lcl.IObject) {
		m.actionMemo.Lines().Add("[保存] 动作已执行")
	})
	m.actionCopy.SetOnExecute(func(lcl.IObject) {
		m.actionMemo.Lines().Add("[复制] 动作已执行")
	})
}

// ---------- 状态栏 ----------

func (m *TMainForm) createStatusBar() {
	m.statusBar = lcl.NewStatusBar(m)
	m.statusBar.SetParent(m)
	m.statusBar.SetName("StatusBar")
	m.statusBar.SetSimplePanel(false)

	p1 := m.statusBar.Panels().AddToStatusPanel()
	p1.SetText("当前语言：简体中文")
	p1.SetWidth(180)

	p2 := m.statusBar.Panels().AddToStatusPanel()
	p2.SetText("就绪")
	p2.SetWidth(120)

	p3 := m.statusBar.Panels().AddToStatusPanel()
	p3.SetText("国际化示例程序 v2.0")
}

// ---------- 弹出菜单 ----------

func (m *TMainForm) createPopupMenu() {
	m.popupMenu = lcl.NewPopupMenu(m)

	item1 := lcl.NewMenuItem(m)
	item1.SetName("PopupCopy")
	item1.SetCaption("复制")
	m.popupMenu.Items().Add(item1)

	item2 := lcl.NewMenuItem(m)
	item2.SetName("PopupPaste")
	item2.SetCaption("粘贴")
	m.popupMenu.Items().Add(item2)

	sep := lcl.NewMenuItem(m)
	sep.SetCaption("-")
	m.popupMenu.Items().Add(sep)

	item3 := lcl.NewMenuItem(m)
	item3.SetName("PopupSelectAll")
	item3.SetCaption("全选")
	m.popupMenu.Items().Add(item3)

	sep2 := lcl.NewMenuItem(m)
	sep2.SetCaption("-")
	m.popupMenu.Items().Add(sep2)

	item4 := lcl.NewMenuItem(m)
	item4.SetName("PopupDelete")
	item4.SetCaption("删除")
	m.popupMenu.Items().Add(item4)

	m.memo.SetPopupMenu(m.popupMenu)
}

// ---------- 事件处理 ----------

func (m *TMainForm) onLangChange(sender lcl.IObject) {
	idx := m.langComboBox.ItemIndex()
	if idx < 0 || int(idx) >= len(supportedLangs) {
		return
	}
	lang := supportedLangs[idx].code
	if lang == m.currentLang {
		return
	}
	m.currentLang = lang

	data, err := os.ReadFile(filepath.Join(m.localeDir, "locale."+lang+".kv"))
	if err != nil {
		fmt.Println("加载翻译文件失败:", err)
		return
	}
	locales.SwitchI18nLang(string(data))

	langPanel := m.statusBar.Panels().ItemsWithIntToStatusPanel(0)
	langPanel.SetText("当前语言：" + supportedLangs[idx].name)
}

func (m *TMainForm) onGreetClick(sender lcl.IObject) {
	api.ShowMessage("你好！感谢你使用本程序。")
}

func (m *TMainForm) onAboutClick(sender lcl.IObject) {
	var buttons types.TMsgDlgButtons
	buttons = types.NewSet(types.MbOK)
	api.MessageDlg(
		"这是一个 LCL 国际化示例程序。",
		types.MtInformation,
		buttons,
		0,
	)
}

func (m *TMainForm) onOpenSubForm(sender lcl.IObject) {
	subForm.Show()
}

// ==================== 子窗体方法 ====================

func (m *TSubForm) FormCreate(sender lcl.IObject) {
	m.SetName("SubForm")
	m.SetCaption("子窗口")
	m.SetPosition(types.PoScreenCenter)
	m.SetWidth(450)
	m.SetHeight(350)

	// 标题
	m.titleLabel = lcl.NewLabel(m)
	m.titleLabel.SetParent(m)
	m.titleLabel.SetName("SubFormTitleLabel")
	m.titleLabel.SetLeft(20)
	m.titleLabel.SetTop(15)
	m.titleLabel.SetCaption("这是子窗口，支持国际化：")
	m.titleLabel.Font().SetSize(14)
	m.titleLabel.Font().SetBold(true)

	// DateEdit
	dateLabel := lcl.NewLabel(m)
	dateLabel.SetParent(m)
	dateLabel.SetName("DateEditLabel")
	dateLabel.SetLeft(20)
	dateLabel.SetTop(50)
	dateLabel.SetCaption("日期选择：")

	m.dateEdit = lcl.NewDateEdit(m)
	m.dateEdit.SetParent(m)
	m.dateEdit.SetName("DateEdit")
	m.dateEdit.SetLeft(100)
	m.dateEdit.SetTop(47)
	m.dateEdit.SetWidth(200)
	m.dateEdit.SetHint("选择日期")
	m.dateEdit.SetText("")

	// TimeEdit
	timeLabel := lcl.NewLabel(m)
	timeLabel.SetParent(m)
	timeLabel.SetName("TimeEditLabel")
	timeLabel.SetLeft(20)
	timeLabel.SetTop(85)
	timeLabel.SetCaption("时间选择：")

	m.timeEdit = lcl.NewTimeEdit(m)
	m.timeEdit.SetParent(m)
	m.timeEdit.SetName("TimeEdit")
	m.timeEdit.SetLeft(100)
	m.timeEdit.SetTop(82)
	m.timeEdit.SetWidth(200)
	m.timeEdit.SetHint("选择时间")
	m.timeEdit.SetText("")

	// FileNameEdit
	fileLabel := lcl.NewLabel(m)
	fileLabel.SetParent(m)
	fileLabel.SetName("FileNameEditLabel")
	fileLabel.SetLeft(20)
	fileLabel.SetTop(120)
	fileLabel.SetCaption("文件选择：")

	m.fileNameEdit = lcl.NewFileNameEdit(m)
	m.fileNameEdit.SetParent(m)
	m.fileNameEdit.SetName("FileNameEdit")
	m.fileNameEdit.SetLeft(100)
	m.fileNameEdit.SetTop(117)
	m.fileNameEdit.SetWidth(250)
	m.fileNameEdit.SetHint("选择文件")
	m.fileNameEdit.SetTextHint("请选择文件")
	m.fileNameEdit.SetDialogTitle("选择文件")
	m.fileNameEdit.SetFilter("文本文件 (*.txt)|*.txt|所有文件 (*.*)|*.*")
	m.fileNameEdit.SetText("")

	// DirectoryEdit
	dirLabel := lcl.NewLabel(m)
	dirLabel.SetParent(m)
	dirLabel.SetName("DirectoryEditLabel")
	dirLabel.SetLeft(20)
	dirLabel.SetTop(155)
	dirLabel.SetCaption("目录选择：")

	m.directoryEdit = lcl.NewDirectoryEdit(m)
	m.directoryEdit.SetParent(m)
	m.directoryEdit.SetName("DirectoryEdit")
	m.directoryEdit.SetLeft(100)
	m.directoryEdit.SetTop(152)
	m.directoryEdit.SetWidth(250)
	m.directoryEdit.SetHint("选择目录")
	m.directoryEdit.SetTextHint("请选择目录")
	m.directoryEdit.SetDialogTitle("选择目录")
	m.directoryEdit.SetText("")

	// 按钮
	m.okBtn = lcl.NewButton(m)
	m.okBtn.SetParent(m)
	m.okBtn.SetName("SubFormOKBtn")
	m.okBtn.SetBounds(100, 200, 90, 30)
	m.okBtn.SetCaption("确定")
	m.okBtn.SetHint("确认操作")
	m.okBtn.SetOnClick(func(sender lcl.IObject) {
		m.statusLabel.SetCaption("已点击确定")
	})

	m.cancelBtn = lcl.NewButton(m)
	m.cancelBtn.SetParent(m)
	m.cancelBtn.SetName("SubFormCancelBtn")
	m.cancelBtn.SetBounds(200, 200, 90, 30)
	m.cancelBtn.SetCaption("取消")
	m.cancelBtn.SetHint("取消操作")
	m.cancelBtn.SetOnClick(func(sender lcl.IObject) {
		m.Close()
	})

	// 状态标签
	m.statusLabel = lcl.NewLabel(m)
	m.statusLabel.SetParent(m)
	m.statusLabel.SetName("SubFormStatusLabel")
	m.statusLabel.SetLeft(20)
	m.statusLabel.SetTop(250)
	m.statusLabel.SetCaption("状态：等待操作")
}
