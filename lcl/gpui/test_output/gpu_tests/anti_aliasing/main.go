// GPU Test: Anti-aliasing - 抗锯齿渲染验证
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

// AntiAliasingTestWidget 自定义控件，用于渲染抗锯齿测试
type AntiAliasingTestWidget struct {
	widget.BaseWidget
}

func NewAntiAliasingTestWidget() *AntiAliasingTestWidget {
	w := &AntiAliasingTestWidget{
		BaseWidget: widget.NewBaseWidget(),
	}
	w.SetOwner(w)
	return w
}

// Render 渲染抗锯齿测试内容
func (w *AntiAliasingTestWidget) Render(ctx *widget.Context) {
	if w == nil || ctx == nil || ctx.Renderer == nil {
		return
	}
	bounds := w.Bounds()
	ctx.Renderer.PushClip(bounds)

	// 测试1: 小圆角矩形（按钮尺寸）
	ctx.Renderer.FillRoundRect(math.NewRect(100, 100, 80, 32), 4, math.NewColor(0.2, 0.5, 1, 1))
	ctx.Renderer.StrokeRoundRect(math.NewRect(100, 100, 80, 32), 4, 1, math.NewColor(0.1, 0.3, 0.8, 1))

	// 测试2: 中圆角矩形（输入框尺寸）- 使用可见颜色
	ctx.Renderer.FillRoundRect(math.NewRect(200, 100, 200, 40), 8, math.NewColor(0.9, 0.95, 1, 1))
	ctx.Renderer.StrokeRoundRect(math.NewRect(200, 100, 200, 40), 8, 2, math.NewColor(0, 0, 0, 0.3))

	// 测试3: 大圆角矩形（卡片尺寸）- 使用可见颜色
	ctx.Renderer.FillRoundRect(math.NewRect(100, 200, 300, 150), 16, math.NewColor(0.95, 0.95, 0.95, 1))
	ctx.Renderer.StrokeRoundRect(math.NewRect(100, 200, 300, 150), 16, 2, math.NewColor(0, 0, 0, 0.2))

	// 测试4: 胶囊形（Tag/Pill）
	ctx.Renderer.FillRoundRect(math.NewRect(500, 100, 120, 32), 16, math.NewColor(0.2, 0.8, 0.4, 1))

	// 测试5: 圆形
	ctx.Renderer.FillCircle(math.NewVec2(600, 300), 60, math.NewColor(1, 0.5, 0, 1))
	ctx.Renderer.StrokeCircle(math.NewVec2(600, 300), 60, 2, math.NewColor(0.8, 0.4, 0, 1))

	// 测试6: 小圆形（头像尺寸）
	ctx.Renderer.FillCircle(math.NewVec2(750, 150), 20, math.NewColor(0.8, 0.2, 0.8, 1))

	// 测试7: 斜线（抗锯齿）
	for i := 0; i < 8; i++ {
		x1 := float32(800 + i*50)
		y1 := float32(100)
		x2 := float32(800 + i*50 + 100)
		y2 := float32(300)
		ctx.Renderer.DrawLine(x1, y1, x2, y2, 2, math.NewColor(0, 0, 0, 0.5))
	}

	// 测试8: 小文字（抗锯齿）- 使用更大字号
	if ctx.Font != nil {
		ctx.Renderer.DrawText("Hello World 你好世界", 100, 450, ctx.Font, math.NewColor(0, 0, 0, 1))
		ctx.Renderer.DrawText("ABCDEFGHIJ 1234567890", 100, 480, ctx.Font, math.NewColor(0, 0, 0, 0.8))
	}

	// 测试9: 混合形状
	ctx.Renderer.FillRoundRect(math.NewRect(500, 400, 200, 100), 20, math.NewColor(0.9, 0.9, 0.9, 1))
	ctx.Renderer.StrokeRoundRect(math.NewRect(500, 400, 200, 100), 20, 2, math.NewColor(0, 0, 0, 0.2))
	ctx.Renderer.FillCircle(math.NewVec2(600, 450), 30, math.NewColor(0.2, 0.5, 1, 1))
	if ctx.Font != nil {
		ctx.Renderer.DrawText("混合", 570, 440, ctx.Font, math.NewColor(1, 1, 1, 1))
	}

	ctx.Renderer.PopClip()
}

func main() {
	libname.UseWS = "gtk3"
	if ws := os.Getenv("GPUI_WS"); ws != "" {
		libname.UseWS = ws
	}
	app := ui.NewApplication("GPU Test: Anti-aliasing", 1600, 900)

	app.OnSetup(func(engine *ui.Engine) {
		setupUI(engine)
	})

	app.Run()
}

