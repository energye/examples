// Package widget provides popup and overlay support
package widget

import (
	"github.com/energye/examples/lcl/gpui/core/math"
	"github.com/energye/examples/lcl/gpui/render/font"
	"github.com/energye/examples/lcl/gpui/render/pipeline"
	"github.com/energye/examples/lcl/gpui/style/color"
	"github.com/energye/examples/lcl/gpui/style/theme"
)

// Popup represents a popup widget
type Popup struct {
	BaseWidget
	content     Widget
	anchor      Widget
	offset      math.Vec2
	bgColor     math.Color
	borderColor math.Color
	shadow      bool
	onClose     func()
}

// NewPopup creates a new popup
func NewPopup(content Widget) *Popup {
	return &Popup{
		BaseWidget:  NewBaseWidget(),
		content:     content,
		bgColor:     color.BgBase,
		borderColor: color.BorderBase,
		shadow:      true,
	}
}

// SetAnchor sets the anchor widget
func (p *Popup) SetAnchor(anchor Widget, offset math.Vec2) {
	p.anchor = anchor
	p.offset = offset
}

// ShowAt shows the popup at a specific position
func (p *Popup) ShowAt(x, y float32) {
	p.SetPos(x, y)
	p.visible = true

	if p.content != nil {
		p.content.SetPos(0, 0)
		p.SetSize(p.content.Width(), p.content.Height())
	}
}

// ShowAtWidget shows the popup anchored to a widget
func (p *Popup) ShowAtWidget(anchor Widget) {
	p.anchor = anchor
	p.updatePosition()
	p.visible = true
}

// Hide hides the popup
func (p *Popup) Hide() {
	p.visible = false
	if p.onClose != nil {
		p.onClose()
	}
}

// SetOnClose sets the close handler
func (p *Popup) SetOnClose(handler func()) {
	p.onClose = handler
}

func (p *Popup) updatePosition() {
	if p.anchor == nil {
		return
	}
	anchorBounds := p.anchor.Bounds()
	x := anchorBounds.X + p.offset.X
	y := anchorBounds.Y + anchorBounds.H + p.offset.Y
	p.SetPos(x, y)
}

// Render renders the popup
func (p *Popup) Render(renderer *pipeline.Renderer) {
	if !p.visible {
		return
	}

	th := theme.GetTheme()

	if p.shadow {
		shadowOffset := math.NewVec2(0, 2)
		renderer.DrawShadow(p.bounds, shadowOffset, 8, math.NewColor(0, 0, 0, 0.15))
	}

	renderer.FillRoundRect(p.bounds, th.RadiusMD, p.bgColor)
	renderer.StrokeRoundRect(p.bounds, th.RadiusMD, 1, p.borderColor)

	if p.content != nil {
		p.content.Render(renderer)
	}
}

// Tooltip is a simple tooltip popup
type Tooltip struct {
	BaseWidget
	text      string
	font      *font.Font
	bgColor   math.Color
	textColor math.Color
}

func NewTooltip(text string, f *font.Font) *Tooltip {
	return &Tooltip{
		BaseWidget: NewBaseWidget(),
		text:       text,
		font:       f,
		bgColor:    math.NewColor(0, 0, 0, 0.8),
		textColor:  color.TextWhite,
	}
}

func (t *Tooltip) ShowAt(x, y float32) {
	if t.font == nil || t.text == "" {
		return
	}

	th := theme.GetTheme()
	padding := th.SpaceSM
	textW, textH := t.font.MeasureText(t.text)

	t.SetPos(x, y-textH-padding*2)
	t.SetSize(textW+padding*2, textH+padding*2)
	t.visible = true
}

func (t *Tooltip) Render(renderer *pipeline.Renderer) {
	if !t.visible || t.font == nil || t.text == "" {
		return
	}

	th := theme.GetTheme()
	renderer.FillRoundRect(t.bounds, th.RadiusSM, t.bgColor)

	padding := th.SpaceSM
	textX := t.bounds.X + padding
	textY := t.bounds.Y + padding
	renderer.DrawText(t.text, textX, textY, t.font, t.textColor)
}

