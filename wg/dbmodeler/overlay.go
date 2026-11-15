package main

// 浮层：右键菜单 / 下拉选项 / 模态弹窗（DDL 预览、帮助、校验报告）。
// 同样是纯自绘对象，没有原生菜单、没有原生对话框。

import (
	"strings"

	"github.com/energye/lcl/types"
	"github.com/energye/lcl/types/colors"
)

type MenuItem struct {
	Text     string
	Hint     string
	Shortcut string
	Enabled  bool
	Checked  bool
	Danger   bool
	Sep      bool
	OnClick  func()
}

func mi(text string, onClick func()) MenuItem {
	return MenuItem{Text: text, Enabled: true, OnClick: onClick}
}
func miSep() MenuItem { return MenuItem{Sep: true} }

type PopupMenu struct {
	baseWidget
	Title   string
	Items   []MenuItem
	hover   int
	OnClose func()
	OnHover func(string)
}

func NewPopupMenu(title string, items []MenuItem) *PopupMenu {
	return &PopupMenu{
		baseWidget: newBase("popup", r4(0, 0, 224, 100), false),
		Title:      title, Items: items, hover: -1,
	}
}

func (m *PopupMenu) itemHeight(it MenuItem) int32 {
	if it.Sep {
		return 9
	}
	return 27
}

func (m *PopupMenu) PlaceAt(x, y, clientW, clientH int32) {
	h := int32(8)
	if m.Title != "" {
		h += 24
	}
	for _, it := range m.Items {
		h += m.itemHeight(it)
	}
	h += 8
	w := int32(226)
	if x+w > clientW-6 {
		x = clientW - w - 6
	}
	if y+h > clientH-6 {
		y = clientH - h - 6
	}
	if x < 6 {
		x = 6
	}
	if y < 6 {
		y = 6
	}
	m.rect = r4(x, y, x+w, y+h)
}

func (m *PopupMenu) itemTop(idx int) int32 {
	y := m.rect.Top + 8
	if m.Title != "" {
		y += 24
	}
	for i := 0; i < idx; i++ {
		y += m.itemHeight(m.Items[i])
	}
	return y
}

func (m *PopupMenu) itemAt(y int32) int {
	for i, it := range m.Items {
		top := m.itemTop(i)
		if y >= top && y < top+m.itemHeight(it) && !it.Sep {
			return i
		}
	}
	return -1
}

func (m *PopupMenu) Paint(p *Painter) {
	if !m.visible {
		return
	}
	r := m.rect
	// 伪阴影：先画一个偏移的深色底，再画面板。
	shadow := r4(r.Left+3, r.Top+4, r.Right+3, r.Bottom+4)
	if p.Theme.Dark {
		p.FillRect(shadow, rgb(8, 10, 14))
	} else {
		p.FillRect(shadow, rgb(190, 199, 214))
	}
	p.RoundRect(r, 7, p.Theme.Surface, p.Theme.Border)
	if m.Title != "" {
		p.Text(r4(r.Left+12, r.Top+6, r.Right-12, r.Top+28),
			m.Title, p.Theme.Muted, 10, AlignLeft, true)
		p.FillRect(r4(r.Left+8, r.Top+29, r.Right-8, r.Top+30), p.Theme.Border)
	}
	for i, it := range m.Items {
		top := m.itemTop(i)
		h := m.itemHeight(it)
		if it.Sep {
			p.FillRect(r4(r.Left+10, top+4, r.Right-10, top+5), p.Theme.Border)
			continue
		}
		if i == m.hover && it.Enabled {
			p.RoundRect(r4(r.Left+4, top, r.Right-4, top+h), 4,
				p.Theme.AccentSoft, p.Theme.AccentSoft)
		}
		c := p.Theme.Text
		if !it.Enabled {
			c = p.Theme.Muted
		} else if it.Danger {
			c = p.Theme.Danger
		}
		if it.Checked {
			p.Text(r4(r.Left+8, top, r.Left+24, top+h), "✓", p.Theme.Accent, 12, AlignLeft, true)
		}
		p.Text(r4(r.Left+27, top, r.Right-64, top+h), it.Text, c, 12, AlignLeft, false)
		if it.Shortcut != "" {
			p.Text(r4(r.Right-78, top, r.Right-10, top+h),
				it.Shortcut, p.Theme.Muted, 10, AlignRight, false)
		}
	}
}

func (m *PopupMenu) Handle(e *Event) bool {
	switch e.Kind {
	case EvMove:
		m.hover = m.itemAt(e.Y)
		if m.hover >= 0 && m.OnHover != nil {
			m.OnHover(m.Items[m.hover].Hint)
		}
	case EvLeave:
		m.hover = -1
	case EvDown:
		if e.Button != types.MbLeft {
			return true
		}
		idx := m.itemAt(e.Y)
		if idx >= 0 && m.Items[idx].Enabled && m.Items[idx].OnClick != nil {
			m.Items[idx].OnClick()
		}
		if m.OnClose != nil {
			m.OnClose()
		}
	}
	return true
}

