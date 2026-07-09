// GPU Test: Complex Shapes - 复杂图形渲染验证
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

// ComplexShapesTestWidget 自定义控件，用于渲染复杂图形测试
type ComplexShapesTestWidget struct {
	widget.BaseWidget
}

func NewComplexShapesTestWidget() *ComplexShapesTestWidget {
	w := &ComplexShapesTestWidget{
		BaseWidget: widget.NewBaseWidget(),
	}
	w.SetOwner(w)
	return w
}

// Render 渲染复杂图形测试内容
func (w *ComplexShapesTestWidget) Render(ctx *widget.Context) {
	if w == nil || ctx == nil || ctx.Renderer == nil {
		return
	}
	bounds := w.Bounds()
	ctx.Renderer.PushClip(bounds)

	// 测试1: 正六边形
	hexPath := pipeline.NewPath()
	for i := 0; i < 6; i++ {
		angle := float64(i) * stdmath.Pi / 3
		x := 200 + 60*float32(stdmath.Cos(angle))
		y := 200 + 60*float32(stdmath.Sin(angle))
		if i == 0 {
			hexPath.MoveTo(x, y)
		} else {
			hexPath.LineTo(x, y)
		}
	}
	hexPath.Close()
	ctx.Renderer.FillPath(hexPath, math.NewColor(0.2, 0.6, 1, 1))
	ctx.Renderer.StrokePath(hexPath, 2, math.NewColor(0, 0, 0, 0.3))

	// 测试2: 五角星
	starPath := pipeline.NewPath()
	for i := 0; i < 10; i++ {
		angle := float64(i)*stdmath.Pi/5 - stdmath.Pi/2
		r := float32(60)
		if i%2 == 1 {
			r = 25
		}
		x := 450 + r*float32(stdmath.Cos(angle))
		y := 200 + r*float32(stdmath.Sin(angle))
		if i == 0 {
			starPath.MoveTo(x, y)
		} else {
			starPath.LineTo(x, y)
		}
	}
	starPath.Close()
	ctx.Renderer.FillPath(starPath, math.NewColor(1, 0.8, 0, 1))

	// 测试3: 菱形
	diamondPath := pipeline.NewPath()
	diamondPath.MoveTo(700, 140)
	diamondPath.LineTo(760, 200)
	diamondPath.LineTo(700, 260)
	diamondPath.LineTo(640, 200)
	diamondPath.Close()
	ctx.Renderer.FillPath(diamondPath, math.NewColor(0.8, 0.2, 0.8, 1))

	// 测试4: 圆环
	ctx.Renderer.FillCircle(math.NewVec2(900, 200), 60, math.NewColor(0.2, 0.8, 0.4, 1))
	ctx.Renderer.FillCircle(math.NewVec2(900, 200), 40, math.NewColor(1, 1, 1, 1))

	// 测试5: 十字形
	crossPath := pipeline.NewPath()
	crossPath.MoveTo(1100, 160)
	crossPath.LineTo(1120, 160)
	crossPath.LineTo(1120, 180)
	crossPath.LineTo(1140, 180)
	crossPath.LineTo(1140, 200)
	crossPath.LineTo(1120, 200)
	crossPath.LineTo(1120, 220)
	crossPath.LineTo(1100, 220)
	crossPath.LineTo(1100, 200)
	crossPath.LineTo(1080, 200)
	crossPath.LineTo(1080, 180)
	crossPath.LineTo(1100, 180)
	crossPath.Close()
	ctx.Renderer.FillPath(crossPath, math.NewColor(1, 0.3, 0.3, 1))

	// 测试6: 箭头形状
	arrowPath := pipeline.NewPath()
	arrowPath.MoveTo(1300, 200)
	arrowPath.LineTo(1350, 160)
	arrowPath.LineTo(1350, 180)
	arrowPath.LineTo(1400, 180)
	arrowPath.LineTo(1400, 220)
	arrowPath.LineTo(1350, 220)
	arrowPath.LineTo(1350, 240)
	arrowPath.Close()
	ctx.Renderer.FillPath(arrowPath, math.NewColor(0.2, 0.5, 1, 1))

	// 测试7: 齿轮形状
	gearPath := pipeline.NewPath()
	gearCx := float32(200)
	gearCy := float32(450)
	gearR := float32(60)
	teeth := 12
	for i := 0; i < teeth*2; i++ {
		angle := float64(i) * stdmath.Pi / float64(teeth)
		r := gearR
		if i%2 == 0 {
			r = gearR * 0.7
		}
		x := gearCx + r*float32(stdmath.Cos(angle))
		y := gearCy + r*float32(stdmath.Sin(angle))
		if i == 0 {
			gearPath.MoveTo(x, y)
		} else {
			gearPath.LineTo(x, y)
		}
	}
	gearPath.Close()
	ctx.Renderer.FillPath(gearPath, math.NewColor(0.5, 0.5, 0.5, 1))
	ctx.Renderer.FillCircle(math.NewVec2(gearCx, gearCy), 20, math.NewColor(0.8, 0.8, 0.8, 1))

	// 测试8: 心形
	heartPath := pipeline.NewPath()
	heartCx := float32(450)
	heartCy := float32(450)
	for i := 0; i <= 360; i++ {
		t := float64(i) * stdmath.Pi / 180
		x := heartCx + 40*float32(16*stdmath.Pow(stdmath.Sin(t), 3))/16
		y := heartCy - 40*float32(13*stdmath.Cos(t)-5*stdmath.Cos(2*t)-2*stdmath.Cos(3*t)-stdmath.Cos(4*t))/16
		if i == 0 {
			heartPath.MoveTo(x, y)
		} else {
			heartPath.LineTo(x, y)
		}
	}
	heartPath.Close()
	ctx.Renderer.FillPath(heartPath, math.NewColor(1, 0.2, 0.4, 1))

	// 测试9: 闪电形状
	lightningPath := pipeline.NewPath()
	lightningPath.MoveTo(700, 380)
	lightningPath.LineTo(720, 420)
	lightningPath.LineTo(710, 420)
	lightningPath.LineTo(730, 480)
	lightningPath.LineTo(710, 440)
	lightningPath.LineTo(720, 440)
	lightningPath.Close()
	ctx.Renderer.FillPath(lightningPath, math.NewColor(1, 0.9, 0, 1))

	// 测试10: 音符形状
	notePath := pipeline.NewPath()
	notePath.MoveTo(900, 380)
	notePath.LineTo(900, 460)
	notePath.CubicTo(900, 480, 920, 480, 920, 460)
	notePath.CubicTo(920, 440, 900, 440, 900, 460)
	notePath.MoveTo(900, 380)
	notePath.LineTo(940, 360)
	notePath.LineTo(940, 400)
	ctx.Renderer.StrokePath(notePath, 3, math.NewColor(0, 0, 0, 1))

	ctx.Renderer.PopClip()
}

