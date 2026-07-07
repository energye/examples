package shader

import "testing"

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
