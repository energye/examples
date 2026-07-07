package widget

import (
	"fmt"

	"github.com/energye/examples/lcl/gpui/core/math"
	"github.com/energye/examples/lcl/gpui/render/font"
	"github.com/energye/examples/lcl/gpui/render/pipeline"
	"github.com/energye/examples/lcl/gpui/style/animation"
	"github.com/energye/examples/lcl/gpui/style/color"
	"github.com/energye/examples/lcl/gpui/style/theme"
)

// Checkbox is a checkbox widget
type Checkbox struct {
	BaseWidget
	checked   bool
	text      string
	font      *font.Font
	checkAnim *animation.Animation
	onClick   func()
}

// NewCheckbox creates a new checkbox
func NewCheckbox(text string, f *font.Font) *Checkbox {
	th := theme.GetTheme()

	return &Checkbox{
		BaseWidget: NewBaseWidget(),
		text:       text,
		font:       f,
		checkAnim:  animation.NewAnimation(0, 1, th.DurationNormal, animation.EaseOut),
	}
}

// Checked returns whether the checkbox is checked
func (cb *Checkbox) Checked() bool {
	return cb.checked
}

// SetChecked sets the checked state
func (cb *Checkbox) SetChecked(checked bool) {
	cb.checked = checked
	if checked {
		cb.checkAnim.PlayForward()
	} else {
		cb.checkAnim.PlayReverse()
	}
}

// SetOnClick sets the click handler
func (cb *Checkbox) SetOnClick(handler func()) {
	cb.onClick = handler
}

// Focusable returns true
func (cb *Checkbox) Focusable() bool {
	return true
}

// Render renders the checkbox
func (cb *Checkbox) Render(renderer *pipeline.Renderer) {
	if !cb.visible {
		return
	}

	th := theme.GetTheme()
	checkT := cb.checkAnim.Value()

	// Checkbox size
	boxSize := float32(16)
	spacing := float32(8)

	// Calculate positions
	boxRect := math.NewRect(cb.bounds.X, cb.bounds.Y+(cb.bounds.H-boxSize)/2, boxSize, boxSize)

	// Colors
	borderColor := color.BorderBase
	checkColor := color.Primary

	if cb.focused {
		borderColor = color.Primary
	}

	// Draw checkbox box
	renderer.StrokeRoundRect(boxRect, th.RadiusSM, 1.5, borderColor)

	// Draw check animation
	if checkT > 0 {
		// Fill background
		fillColor := checkColor.WithAlpha(checkT)
		renderer.FillRoundRect(boxRect, th.RadiusSM, fillColor)

		// Draw checkmark
		if checkT > 0.5 {
			checkAlpha := (checkT - 0.5) * 2 // 0.5-1 -> 0-1
			checkCol := math.NewColor(1, 1, 1, checkAlpha)
			renderer.DrawCheckmark(boxRect, boxSize*0.6, checkCol)
		}
	}

	// Draw text
	if cb.text != "" && cb.font != nil {
		textX := boxRect.X + boxSize + spacing
		textY := cb.bounds.Y + (cb.bounds.H-cb.font.LineHeight())/2
		renderer.DrawText(cb.text, textX, textY, cb.font, color.TextPrimary)
	}
}

// MouseDown handles mouse down
func (cb *Checkbox) MouseDown(x, y float32, button int) bool {
	if !cb.enabled || !cb.bounds.Contains(x, y) {
		return false
	}

	cb.checked = !cb.checked
	if cb.checked {
		cb.checkAnim.PlayForward()
	} else {
		cb.checkAnim.PlayReverse()
	}

	if cb.onClick != nil {
		cb.onClick()
	}

	return true
}

// Progress is a progress bar widget
type Progress struct {
	BaseWidget
	progress float32 // 0-1
	color    math.Color
	showText bool
	font     *font.Font
}

// NewProgress creates a new progress bar
func NewProgress(progress float32, f *font.Font) *Progress {
	return &Progress{
		BaseWidget: NewBaseWidget(),
		progress:   progress,
		color:      color.Primary,
		showText:   true,
		font:       f,
	}
}

// SetProgress sets the progress (0-1)
func (p *Progress) SetProgress(progress float32) {
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}
	p.progress = progress
}

