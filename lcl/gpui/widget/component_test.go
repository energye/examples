package widget

import (
	"testing"

	"github.com/energye/examples/lcl/gpui/core/math"
	"github.com/energye/examples/lcl/gpui/style/token"
)

func TestComponentBaseDefaultControlStyle(t *testing.T) {
	tokens := token.DefaultLight()
	c := NewComponentBase()
	style := c.ResolveControlStyle(&Context{Tokens: tokens})

	if style.Metrics.Height != tokens.Alias.ControlHeight {
		t.Fatalf("height = %v, want %v", style.Metrics.Height, tokens.Alias.ControlHeight)
	}
	if style.Metrics.FontSize != tokens.Global.FontSize {
		t.Fatalf("font size = %v, want %v", style.Metrics.FontSize, tokens.Global.FontSize)
	}
	if style.Palette.Background != tokens.Global.ColorBgContainer {
		t.Fatal("outlined background should use container background")
	}
	if style.Palette.Border != tokens.Global.ColorBorder {
		t.Fatal("outlined border should use global border")
	}
}

func TestComponentBaseControlSizes(t *testing.T) {
	tokens := token.DefaultLight()
	c := NewComponentBase()

	c.SetControlSize(SizeSmall)
	small := c.ResolveControlStyle(&Context{Tokens: tokens})
	if small.Metrics.Height != tokens.Alias.ControlHeightSM || small.Metrics.FontSize != tokens.Global.FontSizeSM {
		t.Fatal("small metrics should use small token values")
	}

	c.SetControlSize(SizeLarge)
	large := c.ResolveControlStyle(&Context{Tokens: tokens})
	if large.Metrics.Height != tokens.Alias.ControlHeightLG || large.Metrics.FontSize != tokens.Global.FontSizeLG {
		t.Fatal("large metrics should use large token values")
	}
	if large.Metrics.Radius != tokens.Global.RadiusLG {
		t.Fatal("large radius should use RadiusLG")
	}
}

func TestComponentBaseStatusPalette(t *testing.T) {
	tokens := token.DefaultLight()
	c := NewComponentBase()
	c.SetStatus(StatusError)
	style := c.ResolveControlStyle(&Context{Tokens: tokens})

	if style.Palette.Border != tokens.Global.ColorError {
		t.Fatal("error status should use error border")
	}
	if style.Palette.FocusBorder != tokens.Global.ColorError {
		t.Fatal("error status should use error focus border")
	}
	if style.Palette.StatusColor != tokens.Global.ColorError {
		t.Fatal("error status should expose error status color")
	}
}

func TestComponentBaseVariantSolidAndStates(t *testing.T) {
	tokens := token.DefaultLight()
	c := NewComponentBase()
	c.SetVariant(VariantSolid)
	style := c.ResolveControlStyle(&Context{Tokens: tokens})

	if style.Palette.Background != tokens.Global.ColorPrimary {
		t.Fatal("solid default background should use primary color")
	}
	if style.Palette.Text != tokens.Global.ColorTextLight {
		t.Fatal("solid text should use light text color")
	}

	c.SetStateFlag(StateActive, true)
	active := c.ResolveControlStyle(&Context{Tokens: tokens})
	if active.Palette.Background == style.Palette.Background {
		t.Fatal("active solid background should differ from normal background")
	}
}

func TestComponentBaseDisabledPaletteOverridesVariant(t *testing.T) {
	tokens := token.DefaultLight()
	c := NewComponentBase()
	c.SetVariant(VariantSolid)
	c.SetEnabled(false)
	style := c.ResolveControlStyle(&Context{Tokens: tokens})

	if style.Palette.Text != tokens.Global.ColorTextDisabled {
		t.Fatal("disabled text should use disabled token")
	}
	if style.Palette.Background != tokens.Alias.ColorFillAlter {
		t.Fatal("disabled background should use fill alter")
	}
	if style.Palette.Border.A <= 0 {
		t.Fatal("disabled border should remain visible")
	}
}

func TestComponentBaseTokenOverride(t *testing.T) {
	tokens := token.DefaultLight()
	override := token.DefaultLight()
	override.Global.ColorPrimary = math.NewColor(1, 0, 0, 1)

	c := NewComponentBase()
	c.SetVariant(VariantSolid)
	c.SetTokenOverride(&override)
	style := c.ResolveControlStyle(&Context{Tokens: tokens})

	if style.Palette.Background != override.Global.ColorPrimary {
		t.Fatal("token override should take precedence over context tokens")
	}
}

func TestControlStyleBoxStyle(t *testing.T) {
	tokens := token.DefaultLight()
	c := NewComponentBase()
	style := c.ResolveControlStyle(&Context{Tokens: tokens})
	box := style.BoxStyle()

	if box.Background != style.Palette.Background {
		t.Fatal("box background should come from control palette")
	}
	if box.BorderColor != style.Palette.Border {
		t.Fatal("box border should come from control palette")
	}
	if box.Radius != style.Metrics.Radius {
		t.Fatal("box radius should come from control metrics")
	}
}
