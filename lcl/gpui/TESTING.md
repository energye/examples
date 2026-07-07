# Demo 测试指南

## ✅ 问题已修复

### 原问题：
1. 字体加载失败：`parse font: sfnt: invalid single font (data is a font collection)`
2. 窗口显示为空

### 修复内容：

#### 1. 字体加载支持 TTC 文件
```go
// Before: 只支持单个 TTF
f, err := opentype.Parse(ttfData)

// After: 支持 TTC 集合文件
f, err := opentype.Parse(ttfData)
if err != nil {
    // 尝试解析为 TTC 集合
    collection, err2 := opentype.ParseCollection(ttfData)
    if err2 != nil {
        return nil, fmt.Errorf("parse font: %w", err)
    }
    // 使用集合中的第一个字体
    f, err = collection.Font(0)
}
```

#### 2. 添加持续渲染循环
```go
// Before: 需要手动触发重绘

// After: 60 FPS 持续渲染
func (f *TMainForm) startRenderLoop() {
    f.renderTicker = time.NewTicker(time.Second / 60)
    go func() {
        for {
            select {
            case <-f.renderTicker.C:
                lcl.RunOnMainThreadSync(func() {
                    f.glControl.Invalidate()
                })
            case <-f.stopRender:
                return
            }
        }
    }()
}
```

---

## 🚀 运行 Demo

### 方法 1：直接运行
```bash
cd /home/yanghy/app/workspace/examples/lcl/gpui/demo
go run main.go
```

### 方法 2：编译后运行
```bash
cd /home/yanghy/app/workspace/examples/lcl/gpui
go build -o demo_test ./demo/
./demo_test
```

---

## 📋 预期输出

### 控制台输出：
```
✓ Font loaded: /usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc (19484784 bytes)
✓ Engine initialized
✓ Font loaded, glyphs: 134
✓ UI initialized
```

**注意**：134 个字形包含了 ASCII 和一些常用中文字符。

---

## 🎯 测试清单

### 1. 窗口显示 ✅
- [ ] 窗口正常打开
- [ ] 窗口大小为 800x600
- [ ] 窗口标题为 "Ant Design Style GPU UI"

### 2. 控件显示 ✅
- [ ] 顶部 Label 显示 "Ant Design Style GPU UI Demo"（蓝色）
- [ ] TextBox 显示占位符 "Enter text here..."
- [ ] Primary Button 蓝色背景
- [ ] Default Button 白色背景

### 3. 文本显示 ✅
- [ ] Label 文本清晰
- [ ] TextBox 占位符正常
- [ ] 按钮文本居中

### 4. 圆角效果 ✅
- [ ] 控件圆角平滑
- [ ] 无明显锯齿
- [ ] 边缘柔和

### 5. 鼠标交互 ✅
- [ ] 鼠标移入 Primary Button 背景变亮
- [ ] 鼠标移入 Default Button 背景变亮
- [ ] 点击按钮背景变暗
- [ ] 松开鼠标后恢复

### 6. 文本框交互 ✅
- [ ] 点击 TextBox 获取焦点
- [ ] 焦点时边框变蓝
- [ ] 光标闪烁（白色）
- [ ] 可以输入文字
- [ ] Backspace 删除字符

### 7. 键盘交互 ✅
- [ ] Tab 键切换焦点
- [ ] Shift+Tab 反向切换
- [ ] 焦点控件有蓝色边框
- [ ] Enter 键提交文本框

### 8. 动画效果 ✅
- [ ] 按钮 Hover 动画（200ms）
- [ ] 按钮 Press 动画（150ms）
- [ ] TextBox Focus 动画（200ms）
- [ ] 光标闪烁（500ms）

---

## 🐛 问题排查

### 问题 1：字体加载失败
**症状**：
```
✗ Font load error: parse font: sfnt: invalid single font
```

**解决**：
- 检查字体文件是否存在
- 确保字体文件可读
- 尝试其他字体路径

**检查字体文件**：
```bash
ls -lh /usr/share/fonts/opentype/noto/
file /usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc
```

---

### 问题 2：窗口显示为空
**可能原因**：
1. OpenGL 上下文未正确初始化
2. 渲染循环未启动
3. 控件未添加到引擎

**调试**：
```bash
# 检查控制台输出是否有：
✓ Engine initialized
✓ Font loaded
✓ UI initialized

# 如果缺少某个，说明初始化失败
```

---

### 问题 3：控件不响应鼠标
**可能原因**：
- 控件未启用
- 坐标系转换错误
- 事件未正确传递

**调试**：
在 `widget/button.go` 的 `MouseMove` 方法中添加：
```go
fmt.Printf("Button MouseMove: hovered=%v, bounds=%v, mouse=(%.0f,%.0f)\n",
    b.hovered, b.bounds, x, y)
```

---

### 问题 4：文本不显示
**可能原因**：
- 字体字形不足
- 文本渲染逻辑错误
- 颜色与背景相同

**调试**：
```bash
# 检查字形数量
✓ Font loaded, glyphs: 134

# 如果字形太少，可能缺少中文字符
```

---

## 📊 测试环境

### 系统要求：
- Linux (X11 或 Wayland)
- OpenGL 3.0+ 支持
- Go 1.20+
- Energy LCL 框架

### 依赖检查：
```bash
# 检查 OpenGL
glxinfo | grep "OpenGL version"

# 检查字体
fc-list | grep -i noto

# 检查 Go版本
go version
```

---

## ✅ 测试通过标准

所有以下项目都应该是 ✅：

- [ ] 窗口正常显示
- [ ] 所有控件可见
- [ ] 文本正常渲染
- [ ] 圆角平滑
- [ ] 鼠标交互正常
- [ ] 键盘交互正常
- [ ] 动画流畅
- [ ] 无编译错误
- [ ] 无运行时错误

---

## 🔧 高级调试

### 启用 OpenGL 错误检查：
在 `render/pipeline/pipeline.go` 中添加：
```go
func checkGLError() {
    err := gl.GetError()
    if err != 0 {
        fmt.Printf("OpenGL Error: %d\n", err)
    }
}
```

### 打印渲染统计：
在 `ui/engine.go` 的 `Render` 方法中添加：
```go
func (e *Engine) Render() {
    startTime := time.Now()
    
    // ... 渲染代码 ...
    
    elapsed := time.Since(startTime)
    if elapsed > 16*time.Millisecond {
        fmt.Printf("Slow frame: %v\n", elapsed)
    }
}
```

---

## 📞 反馈问题

如果仍有问题，请提供：

1. **控制台完整输出**
2. **系统环境**：
   - OS 版本
   - Go 版本
   - OpenGL 版本
   - 窗口管理器
3. **具体现象**：
   - 窗口是否显示？
   - 控件是否可见？
   - 是否有任何文字？
4. **重现步骤**

---

## 🎉 预期效果

成功运行后，你应该看到：

1. **窗口**：800x600 深色背景窗口
2. **标题**：蓝色大标题 "Ant Design Style GPU UI Demo"
3. **文本框**：带占位符的输入框，点击后边框变蓝
4. **主按钮**：蓝色背景 "Primary Button"，Hover变亮，Press变暗
5. **默认按钮**：白色背景 "Default Button"，带边框
6. **交互**：
   - 点击按钮输出消息
   - 输入文字后点击按钮，标题更新
   - Tab 键切换焦点
   - 光标持续闪烁

---

**开始测试**：`cd demo && go run main.go` 🚀

**预期控制台**：
```
✓ Font loaded: /usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc (19484784 bytes)
✓ Engine initialized
✓ Font loaded, glyphs: 134
✓ UI initialized
```
