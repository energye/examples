//go:build !darwin
// +build !darwin

package window

import (
	"github.com/energye/lcl/api"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
	"github.com/energye/lcl/types/colors"
	"github.com/energye/widget/wg"
	"net/url"
	"strings"
)

const isDarwin = false

func (m *BrowserWindow) macOSToolbar() {

}

func (m *BrowserWindow) SetAddrText(val string) {
	lcl.RunOnMainThreadAsync(func(id uint32) {
		m.addr.SetText(val)
		m.addr.SetSelStart(int32(len(val)))
		m.addr.SetFocus()
	})
}

func (m *BrowserWindow) createTitleWidgetControl() {
	// 添加 chromium 按钮
	m.addChromBtn = wg.NewButton(m)
	m.addChromBtn.SetParent(m.box)
	addBtnRect := types.TRect{Left: 5, Top: 5}
	addBtnRect.SetSize(40, 40)
	m.addChromBtn.SetBoundsRect(addBtnRect)
	m.addChromBtn.SetColor(bgColor)
	m.addChromBtn.SetRadius(5)
	m.addChromBtn.SetAlpha(255)
	m.addChromBtn.SetIcon(getResourcePath("add.png"))
	m.addChromBtn.SetOnClick(func(sender lcl.IObject) {
		println("add chromium isMainThread:", api.MainThreadId() == api.CurrentThreadId())
		m.addr.SetText("")
		newChromium := m.createChromium("")
		m.OnChromiumCreateTabSheet(newChromium)
		newChromium.createBrowser(nil)
	})
	// 窗口控制按钮 最小化，最大化，关闭
	m.minBtn = wg.NewButton(m)
	m.minBtn.SetParent(m.box)
	m.minBtn.SetShowHint(true)
	m.minBtn.SetHint("最小化")
	minBtnRect := types.TRect{Left: m.box.Width() - 45*3, Top: 5}
	minBtnRect.SetSize(40, 40)
	m.minBtn.SetBoundsRect(minBtnRect)
	m.minBtn.SetColor(bgColor)
	m.minBtn.SetRadius(5)
	m.minBtn.SetAlpha(255)
	m.minBtn.SetIcon(getResourcePath("btn-min.png"))
	m.minBtn.SetOnClick(func(sender lcl.IObject) {
		m.Minimize()
	})
	m.maxBtn = wg.NewButton(m)
	m.maxBtn.SetParent(m.box)
	m.maxBtn.SetShowHint(true)
	m.maxBtn.SetHint("最大化")
	maxBtnRect := types.TRect{Left: m.box.Width() - 45*2, Top: 5}
	maxBtnRect.SetSize(40, 40)
	m.maxBtn.SetBoundsRect(maxBtnRect)
	m.maxBtn.SetColor(bgColor)
	m.maxBtn.SetRadius(5)
	m.maxBtn.SetAlpha(255)
	m.maxBtn.SetIcon(getResourcePath("btn-max.png"))
	m.maxBtn.SetOnClick(func(sender lcl.IObject) {
		m.Maximize()
	})
	m.closeBtn = wg.NewButton(m)
	m.closeBtn.SetParent(m.box)
	m.closeBtn.SetShowHint(true)
	m.closeBtn.SetHint("关闭")
	closeBtnRect := types.TRect{Left: m.box.Width() - 45, Top: 5}
	closeBtnRect.SetSize(40, 40)
	m.closeBtn.SetBoundsRect(closeBtnRect)
	m.closeBtn.SetColor(bgColor)
	m.closeBtn.SetRadius(5)
	m.closeBtn.SetAlpha(255)
	m.closeBtn.SetIcon(getResourcePath("btn-close.png"))
	m.closeBtn.SetOnClick(func(sender lcl.IObject) {
		if len(m.chroms) == 0 {
			m.Close()
		} else {
			for _, chrom := range m.chroms {
				chrom.closeBrowser()
			}
		}
		m.isWindowButtonClose = true
	})
	// 浏览器控制按钮
	// 后退
	m.backBtn = wg.NewButton(m)
	m.backBtn.SetParent(m.box)
	m.backBtn.SetShowHint(true)
	m.backBtn.SetHint("单击返回")
	backBtnRect := types.TRect{Left: 5, Top: 47}
	backBtnRect.SetSize(40, 40)
	m.backBtn.SetBoundsRect(backBtnRect)
	m.backBtn.SetColor(bgColor)
	m.backBtn.SetRadius(5)
	m.backBtn.SetAlpha(255)
	m.backBtn.SetIcon(getResourcePath("back.png"))
	m.backBtn.SetOnClick(func(sender lcl.IObject) {
		chrom := m.getActiveChrom()
		if chrom != nil && chrom.chromium.CanGoBack() {
			chrom.chromium.GoBack()
		}
	})
	// 前进按钮
	m.forwardBtn = wg.NewButton(m)
	m.forwardBtn.SetParent(m.box)
	m.forwardBtn.SetShowHint(true)
	m.forwardBtn.SetHint("单击前进")
	forwardBtnRect := types.TRect{Left: 50, Top: 47}
	forwardBtnRect.SetSize(40, 40)
	m.forwardBtn.SetBoundsRect(forwardBtnRect)
	m.forwardBtn.SetColor(bgColor)
	m.forwardBtn.SetRadius(5)
	m.forwardBtn.SetAlpha(255)
	m.forwardBtn.SetIcon(getResourcePath("forward.png"))
	m.forwardBtn.SetOnClick(func(sender lcl.IObject) {
		chrom := m.getActiveChrom()
		if chrom != nil && chrom.chromium.CanGoForward() {
			chrom.chromium.GoForward()
		}
	})
	// 刷新按钮
	m.refreshBtn = wg.NewButton(m)
	m.refreshBtn.SetParent(m.box)
	m.refreshBtn.SetShowHint(true)
	m.refreshBtn.SetHint("单击刷新/停止")
	refreshBtnRect := types.TRect{Left: 95, Top: 47}
	refreshBtnRect.SetSize(40, 40)
	m.refreshBtn.SetBoundsRect(refreshBtnRect)
	m.refreshBtn.SetColor(bgColor)
	m.refreshBtn.SetRadius(5)
	m.refreshBtn.SetAlpha(255)
	m.refreshBtn.SetIcon(getResourcePath("refresh.png"))
	m.refreshBtn.SetOnClick(func(sender lcl.IObject) {
		chrom := m.getActiveChrom()
		if chrom != nil {
			if chrom.isLoading {
				chrom.chromium.StopLoad()
			} else {
				chrom.chromium.Reload()
			}
		}
	})
	// 地址栏
	m.createAddrBar()
}

