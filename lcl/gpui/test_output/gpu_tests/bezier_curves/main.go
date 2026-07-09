// GPU Test: Bezier Curves - 贝塞尔曲线渲染验证
package main

import (
	"fmt"
	stdmath "math"
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

// BezierTestWidget 自定义控件，用于渲染贝塞尔曲线测试
type BezierTestWidget struct {
	widget.BaseWidget
}

func NewBezierTestWidget() *BezierTestWidget {
	w := &BezierTestWidget{
		BaseWidget: widget.NewBaseWidget(),
	}
	w.SetOwner(w)
	return w
}

// Render 渲染贝塞尔曲线测试内容
func (w *BezierTestWidget) Render(ctx *widget.Context) {
	if w == nil || ctx == nil || ctx.Renderer == nil {
		return
	}
	bounds := w.Bounds()
	ctx.Renderer.PushClip(bounds)

	// 测试1: 二次贝塞尔曲线
	path1 := pipeline.NewPath()
	path1.MoveTo(100, 200)
	path1.QuadTo(300, 100, 500, 200)
	ctx.Renderer.StrokePath(path1, 3, math.NewColor(0, 0, 1, 1))

	// 测试2: 三次贝塞尔曲线
	path2 := pipeline.NewPath()
	path2.MoveTo(100, 350)
	path2.CubicTo(200, 250, 400, 450, 500, 350)
	ctx.Renderer.StrokePath(path2, 3, math.NewColor(1, 0, 0, 1))

	// 测试3: 复杂贝塞尔曲线（心形）
	path3 := pipeline.NewPath()
	path3.MoveTo(800, 200)
	path3.CubicTo(800, 150, 750, 100, 700, 100)
	path3.CubicTo(650, 100, 600, 150, 600, 200)
	path3.CubicTo(600, 300, 800, 400, 800, 400)
	path3.CubicTo(800, 400, 1000, 300, 1000, 200)
	path3.CubicTo(1000, 150, 950, 100, 900, 100)
	path3.CubicTo(850, 100, 800, 150, 800, 200)
	path3.Close()
	ctx.Renderer.FillPath(path3, math.NewColor(1, 0.2, 0.4, 1))

	// 测试4: 螺旋线 - 从中心开始螺旋
	spiralCx := float32(1200)
	spiralCy := float32(300)
	path4 := pipeline.NewPath()
	// 从中心开始，初始半径为0
	path4.MoveTo(spiralCx, spiralCy)
	for i := 1; i <= 720; i++ {
		t := float64(i) * stdmath.Pi / 180
		r := float64(i) * 0.15 // 半径从0开始增长
		x := spiralCx + float32(r*stdmath.Cos(t))
		y := spiralCy + float32(r*stdmath.Sin(t))
		path4.LineTo(x, y)
	}
	ctx.Renderer.StrokePath(path4, 2, math.NewColor(0, 1, 0, 1))

	ctx.Renderer.PopClip()
}

func main() {
	libname.UseWS = "gtk3"
	if ws := os.Getenv("GPUI_WS"); ws != "" {
		libname.UseWS = ws
	}
	app := ui.NewApplication("GPU Test: Bezier Curves", 1600, 900)

	app.OnSetup(func(engine *ui.Engine) {
		setupUI(engine)
	})

	app.Run()
}

func setupUI(engine *ui.Engine) {
	tokens := token.Current()
	outDir := os.Getenv("GPUI_TEST_OUTPUT")
	if outDir == "" {
		outDir = "lcl/gpui/test_output/gpu_tests/bezier_curves"
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
	title := widget.NewText("GPU Test: Bezier Curves (贝塞尔曲线渲染)")
	title.SetPos(40, 20)
	title.SetSize(800, 40)
	title.SetFont(engine.Font())
	title.SetColor(tokens.Global.ColorText)
	engine.AddWidget(title)

	// 测试标签
	label1 := widget.NewText("1. 二次贝塞尔曲线 (QuadTo)")
	label1.SetPos(40, 80)
	label1.SetSize(400, 20)
	label1.SetFont(engine.Font())
	label1.SetColor(tokens.Global.ColorTextSecondary)
	engine.AddWidget(label1)

	label2 := widget.NewText("2. 三次贝塞尔曲线 (CubicTo)")
	label2.SetPos(40, 250)
	label2.SetSize(400, 20)
	label2.SetFont(engine.Font())
	label2.SetColor(tokens.Global.ColorTextSecondary)
	engine.AddWidget(label2)

	label3 := widget.NewText("3. 复杂贝塞尔曲线 (心形)")
	label3.SetPos(600, 80)
	label3.SetSize(400, 20)
	label3.SetFont(engine.Font())
	label3.SetColor(tokens.Global.ColorTextSecondary)
	engine.AddWidget(label3)

	label4 := widget.NewText("4. 螺旋线 (连续线段)")
	label4.SetPos(1100, 250)
	label4.SetSize(400, 20)
	label4.SetFont(engine.Font())
	label4.SetColor(tokens.Global.ColorTextSecondary)
	engine.AddWidget(label4)

	// 添加自定义渲染控件
	testWidget := NewBezierTestWidget()
	testWidget.SetPos(0, 0)
	testWidget.SetSize(1600, 900)
	engine.AddWidget(testWidget)

	// 预期效果说明
	expected := widget.NewText(`预期效果:
1. 二次贝塞尔曲线: 平滑的抛物线形状，起点和终点正确，控制点影响曲线弯曲
2. 三次贝塞尔曲线: S形曲线，两个控制点分别影响两端的弯曲方向
3. 心形曲线: 由多条三次贝塞尔曲线组成的心形，填充红色，边缘平滑
4. 螺旋线: 从中心向外旋转的螺旋，线条连续无断裂
验证标准: 所有曲线边缘平滑无锯齿，控制点影响正确，填充区域完整`)
	expected.SetPos(40, 650)
	expected.SetSize(1520, 230)
	expected.SetFont(engine.Font())
	expected.SetColor(tokens.Global.ColorText)
	engine.AddWidget(expected)

	// 渲染完成后保存
	go func() {
		time.Sleep(3 * time.Second)
		lcl.RunOnMainThreadAsync(func(id uint32) {
			imgPath := filepath.Join(outDir, "gpu_bezier_curves.png")
			engine.Renderer().SavePNG(imgPath)
			fmt.Printf("✓ 图片已保存: %s\n", imgPath)

			txtPath := filepath.Join(outDir, "gpu_bezier_curves.txt")
			f, _ := os.Create(txtPath)
			if f != nil {
				f.WriteString("GPU Test: Bezier Curves\n")
				f.WriteString("=======================\n\n")
				f.WriteString("渲染内容:\n")
				f.WriteString("1. 二次贝塞尔曲线: (100,200) -> (300,100) -> (500,200)\n")
				f.WriteString("2. 三次贝塞尔曲线: (100,350) -> (200,250) -> (400,450) -> (500,350)\n")
				f.WriteString("3. 心形曲线: 由6条三次贝塞尔曲线组成，填充红色\n")
				f.WriteString("4. 螺旋线: 从(1200,300)开始，360度旋转\n\n")
				f.WriteString("验证标准:\n")
				f.WriteString("- 曲线边缘平滑无锯齿\n")
				f.WriteString("- 控制点影响正确\n")
				f.WriteString("- 填充区域完整\n")
				f.WriteString("- 线条连续无断裂\n")
				f.Close()
				fmt.Printf("✓ 文本数据已保存: %s\n", txtPath)
			}
		})
	}()

	fmt.Println("✓ GPU测试程序已启动: Bezier Curves")
}
