# GPUI 快速开始 - 5分钟创建 GUI 应用

## 🚀 安装

```bash
go get github.com/energye/examples/lcl/gpui
```

## 📝 最简示例（20行）

```go
package main

import (
    "github.com/energye/examples/lcl/gpui/ui"
    "github.com/energye/examples/lcl/gpui/widget"
)

func main() {
    // 1. 创建窗口
    window := ui.NewWindow(ui.WindowConfig{
        Title:  "My App",
        Width:  800,
        Height: 600,
    })

    // 2. 设置 UI
    window.OnShow(func() {
        engine := window.Engine()
        
        // 创建按钮
        btn := widget.NewButton("Click Me", widget.ButtonPrimary, engine.Font())
        btn.SetPos(20, 20)
        btn.SetOnClick(func() {
            println("Button clicked!")
        })
        engine.AddWidget(btn)
    })

    // 3. 运行
    window.Run()
}
```

**就这样！** 只需要 3 步。

---

## 🎯 常用控件

### Label（标签）
```go
label := widget.NewLabel("Hello, World!", engine.Font())
label.SetPos(20, 20)
label.SetColor(color.Primary)
engine.AddWidget(label)
```

### Button（按钮）
```go
btn := widget.NewButton("Click", widget.ButtonPrimary, engine.Font())
btn.SetPos(20, 60)
btn.SetOnClick(func() {
    println("Clicked!")
})
engine.AddWidget(btn)
```

**按钮类型**：
- `ButtonDefault` - 默认
- `ButtonPrimary` - 主色（蓝）
- `ButtonSuccess` - 成功（绿）
- `ButtonWarning` - 警告（黄）
- `ButtonDanger` - 危险（红）

### TextBox（文本框）
```go
textbox := widget.NewTextBox("Enter text...", engine.Font())
textbox.SetPos(20, 110)
textbox.SetSize(300, 32)
textbox.SetOnChange(func(text string) {
    println("Text:", text)
})
engine.AddWidget(textbox)
```

---

## 🎨 完整示例

```go
package main

import (
    "fmt"
    "github.com/energye/examples/lcl/gpui/style/color"
    "github.com/energye/examples/lcl/gpui/ui"
    "github.com/energye/examples/lcl/gpui/widget"
)

func main() {
    window := ui.NewWindow(ui.WindowConfig{
        Title:  "Registration Form",
        Width:  400,
        Height: 300,
    })

    window.OnShow(func() {
        engine := window.Engine()
        font := engine.Font()

        // 标题
        title := widget.NewLabel("User Registration", font)
        title.SetPos(20, 20)
        title.SetColor(color.Primary)
        engine.AddWidget(title)

        // 用户名
        userLabel := widget.NewLabel("Username:", font)
        userLabel.SetPos(20, 60)
        engine.AddWidget(userLabel)

        userInput := widget.NewTextBox("Enter username", font)
        userInput.SetPos(100, 60)
        userInput.SetSize(250, 32)
        engine.AddWidget(userInput)

        // 密码
        passLabel := widget.NewLabel("Password:", font)
        passLabel.SetPos(20, 110)
        engine.AddWidget(passLabel)

        passInput := widget.NewTextBox("Enter password", font)
        passInput.SetPos(100, 110)
        passInput.SetSize(250, 32)
        engine.AddWidget(passInput)

        // 提交按钮
        submitBtn := widget.NewButton("Submit", widget.ButtonPrimary, font)
        submitBtn.SetPos(100, 160)
        submitBtn.SetOnClick(func() {
            username := userInput.Text()
            password := passInput.Text()
            fmt.Printf("Username: %s, Password: %s\n", username, password)
        })
        engine.AddWidget(submitBtn)

        // 取消按钮
        cancelBtn := widget.NewButton("Cancel", widget.ButtonDefault, font)
        cancelBtn.SetPos(220, 160)
        cancelBtn.SetOnClick(func() {
            userInput.SetText("")
            passInput.SetText("")
        })
        engine.AddWidget(cancelBtn)

        // 设置焦点
        engine.SetFocus(userInput)
    })

    window.Run()
}
```

