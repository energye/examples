package math

import (
	"testing"
)

func TestMat4MultiplyIdentity(t *testing.T) {
	m := TranslationMatrix(10, 20, 30)
	result := m.Multiply(IdentityMatrix())
	for i := range result {
		if result[i] != m[i] {
			t.Fatalf("M * I should equal M, got diff at index %d: %f != %f", i, result[i], m[i])
		}
	}
}

func TestMat4MultiplyOrder(t *testing.T) {
	// A = translate(10, 0, 0), B = scale(2, 2, 1)
	// A*B: scale first then translate => point (5,0) -> (10,0) -> (20,0)
	// B*A: translate first then scale => point (5,0) -> (15,0) -> (30,0)
	A := TranslationMatrix(10, 0, 0)
	B := ScaleMatrix(2, 2, 1)

	result := A.Multiply(B)

	// Transform point (5, 0) using result
	px := result[0]*5 + result[4]*0 + result[12]
	py := result[1]*5 + result[5]*0 + result[13]

	// A*B applied to (5,0): scale by 2 -> (10,0), translate by 10 -> (20,0)
	if px != 20 || py != 0 {
		t.Fatalf("A*B applied to (5,0) should be (20,0), got (%f, %f)", px, py)
	}
}

func TestMat4MultiplyTranslation(t *testing.T) {
	// T1 * T2 should compose translations
	T1 := TranslationMatrix(10, 20, 0)
	T2 := TranslationMatrix(30, 40, 0)
	result := T1.Multiply(T2)

	// Should be equivalent to translate(40, 60, 0)
	if result[12] != 40 || result[13] != 60 {
		t.Fatalf("T1*T2 translation should be (40,60), got (%f, %f)", result[12], result[13])
	}
}

func TestMat4MultiplyScale(t *testing.T) {
	S1 := ScaleMatrix(2, 3, 1)
	S2 := ScaleMatrix(4, 5, 1)
	result := S1.Multiply(S2)

	// Should be equivalent to scale(8, 15, 1)
	if result[0] != 8 || result[5] != 15 {
		t.Fatalf("S1*S2 scale should be (8,15), got (%f, %f)", result[0], result[5])
	}
}

func TestMat4TransformPoint(t *testing.T) {
	// Translate by (100, 50, 0) then scale by (2, 2, 1)
	// Apply to point (10, 10): scale first -> (20, 20), translate -> (120, 70)
	T := TranslationMatrix(100, 50, 0)
	S := ScaleMatrix(2, 2, 1)
	TS := T.Multiply(S)

	x := TS[0]*10 + TS[4]*10 + TS[12]
	y := TS[1]*10 + TS[5]*10 + TS[13]

	if x != 120 || y != 70 {
		t.Fatalf("T*S applied to (10,10) should be (120,70), got (%f, %f)", x, y)
	}
}

func TestMat4TransformPointReverse(t *testing.T) {
	// Scale by (2, 2, 1) then translate by (100, 50, 0)
	// Apply to point (10, 10): translate first -> (110, 60), scale -> (220, 120)
	S := ScaleMatrix(2, 2, 1)
	T := TranslationMatrix(100, 50, 0)
	ST := S.Multiply(T)

	x := ST[0]*10 + ST[4]*10 + ST[12]
	y := ST[1]*10 + ST[5]*10 + ST[13]

	if x != 220 || y != 120 {
		t.Fatalf("S*T applied to (10,10) should be (220,120), got (%f, %f)", x, y)
	}
}

func TestColorToHSL(t *testing.T) {
	// Red: HSL(0, 1, 0.5)
	red := NewColor(1, 0, 0, 1)
	hsl := red.ToHSL()
	if hsl.H != 0 || hsl.S != 1 || hsl.L != 0.5 {
		t.Fatalf("Red HSL should be (0, 1, 0.5), got (%f, %f, %f)", hsl.H, hsl.S, hsl.L)
	}

	// Green: HSL(120, 1, 0.5)
	green := NewColor(0, 1, 0, 1)
	hsl = green.ToHSL()
	if hsl.H != 120 || hsl.S != 1 || hsl.L != 0.5 {
		t.Fatalf("Green HSL should be (120, 1, 0.5), got (%f, %f, %f)", hsl.H, hsl.S, hsl.L)
	}

	// Blue: HSL(240, 1, 0.5)
	blue := NewColor(0, 0, 1, 1)
	hsl = blue.ToHSL()
	if hsl.H != 240 || hsl.S != 1 || hsl.L != 0.5 {
		t.Fatalf("Blue HSL should be (240, 1, 0.5), got (%f, %f, %f)", hsl.H, hsl.S, hsl.L)
	}

	// White: HSL(0, 0, 1)
	white := NewColor(1, 1, 1, 1)
	hsl = white.ToHSL()
	if hsl.S != 0 || hsl.L != 1 {
		t.Fatalf("White HSL should have S=0, L=1, got S=%f, L=%f", hsl.S, hsl.L)
	}

	// Black: HSL(0, 0, 0)
	black := NewColor(0, 0, 0, 1)
	hsl = black.ToHSL()
	if hsl.S != 0 || hsl.L != 0 {
		t.Fatalf("Black HSL should have S=0, L=0, got S=%f, L=%f", hsl.S, hsl.L)
	}
}

