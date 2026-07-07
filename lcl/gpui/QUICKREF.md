# GPUI 快速参考 - 底层库 vs 示例程序

## 🎯 核心原则

**底层库负责**：怎么做（How）
- OpenGL 管理
- 渲染管线
- 事件分发
- 资源管理

**示例程序负责**：做什么（What）
- 创建控件
- 设置属性
- 业务逻辑

---

## 📝 代码模板

### 最简示例（60行）：
```go
package main

import (
    "github.com/energye/examples/lcl/gpui/ui"
    "github.com/energye/examples/lcl/gpui/widget"
    "github.com/energye/lcl/lcl"
    "github.com/energye/lcl/types"
)

type MainForm struct {
    lcl.TEngForm
    glControl lcl.IOpenGLControl
    engine    *ui.Engine
    initialized bool
}

var mainForm = &MainForm{}

func main() {
    lcl.Init()
    lcl.RunApp(mainForm)
}

func (f *MainForm) FormCreate(sender lcl.IObject) {
    f.SetCaption("My App")
    f.SetWidth(800)
    f.SetHeight(600)

    // 创建 GL 控件
    f.glControl = lcl.NewOpenGLControl(f)
    f.glControl.SetParent(f)
    f.glControl.SetAlign(types.AlClient)
    f.glControl.SetOnPaint(f.onPaint)

    // 创建引擎
    f.engine = ui.NewEngine()
}

func (f *MainForm) FormShow(sender lcl.IObject) {
    // 初始化
    f.glControl.MakeCurrent(true)
    defer f.glControl.ReleaseContext()
    f.engine.Initialize()
    f.engine.SetSize(800, 600)

    // 加载字体
    font, _ := ui.LoadDefaultFont(14)
    f.engine.SetFont(font)

    // 设置 UI
    f.setupUI()
    f.initialized = true
}

func (f *MainForm) setupUI() {
    engine := f.engine
    font := engine.Font()

    // 添加控件
    label := widget.NewLabel("Hello", font)
    engine.AddWidget(label)

    btn := widget.NewButton("Click", widget.ButtonPrimary, font)
    btn.SetOnClick(func() {
        label.SetText("Clicked!")
    })
    engine.AddWidget(btn)
}

func (f *MainForm) onPaint(sender lcl.IObject) {
    if !f.initialized {
        return
    }
    f.glControl.MakeCurrent(true)
    defer f.glControl.ReleaseContext()
    f.engine.Render()
    f.glControl.SwapBuffers()
}
```

---

## 🔧 底层库 API

### Engine（核心）：
```go
// 创建和初始化
engine := ui.NewEngine()
engine.Initialize()          // 初始化 GL（需要上下文）
engine.Delete()              // 释放资源

// 渲染
engine.Render()              // 渲染一帧（需要上下文）

// 控件管理
engine.AddWidget(widget)     // 添加控件
engine.SetFocus(widget)      // 设置焦点

// 事件处理
engine.HandleMouseDown(x, y, button)
engine.HandleMouseUp(x, y, button)
engine.HandleMouseMove(x, y)
engine.HandleKeyDown(key, mods)
engine.HandleCharInput(char)

// 属性
engine.Font()                // 获取字体
engine.SetFont(font)         // 设置字体
engine.SetSize(w, h)         // 设置尺寸
engine.CursorTime()          // 光标时间（用于动画）
```

### Font（字体）：
```go
// 加载字体
font, err := ui.LoadDefaultFont(14)

// 或者从文件加载
data, _ := os.ReadFile("font.ttc")
ui.SetDefaultFontData(data)
font, _ := ui.LoadDefaultFont(14)

// 使用
engine.SetFont(font)
```

---

## 🧩 控件 API

### Label（标签）：
```go
label := widget.NewLabel("Text", font)
label.SetPos(20, 20)
label.SetSize(200, 24)
label.SetColor(color.Primary)
engine.AddWidget(label)
```

### Button（按钮）：
```go
btn := widget.NewButton("Click", widget.ButtonPrimary, font)
btn.SetPos(20, 60)
btn.SetSize(120, 32)
btn.SetOnClick(func() {
    // 处理点击
})
engine.AddWidget(btn)

// 按钮类型
widget.ButtonDefault  // 默认
widget.ButtonPrimary  // 主色（蓝）
widget.ButtonSuccess  // 成功（绿）
widget.ButtonWarning  // 警告（黄）
widget.ButtonDanger   // 危险（红）
```

### TextBox（文本框）：
```go
textbox := widget.NewTextBox("Placeholder", font)
textbox.SetPos(20, 110)
textbox.SetSize(300, 32)
textbox.SetOnChange(func(text string) {
    // 文本变化
})
textbox.SetOnSubmit(func(text string) {
    // 提交（Enter）
})
engine.AddWidget(textbox)
```

---

## 🎨 颜色系统

