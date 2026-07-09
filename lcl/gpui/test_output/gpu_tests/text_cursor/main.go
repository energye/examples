// GPU Test: DrawTextCursor - 文本光标渲染验证
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

// TextCursorTestWidget 自定义控件，用于渲染文本光标测试
type TextCursorTestWidget struct {
	widget.BaseWidget
}

func NewTextCursorTestWidget() *TextCursorTestWidget {
	w := &TextCursorTestWidget{
		BaseWidget: widget.NewBaseWidget(),
	}
	w.SetOwner(w)
	return w
}

// Render 渲染文本光标测试内容
func (w *TextCursorTestWidget) Render(ctx *widget.Context) {
	if w == nil || ctx == nil || ctx.Renderer == nil || ctx.Font == nil {
		return
	}
	bounds := w.Bounds()
	ctx.Renderer.PushClip(bounds)

	f := ctx.Font

	// 测试1: 不同位置光标
	text := "Hello World"
	textX := float32(100)
	textY := float32(120)
	ctx.Renderer.DrawText(text, textX, textY, f, math.NewColor(0, 0, 0, 1))
	// 开头光标
	ctx.Renderer.DrawTextCursor(textX, textY, f.LineHeight(), 2, math.NewColor(0, 0, 0, 1))
	// 中间光标 (在 "Hello " 之后)
	midWidth := f.TextWidth("Hello ")
	ctx.Renderer.DrawTextCursor(textX+midWidth, textY, f.LineHeight(), 2, math.NewColor(0, 0, 0, 1))
	// 结尾光标
	endWidth := f.TextWidth(text)
	ctx.Renderer.DrawTextCursor(textX+endWidth, textY, f.LineHeight(), 2, math.NewColor(0, 0, 0, 1))

	// 测试2: 不同高度光标
	text2 := "Height Test"
	text2X := float32(100)
	text2Y := float32(280)
	ctx.Renderer.DrawText(text2, text2X, text2Y, f, math.NewColor(0, 0, 0, 0.5))
	ctx.Renderer.DrawTextCursor(text2X, text2Y, 16, 2, math.NewColor(0, 0, 0, 1))
	ctx.Renderer.DrawTextCursor(text2X+200, text2Y, 24, 2, math.NewColor(0, 0, 0, 1))
	ctx.Renderer.DrawTextCursor(text2X+400, text2Y, 32, 2, math.NewColor(0, 0, 0, 1))
	ctx.Renderer.DrawTextCursor(text2X+600, text2Y, 48, 2, math.NewColor(0, 0, 0, 1))

	// 测试3: 不同宽度光标
	text3 := "Width Test"
	text3X := float32(100)
	text3Y := float32(430)
	ctx.Renderer.DrawText(text3, text3X, text3Y, f, math.NewColor(0, 0, 0, 0.5))
	ctx.Renderer.DrawTextCursor(text3X, text3Y, f.LineHeight(), 1, math.NewColor(0, 0, 0, 1))
	ctx.Renderer.DrawTextCursor(text3X+200, text3Y, f.LineHeight(), 2, math.NewColor(0, 0, 0, 1))
	ctx.Renderer.DrawTextCursor(text3X+400, text3Y, f.LineHeight(), 3, math.NewColor(0, 0, 0, 1))
	ctx.Renderer.DrawTextCursor(text3X+600, text3Y, f.LineHeight(), 4, math.NewColor(0, 0, 0, 1))

	// 测试4: 不同颜色光标
	text4 := "Color Test"
	text4X := float32(100)
	text4Y := float32(580)
	ctx.Renderer.DrawText(text4, text4X, text4Y, f, math.NewColor(0, 0, 0, 0.5))
	ctx.Renderer.DrawTextCursor(text4X, text4Y, f.LineHeight(), 2, math.NewColor(0, 0, 0, 1))   // 黑
	ctx.Renderer.DrawTextCursor(text4X+200, text4Y, f.LineHeight(), 2, math.NewColor(1, 0, 0, 1)) // 红
	ctx.Renderer.DrawTextCursor(text4X+400, text4Y, f.LineHeight(), 2, math.NewColor(0, 0, 1, 1)) // 蓝
	ctx.Renderer.DrawTextCursor(text4X+600, text4Y, f.LineHeight(), 2, math.NewColor(0, 1, 0, 1)) // 绿

	ctx.Renderer.PopClip()
}

