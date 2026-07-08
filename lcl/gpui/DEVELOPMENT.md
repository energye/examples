# GPUI 迭代开发文档

> 本文档用于持续跟踪 GPUI 框架的开发迭代，记录每次更新的目标、完成情况、不足之处及下一步计划。

---

## 文档使用说明

- 每次完成一个阶段或重要修复后，在对应条目更新状态
- 状态标记：⬜ 未开始 → 🔄 进行中 → ✅ 已完成 → ⚠️ 已完成但有遗留
- 每次更新后在「迭代日志」追加记录
- 当前阶段完成后，在「阶段总结」记录目标达成和不足

---

## 当前状态总览

| 阶段 | 名称 | 状态 | 开始日期 | 完成日期 |
|------|------|------|----------|----------|
| Phase 0 | 审计与规划 | ✅ 已完成 | 2026-07-08 | 2026-07-08 |
| Phase 1 | 致命 Bug 修复 + 颜色系统统一 | ⬜ 未开始 | - | - |
| Phase 2 | 架构补全 | ⬜ 未开始 | - | - |
| Phase 3 | 功能补全 | ⬜ 未开始 | - | - |
| Phase 4 | 控件库扩展 | ⬜ 未开始 | - | - |

**当前进度**：Phase 0 完成，准备进入 Phase 1

---

## Phase 0：审计与规划 ✅

**目标**：全面审查代码库，识别所有 Bug、架构缺陷和功能缺失

**完成日期**：2026-07-08

### 审计结果汇总

#### 致命 Bug（渲染错误）

| ID | 问题 | 文件 | 状态 |
|----|------|------|------|
| B-001 | `Mat4.Multiply` 列主序乘法顺序反了，所有嵌套变换错误 | `core/math/math.go:235-245` | ⬜ |
| B-002 | 文本渲染忽略基线偏移（ascent/bearingY），字形垂直对齐错乱 | `render/pipeline/primitives.go:26-44` | ⬜ |
| B-003 | `StrokeRect` 四角重叠绘制，半透明边框出现暗角 | `render/pipeline/text.go:53-63` | ⬜ |
| B-004 | 渐变着色器坐标空间混乱，transform 激活时渐变方向错误 | `render/shader/shader.go:398-437` | ⬜ |

#### 严重 Bug（行为错误）

| ID | 问题 | 文件 | 状态 |
|----|------|------|------|
| B-005 | `Container.Remove` 不清理 pointerCapture/hoverChild，被移除控件仍接收事件 | `widget/container.go:58-72` | ⬜ |
| B-006 | Engine 对双击同时派发 MouseDown+DoubleClick+MouseUp，三次激活 | `ui/engine.go:252-268` | ⬜ |
| B-007 | `Box.HandleEvent` 禁用状态下仍吞噬鼠标事件 | `widget/primitives.go:79-95` | ⬜ |

#### 逻辑错误

| ID | 问题 | 文件 | 状态 |
|----|------|------|------|
| B-008 | `takeLine` 在 index=0 处断行错误（`lastSpace > 0` 应为 `>= 0`） | `render/pipeline/text.go:121` | ⬜ |
| B-009 | `FillRoundRectWithBorder` 内圆角半径计算错误（`radius - borderWidth` 应为 `radius - borderWidth/2`） | `render/pipeline/text.go:115-116` | ⬜ |
| B-010 | 圆角 SDF 在 `radius > min(halfW, halfH)` 时退化 | `render/shader/shader.go:330-343` | ⬜ |
| B-011 | `FocusManager.SetFocusable(false)` 不清除已设置的 `focused` 标志 | `widget/base.go:130-149` | ⬜ |
| B-012 | `LayoutContainer` Measure 阶段重复计算 layout | `widget/layout.go:162-175` | ⬜ |

#### 架构缺陷

| ID | 问题 | 影响范围 | 状态 |
|----|------|----------|------|
| A-001 | 三套冲突的颜色/主题系统（color.go v4 / token.go v5 / theme.go 独立） | 整个样式层 | ⬜ |
| A-002 | 两套重复的动画系统（motion/ 和 style/animation/），均未接入控件 | 整个动画层 | ⬜ |
| A-003 | 焦点系统无法跨容器工作，registerFocusable 硬编码 Container/LayoutContainer 类型 | 事件系统 | ⬜ |
| A-004 | Widget 接口缺少生命周期钩子（OnMount/OnUnmount/OnResize/OnStateChanged） | 控件框架 | ⬜ |
| A-005 | SetOwner 必须手动调用，忘记调用导致 HitTest/SetFocusable 行为错误 | 控件框架 | ⬜ |
| A-006 | Application 使用全局 currentApp 变量，不支持多实例 | UI 层 | ⬜ |
| A-007 | 无 DPI 缩放支持，Engine.Context() 硬编码 Scale=1 | 渲染层 | ⬜ |

#### 功能缺失

