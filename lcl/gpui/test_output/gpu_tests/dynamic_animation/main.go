// GPU Test: Dynamic Animation - 动态效果渲染验证
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

// DynamicAnimationTestWidget 自定义控件，用于渲染动态效果测试
type DynamicAnimationTestWidget struct {
	widget.BaseWidget
	startTime time.Time
}

func NewDynamicAnimationTestWidget() *DynamicAnimationTestWidget {
	w := &DynamicAnimationTestWidget{
		BaseWidget: widget.NewBaseWidget(),
		startTime:  time.Now(),
	}
	w.SetOwner(w)
	return w
}

// Render 渲染动态效果测试内容
func (w *DynamicAnimationTestWidget) Render(ctx *widget.Context) {
	if w == nil || ctx == nil || ctx.Renderer == nil {
		return
	}
	bounds := w.Bounds()
	ctx.Renderer.PushClip(bounds)

	elapsed := float64(time.Since(w.startTime).Milliseconds()) / 1000.0

	// 测试1: 旋转的正方形
	angle := elapsed * 2 // 2 radians per second
	cx, cy := float32(200), float32(200)
	size := float32(80)
	// 绘制旋转的正方形（使用4条线）
	for i := 0; i < 4; i++ {
		a1 := float64(i) * stdmath.Pi / 2
		a2 := float64(i+1) * stdmath.Pi / 2
		x1 := cx + size*float32(stdmath.Cos(a1+angle))
		y1 := cy + size*float32(stdmath.Sin(a1+angle))
		x2 := cx + size*float32(stdmath.Cos(a2+angle))
		y2 := cy + size*float32(stdmath.Sin(a2+angle))
		ctx.Renderer.DrawLine(x1, y1, x2, y2, 3, math.NewColor(0.2, 0.5, 1, 1))
	}

	// 测试2: 脉冲圆
	pulse := float32(0.5 + 0.5*stdmath.Sin(elapsed*3))
	radius := 30 + pulse*20
	ctx.Renderer.FillCircle(math.NewVec2(500, 200), radius, math.NewColor(1, 0.3, 0.3, 0.8))

	// 测试3: 移动的点
	x := float32(700 + 100*stdmath.Sin(elapsed*2))
	y := float32(200 + 50*stdmath.Cos(elapsed*3))
	ctx.Renderer.FillCircle(math.NewVec2(x, y), 10, math.NewColor(0, 1, 0, 1))

	// 测试4: 波浪线
	path := pipeline.NewPath()
	path.MoveTo(100, 400)
	for i := 0; i < 400; i++ {
		x := float32(100 + i)
		y := float32(400 + 30*stdmath.Sin(float64(i)*0.05+elapsed*3))
		path.LineTo(x, y)
	}
	ctx.Renderer.StrokePath(path, 2, math.NewColor(0.8, 0.2, 0.8, 1))

	// 测试5: 旋转的星形
	starCx := float32(1000)
	starCy := float32(200)
	starR := float32(60)
	for i := 0; i < 10; i++ {
		a1 := float64(i)*stdmath.Pi/5 + elapsed
		a2 := float64(i+1)*stdmath.Pi/5 + elapsed
		r1 := starR
		r2 := starR * 0.4
		if i%2 == 0 {
			r1, r2 = r2, r1
		}
		x1 := starCx + r1*float32(stdmath.Cos(a1))
		y1 := starCy + r1*float32(stdmath.Sin(a1))
		x2 := starCx + r2*float32(stdmath.Cos(a2))
		y2 := starCy + r2*float32(stdmath.Sin(a2))
		ctx.Renderer.DrawLine(x1, y1, x2, y2, 2, math.NewColor(1, 0.8, 0, 1))
	}

	// 测试6: 渐变色动画
	hue := float32(stdmath.Mod(elapsed*60, 360))
	r := float32(stdmath.Abs(stdmath.Sin(float64(hue) * stdmath.Pi / 180)))
	g := float32(stdmath.Abs(stdmath.Sin(float64(hue+120) * stdmath.Pi / 180)))
	b := float32(stdmath.Abs(stdmath.Sin(float64(hue+240) * stdmath.Pi / 180)))
	ctx.Renderer.FillRoundRect(math.NewRect(1200, 150, 200, 100), 16, math.NewColor(r, g, b, 1))

	ctx.Renderer.PopClip()

	// 触发重绘以维持动画
}

func main() {
	libname.UseWS = "gtk3"
	if ws := os.Getenv("GPUI_WS"); ws != "" {
		libname.UseWS = ws
	}
	app := ui.NewApplication("GPU Test: Dynamic Animation", 1600, 900)

	app.OnSetup(func(engine *ui.Engine) {
		setupUI(engine)
	})

	app.Run()
}

