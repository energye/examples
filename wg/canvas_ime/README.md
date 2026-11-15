# 单画布渲染示例 · 输入法（IME）测试

> 位置：`examples/wg/canvas_ime/`，属于 `examples/wg` 示例集合；由 `go.work` 直接管理，无独立 go.mod。

> 设计目标（原文）：
> 「只有一个画布，然后在上面画各种渲染基础能力和组件的功能实现。还有输入具有焦点的也都使用它实现。不使用有句柄的也就是纯自绘的任意组件。」
>
> 本示例把这句话直接做成了可运行程序：窗口（Form）是唯一持有 HWND 的对象；挂在窗口下的 `ICustomGraphicControl` 是无句柄画布；所有组件都是纯 Go 对象，画在同一张画布上，包括带焦点、能输入中文的输入框。界面标题与副标题也直接用了这段文案。

## 一、它演示了什么

| 渲染基础能力 | 落地位置 |
|---|---|
| 圆角矩形 / 描边 / 状态色 | `Painter.RoundRect`，Button / Input / StatusPanel |
| 垂直渐变 | `Painter.Gradient`，进度条与面板顶部渐变带 |
| 文本度量 / 对齐 / 多行状态文本 | `Painter.Text` / `TextWidth` / `TextHeight` |
| 伪阴影（层叠） | `Painter.Shadow` |
| 光标 / 预编辑下划线 | `Painter.Caret` + `Input.Paint` |
| 统一主题（dark / light 一键切换） | `Theme`，所有 Widget 不自己选颜色 |
| 命中测试 / hover / 捕获 / 冒泡 / 焦点链 / Tab | `Scene` |
| 脏区合并（概念演示） | `Scene.InvalidateRect` + 面板显示 |

组件：`Label`、`Button`、`Input`（焦点 + 中文输入 + IME 预编辑 + 完整键盘编辑）、`ProgressBar`（点击调值）、`StatusPanel`（分层卡片 + IME 实时状态）。**全部零 LCL 实例、零句柄**。

### 键盘编辑功能（纯 Go 实现，没有任何原生编辑框）

| 按键 | 行为 |
|---|---|
| `Backspace` / `Delete` | 删除光标前 / 后的字符；有选区时删除整段选区 |
| `←` `→` `Home` `End` | 移动光标；有选区时先折叠到选区端点 |
| `Shift` + 方向键 / Home / End | 扩展选区（蓝色高亮自绘） |
| `Ctrl` + `←` / `→` | 按词移动光标 |
| `Ctrl+Backspace` | 删除光标前的一个词 |
| `Ctrl+A` | 全选（大小写 VK 码都识别，不依赖键盘布局） |
| `Ctrl+C` / `Ctrl+Insert` / `Ctrl+X` / `Shift+Delete` / `Ctrl+V` / `Shift+Insert` | 复制 / 剪切 / 粘贴（`lcl.Clipboard`，粘贴时换行转空格；剪贴板不可用时不会误删选区） |
| `Ctrl+Z` | 撤销（30 步历史，**同时恢复 caret 与选区位置**） |
| `Enter` | 单行输入框提交，状态面板显示「提交: ...」 |
| `Esc` | 清空 IME 预编辑串并取消选区 |
| `Tab` | 在可聚焦组件间切换 |

光标按 **rune 索引**定位，中文/emoji 不会按字节截断；选区高亮、光标、IME 预编辑串都在同一个自绘渲染流程里。状态面板「编辑」行实时显示 caret、选区范围、文本长度与撤销栈深度，另有「最近按键」与「已过滤控制字符」两行。

#### 已修复：删除键乱码 / 快捷键失效（重要）

`SetOnUTF8KeyPress` **不是**「可打印文本」事件。Windows `TranslateMessage` 会把 `Backspace` 变成 `WM_CHAR 0x08`、`Ctrl+A` 变成 `0x01`、`Enter` 变成 `0x0D`、`Tab` 变成 `0x09`，LCL 把它们**全部**转发到 `UTF8KeyPress`。初版无条件 `insert()` 这些字节，于是：按退格/删除后多出一个控制字符方块（看起来就是乱码），`Ctrl+A` 全选后立刻被 `\x01` 替换掉、`Ctrl+V` 粘贴后立刻被 `\x16` 覆盖、`Ctrl+Z` 撤销后又被 `\x1A` 覆盖——这就是「其它功能也不对」的来源。

