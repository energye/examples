// GPU Test: DrawStrikethrough - 删除线渲染验证
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

// StrikethroughTestWidget 自定义控件，用于渲染删除线测试
type StrikethroughTestWidget struct {
	widget.BaseWidget
}

func NewStrikethroughTestWidget() *StrikethroughTestWidget {
	w := &StrikethroughTestWidget{
		BaseWidget: widget.NewBaseWidget(),
	}
	w.SetOwner(w)
	return w
}

// Render 渲染删除线测试内容
func (w *StrikethroughTestWidget) Render(ctx *widget.Context) {
	if w == nil || ctx == nil || ctx.Renderer == nil {
		return
	}
	bounds := w.Bounds()
	ctx.Renderer.PushClip(bounds)

	// 测试1: 文本删除线
	ctx.Renderer.DrawText("Hello World", 100, 120, ctx.Font, math.NewColor(0, 0, 0, 1))
	ctx.Renderer.DrawStrikethrough(100, 132, 250, 2, math.NewColor(0, 0, 0, 1))

	// 测试2: 不同厚度删除线
	ctx.Renderer.DrawStrikethrough(100, 280, 400, 1, math.NewColor(0, 0, 0, 1))
	ctx.Renderer.DrawStrikethrough(100, 300, 400, 2, math.NewColor(0, 0, 0, 1))
	ctx.Renderer.DrawStrikethrough(100, 320, 400, 3, math.NewColor(0, 0, 0, 1))
	ctx.Renderer.DrawStrikethrough(100, 340, 400, 4, math.NewColor(0, 0, 0, 1))

	// 测试3: 不同颜色删除线
	ctx.Renderer.DrawStrikethrough(100, 430, 200, 2, math.NewColor(0, 0, 0, 1)) // 黑
	ctx.Renderer.DrawStrikethrough(350, 430, 200, 2, math.NewColor(1, 0, 0, 1)) // 红
	ctx.Renderer.DrawStrikethrough(600, 430, 200, 2, math.NewColor(0, 0, 1, 1)) // 蓝
	ctx.Renderer.DrawStrikethrough(850, 430, 200, 2, math.NewColor(0, 1, 0, 1)) // 绿

	// 测试4: 不同长度删除线
	ctx.Renderer.DrawStrikethrough(100, 580, 50, 2, math.NewColor(0, 0, 0, 1))
	ctx.Renderer.DrawStrikethrough(200, 580, 100, 2, math.NewColor(0, 0, 0, 1))
	ctx.Renderer.DrawStrikethrough(350, 580, 200, 2, math.NewColor(0, 0, 0, 1))
	ctx.Renderer.DrawStrikethrough(600, 580, 400, 2, math.NewColor(0, 0, 0, 1))

	ctx.Renderer.PopClip()
}

func main() {
	libname.UseWS = "gtk3"
	if ws := os.Getenv("GPUI_WS"); ws != "" {
		libname.UseWS = ws
	}
	app := ui.NewApplication("GPU Test: DrawStrikethrough", 1600, 900)

	app.OnSetup(func(engine *ui.Engine) {
		setupUI(engine)
	})

	app.Run()
}

func setupUI(engine *ui.Engine) {
	tokens := token.Current()
	outDir := os.Getenv("GPUI_TEST_OUTPUT")
	if outDir == "" {
		outDir = "lcl/gpui/test_output/gpu_tests/strikethrough"
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
	title := widget.NewText("GPU Test: DrawStrikethrough (删除线渲染)")
	title.SetPos(40, 20)
	title.SetSize(800, 40)
	title.SetFont(engine.Font())
	title.SetColor(tokens.Global.ColorText)
	engine.AddWidget(title)

	// 测试标签
	label1 := widget.NewText("1. 文本删除线 (Hello World)")
	label1.SetPos(40, 80)
	label1.SetSize(400, 20)
	label1.SetFont(engine.Font())
	label1.SetColor(tokens.Global.ColorTextSecondary)
	engine.AddWidget(label1)

	label2 := widget.NewText("2. 不同厚度删除线 (thickness=1,2,3,4)")
	label2.SetPos(40, 240)
	label2.SetSize(400, 20)
	label2.SetFont(engine.Font())
	label2.SetColor(tokens.Global.ColorTextSecondary)
	engine.AddWidget(label2)

	label3 := widget.NewText("3. 不同颜色删除线 (黑/红/蓝/绿)")
	label3.SetPos(40, 400)
	label3.SetSize(400, 20)
	label3.SetFont(engine.Font())
	label3.SetColor(tokens.Global.ColorTextSecondary)
	engine.AddWidget(label3)

	label4 := widget.NewText("4. 不同长度删除线 (50,100,200,400)")
	label4.SetPos(40, 550)
	label4.SetSize(400, 20)
	label4.SetFont(engine.Font())
	label4.SetColor(tokens.Global.ColorTextSecondary)
	engine.AddWidget(label4)

	// 添加自定义渲染控件
	testWidget := NewStrikethroughTestWidget()
	testWidget.SetPos(0, 0)
	testWidget.SetSize(1600, 900)
	engine.AddWidget(testWidget)

	// 预期效果说明
	expected := widget.NewText(`预期效果:
1. 文本删除线: 删除线位于文本中间垂直位置，穿过文本，无偏移
2. 不同厚度: 4条删除线，厚度分别为1,2,3,4像素，厚度差异明显可见
3. 不同颜色: 4条删除线，颜色分别为黑/红/蓝/绿，颜色清晰可辨，无混色
4. 不同长度: 4条删除线，长度分别为50,100,200,400像素，长度差异明显可见
验证标准: 所有删除线为水平线条，位于文本中间垂直位置，厚度/颜色/长度符合预期`)
	expected.SetPos(40, 650)
	expected.SetSize(1520, 230)
	expected.SetFont(engine.Font())
	expected.SetColor(tokens.Global.ColorText)
	engine.AddWidget(expected)

	// 渲染完成后保存
	go func() {
		time.Sleep(3 * time.Second)

		lcl.RunOnMainThreadAsync(func(id uint32) {
			imgPath := filepath.Join(outDir, "gpu_strikethrough.png")
			engine.Renderer().SavePNG(imgPath)
			fmt.Printf("✓ 图片已保存: %s\n", imgPath)
		})
	}()

	fmt.Println("✓ GPU测试程序已启动: DrawStrikethrough")
}
