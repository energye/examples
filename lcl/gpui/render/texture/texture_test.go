package texture

import (
	"image"
	"testing"
)

func TestNewFromImageWithoutGLIsSafe(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	if tex := NewFromImage(img); tex != nil {
		t.Fatalf("texture = %#v, want nil without GL functions", tex)
	}
}

func TestTextureUpdateAndDeleteWithoutGLAreSafe(t *testing.T) {
	tex := &Texture{ID: 1, Width: 2, Height: 2}
	tex.Update(image.NewRGBA(image.Rect(0, 0, 4, 4)))
	if tex.Width != 2 || tex.Height != 2 {
		t.Fatalf("texture size changed without GL: %dx%d", tex.Width, tex.Height)
	}
	tex.Delete()
	if tex.ID != 1 {
		t.Fatalf("texture id = %d, want unchanged without GL delete", tex.ID)
	}
}