func setupUI(engine *ui.Engine) {
	tokens := token.Current()
	outDir := os.Getenv("GPUI_TEST_OUTPUT")
	if outDir == "" {
		outDir = "lcl/gpui/test_output/gpu_tests/anti_aliasing"
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
	title := widget.NewText("GPU Test: Anti-aliasing (抗锯齿渲染)")
	title.SetPos(40, 20)
	title.SetSize(800, 40)
	title.SetFont(engine.Font())
	title.SetColor(tokens.Global.ColorText)
	engine.AddWidget(title)

	// 测试标签
	label1 := widget.NewText("1. 小圆角矩形 (按钮尺寸, radius=4)")
	label1.SetPos(40, 80)
	label1.SetSize(400, 20)
	label1.SetFont(engine.Font())
	label1.SetColor(tokens.Global.ColorTextSecondary)
	engine.AddWidget(label1)

	label2 := widget.NewText("2. 中圆角矩形 (输入框尺寸, radius=8)")
	label2.SetPos(200, 80)
	label2.SetSize(400, 20)
	label2.SetFont(engine.Font())
	label2.SetColor(tokens.Global.ColorTextSecondary)
	engine.AddWidget(label2)

	label3 := widget.NewText("3. 大圆角矩形 (卡片尺寸, radius=16)")
	label3.SetPos(40, 180)
	label3.SetSize(400, 20)
	label3.SetFont(engine.Font())
	label3.SetColor(tokens.Global.ColorTextSecondary)
	engine.AddWidget(label3)

	label4 := widget.NewText("4. 胶囊形 (Tag/Pill, radius=16)")
	label4.SetPos(500, 80)
	label4.SetSize(400, 20)
	label4.SetFont(engine.Font())
	label4.SetColor(tokens.Global.ColorTextSecondary)
	engine.AddWidget(label4)

	label5 := widget.NewText("5. 圆形 (头像尺寸)")
	label5.SetPos(550, 250)
	label5.SetSize(400, 20)
	label5.SetFont(engine.Font())
	label5.SetColor(tokens.Global.ColorTextSecondary)
	engine.AddWidget(label5)

	label6 := widget.NewText("6. 小圆形")
	label6.SetPos(730, 130)
	label6.SetSize(400, 20)
	label6.SetFont(engine.Font())
	label6.SetColor(tokens.Global.ColorTextSecondary)
	engine.AddWidget(label6)

	label7 := widget.NewText("7. 斜线 (抗锯齿)")
	label7.SetPos(800, 80)
	label7.SetSize(400, 20)
	label7.SetFont(engine.Font())
	label7.SetColor(tokens.Global.ColorTextSecondary)
	engine.AddWidget(label7)

	label8 := widget.NewText("8. 小文字 (抗锯齿)")
	label8.SetPos(40, 420)
	label8.SetSize(400, 20)
	label8.SetFont(engine.Font())
	label8.SetColor(tokens.Global.ColorTextSecondary)
	engine.AddWidget(label8)

	label9 := widget.NewText("9. 混合形状")
	label9.SetPos(500, 380)
	label9.SetSize(400, 20)
	label9.SetFont(engine.Font())
	label9.SetColor(tokens.Global.ColorTextSecondary)
	engine.AddWidget(label9)

	// 添加自定义渲染控件
	testWidget := NewAntiAliasingTestWidget()
	testWidget.SetPos(0, 0)
	testWidget.SetSize(1600, 900)
	engine.AddWidget(testWidget)

	// 预期效果说明
	expected := widget.NewText(`预期效果:
1. 小圆角矩形: 边缘平滑无锯齿，圆角处无像素化
2. 中圆角矩形: 边缘平滑无锯齿，圆角处无像素化
3. 大圆角矩形: 边缘平滑无锯齿，圆角处无像素化
4. 胶囊形: 两端半圆平滑，无锯齿
5. 圆形: 边缘平滑无锯齿，无椭圆变形
6. 小圆形: 边缘平滑无锯齿
7. 斜线: 线条平滑无锯齿，无阶梯感
8. 小文字: 文字边缘平滑，无锯齿
9. 混合形状: 多种形状组合，边缘都平滑
验证标准: 所有形状边缘平滑无锯齿，无像素化阶梯，文字清晰可读`)
	expected.SetPos(40, 600)
	expected.SetSize(1520, 280)
	expected.SetFont(engine.Font())
	expected.SetColor(tokens.Global.ColorText)
	engine.AddWidget(expected)

	// 渲染完成后保存
	go func() {
		time.Sleep(3 * time.Second)
		lcl.RunOnMainThreadAsync(func(id uint32) {
			imgPath := filepath.Join(outDir, "gpu_anti_aliasing.png")
			engine.Renderer().SavePNG(imgPath)
			fmt.Printf("✓ 图片已保存: %s\n", imgPath)

			txtPath := filepath.Join(outDir, "gpu_anti_aliasing.txt")
			f, _ := os.Create(txtPath)
			if f != nil {
				f.WriteString("GPU Test: Anti-aliasing\n")
				f.WriteString("=======================\n\n")
				f.WriteString("渲染内容:\n")
				f.WriteString("1. 小圆角矩形: (100,100) 80x32, radius=4\n")
				f.WriteString("2. 中圆角矩形: (200,100) 200x40, radius=8\n")
				f.WriteString("3. 大圆角矩形: (100,200) 300x150, radius=16\n")
				f.WriteString("4. 胶囊形: (500,100) 120x32, radius=16\n")
				f.WriteString("5. 圆形: (600,300) radius=60\n")
				f.WriteString("6. 小圆形: (750,150) radius=20\n")
				f.WriteString("7. 斜线: 8条不同角度的线段\n")
				f.WriteString("8. 小文字: 'Hello World 你好世界'\n")
				f.WriteString("9. 混合形状: 圆角矩形+圆形+文字\n\n")
				f.WriteString("验证标准:\n")
				f.WriteString("- 所有形状边缘平滑无锯齿\n")
				f.WriteString("- 无像素化阶梯\n")
				f.WriteString("- 文字清晰可读\n")
				f.Close()
				fmt.Printf("✓ 文本数据已保存: %s\n", txtPath)
			}
		})
	}()

	fmt.Println("✓ GPU测试程序已启动: Anti-aliasing")
}