func TestNewColorFromHSL(t *testing.T) {
	// Red from HSL
	red := NewColorFromHSL(0, 1, 0.5, 1)
	if abs(red.R-1) > 0.01 || red.G > 0.01 || red.B > 0.01 {
		t.Fatalf("HSL(0,1,0.5) should be red, got (%f, %f, %f)", red.R, red.G, red.B)
	}

	// Green from HSL
	green := NewColorFromHSL(120, 1, 0.5, 1)
	if green.R > 0.01 || abs(green.G-1) > 0.01 || green.B > 0.01 {
		t.Fatalf("HSL(120,1,0.5) should be green, got (%f, %f, %f)", green.R, green.G, green.B)
	}

	// Roundtrip: RGB -> HSL -> RGB
	original := NewColor(0.2, 0.6, 0.8, 1)
	hsl := original.ToHSL()
	roundtrip := NewColorFromHSL(hsl.H, hsl.S, hsl.L, original.A)
	if abs(original.R-roundtrip.R) > 0.01 || abs(original.G-roundtrip.G) > 0.01 || abs(original.B-roundtrip.B) > 0.01 {
		t.Fatalf("Roundtrip failed: (%f,%f,%f) -> HSL(%f,%f,%f) -> (%f,%f,%f)",
			original.R, original.G, original.B, hsl.H, hsl.S, hsl.L, roundtrip.R, roundtrip.G, roundtrip.B)
	}
}

func TestLightenHSL(t *testing.T) {
	blue := NewColor(0, 0, 1, 1)
	lighter := blue.LightenHSL(0.2)

	// Lightening blue should increase L while keeping H and S
	hsl := lighter.ToHSL()
	if hsl.L <= 0.5 {
		t.Fatalf("Lightened blue should have L > 0.5, got %f", hsl.L)
	}
	if hsl.H != 240 {
		t.Fatalf("Lightened blue should keep H=240, got %f", hsl.H)
	}
}

func TestSaturate(t *testing.T) {
	gray := NewColor(0.5, 0.5, 0.5, 1)
	saturated := gray.Saturate(0.5)

	hsl := saturated.ToHSL()
	if hsl.S <= 0 {
		t.Fatalf("Saturated gray should have S > 0, got %f", hsl.S)
	}
}

func TestHueRotate(t *testing.T) {
	red := NewColor(1, 0, 0, 1)
	rotated := red.HueRotate(120)

	hsl := rotated.ToHSL()
	// Rotating red by 120° should give green
	if abs(hsl.H-120) > 1 {
		t.Fatalf("Red rotated 120° should have H≈120, got %f", hsl.H)
	}
}

func TestMat4Transpose(t *testing.T) {
	m := Mat4{
		1, 2, 3, 4,
		5, 6, 7, 8,
		9, 10, 11, 12,
		13, 14, 15, 16,
	}
	tr := m.Transpose()

	// Column-major: m[col*4+row]
	// Original: row 0 = [1, 5, 9, 13], row 1 = [2, 6, 10, 14], ...
	// Transposed: row 0 = [1, 2, 3, 4], row 1 = [5, 6, 7, 8], ...
	if tr[0] != 1 || tr[1] != 5 || tr[4] != 2 || tr[5] != 6 {
		t.Fatalf("Transpose incorrect: got %v", tr)
	}
}

func TestMat4Inverse(t *testing.T) {
	// Test inverse of translation matrix
	T := TranslationMatrix(10, 20, 30)
	TInv := T.Inverse()
	I := T.Multiply(TInv)

	// Should be identity (approximately)
	for i := 0; i < 16; i++ {
		expected := float32(0)
		if i%5 == 0 { // diagonal
			expected = 1
		}
		if abs(I[i]-expected) > 0.001 {
			t.Fatalf("T * T^-1 should be identity, got diff at index %d: %f vs %f", i, I[i], expected)
		}
	}
}

func TestMat4InverseScale(t *testing.T) {
	// Test inverse of scale matrix
	S := ScaleMatrix(2, 4, 8)
	SInv := S.Inverse()
	I := S.Multiply(SInv)

	for i := 0; i < 16; i++ {
		expected := float32(0)
		if i%5 == 0 {
			expected = 1
		}
		if abs(I[i]-expected) > 0.001 {
			t.Fatalf("S * S^-1 should be identity at index %d", i)
		}
	}
}

func TestMat4InverseIdentity(t *testing.T) {
	I := IdentityMatrix()
	IInv := I.Inverse()

	for i := 0; i < 16; i++ {
		if abs(I[i]-IInv[i]) > 0.001 {
			t.Fatalf("Identity inverse should be identity at index %d", i)
		}
	}
}

func TestShearMatrix(t *testing.T) {
	// Shear matrix should preserve the homogeneous coordinate
	sh := ShearMatrix(0.5, 0, 0, 0, 0, 0)
	if sh[15] != 1 {
		t.Fatalf("Shear matrix w component should be 1, got %f", sh[15])
	}
}

func abs(x float32) float32 {
	if x < 0 {
		return -x
	}
	return x
}
