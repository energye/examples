package window

import (
	"fmt"
	"github.com/energye/examples/wv/assets"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/pkgs/win"
	"github.com/energye/lcl/types"
	"github.com/energye/lcl/types/colors"
	"github.com/energye/lcl/types/messages"
	wv "github.com/energye/wv/windows"
	"sync"
	"syscall"
	"unsafe"
	"widget/wg"
)

var (
	Window  BrowserWindow
	Load    wv.IWVLoader
	bgColor = colors.RGBToColor(56, 57, 60)
)

type BrowserWindow struct {
	lcl.TEngForm
	box          lcl.IPanel
	mainWindowId int32 // 窗口ID
	windowId     int
	browses      []*Browser  // 当前的chrom列表
	addChromBtn  *wg.TButton // 添加浏览器按钮
	// 浏览器控制按钮
	backBtn    *wg.TButton
	forwardBtn *wg.TButton
	refreshBtn *wg.TButton
	// 地址栏
	addr         lcl.IMemo
	addrRightBtn *wg.TButton
	// 窗口控制按钮
	minBtn   *wg.TButton
	maxBtn   *wg.TButton
	closeBtn *wg.TButton
	// 标题栏相关
	titleHeight           int32 // 标题栏高度
	borderWidth           int32 // 边框宽
	isDown                bool  // 鼠标按下和抬起
	isTitleBar, isDarging bool  // 窗口标题栏
	borderHT              uintptr
	// 窗口关闭锁，一个一个关闭
	browserCloseLock    sync.Mutex
	isWindowButtonClose bool // 点击的窗口关闭按钮
	isChromCloseing     bool // 当前是否有正在关闭的 chrom
	oldWndPrc           uintptr
	normalBounds        types.TRect
	windowState         types.TWindowState
}

func (m *BrowserWindow) FormCreate(sender lcl.IObject) {
	m.SetCaption("ENERGY 3.0 WebView2")
	m.SetPosition(types.PoScreenCenter)
	m.SetWidth(1024)
	m.SetHeight(768)
	m.SetDoubleBuffered(true)

	m.box = lcl.NewPanel(m)
	m.box.SetParent(m)
	m.box.SetBevelOuter(types.BvNone)
	m.box.SetDoubleBuffered(true)
	m.box.SetColor(bgColor)
	m.box.SetWidth(m.Width())
	m.box.SetHeight(m.Height())
	m.box.SetAnchors(types.NewSet(types.AkLeft, types.AkTop, types.AkRight, types.AkBottom))
	m.boxDrag()
	m.createTitleWidgetControl()

	// 窗口显示时创建browser
	m.SetOnShow(func(sender lcl.IObject) {
		if Load.InitializationError() {
			fmt.Println("回调函数 => SetOnShow 初始化失败")
		} else {
			if Load.Initialized() {
				fmt.Println("回调函数 => SetOnShow 初始化成功")
				def := "file://" + assets.GetResourcePath("default.html")
				newBrowser := m.CreateBrowser(def)
				m.OnChromiumCreateTabSheet(newBrowser)
				newBrowser.Create()
			}
		}
	})

	m.SetOnResize(func(sender lcl.IObject) {
		// 重新计算 tab sheet left 和 width
		m.recalculateTabSheet()
		// 更新窗口控制按钮状态
		m.updateWindowControlBtn()
		if chrom := m.getActiveBrowse(); chrom != nil {
			chrom.resize(sender)
		}
	})
}

// 更新窗口控制按钮状态
func (m *BrowserWindow) updateWindowControlBtn() {
	if m.WindowState() == types.WsMaximized {
		m.maxBtn.SetHint("向下还原")
		m.maxBtn.SetIcon(assets.GetResourcePath("btn-max-re.png"))
	} else if m.WindowState() == types.WsNormal {
		m.maxBtn.SetIcon(assets.GetResourcePath("btn-max.png"))
		m.maxBtn.SetHint("最大化")
	}
}

