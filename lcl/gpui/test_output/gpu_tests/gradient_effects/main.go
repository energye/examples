// GPU Test: Gradient Effects - 渐变效果渲染验证
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/energye/lcl/api/libname"
	lcl "github.com/energye/lcl/lcl"

	"github.com/energye/examples/lcl/gpui/core/math"
	"github.com/energye/examples/lcl/gpui/render/pipeline"
	"github.com/energye/examples/lcl/gpui/style/token"
	"github.com/energye/examples/lcl/gpui/ui"
	"github.com/energye/examples/lcl/gpui/widget"
)

// GradientTestWidget 自定义控件，用于渲染渐变效果测试
type GradientTestWidget struct {
	widget.BaseWidget
}

func NewGradientTestWidget() *GradientTestWidget {
	w := &GradientTestWidget{
		BaseWidget: widget.NewBaseWidget(),
	}
	w.SetOwner(w)
	return w
}

// Render 渲染渐变效果测试内容
func (w *GradientTestWidget) Render(ctx *widget.Context) {
	if w == nil || ctx == nil || ctx.Renderer == nil {
		return
	}
	bounds := w.Bounds()
	ctx.Renderer.PushClip(bounds)

	// 测试1: 水平渐变
	ctx.Renderer.FillLinearGradient(
		math.NewRect(100, 100, 400, 100),
		math.NewVec2(100, 100), math.NewVec2(500, 100),
		math.NewColor(1, 0, 0, 1), math.NewColor(0, 0, 1, 1),
	)

	// 测试2: 垂直渐变
	ctx.Renderer.FillLinearGradient(
		math.NewRect(100, 250, 400, 100),
		math.NewVec2(100, 250), math.NewVec2(100, 350),
		math.NewColor(0, 1, 0, 1), math.NewColor(1, 1, 0, 1),
	)

	// 测试3: 对角渐变
	ctx.Renderer.FillLinearGradient(
		math.NewRect(100, 400, 400, 100),
		math.NewVec2(100, 400), math.NewVec2(500, 500),
		math.NewColor(0.5, 0, 1, 1), math.NewColor(1, 0.5, 0, 1),
	)

	// 测试4: 圆角渐变
	ctx.Renderer.FillRoundLinearGradient(
		math.NewRect(600, 100, 400, 100), 20,
		math.NewVec2(600, 100), math.NewVec2(1000, 100),
		math.NewColor(1, 0, 0.5, 1), math.NewColor(0, 1, 0.5, 1),
	)

	// 测试5: 多色渐变（模拟彩虹）
	ctx.Renderer.FillLinearGradient(
		math.NewRect(600, 250, 400, 100),
		math.NewVec2(600, 250), math.NewVec2(1000, 250),
		math.NewColor(1, 0, 0, 1), math.NewColor(0, 0, 1, 1),
	)

	// 测试6: 透明度渐变
	ctx.Renderer.FillLinearGradient(
		math.NewRect(600, 400, 400, 100),
		math.NewVec2(600, 400), math.NewVec2(1000, 400),
		math.NewColor(1, 0, 0, 0), math.NewColor(1, 0, 0, 1),
	)

	// 测试7: 径向渐变模拟（使用多个同心圆）
	for i := 0; i < 10; i++ {
		r := float32(100 - i*10)
		alpha := float32(i) / 10.0
		ctx.Renderer.FillCircle(math.NewVec2(1200, 300), r, math.NewColor(0, 0, 1, alpha))
	}

	ctx.Renderer.PopClip()
}

func main() {
	libname.UseWS = "gtk3"
	if ws := os.Getenv("GPUI_WS"); ws != "" {
		libname.UseWS = ws
	}
	app := ui.NewApplication("GPU Test: Gradient Effects", 1600, 900)

	app.OnSetup(func(engine *ui.Engine) {
		setupUI(engine)
	})

	app.Run()
}

