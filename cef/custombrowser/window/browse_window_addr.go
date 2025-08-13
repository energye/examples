package window

import (
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
	"github.com/energye/lcl/types/colors"
	"net/url"
	"strings"
)

func (m *BrowserWindow) createAddrBar() {
	// 地址栏 + 自绘 panel 主要重写形状和背景
	m.addr = lcl.NewMemo(m)
	m.addr.SetParent(m.box)
	m.addr.SetLeft(160)
	m.addr.SetTop(52)
	m.addr.SetHeight(33)
	m.addr.SetWidth(m.Width() - (m.addr.Left() + 80))
	m.addr.SetBorderStyle(types.BsNone)
	m.addr.SetAnchors(types.NewSet(types.AkLeft, types.AkTop, types.AkRight))
	m.addr.Font().SetSize(16)
	m.addr.Font().SetColor(colors.ClWhite)
	m.addr.SetColor(colors.RGBToColor(56, 57, 60))
	// 阻止 memo 换行
	m.addr.SetOnKeyDown(func(sender lcl.IObject, key *uint16, shift types.TShiftState) {
		k := *key
		if k == 13 || k == 10 {
			*key = 0
			tempUrl := strings.TrimSpace(m.addr.Text())
			if _, err := url.Parse(tempUrl); err != nil || tempUrl == "" {
				tempUrl = "https://energye.github.io/"
			}
			for _, chrom := range m.chroms {
				if chrom.isActive {
					chrom.chromium.LoadURLWithStringFrame(tempUrl, chrom.chromium.Browser().GetMainFrame())
				}
			}
		}
	})
	// 阻止 memo 换行
	m.addr.SetOnChange(func(sender lcl.IObject) {
		text := m.addr.Text()
		text = strings.ReplaceAll(text, "\r", "")
		text = strings.ReplaceAll(text, "\n", "")
		m.addr.SetText(text)
	})

}