// 浏览器创建完添加一个 tab Sheet
func (m *BrowserWindow) OnChromiumCreateTabSheet(newBrowse *Browser) {
	m.browses = append(m.browses, newBrowse)
	newBrowse.windowId = int32(len(m.browses))
	//fmt.Println("OnChromiumCreateTabSheet", "当前chromium数量:", len(m.chroms), "新chromiumID:", newChromium.windowId)
	m.AddTabSheetBtn(newBrowse)
}

func (m *BrowserWindow) AddTabSheetBtn(currentBrowse *Browser) {
	// 当前的设置为激活状态（颜色控制）
	var leftSize int32 = 5
	for _, browse := range m.browses {
		if browse.tabSheetBtn != nil {
			leftSize += browse.tabSheetBtn.Width() + 5
		}
	}

	// 创建新 tabSheet
	newTabSheetBtn := wg.NewButton(m)
	newTabSheetBtn.SetParent(m.box)
	newTabSheetBtn.SetShowHint(true)
	newTabSheetBtn.SetCaption("新建标签页")
	newTabSheetBtn.Font().SetSize(12)
	newTabSheetBtn.Font().SetColor(colors.Cl3DFace)
	newTabSheetRect := types.TRect{Left: leftSize, Top: 5}
	newTabSheetRect.SetSize(0, 0)
	newTabSheetBtn.SetBoundsRect(newTabSheetRect)
	newTabSheetBtn.SetStartColor(colors.RGBToColor(86, 88, 93))
	newTabSheetBtn.SetEndColor(colors.RGBToColor(86, 88, 93))
	newTabSheetBtn.RoundedCorner = newTabSheetBtn.RoundedCorner.Exclude(wg.RcLeftBottom).Exclude(wg.RcRightBottom)
	newTabSheetBtn.SetIconFavorite(assets.GetResourcePath("icon.png"))
	newTabSheetBtn.SetIconClose(assets.GetResourcePath("sheet_close.png"))
	newTabSheetBtn.SetOnCloseClick(func(sender lcl.IObject) {
		if m.isChromCloseing {
			// 当前有正在关闭的浏览器
			println("有正在关闭的浏览器")
			return
		}
		m.isChromCloseing = true
		lcl.RunOnMainThreadAsync(func(id uint32) {
			currentBrowse.CloseBrowse()
		})
	})
	newTabSheetBtn.SetOnClick(func(sender lcl.IObject) {
		// tab sheet 按钮点击
		// 更新其它 tabSheetBtn 不激活, 当前为激活显示
		m.updateOtherTabSheetNoActive(currentBrowse)
		// 更新当前 chromium tabSheetBtn激活
		currentBrowse.updateTabSheetActive(true)
	})
	currentBrowse.isActive = true              // 设置默认激活
	currentBrowse.tabSheetBtn = newTabSheetBtn // 绑定到当前 chromium

	// 更新其它tabSheet 非激活状态(颜色控制)
	m.updateOtherTabSheetNoActive(currentBrowse)

	// 重新计算 tab sheet left 和 width
	m.recalculateTabSheet()
}

func (m *BrowserWindow) removeTabSheetBrowse(browse *Browser) {
	m.browserCloseLock.Lock()
	defer m.browserCloseLock.Unlock()
	var isCloseCurrentActive bool
	if active := m.getActiveBrowse(); active != nil && active == browse {
		isCloseCurrentActive = true
	}
	// 删除当前chrom, 使用 windowId - 1 是当前 chrom 所在下标
	idx := browse.windowId - 1
	// 删除
	m.browses = append(m.browses[:idx], m.browses[idx+1:]...)
	// 重新设置每个 chromium 的 windowID, 在下次删除时能对应上
	for id, chrom := range m.browses {
		chrom.windowId = int32(id + 1)
	}
	if len(m.browses) > 0 {
		// 判断关闭时tabSheet是否为当前激活的
		// 如果是当前激活的，激活最后一个
		if isCloseCurrentActive {
			// 激活最后一个
			lastChrom := m.browses[len(m.browses)-1]
			lastChrom.updateTabSheetActive(true)
			// 其它的不激活
			m.updateOtherTabSheetNoActive(lastChrom)
		}
	} else {
		// 没有 chrom 清空和还原控制按钮、地址栏
		m.resetControlBtn()
	}
	// 重新计算 tab sheet left 和 width
	m.recalculateTabSheet()

	// 正在关闭浏览器完成
	m.isChromCloseing = false

	// 点击窗口的关闭按钮时尝试关闭窗口
	if m.isWindowButtonClose {
		// 尝试关闭窗口, 所有 chrom 都关闭后再关闭窗口
		m.Close()
	}
}

