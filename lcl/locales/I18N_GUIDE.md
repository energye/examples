# LCL 国际化使用指南

通过配置文件实现多语言支持，无需修改代码。

---

## 快速开始

**1. 创建翻译文件** `locale.zh-CN.kv`

```properties
MainForm.Caption=我的应用
SubmitBtn.Caption=提交
NameEdit.TextHint=请输入姓名
```

**2. 加载翻译**

```go
data, _ := os.ReadFile("locale.zh-CN.kv")
locales.SwitchI18nLang(string(data))
```

**3. 切换语言**

```go
data, _ := os.ReadFile("locale.en-US.kv")
locales.SwitchI18nLang(string(data))
```

---

## 配置规则

### 基本格式

```properties
组件名.属性名=翻译值
```

### 下拉列表/选项组

```properties
ComboBox.Items[0]=选项一
ComboBox.Items[1]=选项二
RadioGroup.Items[0]=男
RadioGroup.Items[1]=女
```

### 列表头/状态栏

```properties
ListView.Columns[0].Caption=名称
ListView.Columns[1].Caption=大小
StatusBar.Panels[0].Text=就绪
```

### 带标签编辑框

```properties
UsernameEdit.EditLabel.Caption=用户名：
EmailEdit.EditLabel.Caption=邮箱：
```

### 动作组件

```properties
ActionNew.Caption=新建
ActionNew.Hint=创建新文件
```

---

## 支持的组件

### 常用属性

| 属性 | 说明 |
|------|------|
| `Caption` | 标题/文本（按钮、标签、菜单、窗体等） |
| `Hint` | 提示信息（所有可视组件） |
| `Text` | 文本内容（编辑框、备忘录） |
| `TextHint` | 占位符（编辑框、下拉框） |
| `Title` | 对话框标题 |
| `Items[N]` | 列表项（下拉框、列表框、单选组、复选组） |

### 按钮

```properties
Btn.Caption=确定
Btn.Hint=确认操作
```

### 输入框

```properties
Edit.TextHint=请输入内容
Edit.Hint=输入提示
Memo.Hint=在此输入备注
```

### 下拉框/列表框

```properties
Combo.Items[0]=选项一
Combo.Items[1]=选项二
ListBox.Items[0]=项目一
CheckList.Items[0]=复选项一
```

### 单选组/复选组

```properties
RadioGroup.Caption=请选择
RadioGroup.Items[0]=选项一
RadioGroup.Items[1]=选项二

CheckGroup.Caption=功能选项
CheckGroup.Items[0]=自动保存
CheckGroup.Items[1]=显示行号
```

### 标签/面板/分组

```properties
Label.Caption=标题文本
Panel.Caption=面板内容
GroupBox.Caption=分组标题
TabSheet.Caption=选项卡标题
```

### 菜单

```properties
MenuFile.Caption=文件(&F)
MenuFileNew.Caption=新建(&N)
MenuFileNew.Hint=创建新文件
MenuFileSave.Caption=保存(&S)
```

### 工具栏

```properties
ToolBtnNew.Caption=新建
ToolBtnNew.Hint=创建新文件
```

### 列表头

```properties
ListView.Columns[0].Caption=名称
ListView.Columns[1].Caption=类型
HeaderControl.Sections[0].Text=标题
```

### 状态栏

```properties
StatusBar.Panels[0].Text=就绪
StatusBar.Panels[1].Text=版本 1.0
```

### 对话框

```properties
OpenDialog.Title=打开文件
SaveDialog.Title=保存文件
FindDialog.Title=查找
ColorDialog.Title=选择颜色
```

### 系统托盘

```properties
TrayIcon.Hint=我的应用
TrayIcon.BalloonTitle=通知
TrayIcon.BalloonHint=正在后台运行
```

---

## 支持的 TStrings 组件

| 组件 | 格式 |
|------|------|
| TComboBox | `ComboBox.Items[N]=值` |
| TListBox | `ListBox.Items[N]=值` |
| TCheckListBox | `CheckListBox.Items[N]=值` |
| TRadioGroup | `RadioGroup.Items[N]=值` |
| TCheckGroup | `CheckGroup.Items[N]=值` |

## 支持的 TCollection 组件

| 组件 | 格式 |
|------|------|
| TListView | `ListView.Columns[N].Caption=值` |
| THeaderControl | `HeaderControl.Sections[N].Text=值` |
| TStatusBar | `StatusBar.Panels[N].Text=值` |

## 支持的 SubComponent

| 组件 | 格式 |
|------|------|
| TLabeledEdit | `Edit.EditLabel.Caption=值` |
| TButtonPanel | `Panel.OKButton.Caption=值` |
| TButtonPanel | `Panel.CancelButton.Caption=值` |

---

## API 函数

```go
// 加载翻译配置
locales.SwitchI18nLang(data string) bool

// 切换 LCL 内置语言
locales.SwitchDefaultLang(lang string, forceUpdate bool) bool

// 切换 LCL 框架字符串语言
locales.SwitchLCLLang(lang string) bool
```

---

## 常见问题

**Q: 组件没被翻译？**
- 检查组件是否设置了 `Name`
- 检查配置 key 是否匹配（大小写不敏感）
- 检查文件是否正确加载

**Q: 如何翻译 ComboBox 选项？**
```properties
ComboBox.Items[0]=选项一
ComboBox.Items[1]=选项二
```

**Q: 如何翻译 ListView 列头？**
```properties
ListView.Columns[0].Caption=名称
ListView.Columns[1].Caption=大小
```

**Q: 如何翻译 TLabeledEdit 标签？**
```properties
Edit.EditLabel.Caption=用户名：
```

**Q: 支持哪些字符？**
支持所有 UTF-8 字符：中文、日文、韩文、欧洲语言特殊字符等。

---

## 语言代码

| 代码 | 语言 |
|------|------|
| `zh-CN` | 简体中文 |
| `zh-TW` | 繁体中文 |
| `en-US` | 美国英语 |
| `ja` | 日语 |
| `ko` | 韩语 |
| `de` | 德语 |
| `fr` | 法语 |
| `es` | 西班牙语 |
| `ru` | 俄语 |
