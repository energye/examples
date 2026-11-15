# 单画布 · 复杂动态图形 / 半透明 / 异形 / 圆角 渲染示例

> 位置：`examples/wg/showcase/`，属于 `examples/wg` 示例集合；由 `go.work` 管理，无独立 go.mod。
> 基于 `../canvas_ime` 的单画布架构（Form 是唯一 HWND，`ICustomGraphicControl` 是无句柄画布，所有组件是纯 Go 对象）。

一个 **1440×880 的大窗口**，把前几轮讨论的渲染能力全部做成可运行、可实测的程序：动态图形、真·逐像素半透明、非矩形异形、抗锯齿圆角，以及一整屏密集内容。

## 一、窗口里有什么

```
┌──────────────────────────── 标题 / FPS 徽章 ────────────────────────────┐
│ 动态波形(180点)      │ 环形仪表 + 旋转星形   │ 渲染性能实测面板          │
│ 不规则填充区域       │ 多段圆弧 / 扇形        │ FPS·帧耗时·FFI·缓存命中   │
├─────────────────────┼───────────────────────┼───────────────────────────┤
│ 实时折线(3×120点)    │ 动态柱状图(12根圆角柱) │ 自绘控件：按钮/滑块/开关/进度条 │
├─────────────────────┼───────────────────────┼───────────────────────────┤
│ 极坐标花瓣 + 贝塞尔  │ 环形饼图(4段)         │ 架构与操作说明            │
└─────────────────────┴───────────────────────┴───────────────────────────┘
        背景还有 2 个缓慢漂移的不规则半透明光斑
```

全部 9 张卡片是**抗锯齿圆角 + 逐像素 alpha 的玻璃层**，卡片标题、分隔线、图例、控件都是同一张画布上的纯 Go 对象，**零 LCL 实例、零句柄**。

## 二、四类能力怎么实现的

| 能力 | 实现 |
|---|---|
| **半透明** | `AlphaLayer`：`ILazIntfImage(RiqfRGB\|RiqfAlpha)` → **`PixelData()` 一次取指针，Go 侧整块写 BGRA** → `BeginUpdate/EndUpdate` → `LoadFromIntfImage` → `IBitmap(Pf32bit)` → `Canvas.Draw`。不是逐像素 `SetColors`（那是 O(w×h) 次 FFI），而是 O(1) 次。`bakeGlass` 给每像素算圆角覆盖率 × 全局 opacity；`bakeGlow` 画带极坐标瓣状边界的不规则光斑 |
| **抗锯齿** | 几何图元在 Go 侧按**覆盖率就地混合进不透明 `Pf24bit` 帧缓冲**（`raster.go`），不额外叠 RGBA 图层；文字先入队，等像素批量写回后才用 GDI `TextOut` 回放，保留 ClearType。详见第七节 |
| **异形图形** | `Polygon` 极坐标花瓣(96点)与旋转星形、`Polyline` 波形与 3 条实时曲线、`PolyBezier` 连续贝塞尔丝带、`Pie` 饼图扇区、`Arc` 仪表圆弧、曲线下方 180 点不规则填充；波形填充带垂直双色渐变，曲线/丝带带单遍发光描边 |
| **圆角图形** | 两层：RGBA 烘焙的**逐像素抗锯齿圆角卡片**（`roundedAlpha` 平滑覆盖率），以及软件光栅化的 `RoundRect` 矢量圆角柱/按钮/滑块/开关 |
| **动态图形** | `FrameScheduler` 在 **绝对 60Hz 帧边界**投递帧，`PhaseNow()` 用墙钟算相位，所有组件在 `Paint` 时读取相位计算波形、柱高、仪表进度、星形旋转、花瓣形变。组件自己不开定时器 |

`LayerCache` 按 `(theme, size, radius, opacity)` 共享烘焙结果：9 张卡片 + 2 个光斑只生成 **11 层**，且只在静态层烘焙时被访问一次；稳态每帧完全不触碰缓存。改主题/拖滑块会 `InvalidateAll()` + 重建静态快照。通道顺序用探针像素**运行时探测**（本机 BGRA，stride=w×4），不硬编码平台格式。

## 三、运行

```bash
cd E:/app/workspace/examples

go run ./wg/showcase --selftest    # 不弹窗：通道/批量写/圆角 + 渲染回归 + 帧泵单 outstanding 断言
go run ./wg/showcase --bench       # 真实窗口，帧泵启动后跑 3 秒，打印性能与分段剖析
go run ./wg/showcase --mousetest   # 注入 1300+ WM_MOUSEMOVE，断言动画持续且 FPS 不被抬升
go run ./wg/showcase --schedtest   # 断言 60Hz 锁帧 / 静止零帧 / 脏区单帧
go run ./wg/showcase --resizetest  # 真实 WM_SIZE 连续切换 7 种尺寸，断言不崩不卡
go run ./wg/showcase --inputtest   # 最大化/还原/真实拖拽 + 输入延迟，22 项断言（回归本节第九节）
go run ./wg/showcase --inputdiag   # 只记录原生消息/capture 状态到 inputdiag.log，不跑脚本
go run ./wg/showcase --aaprobe     # 无窗口能力探测：LCL AA 是否有效、ScanLine/IntfImage 读写成本
go run ./wg/showcase --screenshot aa_frame.png   # 从帧缓冲导出精确合成帧 + 3x 局部放大图
go run ./wg/showcase               # 打开窗口
go test ./wg/...                   # 两个示例的纯逻辑单测

go build -o wg/showcase/showcase.exe ./wg/showcase   # examples/.gitignore 已忽略 *.exe
```

| 环境变量 | 作用 |
|---|---|
| `WG_AA=0` | **抗锯齿 A/B 开关**：回到全 GDI 硬边渲染（几何无 AA、无发光、无渐变、无文字裁剪），约 7.0ms/帧 |
| `WG_STATIC=0` | 关闭静态层烘焙（A/B 对照：恢复逐帧 alpha 重绘、光斑漂移、约 30 FPS 老行为；同时自动关闭 AA） |
| `WG_FAST=1` | 关闭离屏帧缓冲（增量绘制可能可见；同时自动关闭 AA） |
| `WG_ERASE=1` | 恢复系统擦背景，复现修复前的空底闪烁 |
| `WG_BLANK=1` | 跳过全部 widget，测量「背景恢复 + 原子提交」的固有成本 |
| `WG_PROFILE=1` | 打印逐组件耗时剖析（`--bench` 默认开启） |
| `WG_SIZE=1000x640` | 配合 `--screenshot`：先把真实窗口改成指定尺寸再导帧，验证自适应布局 |
| `WG_BAKE=1` | 配合 `--screenshot`：导出前模拟滑条改不透明度、触发一次 OnPaint 内的静态重烘焙 |