| ID | 缺失功能 | 所属模块 | 状态 |
|----|----------|----------|------|
| F-001 | 颜色缺少 HSL 空间操作（FromHSL/ToHSL/Saturate/Desaturate） | core/math | ⬜ |
| F-002 | 无 10 级色板生成（Ant Design color-1 到 color-10） | style/token | ⬜ |
| F-003 | Token 派生与 Ant Design 不一致（圆角/间距/字号/暗色模式） | style/token | ⬜ |
| F-004 | 组件 Token 仅覆盖 4 个（Button/Input/Card/Modal），Ant Design 有 60+ | style/token | ⬜ |
| F-005 | Mat4 缺少 Inverse/Transpose/Shear | core/math | ⬜ |
| F-006 | 路径系统无贝塞尔曲线，无法渲染 SVG 图标 | render/pipeline | ⬜ |
| F-007 | 无圆角裁剪（PushClip 仅支持矩形） | render/pipeline | ⬜ |
| F-008 | 无 FBO 离屏渲染（模糊/阴影/backdrop-filter 需要） | render/pipeline | ⬜ |
| F-009 | 无纹理管理 API（image.Image → GPU 纹理） | render/texture | ⬜ |
| F-010 | 布局缺少 flex-shrink / flex-basis / align-self | layout | ⬜ |
| F-011 | flex-grow 分配后不执行 min/max 钳位 | layout/flex | ⬜ |
| F-012 | Grid 不支持 fr 单元和 grid-column/row 跨格 | layout/grid | ⬜ |
| F-013 | Overlay FocusTrap 字段存在但未实现任何逻辑 | overlay | ⬜ |
| F-014 | 无焦点环渲染 | widget | ⬜ |
| F-015 | 无键盘快捷键（Escape/方向键等） | widget/event | ⬜ |
| F-016 | 缺少基础控件（Input/Select/Checkbox/Radio/Switch/Slider/Tabs/Menu/Modal） | widget | ⬜ |

#### 性能问题

| ID | 问题 | 文件 | 状态 |
|----|------|------|------|
| P-001 | `UniformSet.key()` 每次绘制分配字符串+排序 | `render/pipeline/pipeline.go:66-82` | ⬜ |
| P-002 | 阴影用 3-16 个圆角矩形叠加，应改为专用 shadow shader | `render/pipeline/primitives.go:194-222` | ⬜ |
| P-003 | 文本换行每行分配 `[]rune`，应改用 range 迭代 | `render/pipeline/text.go:109,140` | ⬜ |
| P-004 | VBO 每个 batch 从 offset 0 重传，应使用 ring buffer | `render/pipeline/pipeline.go:353-362` | ⬜ |

---

## Phase 1：致命 Bug 修复 + 颜色系统统一 ⬜

**目标**：修复所有导致渲染错误的 Bug，统一颜色系统为单一 Token 体系

**预计工作项**：14 项

### 1.1 数学与渲染 Bug 修复 ⬜

| 任务 | 关联问题 | 负责模块 | 状态 |
|------|----------|----------|------|
| 修复 Mat4.Multiply 列主序乘法 | B-001 | `core/math/math.go` | ⬜ |
| 修复文本基线偏移，添加 ascent/bearingY 计算 | B-002 | `render/pipeline/primitives.go` | ⬜ |
| 修复 StrokeRect 四角重叠，改为不重叠的四段 | B-003 | `render/pipeline/text.go` | ⬜ |
| 修复渐变着色器坐标空间，统一使用屏幕空间 | B-004 | `render/shader/shader.go` | ⬜ |
| 修复 takeLine index=0 断行 | B-008 | `render/pipeline/text.go` | ⬜ |
| 修复 FillRoundRectWithBorder 内圆角半径 | B-009 | `render/pipeline/text.go` | ⬜ |
| 修复圆角 SDF 半径退化（添加 clamp） | B-010 | `render/shader/shader.go` | ⬜ |

### 1.2 颜色系统统一 ⬜

| 任务 | 关联问题 | 负责模块 | 状态 |
|------|----------|----------|------|
| Color 添加 HSL 空间操作（FromHSL/ToHSL/Saturate/Desaturate/HueRotate） | F-001 | `core/math/math.go` | ⬜ |
| 实现 10 级色板生成算法（基于 HSL 亮度阶梯） | F-002 | 新建 `style/token/palette.go` | ⬜ |
| Token 派生修正（圆角/间距/字号对齐 Ant Design v5） | F-003 | `style/token/token.go` | ⬜ |
| 废弃 `style/color/color.go`，统一使用 Token 系统 | A-001 | `style/color/` | ⬜ |
| 废弃 `style/theme/theme.go`，统一使用 Token 系统 | A-001 | `style/theme/` | ⬜ |
| 暗色模式色板重新生成（不只是背景/文本/边框） | F-003 | `style/token/token.go` | ⬜ |

### Phase 1 完成标准

