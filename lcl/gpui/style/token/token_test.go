package token

import "testing"

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