`--screenshot` 直接从离屏帧缓冲读取合成后的精确帧（含 GDI 文字），绕开 Win32 抓屏的 `BitBlt` 限制。仓库里保留了 `aa_frame.png` / `aa_frame_zoom.png`（AA 版）、`gdi_frame.png` / `gdi_frame_zoom.png`（`WG_AA=0` 版）作为 A/B 对照，`edge_before.png` / `edge_after.png`（曲线边缘 8 倍放大，修复前后），以及 `resize_small.png`（1000×640 自适应布局）、`rebake_frame.png`（滑条重烘焙后仍完整）。

运行时库优先级同 `canvas_ime`：`ENERGY_LIB` > exe 同目录 > `~/.energy/runtime/libenergy-amd64.dll` > `E:/app/workspace/lib/libenergy-x86_64-win64.dll`。

交互：点击三个圆角按钮（暂停 / 切主题 / 重置缓存）、**拖动滑块**改全局不透明度（会触发全部 RGBA 图层重烘焙）、点开关；键盘 `空格`/`P` 暂停、`T` 切主题、`+`/`-` 调不透明度。

## 四、从 30 FPS 到稳定 60 FPS（本机 1440×880 真实窗口实测）

| 阶段 | FPS | 整帧 | 关键变化 |
|---|---|---|---|
| ① 原实现（逐帧全量重绘） | 30.1 | 35.0 ms | 9 张卡片 + 2 光斑每帧做 alpha 混合 |
| ② 静态层烘焙 | 40.8 → 64 | 9.2 ms → 7.5 ms | 静态内容只光栅化一次，每帧 1 次不透明 BitBlt |
| ③ 失效驱动 60Hz 帧泵 | **59–60** | 7.5 ms | 帧由墙钟边界被动投递，鼠标不再触发绘制（AA 前基线） |
| ④ 软件抗锯齿 + 发光/渐变 | **60.0–60.4** | 9.3–10.7 ms | 几何按覆盖率就地混合、文字延迟回放；单遍合并发光把 PREPARE 拉回预算内 |

**逐帧剖析（优化前，1440×880 / 21 组件 / `--bench`）**：`clear 1.64ms | blit 1.97ms | widgets 31.34ms`，其中 9 张玻璃卡片 + 2 个光斑合计 **24.6ms（79%）**，而全部图表矢量绘制仅 2.2ms、文字 3.5ms。结论很直接：半透明卡片**从不变化**，却每帧重做一遍逐像素 alpha 混合。

**优化后（AA 前）**：`背景恢复 0.6ms | widgets 4.8ms | PRESENT 2.1ms | PREPARE+PRESENT ≈ 7.5ms`，画布调用 607 → 445 次。开启软件 AA 后画布调用进一步降到 **84 次/帧**，整帧 9.3–10.7ms；逐项数据见第七节。

`--selftest` 输出：

```
[SELFTEST] channel order R=2 G=1 B=0 A=3  stride=256
[SELFTEST] bulk-write readback {Red:16448 Green:32896 Blue:49344 Alpha:65535} -> PASS
[SELFTEST] bitmap valid: true
[SELFTEST] roundedAlpha corner/center: PASS
[SELFTEST] back buffer pixel format           PASS  pf=6 bitsPerPixel=24 (want Pf24bit/24)
[SELFTEST] Painter.brush keeps BsClear        PASS  style=1 (want BsClear=1)
[SELFTEST] Painter.brush keeps BsSolid+colour PASS  style=0 color=0x3c28e6
[SELFTEST] offscreen GDI fill colour          PASS  R=e6 G=28 B=3c (want e6/28/3c)
[SELFTEST] frame pump: single outstanding slot PASS  second post refused=true FrameDone cleared=true
RESULT: PASS
```

## 四·二、帧调度模型（对齐成熟 UI 框架）

用户提的四个问题，本质是同一个：**帧不应该由「谁想画」决定，而应该由显示时钟决定**。下表是与成熟框架的对照：

| 概念 | Flutter / Chromium / WPF / Android | 本示例 |
|---|---|---|
| 标记变化 | `scheduleFrame` / `InvalidateVisual` / `invalidate(Rect)` | `RequestFrame()` / `InvalidateRect()` |
| 帧时钟 | VSync / `BeginFrame` / `Choreographer` / `CompositionTarget.Rendering` | `FrameScheduler` 绝对 60Hz 墙钟（16.667ms） |
| 主线程消费 | 帧回调里 build→layout→paint | `WM_APP+1` → `RenderFrame()` → `OnPaint` |
| 整帧准备 | layer tree / display list / command buffer | 所有图元批量画进 `backBuf`（**PREPARE**） |
| 到点提交 | compositor / `Present` / `swapBuffers` | 一次整窗 Blit 或一次脏区 CopyRect（**PRESENT**） |
| 静止 | 无 VSync 请求，零绘制 | 动画停止即停泵，`--schedtest` 实测 1 秒 0 帧 |
| 脏区 | damage rect / `invalidate(Rect)` | `dirty` 合并矩形，静止时只恢复/提交该区域 |

**16ms 内算好批量，到点才取渲染**：每个帧槽里，先用墙钟 `PhaseNow()` 定同一时刻的动画相位，把所有动态几何、文字、图层**一次性批量**光栅化到离屏 `Pf24bit` 帧缓冲；这一步用户完全看不见。然后在帧边界做唯一一次原子提交。两种帧源正是「有动态的场景」（连续 60Hz）与「要刷新场景」（一次性脏区帧），其余时间不绘制。

**为什么不用 `TTimer`**：Windows `SetTimer` 被量化（本机实测 16ms→39 FPS、12ms→78 FPS、14ms 偶发 16 FPS），帧率会跟着量化档位跳。现在帧节拍由 Go 侧绝对时间轴控制，`next += 16.667ms` 不累积漂移；帧泵是**单 outstanding 闭环**（`CAS(0,1)` 拒绝二次投递、等 `FrameDone` 再投下一槽，慢帧平移时槽而非丢整槽），所以不会出现补帧爆发；「3 槽没回就清零 `inflight`」的短超时看门狗已删除——它在静态重烘焙期间会把 post 队列灌满、饿死输入（见第九节）。`timeBeginPeriod(1)` 用来抬高系统时钟精度。

