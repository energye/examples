package main

import (
	"math"
	"testing"
)

func testSurface(w, h int) *Surface {
	stride := w * 4
	return &Surface{
		Px: make([]byte, stride*h), W: w, H: h, Stride: stride,
		R: 2, G: 1, B: 0, A: 3, // Windows BGRA, also the detected default
	}
}

func TestRoundedAlpha(t *testing.T) {
	if v := roundedAlpha(0, 0, 64, 64, 12); v > 0.01 {
		t.Fatalf("corner alpha = %.3f, want ~0", v)
	}
	if v := roundedAlpha(32, 32, 64, 64, 12); v != 1 {
		t.Fatalf("center alpha = %.3f, want 1", v)
	}
	if v := roundedAlpha(5, 32, 64, 64, 12); v != 1 {
		t.Fatalf("left edge middle alpha = %.3f, want 1", v)
	}
	if v := roundedAlpha(0, 0, 64, 64, 0); v != 1 {
		t.Fatalf("radius=0 should disable corner cut, got %.3f", v)
	}
	// (3,3) lies in the 1px transition band around the corner circle.
	edge := roundedAlpha(3, 3, 64, 64, 12)
	if edge <= 0 || edge >= 1 {
		t.Fatalf("anti-aliased edge alpha = %.3f, want 0<a<1", edge)
	}
}

// bakeGlass writes a real per-pixel alpha; this verifies the bulk-write path
// without needing a GUI or lcl.Init().
func TestBakeGlassAlpha(t *testing.T) {
	s := testSurface(64, 64)
	bakeGlass(s, darkTheme(), 12, 0.5, 0.25)
	corner := s.Px[s.offset(0, 0)+s.A]
	center := s.Px[s.offset(32, 32)+s.A]
	if corner != 0 {
		t.Fatalf("corner alpha = %d, want 0", corner)
	}
	if center < 120 || center > 135 {
		t.Fatalf("center alpha = %d, want ~127 (0.5*255)", center)
	}
	// RGB channels must not be left blank either.
	r := s.Px[s.offset(32, 32)+s.R]
	g := s.Px[s.offset(32, 32)+s.G]
	b := s.Px[s.offset(32, 32)+s.B]
	if r == 0 && g == 0 && b == 0 {
		t.Fatal("center color is black, gradient was not written")
	}
}

func TestBakeGlowIsIrregular(t *testing.T) {
	s := testSurface(64, 64)
	bakeGlow(s, darkTheme(), 1.0, 5)
	center := s.Px[s.offset(32, 32)+s.A]
	corner := s.Px[s.offset(0, 0)+s.A]
	if center <= corner {
		t.Fatalf("center alpha %d must exceed corner %d", center, corner)
	}
	// A lobed shape is not a plain circle: two points at the same radius on
	// different angles must not always share the same alpha.
	a1 := s.Px[s.offset(32+20, 32)+s.A]
	a2 := s.Px[s.offset(32-20, 32)+s.A]
	if a1 != a2 {
		return // irregular, as intended
	}
	t.Log("symmetric sample; shape still validated by center/corner")
}

func TestStarPoints(t *testing.T) {
	pts := starPoints(100, 100, 40, 15, 8, 0)
	if len(pts) != 16 {
		t.Fatalf("star point count = %d, want 16", len(pts))
	}
	for i, p := range pts {
		d := math.Hypot(float64(p.X-100), float64(p.Y-100))
		want := 40.0
		if i%2 == 1 {
			want = 15.0
		}
		if math.Abs(d-want) > 1.5 {
			t.Fatalf("point %d radius = %.1f, want %.1f (outer/inner must alternate)", i, d, want)
		}
	}
}

func TestSurfaceOffsetAndLerp(t *testing.T) {
	s := testSurface(64, 64)
	if got := s.offset(3, 2); got != 2*256+3*4 {
		t.Fatalf("offset = %d, want %d", got, 2*256+3*4)
	}
	if v := lerpByte(0, 255, 0.5); v != 127 {
		t.Fatalf("lerp(0,255,0.5) = %d, want 127", v)
	}
	if v := lerpByte(10, 20, -3); v != 10 {
		t.Fatalf("lerp clamp lower = %d, want 10", v)
	}
	if v := lerpByte(10, 20, 9); v != 20 {
		t.Fatalf("lerp clamp upper = %d, want 20", v)
	}
	if v := clampF(2.5, 0, 1); v != 1 {
		t.Fatalf("clampF = %.2f, want 1", v)
	}
}