func setupUI(engine *ui.Engine) {
	tokens := token.Current()
	outDir := os.Getenv("GPUI_TEST_OUTPUT")
	if outDir == "" {
		outDir = "lcl/gpui/test_output/gpu_tests/gradient_effects"
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
	title := widget.NewText("GPU Test: Gradient Effects (渐变效果渲染)")
	title.SetPos(40, 20)
	title.SetSize(800, 40)
	title.SetFont(engine.Font())
	title.SetColor(tokens.Global.ColorText)
	engine.AddWidget(title)

	// 测试标签
	label1 := widget.NewText("1. 水平渐变 (红→蓝)")
	label1.SetPos(40, 80)
	label1.SetSize(400, 20)
	label1.SetFont(engine.Font())
	label1.SetColor(tokens.Global.ColorTextSecondary)
	engine.AddWidget(label1)

	label2 := widget.NewText("2. 垂直渐变 (绿→黄)")
	label2.SetPos(40, 230)
	label2.SetSize(400, 20)
	label2.SetFont(engine.Font())
	label2.SetColor(tokens.Global.ColorTextSecondary)
	engine.AddWidget(label2)

	label3 := widget.NewText("3. 对角渐变 (紫→橙)")
	label3.SetPos(40, 380)
	label3.SetSize(400, 20)
	label3.SetFont(engine.Font())
	label3.SetColor(tokens.Global.ColorTextSecondary)
	engine.AddWidget(label3)

	label4 := widget.NewText("4. 圆角渐变")
	label4.SetPos(600, 80)
	label4.SetSize(400, 20)
	label4.SetFont(engine.Font())
	label4.SetColor(tokens.Global.ColorTextSecondary)
	engine.AddWidget(label4)

	label5 := widget.NewText("5. 彩虹渐变")
	label5.SetPos(600, 230)
	label5.SetSize(400, 20)
	label5.SetFont(engine.Font())
	label5.SetColor(tokens.Global.ColorTextSecondary)
	engine.AddWidget(label5)

	label6 := widget.NewText("6. 透明度渐变")
	label6.SetPos(600, 380)
	label6.SetSize(400, 20)
	label6.SetFont(engine.Font())
	label6.SetColor(tokens.Global.ColorTextSecondary)
	engine.AddWidget(label6)

	label7 := widget.NewText("7. 径向渐变模拟")
	label7.SetPos(1100, 230)
	label7.SetSize(400, 20)
	label7.SetFont(engine.Font())
	label7.SetColor(tokens.Global.ColorTextSecondary)
	engine.AddWidget(label7)

	// 添加自定义渲染控件
	testWidget := NewGradientTestWidget()
	testWidget.SetPos(0, 0)
	testWidget.SetSize(1600, 900)
	engine.AddWidget(testWidget)

	// 预期效果说明
	expected := widget.NewText(`预期效果:
1. 水平渐变: 从左到右红到蓝的平滑过渡，无色带
2. 垂直渐变: 从上到下绿到黄的平滑过渡，无色带
3. 对角渐变: 从左上到右下紫到橙的平滑过渡，方向正确
4. 圆角渐变: 带圆角的水平渐变，圆角处无断裂
5. 彩虹渐变: 红到蓝的平滑过渡
6. 透明度渐变: 从完全透明到完全不透明的平滑过渡
7. 径向渐变模拟: 同心圆叠加模拟径向渐变效果
验证标准: 所有渐变过渡平滑无色带，方向正确，圆角处无断裂`)
	expected.SetPos(40, 600)
	expected.SetSize(1520, 280)
	expected.SetFont(engine.Font())
	expected.SetColor(tokens.Global.ColorText)
	engine.AddWidget(expected)

	// 渲染完成后保存
	go func() {
		time.Sleep(3 * time.Second)
		lcl.RunOnMainThreadAsync(func(id uint32) {
			imgPath := filepath.Join(outDir, "gpu_gradient_effects.png")
			engine.Renderer().SavePNG(imgPath)
			fmt.Printf("✓ 图片已保存: %s\n", imgPath)

			txtPath := filepath.Join(outDir, "gpu_gradient_effects.txt")
			f, _ := os.Create(txtPath)
			if f != nil {
				f.WriteString("GPU Test: Gradient Effects\n")
				f.WriteString("==========================\n\n")
				f.WriteString("渲染内容:\n")
				f.WriteString("1. 水平渐变: (100,100) -> (500,100), 红→蓝\n")
				f.WriteString("2. 垂直渐变: (100,250) -> (100,350), 绿→黄\n")
				f.WriteString("3. 对角渐变: (100,400) -> (500,500), 紫→橙\n")
				f.WriteString("4. 圆角渐变: (600,100) -> (1000,100), radius=20\n")
				f.WriteString("5. 彩虹渐变: (600,250) -> (1000,250), 红→蓝\n")
				f.WriteString("6. 透明度渐变: (600,400) -> (1000,400), 透明→不透明\n")
				f.WriteString("7. 径向渐变模拟: 中心(1200,300), 10层同心圆\n\n")
				f.WriteString("验证标准:\n")
				f.WriteString("- 渐变过渡平滑无色带\n")
				f.WriteString("- 方向正确\n")
				f.WriteString("- 圆角处无断裂\n")
				f.WriteString("- 透明度渐变正确\n")
				f.Close()
				fmt.Printf("✓ 文本数据已保存: %s\n", txtPath)
			}
		})
	}()

	fmt.Println("✓ GPU测试程序已启动: Gradient Effects")
}
