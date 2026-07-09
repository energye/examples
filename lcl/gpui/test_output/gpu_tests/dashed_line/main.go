// GPU Test: DrawDashedLine - 虚线渲染验证
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

// DashedLineTestWidget 自定义控件，用于渲染虚线测试
type DashedLineTestWidget struct {
	widget.BaseWidget
}

func NewDashedLineTestWidget() *DashedLineTestWidget {
	w := &DashedLineTestWidget{
		BaseWidget: widget.NewBaseWidget(),
	}
	w.SetOwner(w)
	return w
}

// Render 渲染虚线测试内容
func (w *DashedLineTestWidget) Render(ctx *widget.Context) {
	if w == nil || ctx == nil || ctx.Renderer == nil {
		return
	}
	bounds := w.Bounds()
	ctx.Renderer.PushClip(bounds)

	// 测试1: 水平虚线
	ctx.Renderer.DrawDashedLine(100, 120, 1500, 120, 2, 10, 5, math.NewColor(0, 0, 0, 1))

	// 测试2: 垂直虚线
	ctx.Renderer.DrawDashedLine(800, 150, 800, 600, 3, 15, 8, math.NewColor(1, 0, 0, 1))

	// 测试3: 对角虚线
	ctx.Renderer.DrawDashedLine(100, 350, 1500, 550, 2, 20, 10, math.NewColor(0, 0, 1, 1))

	// 测试4: 不同宽度虚线
	ctx.Renderer.DrawDashedLine(100, 450, 1500, 450, 1, 10, 5, math.NewColor(0, 0, 0, 1))
	ctx.Renderer.DrawDashedLine(100, 470, 1500, 470, 2, 10, 5, math.NewColor(0, 0, 0, 1))
	ctx.Renderer.DrawDashedLine(100, 490, 1500, 490, 4, 10, 5, math.NewColor(0, 0, 0, 1))
	ctx.Renderer.DrawDashedLine(100, 510, 1500, 510, 8, 10, 5, math.NewColor(0, 0, 0, 1))

	// 测试5: 不同颜色虚线
	ctx.Renderer.DrawDashedLine(100, 570, 400, 570, 2, 10, 5, math.NewColor(1, 0, 0, 1))
	ctx.Renderer.DrawDashedLine(500, 570, 800, 570, 2, 10, 5, math.NewColor(0, 1, 0, 1))
	ctx.Renderer.DrawDashedLine(900, 570, 1200, 570, 2, 10, 5, math.NewColor(0, 0, 1, 1))
	ctx.Renderer.DrawDashedLine(1300, 570, 1500, 570, 2, 10, 5, math.NewColor(1, 1, 0, 1))

	ctx.Renderer.PopClip()
}

func main() {
	libname.UseWS = "gtk3"
	if ws := os.Getenv("GPUI_WS"); ws != "" {
		libname.UseWS = ws
	}
	app := ui.NewApplication("GPU Test: DrawDashedLine", 1600, 900)

	app.OnSetup(func(engine *ui.Engine) {
		setupUI(engine)
	})

	app.Run()
}

func setupUI(engine *ui.Engine) {
	tokens := token.Current()
	outDir := os.Getenv("GPUI_TEST_OUTPUT")
	if outDir == "" {
		outDir = "lcl/gpui/test_output/gpu_tests/dashed_line"
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
	title := widget.NewText("GPU Test: DrawDashedLine (虚线渲染)")
	title.SetPos(40, 20)
	title.SetSize(800, 40)
	title.SetFont(engine.Font())
	title.SetColor(tokens.Global.ColorText)
	engine.AddWidget(title)

	// 测试标签
	label1 := widget.NewText("1. 水平虚线 (dashLen=10, gapLen=5, width=2)")
	label1.SetPos(40, 80)
	label1.SetSize(400, 20)
	label1.SetFont(engine.Font())
	label1.SetColor(tokens.Global.ColorTextSecondary)
	engine.AddWidget(label1)

	label2 := widget.NewText("2. 垂直虚线 (dashLen=15, gapLen=8, width=3)")
	label2.SetPos(40, 200)
	label2.SetSize(400, 20)
	label2.SetFont(engine.Font())
	label2.SetColor(tokens.Global.ColorTextSecondary)
	engine.AddWidget(label2)

	label3 := widget.NewText("3. 对角虚线 (dashLen=20, gapLen=10, width=2)")
	label3.SetPos(40, 320)
	label3.SetSize(400, 20)
	label3.SetFont(engine.Font())
	label3.SetColor(tokens.Global.ColorTextSecondary)
	engine.AddWidget(label3)

	label4 := widget.NewText("4. 不同宽度虚线 (width=1,2,4,8)")
	label4.SetPos(40, 420)
	label4.SetSize(400, 20)
	label4.SetFont(engine.Font())
	label4.SetColor(tokens.Global.ColorTextSecondary)
	engine.AddWidget(label4)

	label5 := widget.NewText("5. 不同颜色虚线 (红/绿/蓝/黄)")
	label5.SetPos(40, 540)
	label5.SetSize(400, 20)
	label5.SetFont(engine.Font())
	label5.SetColor(tokens.Global.ColorTextSecondary)
	engine.AddWidget(label5)

	// 添加自定义渲染控件
	testWidget := NewDashedLineTestWidget()
	testWidget.SetPos(0, 0)
	testWidget.SetSize(1600, 900)
	engine.AddWidget(testWidget)

	// 预期效果说明
	expected := widget.NewText(`预期效果:
1. 水平虚线: 从左到右的虚线，dash和gap交替出现，线条均匀，无断裂或重叠
2. 垂直虚线: 从上到下的虚线，dash和gap交替出现，线条均匀，无断裂或重叠
3. 对角虚线: 从左上到右下的虚线，dash和gap交替出现，线条均匀，无断裂或重叠
4. 不同宽度: 4条虚线，宽度分别为1,2,4,8像素，宽度差异明显可见
5. 不同颜色: 4条虚线，颜色分别为红/绿/蓝/黄，颜色清晰可辨，无混色
验证标准: 所有虚线的dash长度一致，gap长度一致，无断裂或重叠，宽度和颜色符合预期`)
	expected.SetPos(40, 620)
	expected.SetSize(1520, 260)
	expected.SetFont(engine.Font())
	expected.SetColor(tokens.Global.ColorText)
	engine.AddWidget(expected)

	// 渲染完成后保存
	go func() {
		time.Sleep(3 * time.Second)

		lcl.RunOnMainThreadAsync(func(id uint32) {
			imgPath := filepath.Join(outDir, "gpu_dashed_line.png")
			engine.Renderer().SavePNG(imgPath)
			fmt.Printf("✓ 图片已保存: %s\n", imgPath)
		})
	}()

	fmt.Println("✓ GPU测试程序已启动: DrawDashedLine")
}
