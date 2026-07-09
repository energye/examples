package pipeline

import (
	"image"
	"image/color"
	"path/filepath"
	stdmath "math"
	"testing"

	"github.com/energye/examples/lcl/gpui/core/math"
)

// TestDrawDashedLine verifies that DrawDashedLine produces correct dash segments.
func TestDrawDashedLine(t *testing.T) {
	outDir := snapshotOutputDir(t)
	img := image.NewRGBA(image.Rect(0, 0, 640, 420))
	fillImage(img, rgba(255, 255, 255, 255))

	// Draw horizontal dashed line
	drawCPUDashedLine(img, 40, 60, 600, 60, 2, 10, 5, rgba(0, 0, 0, 255))
	// Draw vertical dashed line
	drawCPUDashedLine(img, 320, 40, 320, 380, 2, 10, 5, rgba(255, 0, 0, 255))
	// Draw diagonal dashed line
	drawCPUDashedLine(img, 40, 40, 600, 380, 2, 15, 8, rgba(0, 0, 255, 255))

	writePNG(t, filepath.Join(outDir, "dashed_line.png"), img)

	// Verify the image has non-white pixels (dashes were drawn)
	nonWhite := 0
	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			if r < 60000 || g < 60000 || b < 60000 {
				nonWhite++
			}
		}
	}
	if nonWhite < 100 {
		t.Errorf("DrawDashedLine: expected at least 100 non-white pixels, got %d", nonWhite)
	}
}

// TestDrawArrow verifies that DrawArrow produces correct triangle shapes.
func TestDrawArrow(t *testing.T) {
	outDir := snapshotOutputDir(t)
	img := image.NewRGBA(image.Rect(0, 0, 640, 420))
	fillImage(img, rgba(255, 255, 255, 255))

	// Draw arrows in 4 directions
	drawCPUArrow(img, 160, 100, 40, 0, rgba(22, 119, 255, 255))  // Up
	drawCPUArrow(img, 480, 100, 40, 1, rgba(82, 196, 26, 255))   // Right
	drawCPUArrow(img, 160, 320, 40, 2, rgba(255, 77, 79, 255))   // Down
	drawCPUArrow(img, 480, 320, 40, 3, rgba(250, 173, 20, 255))  // Left

	writePNG(t, filepath.Join(outDir, "arrow.png"), img)

	// Verify non-white pixels exist in arrow regions
	nonWhite := 0
	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			if r < 60000 || g < 60000 || b < 60000 {
				nonWhite++
			}
		}
	}
	if nonWhite < 50 {
		t.Errorf("DrawArrow: expected at least 50 non-white pixels, got %d", nonWhite)
	}
}

// TestDrawTextCursor verifies that DrawTextCursor produces a visible cursor.
func TestDrawTextCursor(t *testing.T) {
	outDir := snapshotOutputDir(t)
	img := image.NewRGBA(image.Rect(0, 0, 640, 420))
	fillImage(img, rgba(255, 255, 255, 255))

	// Draw cursor at different positions
	drawCPURect(img, rectF{100-1, 100, 2, 20}, rgba(0, 0, 0, 255))
	drawCPURect(img, rectF{200-1, 150, 2, 24}, rgba(22, 119, 255, 255))
	drawCPURect(img, rectF{300-1, 200, 2, 28}, rgba(255, 0, 0, 255))

	writePNG(t, filepath.Join(outDir, "text_cursor.png"), img)

	// Verify cursor pixels exist
	nonWhite := 0
	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			if r < 60000 || g < 60000 || b < 60000 {
				nonWhite++
			}
		}
	}
	if nonWhite < 30 {
		t.Errorf("DrawTextCursor: expected at least 30 non-white pixels, got %d", nonWhite)
	}
}

// TestDrawSelectionHighlight verifies selection highlight rendering.
func TestDrawSelectionHighlight(t *testing.T) {
	outDir := snapshotOutputDir(t)
	img := image.NewRGBA(image.Rect(0, 0, 640, 420))
	fillImage(img, rgba(255, 255, 255, 255))

	// Draw selection highlights
	drawCPURect(img, rectF{50, 100, 200, 20}, rgba(22, 119, 255, 64))  // Selection
	drawCPURect(img, rectF{50, 150, 300, 20}, rgba(22, 119, 255, 64))  // Longer selection

	writePNG(t, filepath.Join(outDir, "selection_highlight.png"), img)

	// Verify highlight pixels exist
	nonWhite := 0
	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			if r < 60000 || g < 60000 || b < 60000 {
				nonWhite++
			}
		}
	}
	if nonWhite < 100 {
		t.Errorf("DrawSelectionHighlight: expected at least 100 non-white pixels, got %d", nonWhite)
	}
}

// TestDrawUnderline verifies underline rendering.
func TestDrawUnderline(t *testing.T) {
	outDir := snapshotOutputDir(t)
	img := image.NewRGBA(image.Rect(0, 0, 640, 420))
	fillImage(img, rgba(255, 255, 255, 255))

	// Draw underlines at different positions
	drawCPURect(img, rectF{50, 120, 200, 2}, rgba(0, 0, 0, 255))
	drawCPURect(img, rectF{50, 170, 300, 2}, rgba(22, 119, 255, 255))

	writePNG(t, filepath.Join(outDir, "underline.png"), img)

	// Verify underline pixels exist
	nonWhite := 0
	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			if r < 60000 || g < 60000 || b < 60000 {
				nonWhite++
			}
		}
	}
	if nonWhite < 20 {
		t.Errorf("DrawUnderline: expected at least 20 non-white pixels, got %d", nonWhite)
	}
}