- [ ] Mat4 变换嵌套顺序正确（可通过嵌套 Container 验证）
- [ ] 文本垂直对齐正确（不同字形基线一致）
- [ ] 半透明边框无暗角
- [ ] 渐变在有 transform 时方向正确
- [ ] 颜色系统只有一个入口（token.Current()）
- [ ] Lighten/Darken 在 HSL 空间操作
- [ ] 10 级色板可正确生成
- [ ] 暗色模式下组件颜色正确调整

---

## Phase 2：架构补全 ⬜

**目标**：修复控件层 Bug，补全架构缺陷，使框架具备生产级基础

**预计工作项**：12 项

### 2.1 控件层 Bug 修复 ⬜

| 任务 | 关联问题 | 负责模块 | 状态 |
|------|----------|----------|------|
| Container.Remove 清理 pointerCapture/hoverChild | B-005 | `widget/container.go` | ⬜ |
| LayoutContainer.Remove 同步修复 | B-005 | `widget/layout.go` | ⬜ |
| Engine 双击事件改为只派发 DoubleClick（不同时派发 MouseDown） | B-006 | `ui/engine.go` | ⬜ |
| Box.HandleEvent 禁用时不再吞噬事件 | B-007 | `widget/primitives.go` | ⬜ |
| FocusManager.SetFocusable(false) 清除 focused 标志 | B-011 | `widget/base.go` | ⬜ |
| LayoutContainer.Measure 缓存结果避免重复计算 | B-012 | `widget/layout.go` | ⬜ |

### 2.2 焦点系统重构 ⬜

| 任务 | 关联问题 | 负责模块 | 状态 |
|------|----------|----------|------|
| 焦点管理改为全局协调（Engine 层单一 FocusManager） | A-003 | `widget/focus.go` + `ui/engine.go` | ⬜ |
| registerFocusable 改用 Children() 接口而非类型硬编码 | A-003 | `widget/container.go` | ⬜ |
| Portal 关闭时自动恢复之前的焦点 | A-003 | `widget/portal.go` | ⬜ |

### 2.3 生命周期钩子 ⬜

| 任务 | 关联问题 | 负责模块 | 状态 |
|------|----------|----------|------|
| Widget 接口添加 OnMount/OnUnmount 可选接口 | A-004 | `widget/types.go` | ⬜ |
| Container.Add 触发 OnMount，Remove 触发 OnUnmount | A-004 | `widget/container.go` | ⬜ |
| Layout 时 bounds 变化触发 OnResize 回调 | A-004 | `widget/base.go` | ⬜ |

### Phase 2 完成标准

- [ ] 移除子控件后不再有悬挂引用或事件泄漏
- [ ] 双击只触发一次回调
- [ ] 禁用控件不阻止事件传递
- [ ] Tab 键可以跨 Container 导航
- [ ] Portal 关闭后焦点正确恢复
- [ ] 第三方容器类型的子控件可正确参与焦点系统
- [ ] 控件有 Mount/Unmount/Resize 回调

---

## Phase 3：功能补全 ⬜

**目标**：补全渲染管线和布局引擎的缺失能力，使框架能支撑完整 UI

**预计工作项**：10 项

### 3.1 渲染管线增强 ⬜

| 任务 | 关联问题 | 负责模块 | 状态 |
|------|----------|----------|------|
| 贝塞尔曲线路径（Quadratic/Cubic Bezier） | F-006 | `render/pipeline/path.go` | ⬜ |
| 圆角裁剪（PushClip 支持 rounded rect） | F-007 | `render/pipeline/pipeline.go` | ⬜ |
| FBO 离屏渲染支持 | F-008 | `render/pipeline/` + `core/gl/` | ⬜ |
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
| Checkbox（复选框，支持半选） | F-016 | 新建 `widget/checkbox.go` | ⬜ |
| Radio（单选框） | F-016 | 新建 `widget/radio.go` | ⬜ |
| Switch（开关） | F-016 | 新建 `widget/switch.go` | ⬜ |
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

---

## 迭代日志

> 每次更新后在此追加记录，格式：日期 | 阶段 | 完成内容 | 当前不足 | 下一步

### 2026-07-08 | Phase 0 | 审计完成

**完成内容**：
- 对渲染管线、控件层、事件系统、Token 系统、布局引擎、Overlay 系统、动画系统进行全面代码审查
- 识别 4 个致命 Bug、3 个严重 Bug、5 个逻辑错误、7 个架构缺陷、16 个功能缺失、4 个性能问题
- 建立迭代开发文档框架

**当前不足**：
- 所有识别的问题均未修复
- 三套颜色系统并存，互相冲突
- 两套动画系统并存，均未接入控件
- 焦点系统无法跨容器工作
- 缺少生命周期钩子

**下一步**：进入 Phase 1，修复致命 Bug，统一颜色系统

**不足所在环节**：
- 渲染层：core/math（Mat4 乘法）、render/pipeline（文本基线/边框绘制）、render/shader（渐变坐标/SDF 退化）
- 样式层：style/color 与 style/token 两套系统冲突，style/theme 第三套系统独立
- 控件层：widget/container（指针捕获清理）、widget/primitives（事件吞噬）
- UI 层：engine（双击事件逻辑）

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
