package widget

import (
	"testing"

	"github.com/energye/examples/lcl/gpui/core/math"
	"github.com/energye/examples/lcl/gpui/style/token"
)

func TestButtonDefaultMeasureUsesTokenHeight(t *testing.T) {
	tokens := token.DefaultLight()
	button := NewButton("OK")

	size := button.Measure(&Context{Tokens: tokens}, Constraints{})
	if size.X != tokens.Alias.ControlHeight {
		t.Fatalf("width = %v, want min touch width %v", size.X, tokens.Alias.ControlHeight)
	}
	if size.Y != tokens.Alias.ControlHeight {
		t.Fatalf("height = %v, want %v", size.Y, tokens.Alias.ControlHeight)
	}
}

func TestButtonBlockMeasureUsesAvailableWidth(t *testing.T) {
	button := NewButton("Submit")
	button.Block = true

	size := button.Measure(nil, Constraints{Max: math.NewVec2(240, 0)})
	if size.X != 240 {
		t.Fatalf("block width = %v, want 240", size.X)
	}
}

func TestButtonPrimaryStyle(t *testing.T) {
	tokens := token.DefaultLight()
	button := NewButton("Primary")
	button.Kind = ButtonPrimary
	style := button.buttonStyle(&Context{Tokens: tokens})

	if style.Palette.Background != tokens.Global.ColorPrimary {
		t.Fatal("primary button should use primary background")
	}
	if style.Palette.Text != tokens.Global.ColorTextLight {
		t.Fatal("primary button should use light text")
	}
}

func TestButtonDangerStyle(t *testing.T) {
	tokens := token.DefaultLight()
	button := NewButton("Delete")
	button.Danger = true
	style := button.buttonStyle(&Context{Tokens: tokens})

	if style.Palette.Border != tokens.Global.ColorError {
		t.Fatal("danger button should use error border")
	}
	if style.Palette.Text != tokens.Global.ColorError {
		t.Fatal("danger default button should use error text")
	}
}

func TestButtonLinkStyle(t *testing.T) {
	tokens := token.DefaultLight()
	button := NewButton("Link")
	button.Kind = ButtonLink
	style := button.buttonStyle(&Context{Tokens: tokens})

	if style.Palette.Border.A != 0 || style.Palette.Background.A != 0 {
		t.Fatal("link button should not draw background or border")
	}
	if style.Palette.Text != tokens.Global.ColorPrimary {
		t.Fatal("link button should use primary text")
	}
}

func TestButtonPointerClick(t *testing.T) {
	button := NewButton("Save")
	clicks := 0
	button.SetOnClick(func() {
		clicks++
	})

	button.HandleEvent(nil, Event{Type: EventMouseDown, Button: 1})
	button.HandleEvent(nil, Event{Type: EventMouseUp, Button: 1})
	if clicks != 1 {
		t.Fatalf("clicks = %d, want 1", clicks)
	}
}

func TestButtonKeyboardClick(t *testing.T) {
	button := NewButton("Save")
	button.Focus()
	clicks := 0
	button.SetOnClick(func() {
		clicks++
	})

	if !button.HandleEvent(nil, Event{Type: EventKeyDown, Key: keyEnter}) {
		t.Fatal("focused enter should activate button")
	}
	if clicks != 1 {
		t.Fatalf("clicks = %d, want 1", clicks)
	}
}

func TestButtonLoadingBlocksClick(t *testing.T) {
	button := NewButton("Save")
	button.SetLoading(true)
	clicks := 0
	button.SetOnClick(func() {
		clicks++
	})

	if !button.HandleEvent(nil, Event{Type: EventMouseDown, Button: 1}) {
		t.Fatal("loading button should consume activation event")
	}
	button.HandleEvent(nil, Event{Type: EventMouseUp, Button: 1})
	if clicks != 0 {
		t.Fatalf("clicks = %d, want 0", clicks)
	}
}

func TestButtonMouseUpOutsideClearsActiveWithoutClick(t *testing.T) {
	root := NewContainer()
	root.Layout(nil, math.NewRect(0, 0, 200, 120))
	button := NewButton("Save")
	button.SetBounds(math.NewRect(10, 10, 80, 32))
	clicks := 0
	button.SetOnClick(func() {
		clicks++
	})
	root.Add(button)

	root.HandleEvent(nil, Event{Type: EventMouseDown, X: 20, Y: 20, Button: 1})
	if !button.HasState(StateActive) {
		t.Fatal("button should become active after mouse down")
	}
	root.HandleEvent(nil, Event{Type: EventMouseMove, X: 160, Y: 90})
	if !button.HasState(StateActive) {
		t.Fatal("button should stay active while captured pointer is pressed")
	}
	root.HandleEvent(nil, Event{Type: EventMouseUp, X: 160, Y: 90, Button: 1})
	if button.HasState(StateActive) {
		t.Fatal("button should remain inactive after outside mouse up")
	}
	if clicks != 0 {
		t.Fatalf("clicks = %d, want 0", clicks)
	}
}