// TestDrawStrikethrough verifies strikethrough rendering.
func TestDrawStrikethrough(t *testing.T) {
	outDir := snapshotOutputDir(t)
	img := image.NewRGBA(image.Rect(0, 0, 640, 420))
	fillImage(img, rgba(255, 255, 255, 255))

	// Draw strikethrough lines
	drawCPURect(img, rectF{50, 110, 200, 2}, rgba(0, 0, 0, 255))
	drawCPURect(img, rectF{50, 160, 300, 2}, rgba(255, 0, 0, 255))

	writePNG(t, filepath.Join(outDir, "strikethrough.png"), img)

	// Verify strikethrough pixels exist
	nonWhite := 0
	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			if r < 60000 || g < 60000 || b < 60000 {
				nonWhite++
			}
		}
	}
	if nonWhite < 20 {
		t.Errorf("DrawStrikethrough: expected at least 20 non-white pixels, got %d", nonWhite)
	}
}

// TestDrawFilledTriangle verifies triangle rendering.
func TestDrawFilledTriangle(t *testing.T) {
	outDir := snapshotOutputDir(t)
	img := image.NewRGBA(image.Rect(0, 0, 640, 420))
	fillImage(img, rgba(255, 255, 255, 255))

	// Draw a filled triangle
	p1 := math.NewVec2(320, 50)
	p2 := math.NewVec2(200, 200)
	p3 := math.NewVec2(440, 200)
	drawCPUTriangle(img, p1, p2, p3, rgba(22, 119, 255, 255))

	writePNG(t, filepath.Join(outDir, "filled_triangle.png"), img)

	// Verify triangle pixels exist
	nonWhite := 0
	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			if r < 60000 || g < 60000 || b < 60000 {
				nonWhite++
			}
		}
	}
	if nonWhite < 100 {
		t.Errorf("DrawFilledTriangle: expected at least 100 non-white pixels, got %d", nonWhite)
	}
}

// Helper: drawCPUDashedLine draws a dashed line on a CPU image.
func drawCPUDashedLine(img *image.RGBA, x1, y1, x2, y2, width, dashLen, gapLen float32, c color.RGBA) {
	dx := x2 - x1
	dy := y2 - y1
	length := float32(stdmath.Sqrt(float64(dx*dx + dy*dy)))
	if length < 0.001 {
		return
	}
	nx := dx / length
	ny := dy / length
	segmentLen := dashLen + gapLen
	dist := float32(0)
	for dist < length {
		segStart := dist
		segEnd := dist + dashLen
		if segEnd > length {
			segEnd = length
		}
		if segStart < length {
			sx := x1 + nx*segStart
			sy := y1 + ny*segStart
			ex := x1 + nx*segEnd
			_ = y1 + ny*segEnd // ey unused
			drawCPURect(img, rectF{sx, sy - width/2, ex - sx, width}, c)
		}
		dist += segmentLen
	}
}

// Helper: drawCPUArrow draws an arrow on a CPU image.
func drawCPUArrow(img *image.RGBA, cx, cy, size float32, direction int, c color.RGBA) {
	halfSize := size * 0.5
	var p1, p2, p3 math.Vec2
	switch direction {
	case 0: // Up
		p1 = math.NewVec2(cx, cy-halfSize)
		p2 = math.NewVec2(cx-halfSize, cy+halfSize)
		p3 = math.NewVec2(cx+halfSize, cy+halfSize)
	case 1: // Right
		p1 = math.NewVec2(cx+halfSize, cy)
		p2 = math.NewVec2(cx-halfSize, cy-halfSize)
		p3 = math.NewVec2(cx-halfSize, cy+halfSize)
	case 2: // Down
		p1 = math.NewVec2(cx, cy+halfSize)
		p2 = math.NewVec2(cx-halfSize, cy-halfSize)
		p3 = math.NewVec2(cx+halfSize, cy-halfSize)
	case 3: // Left
		p1 = math.NewVec2(cx-halfSize, cy)
		p2 = math.NewVec2(cx+halfSize, cy-halfSize)
		p3 = math.NewVec2(cx+halfSize, cy+halfSize)
	}
	drawCPUTriangle(img, p1, p2, p3, c)
}

// Helper: drawCPUTriangle draws a filled triangle on a CPU image.
func drawCPUTriangle(img *image.RGBA, p1, p2, p3 math.Vec2, c color.RGBA) {
	minX := int(min3(p1.X, p2.X, p3.X))
	maxX := int(max3(p1.X, p2.X, p3.X)) + 1
	minY := int(min3(p1.Y, p2.Y, p3.Y))
	maxY := int(max3(p1.Y, p2.Y, p3.Y)) + 1

	for y := maxInt(0, minY); y < minInt(img.Bounds().Dy(), maxY); y++ {
		for x := maxInt(0, minX); x < minInt(img.Bounds().Dx(), maxX); x++ {
			px := float32(x) + 0.5
			py := float32(y) + 0.5
			if pointInTriangle(math.NewVec2(px, py), p1, p2, p3) {
				blendPixel(img, x, y, c)
			}
		}
	}
}

func min3(a, b, c float32) float32 {
	m := a
	if b < m {
		m = b
	}
	if c < m {
		m = c
	}
	return m
}

func max3(a, b, c float32) float32 {
	m := a
	if b > m {
		m = b
	}
	if c > m {
		m = c
	}
	return m
}