**要不要接真 VSync**：当前**不需要**，也不该现在要。Windows 8+ 起 DWM 强制合成，GDI 的 BitBlt 只写进窗口 surface 由 DWM 在合成周期取用，**不会撕裂**；真 VSync 唯一能改善的是软件墙钟与合成点的相位漂移，而实测抖动还包含 goroutine 调度与消息排队。`IDXGIOutput::WaitForVBlank` 会阻塞线程，放进主线程绘制路径还可能把输入重新饿死；`DwmFlush()` 的官方语义是「等排队的 DirectX 表面更新绘制到屏幕」，并不是等下一个垂直空白，不能拿来冒充。只有在上 GPU 后端 / 120Hz+ 高刷 / 多显示器不同刷新率 / 连续高速动画时才该重新评估，届时首选 DirectComposition 的 Compositor clock API，且等待必须放在调度 goroutine 而非主线程。判断依据与触发条件见 `wg-canvas-arch/ARCHITECTURE.md` §4·二.5。

**鼠标为什么不再抬升 FPS**：`dispatch()` 现在只更新 hover/拖动状态，动画运行期间**完全不请求帧**（帧泵已经在 60Hz 跑）；静止且状态真的变化时才合并脏区并请求恰好一帧；纯移动无状态变化则零帧。实测 `--mousetest` 注入 1330 条 `WM_MOUSEMOVE`：修复前 126 FPS，修复后 **60 FPS**、相位正常推进。

**`--schedtest` 实测**（真实窗口 + 真实消息循环）：

```
动画中 FPS: 60.0 (期望 ~60)
暂停 1s 内绘制帧数: 0 (期望 0)   [诊断 active=false 帧泵投递增量=0]
脏区请求产生帧数: 1 (期望 1)   其中脏区帧: 1 (期望 1)
恢复 1s 内绘制帧数: 59 (期望 ~60)   实测 58.9 FPS
RESULT: PASS — 60Hz 锁帧 / 静止零帧 / 脏区单帧 全部成立
```

> 取舍：静态层烘焙把 2 个光斑的缓慢漂移固定成背景装饰（`WG_STATIC=0` 可恢复逐帧漂移，代价是回到 ~30 FPS）。静态层重建只在主题 / 不透明度 / 尺寸 / 缓存重置时发生，并在帧泵启动前预烘，避免抢占启动帧槽。

## 五、已修复：鼠标在窗口内移动时动画停止

**现象**：鼠标持续移动时波形、仪表、星形全部冻结，停止移动后恢复。

**根因**：Windows 的 `WM_TIMER` 和 `WM_PAINT` 都是**低优先级消息** —— 只在消息队列为空时才生成。持续的 `WM_MOUSEMOVE` 会把两者饿死。而原来的动画相位是 `Tick()` 里 `Phase += dt`（dt 来自 Timer 两次回调的间隔），Timer 一停 `Phase` 就不再推进，即便偶尔有重绘也永远是同一帧，看起来就是"渲染停止"。此外 `dispatch()` 末尾对每次鼠标移动都无条件 `Invalidate()`，进一步堆积绘制请求。

**修复**（`scene.go`）：

| 改动 | 作用 |
|---|---|
| `PhaseNow()` 用墙钟时间计算相位，`OnPaint` 开头也刷新 | 相位不再依赖 Timer 触发次数，丢帧不会冻结或倒退；**这是关键修复** |
| 现行版本：鼠标事件**只改状态、不请求帧**；帧由 `FrameScheduler` 在 60Hz 边界被动投递 | 彻底消除「鼠标一动 FPS 就涨」；静止时纯移动零帧 |
| FPS / Frame 统计在 `OnPaint()` 里按真实绘制计数，静止期不参与平均 | 统计真实动画帧数，不把调度 tick 或空闲时间当帧 |
| `SetPaused()` 累计暂停时长 | 暂停/继续后墙钟相位仍然连续 |

**对照实验**（`--mousetest`，3 秒注入约 1330 条 `WM_MOUSEMOVE`）：

| 方案 | 相位推进 | 实际绘制 | FPS |
|---|---|---|---|
| 早期：墙钟 + 同步 `Repaint`（每次鼠标都触发） | 3.01 s | 357 帧 | **126.4**（被鼠标抬升） |
| 中期：14ms 闸门 + 8ms TTimer | 2.99 s | 196 帧 | 64.8（受 SetTimer 量化影响） |
| 现行：失效驱动 60Hz 帧泵 | 3.00 s | **180 帧** | **60.0** |

对照说明：墙钟相位解决了「动画停止」；帧泵解决了「FPS 被输入事件抬升」和「帧率随定时器量化跳变」。三者叠加输出 `RESULT: PASS — 鼠标移动期间动画持续运行`。

## 六、已修复：窗口一直闪 + 内容全是暗灰色、不彩色

**现象**：上一轮交接里遗留的两个问题。做完离屏双缓冲后用户实测仍然闪，且整屏只有暗灰、没有彩色内容。

**结论：不是配色问题，是三个互相独立的渲染 bug 叠加。** 上一轮只做了「整帧画进位图再 Blit」，但没检查位图格式、画刷状态和消息时序。

| # | 根因（均已实证） | 造成的现象 | 修复 |
|---|---|---|---|
| 1 | LCL 的 `TBrush.SetColor()` 会把 `Style` **重置回 `bsSolid`**。`Painter.brush()` 原来先 `SetStyle` 后 `SetColor`，所以 `BsClear` 从未真正生效 | 每次 `TextOut` 都填一整块不透明色底：标题变白色实心条、副标题变灰条、FPS 徽章变青色块 | `painter.go`：改为**先 `SetColor` 再 `SetStyle`**（顺序不能交换） |
| 2 | 离屏帧缓冲用了 `Pf32bit`。GDI 的 `FillRect` / `Polyline` / `TextOut` 只写 RGB，32 位 DIB 的第 4 字节恒为 0；LCL 把这种位图 `Draw` 到屏幕 DC 时按 `AC_SRC_ALPHA` 混合 | 所有矢量图元（曲线、柱、饼、文字）被当作全透明而消失，只剩自带 alpha 的玻璃卡片 —— 这正是「内容全是暗灰色、不彩色」 | `scene.go` 的 `newBackBuffer()` 改用 **`Pf24bit`**（无 alpha 通道），最后一次 Blit 是纯拷贝 |
| 3 | 每帧 `WM_ERASEBKGND` 先把窗口擦成窗体底色，而离屏整帧渲染要 30ms+；这 30ms 里屏幕上是空底，成品只在最后一次 Blit 后存在一瞬 | 持续闪烁；且大部分时间窗口里只有深色底 | `main.go` 的 `SetOnWndProc`：在**首帧成功绘制之后**抑制 `WM_ERASEBKGND`（`msg.Result = 1`），其余消息一律 `InheritedWndProc(msg)` 透传 |