// 清空地址栏 和 还原控制按钮
func (m *BrowserWindow) resetControlBtn() {
	m.addr.SetText("")
	m.backBtn.IsDisable = true
	m.forwardBtn.IsDisable = true
	m.backBtn.SetIcon(assets.GetResourcePath("back_disable.png"))
	m.backBtn.Invalidate()
	m.forwardBtn.SetIcon(assets.GetResourcePath("forward_disable.png"))
	m.forwardBtn.Invalidate()
	m.refreshBtn.SetIcon(assets.GetResourcePath("refresh.png"))
	m.refreshBtn.Invalidate()
	m.updateWindowCaption("")
}

var (
	tabSheetBtnHeight    int32 = 40
	tabSheetBtnRightSize int32 = 40 * 6 // 添加按钮，最小化，最大化，关闭按钮的预留位置
)

// 重新计算 tab sheet left 和 width
func (m *BrowserWindow) recalculateTabSheet() {
	var (
		minWidth, maxWidth int32 = 40, 230       // 最小宽度, 最大宽度
		width                    = m.box.Width() // 当前窗口宽
		leftSize           int32 = 0             // 默认 间距
	)
	areaWidth := width - tabSheetBtnRightSize // 区域可用宽度
	count := int32(len(m.browses))
	if count == 0 {
		count = 1
	}
	avgWidth := areaWidth / int32(count)
	if avgWidth <= minWidth {
		avgWidth = minWidth
	}
	if avgWidth >= maxWidth {
		avgWidth = maxWidth
	}

	for _, chrom := range m.browses {
		if chrom.tabSheetBtn != nil {
			chrom.tabSheetBtn.SetBounds(leftSize+5, 5, avgWidth, tabSheetBtnHeight)
			leftSize += avgWidth
		}
	}

	m.updateBtnLeft()
}

// 更新 添加按钮位置
func (m *BrowserWindow) updateBtnLeft() {
	var leftSize int32 = 0
	for _, chrom := range m.browses {
		if chrom.tabSheetBtn != nil {
			leftSize += chrom.tabSheetBtn.Width()
		}
	}
	// 添加浏览器按钮, 保持在最后
	if m.addChromBtn != nil {
		m.addChromBtn.SetLeft(leftSize + 10)
	}
	// 窗口 最小化、最大化，关闭按钮
	if m.minBtn != nil {
		m.minBtn.SetLeft(m.box.Width() - 45*3)
		m.maxBtn.SetLeft(m.box.Width() - 45*2)
		m.closeBtn.SetLeft(m.box.Width() - 45)
	}
}
func (m *BrowserWindow) SetAddrText(val string) {
	lcl.RunOnMainThreadAsync(func(id uint32) {
		m.addr.SetText(val)
		m.addr.SetSelStart(int32(len(val)))
		m.addr.SetFocus()
	})
}

func (m *BrowserWindow) updateWindowCaption(title string) {
	lcl.RunOnMainThreadAsync(func(id uint32) {
		if title == "" {
			title = "ENERGY-3.0-浏览器"
		}
		m.SetCaption(title)
	})
}

