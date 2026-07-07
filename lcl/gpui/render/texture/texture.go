package texture

import (
	"image"
	"image/draw"
	"unsafe"

	"github.com/energye/examples/lcl/gpui/core/gl"
)

// Texture wraps an OpenGL 2D texture resource.
type Texture struct {
	ID     uint32
	Width  int
	Height int
}

// NewFromImage uploads an image as an RGBA texture.
func NewFromImage(img image.Image) *Texture {
	if img == nil {
		return nil
	}

	rgba := toRGBA(img)
	bounds := rgba.Bounds()

	var id uint32
	gl.GenTextures(1, &id)
	gl.BindTexture(gl.GL_TEXTURE_2D, id)
	gl.TexParameteri(gl.GL_TEXTURE_2D, gl.GL_TEXTURE_MIN_FILTER, gl.GL_LINEAR)
	gl.TexParameteri(gl.GL_TEXTURE_2D, gl.GL_TEXTURE_MAG_FILTER, gl.GL_LINEAR)
	gl.TexParameteri(gl.GL_TEXTURE_2D, gl.GL_TEXTURE_WRAP_S, gl.GL_CLAMP_TO_EDGE)
	gl.TexParameteri(gl.GL_TEXTURE_2D, gl.GL_TEXTURE_WRAP_T, gl.GL_CLAMP_TO_EDGE)
	gl.TexImage2D(
		gl.GL_TEXTURE_2D,
		0,
		int32(gl.GL_RGBA),
		int32(bounds.Dx()),
		int32(bounds.Dy()),
		0,
		gl.GL_RGBA,
		gl.GL_UNSIGNED_BYTE,
		unsafePtr(rgba.Pix),
	)

	return &Texture{ID: id, Width: bounds.Dx(), Height: bounds.Dy()}
}

// Update replaces the texture contents.
func (t *Texture) Update(img image.Image) {
	if t == nil || t.ID == 0 || img == nil {
		return
	}

	rgba := toRGBA(img)
	bounds := rgba.Bounds()

	gl.BindTexture(gl.GL_TEXTURE_2D, t.ID)
	gl.TexImage2D(
		gl.GL_TEXTURE_2D,
		0,
		int32(gl.GL_RGBA),
		int32(bounds.Dx()),
		int32(bounds.Dy()),
		0,
		gl.GL_RGBA,
		gl.GL_UNSIGNED_BYTE,
		unsafePtr(rgba.Pix),
	)
	t.Width = bounds.Dx()
	t.Height = bounds.Dy()
}

// Delete releases the OpenGL texture.
func (t *Texture) Delete() {
	if t == nil || t.ID == 0 {
		return
	}
	gl.DeleteTextures(1, &t.ID)
	t.ID = 0
	t.Width = 0
	t.Height = 0
}

func toRGBA(img image.Image) *image.RGBA {
	if rgba, ok := img.(*image.RGBA); ok {
		return rgba
	}

	bounds := img.Bounds()
	rgba := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	draw.Draw(rgba, rgba.Bounds(), img, bounds.Min, draw.Src)
	return rgba
}

func unsafePtr(p []byte) uintptr {
	if len(p) == 0 {
		return 0
	}
	return uintptr(unsafe.Pointer(&p[0]))
}