### 实证过程（不靠猜）

| 实验 | 结果 |
|---|---|
| 无窗口位图回读：GDI 画进 `Pf32bit` 的像素 | `raw=[3c 28 e6 00]`、`TFPColor A=0000`；而 `LoadFromIntfImage` 来的 RGBA 层是 `[3c 28 e0 80]`、`A=8080` → 确认 GDI 不写 alpha |
| 画刷状态回读（文字区域统计白色像素数） | `SetStyle(BsClear)` → `SetColor` 后读回 `Style=0`、白像素 1380（整块色底）；`SetColor` → `SetStyle(BsClear)` 后 `Style=1`、白像素 79（只有字形） |
| OnPaint 内部 `GetPixel` 读屏幕 DC | 与离屏缓冲逐点完全一致（如 `0x004e372f`）→ **Blit 是成功的**，空白不来自 Blit，而来自之后的擦背景 |
| `Pf32bit` vs `Pf24bit` 对照截图（都已抑制擦背景） | `Pf32bit`：仪表/饼图/柱状图位置全是黑洞，无文字；`Pf24bit`：9 张卡片、全部彩色图表、中文标题完整 |
| 连续 40 帧屏幕采样（统计「只有窗体底色」的空底帧） | `WG_ERASE=1`（修复前行为）：**26/40 帧全空底**；修复后：**0/40** |

`--selftest` 新增 4 项**无窗口回归断言**锁住这些结论：帧缓冲必须是 `Pf24bit/24bpp`、`Painter.brush` 必须保持 `BsClear` 和 `BsSolid+颜色`、离屏 GDI 填充颜色必须回读正确。任何一条被改回去都会立刻 FAIL。改用帧泵后系统不再逐帧发送 `WM_ERASEBKGND`（抑制计数为 0），擦背景风险由「整帧离屏 + 原子提交」本身消除；`WG_ERASE=1` 仍可强制复现。

### 关于「别再吞 WM_ERASEBKGND」的修正

上一轮的结论是「吞掉 `WM_ERASEBKGND` 会让窗口初始化卡死，别再试」——**无条件吞确实会卡死**，但原因是从窗口创建的第一条消息起就不透传。现在用 `s.Frame > 0` 守卫：**初始化阶段的擦背景全部照常透传，只在首帧成功绘制之后才接管**，其余所有消息仍一律 `InheritedWndProc(msg)`。实测 `--bench` / `--mousetest` 都正常 3 秒退出。`WG_ERASE=1` 可随时恢复旧行为做 A/B 对照。

### 附带修掉的崩溃

挂上 `SetOnWndProc` 后，`--bench` / `--mousetest` 从 worker goroutine 直接 `os.Exit()` 会和正在执行的 Go↔DLL 回调撞车，触发 `fatal error: exitsyscall: syscall frame is no longer valid`。改为 `lcl.RunOnMainThreadAsync` → `lcl.Application.Terminate()`，由 `main()` 在消息循环结束后退出。

### 修复后的实测

该轮实测：`--bench` 稳定 **59–60 FPS**、PREPARE+PRESENT ≈ 7.5 ms、445 次画布调用；`--schedtest` 4 项断言全 PASS；`--mousetest` 180 帧 / 3s / PASS；`--selftest` 8 项 PASS；`go test -count=1 ./wg/...` 两包 ok；连续 40 帧屏幕采样 **0 帧空底**。本轮（软件 AA）没有改动 `WM_ERASEBKGND` 抑制与 PRESENT 路径，画面正确性改由帧缓冲导出的精确帧确认；Win32 `BitBlt` 屏幕抓图在本机受限，因此未重复 40 帧采样。

## 七、软件抗锯齿：图形质量与帧稳定性

用户要求「增加抗锯齿，并保证性能稳定」。结论是：**当前 LCL/GDI 后端不提供图元级 AA，且任何额外叠一层 RGBA 的方案都会直接击穿 16.67ms 预算**，只能把覆盖率混合做进已有的不透明帧缓冲。

### 先证伪两条常规路线（`--aaprobe` 实测）

| 方案 | 实测 | 结论 |
|---|---|---|
| `TCanvas.SetAntialiasingMode(AmOn)` | Polyline / Polygon / Ellipse / Arc / Pie 的 `AmOff / AmOn / AmDontCare` 中间灰阶像素**全部为 0** | GDI widgetset 无效，必须自己做软件 AA |
| 额外叠一层 RGBA 做混合后 `Draw` | 整窗 alpha `Draw` **24.72ms**；12 个局部层也要 **25.6ms/帧**；而不透明 `CopyRect` 整窗只要 **0.48ms** | 差 50 倍，绝不可行 |
| `TBitmap.ScanLine` 直写像素 | `Pf24bit` 帧缓冲的 raw image 与 GDI HDC **不同步**（GDI 填充后 raw 全黑，Go 写白后 GDI 看不见）；直写的 `Pf32bit` 位图 `Draw` 时还不会被 alpha 混合 | 不能拿 ScanLine 当帧缓冲 |

### 可行管线（`CreateIntfImage` 批量读写）

```
静态快照(Pf24bit) ──CreateIntfImage().PixelData()──> Go 像素块   ~1.01ms
        │                                              │
        │                          几何按覆盖率就地 read-modify-write
        │                                              │
        └────────背景恢复◄──── LoadFromIntfImage() ───────┘   ~1.00ms
                              ↓
                    GDI TextOut 回放中文（保留 ClearType）
                              ↓
                    一次不透明 CopyRect / BitBlt（PRESENT）   ~2.0ms
```

