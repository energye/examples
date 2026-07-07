package pipeline

import (
	"image"
	"image/png"
	"os"
	"path/filepath"
	"unsafe"

	"github.com/energye/examples/lcl/gpui/core/gl"
)

// CaptureRGBA reads the current framebuffer into a top-left-origin RGBA image.
func (r *Renderer) CaptureRGBA() *image.RGBA {
	if r == nil {
		return image.NewRGBA(image.Rect(0, 0, 1, 1))
	}
	width := int(r.width)
	height := int(r.height)
	if width <= 0 || height <= 0 {
		return image.NewRGBA(image.Rect(0, 0, 1, 1))
	}
	if gl.ReadPixels == nil {
		return image.NewRGBA(image.Rect(0, 0, 1, 1))
	}

	raw := make([]byte, width*height*4)
	gl.ReadPixels(0, 0, int32(width), int32(height), gl.GL_RGBA, gl.GL_UNSIGNED_BYTE, uintptr(unsafe.Pointer(&raw[0])))

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	stride := width * 4
	for y := 0; y < height; y++ {
		srcStart := (height - 1 - y) * stride
		dstStart := y * img.Stride
		copy(img.Pix[dstStart:dstStart+stride], raw[srcStart:srcStart+stride])
	}
	return img
}

// SavePNG writes the current framebuffer to a PNG file.
func (r *Renderer) SavePNG(path string) error {
	img := r.CaptureRGBA()
	dir := filepath.Dir(path)
	if dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}