func main() {
	libname.UseWS = "gtk3"
	if ws := os.Getenv("GPUI_WS"); ws != "" {
		libname.UseWS = ws
	}
	app := ui.NewApplication("GPU Test: Complex Shapes", 1600, 900)

	app.OnSetup(func(engine *ui.Engine) {
		setupUI(engine)
	})

	app.Run()
}

func setupUI(engine *ui.Engine) {
	tokens := token.Current()
	outDir := os.Getenv("GPUI_TEST_OUTPUT")
	if outDir == "" {
		outDir = "lcl/gpui/test_output/gpu_tests/complex_shapes"
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
	title := widget.NewText("GPU Test: Complex Shapes (复杂图形渲染)")
	title.SetPos(40, 20)
	title.SetSize(800, 40)
	title.SetFont(engine.Font())
	title.SetColor(tokens.Global.ColorText)
	engine.AddWidget(title)

	// 测试标签
	label1 := widget.NewText("1. 正六边形")
	label1.SetPos(40, 80)
	label1.SetSize(400, 20)
	label1.SetFont(engine.Font())
	label1.SetColor(tokens.Global.ColorTextSecondary)
	engine.AddWidget(label1)

	label2 := widget.NewText("2. 五角星")
	label2.SetPos(350, 80)
	label2.SetSize(400, 20)
	label2.SetFont(engine.Font())
	label2.SetColor(tokens.Global.ColorTextSecondary)
	engine.AddWidget(label2)

	label3 := widget.NewText("3. 菱形")
	label3.SetPos(600, 80)
	label3.SetSize(400, 20)
	label3.SetFont(engine.Font())
	label3.SetColor(tokens.Global.ColorTextSecondary)
	engine.AddWidget(label3)

	label4 := widget.NewText("4. 圆环")
	label4.SetPos(850, 80)
	label4.SetSize(400, 20)
	label4.SetFont(engine.Font())
	label4.SetColor(tokens.Global.ColorTextSecondary)
	engine.AddWidget(label4)

	label5 := widget.NewText("5. 十字形")
	label5.SetPos(1050, 80)
	label5.SetSize(400, 20)
	label5.SetFont(engine.Font())
	label5.SetColor(tokens.Global.ColorTextSecondary)
	engine.AddWidget(label5)

	label6 := widget.NewText("6. 箭头形状")
	label6.SetPos(1250, 80)
	label6.SetSize(400, 20)
	label6.SetFont(engine.Font())
	label6.SetColor(tokens.Global.ColorTextSecondary)
	engine.AddWidget(label6)

	label7 := widget.NewText("7. 齿轮形状")
	label7.SetPos(40, 350)
	label7.SetSize(400, 20)
	label7.SetFont(engine.Font())
	label7.SetColor(tokens.Global.ColorTextSecondary)
	engine.AddWidget(label7)

	label8 := widget.NewText("8. 心形")
	label8.SetPos(350, 350)
	label8.SetSize(400, 20)
	label8.SetFont(engine.Font())
	label8.SetColor(tokens.Global.ColorTextSecondary)
	engine.AddWidget(label8)

	label9 := widget.NewText("9. 闪电形状")
	label9.SetPos(600, 350)
	label9.SetSize(400, 20)
	label9.SetFont(engine.Font())
	label9.SetColor(tokens.Global.ColorTextSecondary)
	engine.AddWidget(label9)

	label10 := widget.NewText("10. 音符形状")
	label10.SetPos(850, 350)
	label10.SetSize(400, 20)
	label10.SetFont(engine.Font())
	label10.SetColor(tokens.Global.ColorTextSecondary)
	engine.AddWidget(label10)

	// 添加自定义渲染控件
	testWidget := NewComplexShapesTestWidget()
	testWidget.SetPos(0, 0)
	testWidget.SetSize(1600, 900)
	engine.AddWidget(testWidget)

	// 预期效果说明
	expected := widget.NewText(`预期效果:
1. 正六边形: 6条等长边，角度正确，填充和描边
2. 五角星: 5个角正确，填充金色
3. 菱形: 4条等长边，角度正确
4. 圆环: 同心圆叠加形成环形
5. 十字形: 对称的十字形状
6. 箭头形状: 指向右的箭头
7. 齿轮形状: 12个齿的齿轮，中心有圆孔
8. 心形: 数学公式生成的心形曲线
9. 闪电形状: 锯齿形的闪电
10. 音符形状: 音符轮廓
验证标准: 所有形状边缘平滑无锯齿，填充完整，形状正确`)
	expected.SetPos(40, 600)
	expected.SetSize(1520, 280)
	expected.SetFont(engine.Font())
	expected.SetColor(tokens.Global.ColorText)
	engine.AddWidget(expected)

	// 渲染完成后保存
	go func() {
		time.Sleep(3 * time.Second)
		lcl.RunOnMainThreadAsync(func(id uint32) {
			imgPath := filepath.Join(outDir, "gpu_complex_shapes.png")
			engine.Renderer().SavePNG(imgPath)
			fmt.Printf("✓ 图片已保存: %s\n", imgPath)

			txtPath := filepath.Join(outDir, "gpu_complex_shapes.txt")
			f, _ := os.Create(txtPath)
			if f != nil {
				f.WriteString("GPU Test: Complex Shapes\n")
				f.WriteString("========================\n\n")
				f.WriteString("渲染内容:\n")
				f.WriteString("1. 正六边形: 中心(200,200), 半径60\n")
				f.WriteString("2. 五角星: 中心(450,200), 外半径60, 内半径25\n")
				f.WriteString("3. 菱形: 中心(700,200), 60x60\n")
				f.WriteString("4. 圆环: 中心(900,200), 外半径60, 内半径40\n")
				f.WriteString("5. 十字形: 中心(1100,200), 60x60\n")
				f.WriteString("6. 箭头形状: 中心(1300,200), 100x80\n")
				f.WriteString("7. 齿轮形状: 中心(200,450), 半径60, 12齿\n")
				f.WriteString("8. 心形: 中心(450,450), 大小40\n")
				f.WriteString("9. 闪电形状: (700,380) -> (730,480)\n")
				f.WriteString("10. 音符形状: (900,380) -> (940,400)\n\n")
				f.WriteString("验证标准:\n")
				f.WriteString("- 所有形状边缘平滑无锯齿\n")
				f.WriteString("- 填充完整无遗漏\n")
				f.WriteString("- 形状正确无变形\n")
				f.Close()
				fmt.Printf("✓ 文本数据已保存: %s\n", txtPath)
			}
		})
	}()

	fmt.Println("✓ GPU测试程序已启动: Complex Shapes")
}