// ---- Modal ----

type ModalButton struct {
	Text    string
	Primary bool
	OnClick func()
}

type Modal struct {
	baseWidget
	Title    string
	Lines    []string
	Subtitle string
	Buttons  []ModalButton
	OnClose  func()
	Scroll   int32
	lineH    int32
	panel    types.TRect
	hoverBtn int
	dragBar  bool
	dragDY   int32
}

func NewModal(title string, lines []string) *Modal {
	return &Modal{
		baseWidget: newBase("modal", r4(0, 0, 100, 100), false),
		Title:      title, Lines: lines, lineH: 18, hoverBtn: -1,
	}
}

func (d *Modal) SetText(s string) {
	d.Lines = strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), "\n")
	d.Scroll = 0
}

func (d *Modal) Layout(client types.TRect) {
	d.rect = client
	w := int32(760)
	if client.Width()-80 < w {
		w = client.Width() - 80
	}
	if w < 320 {
		w = 320
	}
	h := int32(560)
	if client.Height()-80 < h {
		h = client.Height() - 80
	}
	if h < 240 {
		h = 240
	}
	l := client.Left + (client.Width()-w)/2
	t := client.Top + (client.Height()-h)/2
	d.panel = r4(l, t, l+w, t+h)
}

func (d *Modal) contentRect() types.TRect {
	p := d.panel
	return r4(p.Left+18, p.Top+58, p.Right-18, p.Bottom-56)
}

func (d *Modal) maxScroll() int32 {
	need := int32(len(d.Lines)) * d.lineH
	avail := d.contentRect().Height()
	if need <= avail {
		return 0
	}
	return need - avail
}

func (d *Modal) clampScroll() {
	if d.Scroll < 0 {
		d.Scroll = 0
	}
	if m := d.maxScroll(); d.Scroll > m {
		d.Scroll = m
	}
}

func (d *Modal) barTrack() types.TRect {
	cr := d.contentRect()
	return r4(cr.Right-6, cr.Top, cr.Right-1, cr.Bottom)
}

func (d *Modal) buttonRect(i int) types.TRect {
	p := d.panel
	w := int32(104)
	x := p.Right - 18 - w - int32(i)*(w+10)
	return r4(x, p.Bottom-44, x+w, p.Bottom-16)
}

func (d *Modal) Paint(p *Painter) {
	if !d.visible {
		return
	}
	mask := colors.RGBToColor(184, 194, 210)
	if p.Theme.Dark {
		mask = colors.RGBToColor(10, 12, 17)
	}
	p.FillRect(d.rect, mask)
	panel := d.panel
	p.RoundRect(panel, 10, p.Theme.Surface, p.Theme.Border)
	p.Gradient(r4(panel.Left+1, panel.Top+1, panel.Right-1, panel.Top+46),
		groupColor(*p.Theme, "core"), p.Theme.AccentDown)
	p.Text(r4(panel.Left+18, panel.Top+4, panel.Right-50, panel.Top+42),
		d.Title, colors.RGBToColor(255, 255, 255), 15, AlignLeft, true)
	if d.Subtitle != "" {
		p.Text(r4(panel.Left+18, panel.Top+44, panel.Right-18, panel.Top+58),
			d.Subtitle, p.Theme.Muted, 10, AlignLeft, false)
	}
	// 关闭按钮
	closeR := r4(panel.Right-38, panel.Top+8, panel.Right-12, panel.Top+36)
	p.Text(closeR, "×", colors.RGBToColor(255, 255, 255), 17, AlignCenter, false)

	cr := d.contentRect()
	p.RoundRect(r4(cr.Left-6, cr.Top-4, cr.Right+6, cr.Bottom+4), 5,
		p.Theme.Bg, p.Theme.Border)
	p.SetClip(cr)
	d.clampScroll()
	first := int(d.Scroll / d.lineH)
	for idx := first; idx < len(d.Lines); idx++ {
		y := cr.Top + int32(idx)*d.lineH - d.Scroll
		if y+d.lineH <= cr.Top || y >= cr.Bottom {
			continue
		}
		line := d.Lines[idx]
		c := p.Theme.Text
		if strings.HasPrefix(strings.TrimSpace(line), "--") {
			c = p.Theme.Muted
		} else if strings.HasPrefix(strings.ToUpper(strings.TrimSpace(line)), "CREATE ") {
			c = p.Theme.Accent
		} else if strings.HasPrefix(strings.ToUpper(strings.TrimSpace(line)), "ALTER ") {
			c = p.Theme.FK
		}
		p.Text(r4(cr.Left+30, y, cr.Right-8, y+d.lineH), line, c, 11, AlignLeft, false)
		p.Text(r4(cr.Left, y, cr.Left+26, y+d.lineH),
			fmtInt(idx+1), p.Theme.Muted, 10, AlignRight, false)
	}
	p.ClearClip()

	// 内容滚动条（可拖动）
	if d.maxScroll() > 0 {
		track := d.barTrack()
		p.FillRect(track, p.Theme.SurfaceAlt)
		thumb := scrollThumb(track, d.Scroll, d.maxScroll())
		p.RoundRect(thumb, 3, p.Theme.Accent, p.Theme.Accent)
	}
	for i, b := range d.Buttons {
		br := d.buttonRect(i)
		bg, fg := p.Theme.SurfaceAlt, p.Theme.Text
		if b.Primary {
			bg, fg = p.Theme.Accent, p.Theme.Surface
		}
		if i == d.hoverBtn {
			bg = p.Theme.AccentDown
			fg = p.Theme.Surface
		}
		p.RoundRect(br, 5, bg, p.Theme.Border)
		p.Text(br, b.Text, fg, 12, AlignCenter, b.Primary)
	}
}