func setupUI(engine *ui.Engine) {
	tokens := token.Current()
	outDir := os.Getenv("GPUI_TEST_OUTPUT")
	if outDir == "" {
		outDir = "lcl/gpui/test_output/gpu_tests/dynamic_animation"
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
	title := widget.NewText("GPU Test: Dynamic Animation (动态效果渲染)")
	title.SetPos(40, 20)
	title.SetSize(800, 40)
	title.SetFont(engine.Font())
	title.SetColor(tokens.Global.ColorText)
	engine.AddWidget(title)

	// 测试标签
	label1 := widget.NewText("1. 旋转正方形")
	label1.SetPos(40, 80)
	label1.SetSize(400, 20)
	label1.SetFont(engine.Font())
	label1.SetColor(tokens.Global.ColorTextSecondary)
	engine.AddWidget(label1)

	label2 := widget.NewText("2. 脉冲圆")
	label2.SetPos(400, 80)
	label2.SetSize(400, 20)
	label2.SetFont(engine.Font())
	label2.SetColor(tokens.Global.ColorTextSecondary)
	engine.AddWidget(label2)

	label3 := widget.NewText("3. 移动的点")
	label3.SetPos(600, 80)
	label3.SetSize(400, 20)
	label3.SetFont(engine.Font())
	label3.SetColor(tokens.Global.ColorTextSecondary)
	engine.AddWidget(label3)

	label4 := widget.NewText("4. 波浪线")
	label4.SetPos(40, 350)
	label4.SetSize(400, 20)
	label4.SetFont(engine.Font())
	label4.SetColor(tokens.Global.ColorTextSecondary)
	engine.AddWidget(label4)

	label5 := widget.NewText("5. 旋转星形")
	label5.SetPos(900, 80)
	label5.SetSize(400, 20)
	label5.SetFont(engine.Font())
	label5.SetColor(tokens.Global.ColorTextSecondary)
	engine.AddWidget(label5)

	label6 := widget.NewText("6. 渐变色动画")
	label6.SetPos(1200, 130)
	label6.SetSize(400, 20)
	label6.SetFont(engine.Font())
	label6.SetColor(tokens.Global.ColorTextSecondary)
	engine.AddWidget(label6)

	// 添加自定义渲染控件
	testWidget := NewDynamicAnimationTestWidget()
	testWidget.SetPos(0, 0)
	testWidget.SetSize(1600, 900)
	engine.AddWidget(testWidget)

	// 预期效果说明
	expected := widget.NewText(`预期效果:
1. 旋转正方形: 正方形持续旋转，线条平滑无断裂
2. 脉冲圆: 圆形大小周期性变化，边缘平滑
3. 移动的点: 点沿椭圆轨迹移动，运动平滑
4. 波浪线: 正弦波持续波动，线条平滑无断裂
5. 旋转星形: 五角星持续旋转，线条平滑
6. 渐变色动画: 颜色持续变化，过渡平滑
验证标准: 所有动画流畅无卡顿，边缘平滑无锯齿，颜色过渡自然`)
	expected.SetPos(40, 600)
	expected.SetSize(1520, 280)
	expected.SetFont(engine.Font())
	expected.SetColor(tokens.Global.ColorText)
	engine.AddWidget(expected)

	// 渲染完成后保存
	go func() {
		time.Sleep(3 * time.Second)
		lcl.RunOnMainThreadAsync(func(id uint32) {
			imgPath := filepath.Join(outDir, "gpu_dynamic_animation.png")
			engine.Renderer().SavePNG(imgPath)
			fmt.Printf("✓ 图片已保存: %s\n", imgPath)

			txtPath := filepath.Join(outDir, "gpu_dynamic_animation.txt")
			f, _ := os.Create(txtPath)
			if f != nil {
				f.WriteString("GPU Test: Dynamic Animation\n")
				f.WriteString("===========================\n\n")
				f.WriteString("渲染内容:\n")
				f.WriteString("1. 旋转正方形: 中心(200,200), 大小80, 旋转速度2rad/s\n")
				f.WriteString("2. 脉冲圆: 中心(500,200), 半径30-50, 脉冲速度3Hz\n")
				f.WriteString("3. 移动的点: 椭圆轨迹(700,200), 半径100x50\n")
				f.WriteString("4. 波浪线: (100,400) -> (500,400), 振幅30, 频率0.05\n")
				f.WriteString("5. 旋转星形: 中心(1000,200), 半径60, 旋转速度1rad/s\n")
				f.WriteString("6. 渐变色动画: (1200,150) 200x100, 颜色循环\n\n")
				f.WriteString("验证标准:\n")
				f.WriteString("- 所有动画流畅无卡顿\n")
				f.WriteString("- 边缘平滑无锯齿\n")
				f.WriteString("- 颜色过渡自然\n")
				f.WriteString("- 运动轨迹正确\n")
				f.Close()
				fmt.Printf("✓ 文本数据已保存: %s\n", txtPath)
			}
		})
	}()

	fmt.Println("✓ GPU测试程序已启动: Dynamic Animation")
}
