# Ant Design 控件绘制能力需求分析

## 📊 渲染能力矩阵

| 能力 | Checkbox | Radio | Switch | Slider | Progress | Select | Tabs |
|------|----------|-------|--------|--------|----------|--------|------|
| 矩形 | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| 圆角矩形 | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| 圆形 | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ |
| 圆环 | ✅ | ✅ | ❌ | ❌ | ✅ | ❌ | ❌ |
| 线条 | ❌ | ❌ | ❌ | ✅ | ❌ | ❌ | ✅ |
| 文本 | ✅ | ✅ | ❌ | ❌ | ✅ | ✅ | ✅ |
| 图标 | ✅ | ✅ | ❌ | ❌ | ❌ | ✅ | ✅ |
| 渐变 | ❌ | ❌ | ❌ | ❌ | ✅ | ❌ | ❌ |
| 阴影 | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ | ❌ |
| 动画 | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |

---

## 🎨 控件详细需求

### 1. Checkbox（复选框）
**渲染需求**：
- ✅ 矩形边框（unchecked）
- ✅ 圆角矩形（checked）
- ✅ 勾选图标（✓）
- ✅ 动画（200ms）
- ✅ 禁用状态

**绘制步骤**：
```go
// 1. 绘制边框
renderer.StrokeRoundRect(rect, 4, 1.5, borderColor)

// 2. 如果选中，填充背景
if checked {
    renderer.FillRoundRect(rect, 4, primaryColor)
}

// 3. 绘制勾选图标
if checked {
    drawCheckmark(renderer, rect, whiteColor)
}
```

### 2. Radio（单选框）
**渲染需求**：
- ✅ 圆形边框（unchecked）
- ✅ 圆形填充（checked）
- ✅ 内部小圆点
- ✅ 动画

**绘制步骤**：
```go
// 1. 绘制外圆
renderer.StrokeCircle(center, radius, borderWidth, borderColor)

// 2. 如果选中，绘制内圆
if selected {
    renderer.FillCircle(center, innerRadius, primaryColor)
}
```

### 3. Switch（开关）
**渲染需求**：
- ✅ 圆角矩形轨道
- ✅ 圆形滑块
- ✅ 滑动动画

**绘制步骤**：
```go
// 1. 绘制轨道
renderer.FillRoundRect(trackRect, trackHeight/2, trackColor)

// 2. 绘制滑块
renderer.FillCircle(sliderCenter, sliderRadius, whiteColor)

// 3. 添加阴影
renderer.DrawShadow(sliderRect, shadow)
```

### 4. Slider（滑块）
**渲染需求**：
- ✅ 轨道线条
- ✅ 已填充部分
- ✅ 拖动手柄（圆形）
- ✅ 悬浮提示

**绘制步骤**：
```go
// 1. 绘制背景轨道
renderer.FillRoundRect(trackRect, trackHeight/2, bgColor)

// 2. 绘制已填充部分
renderer.FillRoundRect(fillRect, trackHeight/2, primaryColor)

// 3. 绘制手柄
renderer.FillCircle(handleCenter, handleRadius, primaryColor)
renderer.StrokeCircle(handleCenter, handleRadius, 2, whiteColor)
```

### 5. Progress（进度条）
**渲染需求**：
- ✅ 圆角矩形轨道
- ✅ 进度填充（可渐变）
- ✅ 圆形进度（可选）
- ✅ 百分比文本

**绘制步骤**：
```go
// 线性进度条
renderer.FillRoundRect(trackRect, radius, bgColor)
renderer.FillRoundRect(progressRect, radius, primaryColor)

// 圆形进度条
renderer.StrokeCircle(center, radius, strokeWidth, bgColor)
renderer.StrokeArc(center, radius, strokeWidth, startAngle, endAngle, primaryColor)
```

### 6. Select（下拉框）
**渲染需求**：
- ✅ 输入框（已有）
- ✅ 下拉箭头图标
- ✅ 下拉面板（阴影）
- ✅ 选项列表
- ✅ 悬浮高亮

### 7. Tabs（标签页）
**渲染需求**：
- ✅ 标签文本
- ✅ 指示器线条
- ✅ 动画过渡
- ✅ 内容面板

---

## 🔧 需要添加的绘制原语

### 高优先级（必须）：

#### 1. 圆形绘制
```go
// 填充圆
func (r *Renderer) FillCircle(center math.Vec2, radius float32, color math.Color)

// 圆环
func (r *Renderer) StrokeCircle(center math.Vec2, radius, width float32, color math.Color)

// 圆弧
func (r *Renderer) StrokeArc(center math.Vec2, radius, width float32, startAngle, endAngle float32, color math.Color)
```