func (m *BrowserWindow) createAddrBar() {
	color := colors.RGBToColor(86, 88, 93)
	top := int32(50)
	height := int32(33)
	// 地址栏 + 自绘 panel 主要重写形状和背景
	var (
		addrLeft  *wg.TButton
		addrRight *wg.TButton
	)
	m.addr = lcl.NewMemo(m)
	m.addr.SetParent(m.box)
	m.addr.SetLeft(160)
	m.addr.SetTop(top)
	m.addr.SetHeight(height)
	m.addr.SetWidth(m.Width() - (m.addr.Left() + 80))
	m.addr.SetBorderStyle(types.BsNone)
	m.addr.SetColor(color)
	m.addr.SetAnchors(types.NewSet(types.AkLeft, types.AkTop, types.AkRight))
	m.addr.Font().SetSize(17)
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
		addrLeft.SetColor(dkColor)
		addrRight.SetColor(dkColor)
		addrLeft.Invalidate()
		addrRight.Invalidate()
	}
	addrLeave := func(sender lcl.IObject) {
		m.addr.SetColor(color)
		addrLeft.SetColor(color)
		addrRight.SetColor(color)
		addrLeft.Invalidate()
		addrRight.Invalidate()
	}

	addrLeft = wg.NewButton(m)
	addrLeft.SetParent(m.box)
	addrLeftRect := types.TRect{Left: 140, Top: top}
	addrLeftRect.SetSize(30, height)
	addrLeft.SetBoundsRect(addrLeftRect)
	addrLeft.SetColor(color)
	addrLeft.SetRadius(15)
	addrLeft.SetAlpha(255)
	addrLeft.SetDisable(true)
	addrLeft.RoundedCorner = addrLeft.RoundedCorner.Exclude(wg.RcRightBottom).Exclude(wg.RcRightTop)
	addrLeft.SetOnClick(addrFocus)
	addrLeft.SetOnMouseEnter(addrEnter)
	addrLeft.SetOnMouseLeave(addrLeave)

	addrRight = wg.NewButton(m)
	addrRight.SetParent(m.box)
	addrRightRect := types.TRect{Left: m.addr.Left() + m.addr.Width(), Top: top}
	addrRightRect.SetSize(30, height)
	addrRight.SetBoundsRect(addrRightRect)
	addrRight.SetColor(color)
	addrRight.SetRadius(15)
	addrRight.SetAlpha(255)
	addrRight.SetDisable(true)
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
	m.addrRightBtn.SetColor(bgColor)
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