// SetColor sets the progress color
func (p *Progress) SetColor(c math.Color) {
	p.color = c
}

// Render renders the progress bar
func (p *Progress) Render(renderer *pipeline.Renderer) {
	if !p.visible {
		return
	}

	th := theme.GetTheme()
	_ = th
	radius := p.bounds.H / 2

	// Background track
	renderer.FillRoundRect(p.bounds, radius, color.BgDark)

	// Progress fill
	fillWidth := p.bounds.W * p.progress
	if fillWidth > 0 {
		fillRect := math.NewRect(p.bounds.X, p.bounds.Y, fillWidth, p.bounds.H)
		renderer.FillRoundRect(fillRect, radius, p.color)
	}

	// Text
	if p.showText && p.font != nil {
		text := fmt.Sprintf("%.0f%%", p.progress*100)
		textW, textH := p.font.MeasureText(text)
		textX := p.bounds.X + (p.bounds.W-textW)/2
		textY := p.bounds.Y + (p.bounds.H-textH)/2

		// Choose text color based on progress
		textColor := color.TextPrimary
		if p.progress > 0.5 {
			textColor = color.TextWhite
		}

		renderer.DrawText(text, textX, textY, p.font, textColor)
	}
}

// Switch is a toggle switch widget
type Switch struct {
	BaseWidget
	checked   bool
	trackAnim *animation.Animation
	onClick   func()
}

// NewSwitch creates a new switch
func NewSwitch() *Switch {
	th := theme.GetTheme()

	return &Switch{
		BaseWidget: NewBaseWidget(),
		trackAnim:  animation.NewAnimation(0, 1, th.DurationNormal, animation.EaseOut),
	}
}

// Checked returns whether the switch is on
func (s *Switch) Checked() bool {
	return s.checked
}

// SetChecked sets the checked state
func (s *Switch) SetChecked(checked bool) {
	s.checked = checked
	if checked {
		s.trackAnim.PlayForward()
	} else {
		s.trackAnim.PlayReverse()
	}
}

// SetOnClick sets the click handler
func (s *Switch) SetOnClick(handler func()) {
	s.onClick = handler
}

// Focusable returns true
func (s *Switch) Focusable() bool {
	return true
}

// Render renders the switch
func (s *Switch) Render(renderer *pipeline.Renderer) {
	if !s.visible {
		return
	}

	th := theme.GetTheme()
	_ = th
	trackT := s.trackAnim.Value()

	// Track dimensions
	trackHeight := float32(22)
	trackWidth := float32(44)
	trackRect := math.NewRect(
		s.bounds.X,
		s.bounds.Y+(s.bounds.H-trackHeight)/2,
		trackWidth,
		trackHeight,
	)
	trackRadius := trackHeight / 2

	// Slider dimensions
	sliderRadius := float32(8)
	sliderMargin := float32(3)

	// Calculate slider position
	sliderX := trackRect.X + sliderMargin + sliderRadius + (trackWidth-2*sliderMargin-2*sliderRadius)*trackT
	sliderY := trackRect.Y + trackHeight/2

	// Track color (interpolate between off and on)
	offColor := color.BorderBase
	onColor := color.Primary
	trackColor := offColor.Lerp(onColor, trackT)

	// Draw track
	renderer.FillRoundRect(trackRect, trackRadius, trackColor)

	// Draw slider
	sliderCenter := math.NewVec2(sliderX, sliderY)
	renderer.FillCircle(sliderCenter, sliderRadius, math.NewColor(1, 1, 1, 1))

	// Slider shadow
	if s.focused {
		shadowColor := color.Primary.WithAlpha(0.3)
		renderer.DrawShadow(
			math.NewRect(sliderX-sliderRadius, sliderY-sliderRadius, sliderRadius*2, sliderRadius*2),
			math.NewVec2(0, 1),
			4,
			shadowColor,
		)
	}
}

// MouseDown handles mouse down
func (s *Switch) MouseDown(x, y float32, button int) bool {
	if !s.enabled || !s.bounds.Contains(x, y) {
		return false
	}

	s.checked = !s.checked
	if s.checked {
		s.trackAnim.PlayForward()
	} else {
		s.trackAnim.PlayReverse()
	}

	if s.onClick != nil {
		s.onClick()
	}

	return true
}
