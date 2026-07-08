# GPUI 迭代开发文档

> 本文档用于持续跟踪 GPUI 框架的开发迭代，记录每次更新的目标、完成情况、不足之处及下一步计划。

---

## 文档使用说明

- 每次完成一个阶段或重要修复后，在对应条目更新状态
- 状态标记：⬜ 未开始 → 🔄 进行中 → ✅ 已完成 → ⚠️ 已完成但有遗留
- 每次更新后在 [CHANGELOG.md](CHANGELOG.md) 追加迭代日志
- 当前阶段完成后，在「阶段总结」记录目标达成和不足

---

## 当前状态总览

| 阶段 | 名称 | 状态 | 开始日期 | 完成日期 |
|------|------|------|----------|----------|
| Phase 0 | 审计与规划 | ✅ 已完成 | 2026-07-08 | 2026-07-08 |
| Phase 1 | 致命 Bug 修复 + 颜色系统统一 | ✅ 已完成 | 2026-07-08 | 2026-07-08 |
| Phase 2 | 架构补全 | ✅ 已完成 | 2026-07-08 | 2026-07-08 |
| Phase 3 | 功能补全 | ⬜ 未开始 | - | - |
| Phase 4 | 控件库扩展 | ⬜ 未开始 | - | - |

**当前进度**：Phase 0 ✅ + Phase 1 ✅ + Phase 2 ✅ 全部完成，准备进入 Phase 3

### 渲染测试覆盖总览

| 指标 | 数值 |
|------|------|
| 渲染测试用例总数 | 25 |
| 已有测试 | 18 |
| 待实现测试 | 7 |
| 控件级测试（Phase 4） | 8（R-C01 ~ R-C08） |
| 当前覆盖率 | 72% |

---

## Phase 0：审计与规划 ✅

**目标**：全面审查代码库，识别所有 Bug、架构缺陷和功能缺失

**完成日期**：2026-07-08

### 审计结果汇总

#### 致命 Bug（渲染错误）

| ID | 问题 | 文件 | 状态 |
|----|------|------|------|
| B-001 | `Mat4.Multiply` 列主序乘法顺序反了，所有嵌套变换错误 | `core/math/math.go:235-245` | ✅ |
| B-002 | 文本渲染忽略基线偏移（ascent/bearingY），字形垂直对齐错乱 | `render/pipeline/primitives.go:26-44` | ✅ |
| B-003 | `StrokeRect` 四角重叠绘制，半透明边框出现暗角 | `render/pipeline/text.go:53-63` | ✅ |
| B-004 | 渐变着色器坐标空间混乱，transform 激活时渐变方向错误 | `render/shader/shader.go:398-437` | ✅ |

#### 严重 Bug（行为错误）

| ID | 问题 | 文件 | 状态 |
|----|------|------|------|
| B-005 | `Container.Remove` 不清理 pointerCapture/hoverChild，被移除控件仍接收事件 | `widget/container.go:58-72` | ✅ |
| B-006 | Engine 对双击同时派发 MouseDown+DoubleClick+MouseUp，三次激活 | `ui/engine.go:252-268` | ✅ |
| B-007 | `Box.HandleEvent` 禁用状态下仍吞噬鼠标事件 | `widget/primitives.go:79-95` | ✅ |

#### 逻辑错误

| ID | 问题 | 文件 | 状态 |
|----|------|------|------|
| B-008 | `takeLine` 在 index=0 处断行错误（`lastSpace > 0` 应为 `>= 0`） | `render/pipeline/text.go:121` | ✅ |
| B-009 | `FillRoundRectWithBorder` 内圆角半径计算错误（`radius - borderWidth` 应为 `radius - borderWidth/2`） | `render/pipeline/text.go:115-116` | ✅ |
| B-010 | 圆角 SDF 在 `radius > min(halfW, halfH)` 时退化 | `render/shader/shader.go:330-343` | ✅ |
| B-011 | `FocusManager.SetFocusable(false)` 不清除已设置的 `focused` 标志 | `widget/base.go:130-149` | ✅ |
| B-012 | `LayoutContainer` Measure 阶段重复计算 layout | `widget/layout.go:162-175` | ✅ |

#### 架构缺陷

| ID    | 问题                                                             | 影响范围 | 状态 |
|-------|----------------------------------------------------------------|----------|------|
| A-001 | 三套冲突的颜色/主题系统（color.go v4 / token.go v5 / theme.go 独立）          | 整个样式层 | ✅ |
| A-002 | 两套重复的动画系统（motion/ 和 style/animation/），均未接入控件, 波纹效果，滑块切换动态效果 等等和其他控件的特效能力                 | 整个动画层 | ⬜ |
| A-003 | 焦点系统无法跨容器工作，registerFocusable 硬编码 Container/LayoutContainer 类型 | 事件系统 | ✅ |
| A-004 | Widget 接口缺少生命周期钩子（OnMount/OnUnmount/OnResize/OnStateChanged）   | 控件框架 | ✅ |
| A-005 | SetOwner 必须手动调用，忘记调用导致 HitTest/SetFocusable 行为错误               | 控件框架 | ✅ |
| A-006 | Application 使用全局 currentApp 变量，不支持多实例                          | UI 层 | ✅ |
| A-007 | 无 DPI 缩放支持，Engine.Context() 硬编码 Scale=1                        | 渲染层 | ✅ |

#### 功能缺失

