package pipeline

import (
	"testing"

	coremath "github.com/energye/examples/lcl/gpui/core/math"
	"github.com/energye/examples/lcl/gpui/render/shader"
)

func TestBatchSplitsWhenCapacityWouldOverflow(t *testing.T) {
	bm := NewBatchManager(4, 6)
	shaderProg := &shader.ShaderProgram{ID: 1, Name: "test"}
	verts := QuadVertices(
		coremath.NewRect(0, 0, 10, 10),
		coremath.NewRect(0, 0, 1, 1),
		coremath.NewColor(1, 1, 1, 1),
	)

	bm.AddQuad(shaderProg, 0, verts)
	bm.AddQuad(shaderProg, 0, verts)

	if len(bm.batches) != 1 {
		t.Fatalf("completed batches = %d, want 1", len(bm.batches))
	}
	if bm.current == nil {
		t.Fatal("expected current batch")
	}
	if len(bm.batches[0].Verts) != 4 || len(bm.current.Verts) != 4 {
		t.Fatalf("batch vertex counts = %d and %d, want 4 and 4", len(bm.batches[0].Verts), len(bm.current.Verts))
	}
}

func TestBatchManagerNilAndTinyCapacityAreSafe(t *testing.T) {
	var bm *BatchManager
	shaderProg := &shader.ShaderProgram{ID: 1, Name: "test"}
	verts := QuadVertices(
		coremath.NewRect(0, 0, 10, 10),
		coremath.NewRect(0, 0, 1, 1),
		coremath.NewColor(1, 1, 1, 1),
	)

	bm.AddQuad(shaderProg, 0, verts)
	bm.Reset()

	tiny := NewBatchManager(2, 2)
	tiny.AddQuad(shaderProg, 0, verts)
	if len(tiny.batches) != 0 {
		t.Fatalf("completed batches = %d, want no empty completed batch", len(tiny.batches))
	}
	if tiny.current == nil || len(tiny.current.Verts) != 4 {
		t.Fatalf("current batch should contain the oversized primitive")
	}
}

func TestNilRendererAccessorsAreSafe(t *testing.T) {
	var r *Renderer
	r.BeginFrame(100, 100)
	r.EndFrame()
	r.Flush()
	r.Delete()
	r.PushTransform(coremath.IdentityMatrix())
	r.PopTransform()
	r.PushClip(coremath.NewRect(0, 0, 10, 10))
	r.PopClip()
	if _, ok := r.CurrentTransform(); ok {
		t.Fatal("nil renderer should not have a transform")
	}
	if _, ok := r.CurrentClip(); ok {
		t.Fatal("nil renderer should not have a clip")
	}
	if r.ShaderManager() != nil {
		t.Fatal("nil renderer shader manager should be nil")
	}
	if r.ProjectionMatrix() != nil {
		t.Fatal("nil renderer projection matrix should be nil")
	}
}

func TestBeginFrameWithoutInitResetsCPUState(t *testing.T) {
	r := NewRenderer()
	r.clipStack = append(r.clipStack, coremath.NewRect(0, 0, 10, 10))
	r.transformStack = append(r.transformStack, coremath.IdentityMatrix())
	r.batch.current = &Batch{Verts: []Vertex{{X: 1}}, Indices: []uint32{0}}

	r.BeginFrame(320, 240)

	if r.width != 320 || r.height != 240 {
		t.Fatalf("size = (%v,%v), want (320,240)", r.width, r.height)
	}
	if len(r.clipStack) != 0 {
		t.Fatalf("clip stack length = %d, want 0", len(r.clipStack))
	}
	if len(r.transformStack) != 0 {
		t.Fatalf("transform stack length = %d, want 0", len(r.transformStack))
	}
	if r.batch.current != nil || len(r.batch.batches) != 0 {
		t.Fatal("batch state should be reset")
	}
}

func TestUniformFastKeyDistinguishesFractionalValues(t *testing.T) {
	a := UniformSet{"uLineWidth": FloatUniform(0.25)}
	b := UniformSet{"uLineWidth": FloatUniform(0.75)}
	if a.fastKey() == b.fastKey() {
		t.Fatal("fractional uniform values should not collide by integer truncation")
	}
}

func TestUniformFastKeyIsOrderIndependent(t *testing.T) {
	a := UniformSet{
		"uLineStart": Vec2Uniform(0.25, 1.5),
		"uLineEnd":   Vec2Uniform(100.75, 40.125),
		"uLineWidth": FloatUniform(2.5),
	}
	b := UniformSet{
		"uLineWidth": FloatUniform(2.5),
		"uLineEnd":   Vec2Uniform(100.75, 40.125),
		"uLineStart": Vec2Uniform(0.25, 1.5),
	}
	if a.fastKey() != b.fastKey() {
		t.Fatal("equivalent uniform sets should have the same fast key")
	}
}

func TestCaptureRGBAWithoutReadPixelsIsSafe(t *testing.T) {
	r := NewRenderer()
	r.width = 16
	r.height = 16
	img := r.CaptureRGBA()
	if img.Bounds().Dx() != 1 || img.Bounds().Dy() != 1 {
		t.Fatalf("capture size = %dx%d, want 1x1 without ReadPixels", img.Bounds().Dx(), img.Bounds().Dy())
	}
}