关键：**AA 不引入新的图层，也不引入新的 alpha 通道**；它只是把「往 GDI 画图元」换成「往 Go 像素块按覆盖率混色」，因此 PRESENT 依旧是一次纯拷贝。`Painter` 是唯一接触 `ICanvas` 的封装，按 `frame != nil` 在两套后端间切换，所有 widget 无需改动。文字不能进软件光栅器（会丢 ClearType），所以 `Text()` 把命令缓进 `textCmds`，等 `LoadFromIntfImage` 写完像素后再回放。

### 光栅器（`raster.go`）与零分配热路径

| 手段 | 作用 |
|---|---|
| 距离场覆盖率（叉积 `\|ux·dy−uy·dx\|/L` 代替逐像素 `Hypot`） | 任意宽度线段、圆头端点的亚像素 AA |
| 轴对齐线快速路径 `axisLine` | 网格/分隔线沿长度方向覆盖率恒定，内层只做 `blend` |
| 活动边表 `rasterEdge`（先按 `yLo` 排序，逐行只处理 2–4 条活动边） | 180 点波形多边形从 `O(边数×高度)` 降为 `O(N log N + 高度×活动边)` |
| 扇区判定用叉积替代 `atan2` | 饼图/圆弧扇区判定去掉三角函数 |
| `polygonGradient` 复用扫描行 | 垂直双色渐变只是每行一次 lerp，**近似免费** |
| **单遍发光** `lineCov` | halo 与实线同色，按 over 合成 `core + glow − core·glow` 从**一次**距离算出两档覆盖率，不再把整条曲线走两遍 |
| **描线覆盖率取 MAX**（`strokeBegin` / `putCov` / `strokeEnd`） | **一条折线是一个图元**：180+ 线段的圆头端点原本会在同一像素反复 over 合成，边缘覆盖度非单调、出现周期性亮珠，看起来就是"锯齿"。现在描线期间把覆盖率写进稀疏索引缓冲取 max，结束后才逐像素混合一次；代际标记免去整块清零，只在被触碰的像素上收尾 |
| 复用点数数组 / bezier 缓冲 / 状态行 slice / `time.Timer`，文本 4 次/秒刷新 | 每帧约 20 次堆分配降到接近 0，消除 GC 抖动 |
| `fitText` 按 rect 宽度裁剪 + 省略号（带 memo） | `TextOut` 无裁剪，长说明文字原本会画到相邻卡片上；顺带把图例文案改成当前 AA 架构的短描述 |

### 实测：单遍发光 + 覆盖取 MAX 后仍守住预算

| 指标 | 两遍发光（初版） | **单遍合并 + MAX（现行）** | `WG_AA=0` GDI 基线 |
|---|---|---|---|
| widgets（含 AA 光栅化） | 6.85–6.99 ms | **5.99–6.39 ms** | 4.27 ms（无 AA） |
| PREPARE+PRESENT | 10.63–11.01 ms | **9.66–10.41 ms** | 7.02 ms |
| PREPARE p50 / p95 | 12.95–13.40 / 14.10–17.02 ms | **11.97–12.13 / 13.98–15.99 ms** | 6.45 / 11.34 ms |
| 帧间隔 p50 / p95 | 16.56–16.95 / 18.96–19.49 ms | **16.48–16.61 / 19.09–20.99 ms** | 16.67 / 19.40 ms |
| jank(>20ms) | 1.1–3.9% | **2.8%**（安静复测） | 3.9% |

> 覆盖率累积第一版用「整块缓冲清零 + 按行 min/max 扫描」，在波形起伏处行范围几乎等于整窗宽度，widgets 涨到 7.06ms；改成**稀疏触碰索引 + 代际标记**（只遍历真正被覆盖的像素、免清零）后回落到 6.0–6.4ms，画质不变。
>
> 本表是「覆盖取 MAX」阶段的快照；之后第八节的四个回归修复又统一了 `lineCov`、补回 halo 像素并加入自适应布局，**最新数值见第八节末尾的实测表**。

`--bench` 安静复测 **60.1 FPS**、widgets 6.20ms、PREPARE+PRESENT 9.95ms；`--schedtest`、`--mousetest`（1330 条鼠标消息下 60.1 FPS）、`--selftest`、`go test -count=1 ./wg/...`、`WG_AA=0` / `WG_FAST=1` / `WG_STATIC=0` 全 PASS。**残留抖动不是渲染成本**：帧间隔 p95 的波动明显大于 PREPARE 的波动，来自软件墙钟与系统调度；连续启动多个进程时首轮会偏高。这不是缺陷：Windows 8+ DWM 已负责呈现同步，当前 GDI 方案不需要真 VSync（判断与触发条件见架构文档 §4·二.5）；`--schedtest` 因此改用帧间隔中位数判定，避免把外部负载顿挫误判成帧泵故障。

**A/B 画面**：`aa_frame_zoom.png` 里曲线/圆点边缘有平滑中间色、下方填充是垂直渐变、曲线带柔光；`gdi_frame_zoom.png` 是明显阶梯 + 纯色填充。9 张卡片、全部图表、中文文字在两版里都完整无缺失。

**「曲线还有锯齿」的定位与修复**：把曲线穿越处逐像素读出 B 通道，发现 AA 其实生效（不是硬阶 255↔184），但过渡序列是 `132→146→146→211→255→255→226→186→185→173→143`——**有平台、非单调**。原因是密集折线的每个圆头端点都在同一像素被 over 合成多次，形成周期性亮珠与毛边，肉眼就是"锯齿"。改为整条折线取 MAX 覆盖率后，剖面变成单调的 `178→255→…→206→145→136`，8 倍放大对照见 `edge_before.png` / `edge_after.png`：亮珠消失、斜边成为连续均匀亮带。

## 八、四个渲染回归：曲线 / 滑条 / 布局 / resize

用户同时报告了四个现象：① 波形、3 条曲线、贝塞尔线条有锯齿、粗细不均、转弯处断开；② 拖滑条后第 9 个图形内容消失；③ 窗口内所有渲染块的内容偏下；④ 调整窗口大小直接崩死/卡死。逐个用实证定位，不靠目测。

### ① 曲线：`segment()` 两个分支用了两套覆盖率公式

