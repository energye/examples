// GPU Test: DrawSelectionHighlight - 选择高亮渲染验证
package main

import (
	"fmt"
	"github.com/energye/lcl/lcl"
	"os"
	"path/filepath"
	"time"

	"github.com/energye/lcl/api/libname"

	"github.com/energye/examples/lcl/gpui/core/math"
	"github.com/energye/examples/lcl/gpui/render/pipeline"
	"github.com/energye/examples/lcl/gpui/style/token"
	"github.com/energye/examples/lcl/gpui/ui"
	"github.com/energye/examples/lcl/gpui/widget"
)

// SelectionHighlightTestWidget 自定义控件，用于渲染选择高亮测试
type SelectionHighlightTestWidget struct {
	widget.BaseWidget
}

func NewSelectionHighlightTestWidget() *SelectionHighlightTestWidget {
	w := &SelectionHighlightTestWidget{
		BaseWidget: widget.NewBaseWidget(),
	}
	w.SetOwner(w)
	return w
}

// Render 渲染选择高亮测试内容
func (w *SelectionHighlightTestWidget) Render(ctx *widget.Context) {
	if w == nil || ctx == nil || ctx.Renderer == nil || ctx.Font == nil {
		return
	}
	bounds := w.Bounds()
	ctx.Renderer.PushClip(bounds)

	f := ctx.Font
	lineHeight := f.LineHeight()

	// 测试1: 部分文本选择 - 选择 "World"
	text1 := "Hello World"
	text1X := float32(100)
	text1Y := float32(120)
	ctx.Renderer.DrawText(text1, text1X, text1Y, f, math.NewColor(0, 0, 0, 1))
	// 计算 "Hello " 的宽度作为选择起点
	helloWidth := f.TextWidth("Hello ")
	worldWidth := f.TextWidth("World")
	// 高亮区域垂直居中于文本，上下各留2px边距
	highlightPadding := float32(2)
	ctx.Renderer.DrawSelectionHighlight(
		math.NewRect(text1X+helloWidth, text1Y-highlightPadding, worldWidth, lineHeight+highlightPadding*2),
		math.NewColor(0.2, 0.5, 1, 0.4),
	)

	// 测试2: 全部文本选择
	text2 := "Hello World"
	text2X := float32(100)
	text2Y := float32(280)
	text2Width := f.TextWidth(text2)
	ctx.Renderer.DrawText(text2, text2X, text2Y, f, math.NewColor(0, 0, 0, 1))
	ctx.Renderer.DrawSelectionHighlight(
		math.NewRect(text2X, text2Y-highlightPadding, text2Width, lineHeight+highlightPadding*2),
		math.NewColor(0.2, 0.5, 1, 0.4),
	)

	// 测试3: 不同透明度选择 - 使用更高透明度确保可见
	alphaY := float32(430)
	ctx.Renderer.DrawSelectionHighlight(math.NewRect(100, alphaY, 200, lineHeight+highlightPadding*2), math.NewColor(0.2, 0.5, 1, 0.2))
	ctx.Renderer.DrawSelectionHighlight(math.NewRect(350, alphaY, 200, lineHeight+highlightPadding*2), math.NewColor(0.2, 0.5, 1, 0.35))
	ctx.Renderer.DrawSelectionHighlight(math.NewRect(600, alphaY, 200, lineHeight+highlightPadding*2), math.NewColor(0.2, 0.5, 1, 0.5))
	ctx.Renderer.DrawSelectionHighlight(math.NewRect(850, alphaY, 200, lineHeight+highlightPadding*2), math.NewColor(0.2, 0.5, 1, 0.7))

	// 测试4: 不同颜色选择
	colorY := float32(580)
	ctx.Renderer.DrawSelectionHighlight(math.NewRect(100, colorY, 200, lineHeight+highlightPadding*2), math.NewColor(0, 0, 1, 0.4))  // 蓝
	ctx.Renderer.DrawSelectionHighlight(math.NewRect(350, colorY, 200, lineHeight+highlightPadding*2), math.NewColor(0, 1, 0, 0.4))  // 绿
	ctx.Renderer.DrawSelectionHighlight(math.NewRect(600, colorY, 200, lineHeight+highlightPadding*2), math.NewColor(1, 0, 0, 0.4))  // 红
	ctx.Renderer.DrawSelectionHighlight(math.NewRect(850, colorY, 200, lineHeight+highlightPadding*2), math.NewColor(1, 1, 0, 0.4))  // 黄

	ctx.Renderer.PopClip()
}

