package pipeline

import (
	"testing"

	coremath "github.com/energye/examples/lcl/gpui/core/math"
)

func TestSVGIconCacheReusesGeometry(t *testing.T) {
	cache := NewSVGIconCache()
	viewBox := coremath.NewRect(0, 0, 100, 100)
	path := "M0 0 H100 V100 H0 Z"

	a, err := cache.Get(path, viewBox, FillRuleNonZero)
	if err != nil {
		t.Fatalf("cache Get returned error: %v", err)
	}
	b, err := cache.Get(path, viewBox, FillRuleNonZero)
	if err != nil {
		t.Fatalf("cache Get returned error: %v", err)
	}
	if a != b {
		t.Fatal("expected identical cache entry for same key")
	}
	if cache.Len() != 1 {
		t.Fatalf("cache len = %d, want 1", cache.Len())
	}
}

func TestSVGIconCacheSeparatesFillRuleAndViewBox(t *testing.T) {
	cache := NewSVGIconCache()
	path := "M0 0 H100 V100 H0 Z"

	if _, err := cache.Get(path, coremath.NewRect(0, 0, 100, 100), FillRuleNonZero); err != nil {
		t.Fatalf("cache Get returned error: %v", err)
	}
	if _, err := cache.Get(path, coremath.NewRect(0, 0, 100, 100), FillRuleEvenOdd); err != nil {
		t.Fatalf("cache Get returned error: %v", err)
	}
	if _, err := cache.Get(path, coremath.NewRect(0, 0, 1024, 1024), FillRuleNonZero); err != nil {
		t.Fatalf("cache Get returned error: %v", err)
	}
	if cache.Len() != 3 {
		t.Fatalf("cache len = %d, want 3", cache.Len())
	}
}