```go
import "github.com/energye/examples/lcl/gpui/style/color"

// 主色
color.Primary       // #1890ff
color.PrimaryHover  // #40a9ff
color.PrimaryActive // #096dd9

// 语义色
color.Success       // #52c41a
color.Warning       // #faad14
color.Error         // #ff4d4f

// 文本色
color.TextPrimary    // rgba(0,0,0,0.85)
color.TextSecondary  // rgba(0,0,0,0.45)
color.TextDisabled   // rgba(0,0,0,0.25)
```

---

## 📐 布局

### 定位：
```go
widget.SetPos(x, y)   // 设置位置
widget.SetSize(w, h)  // 设置大小
```

### 常见布局：
```go
// 水平排列
btn1.SetPos(20, 100)
btn2.SetPos(150, 100)
btn3.SetPos(280, 100)

// 垂直排列
label1.SetPos(20, 20)
label2.SetPos(20, 60)
label3.SetPos(20, 100)
```

---

## 🎯 焦点管理

```go
// 设置焦点
engine.SetFocus(textbox)

// Tab 切换（自动支持）
// 按 Tab 切换到下一个焦点控件
// 按 Shift+Tab 切换到上一个

// 焦点效果
// - 文本框：蓝色边框 + 光标闪烁
// - 按钮：蓝色边框
```

---

## 🖱️ 事件处理

### 鼠标事件：
```go
// 在 onPaint 等回调中
engine.HandleMouseDown(x, y, button)  // button: 0=左, 2=右
engine.HandleMouseUp(x, y, button)
engine.HandleMouseMove(x, y)
```

### 键盘事件：
```go
engine.HandleKeyDown(key, mods)  // mods: 1=Shift, 2=Ctrl
engine.HandleCharInput(char)
```

### 控件回调：
```go
btn.SetOnClick(func() { ... })
textbox.SetOnChange(func(text string) { ... })
textbox.SetOnSubmit(func(text string) { ... })
```

---

## 🎨 动画系统

### 自动动画：
- ✅ 按钮 Hover：200ms 变亮
- ✅ 按钮 Press：150ms 变暗
- ✅ 文本框 Focus：200ms 边框变蓝
- ✅ 光标：500ms 闪烁循环

### 自定义动画：
```go
import "github.com/energye/examples/lcl/gpui/style/animation"

// 创建动画
anim := animation.NewAnimation(0, 1, 200*time.Millisecond, animation.EaseOut)

// 循环动画
loopAnim := animation.NewLoopAnimation(0, 1, 500*time.Millisecond, animation.Linear)

// 使用
value := anim.Value()
```

---

## 📊 代码对比

### 旧版（300+ 行）：
```go
type MainForm struct {
    glControl    lcl.IOpenGLControl
    engine       *ui.Engine
    renderTicker *time.Ticker
    stopRender   chan bool
    // ...很多字段
}

func (f *MainForm) FormCreate(sender lcl.IObject) {
    // 50+ 行：GL 初始化、事件绑定
}

func (f *MainForm) startRenderLoop() {
    // 20+ 行：渲染循环
}

func (f *MainForm) OnPaint(sender lcl.IObject) {
    // 10+ 行：渲染逻辑
}
```

### 新版（60 行）：
```go
type MainForm struct {
    lcl.TEngForm
    glControl lcl.IOpenGLControl
    engine    *ui.Engine
    // 最少的字段
}

func (f *MainForm) setupUI() {
    // 只关注业务逻辑
    label := widget.NewLabel("Hello", engine.Font())
    engine.AddWidget(label)
}
```

---

## ✅ 最佳实践

### 1. 职责分离
- **底层库**：框架层功能
- **示例程序**：业务逻辑

### 2. API 使用
```go
// ✅ 正确
engine.AddWidget(widget)
engine.Render()

// ❌ 错误
// 直接操作 GL
// 管理渲染循环
```

### 3. 资源管理
```go
// 底层库自动管理
engine.Delete()  // 自动清理

// 不需要手动：
// - 停止定时器
// - 清理 goroutine
// - 释放 GL 资源
```

---

## 🐛 常见问题

### Q: 编译错误 "undefined: xxx"？
A: 检查导入路径是否正确。

### Q: 窗口不显示？
A: 确保调用了 `engine.Initialize()` 和 `engine.Render()`。

### Q: 控件不响应？
A: 确保调用了 `engine.AddWidget(widget)`。

### Q: 字体不显示？
A: 检查字体文件路径，确保调用了 `ui.SetDefaultFontData()`。

### Q: 光标不闪烁？
A: 使用 `widget.NewLoopAnimation` 创建循环动画。

---

## 📚 文档

- `README.md` - 项目说明
- `ARCHITECTURE.md` - 架构设计
- `ARCHITECTURE_FINAL.md` - 重构总结
- `QUICKSTART.md` - 快速开始
- **本文件** - 快速参考

---

## 🚀 快速开始

```bash
cd /home/yanghy/app/workspace/examples/lcl/gpui/demo
go run main.go
```

**只需 3 步**：
1. 创建引擎
2. 添加控件
3. 渲染

**开始创建你的 GUI 应用！🎨**