func main() {
	libname.UseWS = "gtk3"
	if ws := os.Getenv("GPUI_WS"); ws != "" {
		libname.UseWS = ws
	}
	app := ui.NewApplication("GPU Test: DrawSelectionHighlight", 1600, 900)

	app.OnSetup(func(engine *ui.Engine) {
		setupUI(engine)
	})

	app.Run()
}

func setupUI(engine *ui.Engine) {
	tokens := token.Current()
	outDir := os.Getenv("GPUI_TEST_OUTPUT")
	if outDir == "" {
		outDir = "lcl/gpui/test_output/gpu_tests/selection_highlight"
	}
	os.MkdirAll(outDir, 0o755)

	// 背景面板
	panel := widget.NewBox(pipeline.BoxStyle{
		Background:  tokens.Global.ColorBgContainer,
		BorderColor: tokens.Global.ColorBorder,
		BorderWidth: 1,
		Radius:      tokens.Global.RadiusLG,
	})
	panel.SetPos(0, 0)
	panel.SetSize(1600, 900)
	engine.AddWidget(panel)

	// 标题
	title := widget.NewText("GPU Test: DrawSelectionHighlight (选择高亮渲染)")
	title.SetPos(40, 20)
	title.SetSize(800, 40)
	title.SetFont(engine.Font())
	title.SetColor(tokens.Global.ColorText)
	engine.AddWidget(title)

	// 测试标签
	label1 := widget.NewText("1. 部分文本选择 (选择 'World')")
	label1.SetPos(40, 80)
	label1.SetSize(400, 20)
	label1.SetFont(engine.Font())
	label1.SetColor(tokens.Global.ColorTextSecondary)
	engine.AddWidget(label1)

	label2 := widget.NewText("2. 全部文本选择 (选择 'Hello World')")
	label2.SetPos(40, 240)
	label2.SetSize(400, 20)
	label2.SetFont(engine.Font())
	label2.SetColor(tokens.Global.ColorTextSecondary)
	engine.AddWidget(label2)

	label3 := widget.NewText("3. 不同透明度选择 (alpha=0.1,0.2,0.3,0.5)")
	label3.SetPos(40, 400)
	label3.SetSize(400, 20)
	label3.SetFont(engine.Font())
	label3.SetColor(tokens.Global.ColorTextSecondary)
	engine.AddWidget(label3)

	label4 := widget.NewText("4. 不同颜色选择 (蓝/绿/红/黄)")
	label4.SetPos(40, 550)
	label4.SetSize(400, 20)
	label4.SetFont(engine.Font())
	label4.SetColor(tokens.Global.ColorTextSecondary)
	engine.AddWidget(label4)

	// 添加自定义渲染控件
	testWidget := NewSelectionHighlightTestWidget()
	testWidget.SetPos(0, 0)
	testWidget.SetSize(1600, 900)
	engine.AddWidget(testWidget)

	// 预期效果说明
	expected := widget.NewText(`预期效果:
1. 部分选择: 只有 'World' 部分有高亮背景，'Hello ' 部分无高亮，边界清晰
2. 全部选择: 整个 'Hello World' 都有高亮背景，覆盖完整无遗漏
3. 不同透明度: 4个选择区域，透明度分别为0.1,0.2,0.3,0.5，透明度差异明显可见
4. 不同颜色: 4个选择区域，颜色分别为蓝/绿/红/黄，颜色清晰可辨，无混色
验证标准: 选择高亮为半透明矩形，覆盖选中文本区域，不影响文本可读性`)
	expected.SetPos(40, 650)
	expected.SetSize(1520, 230)
	expected.SetFont(engine.Font())
	expected.SetColor(tokens.Global.ColorText)
	engine.AddWidget(expected)

	// 渲染完成后保存
	go func() {
		time.Sleep(3 * time.Second)
		lcl.RunOnMainThreadAsync(func(id uint32) {
			imgPath := filepath.Join(outDir, "gpu_selection_highlight.png")
			engine.Renderer().SavePNG(imgPath)
			fmt.Printf("✓ 图片已保存: %s\n", imgPath)
		})
	}()

	fmt.Println("✓ GPU测试程序已启动: DrawSelectionHighlight")
}
