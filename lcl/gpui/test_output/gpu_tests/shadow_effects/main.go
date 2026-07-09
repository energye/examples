// GPU Test: Shadow Effects - 阴影效果渲染验证
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

// ShadowTestWidget 自定义控件，用于渲染阴影效果测试
type ShadowTestWidget struct {
	widget.BaseWidget
}

func NewShadowTestWidget() *ShadowTestWidget {
	w := &ShadowTestWidget{
		BaseWidget: widget.NewBaseWidget(),
	}
	w.SetOwner(w)
	return w
}

// Render 渲染阴影效果测试内容
func (w *ShadowTestWidget) Render(ctx *widget.Context) {
	if w == nil || ctx == nil || ctx.Renderer == nil {
		return
	}
	bounds := w.Bounds()
	ctx.Renderer.PushClip(bounds)

	// 使用浅灰色背景以便观察阴影
	ctx.Renderer.FillRect(bounds, math.NewColor(0.95, 0.95, 0.95, 1))

	// 测试1: 小阴影（按钮样式）- 增加alpha值
	ctx.Renderer.DrawShadow(
		math.NewRect(100, 100, 200, 60),
		math.NewVec2(0, 3), 8,
		math.NewColor(0, 0, 0, 0.25),
	)
	ctx.Renderer.FillRoundRect(math.NewRect(100, 100, 200, 60), 8, math.NewColor(0.2, 0.5, 1, 1))

	// 测试2: 中阴影（卡片样式）- 增加alpha值
	ctx.Renderer.DrawShadow(
		math.NewRect(400, 100, 300, 150),
		math.NewVec2(0, 6), 16,
		math.NewColor(0, 0, 0, 0.3),
	)
	ctx.Renderer.FillRoundRect(math.NewRect(400, 100, 300, 150), 12, math.NewColor(1, 1, 1, 1))

	// 测试3: 大阴影（弹窗样式）- 增加alpha值
	ctx.Renderer.DrawShadow(
		math.NewRect(800, 100, 400, 200),
		math.NewVec2(0, 10), 30,
		math.NewColor(0, 0, 0, 0.35),
	)
	ctx.Renderer.FillRoundRect(math.NewRect(800, 100, 400, 200), 16, math.NewColor(1, 1, 1, 1))

	// 测试4: 彩色阴影 - 增加alpha值
	ctx.Renderer.DrawShadow(
		math.NewRect(100, 350, 200, 100),
		math.NewVec2(0, 6), 16,
		math.NewColor(0.2, 0.5, 1, 0.5),
	)
	ctx.Renderer.FillRoundRect(math.NewRect(100, 350, 200, 100), 12, math.NewColor(1, 1, 1, 1))

	// 测试5: 偏移阴影 - 增加alpha值
	ctx.Renderer.DrawShadow(
		math.NewRect(400, 350, 200, 100),
		math.NewVec2(10, 10), 20,
		math.NewColor(0, 0, 0, 0.35),
	)
	ctx.Renderer.FillRoundRect(math.NewRect(400, 350, 200, 100), 12, math.NewColor(1, 1, 1, 1))

	// 测试6: 多层阴影 - 增加alpha值
	ctx.Renderer.DrawShadow(
		math.NewRect(800, 350, 200, 100),
		math.NewVec2(0, 3), 6,
		math.NewColor(0, 0, 0, 0.2),
	)
	ctx.Renderer.DrawShadow(
		math.NewRect(800, 350, 200, 100),
		math.NewVec2(0, 10), 24,
		math.NewColor(0, 0, 0, 0.3),
	)
	ctx.Renderer.FillRoundRect(math.NewRect(800, 350, 200, 100), 12, math.NewColor(1, 1, 1, 1))

	// 测试7: 圆形阴影 - 使用圆形阴影绘制
	ctx.Renderer.DrawShadow(
		math.NewRect(1200, 150, 100, 100),
		math.NewVec2(0, 6), 16,
		math.NewColor(0, 0, 0, 0.3),
	)
	ctx.Renderer.FillCircle(math.NewVec2(1250, 200), 50, math.NewColor(0.2, 0.8, 0.4, 1))

	ctx.Renderer.PopClip()
}

func main() {
	libname.UseWS = "gtk3"
	if ws := os.Getenv("GPUI_WS"); ws != "" {
		libname.UseWS = ws
	}
	app := ui.NewApplication("GPU Test: Shadow Effects", 1600, 900)

	app.OnSetup(func(engine *ui.Engine) {
		setupUI(engine)
	})

	app.Run()
}

