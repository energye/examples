package pipeline

import (
	"image"
	"image/color"
	"image/png"
	stdmath "math"
	"os"
	"path/filepath"
	"testing"

	"github.com/energye/examples/lcl/gpui/core/math"
)

func TestWriteCoreDrawingSnapshots(t *testing.T) {
	outDir := snapshotOutputDir(t)

	shapes := image.NewRGBA(image.Rect(0, 0, 640, 420))
	fillImage(shapes, rgba(245, 247, 250, 255))
	drawCPUShadow(shapes, rectF{40, 40, 210, 88}, 12, rgba(0, 0, 0, 45))
	drawCPURoundRect(shapes, rectF{40, 40, 210, 88}, 14, rgba(255, 255, 255, 255))
	drawCPURoundStroke(shapes, rectF{40, 40, 210, 88}, 14, 2, rgba(22, 119, 255, 255))
	drawCPURoundLinearGradient(shapes, rectF{290, 40, 290, 88}, 14, rgba(22, 119, 255, 255), rgba(82, 196, 26, 255))
	drawCPURoundStroke(shapes, rectF{290, 40, 290, 88}, 14, 2, rgba(0, 0, 0, 80))
	drawCPURoundRect(shapes, rectF{70, 175, 120, 120}, 60, rgba(250, 173, 20, 255))
	drawCPURoundStroke(shapes, rectF{230, 175, 120, 120}, 60, 8, rgba(235, 47, 150, 255))
	drawCPURoundRect(shapes, rectF{400, 180, 170, 70}, 6, rgba(19, 194, 194, 255))
	drawCPURoundStroke(shapes, rectF{400, 180, 170, 70}, 6, 3, rgba(8, 96, 96, 255))
	writePNG(t, filepath.Join(outDir, "core_shapes.png"), shapes)

	paths := image.NewRGBA(image.Rect(0, 0, 640, 420))
	fillImage(paths, rgba(255, 255, 255, 255))
	drawGrid(paths, rgba(235, 235, 235, 255), 20)
	heart, err := ParseSVGPath("M320 120 C260 60 160 100 190 210 C215 300 320 345 320 345 C320 345 425 300 450 210 C480 100 380 60 320 120 Z")
	if err != nil {
		t.Fatalf("parse heart path: %v", err)
	}
	drawCPUPath(paths, heart, rgba(255, 77, 79, 235))

	arrow := NewPath()
	arrow.MoveTo(120, 300)
	arrow.LineTo(230, 300)
	arrow.LineTo(230, 255)
	arrow.LineTo(330, 330)
	arrow.LineTo(230, 405)
	arrow.LineTo(230, 360)
	arrow.LineTo(120, 360)
	arrow.Close()
	drawCPUPath(paths, arrow, rgba(22, 119, 255, 220))
	writePNG(t, filepath.Join(outDir, "svg_path.png"), paths)
}

func snapshotOutputDir(t *testing.T) string {
	t.Helper()
	outDir := os.Getenv("GPUI_TEST_OUTPUT")
	if outDir == "" {
		outDir = filepath.Clean("../../test_output/render_core")
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		t.Fatalf("create snapshot output dir: %v", err)
	}
	return outDir
}

func writePNG(t *testing.T, path string, img image.Image) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create png %s: %v", path, err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatalf("encode png %s: %v", path, err)
	}
	info, err := f.Stat()
	if err != nil {
		t.Fatalf("stat png %s: %v", path, err)
	}
	if info.Size() == 0 {
		t.Fatalf("png %s is empty", path)
	}
	t.Logf("wrote %s", path)
}

type rectF struct {
	x, y, w, h float32
}

func rgba(r, g, b, a uint8) color.RGBA {
	return color.RGBA{R: r, G: g, B: b, A: a}
}

func fillImage(img *image.RGBA, c color.RGBA) {
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			img.SetRGBA(x, y, c)
		}
	}
}

func drawGrid(img *image.RGBA, c color.RGBA, step int) {
	b := img.Bounds()
	for x := b.Min.X; x < b.Max.X; x += step {
		for y := b.Min.Y; y < b.Max.Y; y++ {
			blendPixel(img, x, y, c)
		}
	}
	for y := b.Min.Y; y < b.Max.Y; y += step {
		for x := b.Min.X; x < b.Max.X; x++ {
			blendPixel(img, x, y, c)
		}
	}
}

func drawCPURoundRect(img *image.RGBA, r rectF, radius float32, c color.RGBA) {
	drawRoundSDF(img, r, radius, 0, c, true)
}

func drawCPURoundStroke(img *image.RGBA, r rectF, radius, width float32, c color.RGBA) {
	drawRoundSDF(img, r, radius, width, c, false)
}

func drawCPUShadow(img *image.RGBA, r rectF, blur float32, c color.RGBA) {
	steps := int(blur / 2)
	if steps < 3 {
		steps = 3
	}
	for i := steps; i >= 1; i-- {
		t := float32(i) / float32(steps)
		alpha := uint8(float32(c.A) * (1 - t) * (1 - t) / float32(steps) * 2)
		expand := blur * t
		drawCPURoundRect(img, rectF{r.x - expand, r.y - expand + blur*0.25, r.w + expand*2, r.h + expand*2}, 14+expand, rgba(c.R, c.G, c.B, alpha))
	}
}