func fmtInt(n int) string {
	if n < 10 {
		return " " + string(rune('0'+n))
	}
	if n < 100 {
		return string(rune('0'+n/10)) + string(rune('0'+n%10))
	}
	// 行号超过 99 时用简单十进制，避免引入额外依赖
	return itoa(n)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}

func (d *Modal) hitButton(x, y int32) int {
	for i := range d.Buttons {
		r := d.buttonRect(i)
		if x >= r.Left && x < r.Right && y >= r.Top && y < r.Bottom {
			return i
		}
	}
	return -1
}

func (d *Modal) Handle(e *Event) bool {
	switch e.Kind {
	case EvMove:
		if d.dragBar {
			track := d.barTrack()
			thumb := scrollThumb(track, d.Scroll, d.maxScroll())
			d.Scroll = scrollFromThumb(track, thumb.Height(), e.Y, d.dragDY, d.maxScroll())
			d.clampScroll()
			return true
		}
		d.hoverBtn = d.hitButton(e.X, e.Y)
	case EvLeave:
		d.hoverBtn = -1
		d.dragBar = false
	case EvUp:
		d.dragBar = false
	case EvWheel:
		d.Scroll -= e.Wheel / 3
		d.clampScroll()
	case EvKey:
		switch e.Key {
		case vkEscape:
			if d.OnClose != nil {
				d.OnClose()
			}
		case vkUp:
			d.Scroll -= d.lineH
			d.clampScroll()
		case vkDown:
			d.Scroll += d.lineH
			d.clampScroll()
		case vkHome:
			d.Scroll = 0
		case vkEnd:
			d.Scroll = d.maxScroll()
		case 33: // PageUp
			d.Scroll -= d.lineH * 12
			d.clampScroll()
		case 34: // PageDown
			d.Scroll += d.lineH * 12
			d.clampScroll()
		}
	case EvDown:
		if e.Button != types.MbLeft {
			return true
		}
		if max := d.maxScroll(); max > 0 && e.X >= d.barTrack().Left && e.X <= d.barTrack().Right {
			track := d.barTrack()
			thumb := scrollThumb(track, d.Scroll, max)
			if e.Y >= thumb.Top && e.Y < thumb.Bottom {
				d.dragBar = true
				d.dragDY = e.Y - thumb.Top
			} else {
				if e.Y < thumb.Top {
					d.Scroll -= d.contentRect().Height()
				} else {
					d.Scroll += d.contentRect().Height()
				}
				d.clampScroll()
				d.dragBar = true
				d.dragDY = thumb.Height() / 2
			}
			return true
		}
		if idx := d.hitButton(e.X, e.Y); idx >= 0 && d.Buttons[idx].OnClick != nil {
			d.Buttons[idx].OnClick()
			return true
		}
		closeR := r4(d.panel.Right-38, d.panel.Top+8, d.panel.Right-12, d.panel.Top+36)
		if e.X >= closeR.Left && e.X < closeR.Right && e.Y >= closeR.Top && e.Y < closeR.Bottom {
			if d.OnClose != nil {
				d.OnClose()
			}
			return true
		}
		// 点击遮罩（面板外）也关闭
		if e.X < d.panel.Left || e.X > d.panel.Right || e.Y < d.panel.Top || e.Y > d.panel.Bottom {
			if d.OnClose != nil {
				d.OnClose()
			}
		}
	}
	return true
}