func setupUI(engine *ui.Engine) {
	tokens := token.Current()
	outDir := os.Getenv("GPUI_TEST_OUTPUT")
	if outDir == "" {
		outDir = "lcl/gpui/test_output/gpu_tests/shadow_effects"
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
	title := widget.NewText("GPU Test: Shadow Effects (阴影效果渲染)")
	title.SetPos(40, 20)
	title.SetSize(800, 40)
	title.SetFont(engine.Font())
	title.SetColor(tokens.Global.ColorText)
	engine.AddWidget(title)

	// 测试标签
	label1 := widget.NewText("1. 小阴影 (按钮样式)")
	label1.SetPos(40, 80)
	label1.SetSize(400, 20)
	label1.SetFont(engine.Font())
	label1.SetColor(tokens.Global.ColorTextSecondary)
	engine.AddWidget(label1)

	label2 := widget.NewText("2. 中阴影 (卡片样式)")
	label2.SetPos(400, 80)
	label2.SetSize(400, 20)
	label2.SetFont(engine.Font())
	label2.SetColor(tokens.Global.ColorTextSecondary)
	engine.AddWidget(label2)

	label3 := widget.NewText("3. 大阴影 (弹窗样式)")
	label3.SetPos(800, 80)
	label3.SetSize(400, 20)
	label3.SetFont(engine.Font())
	label3.SetColor(tokens.Global.ColorTextSecondary)
	engine.AddWidget(label3)

	label4 := widget.NewText("4. 彩色阴影")
	label4.SetPos(40, 320)
	label4.SetSize(400, 20)
	label4.SetFont(engine.Font())
	label4.SetColor(tokens.Global.ColorTextSecondary)
	engine.AddWidget(label4)

	label5 := widget.NewText("5. 偏移阴影")
	label5.SetPos(400, 320)
	label5.SetSize(400, 20)
	label5.SetFont(engine.Font())
	label5.SetColor(tokens.Global.ColorTextSecondary)
	engine.AddWidget(label5)

	label6 := widget.NewText("6. 多层阴影")
	label6.SetPos(800, 320)
	label6.SetSize(400, 20)
	label6.SetFont(engine.Font())
	label6.SetColor(tokens.Global.ColorTextSecondary)
	engine.AddWidget(label6)

	label7 := widget.NewText("7. 圆形阴影")
	label7.SetPos(1200, 80)
	label7.SetSize(400, 20)
	label7.SetFont(engine.Font())
	label7.SetColor(tokens.Global.ColorTextSecondary)
	engine.AddWidget(label7)

	// 添加自定义渲染控件
	testWidget := NewShadowTestWidget()
	testWidget.SetPos(0, 0)
	testWidget.SetSize(1600, 900)
	engine.AddWidget(testWidget)

	// 预期效果说明
	expected := widget.NewText(`预期效果:
1. 小阴影: 轻微的底部阴影，适合按钮控件
2. 中阴影: 中等强度的阴影，适合卡片控件
3. 大阴影: 较强的阴影，适合弹窗控件
4. 彩色阴影: 蓝色调的阴影，增加视觉层次
5. 偏移阴影: 向右下方偏移的阴影，模拟光源方向
6. 多层阴影: 两层阴影叠加，增加深度感
7. 圆形阴影: 圆形控件的阴影，边缘柔和
验证标准: 所有阴影边缘柔和无锯齿，大小和偏移符合预期，颜色正确`)
	expected.SetPos(40, 600)
	expected.SetSize(1520, 280)
	expected.SetFont(engine.Font())
	expected.SetColor(tokens.Global.ColorText)
	engine.AddWidget(expected)

	// 渲染完成后保存
	go func() {
		time.Sleep(3 * time.Second)
		lcl.RunOnMainThreadAsync(func(id uint32) {
			imgPath := filepath.Join(outDir, "gpu_shadow_effects.png")
			engine.Renderer().SavePNG(imgPath)
			fmt.Printf("✓ 图片已保存: %s\n", imgPath)

			txtPath := filepath.Join(outDir, "gpu_shadow_effects.txt")
			f, _ := os.Create(txtPath)
			if f != nil {
				f.WriteString("GPU Test: Shadow Effects\n")
				f.WriteString("========================\n\n")
				f.WriteString("渲染内容:\n")
				f.WriteString("1. 小阴影: (100,100) 200x60, offset=(0,2), blur=4, alpha=0.1\n")
				f.WriteString("2. 中阴影: (400,100) 300x150, offset=(0,4), blur=12, alpha=0.15\n")
				f.WriteString("3. 大阴影: (800,100) 400x200, offset=(0,8), blur=24, alpha=0.2\n")
				f.WriteString("4. 彩色阴影: (100,350) 200x100, color=blue, alpha=0.3\n")
				f.WriteString("5. 偏移阴影: (400,350) 200x100, offset=(8,8), blur=16\n")
				f.WriteString("6. 多层阴影: (800,350) 200x100, 两层阴影叠加\n")
				f.WriteString("7. 圆形阴影: (1200,150) 100x100, 圆形阴影\n\n")
				f.WriteString("验证标准:\n")
				f.WriteString("- 阴影边缘柔和无锯齿\n")
				f.WriteString("- 大小和偏移符合预期\n")
				f.WriteString("- 颜色正确\n")
				f.WriteString("- 多层阴影叠加效果正确\n")
				f.Close()
				fmt.Printf("✓ 文本数据已保存: %s\n", txtPath)
			}
		})
	}()

	fmt.Println("✓ GPU测试程序已启动: Shadow Effects")
}