func drawRoundSDF(img *image.RGBA, r rectF, radius, width float32, c color.RGBA, fill bool) {
	minX := maxInt(0, int(stdmath.Floor(float64(r.x-radius-width-2))))
	minY := maxInt(0, int(stdmath.Floor(float64(r.y-radius-width-2))))
	maxX := minInt(img.Bounds().Dx(), int(stdmath.Ceil(float64(r.x+r.w+radius+width+2))))
	maxY := minInt(img.Bounds().Dy(), int(stdmath.Ceil(float64(r.y+r.h+radius+width+2))))

	for y := minY; y < maxY; y++ {
		for x := minX; x < maxX; x++ {
			px := float32(x) + 0.5
			py := float32(y) + 0.5
			d := roundRectDistance(px, py, r, radius)
			alpha := float32(0)
			if fill {
				alpha = clamp01(0.5 - d)
			} else {
				outer := clamp01(0.5 - d)
				inner := clamp01(d + width + 0.5)
				alpha = outer * inner
			}
			if alpha > 0 {
				src := c
				src.A = uint8(float32(c.A) * alpha)
				blendPixel(img, x, y, src)
			}
		}
	}
}

func roundRectDistance(px, py float32, r rectF, radius float32) float32 {
	cx := r.x + r.w*0.5
	cy := r.y + r.h*0.5
	qx := abs32(px-cx) - (r.w*0.5 - radius)
	qy := abs32(py-cy) - (r.h*0.5 - radius)
	ox := max32(qx, 0)
	oy := max32(qy, 0)
	return float32(stdmath.Sqrt(float64(ox*ox+oy*oy))) + min32(max32(qx, qy), 0) - radius
}

func drawCPULinearGradient(img *image.RGBA, r rectF, start, end color.RGBA) {
	minX := maxInt(0, int(r.x))
	minY := maxInt(0, int(r.y))
	maxX := minInt(img.Bounds().Dx(), int(r.x+r.w))
	maxY := minInt(img.Bounds().Dy(), int(r.y+r.h))
	for y := minY; y < maxY; y++ {
		for x := minX; x < maxX; x++ {
			t := float32(x-minX) / max32(float32(maxX-minX), 1)
			c := color.RGBA{
				R: uint8(float32(start.R) + (float32(end.R)-float32(start.R))*t),
				G: uint8(float32(start.G) + (float32(end.G)-float32(start.G))*t),
				B: uint8(float32(start.B) + (float32(end.B)-float32(start.B))*t),
				A: uint8(float32(start.A) + (float32(end.A)-float32(start.A))*t),
			}
			blendPixel(img, x, y, c)
		}
	}
}

func drawCPURoundLinearGradient(img *image.RGBA, r rectF, radius float32, start, end color.RGBA) {
	minX := maxInt(0, int(stdmath.Floor(float64(r.x))))
	minY := maxInt(0, int(stdmath.Floor(float64(r.y))))
	maxX := minInt(img.Bounds().Dx(), int(stdmath.Ceil(float64(r.x+r.w))))
	maxY := minInt(img.Bounds().Dy(), int(stdmath.Ceil(float64(r.y+r.h))))
	for y := minY; y < maxY; y++ {
		for x := minX; x < maxX; x++ {
			d := roundRectDistance(float32(x)+0.5, float32(y)+0.5, r, radius)
			alpha := clamp01(0.5 - d)
			if alpha <= 0 {
				continue
			}
			t := float32(x-minX) / max32(float32(maxX-minX), 1)
			c := color.RGBA{
				R: uint8(float32(start.R) + (float32(end.R)-float32(start.R))*t),
				G: uint8(float32(start.G) + (float32(end.G)-float32(start.G))*t),
				B: uint8(float32(start.B) + (float32(end.B)-float32(start.B))*t),
				A: uint8((float32(start.A) + (float32(end.A)-float32(start.A))*t) * alpha),
			}
			blendPixel(img, x, y, c)
		}
	}
}

func drawCPUPath(img *image.RGBA, path *Path, c color.RGBA) {
	for _, points := range pathSubpaths(path) {
		triangles := triangulateSimplePolygon(points)
		if len(triangles) == 0 {
			continue
		}
		for _, tri := range triangles {
			drawTriangle(img, points[tri[0]], points[tri[1]], points[tri[2]], c)
		}
	}
}

func drawTriangle(img *image.RGBA, a, b, c math.Vec2, col color.RGBA) {
	minX := maxInt(0, int(stdmath.Floor(float64(min32(a.X, min32(b.X, c.X))))))
	minY := maxInt(0, int(stdmath.Floor(float64(min32(a.Y, min32(b.Y, c.Y))))))
	maxX := minInt(img.Bounds().Dx(), int(stdmath.Ceil(float64(max32(a.X, max32(b.X, c.X))))))
	maxY := minInt(img.Bounds().Dy(), int(stdmath.Ceil(float64(max32(a.Y, max32(b.Y, c.Y))))))
	for y := minY; y < maxY; y++ {
		for x := minX; x < maxX; x++ {
			p := math.NewVec2(float32(x)+0.5, float32(y)+0.5)
			if pointInTriangle(p, a, b, c) {
				blendPixel(img, x, y, col)
			}
		}
	}
}

func blendPixel(img *image.RGBA, x, y int, src color.RGBA) {
	if src.A == 0 {
		return
	}
	dst := img.RGBAAt(x, y)
	sa := float32(src.A) / 255
	da := float32(dst.A) / 255
	outA := sa + da*(1-sa)
	if outA <= 0 {
		img.SetRGBA(x, y, color.RGBA{})
		return
	}
	out := color.RGBA{
		R: uint8((float32(src.R)*sa + float32(dst.R)*da*(1-sa)) / outA),
		G: uint8((float32(src.G)*sa + float32(dst.G)*da*(1-sa)) / outA),
		B: uint8((float32(src.B)*sa + float32(dst.B)*da*(1-sa)) / outA),
		A: uint8(outA * 255),
	}
	img.SetRGBA(x, y, out)
}

func clamp01(v float32) float32 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func abs32(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}

func min32(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

func max32(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
