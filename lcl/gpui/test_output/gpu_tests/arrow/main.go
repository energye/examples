// GPU Test: DrawArrow - 箭头渲染验证
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

// ArrowTestWidget 自定义控件，用于渲染箭头测试
type ArrowTestWidget struct {
	widget.BaseWidget
}

func NewArrowTestWidget() *ArrowTestWidget {
	w := &ArrowTestWidget{
		BaseWidget: widget.NewBaseWidget(),
	}
	w.SetOwner(w)
	return w
}

// Render 渲染箭头测试内容
func (w *ArrowTestWidget) Render(ctx *widget.Context) {
	if w == nil || ctx == nil || ctx.Renderer == nil {
		return
	}
	bounds := w.Bounds()
	ctx.Renderer.PushClip(bounds)

	// 测试1: 四个方向箭头
	ctx.Renderer.DrawArrow(math.NewVec2(200, 180), 60, 0, math.NewColor(0, 0, 1, 1)) // 上
	ctx.Renderer.DrawArrow(math.NewVec2(400, 180), 60, 1, math.NewColor(0, 0, 1, 1)) // 右
	ctx.Renderer.DrawArrow(math.NewVec2(600, 180), 60, 2, math.NewColor(0, 0, 1, 1)) // 下
	ctx.Renderer.DrawArrow(math.NewVec2(800, 180), 60, 3, math.NewColor(0, 0, 1, 1)) // 左

	// 测试2: 不同大小箭头
	ctx.Renderer.DrawArrow(math.NewVec2(200, 400), 20, 2, math.NewColor(0, 0, 1, 1))
	ctx.Renderer.DrawArrow(math.NewVec2(400, 400), 40, 2, math.NewColor(0, 0, 1, 1))
	ctx.Renderer.DrawArrow(math.NewVec2(600, 400), 60, 2, math.NewColor(0, 0, 1, 1))
	ctx.Renderer.DrawArrow(math.NewVec2(800, 400), 80, 2, math.NewColor(0, 0, 1, 1))

	// 测试3: 不同颜色箭头
	ctx.Renderer.DrawArrow(math.NewVec2(200, 600), 60, 2, math.NewColor(1, 0, 0, 1))  // 红
	ctx.Renderer.DrawArrow(math.NewVec2(500, 600), 60, 2, math.NewColor(0, 1, 0, 1))  // 绿
	ctx.Renderer.DrawArrow(math.NewVec2(800, 600), 60, 2, math.NewColor(0, 0, 1, 1))  // 蓝
	ctx.Renderer.DrawArrow(math.NewVec2(1100, 600), 60, 2, math.NewColor(1, 1, 0, 1)) // 黄

	ctx.Renderer.PopClip()
}

func main() {
	libname.UseWS = "gtk3"
	if ws := os.Getenv("GPUI_WS"); ws != "" {
		libname.UseWS = ws
	}
	app := ui.NewApplication("GPU Test: DrawArrow", 1600, 900)

	app.OnSetup(func(engine *ui.Engine) {
		setupUI(engine)
	})

	app.Run()
}

func setupUI(engine *ui.Engine) {
	tokens := token.Current()
	outDir := os.Getenv("GPUI_TEST_OUTPUT")
	if outDir == "" {
		outDir = "lcl/gpui/test_output/gpu_tests/arrow"
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
	title := widget.NewText("GPU Test: DrawArrow (箭头渲染)")
	title.SetPos(40, 20)
	title.SetSize(800, 40)
	title.SetFont(engine.Font())
	title.SetColor(tokens.Global.ColorText)
	engine.AddWidget(title)

	// 测试标签
	label1 := widget.NewText("1. 四个方向箭头 (上/右/下/左)")
	label1.SetPos(40, 80)
	label1.SetSize(400, 20)
	label1.SetFont(engine.Font())
	label1.SetColor(tokens.Global.ColorTextSecondary)
	engine.AddWidget(label1)

	label2 := widget.NewText("2. 不同大小箭头 (size=20,40,60,80)")
	label2.SetPos(40, 300)
	label2.SetSize(400, 20)
	label2.SetFont(engine.Font())
	label2.SetColor(tokens.Global.ColorTextSecondary)
	engine.AddWidget(label2)

	label3 := widget.NewText("3. 不同颜色箭头 (红/绿/蓝/黄)")
	label3.SetPos(40, 520)
	label3.SetSize(400, 20)
	label3.SetFont(engine.Font())
	label3.SetColor(tokens.Global.ColorTextSecondary)
	engine.AddWidget(label3)

	// 添加自定义渲染控件
	testWidget := NewArrowTestWidget()
	testWidget.SetPos(0, 0)
	testWidget.SetSize(1600, 900)
	engine.AddWidget(testWidget)

	// 预期效果说明
	expected := widget.NewText(`预期效果:
1. 四个方向: 箭头分别指向上/右/下/左，方向正确，三角形完整无缺角
2. 不同大小: 4个箭头，大小分别为20,40,60,80像素，大小差异明显可见
3. 不同颜色: 4个箭头，颜色分别为红/绿/蓝/黄，颜色清晰可辨，无混色
验证标准: 所有箭头为等腰三角形，方向正确，大小和颜色符合预期，边缘平滑无锯齿`)
	expected.SetPos(40, 680)
	expected.SetSize(1520, 200)
	expected.SetFont(engine.Font())
	expected.SetColor(tokens.Global.ColorText)
	engine.AddWidget(expected)

	// 渲染完成后保存
	go func() {
		time.Sleep(3 * time.Second)
		lcl.RunOnMainThreadAsync(func(id uint32) {
			imgPath := filepath.Join(outDir, "gpu_arrow.png")
			engine.Renderer().SavePNG(imgPath)
			fmt.Printf("✓ 图片已保存: %s\n", imgPath)
		})
	}()

	fmt.Println("✓ GPU测试程序已启动: DrawArrow")
}