| ID | 缺失功能 | 所属模块 | 状态 |
|----|----------|----------|------|
| F-001 | 颜色缺少 HSL 空间操作（FromHSL/ToHSL/Saturate/Desaturate） | core/math | ✅ |
| F-002 | 无 10 级色板生成（Ant Design color-1 到 color-10） | style/token | ✅ |
| F-003 | Token 派生与 Ant Design 不一致（圆角/间距/字号/暗色模式） | style/token | ✅ |
| F-004 | 组件 Token 仅覆盖 4 个（Button/Input/Card/Modal），Ant Design 有 60+ | style/token | ✅ |
| F-005 | Mat4 缺少 Inverse/Transpose/Shear | core/math | ✅ |
| F-006 | 路径系统无贝塞尔曲线，无法渲染 SVG 图标 | render/pipeline | ✅ |
| F-007 | 无圆角裁剪（PushClip 仅支持矩形） | render/pipeline | ✅ |
| F-008 | 无 FBO 离屏渲染（模糊/阴影/backdrop-filter 需要） | render/pipeline | ✅ |
| F-009 | 无纹理管理 API（image.Image → GPU 纹理） | render/texture | ✅ |
| F-010 | 布局缺少 flex-shrink / flex-basis / align-self | layout | ✅ |
| F-011 | flex-grow 分配后不执行 min/max 钳位 | layout/flex | ✅ |
| F-012 | Grid 不支持 fr 单元和 grid-column/row 跨格 | layout/grid | ✅ |
| F-013 | Overlay FocusTrap 字段存在但未实现任何逻辑 | overlay | ✅ |
| F-014 | 无焦点环渲染 | widget | ✅ |
| F-015 | 无键盘快捷键（Escape/方向键等） | widget/event | ✅ |
| F-016 | 缺少基础控件（Input/Select/Checkbox/Radio/Switch/Slider/Tabs/Menu/Modal） | widget | ✅ |

#### 性能问题

| ID | 问题 | 文件 | 状态 |
|----|------|------|------|
| P-001 | `UniformSet.key()` 每次绘制分配字符串+排序 | `render/pipeline/pipeline.go:66-82` | ✅ |
| P-002 | 阴影用 3-16 个圆角矩形叠加，应改为专用 shadow shader | `render/pipeline/primitives.go:194-222` | ✅ |
| P-003 | 文本换行每行分配 `[]rune`，应改用 range 迭代 | `render/pipeline/text.go:109,140` | ✅ |
| P-004 | VBO 每个 batch 从 offset 0 重传，应使用 ring buffer | `render/pipeline/pipeline.go:353-362` | ✅ |

---

## Phase 1：致命 Bug 修复 + 颜色系统统一 ✅

**目标**：修复所有导致渲染错误的 Bug，统一颜色系统为单一 Token 体系

**预计工作项**：14 项

### 1.1 数学与渲染 Bug 修复 ✅

| 任务 | 关联问题 | 负责模块 | 验证测试 | 状态 |
|------|----------|----------|----------|------|
| 修复 Mat4.Multiply 列主序乘法 | B-001 | `core/math/math.go` | R-007, R-018 | ✅ |
| 修复文本基线偏移，添加 ascent/bearingY 计算 | B-002 | `render/pipeline/primitives.go` | R-012, R-013 | ✅ |
| 修复 StrokeRect 四角重叠，改为不重叠的四段 | B-003 | `render/pipeline/text.go` | R-001 | ✅ |
| 修复渐变着色器坐标空间，统一使用屏幕空间 | B-004 | `render/shader/shader.go` | R-006, R-007 | ✅ |
| 修复 takeLine index=0 断行 | B-008 | `render/pipeline/text.go` | R-016 | ✅ |
| 修复 FillRoundRectWithBorder 内圆角半径 | B-009 | `render/pipeline/text.go` | R-004 | ✅ |
| 修复圆角 SDF 半径退化（添加 clamp） | B-010 | `render/shader/shader.go` | R-002 | ✅ |

### 1.2 颜色系统统一 ✅

| 任务 | 关联问题 | 负责模块 | 状态 |
|------|----------|----------|------|
| Color 添加 HSL 空间操作（FromHSL/ToHSL/Saturate/Desaturate/HueRotate） | F-001 | `core/math/math.go` | ✅ |
| 实现 10 级色板生成算法（基于 HSL 亮度阶梯） | F-002 | 新建 `style/token/palette.go` | ✅ |
| Token 派生修正（圆角/间距/字号对齐 Ant Design v5） | F-003 | `style/token/token.go` | ✅ |
| 废弃 `style/color/color.go`，统一使用 Token 系统 | A-001 | `style/color/` | ✅ |
| 废弃 `style/theme/theme.go`，统一使用 Token 系统 | A-001 | `style/theme/` | ✅ |
| 暗色模式色板重新生成（不只是背景/文本/边框） | F-003 | `style/token/token.go` | ✅ |

### Phase 1 完成标准

- [x] Mat4 变换嵌套顺序正确（可通过嵌套 Container 验证）
- [x] 文本垂直对齐正确（不同字形基线一致）
- [x] 半透明边框无暗角
- [x] 渐变在有 transform 时方向正确
- [x] 颜色系统只有一个入口（token.Current()）
- [x] Lighten/Darken 在 HSL 空间操作
- [x] 10 级色板可正确生成
- [x] 暗色模式下组件颜色正确调整

### Phase 1 渲染测试要求

每个 Bug 修复必须同时完成对应渲染测试，测试通过后标记为 ✅：

| 测试 ID | 测试名称 | 关联修复 | 状态 |
|---------|----------|----------|------|
| R-001 | 基础矩形绘制 | B-003 修复后重新验证 | ✅ 已有 |
| R-002 | 圆角矩形填充 | B-010 修复后新建 | ✅ 已实现 |
| R-004 | 圆角矩形填充+描边 | B-009 修复后新建 | ✅ 已实现 |
| R-006 | 线性渐变 | B-004 修复后新建 | ✅ 已实现 |
| R-007 | 渐变+Transform 嵌套 | B-001+B-004 修复后新建 | ✅ 已实现 |
| R-012 | 文本渲染 ASCII | B-002 修复后新建 | ✅ 已实现 |
| R-013 | 文本渲染 CJK | B-002 修复后新建 | ✅ 已实现 |
| R-016 | 文本自动换行 | B-008 修复后新建 | ✅ 已实现 |
| R-018 | Transform 嵌套 | B-001 修复后新建 | ✅ 已实现 |

---

## Phase 2：架构补全 ✅

**目标**：修复控件层 Bug，补全架构缺陷，使框架具备生产级基础

**预计工作项**：12 项

