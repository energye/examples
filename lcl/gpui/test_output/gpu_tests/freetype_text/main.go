// GPU Test: Freetype text rendering - Freetype 文本渲染验证
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/energye/lcl/api/libname"
	lcl "github.com/energye/lcl/lcl"
	xfont "golang.org/x/image/font"

	"github.com/energye/examples/lcl/gpui/core/math"
	renderfont "github.com/energye/examples/lcl/gpui/render/font"
	"github.com/energye/examples/lcl/gpui/render/pipeline"
	"github.com/energye/examples/lcl/gpui/style/token"
	"github.com/energye/examples/lcl/gpui/ui"
	"github.com/energye/examples/lcl/gpui/widget"
)

type FreetypeTextWidget struct {
	widget.BaseWidget
	regular *renderfont.Font
	bold    *renderfont.Font
	italic  *renderfont.Font
	spaced  *renderfont.Font
}

func NewFreetypeTextWidget(regular, bold, italic, spaced *renderfont.Font) *FreetypeTextWidget {
	w := &FreetypeTextWidget{
		BaseWidget: widget.NewBaseWidget(),
		regular:    regular,
		bold:       bold,
		italic:     italic,
		spaced:     spaced,
	}
	w.SetOwner(w)
	return w
}

func (w *FreetypeTextWidget) Render(ctx *widget.Context) {
	if w == nil || ctx == nil || ctx.Renderer == nil {
		return
	}
	f := firstFont(w.regular, ctx.Font)
	bold := firstFont(w.bold, f)
	italic := firstFont(w.italic, f)
	spaced := firstFont(w.spaced, f)

	ctx.Renderer.PushClip(w.Bounds())

	ctx.Renderer.DrawText("8pt: crisp small text 你好世界 1234567890", 64, 92, f, math.NewColor(0.05, 0.05, 0.05, 1))
	ctx.Renderer.DrawText("Default: Freetype AA text rendering / 矢量抗锯齿文本", 64, 132, f, math.NewColor(0.02, 0.18, 0.34, 1))
	ctx.Renderer.DrawText("Bold synthesis: The quick brown fox / 粗体文本", 64, 172, bold, math.NewColor(0.48, 0.12, 0.08, 1))
	ctx.Renderer.DrawText("Italic synthesis: slanted glyph edges / 斜体文本", 64, 212, italic, math.NewColor(0.08, 0.28, 0.12, 1))
	ctx.Renderer.DrawText("Letter spacing: A V A T A R  字 间 距", 64, 252, spaced, math.NewColor(0.28, 0.12, 0.42, 1))

	rect := math.NewRect(64, 306, 520, 150)
	ctx.Renderer.StrokeRoundRect(rect, 8, 1, math.NewColor(0.55, 0.62, 0.72, 1))
	ctx.Renderer.DrawTextInRect(
		"Left aligned multiline text uses the same freetype atlas cache and wraps cleanly. 多行文本自动换行，边缘应平滑无毛边。",
		math.NewRect(rect.X+14, rect.Y+12, rect.W-28, rect.H-24),
		pipeline.TextOptions{Font: f, Color: math.NewColor(0.12, 0.14, 0.18, 1), Align: pipeline.TextAlignLeft, MaxLines: 4, LineHeight: 28},
	)

	centerRect := math.NewRect(640, 306, 360, 86)
	rightRect := math.NewRect(640, 420, 360, 86)
	ctx.Renderer.StrokeRoundRect(centerRect, 8, 1, math.NewColor(0.55, 0.62, 0.72, 1))
	ctx.Renderer.StrokeRoundRect(rightRect, 8, 1, math.NewColor(0.55, 0.62, 0.72, 1))
	ctx.Renderer.DrawTextInRect("Centered text\n居中对齐", centerRect, pipeline.TextOptions{
		Font: f, Color: math.NewColor(0.05, 0.22, 0.36, 1), Align: pipeline.TextAlignCenter, MaxLines: 2, LineHeight: 30,
	})
	ctx.Renderer.DrawTextInRect("Right aligned text\n右对齐", rightRect, pipeline.TextOptions{
		Font: f, Color: math.NewColor(0.36, 0.16, 0.04, 1), Align: pipeline.TextAlignRight, MaxLines: 2, LineHeight: 30,
	})

	ctx.Renderer.DrawTextInRect(
		"Ellipsis should truncate this long freetype rendered line without changing caller APIs.",
		math.NewRect(64, 520, 460, 34),
		pipeline.TextOptions{Font: f, Color: math.NewColor(0.12, 0.12, 0.12, 1), Ellipsis: true, MaxLines: 1},
	)

	ctx.Renderer.PopClip()
}