`raster.go` 的 `segment()` 在 `|dx| >= |dy|` 分支写的是 `clampCov(half+0.5-d)`，而 `|dy| > |dx|` 分支写的是 `lineCov(d, half)`。发光描线传入的 `half` 是 **halo 半宽**（3~3.5px），于是水平主导的线段被画成一条 6–8px 的**全实心带**，陡峭段却是 2px 实线 + 淡光晕；斜率每跨越 ±1 宽度就突变一次 —— 肉眼看到的正是「有粗有细、转弯断开、锯齿」。改为两分支统一 `lineCov`。

另外 `strokeBegin` 的 bbox padding 只有 `half+3`，而 `segment` 的遍历范围是 `(half+1)/cos+1`（45° 时最大）。超出的像素被 `putCov` 静默丢弃，转弯处 halo 被剪断。padding 改为 `sqrt2*(half+1.5)+2`。

`edge_before.png` / `edge_after.png`（8 倍最近邻）：修复前是粗实心阶梯带，修复后是粗细均匀、有中间色过渡和柔光的连续细线。

### ② 滑条内容消失 + ④ 崩死：其实是同一个 bug 的两种表现

`--resizetest` 抓到真实崩溃栈（不是猜测）：

```text
unexpected fault address ... fatal error: fault
main.(*FramePixels).blend        raster.go:195
main.(*FramePixels).fillRect     raster.go:842
main.(*Painter).FillRect         painter.go:136
main.(*Scene).ensureStaticBuffer staticlayer.go:45
main.(*Scene).bindCanvas.func1   scene.go:382
```

静态层重烘焙可能发生在 **OnPaint 内部**（拖滑条改 `Opacity` → `InvalidateStatic`，或 `WM_SIZE` 改客户区）。此时 `painter.frame` 仍指向上一帧的 AA 表面，于是：

- `FillRect` 走软件光栅器写进 `framePix.px` 而不是正在烘焙的 GDI 位图；resize 时 `frameImg` 刚被释放，这块 Go slice 已悬空 → **段错误**。
- 静态 widget（图例）的 `Text` 进入 `textCmds` 延迟队列，烘焙前后两次 `Reset()` 把它清空 → 第 9 张卡片「架构与操作」的 6 行文字消失。首次启动的预烘焙发生在 OnPaint 之外（`frame == nil`），所以初始画面正常，只有**二次烘焙**才暴露 —— 这就是「调滑条才消失」。

三处修复：`ensureStaticBuffer` 临时 `p.frame = nil` 强制 GDI 后端、结束恢复现场；`ensureBackBuffer` 释放旧位图前先 `freeAASurface()`（`frameImg` 由旧位图 `CreateIntfImage` 得来，顺序反了就是悬空 DIB）；`freeAASurface` 把 `framePix.px = nil`，并给 `blend` / `copyFrom` 加空缓冲守卫。`WG_BAKE=1 --screenshot` 导出的 `rebake_frame.png` 证明重烘焙后 6 行文字与 9 张卡片完整。

### ③ 内容偏下 → 自适应布局

逐卡片统计「彩色像素垂直重心」：`仪表 -34 / 折线 -34 / 控件 -33 / 柱状 -19 / 饼图 -20 / 花瓣 -18 / 说明 -17 / 性能 -17 px`，8/9 张其实是**偏上**的，所以问题在整体网格（卡片 96..828，窗口 880，上 96 / 下 52）与硬编码绝对坐标，不在卡片内部。

`layout.go` 把 1440×880 设计画布按 `min(cw/1440, ch/880)` 等比缩放并居中，内容块整体上提 16px 让上下留白均衡；`Scene.relayout` 在每次客户区尺寸变化时**一次性**重算所有 widget rect（`Controls` 的子控件同步重映射，保证 hit test 仍正确）；字号、多行行距、`inset`、圆角半径、控件子偏移统一乘 `Scene.S()`，否则卡片缩小后固定字号文字会溢出到相邻卡片（1000×640 首版就出现过）。

`resize_small.png`（1000×640，scale=0.694）：9 张卡片、性能面板、6 行说明、控件全部完整且互不重叠。

### ④ 卡死：每个中间尺寸都重烘焙

9 张 RGBA 玻璃卡片烘焙一次约 85–100ms，拖动窗口边缘会连续发 `WM_SIZE`，每个尺寸都重建一次 → 消息积压（诊断中放大到 1600×960 时 350ms 内 **0 帧**，随后 38 帧一次性涌出）。`WM_SIZE` / `WM_SIZING` 打时间戳，**180ms 冷却期**内保留旧快照、只画轻量背景并临时走 GDI 路径，尺寸稳定后才烘焙一次；AA 判定同时加上 `staticW/staticH == 当前尺寸` 且 `len(staticPx) == len(framePix.px)`，杜绝尺寸不匹配时的越界读。

`--resizetest` 7 次真实尺寸切换：修复前 step3 崩溃；加入冷却后每步 10–25 帧、无 WARN、`RESULT: PASS`。

> 冷却逻辑还暴露了**第二次偶发崩溃**：栈落在 `captureStatic` 的 `copy(staticPx, px)`（`runtime.memmove` SIGSEGV）。冷却期内静态快照仍是旧尺寸，而代码用**新客户区高度** `r.Height()` 算 `need` 去读旧 DIB → 越界读。`captureStatic` 改为按快照自身的 `staticW/staticH` 计算，且尺寸与帧缓冲不一致时直接跳过（此时 `aa` 判定本就为 false，不会用到旧像素）。修复后连续 3 次 `--resizetest` 全部 `exit=0 / RESULT: PASS`。

### 本轮实测（1440×880 真实窗口）

| 指标 | 安静复测 ×3 |
|---|---|
| FPS | **59.9 / 60.0 / 60.0** |
| widgets（含 AA 光栅化） | 6.35–6.63 ms |
| PREPARE+PRESENT | 10.04–11.27 ms |
| PREPARE p50 / p95 | 13.27–13.62 / 16.13–18.79 ms |
| 帧间隔 p50 / p95 | 16.44–16.57 / 19.01–20.87 ms |
| jank(>20ms) | 安静 2.2%；连续启动其它进程时首轮 10.7% |

比上一轮多出的约 1ms 来自 `strokeBegin` 补回原本被 bbox 裁掉的 halo 像素（修复断开的必要成本），仍在 16.67ms 预算内。`go build ./wg/...`、`go test -count=1 ./wg/...`、`--selftest`、`--schedtest`、`--mousetest`、`--aaprobe`、`--resizetest`、`WG_AA=0` / `WG_FAST=1` / `WG_STATIC=0` / `WG_BAKE=1` / `WG_SIZE=...` 全 PASS。