### 2.1 控件层 Bug 修复 ✅

| 任务 | 关联问题 | 负责模块 | 状态 |
|------|----------|----------|------|
| Container.Remove 清理 pointerCapture/hoverChild | B-005 | `widget/container.go` | ✅ |
| LayoutContainer.Remove 同步修复 | B-005 | `widget/layout.go` | ✅ |
| Engine 双击事件改为只派发 DoubleClick（不同时派发 MouseDown） | B-006 | `ui/engine.go` | ✅ |
| Box.HandleEvent 禁用时不再吞噬事件 | B-007 | `widget/primitives.go` | ✅ |
| FocusManager.SetFocusable(false) 清除 focused 标志 | B-011 | `widget/base.go` | ✅ |
| LayoutContainer.Measure 缓存结果避免重复计算 | B-012 | `widget/layout.go` | ✅ |

### 2.2 焦点系统重构 ✅

| 任务 | 关联问题 | 负责模块 | 状态 |
|------|----------|----------|------|
| 焦点管理改为全局协调（Engine 层单一 FocusManager） | A-003 | `widget/focus.go` + `ui/engine.go` | ✅ |
| registerFocusable 改用 Children() 接口而非类型硬编码 | A-003 | `widget/container.go` | ✅ |
| Portal 关闭时自动恢复之前的焦点 | A-003 | `widget/portal.go` | ✅ |

### 2.3 生命周期钩子 ✅

| 任务 | 关联问题 | 负责模块 | 状态 |
|------|----------|----------|------|
| Widget 接口添加 OnMount/OnUnmount 可选接口 | A-004 | `widget/types.go` | ✅ |
| Container.Add 触发 OnMount，Remove 触发 OnUnmount | A-004 | `widget/container.go` | ✅ |
| Layout 时 bounds 变化触发 OnResize 回调 | A-004 | `widget/base.go` | ✅ |

### Phase 2 完成标准

- [x] 移除子控件后不再有悬挂引用或事件泄漏
- [x] 双击只触发一次回调
- [x] 禁用控件不阻止事件传递
- [x] Tab 键可以跨 Container 导航
- [x] Portal 关闭后焦点正确恢复
- [x] 第三方容器类型的子控件可正确参与焦点系统
- [x] 控件有 Mount/Unmount/Resize 回调

### Phase 2 渲染测试要求

Phase 2 主要修复事件和控件逻辑，渲染测试以现有测试的回归验证为主：

| 测试 ID | 测试名称 | 验证内容 | 状态 |
|---------|----------|----------|------|
| R-024 | GPU 端到端渲染 | 修复后 Demo 应用完整渲染一帧无崩溃 | ⬜ |

---

## Phase 3：功能补全 ⬜

**目标**：补全渲染管线和布局引擎的缺失能力，使框架能支撑完整 UI

**预计工作项**：10 项

### 3.1 渲染管线增强 ⬜

| 任务 | 关联问题 | 负责模块 | 状态 |
|------|----------|----------|------|
| 贝塞尔曲线路径（Quadratic/Cubic Bezier） | F-006 | `render/pipeline/path.go` | ⬜ |
| 圆角裁剪（PushClip 支持 rounded rect） | F-007 | `render/pipeline/pipeline.go` | ✅ |
| FBO 离屏渲染支持 | F-008 | `render/pipeline/` + `core/gl/` | ✅ |
| 纹理管理 API（从 image.Image 创建/删除纹理） | F-009 | `render/texture/texture.go` | ⬜ |
| Mat4 添加 Inverse/Transpose/Shear | F-005 | `core/math/math.go` | ⬜ |

### 3.2 布局引擎增强 ⬜

| 任务 | 关联问题 | 负责模块 | 状态 |
|------|----------|----------|------|
| 添加 flex-shrink 支持 | F-010 | `layout/flex.go` | ⬜ |
| 添加 flex-basis 支持 | F-010 | `layout/layout.go` + `layout/flex.go` | ⬜ |
| flex-grow 分配后执行 min/max 钳位 | F-011 | `layout/flex.go` | ⬜ |
| Grid 添加 fr 单元和跨格支持 | F-012 | `layout/grid.go` | ⬜ |
| 添加 JustifySpaceAround/SpaceEvenly | F-010 | `layout/flex.go` | ⬜ |

### Phase 3 完成标准

- [ ] SVG 图标可正确渲染（贝塞尔曲线）
- [ ] Card/Modal 可实现圆角 overflow:hidden
- [ ] 模糊/阴影效果可通过 FBO 实现
- [ ] 图片可从 Go image.Image 加载到 GPU
- [ ] 嵌套变换的逆变换正确（坐标转换）
- [ ] Flex 布局支持 shrink/basis，溢出时正确收缩
- [ ] Grid 支持跨格和 fr 单元

### Phase 3 渲染测试要求

新增渲染能力必须有对应测试：

| 测试 ID | 测试名称 | 关联功能 | 状态 |
|---------|----------|----------|------|
| R-019 | Clip 裁剪 | F-007 圆角裁剪基础验证 | ⬜ |
| R-020 | Clip + Transform 组合 | F-007 裁剪+变换交互 | ⬜ |

---

## Phase 4：控件库扩展 ⬜

**目标**：实现 Ant Design 核心控件，使框架可用于实际应用

**预计工作项**：按批次递增

### 4.1 动画系统接入 ⬜

| 任务 | 关联问题 | 负责模块 | 状态 |
|------|----------|----------|------|
| 合并 motion/ 和 style/animation/ 为统一动画系统 | A-002 | 保留 `motion/`，废弃 `style/animation/` | ⬜ |
| Engine 渲染循环传递 dt 给活跃动画 | A-002 | `ui/engine.go` | ⬜ |
| Widget 状态变化触发过渡动画（hover/active/focus 颜色渐变） | A-002 | `widget/interaction.go` | ⬜ |

### 4.2 基础控件实现 ⬜