---

## ✨ 事件处理

### 按钮点击
```go
btn.SetOnClick(func() {
    // 处理点击
})
```

### 文本变化
```go
textbox.SetOnChange(func(text string) {
    fmt.Println("Text changed:", text)
})
```

### 提交事件
```go
textbox.SetOnSubmit(func(text string) {
    fmt.Println("Submitted:", text)
})
```

---

## 🎯 焦点管理

### 设置焦点
```go
engine.SetFocus(textbox)
```

### Tab 切换
- 自动支持 Tab 键切换焦点
- 自动支持 Shift+Tab 反向切换
- 焦点控件有蓝色边框

---

## 🎨 颜色系统

```go
import "github.com/energye/examples/lcl/gpui/style/color"

// 主色
color.Primary      // #1890ff
color.PrimaryHover // #40a9ff

// 语义色
color.Success      // #52c41a
color.Warning      // #faad14
color.Error        // #ff4d4f

// 文本色
color.TextPrimary   // rgba(0,0,0,0.85)
color.TextSecondary // rgba(0,0,0,0.45)
```

---

## 📐 布局技巧

### 手动定位
```go
widget.SetPos(20, 60)  // x, y
widget.SetSize(200, 32) // width, height
```

### 常见布局
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

## 🔧 高级用法

### 自定义渲染
```go
engine.SetRenderHandler(func(renderer *pipeline.Renderer) {
    // 自定义绘制
    renderer.FillRect(math.NewRect(10, 10, 100, 100), color.Primary)
})
```

### 多窗口
```go
window1 := ui.NewWindow(ui.WindowConfig{Title: "Window 1", Width: 400, Height: 300})
window2 := ui.NewWindow(ui.WindowConfig{Title: "Window 2", Width: 400, Height: 300})

window1.OnShow(func() { setupUI1(window1.Engine()) })
window2.OnShow(func() { setupUI2(window2.Engine()) })

// 只能有一个 Run()
window1.Run()
```

---

## 📊 代码模板

### 基础模板
```go
package main

import (
    "github.com/energye/examples/lcl/gpui/ui"
    "github.com/energye/examples/lcl/gpui/widget"
)

func main() {
    window := ui.NewWindow(ui.WindowConfig{
        Title:  "App",
        Width:  800,
        Height: 600,
    })

    window.OnShow(func() {
        engine := window.Engine()
        // TODO: 创建控件
    })

    window.Run()
}
```

### 表单模板
```go
func setupForm(engine *ui.Engine) {
    font := engine.Font()
    
    // 标题
    title := widget.NewLabel("Form Title", font)
    title.SetPos(20, 20)
    engine.AddWidget(title)

    // 输入字段
    y := float32(60)
    for _, field := range []string{"Name", "Email", "Phone"} {
        label := widget.NewLabel(field+":", font)
        label.SetPos(20, y)
        engine.AddWidget(label)

        input := widget.NewTextBox("Enter "+field, font)
        input.SetPos(120, y)
        input.SetSize(250, 32)
        engine.AddWidget(input)

        y += 50
    }

    // 按钮
    btn := widget.NewButton("Submit", widget.ButtonPrimary, font)
    btn.SetPos(120, y)
    engine.AddWidget(btn)
}
```

---

## 🐛 常见问题

### Q: 字体不显示？
A: 检查字体文件路径，确保有中文字体。

### Q: 控件不响应？
A: 确保调用了 `engine.AddWidget(widget)`

### Q: 窗口不显示？
A: 确保调用了 `window.Run()`

### Q: 如何调试？
A: 添加 `fmt.Println` 在回调中。

---

## 📚 文档

- `README.md` - 项目说明
- `ARCHITECTURE.md` - 架构设计
- `REFACTOR_COMPLETE.md` - 重构总结
- `QUICKSTART.md` - 快速开始（本文件）

---

## 🎉 开始开发

```bash
cd /home/yanghy/app/workspace/examples/lcl/gpui/demo
go run main.go
```

**只需 3 步**：
1. 创建窗口
2. 设置 UI
3. 运行

**开始创建你的 GUI 应用吧！🚀**