实测证据（`--keydiag` 阶段的原始事件日志）：

```
KeyDown   key=46 0x002E   shift=0x00000000     // Delete 是 VK 码，正确
KeyDown   key=8  0x0008   shift=0x00000000     // Backspace
UTF8Key   len=1 bytes=08 str="\b"              // 控制字符也进来了
UTF8Key   len=1 bytes=01 str="\x01"            // Ctrl+A
UTF8Key   len=1 bytes=0D str="\r"              // Enter
UTF8Key   len=1 bytes=09 str="\t"              // Tab
```

修复方式：在**事件边界**用 `sanitizeText()` 统一拦截 C0(0x00–0x1F)、DEL(0x7F)、C1(0x80–0x9F)、无效 UTF-8 与 shortstring NUL 填充，并把干净值写回 `*text` 阻止 LCL 继续转发；这些键的语义全部由 `SetOnKeyDown` 的 VK 码处理。中文/emoji 上屏字符的 rune 均 ≥ 0xA0，不受影响。`Input.insert()` 里再过滤一次做纵深防御。剪贴板访问也加了 nil / 异常保护。

## 二、输入法测试是怎么做的

窗口级 `TEngForm.SetOnWndProc` 拦截全部消息，末尾**必须** `InheritedWndProc(msg)` 透传；`SetKeyPreview(true)` + `SetOnKeyDown` / `SetOnUTF8KeyPress` 负责键盘与上屏字符。

| 环节 | 实现 |
|---|---|
| 预编辑串 | `ImmGetContext` → `ImmGetCompositionStringW(GCS_COMPSTR)` → `ImmReleaseContext`，结果交给 `Input` 用黄色 + 下划线绘制 |
| 组合光标位置 | `ImmGetCompositionStringW(GCS_CURSORPOS)` |
| 上屏结果 | `GCS_RESULTSTR` 仅显示在面板（实际插入走 `UTF8KeyPress`，避免重复） |
| 候选框 / 组合框定位 | `TControl.ClientToParent` 把画布坐标转成 **Form 客户区坐标**（微软文档：两个结构体的 `ptCurrentPos` 都相对包含窗口左上角，**不是屏幕坐标**），再调 `ImmSetCompositionWindow(CFS_POINT\|CFS_FORCE_POSITION)` + `ImmSetCandidateWindow(CFS_CANDIDATEPOS)`；在 `InheritedWndProc` 之后设置，并在焦点变化 / 上屏后预先设置 |
| 定位诊断 | 面板显示 `client(x,y) screen(x,y) comp=1 cand=1`，可直接判断 API 返回码与坐标是否合理 |
| 无输入法自检 | 「模拟组合串」按钮直接注入 `ni'hao`，验证预编辑渲染与定位链路 |
| 实时观测 | 面板显示组合串、上屏结果、消息计数、最近 3 条消息、最近按键与被过滤的控制字符数；控制台同步打印 `[IME] ...` / `[KEY] ...` |
| 高频消息 | `WM_IME_NOTIFY (0x282)` 只计数、每 30 条刷新一次面板，不逐条打印/重绘，避免刷屏挤掉有用的 COMPOSITION 日志 |

## 三、运行

```bash
cd E:/app/workspace/examples        # go.work 已 use ./examples
go run ./wg/canvas_ime --selftest     # 不弹窗，只验证 DLL 导出与 IME 接口
go run ./wg/canvas_ime --keyselftest  # 自动注入真实按键消息并断言（约 3 秒后自动退出）
go run ./wg/canvas_ime                # 打开窗口
go test ./wg/canvas_ime               # 编辑逻辑单测（9 个用例）

# 或生成可执行文件（examples/.gitignore 已忽略 *.exe）
go build -o wg/canvas_ime/canvas_ime.exe ./wg/canvas_ime
wg/canvas_ime/canvas_ime.exe
```