func firstFont(primary, fallback *renderfont.Font) *renderfont.Font {
	if primary != nil {
		return primary
	}
	return fallback
}

func main() {
	libname.UseWS = "gtk3"
	if ws := os.Getenv("GPUI_WS"); ws != "" {
		libname.UseWS = ws
	}

	app := ui.NewApplication("GPU Test: Freetype Text", 1100, 680)
	app.OnSetup(func(engine *ui.Engine) {
		setupUI(engine)
	})
	app.Run()
}

func setupUI(engine *ui.Engine) {
	tokens := token.Current()
	outDir := os.Getenv("GPUI_TEST_OUTPUT")
	if outDir == "" {
		outDir = "lcl/gpui/test_output/gpu_tests/freetype_text"
	}
	os.MkdirAll(outDir, 0o755)

	panel := widget.NewBox(pipeline.BoxStyle{
		Background:  tokens.Global.ColorBgContainer,
		BorderColor: tokens.Global.ColorBorder,
		BorderWidth: 1,
		Radius:      tokens.Global.RadiusLG,
	})
	panel.SetPos(0, 0)
	panel.SetSize(1100, 680)
	engine.AddWidget(panel)

	title := widget.NewText("GPU Test: Freetype Text Rendering")
	title.SetPos(40, 24)
	title.SetSize(900, 36)
	title.SetFont(engine.Font())
	title.SetColor(tokens.Global.ColorText)
	engine.AddWidget(title)

	regular, bold, italic, spaced := buildTestFonts(engine.Font())
	testWidget := NewFreetypeTextWidget(regular, bold, italic, spaced)
	testWidget.SetPos(0, 0)
	testWidget.SetSize(1100, 680)
	engine.AddWidget(testWidget)

	go func() {
		time.Sleep(3 * time.Second)
		lcl.RunOnMainThreadAsync(func(id uint32) {
			imgPath := filepath.Join(outDir, "gpu_freetype_text.png")
			if err := engine.Renderer().SavePNG(imgPath); err != nil {
				fmt.Println("GPU snapshot error:", err)
				return
			}
			fmt.Printf("✓ 图片已保存: %s\n", imgPath)

			txtPath := filepath.Join(outDir, "gpu_freetype_text.txt")
			f, _ := os.Create(txtPath)
			if f != nil {
				f.WriteString("GPU Test: Freetype text rendering\n")
				f.WriteString("=================================\n\n")
				f.WriteString("Coverage:\n")
				f.WriteString("- Small text anti-aliasing\n")
				f.WriteString("- CJK and Latin mixed text\n")
				f.WriteString("- Bold and italic synthetic styles\n")
				f.WriteString("- Letter spacing\n")
				f.WriteString("- Multiline wrapping, center/right alignment, ellipsis\n")
				f.Close()
				fmt.Printf("✓ 文本数据已保存: %s\n", txtPath)
			}
		})
	}()

	fmt.Println("✓ GPU测试程序已启动: Freetype Text")
}

func buildTestFonts(fallback *renderfont.Font) (*renderfont.Font, *renderfont.Font, *renderfont.Font, *renderfont.Font) {
	if ui.DefaultFontData == nil {
		return fallback, fallback, fallback, fallback
	}

	regular, err := newDefaultStyledFont(renderfont.FontStyle{
		Size: 10, DPI: 96, Hinting: xfont.HintingFull,
	})
	if err != nil {
		regular = fallback
	}
	bold, err := newDefaultStyledFont(renderfont.FontStyle{
		Size: 18, DPI: 96, Hinting: xfont.HintingFull, Bold: true,
	})
	if err != nil {
		bold = fallback
	}
	italic, err := newDefaultStyledFont(renderfont.FontStyle{
		Size: 18, DPI: 96, Hinting: xfont.HintingFull, Italic: true,
	})
	if err != nil {
		italic = fallback
	}
	spaced, err := newDefaultStyledFont(renderfont.FontStyle{
		Size: 18, DPI: 96, Hinting: xfont.HintingFull, LetterSpacing: 3,
	})
	if err != nil {
		spaced = fallback
	}
	return regular, bold, italic, spaced
}

func newDefaultStyledFont(style renderfont.FontStyle) (*renderfont.Font, error) {
	f, err := renderfont.NewFontStyled(ui.DefaultFontData, style)
	if err != nil {
		return nil, err
	}
	fallbacks := make([]*renderfont.Font, 0, len(ui.DefaultFontFallbackData))
	for _, data := range ui.DefaultFontFallbackData {
		fallback, err := renderfont.NewFontStyled(data, style)
		if err == nil && fallback != nil {
			fallbacks = append(fallbacks, fallback)
		}
	}
	f.SetFallbacks(fallbacks...)
	return f, nil
}
