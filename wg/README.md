# wg 示例集合

基于「单画布纯自绘」架构的 `wg` 组件示例集合。

> 设计目标（原文）：「只有一个画布，然后在上面画各种渲染基础能力和组件的功能实现。还有输入具有焦点的也都使用它实现。不使用有句柄的也就是纯自绘的任意组件。」

窗口（Form）是唯一持有 HWND 的对象；挂在窗口下的 `ICustomGraphicControl` 是无句柄画布；所有组件都是纯 Go 对象，画在同一张画布上——包括带焦点、能输入中文的输入框。

| 示例 | 内容 | 运行 |
|---|---|---|
| [canvas_ime](./canvas_ime) | 单画布渲染（圆角/渐变/阴影/主题/进度条/状态面板）+ 输入法 IME 实测（预编辑串、候选框定位、UTF-8 上屏）+ 纯自绘输入框的完整键盘编辑（选区/剪贴板/撤销/按词移动）与控制字符过滤 | `go run ./wg/canvas_ime` |
| [showcase](./showcase) | 1440×880 设计画布（随窗口等比自适应）：动态波形/折线/柱状/仪表/饼图、真·逐像素半透明玻璃卡片、异形花瓣与星形与贝塞尔、抗锯齿圆角；软件覆盖率抗锯齿 + 渐变/发光描边；静态层烘焙 + 失效驱动 60Hz 帧泵 + 脏区局部提交 + resize 去抖，稳定 60 FPS | `go run ./wg/showcase` |

在 `E:/app/workspace/examples` 下执行，`go.work` 已 `use ./examples`，无需独立 go.mod。自检与回归：

```bash
go run ./wg/canvas_ime --selftest     # 依赖与 DLL 导出
go run ./wg/canvas_ime --keyselftest  # 自动注入真实按键并断言编辑器状态
go run ./wg/showcase --selftest       # RGBA 批量写 / 通道顺序 / 圆角覆盖率 / 4 项渲染回归
go run ./wg/showcase --bench          # 真实窗口跑 3 秒打印 FPS 与帧耗时
go run ./wg/showcase --mousetest      # 注入 WM_MOUSEMOVE 断言动画持续且 FPS 不被抬升
go run ./wg/showcase --schedtest      # 断言 60Hz 锁帧 / 静止零帧 / 脏区单帧
go run ./wg/showcase --resizetest     # 真实 WM_SIZE 连续切换 7 种尺寸，断言不崩不卡
go run ./wg/showcase --aaprobe        # 无窗口探测 LCL AA 能力与像素读写成本
go run ./wg/showcase --screenshot aa_frame.png  # 从帧缓冲导出精确合成帧（含 GDI 文字）
go test ./wg/...                      # 两个示例的纯逻辑单测
```

> **渲染管线约束**（来自 showcase 的实测，做组件库时必须沿用）：① 画刷**先 `SetColor` 再 `SetStyle`**（LCL 的 `SetColor` 会重置 Style）；② 离屏帧缓冲用**不带 alpha 通道的 `Pf24bit`**，GDI 只写 RGB，用 `Pf32bit` 会让全部矢量图元在最终 Blit 时被当成全透明；③ 单画布整帧渲染必须**抑制 `WM_ERASEBKGND`**（首帧之后才接管，其余消息一律透传），否则渲染期里窗口一直是空底。详见 `showcase/README.md` 第六节。
>
> **帧调度约束**（同样必须沿用）：④ 静态内容**只烘焙一次**成不透明快照，每帧整窗 BitBlt 恢复 + 只重画动态组件；逐帧重做 alpha 混合会吃掉 79% 的帧预算。⑤ 帧由**失效驱动 + 墙钟锁 60Hz**、被动投递到主线程（`WM_APP+1`），输入事件只标脏区不请求帧，动画停止即零帧；不要用 `TTimer` 当帧时钟（Windows 量化会让帧率在 39/60/78 之间跳）。每帧分 **PREPARE**（离屏批量光栅化）与 **PRESENT**（边界原子提交）两段。详见 `showcase/README.md` 第四·二节。
>
> **抗锯齿 / 文本约束**：⑥ 当前 GDI widgetset 的 `AntialiasingMode` 对图元**无效**，且额外叠一层 RGBA 做 alpha `Draw` 要 24.7ms/帧（击穿预算），所以 AA 必须**按覆盖率就地混合进不透明帧缓冲**：`CreateIntfImage().PixelData()` 批量读出静态快照 → Go 侧光栅化 → `LoadFromIntfImage` 批量写回 → 最后一次不透明提交。文字**不能**进软件光栅器（会丢 ClearType），须延迟到写回之后用 GDI `TextOut` 回放；`TextOut` 本身无裁剪，标签必须按 rect 宽度裁剪。折线/曲线是**一个图元**：密集顶点上的圆头端点若逐段 over 合成，会让边缘覆盖率非单调、出现周期性亮珠（肉眼即"锯齿"），必须先累积覆盖率取 MAX、最后逐像素混合一次。⑦ 不透明 `CopyRect` 整窗约 0.48ms，alpha `Draw` 约 24.7ms（50 倍差）；`Pf24bit` 帧缓冲的 `ScanLine` raw image 与 GDI HDC 不同步，不要拿它当帧缓冲。详见 `showcase/README.md` 第七节。
>
> **生命周期与布局约束**：⑧ 静态层**二次烘焙必须切回 GDI 后端**（临时 `painter.frame = nil`）——重烘焙可能发生在 OnPaint 内，此时 AA 表面要么属于上一帧、要么已被 resize 释放，写进去就是段错误，且延迟文字队列会被 `Reset()` 清空（表现为「拖滑条后某张卡片内容消失」）。释放顺序必须是 **先 `freeAASurface()` 再 `Free()` 旧位图**（`frameImg` 由位图 `CreateIntfImage` 得来），并给像素视图加空缓冲守卫。⑨ 布局用**设计画布 → 客户区等比缩放 + 居中**，坐标不能硬编码在窗口像素上；字号/行距/inset/子控件偏移要同步缩放。⑩ `WM_SIZE`/`WM_SIZING` 要**去抖**：一次 RGBA 静态烘焙约 85–100ms，拖动窗口时每个中间尺寸都烘焙会饿死消息循环；冷却期保留旧快照走轻量路径，稳定后烘焙一次，且 AA 判定必须校验快照尺寸与帧缓冲一致；`unsafe.Slice` 读回 DIB 的长度只能由**该位图自身尺寸**推导，不能用当前客户区尺寸（冷却期两者短暂不一致会偶发 SIGSEGV）。详见 `showcase/README.md` 第八节。

后续可继续在此集合下补充组件示例（如 `tab`、`list`、`slider`），并复用 `canvas_ime` 里的 `Scene` / `Painter` 骨架；成熟后可上移到 `energy/lcl/wg`。