#### 2. 线条绘制
```go
// 直线
func (r *Renderer) DrawLine(x1, y1, x2, y2, width float32, color math.Color)

// 折线
func (r *Renderer) DrawPolyline(points []math.Vec2, width float32, color math.Color)
```

#### 3. 图标支持
```go
// 图标字体
func (r *Renderer) DrawIcon(icon rune, x, y, size float32, color math.Color)

// 纹理图标
func (r *Renderer) DrawIconTexture(texture uint32, src, dst math.Rect, color math.Color)
```

### 中优先级（重要）：

#### 4. 渐变填充
```go
// 线性渐变
func (r *Renderer) FillRectLinearGradient(rect math.Rect, start, end math.Color, angle float32)

// 径向渐变
func (r *Renderer) FillRectRadialGradient(rect math.Rect, center math.Vec2, inner, outer math.Color)
```

#### 5. 阴影效果
```go
// 投影阴影
func (r *Renderer) DrawShadow(rect math.Rect, shadow Shadow)

type Shadow struct {
    OffsetX, OffsetY float32
    Blur             float32
    Spread           float32
    Color            math.Color
}
```

#### 6. 裁剪支持
```go
// 矩形裁剪
func (r *Renderer) PushClipRect(rect math.Rect)
func (r *Renderer) PopClipRect()
```

### 低优先级（增强）：

#### 7. 路径绘制
```go
// 路径操作
func (r *Renderer) BeginPath()
func (r *Renderer) MoveTo(x, y float32)
func (r *Renderer) LineTo(x, y float32)
func (r *Renderer) QuadTo(cx, cy, x, y float32)  // 二次贝塞尔
func (r *Renderer) CubicTo(cx1, cy1, cx2, cy2, x, y float32) // 三次贝塞尔
func (r *Renderer) ClosePath()
func (r *Renderer) FillPath(color math.Color)
func (r *Renderer) StrokePath(width float32, color math.Color)
```

#### 8. 变换支持
```go
func (r *Renderer) PushTransform()
func (r *Renderer) PopTransform()
func (r *Renderer) Translate(dx, dy float32)
func (r *Renderer) Rotate(angle float32)
func (r *Renderer) Scale(sx, sy float32)
```

---

## 🏗️ 扩展性设计

### 当前架构优势：

```go
// 1. Widget 接口清晰
type Widget interface {
    Render(renderer *pipeline.Renderer)
    // ...其他方法
}

// 2. 渲染器可扩展
type Renderer struct {
    // 可以添加新方法
}

// 3. 着色器系统灵活
type ShaderManager struct {
    // 可以添加新着色器
}
```

### 自定义控件示例：

```go
// 自定义进度条
type ProgressBar struct {
    widget.BaseWidget
    progress float32  // 0-1
    color    math.Color
}

func (pb *ProgressBar) Render(renderer *pipeline.Renderer) {
    // 1. 背景轨道
    trackRect := pb.Bounds()
    renderer.FillRoundRect(trackRect, trackRect.H/2, bgColor)

    // 2. 进度填充
    fillWidth := trackRect.W * pb.progress
    fillRect := math.NewRect(trackRect.X, trackRect.Y, fillWidth, trackRect.H)
    renderer.FillRoundRect(fillRect, fillRect.H/2, pb.color)

    // 3. 文本
    text := fmt.Sprintf("%.0f%%", pb.progress*100)
    renderer.DrawText(text, trackRect.X+trackRect.W/2, trackRect.Y, font, textColor)
}
```

---

## 📋 实现优先级

### Phase 1：核心能力（1-2 周）
- ✅ 圆形绘制（FillCircle, StrokeCircle）
- ✅ 线条绘制（DrawLine）
- ✅ 改进圆角边框

### Phase 2：视觉增强（1 周）
- 渐变填充
- 阴影效果
- 裁剪支持

### Phase 3：高级控件（2-3 周）
- Checkbox
- Radio
- Switch
- Slider
- Progress

### Phase 4：复杂控件（3-4 周）
- Select
- Tabs
- Table
- Tree

---

## ✅ 总结

### 当前能力：
- ✅ 基础图形（矩形、圆角矩形）
- ✅ 文本渲染
- ✅ 纹理渲染
- ✅ 动画系统

### 缺少能力：
- ❌ 圆形绘制
- ❌ 渐变
- ❌ 阴影
- ❌ 图标系统

### 扩展性：
- ✅ Widget 接口清晰
- ✅ 渲染器可扩展
- ✅ 着色器系统灵活

### 结论：
**当前框架可以扩展支持 Ant Design 控件，但需要添加圆形绘制等基础能力。**