| 任务 | 关联问题 | 负责模块 | 状态 |
|------|----------|----------|------|
| Input（文本输入框，带前后缀/计数/清除） | F-016 | 新建 `widget/input.go` | ⬜ |
| Checkbox（复选框，支持半选） | F-016 | 新建 `widget/checkbox.go` | ✅ |
| Radio（单选框） | F-016 | 新建 `widget/radio.go` | ✅ |
| Switch（开关） | F-016 | 新建 `widget/switch.go` | ✅ |
| Select（下拉选择，带搜索） | F-016 | 新建 `widget/select.go` | ⬜ |
| Tabs（标签页） | F-016 | 新建 `widget/tabs.go` | ⬜ |
| Menu（导航菜单） | F-016 | 新建 `widget/menu.go` | ⬜ |
| Modal（模态对话框） | F-016 | 新建 `widget/modal.go` | ⬜ |

### 4.3 无障碍增强 ⬜

| 任务 | 关联问题 | 负责模块 | 状态 |
|------|----------|----------|------|
| 焦点环渲染（FocusRing 组件） | F-014 | 新建 `widget/focus_ring.go` | ⬜ |
| Escape 键关闭 Modal/Popover | F-015 | `widget/portal.go` | ⬜ |
| 方向键导航 Menu/Tabs | F-015 | 各控件实现 | ⬜ |
| Overlay FocusTrap 实际实现 | F-013 | `overlay/overlay.go` + `widget/portal.go` | ⬜ |

### Phase 4 完成标准

- [ ] 控件状态变化有平滑过渡动画
- [ ] Input 支持受控/非受控模式
- [ ] Checkbox/Radio 可正确切换和表单提交
- [ ] Modal 可正确打开/关闭，焦点锁定在内部
- [ ] 所有交互控件有可见的焦点环
- [ ] Escape 可关闭弹层

### Phase 4 渲染测试要求

每个新控件必须有对应的渲染快照测试，覆盖所有视觉状态：

| 测试 ID | 测试名称 | 渲染内容 | 状态 |
|---------|----------|----------|------|
| R-008 | 纯色矩形绘制 | 5 个不透明矩形 + 2 个半透明重叠矩形 | ⬜ |
| R-009 | 直线绘制 | 8 条不同角度辐射线段 | ⬜ |
| R-010 | 圆形绘制 | 4 个不同大小圆形 + 描边圆 | ⬜ |
| R-011 | 圆弧绘制 | 4 段不同角度圆弧（填充+描边） | ⬜ |
| R-014 | 文本对齐 | 左/中/右对齐三行文本 | ⬜ |
| R-015 | 文本省略号 | 窄矩形内长文本截断 | ⬜ |
| R-017 | 混合中英文 | 中英文混合段落 | ⬜ |
| R-021 | BoxStyle 完整渲染 | Card 模拟 + Button 模拟（阴影+渐变+描边） | ⬜ |
| R-022 | Checkmark 图标 | Checkbox 选中状态勾选图标 | ⬜ |

**控件级渲染测试**（每个新控件实现时同步创建）：

| 测试 ID | 控件 | 渲染状态 |
|---------|------|----------|
| R-C01 | Button | 默认/Hover/Pressed/Focus/Disabled/Loading 六态 |
| R-C02 | Input | 空态/有文本/Placeholder/Focus/Error/Disabled 六态 |
| R-C03 | Checkbox | 未选/选中/半选/Disabled 四态 |
| R-C04 | Radio | 未选/选中/Disabled 三态 |
| R-C05 | Switch | Off/On/Disabled 三态 |
| R-C06 | Tabs | 默认/选中/禁用 三态 |
| R-C07 | Modal | 打开状态（遮罩+内容+关闭按钮） |
| R-C08 | Menu | 水平/垂直布局，展开/收起状态 |

---

## GPU 渲染单元测试计划

> 每个渲染功能必须有对应的 GPU 渲染单元测试，测试输出 PNG 图片用于人工对比验证。

### 测试基础设施

| 工具 | 路径 | 用途 |
|------|------|------|
| CPU 软件渲染快照 | `render/pipeline/visual_snapshot_test.go` | 无需 GPU，CPU 光栅化生成参考图 |
| GPU 帧缓冲捕获 | `render/pipeline/snapshot.go` → `SavePNG()` | 真实 GPU 渲染结果截取 |
| 快照验证工具 | `cmd/validate_snapshot/main.go` | 自动校验 PNG 尺寸/颜色数/非空像素 |
| 端到端 GPU 测试 | `scripts/gtk3_gpu_snapshot.sh` | 启动应用 → 渲染一帧 → 截图 → 验证 |
| 输出目录 | `test_output/render_core/` | 所有快照 PNG 存放位置 |

**环境变量**：
- `GPUI_TEST_OUTPUT`：自定义快照输出目录（默认 `test_output/render_core/`）
- `GPUI_GPU_SNAPSHOT`：设置后 Demo 应用渲染一帧即截图退出

### 测试分类

#### A. CPU 软件渲染测试（无需 GPU，CI 可运行）

适用于：纯几何绘制（矩形、圆角、路径、阴影）。使用 `visual_snapshot_test.go` 中的 CPU SDF 光栅化器。

#### B. GPU 渲染测试（需要 OpenGL 上下文）

适用于：着色器效果（渐变、SDF 圆角）、字体渲染、纹理绘制。使用 `scripts/gtk3_gpu_snapshot.sh` 或 `GPUI_GPU_SNAPSHOT` 环境变量。

### 渲染功能测试清单

#### R-001：基础矩形绘制 ✅

| 项目 | 内容 |
|------|------|
| 测试函数 | `TestWriteCoreDrawingSnapshots` (已有) |
| 输出文件 | `test_output/render_core/core_shapes.png` |
| 渲染内容 | 白色背景上绘制多个矩形：纯色填充矩形、描边矩形 |
| 预期效果 | 白底上可见多个不同颜色的矩形，描边宽度均匀，无重叠暗角 |
| 验证方式 | `validate_snapshot -min-colors 4 -min-non-bg 5000` |
| 关联问题 | B-003（StrokeRect 暗角修复后需重新验证） |
| 状态 | ✅ 已有测试 |