`--keyselftest` 打开**真实的示例窗口**（同一套事件绑定），用 `PostMessage` 注入 `WM_KEYDOWN/WM_CHAR/WM_KEYUP`，断言编辑器文本，无需人工敲键。本机实测：

```
==== KEYSELFTEST RESULT ====
初始值              got="abcd" want="abcd"  PASS
Backspace           got="abc"  want="abc"   PASS
Home + Delete       got="bc"   want="bc"    PASS
控制字符全部过滤      got="bc"   want="bc"    PASS
方向键边界安全        got="bc"   want="bc"    PASS
过滤控制字符数: 13 (应 >= 6)
RESULT: PASS
```

> 注：注入的 `WM_CHAR` 不含真实扫描码/重复计数，LCL 会额外派发几次 `UTF8KeyPress`，所以「过滤控制字符数」比注入条数多；这是注入伪影，控制字符全部被丢弃，不影响断言，真实键盘不会这样。

运行时库优先级：`ENERGY_LIB` 环境变量 > exe 同目录 `libenergy-amd64.dll` > `~/.energy/runtime/libenergy-amd64.dll` > `E:/app/workspace/lib/libenergy-x86_64-win64.dll`。程序在 `lcl.Init()` 前把找到的路径写入 `libname.LibName`。若加载失败：

```bash
ENERGY_LIB=E:/app/workspace/lib/libenergy-x86_64-win64.dll go run ./wg/canvas_ime
```

本机实测自检输出：

```
[SELFTEST] runtime library: C:\Users\Administrator\.energy\runtime\libenergy-amd64.dll
[SELFTEST] library exports: OK: 6 runtime library exports resolved
[SELFTEST] IME bindings: OK: 6 imm32/user32 IME procedures resolved
[SELFTEST] GCS_RESULTSTR=0x800
```

被校验的 6 个库导出：`TCustomGraphicControl_Create/OnPaint/OnMouseDown`、`TEngForm_InheritedWndProc/OnWndProc`、`TWinControl_SetFocus`。

## 四、手动测试清单（真实中文输入法）

1. 运行 `go run ./wg/canvas_ime`（或双击生成的 `canvas_ime.exe`）；确认标题栏是「单画布渲染示例 · 输入法(IME)测试」。
2. 点一下「自绘输入框」（圆角边框变蓝 = 获得焦点；Tab 可在输入框与各按钮间轮换）。
3. 切换到中文输入法（微软拼音 / 搜狗等），敲 `nihao`：
   - 输入框内应出现**黄色带下划线的预编辑串**；
   - 候选框应出现在自绘光标下方（由 `ImmSetCompositionWindow` / `ImmSetCandidateWindow` 定位）；
   - 面板 `IME 定位` 行应显示 `client(x,y) ... comp=1 cand=1`（两个 API 均返回成功）；
   - 面板 `IME 组合串` 显示拼音、`IME 消息数` 递增、控制台打印 `COMPOSITION lParam=0x...`。
4. 按空格 / 选字上屏：
   - 汉字插入输入框；
   - 面板 `上屏结果` 显示 `UTF8KeyPress: 你好`（或 `GCS_RESULTSTR: ...`）；
   - 预编辑串清空（`ENDCOMPOSITION`）。
5. 验证 Backspace 删除、Tab 切换焦点、Esc 清空组合串。
6. 点「自绘按钮」（进度条跳到 75%）、点击进度条任意位置、点「切换主题」（全窗口统一换色）、点「模拟组合串」（无需输入法即可看预编辑渲染）。
7. 键盘编辑：输入一段中文 → `Home`/`End`、`←`/`→` 移光标 → `Shift+←` 选几个字（看高亮）→ `Ctrl+C` → `End` → `Ctrl+V` 粘贴 → `Ctrl+Z` 撤销 → `Delete` 删光标后字符、`Backspace` 删光标前字符（**不应再出现乱码方块**）→ `Ctrl+A` 全选后直接打字替换 → `Enter` 提交（面板显示「提交: ...」）→ `Esc` 取消选区。再试 `Ctrl+←`/`Ctrl+→` 按词跳、`Ctrl+Backspace` 删词。每一步都可对照面板「编辑」「最近按键」「已过滤控制字符」三行，控制台也会打印 `[KEY] ...`。

