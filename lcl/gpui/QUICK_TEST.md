# 快速测试指南

## 🚀 运行 Demo

```bash
cd /home/yanghy/app/workspace/examples/lcl/gpui/demo
go run main.go
```

## ✅ 预期控制台输出

```
✓ Font loaded: /usr/share/fonts/opentype/noto/NotoSansCJK-Light.ttc (18160344 bytes)
✓ Engine initialized
✓ Font loaded, glyphs: 134
✓ UI initialized
```

**注意**：加载了 **NotoSansCJK-Light** 字体（更轻薄、更清晰）

---

## 🧪 测试清单

### 1. 光标测试（30秒）

#### 步骤：
1. 点击 TextBox
2. 观察光标
3. 输入几个字符
4. 等待 2-3秒

#### 检查项：
- [ ] ✅ 光标出现在文本框内
- [ ] ✅ 光标距离左边框约 8px
- [ ] ✅ 光标持续闪烁（约 1秒显示，1秒隐藏）
- [ ] ✅ 输入文字后光标向右移动
- [ ] ✅ 光标闪烁不中断

**预期**：光标应该像标准文本框一样持续闪烁

---

### 2. 字体测试（20秒）

#### 步骤：
1. 查看 Label 文本
2. 查看 Button 文本
3. 查看 TextBox 占位符

#### 检查项：
- [ ] ✅ 中文字符清晰显示
- [ ] ✅ 字体平滑无锯齿
- [ ] ✅ 文本对齐正确
- [ ] ✅ 类似微软雅黑效果（轻薄、清晰）

**预期**：字体应该清晰、轻薄、专业

---

### 3. 窗口调整测试（30秒）

#### 步骤：
1. 拖动窗口右下角调整大小
2. 快速拖动几次
3. 缓慢拖动
4. 最小化再恢复

#### 检查项：
- [ ] ✅ 控件位置固定不动
- [ ] ✅ 无上下抖动
- [ ] ✅ 渲染流畅（无闪烁）
- [ ] ✅ 窗口边缘拖动平滑

**预期**：控件应该固定在原位，不随窗口调整而抖动

---

### 4. 交互测试（40秒）

#### 按钮测试：
1. 鼠标移入 Primary Button
2. 观察背景变亮
3. 点击按钮
4. 观察背景变暗
5. 松开鼠标

#### 检查项：
- [ ] ✅ Hover：背景变亮（200ms 动画）
- [ ] ✅ Press：背景变暗（150ms 动画）
- [ ] ✅ Click：触发事件，Label 更新
- [ ] ✅ 控制台输出 "Button clicked!"

#### TextBox 测试：
1. 点击 TextBox
2. 观察边框变蓝
3. 输入文字
4. 按 Backspace 删除

#### 检查项：
- [ ] ✅ Focus：边框变蓝（200ms 动画）
- [ ] ✅ Input：文字正确显示
- [ ] ✅ Delete：Backspace 删除字符
- [ ] ✅ Cursor：光标跟随移动

---

### 5. 焦点切换测试（20秒）

#### 步骤：
1. 点击 TextBox
2. 按 Tab 键
3. 按 Tab 键
4. 按 Shift+Tab

#### 检查项：
- [ ] ✅ Tab：焦点切换到下一个控件
- [ ] ✅ Shift+Tab：焦点切换到上一个控件
- [ ] ✅ 焦点控件有蓝色边框
- [ ] ✅ 焦点切换平滑

---

## 🐛 问题排查

### 问题 1：光标不闪烁
**检查**：
- 是否点击了 TextBox？
- 是否等待足够时间（500ms）？
- 控制台是否有错误？

**调试**：
在 textbox.go 的 Render 中添加：
```go
fmt.Printf("Cursor: focused=%v, cursorT=%.2f\n", tb.focused, cursorT)
```

---

### 问题 2：字体不清晰
**检查**：
```
✓ Font loaded: /usr/share/fonts/opentype/noto/NotoSansCJK-Light.ttc
```
应该显示 "Light" 版本

**备选字体**：
如果 Light 不够好，尝试修改 demo/main.go：
```go
"/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc",  // 标准版
"/usr/share/fonts/truetype/wqy/wqy-microhei.ttc",         // 文泉驿
```

---

### 问题 3：窗口抖动
**检查**：
- 是否快速拖动窗口？
- 是否在调整大小时观察控件？

**原因**：
可能是因为渲染循环和窗口调整不同步

**解决**：
确保 OnPaint 中更新尺寸：
```go
w := float32(f.Width())
h := float32(f.Height())
f.engine.SetSize(w, h)
```

---

## 📊 性能指标

### 渲染性能：
- **FPS**: 60（流畅）
- **光标闪烁**: 1Hz（500ms开/500ms关）
- **动画**: 200ms（Hover/Focus）

### 内存使用：
- **字体图集**: 2048x2048
- **字形数量**: 134（ASCII + 常用中文）
- **纹理内存**: ~16MB

---

## ✨ 成功标准

所有测试通过：
- ✅ 光标持续闪烁
- ✅ 字体清晰美观
- ✅ 窗口调整稳定
- ✅ 交互响应正常
- ✅ 动画流畅

---

## 📞 反馈问题

如果仍有问题，请提供：

1. **操作系统**：Linux/macOS/Windows
2. **字体文件**：
   ```bash
   ls /usr/share/fonts/opentype/noto/
   ```
3. **控制台输出**：完整输出
4. **具体现象**：详细描述
5. **重现步骤**：如何操作

---

## 🎉 预期效果

运行成功后，你应该看到：

1. **窗口**：800x600 深色背景
2. **标题**：蓝色 "Ant Design Style GPU UI Demo"
3. **TextBox**：
   - 占位符 "Enter text here..."
   - 点击后边框变蓝
   - 光标持续闪烁
4. **Primary Button**：蓝色背景，Hover变亮
5. **Default Button**：白色背景，带边框
6. **交互**：
   - 点击按钮输出消息
   - 输入文字更新标题
   - Tab 切换焦点

---

**开始测试**：`cd demo && go run main.go` 🚀

**预期字体**：NotoSansCJK-Light（更清晰）