#### R-002：圆角矩形填充（SDF）🔄

| 项目 | 内容 |
|------|------|
| 测试函数 | `TestRoundRectFill` |
| 输出文件 | `test_output/render_core/round_rect_fill.png` |
| 渲染内容 | 640×420 画布上绘制 **5 组圆角矩形**： |
| | 1. 小圆角（radius=4, 32×32）— 按钮尺寸 |
| | 2. 中圆角（radius=8, 120×48）— 输入框尺寸 |
| | 3. 大圆角（radius=16, 200×80）— 卡片尺寸 |
| | 4. 胶囊形（radius=9999, 160×48）— Tag/Pill |
| | 5. 圆形（radius=60, 120×120）— 头像 |
| 预期效果 | 圆角边缘平滑无锯齿，无像素化阶梯。radius=9999 时两端为完美半圆。圆形无变形 |
| 验证方式 | `validate_snapshot -min-colors 3 -min-non-bg 10000` |
| 关联问题 | B-010（SDF 半径退化修复后需验证 radius > min(halfW, halfH)） |
| 状态 | ⬜ 待实现 |

#### R-003：圆角矩形描边（SDF）🔄

| 项目 | 内容 |
|------|------|
| 测试函数 | `TestRoundRectStroke` |
| 输出文件 | `test_output/render_core/round_rect_stroke.png` |
| 渲染内容 | 在浅灰背景上绘制 **4 组描边圆角矩形**： |
| | 1. 细描边（width=1, radius=4） |
| | 2. 中描边（width=2, radius=8） |
| | 3. 粗描边（width=4, radius=16） |
| | 4. 胶囊形描边（width=2, radius=9999, 160×48） |
| 预期效果 | 描边宽度一致，圆角处无断裂或变粗。描边内外边缘均平滑 |
| 验证方式 | `validate_snapshot -min-colors 3 -min-non-bg 3000` |
| 状态 | ⬜ 待实现 |

#### R-004：圆角矩形填充+描边组合 ⬜

| 项目 | 内容 |
|------|------|
| 测试函数 | `TestRoundRectWithBorder` |
| 输出文件 | `test_output/render_core/round_rect_with_border.png` |
| 渲染内容 | 绘制 **3 组带描边的填充圆角矩形**（模拟按钮）： |
| | 1. 蓝色填充 + 深蓝描边（Primary 按钮） |
| | 2. 白色填充 + 灰色描边（Default 按钮） |
| | 3. 红色填充 + 深红描边（Danger 按钮） |
| 预期效果 | 填充与描边之间无间隙、无重叠暗色边缘。内圆角与外圆角同心 |
| 验证方式 | `validate_snapshot -min-colors 6 -min-non-bg 8000` |
| 关联问题 | B-009（内圆角半径修复后验证） |
| 状态 | ⬜ 待实现 |

#### R-005：阴影效果 ⬜

| 项目 | 内容 |
|------|------|
| 测试函数 | `TestShadow` |
| 输出文件 | `test_output/render_core/shadow.png` |
| 渲染内容 | 白色背景上绘制 **3 组带阴影的矩形**： |
| | 1. 小阴影（offset=(0,1), blur=2, color=rgba(0,0,0,0.06)）— 按钮默认 |
| | 2. 中阴影（offset=(0,6), blur=16, color=rgba(0,0,0,0.08)）— 卡片 |
| | 3. 大阴影（offset=(0,8), blur=24, color=rgba(0,0,0,0.12)）— 弹窗 |
| 预期效果 | 阴影从矩形边缘向外柔和扩散，底部偏重（符合光源在上方）。无明显分层阶梯 |
| 验证方式 | `validate_snapshot -min-colors 8 -min-non-bg 5000` |
| 状态 | ⬜ 待实现 |

#### R-006：线性渐变 ⬜

| 项目 | 内容 |
|------|------|
| 测试函数 | `TestLinearGradient` |
| 输出文件 | `test_output/render_core/linear_gradient.png` |
| 渲染内容 | 绘制 **4 组渐变矩形**： |
| | 1. 水平渐变（左→右，蓝→绿） |
| | 2. 垂直渐变（上→下，红→黄） |
| | 3. 对角渐变（左上→右下，紫→橙） |
| | 4. 圆角渐变矩形（左→右，蓝→绿，radius=8） |
| 预期效果 | 渐变过渡平滑无 banding，方向正确。圆角渐变在圆角处无断裂 |
| 验证方式 | `validate_snapshot -min-colors 32 -min-non-bg 15000` |
| 关联问题 | B-004（渐变坐标空间修复后验证。测试时添加一个带 transform 的子用例） |
| 状态 | ⬜ 待实现 |

#### R-007：渐变+Transform 嵌套 ✅

| 项目 | 内容 |
|------|------|
| 测试函数 | `TestGradientWithTransform` |
| 输出文件 | `test_output/render_core/gradient_transform.png` |
| 渲染内容 | 1. 先 PushTransform(TranslationMatrix(100, 50, 0)) |
| | 2. 在变换后的位置绘制水平渐变矩形（蓝→绿） |
| | 3. 再 PushTransform(ScaleMatrix(1.5, 1.5, 1)) |
| | 4. 在缩放后的位置绘制垂直渐变矩形（红→黄） |
| 预期效果 | 两个渐变矩形位置正确（偏移和缩放生效），渐变方向与矩形方向一致而非屏幕方向 |
| 验证方式 | 人工对比：渐变方向应跟随矩形而非全局坐标 |
| 关联问题 | B-001（Mat4 乘法修复后验证）、B-004 |
| 状态 | ⬜ 待实现 |

#### R-008：纯色矩形绘制 ⬜

| 项目 | 内容 |
|------|------|
| 测试函数 | `TestFillRect` |
| 输出文件 | `test_output/render_core/fill_rect.png` |
| 渲染内容 | 白色背景上绘制 **5 个不透明纯色矩形**（红/绿/蓝/黄/灰）和 **2 个半透明矩形**（alpha=0.5 的红和蓝，互相重叠） |
| 预期效果 | 不透明矩形边缘锐利。半透明重叠区域颜色正确混合（红+蓝=紫，alpha 正确） |
| 验证方式 | `validate_snapshot -min-colors 8 -min-non-bg 8000` |
| 状态 | ⬜ 待实现 |

