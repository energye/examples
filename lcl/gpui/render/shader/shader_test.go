package shader

import (
	"strings"
	"testing"
)

func TestLoadShaderWithoutGLReturnsError(t *testing.T) {
	sm := NewShaderManager()
	if shader, err := sm.LoadShader("test", "vertex", "fragment"); err == nil || shader != nil {
		t.Fatalf("LoadShader without GL = (%#v, %v), want nil error", shader, err)
	}
}

func TestShaderManagerSettersWithoutCurrentShaderAreSafe(t *testing.T) {
	sm := NewShaderManager()
	sm.UseShader(nil)
	sm.SetFloat("x", 1)
	sm.SetVec2("x", 1, 2)
	sm.SetVec4("x", 1, 2, 3, 4)
	sm.SetInt("x", 1)
	sm.SetMat4("x", nil)
	sm.Delete()
}

func TestRoundedRectShadersUseSoftScreenSpaceAntialiasing(t *testing.T) {
	rounded := BuiltinShaderSources["rounded_rect"][1]
	if !strings.Contains(rounded, "max(length(vec2(dFdx(d), dFdy(d))), 0.75)") {
		t.Fatal("rounded rect shader should use a minimum screen-space antialiasing width")
	}
	if !strings.Contains(rounded, "smoothstep(-aa, aa, d)") {
		t.Fatal("rounded rect shader should blend both sides of the SDF edge")
	}

	stroke := BuiltinShaderSources["rounded_rect_stroke"][1]
	if !strings.Contains(stroke, "strokeCenter = -halfWidth") {
		t.Fatal("rounded rect stroke should be centered on the shape edge")
	}
	if !strings.Contains(stroke, "abs(d - strokeCenter)") {
		t.Fatal("rounded rect stroke should use distance from the stroke centerline")
	}
}
