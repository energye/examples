package token

import (
	"sync"
	"testing"

	"github.com/energye/examples/lcl/gpui/core/math"
)

func TestDeriveLightTokens(t *testing.T) {
	tokens := DefaultLight()
	if tokens.Mode != ModeLight {
		t.Fatalf("mode = %v, want light", tokens.Mode)
	}
	if tokens.Global.ColorPrimary != tokens.Seed.ColorPrimary {
		t.Fatal("global primary should come from seed")
	}
	if tokens.Global.SpaceMD != tokens.Seed.SizeUnit*4 {
		t.Fatalf("SpaceMD = %v, want %v", tokens.Global.SpaceMD, tokens.Seed.SizeUnit*4)
	}
	if tokens.Components.Button.Height != tokens.Alias.ControlHeight {
		t.Fatal("button height should use control height alias")
	}
}

func TestDeriveDarkTokens(t *testing.T) {
	light := DefaultLight()
	dark := DefaultDark()
	if dark.Mode != ModeDark {
		t.Fatalf("mode = %v, want dark", dark.Mode)
	}
	if light.Global.ColorBgBase == dark.Global.ColorBgBase {
		t.Fatal("dark bg base should differ from light bg base")
	}
	if dark.Global.ColorText.A <= 0 {
		t.Fatal("dark text should be visible")
	}
}

func TestDeriveNormalizesPartialSeed(t *testing.T) {
	primary := math.NewColor(0.1, 0.2, 0.3, 1)
	tokens := Derive(SeedToken{ColorPrimary: primary}, ModeLight)

	if tokens.Seed.ColorPrimary != primary {
		t.Fatal("explicit primary color should be preserved")
	}
	if tokens.Seed.FontSize != DefaultSeed().FontSize || tokens.Global.FontSize <= 0 {
		t.Fatalf("font size = %v, want default positive size", tokens.Global.FontSize)
	}
	if tokens.Global.RadiusSM < 0 {
		t.Fatalf("small radius = %v, want non-negative default-derived radius", tokens.Global.RadiusSM)
	}
	if tokens.Components.Button.Height <= 0 {
		t.Fatalf("button height = %v, want positive control height", tokens.Components.Button.Height)
	}
}

func TestCurrentTokens(t *testing.T) {
	Reset()
	light := Current()
	dark := DefaultDark()
	SetCurrent(dark)
	if Current().Mode != ModeDark {
		t.Fatal("current mode should be dark")
	}
	SetCurrent(light)
	if Current().Mode != ModeLight {
		t.Fatal("current mode should be light")
	}
}

func TestCurrentTokensConcurrentAccess(t *testing.T) {
	Reset()
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			_ = Current()
		}()
		go func(i int) {
			defer wg.Done()
			if i%2 == 0 {
				SetCurrent(DefaultLight())
			} else {
				SetCurrent(DefaultDark())
			}
		}(i)
	}
	wg.Wait()
	Reset()
}