// DropdownItem represents a dropdown item
type DropdownItem struct {
	Text    string
	Value   interface{}
	Enabled bool
}

// Dropdown is a dropdown popup
type Dropdown struct {
	Popup
	items       []DropdownItem
	selectedIdx int
	onSelect    func(idx int)
	itemHeight  float32
	maxVisible  int
	scrollY     float32
	hoveredIdx  int
	font        *font.Font
}

func NewDropdown(f *font.Font) *Dropdown {
	return &Dropdown{
		Popup: Popup{
			BaseWidget:  NewBaseWidget(),
			bgColor:     color.BgBase,
			borderColor: color.BorderBase,
			shadow:      true,
		},
		items:      make([]DropdownItem, 0),
		itemHeight: 32,
		maxVisible: 8,
		hoveredIdx: -1,
		font:       f,
	}
}

func (d *Dropdown) SetItems(items []DropdownItem) {
	d.items = items
	d.updateSize()
}

func (d *Dropdown) SetOnSelect(handler func(idx int)) {
	d.onSelect = handler
}

func (d *Dropdown) SelectedIndex() int {
	return d.selectedIdx
}

func (d *Dropdown) SelectedItem() *DropdownItem {
	if d.selectedIdx >= 0 && d.selectedIdx < len(d.items) {
		return &d.items[d.selectedIdx]
	}
	return nil
}

func (d *Dropdown) updateSize() {
	th := theme.GetTheme()
	padding := th.SpaceSM

	maxWidth := float32(100)
	for _, item := range d.items {
		itemWidth := float32(len(item.Text))*8 + padding*2
		if itemWidth > maxWidth {
			maxWidth = itemWidth
		}
	}

	visibleItems := len(d.items)
	if visibleItems > d.maxVisible {
		visibleItems = d.maxVisible
	}
	height := float32(visibleItems) * d.itemHeight
	d.SetSize(maxWidth, height)
}

func (d *Dropdown) Render(renderer *pipeline.Renderer) {
	if !d.visible || len(d.items) == 0 {
		return
	}

	th := theme.GetTheme()

	if d.shadow {
		shadowOffset := math.NewVec2(0, 2)
		renderer.DrawShadow(d.bounds, shadowOffset, 8, math.NewColor(0, 0, 0, 0.15))
	}

	renderer.FillRoundRect(d.bounds, th.RadiusMD, d.bgColor)
	renderer.StrokeRoundRect(d.bounds, th.RadiusMD, 1, d.borderColor)

	for i, item := range d.items {
		itemY := d.bounds.Y + float32(i)*d.itemHeight - d.scrollY
		itemRect := math.NewRect(d.bounds.X, itemY, d.bounds.W, d.itemHeight)

		if itemY+d.itemHeight < d.bounds.Y || itemY > d.bounds.Y+d.bounds.H {
			continue
		}

		itemBg := d.bgColor
		if i == d.hoveredIdx {
			itemBg = color.PrimaryBg
		} else if i == d.selectedIdx {
			itemBg = color.BgDark
		}
		renderer.FillRect(itemRect, itemBg)

		if d.font != nil {
			textColor := color.TextPrimary
			if !item.Enabled {
				textColor = color.TextDisabled
			}
			textX := itemRect.X + th.SpaceSM
			textY := itemRect.Y + (d.itemHeight-d.font.LineHeight())/2
			renderer.DrawText(item.Text, textX, textY, d.font, textColor)
		}
	}
}

func (d *Dropdown) MouseDown(x, y float32, button int) bool {
	if !d.enabled || !d.bounds.Contains(x, y) {
		d.Hide()
		return false
	}

	idx := int((y - d.bounds.Y + d.scrollY) / d.itemHeight)
	if idx >= 0 && idx < len(d.items) && d.items[idx].Enabled {
		d.selectedIdx = idx
		d.Hide()
		if d.onSelect != nil {
			d.onSelect(idx)
		}
		return true
	}

	return true
}

func (d *Dropdown) MouseMove(x, y float32) bool {
	if !d.bounds.Contains(x, y) {
		d.hoveredIdx = -1
		return false
	}

	idx := int((y - d.bounds.Y + d.scrollY) / d.itemHeight)
	if idx >= 0 && idx < len(d.items) {
		d.hoveredIdx = idx
	} else {
		d.hoveredIdx = -1
	}

	return true
}