func main() {
	libname.UseWS = "gtk3"
	if ws := os.Getenv("GPUI_WS"); ws != "" {
		libname.UseWS = ws
	}
	app := ui.NewApplication("GPU Test: DrawTextCursor", 1600, 900)

	app.OnSetup(func(engine *ui.Engine) {
		setupUI(engine)
	})

	app.Run()
}

func setupUI(engine *ui.Engine) {
	tokens := token.Current()
	outDir := os.Getenv("GPUI_TEST_OUTPUT")
	if outDir == "" {
		outDir = "lcl/gpui/test_output/gpu_tests/text_cursor"
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
	title := widget.NewText("GPU Test: DrawTextCursor (文本光标渲染)")
	title.SetPos(40, 20)
	title.SetSize(800, 40)
	title.SetFont(engine.Font())
	title.SetColor(tokens.Global.ColorText)
	engine.AddWidget(title)

	// 测试标签
	label1 := widget.NewText("1. 不同位置光标 (文本开头/中间/结尾)")
	label1.SetPos(40, 80)
	label1.SetSize(400, 20)
	label1.SetFont(engine.Font())
	label1.SetColor(tokens.Global.ColorTextSecondary)
	engine.AddWidget(label1)

	label2 := widget.NewText("2. 不同高度光标 (height=16,24,32,48)")
	label2.SetPos(40, 240)
	label2.SetSize(400, 20)
	label2.SetFont(engine.Font())
	label2.SetColor(tokens.Global.ColorTextSecondary)
	engine.AddWidget(label2)

	label3 := widget.NewText("3. 不同宽度光标 (width=1,2,3,4)")
	label3.SetPos(40, 400)
	label3.SetSize(400, 20)
	label3.SetFont(engine.Font())
	label3.SetColor(tokens.Global.ColorTextSecondary)
	engine.AddWidget(label3)

	label4 := widget.NewText("4. 不同颜色光标 (黑/红/蓝/绿)")
	label4.SetPos(40, 550)
	label4.SetSize(400, 20)
	label4.SetFont(engine.Font())
	label4.SetColor(tokens.Global.ColorTextSecondary)
	engine.AddWidget(label4)

	// 添加自定义渲染控件
	testWidget := NewTextCursorTestWidget()
	testWidget.SetPos(0, 0)
	testWidget.SetSize(1600, 900)
	engine.AddWidget(testWidget)

	// 预期效果说明
	expected := widget.NewText(`预期效果:
1. 不同位置: 光标位于文本开头/中间/结尾，位置准确无偏移
2. 不同高度: 4个光标，高度分别为16,24,32,48像素，高度差异明显可见
3. 不同宽度: 4个光标，宽度分别为1,2,3,4像素，宽度差异明显可见
4. 不同颜色: 4个光标，颜色分别为黑/红/蓝/绿，颜色清晰可辨
验证标准: 所有光标为垂直线条，位置/大小/颜色符合预期，边缘清晰无模糊`)
	expected.SetPos(40, 650)
	expected.SetSize(1520, 230)
	expected.SetFont(engine.Font())
	expected.SetColor(tokens.Global.ColorText)
	engine.AddWidget(expected)

	// 渲染完成后保存
	go func() {
		time.Sleep(3 * time.Second)

		lcl.RunOnMainThreadAsync(func(id uint32) {
			imgPath := filepath.Join(outDir, "gpu_text_cursor.png")
			engine.Renderer().SavePNG(imgPath)
			fmt.Printf("✓ 图片已保存: %s\n", imgPath)
		})
	}()

	fmt.Println("✓ GPU测试程序已启动: DrawTextCursor")
}