## 九、已修复：调整大小 / 最大化后标题栏、还原、鼠标全部失效（帧消息饿死输入）

**现象**：拖动边框或最大化之后，画面还在动，但标题栏按钮、还原、窗口内所有鼠标操作都不响应——「像卡死，但没卡死，因为图形还在渲染」。

### 先证伪「鼠标 capture 卡住」

交接文档里怀疑 `WM_SIZING` 模态循环吞掉 `WM_LBUTTONUP` 导致 `s.capture` 永不清除。新增 `--inputdiag` 记录 `WM_SIZE` / `WM_SYSCOMMAND` / `WM_ENTERSIZEMOVE` / `WM_EXITSIZEMOVE` / `WM_CANCELMODE` / `WM_CAPTURECHANGED` / 鼠标按钮，并同时打印 `GetCapture()`、LCL `host.MouseCapture()`、`Scene.capture`、`WM_NCHITTEST` 结果，实测最大化后的状态是：

```
STATE max-held  client=2880x1557 zoomed=true osCapture=0x0 lclCapture=false sceneCapture=nil
PROBE max-held  NCHITTEST(caption) -> 2 (HTCAPTION)
```

即 **OS 捕获、LCL 捕获、场景捕获全部为空，标题栏命中测试完全正常**；真实 `SendInput` 拖边框也能正常进出模态循环。capture 假设不成立。

### 真正的根因：Win32 消息队列优先级

`GetMessage` 取消息的顺序是：sent → **posted** → 硬件输入（鼠标/键盘/非客户区）→ `WM_PAINT` → `WM_TIMER`。我们的帧泵用 `PostMessageW(WM_APP+1)` 投递帧槽，属于 **posted 消息，优先级高于一切鼠标输入**。

在 200% DPI 的屏幕上最大化后客户区是 **2880×1557**（约设计画布的 4 倍），单帧 PREPARE+PRESENT ≈ **30ms**，超过 16.67ms 预算。此时旧帧泵：

1. 允许 **2 个 outstanding** 帧消息，且每 16.67ms 无条件再投一个 → post 队列永远不为空，`GetMessage` 永远取不到输入；
2. 「3 个槽位没收到 `FrameDone` 就判定消息丢失、把 `inflight` 清零」的看门狗，在 ~190ms 的静态层重烘焙期间反复触发，**又往队列里灌进几十条帧消息**，形成长期积压；
3. 结果窗口一路处理 `WM_APP+1`（图形继续渲染），而标题栏点击、`SC_RESTORE`、鼠标消息全部排队饿死。

**决定性 A/B**（`--inputtest` Step 7，同为最大化状态）：注入一次无副作用的点击，测到派发延迟

| | 帧泵开启 | 帧泵关闭 |
|---|---|---|
| 修复前 | **382 ms** | 14 ms |
| 修复后 | **16–32 ms** | 3–4 ms |

修复前最大化状态下点「切换主题」按钮 `light -> light` **FAIL**（点击根本没被派发）；标题栏 `WM_NCHITTEST` 用 `SendMessage` 却能立刻返回 `HTCAPTION`——这正是「sent 消息能过、posted/硬件输入过不去」的典型特征。

### 修复（`framescheduler.go` / `main.go` / `scene.go`）

| 改动 | 原因 |
|---|---|
| **最多 1 个 outstanding 帧消息**（`post()` 用 `CAS(0,1)` 拒绝第二次投递） | post 队列在两帧之间必然为空，输入才有机会被 `GetMessage` 取到 |
| 删掉「3 槽没回就清零 `inflight`」的伪丢包看门狗，改成 3 秒病理超时 | 长帧/重烘焙不是丢消息；原逻辑正是积压来源 |
| 泵在投递前 `waitConsumed()` 等上一帧 `FrameDone`，慢帧**平移时槽**而不是丢整槽 | 闭环调度；既无追赶帧爆发，也不会因一次慢帧连锁掉帧（`--schedtest` 曾误报 39 FPS） |
| 超过预算的帧之后留 `inputBreath = 1.5ms` 空队列窗口 | 让主线程进入 `GetMessage` 时取走挂起的鼠标/键盘/NC 消息；只对慢帧收费，正常 60Hz 不变 |
| `FrameDone(costUs)` 回报上一帧成本 | 让「是否让路给输入」由真实帧耗时决定，不靠猜 |
| `WM_CANCELMODE` / `WM_ENTERSIZEMOVE` / `WM_EXITSIZEMOVE` / `WM_ACTIVATE(deactivate)` 清 `Scene.capture`；`EvMouseMove` 时用 `GetAsyncKeyState(VK_LBUTTON)` 判断真实按键，若已松开则补发 `EvMouseUp` | 修掉诊断顺带发现的次生问题：模态循环吞掉 `WM_LBUTTONUP` 后 LCL 会静默释放自己的 capture（日志里 `osCapture=0 / lclCapture=false / sceneCapture=控件`），场景里的滑块会永远跟着鼠标改不透明度。`WM_CAPTURECHANGED` 不可靠（它也会在正常 MouseUp 前到达，会破坏拖拽结束语义） |
| **帧 = 一个帧槽**，计数与 FPS 统计移到 `RenderFrame()`；`SetActive(false)` 后投递前再查一次 active | LCL 的 `TControl.Repaint` = `Invalidate + Update`，一个帧槽会产生一次同步绘制 + 一条事后到达的低优先级 `WM_PAINT`。闭环泵让这条消息在暂停期间冒出来，旧代码把 `OnPaint` 当帧计数会误报「静止期 24 帧」；按帧槽计数后静止严格零帧，也杜绝了暂停后仍投递一帧/连锁 one-shot |
| `--schedtest` 排除启动预热（先等 30 帧），动画段改 2s，判据改用**两段合并的帧间隔中位数**（14–20ms，≥100 样本） | 绝对帧数在 1–2s 窗口里会被外部 CPU/杀软顿挫吃掉几十帧，产生 12–39 FPS 的假 FAIL（`frameProfile.sample()` 已丢弃 >100ms 的空档）。中位数断言的是「帧泵锁 60Hz」这个真实契约；真坏到 30Hz 会让两段都变坏，依然 FAIL |

### 复测