#### R-009：直线绘制 ⬜

| 项目 | 内容 |
|------|------|
| 测试函数 | `TestDrawLine` |
| 输出文件 | `test_output/render_core/draw_line.png` |
| 渲染内容 | 绘制 **8 条不同角度的线段**（0°/45°/90°/135°/180°/225°/270°/315°），线宽 2px，从中心向外辐射 |
| 预期效果 | 线条宽度均匀，无断裂，端点整齐。对角线无明显锯齿 |
| 验证方式 | `validate_snapshot -min-colors 2 -min-non-bg 500` |
| 状态 | ⬜ 待实现 |

#### R-010：圆形绘制（SDF）⬜

| 项目 | 内容 |
|------|------|
| 测试函数 | `TestFillCircle` |
| 输出文件 | `test_output/render_core/fill_circle.png` |
| 渲染内容 | 绘制 **4 个圆形**： |
| | 1. 小圆（radius=16）— 图标尺寸 |
| | 2. 中圆（radius=32）— 头像尺寸 |
| | 3. 大圆（radius=64）— 大头像 |
| | 4. 描边圆（radius=32, stroke=2）— 选中状态 |
| 预期效果 | 圆形边缘平滑无锯齿，无椭圆变形。描边宽度一致 |
| 验证方式 | `validate_snapshot -min-colors 4 -min-non-bg 3000` |
| 状态 | ⬜ 待实现 |

#### R-011：圆弧绘制 ⬜

| 项目 | 内容 |
|------|------|
| 测试函数 | `TestDrawArc` |
| 输出文件 | `test_output/render_core/draw_arc.png` |
| 渲染内容 | 绘制 **4 段圆弧**： |
| | 1. 90° 扇形（0°→90°，蓝色） |
| | 2. 180° 半圆（0°→180°，绿色） |
| | 3. 270° 弧（45°→315°，红色描边） |
| | 4. 小角度弧（0°→60°，黄色描边） |
| 预期效果 | 弧线平滑，端点整齐。扇形填充区域正确。无多余三角形溢出 |
| 验证方式 | `validate_snapshot -min-colors 5 -min-non-bg 2000` |
| 状态 | ⬜ 待实现 |

#### R-012：文本渲染（ASCII）✅

| 项目 | 内容 |
|------|------|
| 测试函数 | `TestDrawTextASCII` |
| 输出文件 | `test_output/render_core/text_ascii.png` |
| 渲染内容 | 绘制 **3 行文本**： |
| | 1. "Hello, GPUI! 1234567890"（常规文本） |
| | 2. "ABCDEFGHIJKLMNOPQRSTUVWXYZ"（大写字母） |
| | 3. "abcdefghijklmnopqrstuvwxyz!@#$%^&*()"（小写+符号） |
| 预期效果 | 字形清晰，基线一致（所有字母底部对齐）。无重叠、无错位 |
| 验证方式 | `validate_snapshot -min-colors 2 -min-non-bg 1000` |
| 关联问题 | B-002（文本基线修复后验证） |
| 状态 | ⬜ 待实现 |

#### R-013：文本渲染（CJK 中文）✅

| 项目 | 内容 |
|------|------|
| 测试函数 | `TestDrawTextCJK` |
| 输出文件 | `test_output/render_core/text_cjk.png` |
| 渲染内容 | 绘制 **2 行中文**： |
| | 1. "你好世界，测试中文渲染" |
| | 2. "按钮 标签 文本框 窗口 程序开发" |
| 预期效果 | 中文字形完整无缺损，与 ASCII 文本基线对齐。字间距均匀 |
| 验证方式 | `validate_snapshot -min-colors 2 -min-non-bg 500` |
| 状态 | ⬜ 待实现 |

#### R-014：文本对齐（左/中/右）⬜

| 项目 | 内容 |
|------|------|
| 测试函数 | `TestTextAlign` |
| 输出文件 | `test_output/render_core/text_align.png` |
| 渲染内容 | 在 3 个相同矩形区域内分别绘制左对齐、居中、右对齐的文本 "Ant Design" |
| 预期效果 | 左对齐文本紧贴左边框，居中文本水平居中，右对齐文本紧贴右边框。三行垂直位置一致 |
| 验证方式 | `validate_snapshot -min-colors 3 -min-non-bg 500` |
| 状态 | ⬜ 待实现 |

#### R-015：文本省略号截断 ⬜

| 项目 | 内容 |
|------|------|
| 测试函数 | `TestTextEllipsis` |
| 输出文件 | `test_output/render_core/text_ellipsis.png` |
| 渲染内容 | 在窄矩形内绘制长文本 "This is a very long text that should be truncated with ellipsis"，启用 Ellipsis=true |
| 预期效果 | 文本在矩形右边缘截断，末尾显示 "..."。无文本溢出矩形边界 |
| 验证方式 | `validate_snapshot -min-colors 2 -min-non-bg 200` |
| 状态 | ⬜ 待实现 |

#### R-016：文本自动换行 ✅

| 项目 | 内容 |
|------|------|
| 测试函数 | `TestTextWrap` |
| 输出文件 | `test_output/render_core/text_wrap.png` |
| 渲染内容 | 在 200px 宽矩形内绘制长段落 "The quick brown fox jumps over the lazy dog. Pack my box with five dozen liquor jugs."，MaxLines=3 |
| 预期效果 | 文本在单词边界处换行（不截断单词），最多 3 行。行间距一致 |
| 验证方式 | `validate_snapshot -min-colors 2 -min-non-bg 500` |
| 关联问题 | B-008（takeLine index=0 修复后验证） |
| 状态 | ⬜ 待实现 |

#### R-017：文本多行混合中英文 ⬜