// Modal is a modal dialog overlay
type Modal struct {
	BaseWidget
	content   Widget
	title     string
	titleFont *font.Font
	bgColor   math.Color
	maskColor math.Color
	onClose   func()
	closable  bool
}

func NewModal(title string, content Widget, titleFont *font.Font) *Modal {
	return &Modal{
		BaseWidget: NewBaseWidget(),
		content:    content,
		title:      title,
		titleFont:  titleFont,
		bgColor:    color.BgBase,
		maskColor:  math.NewColor(0, 0, 0, 0.45),
		closable:   true,
	}
}

func (m *Modal) Show(parentBounds math.Rect) {
	modalW := float32(400)
	modalH := float32(300)
	x := parentBounds.X + (parentBounds.W-modalW)/2
	y := parentBounds.Y + (parentBounds.H-modalH)/2

	m.SetPos(x, y)
	m.SetSize(modalW, modalH)
	m.visible = true
}

func (m *Modal) Hide() {
	m.visible = false
	if m.onClose != nil {
		m.onClose()
	}
}

func (m *Modal) SetOnClose(handler func()) {
	m.onClose = handler
}

func (m *Modal) Render(renderer *pipeline.Renderer, parentBounds math.Rect) {
	if !m.visible {
		return
	}

	th := theme.GetTheme()

	renderer.FillRect(parentBounds, m.maskColor)

	shadowOffset := math.NewVec2(0, 4)
	renderer.DrawShadow(m.bounds, shadowOffset, 16, math.NewColor(0, 0, 0, 0.2))

	renderer.FillRoundRect(m.bounds, th.RadiusLG, m.bgColor)

	titleHeight := float32(48)
	titleRect := math.NewRect(m.bounds.X, m.bounds.Y, m.bounds.W, titleHeight)
	renderer.FillRoundRect(titleRect, th.RadiusLG, color.Primary)

	if m.titleFont != nil && m.title != "" {
		textW, textH := m.titleFont.MeasureText(m.title)
		textX := m.bounds.X + (m.bounds.W-textW)/2
		textY := m.bounds.Y + (titleHeight-textH)/2
		renderer.DrawText(m.title, textX, textY, m.titleFont, color.TextWhite)
	}

	if m.closable {
		closeSize := float32(16)
		closeX := m.bounds.X + m.bounds.W - closeSize - th.SpaceSM
		closeY := m.bounds.Y + (titleHeight-closeSize)/2
		closeRect := math.NewRect(closeX, closeY, closeSize, closeSize)

		renderer.DrawLine(closeRect.X, closeRect.Y, closeRect.X+closeRect.W, closeRect.Y+closeRect.H, 2, color.TextWhite)
		renderer.DrawLine(closeRect.X+closeRect.W, closeRect.Y, closeRect.X, closeRect.Y+closeRect.H, 2, color.TextWhite)
	}

	if m.content != nil {
		contentRect := math.NewRect(
			m.bounds.X+th.SpaceMD,
			m.bounds.Y+titleHeight+th.SpaceMD,
			m.bounds.W-th.SpaceMD*2,
			m.bounds.H-titleHeight-th.SpaceMD*2,
		)
		m.content.SetBounds(contentRect)
		m.content.Render(renderer)
	}
}

func (m *Modal) MouseDown(x, y float32, button int) bool {
	if !m.visible {
		return false
	}

	if !m.bounds.Contains(x, y) {
		if m.closable {
			m.Hide()
		}
		return true
	}

	th := theme.GetTheme()
	titleHeight := float32(48)
	closeSize := float32(16)
	closeX := m.bounds.X + m.bounds.W - closeSize - th.SpaceSM
	closeY := m.bounds.Y + (titleHeight-closeSize)/2
	closeRect := math.NewRect(closeX, closeY, closeSize, closeSize)

	if closeRect.Contains(x, y) && m.closable {
		m.Hide()
		return true
	}

	if m.content != nil {
		return m.content.MouseDown(x, y, button)
	}

	return true
}