`--inputtest` 固化了最大化 → 还原 → 丢失 MouseUp → 真实 `SendInput` 边框拖拽 → 标题栏双击 → 输入延迟 A/B 共 **22 项断言**，连跑 3 次全 PASS；其中「最大化后点击生效」修复前必 FAIL。性能不变：

```
--bench        60.0–60.4 FPS   PREPARE+PRESENT 10.3–11.7ms   PREPARE p50 12.5–13.5ms
--schedtest    ×8 全 PASS（暂停 0 帧 / 脏区 1 帧 / 合并帧间隔中位数 16.5–17.1ms）
--mousetest    ×3 全 PASS（1264–1268 条 WM_MOUSEMOVE，58.5–60.0 FPS）
--resizetest   7 步每步 20 帧，PASS     --inputtest 22 项 ×3 全 PASS
--selftest / go test ./wg/... / WG_AA=0 / WG_FAST=1 / WG_STATIC=0 / WG_BAKE=1 / WG_SIZE=... 全 PASS
```

**残留限制**：最大化瞬间仍需在主线程烘焙一次 2880×1557 的静态快照（~190ms），期间会有一次可感知的顿挫；但队列不再积压，输入在烘焙结束后立刻恢复。生产后端应把静态层按 DPI 降采样烘焙或改为 GPU 合成。

## 十、文件

| 文件 | 作用 |
|---|---|
| `main.go` | 1440×880 窗口、库探测、设计坐标布局、`WM_ERASEBKGND` 抑制、`WM_SIZE` 去抖时间戳、取消类消息清 capture、帧泵启动、全部命令行模式分派（`--selftest` / `--bench` / `--mousetest` / `--schedtest` / `--resizetest` / `--inputtest` / `--inputdiag` / `--aaprobe` / `--screenshot`）、键盘绑定、主线程优雅退出 |
| `framescheduler.go` | **帧泵核心**：绝对 60Hz 墙钟时间轴、被动 `WM_APP+1` 投递、动画/脏区双帧源、慢帧跳槽不追帧、**单 outstanding 闭环投递 + 慢帧后让路给输入**（见第九节） |
| `framescheduler_windows.go` / `_other.go` | `PostMessageW` 把帧槽交给主线程 / 非 Windows stub |
| `staticlayer.go` | **静态层**：把 9 卡片 + 2 光斑 + 图例一次性烘成不透明 `Pf24bit` 快照；重烘焙时强制切回 GDI 后端（见第八节 ②）；尺寸不匹配时拒绝 Blit |
| `layout.go` | **自适应布局**：1440×880 设计画布 → 当前客户区的等比缩放 + 居中，内容块整体上提 16px |
| `scheduler_windows.go` / `_other.go` | `timeBeginPeriod(1)` 抬高系统定时器精度 / 非 Windows stub |
| `schedtest.go` | `--schedtest`：实证 60Hz 锁帧 / 静止零帧 / 脏区单帧 |
| `resizetest.go` | `--resizetest`：真实 `WM_SIZE` 连续切换 7 种尺寸，断言每步都有帧、主线程未卡死，并打印客户区几何 |
| `profile.go` | 逐组件、逐阶段（PREPARE / PRESENT）耗时剖析 |
| `layer.go` | `Theme`、`Surface`、`AlphaLayer`（批量 BGRA 写入）、`LayerCache`、`bakeGlass` / `bakeGlow`、圆角覆盖率 |
| `raster.go` | **软件 AA 光栅器** `FramePixels`：直接绑定帧缓冲像素块，按覆盖率就地混色；距离场线段（两分支统一 `lineCov`）/ 活动边表多边形 / 椭圆 / 圆弧 / 扇区 / 圆角矩形 / 渐变填充 / 单遍发光 / 描线覆盖率取 MAX，全部零额外图层 |
| `painter.go` | `Painter` —— 唯一接触 `lcl.ICanvas` 的封装；GDI 与软件 AA 双后端分派、状态合并、文本度量/裁剪缓存、延迟文字回放、`Scale` 自适应字号。**画刷先写颜色后写样式**（`SetColor` 会重置 Style） |
| `scene.go` | 单画布宿主、`newBackBuffer(Pf24bit)`、AA 表面的创建/释放顺序与尺寸一致性校验、静态快照批量读 / 写、`relayout`、180ms resize 冷却、脏区合并与局部提交、事件分发（**输入不请求帧**）、`releaseStaleCapture` 硬件按键自愈 |
| `widgets.go` | 波形/折线/仪表/花瓣/柱状/饼图/玻璃卡片/光斑/性能面板/控件/图例；`IsStatic()` 声明静态层归属；`SetRect` 支持子控件重映射；所有固定像素常量走 `Scene.S()` |
| `aaprobe.go` | `--aaprobe` 无窗口能力探测：LCL AA 是否有效、`ScanLine` 与 `IntfImage` 读写语义、alpha Draw vs 不透明 CopyRect 成本 |
| `screenshot.go` | `--screenshot` 从帧缓冲导出精确合成帧 PNG + 3x 局部放大图；支持 `WG_SIZE` 改尺寸、`WG_BAKE` 触发二次烘焙 |
| `showcase_test.go` | 5 个纯逻辑单测，不需要 GUI |
| `mousetest_windows.go` / `mousetest_other.go` | `--mousetest` 的 `WM_MOUSEMOVE` 注入与断言 / 非 Windows stub |
| `inputdiag.go` / `inputdiag_other.go` | `--inputdiag` 原生消息/capture 证据日志；`--inputtest` 的 22 项回归断言、真实 `SendInput` 拖拽、输入延迟 A/B / 非 Windows stub |
| `inputguard_windows.go` / `inputguard_other.go` | `GetAsyncKeyState(VK_LBUTTON)` 读取真实按键状态，识别被模态循环吞掉的 `WM_LBUTTONUP` / 非 Windows stub |

**本示例是可行性测试，不是产品代码**。实证结论已收口为正式架构：`C:/Users/Administrator/WorkBuddy/worksapce/wg-canvas-arch/ARCHITECTURE.md`（第 3 轮实证定稿：五层架构、POC 证伪清单、契约 ①–⑰），施工蓝图见同目录 `API.md`（包结构 / 接口 / 迁移映射 / 验收命令）。索引与 P1 抽取步骤见 `../README.md` 末节。单画布 + 输入法的最小示例见 `../canvas_ime`。本示例使用 `examples` 模块已有依赖，未修改 `go.mod`。
