// Package widget provides UI widgets
package widget

import (
	"github.com/energye/examples/lcl/gpui/core/math"
	"github.com/energye/examples/lcl/gpui/render/font"
	"github.com/energye/examples/lcl/gpui/render/pipeline"
	"github.com/energye/examples/lcl/gpui/style/animation"
	"github.com/energye/examples/lcl/gpui/style/color"
	"github.com/energye/examples/lcl/gpui/style/theme"
)

// ButtonType represents button type
type ButtonType int

const (
	ButtonDefault ButtonType = iota
	ButtonPrimary
	ButtonSuccess
	ButtonWarning
	ButtonDanger
)

// ButtonSize represents button size
type ButtonSize int

const (
	ButtonSizeSM ButtonSize = iota
	ButtonSizeMD
	ButtonSizeLG
	ButtonSizeXL
)

// Button is a clickable button widget
type Button struct {
	BaseWidget

	// Properties
	text    string
	btnType ButtonType
	btnSize ButtonSize
	font    *font.Font

	// State
	hovered bool
	pressed bool

	// Animation
	hoverAnim *animation.Animation
	pressAnim *animation.Animation

	// Event handlers
	onClick func()
}

// NewButton creates a new button
func NewButton(text string, btnType ButtonType, f *font.Font) *Button {
	th := theme.GetTheme()

	btn := &Button{
		BaseWidget: NewBaseWidget(),
		text:       text,
		btnType:    btnType,
		btnSize:    ButtonSizeMD,
		font:       f,
	}

	// Create animations
	btn.hoverAnim = animation.NewAnimation(0, 1, th.DurationNormal, animation.EaseOut)
	btn.pressAnim = animation.NewAnimation(1, 0.8, th.DurationFast, animation.EaseIn)

	// Set size based on theme
	btn.SetSize(0, th.Button.HeightMD) // Width will be auto-calculated

	return btn
}

// Text returns the button text
func (b *Button) Text() string {
	return b.text
}

// SetText sets the button text
func (b *Button) SetText(text string) {
	b.text = text
}

// SetOnClick sets the click handler
func (b *Button) SetOnClick(handler func()) {
	b.onClick = handler
}

// Focusable returns true (button can receive focus)
func (b *Button) Focusable() bool {
	return true
}

// HandleEvent handles a generic UI event.
func (b *Button) HandleEvent(event UIEvent) bool {
	return dispatchLegacyEvent(b, event)
}

// Render renders the button
func (b *Button) Render(renderer *pipeline.Renderer) {
	if !b.visible {
		return
	}

	th := theme.GetTheme()

	// Calculate button width if auto
	if b.bounds.W <= 0 && b.font != nil {
		textW := b.font.TextWidth(b.text)
		paddingH := float32(16) // Fixed padding
		b.bounds.W = textW + paddingH*2
	}

	// Update animations (MUST call every frame)
	hoverT := b.hoverAnim.Value()
	pressT := b.pressAnim.Value()

	// Calculate colors
	bg, txt, border := b.calculateColors(hoverT, pressT)

	// Draw background
	renderer.FillRoundRect(b.bounds, th.Button.Radius, bg)

	// Draw border
	borderW := th.Button.BorderW
	renderer.StrokeRoundRect(b.bounds, th.Button.Radius, borderW, border)

	// Draw focus ring
	if b.HasState(StateFocus) {
		focusRect := b.bounds.Expand(2)
		renderer.StrokeRoundRect(focusRect, th.Button.Radius+2, 2, color.Primary)
	}

	// Draw text centered
	if b.text != "" && b.font != nil {
		textW, textH := b.font.MeasureText(b.text)
		textX := b.bounds.X + (b.bounds.W-textW)/2
		textY := b.bounds.Y + (b.bounds.H-textH)/2
		renderer.DrawText(b.text, textX, textY, b.font, txt)
	}
}

// calculateColors calculates the current colors based on state and animation
func (b *Button) calculateColors(hoverT, pressT float32) (bg, txt, border math.Color) {
	th := theme.GetTheme()

	// Get base colors for button type
	var baseBg, baseTxt, baseBorder math.Color
	switch b.btnType {
	case ButtonPrimary:
		baseBg = th.Button.Primary.Background
		baseTxt = th.Button.Primary.Text
		baseBorder = th.Button.Primary.Border
	case ButtonSuccess:
		baseBg = th.Button.Success.Background
		baseTxt = th.Button.Success.Text
		baseBorder = th.Button.Success.Border
	case ButtonWarning:
		baseBg = th.Button.Warning.Background
		baseTxt = th.Button.Warning.Text
		baseBorder = th.Button.Warning.Border
	case ButtonDanger:
		baseBg = th.Button.Danger.Background
		baseTxt = th.Button.Danger.Text
		baseBorder = th.Button.Danger.Border
	default:
		baseBg = th.Button.Default.Background
		baseTxt = th.Button.Default.Text
		baseBorder = th.Button.Default.Border
	}

	// Apply disabled state
	if b.HasState(StateDisabled) {
		baseBg = baseBg.WithAlpha(0.5)
		baseTxt = baseTxt.WithAlpha(0.5)
		baseBorder = baseBorder.WithAlpha(0.5)
		return baseBg, baseTxt, baseBorder
	}

	// Apply hover effect
	if hoverT > 0 {
		baseBg = baseBg.Lighten(0.05 * hoverT)
		baseBorder = baseBorder.Lighten(0.05 * hoverT)
	}

	// Apply press effect
	if pressT < 1 {
		baseBg = baseBg.Darken(0.1 * (1 - pressT))
	}

	return baseBg, baseTxt, baseBorder
}

// MouseDown handles mouse down
func (b *Button) MouseDown(x, y float32, button int) bool {
	if !b.enabled || !b.bounds.Contains(x, y) {
		return false
	}

	b.pressed = true
	b.SetStateFlag(StateActive, true)
	b.pressAnim.PlayForward()
	return true
}

// MouseUp handles mouse up
func (b *Button) MouseUp(x, y float32, button int) bool {
	if b.pressed && b.hovered && b.enabled {
		if b.onClick != nil {
			b.onClick()
		}
	}

	b.pressed = false
	b.SetStateFlag(StateActive, false)
	b.pressAnim.PlayReverse()
	return true
}

// MouseMove handles mouse move
func (b *Button) MouseMove(x, y float32) bool {
	wasHovered := b.hovered
	b.hovered = b.bounds.Contains(x, y) && b.enabled
	b.SetStateFlag(StateHover, b.hovered)

	if b.hovered != wasHovered {
		if b.hovered {
			b.hoverAnim.PlayForward()
		} else {
			b.hoverAnim.PlayReverse()
		}
	}

	return b.hovered || wasHovered
}

// KeyDown handles key down
func (b *Button) KeyDown(key int, mods int) bool {
	if !b.focused {
		return false
	}

	// Enter or Space triggers click
	if key == 13 || key == 32 { // Enter or Space
		if b.enabled && b.onClick != nil {
			b.onClick()
		}
		return true
	}

	return false
}
