package window

import (
	"github.com/energye/examples/wv/assets"
	"github.com/energye/lcl/api"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/pkgs/win"
	"github.com/energye/lcl/types"
	"github.com/energye/lcl/types/colors"
	"github.com/energye/lcl/types/messages"
	"github.com/energye/widget/wg"
	"net/url"
	"strings"
)

func (m *BrowserWindow) Minimize() {
	m.SetWindowState(types.WsMinimized)
}

func (m *BrowserWindow) Maximize() {
	if m.WindowState() == types.WsNormal {
		m.SetWindowState(types.WsMaximized)
	} else {
		m.SetWindowState(types.WsNormal)
	}
}

func (m *BrowserWindow) FullScreen() {
	if m.WindowState() == types.WsMinimized || m.WindowState() == types.WsMaximized {
		if win.ReleaseCapture() {
			win.SendMessage(m.Handle(), messages.WM_SYSCOMMAND, messages.SC_RESTORE, 0)
		}
	}
	m.windowState = types.WsFullScreen
	m.normalBounds = m.BoundsRect()
	monitorRect := m.Monitor().BoundsRect()
	win.SetWindowPos(m.Handle(), win.HWND_TOP, monitorRect.Left, monitorRect.Top, monitorRect.Width(), monitorRect.Height(), win.SWP_NOOWNERZORDER|win.SWP_FRAMECHANGED)
}

func (m *BrowserWindow) ExitFullScreen() {
	if m.IsFullScreen() {
		m.windowState = types.WsNormal
		m.SetWindowState(types.WsNormal)
		m.SetBoundsRect(m.normalBounds)
	}
}

func (m *BrowserWindow) IsFullScreen() bool {
	return m.windowState == types.WsFullScreen
}

func (m *BrowserWindow) boxDblClick(sender lcl.IObject) {
	if m.isTitleBar {
		if m.WindowState() == types.WsNormal {
			m.SetWindowState(types.WsMaximized)
		} else {
			m.SetWindowState(types.WsNormal)
		}
	}
}

func (m *BrowserWindow) boxMouseMove(sender lcl.IObject, shift types.TShiftState, x, y int32) {
	lcl.Screen.SetCursor(types.CrDefault)
	// 判断鼠标所在区域
	rect := m.BoundsRect()
	if x > m.borderWidth && y > m.borderWidth && x < rect.Width()-m.borderWidth && y < rect.Height()-m.borderWidth && y < m.titleHeight {
		// 标题栏部分
		if m.isDown {
			if win.ReleaseCapture() {
				win.PostMessage(m.Handle(), messages.WM_NCLBUTTONDOWN, messages.HTCAPTION, 0)
			}
		}
		m.borderHT = 0 // 重置边框标记
		m.isTitleBar = true
	} else {
		m.isTitleBar = false
		// 边框区域判断 (8个区域)
		switch {
		// 角落区域 (优先判断)
		case x < m.borderWidth && y < m.borderWidth:
			m.borderHT = messages.HTTOPLEFT
			lcl.Screen.SetCursor(types.CrSizeNWSE)
		case x > rect.Width()-m.borderWidth && y < m.borderWidth:
			m.borderHT = messages.HTTOPRIGHT
			lcl.Screen.SetCursor(types.CrSizeNESW)
		case x < m.borderWidth && y > rect.Height()-m.borderWidth:
			m.borderHT = messages.HTBOTTOMLEFT
			lcl.Screen.SetCursor(types.CrSizeNESW)
		case x > rect.Width()-m.borderWidth && y > rect.Height()-m.borderWidth:
			m.borderHT = messages.HTBOTTOMRIGHT
			lcl.Screen.SetCursor(types.CrSizeNWSE)
		// 边缘区域
		case y < m.borderWidth:
			m.borderHT = messages.HTTOP
			lcl.Screen.SetCursor(types.CrSizeNS)
		case y > rect.Height()-m.borderWidth:
			m.borderHT = messages.HTBOTTOM
			lcl.Screen.SetCursor(types.CrSizeNS)
		case x < m.borderWidth:
			m.borderHT = messages.HTLEFT
			lcl.Screen.SetCursor(types.CrSizeWE)
		case x > rect.Width()-m.borderWidth:
			m.borderHT = messages.HTRIGHT
			lcl.Screen.SetCursor(types.CrSizeWE)
		default:
			m.borderHT = 0 // 客户区
		}
	}
}

func (m *BrowserWindow) boxMouseDown(sender lcl.IObject, button types.TMouseButton, shift types.TShiftState, x, y int32) {
	m.isDown = true
	if m.borderHT != 0 {
		if win.ReleaseCapture() {
			win.PostMessage(m.Handle(), messages.WM_NCLBUTTONDOWN, m.borderHT, 0)
		}
	}
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
	m.addChromBtn.SetIcon(assets.GetResourcePath("add.png"))
	m.addChromBtn.SetOnClick(func(sender lcl.IObject) {
		println("add chromium isMainThread:", api.MainThreadId() == api.CurrentThreadId())
		m.addr.SetText("")
		newChromium := m.CreateBrowser("")
		m.OnChromiumCreateTabSheet(newChromium)
		newChromium.Create()
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
	m.minBtn.SetIcon(assets.GetResourcePath("btn-min.png"))
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
	m.maxBtn.SetIcon(assets.GetResourcePath("btn-max.png"))
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
	m.closeBtn.SetIcon(assets.GetResourcePath("btn-close.png"))
	m.closeBtn.SetOnClick(func(sender lcl.IObject) {
		if len(m.browses) == 0 {
			m.Close()
		} else {
			// 稳妥的关闭方式
			for {
				count := len(m.browses)
				if count == 0 {
					break
				}
				m.browses[0].CloseBrowse()
			}
			// 最后关闭窗口
			m.Close()
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
	m.backBtn.SetIcon(assets.GetResourcePath("back.png"))
	m.backBtn.SetOnClick(func(sender lcl.IObject) {
		browse := m.getActiveBrowse()
		if browse != nil && browse.browser.CanGoBack() {
			browse.browser.GoBack()
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
	m.forwardBtn.SetIcon(assets.GetResourcePath("forward.png"))
	m.forwardBtn.SetOnClick(func(sender lcl.IObject) {
		browse := m.getActiveBrowse()
		if browse != nil && browse.browser.CanGoForward() {
			browse.browser.GoForward()
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
	m.refreshBtn.SetIcon(assets.GetResourcePath("refresh.png"))
	m.refreshBtn.SetOnClick(func(sender lcl.IObject) {
		browse := m.getActiveBrowse()
		if browse != nil {
			if browse.isLoading {
				browse.browser.Stop()
			} else {
				browse.browser.Refresh()
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

			if uri, err := url.Parse(tempUrl); err != nil || tempUrl == "" {
				tempUrl = "https://energye.github.io/"
			} else {
				if uri.Scheme == "" {
					tempUrl = "http://" + tempUrl
				}
			}
			if browse := m.getActiveBrowse(); browse != nil {
				browse.browser.Navigate(tempUrl)
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
	m.addrRightBtn.SetIcon(assets.GetResourcePath("addr-right-btn.png"))
	m.addrRightBtn.SetOnClick(func(sender lcl.IObject) {
		if browse := m.getActiveBrowse(); browse != nil {
			browse.browser.Navigate("https://energye.github.io")
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