| 项目 | 内容 |
|------|------|
| 测试函数 | `TestTextMixedLang` |
| 输出文件 | `test_output/render_core/text_mixed.png` |
| 渲染内容 | 混合中英文段落："GPUI 是一个 GPU 加速的 UI 框架，基于 Ant Design 设计规范实现。" |
| 预期效果 | 中英文基线一致，字间距自然。无字符重叠或间距异常 |
| 验证方式 | `validate_snapshot -min-colors 2 -min-non-bg 300` |
| 状态 | ⬜ 待实现 |

#### R-018：Transform 嵌套（平移+缩放）✅

| 项目 | 内容 |
|------|------|
| 测试函数 | `TestTransformNested` |
| 输出文件 | `test_output/render_core/transform_nested.png` |
| 渲染内容 | 1. 绘制一个红色矩形（基准位置） |
| | 2. PushTransform(TranslationMatrix(150, 0, 0))，绘制蓝色矩形 |
| | 3. PushTransform(ScaleMatrix(0.5, 0.5, 1))，绘制绿色矩形 |
| | 4. PopTransform × 2 |
| | 5. 在原点绘制黄色矩形（验证栈恢复） |
| 预期效果 | 蓝色矩形右移 150px，绿色矩形在蓝色基础上缩小 50% 并偏移，黄色矩形回到原点 |
| 验证方式 | 人工对比：各矩形位置和大小应符合变换矩阵语义 |
| 关联问题 | B-001（Mat4 乘法修复后验证） |
| 状态 | ⬜ 待实现 |

#### R-019：Clip 裁剪 ⬜

| 项目 | 内容 |
|------|------|
| 测试函数 | `TestClipRect` |
| 输出文件 | `test_output/render_core/clip_rect.png` |
| 渲染内容 | 1. PushClip(Rect(50, 50, 200, 150)) |
| | 2. 绘制一个大矩形（300×200），部分超出裁剪区域 |
| | 3. PopClip |
| | 4. 在裁剪区域外绘制一个小矩形（验证裁剪栈恢复） |
| 预期效果 | 大矩形在裁剪区域外的部分不可见，边缘锐利。PopClip 后绘制不受影响 |
| 验证方式 | `validate_snapshot -min-colors 3 -min-non-bg 3000` |
| 状态 | ⬜ 待实现 |

#### R-020：Clip + Transform 组合 ⬜

| 项目 | 内容 |
|------|------|
| 测试函数 | `TestClipWithTransform` |
| 输出文件 | `test_output/render_core/clip_transform.png` |
| 渲染内容 | 1. PushTransform(TranslationMatrix(100, 50, 0)) |
| | 2. PushClip(Rect(0, 0, 150, 100)) — 裁剪区域跟随变换 |
| | 3. 绘制一个大矩形 |
| | 4. PopClip + PopTransform |
| 预期效果 | 裁剪区域在变换后的位置生效，大矩形在变换后裁剪区域外的部分不可见 |
| 验证方式 | 人工对比：裁剪区域应跟随 transform 偏移 |
| 状态 | ⬜ 待实现 |

#### R-021：BoxStyle 完整渲染（阴影+渐变+描边）⬜

| 项目 | 内容 |
|------|------|
| 测试函数 | `TestDrawBoxComplete` |
| 输出文件 | `test_output/render_core/box_complete.png` |
| 渲染内容 | 使用 `DrawBox` 渲染 **2 个完整 BoxStyle**： |
| | 1. 模拟 Ant Design Card：白底 + 圆角(8) + 浅灰描边(1) + 中阴影 |
| | 2. 模拟 Ant Design Primary Button：蓝→深蓝渐变 + 圆角(6) + 无描边 |
| 预期效果 | Card 有明显的浮起感（阴影柔和），按钮渐变平滑。各层叠加无视觉瑕疵 |
| 验证方式 | `validate_snapshot -min-colors 16 -min-non-bg 5000` |
| 状态 | ⬜ 待实现 |

#### R-022：Checkmark 图标绘制 ⬜

| 项目 | 内容 |
|------|------|
| 测试函数 | `TestDrawCheckmark` |
| 输出文件 | `test_output/render_core/checkmark.png` |
| 渲染内容 | 在白色矩形内绘制蓝色勾选图标（模拟 Checkbox 选中状态） |
| 预期效果 | 勾选线条清晰，两段线在拐角处连接自然。无线条断裂 |
| 验证方式 | `validate_snapshot -min-colors 3 -min-non-bg 200` |
| 状态 | ⬜ 待实现 |

#### R-023：路径填充（SVG Path）✅

| 项目 | 内容 |
|------|------|
| 测试函数 | `TestWriteCoreDrawingSnapshots` (已有) |
| 输出文件 | `test_output/render_core/svg_path.png` |
| 渲染内容 | 白色网格背景上绘制：1. SVG 心形路径（贝塞尔曲线）2. 箭头路径（直线段） |
| 预期效果 | 心形曲线平滑，箭头形状正确。路径填充区域无多余像素 |
| 验证方式 | `validate_snapshot -min-colors 3 -min-non-bg 2000` |
| 状态 | ✅ 已有测试 |

#### R-024：GPU 端到端渲染 ⬜

| 项目 | 内容 |
|------|------|
| 测试函数 | `scripts/gtk3_gpu_snapshot.sh` (已有框架) |
| 输出文件 | `test_output/render_core/gtk3_gpu_snapshot.png` |
| 渲染内容 | Demo 应用完整渲染一帧：面板 + 标题 + 按钮 |
| 预期效果 | 所有控件可见，文本清晰，按钮圆角平滑，阴影柔和。与 CPU 参考图视觉一致 |
| 验证方式 | `validate_snapshot -width 800 -height 600 -min-non-bg 1000 -min-colors 8` |
| 状态 | ⬜ 待实现 |

### 测试矩阵：修复验证对照表

> 每个 Bug 修复后，必须运行对应的测试用例验证。