## 五、预期差异与排错

- **删除/退格后出现乱码方块，或 Ctrl+A/C/V/Z 看起来失效**：原因是 `UTF8KeyPress` 会连控制字符一起送达（见第一节「已修复」）。本示例已在事件边界过滤并写回 `*text`，面板「已过滤控制字符」计数应递增、`[KEY]` 控制台会打印 `过滤控制字符 08` 等；若计数不涨，说明过滤层被绕过。可跑 `go run ./wg/canvas_ime --keyselftest` 复现验证。
- **候选框没跟随光标**：先看面板 `IME 定位` 行。`comp=1 cand=1` 表示 `ImmSetCompositionWindow` / `ImmSetCandidateWindow` 都返回成功，且 `client(x,y)` 应接近输入框光标的客户区坐标；`WAIT: no IME context` 表示窗口还没拿到焦点。若坐标看着像屏幕坐标（明显偏大、等于窗口在桌面上的位置），就是误把 `ClientToScreen` 结果传给了 IMM32 —— 本示例初版正是这个 bug，已改为客户区坐标。若坐标正确但候选框仍不跟随，多半是 **TSF 输入法**（微软拼音等）自己渲染候选窗口、不响应 IMM32 定位，需要走 TSF `ITfContextView` / `ITfContextOwner` 或隐藏原生 Edit 代理；这是纯自绘输入框最硬的一块，属于输入法侧限制，不是"没实现"。
- **字符重复**：个别输入法同时产生 `WM_IME_CHAR` 与 `UTF8KeyPress`。本示例只在 `UTF8KeyPress` 插入文本，`GCS_RESULTSTR` 只做展示；若仍重复，面板会同时出现两条记录，便于判断来源。
- **无 `WM_IME_COMPOSITION`**：说明该输入法走的是 `WM_IME_CHAR` 路径；此时预编辑串不显示，但汉字仍可上屏。可按需在 `WM_IME_CHAR` 分支补读。
- **窗口不刷新 / 不响应**：只可能是 `SetOnWndProc` 回调漏调 `m.InheritedWndProc(msg)`，本示例已在每个分支之后统一调用。
- **离屏绘制**：`ICustomGraphicControl.Canvas()` 只在 `OnPaint` 内有效（LCL 临时绑定父窗口 DC）；需要缓存/离屏时用 `IBitmap.Canvas()`，不要保存画布引用。

## 六、文件

| 文件 | 作用 |
|---|---|
| `main.go` | 窗口、库探测、`--selftest` / `--keyselftest`、WndProc/键盘/UTF8 绑定（含控制字符过滤）、组件布局 |
| `scene.go` | 单画布宿主、命中测试、焦点链、事件分派、脏区、IME 状态、按键日志 |
| `render.go` | `Painter` 与 `Theme`，唯一接触 `lcl.ICanvas` 的地方 |
| `widgets.go` | 5 个纯 Go 组件；`sanitizeText`、VK 映射、选区/撤销/剪贴板/按词移动等编辑逻辑 |
| `ime_windows.go` | imm32/user32 绑定、组合串读取、候选框定位、自检 |
| `ime_other.go` | 非 Windows stub（保证跨平台编译通过） |
| `ime_consts.go` | 跨平台共享的 IME 常量 |
| `keydiag_windows.go` / `keydiag_other.go` | `--keyselftest` 的 Win32 消息注入与断言 / 非 Windows stub |
| `widgets_test.go` | 9 个编辑逻辑单测（rune 光标、控制字符、组合键、撤销、按词移动） |

架构设计见 WorkBuddy 工作区 `wg-canvas-arch/ARCHITECTURE.md`，最小骨架见 `wg-canvas-arch/poc/`。本示例使用 `examples` 模块已有的依赖（`golang.org/x/sys v0.30.0`），无需修改 `examples/go.mod`。