// 更新其它 tab sheet 状态
func (m *BrowserWindow) updateOtherTabSheetNoActive(currentBrowse *Browser) {
	for _, browse := range m.browses {
		if browse != currentBrowse {
			browse.updateTabSheetActive(false)
		}
	}
}
func (m *BrowserWindow) updateRefreshBtn(chromium *Browser, isLoading bool) {
	if isLoading {
		lcl.RunOnMainThreadAsync(func(id uint32) {
			chromium.mainWindow.refreshBtn.SetIcon(assets.GetResourcePath("stop.png"))
		})
	} else {
		lcl.RunOnMainThreadAsync(func(id uint32) {
			chromium.mainWindow.refreshBtn.SetIcon(assets.GetResourcePath("refresh.png"))
		})
	}
}

// 获得当前激活的 chrom
func (m *BrowserWindow) getActiveBrowse() *Browser {
	var result *Browser
	for _, chrom := range m.browses {
		if chrom.isActive {
			result = chrom
		}
	}
	return result
}

func (m *BrowserWindow) FormAfterCreate(sender lcl.IObject) {
	m.HookWndProcMessage()
}

func (m *BrowserWindow) wndProc(hwnd types.HWND, message uint32, wParam, lParam uintptr) uintptr {
	switch message {
	case messages.WM_DPICHANGED:
		if !lcl.Application.Scaled() {
			newWindowSize := (*types.TRect)(unsafe.Pointer(lParam))
			win.SetWindowPos(m.Handle(), uintptr(0),
				newWindowSize.Left, newWindowSize.Top, newWindowSize.Right-newWindowSize.Left, newWindowSize.Bottom-newWindowSize.Top,
				win.SWP_NOZORDER|win.SWP_NOACTIVATE)
		}
		return 0 // 确保处理WM_DPICHANGED后返回

	case messages.WM_ACTIVATE:
		win.ExtendFrameIntoClientArea(m.Handle(), win.Margins{CxLeftWidth: 1, CxRightWidth: 1, CyTopHeight: 1, CyBottomHeight: 1})
		return 0

	case messages.WM_NCCALCSIZE:
		if wParam != 0 {
			isMaximize := uint32(win.GetWindowLong(m.Handle(), win.GWL_STYLE))&win.WS_MAXIMIZE != 0
			if isMaximize {
				rect := (*types.TRect)(unsafe.Pointer(lParam))
				monitor := win.MonitorFromRect(rect, win.MONITOR_DEFAULTTONULL)
				if monitor != 0 {
					var monitorInfo types.TMonitorInfo
					monitorInfo.CbSize = types.DWORD(unsafe.Sizeof(monitorInfo))
					if win.GetMonitorInfo(monitor, &monitorInfo) {
						*rect = monitorInfo.RcWork
					}
				}
			}
			return 0 // 移除标准边框
		}

	}

	return win.CallWindowProc(m.oldWndPrc, uintptr(hwnd), message, wParam, lParam)
}

func (m *BrowserWindow) HookWndProcMessage() {
	wndProcCallback := syscall.NewCallback(m.wndProc)
	m.oldWndPrc = win.SetWindowLongPtr(m.Handle(), win.GWL_WNDPROC, wndProcCallback)
	// trigger WM_NCCALCSIZE
	// https://learn.microsoft.com/en-us/windows/win32/dwm/customframe#removing-the-standard-frame
	clientRect := m.BoundsRect()
	win.SetWindowPos(m.Handle(), 0, clientRect.Left, clientRect.Top, clientRect.Right-clientRect.Left, clientRect.Bottom-clientRect.Top, win.SWP_FRAMECHANGED|win.SWP_NOACTIVATE)
}

// box 容器 窗口 拖拽 大小调整
func (m *BrowserWindow) boxDrag() {
	m.titleHeight = 45 // 标题栏高度
	m.borderWidth = 5  // 边框宽

	m.box.SetOnMouseMove(m.boxMouseMove)
	m.box.SetOnDblClick(m.boxDblClick)
	m.box.SetOnMouseDown(m.boxMouseDown)
	m.box.SetOnMouseUp(func(sender lcl.IObject, button types.TMouseButton, shift types.TShiftState, x, y int32) {
		m.isDown = false
	})
}