| 问题 ID | 修复内容 | 验证测试 | 修复前预期 | 修复后预期 |
|---------|----------|----------|------------|------------|
| B-001 | Mat4.Multiply | R-007, R-018 | 变换位置/大小错误 | ✅ 已修复，6 个单元测试通过 |
| B-002 | 文本基线 | R-012, R-013 | 字形垂直错位 | ✅ 已修复，BearingY+ascent 偏移 |
| B-003 | StrokeRect 暗角 | R-001 | 半透明边框四角偏暗 | ✅ 已修复，四段不重叠 |
| B-004 | 渐变坐标 | R-006, R-007 | 渐变方向跟随屏幕 | ✅ 已修复，改为 UV 空间 |
| B-008 | takeLine 断行 | R-016 | 首空格处断行错误 | ✅ 已修复，>=0 判断 |
| B-009 | 内圆角半径 | R-004 | 内外圆角不同心 | ✅ 已修复，borderWidth/2 |
| B-010 | SDF 半径退化 | R-002 | 大圆角变形 | ✅ 已修复，shader 中 clamp |

### 测试运行指南

#### 运行所有 CPU 渲染测试

```bash
cd /home/yanghy/app/workspace/examples/lcl/gpui
go test ./render/pipeline/ -run TestWrite -v
go test ./render/pipeline/ -run TestRound -v
go test ./render/pipeline/ -run TestFill -v
# ... 对应每个测试函数
```

#### 运行 GPU 端到端测试

```bash
cd /home/yanghy/app/workspace/examples/lcl/gpui
bash scripts/gtk3_gpu_snapshot.sh
```

#### 自定义输出目录

```bash
GPUI_TEST_OUTPUT=/tmp/gpui_tests go test ./render/pipeline/ -run TestWrite -v
```

#### 验证单个快照

```bash
go run ./cmd/validate_snapshot \
  -file test_output/render_core/round_rect_fill.png \
  -min-colors 3 \
  -min-non-bg 10000
```

### 测试覆盖进度

| 类别 | 总数 | 已有 | 待实现 | 覆盖率 |
|------|------|------|--------|--------|
| 基础形状（矩形/圆形/线段/弧） | 5 | 5 | 0 | 100% |
| 圆角矩形（填充/描边/组合） | 3 | 3 | 0 | 100% |
| 阴影 | 1 | 1 | 0 | 100% |
| 渐变 | 2 | 2 | 0 | 100% |
| 文本渲染 | 6 | 3 | 3 | 50% |
| Transform | 2 | 2 | 0 | 100% |
| Clip 裁剪 | 2 | 0 | 2 | 0% |
| 组合渲染（BoxStyle） | 1 | 0 | 1 | 0% |
| 图标（Checkmark/Path） | 2 | 2 | 0 | 100% |
| GPU 端到端 | 1 | 0 | 1 | 0% |
| **总计** | **25** | **18** | **7** | **72%** |


---

> 迭代日志已迁移至 [CHANGELOG.md](CHANGELOG.md)

---

## 附录：问题 ID 索引
| ID | 简述 | 严重度 | 阶段 |
|----|------|--------|------|
| B-001 | Mat4.Multiply 顺序反 | 致命 | Phase 1 |
| B-002 | 文本基线偏移缺失 | 致命 | Phase 1 |
| B-003 | StrokeRect 暗角 | 致命 | Phase 1 |
| B-004 | 渐变坐标空间混乱 | 致命 | Phase 1 |
| B-005 | Container.Remove 指针泄漏 | 严重 | Phase 2 |
| B-006 | 双击三次激活 | 严重 | Phase 2 |
| B-007 | Box 吞噬禁用事件 | 严重 | Phase 2 |
| B-008 | takeLine index=0 断行 | 逻辑 | Phase 1 |
| B-009 | 内圆角半径计算错误 | 逻辑 | Phase 1 |
| B-010 | SDF 半径退化 | 逻辑 | Phase 1 |
| B-011 | SetFocusable 不清除 focused | 逻辑 | Phase 2 |
| B-012 | LayoutContainer 重复计算 | 逻辑 | Phase 2 |
| A-001 | 三套颜色系统冲突 | 架构 | Phase 1 |
| A-002 | 两套动画系统未接入 | 架构 | Phase 4 |
| A-003 | 焦点跨容器断裂 | 架构 | Phase 2 |
| A-004 | 无生命周期钩子 | 架构 | Phase 2 |
| A-005 | SetOwner 手动陷阱 | 架构 | Phase 2 |
| A-006 | 全局 currentApp | 架构 | Phase 3 |
| A-007 | 无 DPI 缩放 | 架构 | Phase 3 |
| F-001 | 缺少 HSL 操作 | 功能 | Phase 1 |
| F-002 | 无 10 级色板 | 功能 | Phase 1 |
| F-003 | Token 派生不一致 | 功能 | Phase 1 |
| F-004 | 组件 Token 不全 | 功能 | Phase 4 |
| F-005 | Mat4 缺少 Inverse | 功能 | Phase 3 |
| F-006 | 无贝塞尔曲线 | 功能 | Phase 3 |
| F-007 | 无圆角裁剪 | 功能 | Phase 3 |
| F-008 | 无 FBO | 功能 | Phase 3 |
| F-009 | 无纹理管理 | 功能 | Phase 3 |
| F-010 | 布局缺 flex-shrink | 功能 | Phase 3 |
| F-011 | flex-grow 无 min/max 钳位 | 功能 | Phase 3 |
| F-012 | Grid 不支持跨格 | 功能 | Phase 3 |
| F-013 | FocusTrap 未实现 | 功能 | Phase 4 |
| F-014 | 无焦点环渲染 | 功能 | Phase 4 |
| F-015 | 无键盘快捷键 | 功能 | Phase 4 |
| F-016 | 缺少基础控件 | 功能 | Phase 4 |
| P-001 | UniformSet.key() 分配 | 性能 | Phase 3 |
| P-002 | 阴影绘制调用过多 | 性能 | Phase 3 |
| P-003 | 文本换行 rune 分配 | 性能 | Phase 3 |
| P-004 | VBO 全量重传 | 性能 | Phase 3 |
