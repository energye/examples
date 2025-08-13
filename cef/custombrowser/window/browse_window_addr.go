package window

import (
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
	"github.com/energye/lcl/types/colors"
	"net/url"
	"strings"
	"widget/wg"
)

func (m *BrowserWindow) createAddrBar() {
	color := colors.RGBToColor(86, 88, 93)
	top := int32(50)
	// 地址栏 + 自绘 panel 主要重写形状和背景
	var (
		addrLeft  *wg.TButton
		addrRight *wg.TButton
	)
	m.addr = lcl.NewMemo(m)
	m.addr.SetParent(m.box)
	m.addr.SetLeft(160)
	m.addr.SetTop(top)
	m.addr.SetHeight(30)
	m.addr.SetWidth(m.Width() - (m.addr.Left() + 80))
	m.addr.SetBorderStyle(types.BsNone)
	m.addr.SetColor(color)
	m.addr.SetAnchors(types.NewSet(types.AkLeft, types.AkTop, types.AkRight))
	m.addr.Font().SetSize(17)
	m.addr.Font().SetHeight(-22)
	m.addr.Font().SetColor(colors.ClWhite)
	m.addr.SetWordWrap(false)
	m.addr.SetWantReturns(false)
	m.addr.SetWantTabs(false)
	// 阻止 memo 换行
	m.addr.SetOnKeyDown(func(sender lcl.IObject, key *uint16, shift types.TShiftState) {
		k := *key
		if k == 13 || k == 10 {
			//*key = 0
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

	addrFocus := func(sender lcl.IObject) {
		m.addr.SetSelStart(int32(len(m.addr.Text())))
		m.addr.SetFocus()
	}

	addrEnter := func(sender lcl.IObject) {
		dkColor := wg.DarkenColor(color, -0.2)
		m.addr.SetColor(dkColor)
		addrLeft.SetStartColor(dkColor)
		addrLeft.SetEndColor(dkColor)
		addrRight.SetStartColor(dkColor)
		addrRight.SetEndColor(dkColor)
		addrLeft.Invalidate()
		addrRight.Invalidate()
	}
	addrLeave := func(sender lcl.IObject) {
		m.addr.SetColor(color)
		addrLeft.SetStartColor(color)
		addrLeft.SetEndColor(color)
		addrRight.SetStartColor(color)
		addrRight.SetEndColor(color)
		addrLeft.Invalidate()
		addrRight.Invalidate()
	}

	addrLeft = wg.NewButton(m)
	addrLeft.SetParent(m.box)
	addrLeftRect := types.TRect{Left: 140, Top: top}
	addrLeftRect.SetSize(30, 30)
	addrLeft.SetBoundsRect(addrLeftRect)
	addrLeft.SetStartColor(color)
	addrLeft.SetEndColor(color)
	addrLeft.SetRadius(15)
	addrLeft.SetAlpha(255)
	addrLeft.IsDisable = true
	addrLeft.RoundedCorner = addrLeft.RoundedCorner.Exclude(wg.RcRightBottom).Exclude(wg.RcRightTop)
	addrLeft.SetOnClick(addrFocus)
	addrLeft.SetOnMouseEnter(addrEnter)
	addrLeft.SetOnMouseLeave(addrLeave)

	addrRight = wg.NewButton(m)
	addrRight.SetParent(m.box)
	addrRightRect := types.TRect{Left: m.addr.Left() + m.addr.Width(), Top: top}
	addrRightRect.SetSize(30, 30)
	addrRight.SetBoundsRect(addrRightRect)
	addrRight.SetStartColor(color)
	addrRight.SetEndColor(color)
	addrRight.SetRadius(15)
	addrRight.SetAlpha(255)
	addrRight.IsDisable = true
	addrRight.RoundedCorner = addrRight.RoundedCorner.Exclude(wg.RcLeftBottom).Exclude(wg.RcLeftTop)
	addrRight.SetOnClick(addrFocus)
	addrRight.SetOnMouseEnter(addrEnter)
	addrRight.SetOnMouseLeave(addrLeave)

	// 地址栏右边的 logo 按钮
	m.addrRightBtn = wg.NewButton(m)
	m.addrRightBtn.SetParent(m.box)
	m.addrRightBtn.SetShowHint(true)
	m.addrRightBtn.SetHint("   GO  \nENERGY")
	addrRightBtnRect := types.TRect{Left: m.box.Width() - (40 + 5), Top: 47}
	addrRightBtnRect.SetSize(40, 40)
	m.addrRightBtn.SetBoundsRect(addrRightBtnRect)
	m.addrRightBtn.SetStartColor(colors.RGBToColor(56, 57, 60))
	m.addrRightBtn.SetEndColor(colors.RGBToColor(50, 60, 70))
	m.addrRightBtn.SetRadius(35)
	m.addrRightBtn.SetAlpha(255)
	m.addrRightBtn.SetIcon(getResourcePath("addr-right-btn.png"))
	m.addrRightBtn.SetOnClick(func(sender lcl.IObject) {
		if chrom := m.getActiveChrom(); chrom != nil {
			chrom.chromium.LoadURLWithStringFrame("https://energye.github.io", chrom.chromium.Browser().GetMainFrame())
		}
	})

	m.addr.SetOnResize(func(sender lcl.IObject) {
		addrRight.SetLeft(m.addr.Left() + m.addr.Width())
		// 地址栏右侧按钮
		m.addrRightBtn.SetLeft(m.box.Width() - (m.addrRightBtn.Width() + 5))
	})

	m.addr.SetOnMouseEnter(addrEnter)
	m.addr.SetOnMouseLeave(addrLeave)
}