func (m *BrowserWindow) updateRefreshBtn(chromium *Chromium, isLoading bool) {
	if isLoading {
		lcl.RunOnMainThreadAsync(func(id uint32) {
			chromium.mainWindow.refreshBtn.SetIcon(getResourcePath("stop.png"))
		})
	} else {
		lcl.RunOnMainThreadAsync(func(id uint32) {
			chromium.mainWindow.refreshBtn.SetIcon(getResourcePath("refresh.png"))
		})
	}
}

// 清空地址栏 和 还原控制按钮
func (m *BrowserWindow) resetControlBtn() {
	m.addr.SetText("")
	m.backBtn.SetDisable(true)
	m.forwardBtn.SetDisable(true)
	m.backBtn.SetIcon(getResourcePath("back_disable.png"))
	m.backBtn.Invalidate()
	m.forwardBtn.SetIcon(getResourcePath("forward_disable.png"))
	m.forwardBtn.Invalidate()
	m.refreshBtn.SetIcon(getResourcePath("refresh.png"))
	m.refreshBtn.Invalidate()
	m.updateWindowCaption("")
}

func (m *Chromium) updateTabSheetActive(isActive bool) {
	if isActive {
		activeColor := colors.RGBToColor(86, 88, 93)
		m.tabSheetBtn.SetColor(activeColor)
		m.tabSheet.SetVisible(true)
		m.isActive = true
		m.mainWindow.SetAddrText(m.currentURL)
		m.mainWindow.updateWindowCaption(m.currentTitle)
		m.resize(nil)
	} else {
		notActiveColor := bgColor //colors.RGBToColor(56, 57, 60)
		m.tabSheetBtn.SetColor(notActiveColor)
		m.tabSheet.SetVisible(false)
		m.isActive = false
	}
	m.tabSheetBtn.Invalidate()
	// 根据当前 chromium 浏览器加载状态更新浏览器控制按钮
	m.updateBrowserControlBtn()
}

// 根据当前 chromium 浏览器加载状态更新浏览器控制按钮
func (m *Chromium) updateBrowserControlBtn() {
	m.mainWindow.backBtn.SetDisable(!m.canGoBack)
	m.mainWindow.forwardBtn.SetDisable(!m.canGoForward)
	backDisable := !m.canGoBack
	forwardDisable := !m.canGoForward
	lcl.RunOnMainThreadAsync(func(id uint32) {
		// 退回按钮
		if backDisable {
			// 禁用
			m.mainWindow.backBtn.SetIcon(getResourcePath("back_disable.png"))
		} else {
			m.mainWindow.backBtn.SetIcon(getResourcePath("back.png"))
		}
		m.mainWindow.backBtn.Invalidate()
		// 前进按钮
		if forwardDisable {
			// 禁用
			m.mainWindow.forwardBtn.SetIcon(getResourcePath("forward_disable.png"))
		} else {
			m.mainWindow.forwardBtn.SetIcon(getResourcePath("forward.png"))
		}
		m.mainWindow.forwardBtn.Invalidate()
	})
}

//func (m *BrowserWindow) macOSToolbar() {
//	// 空实现
//}
