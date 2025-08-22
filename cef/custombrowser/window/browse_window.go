package window

import (
	"github.com/energye/lcl/api"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/tool"
	"github.com/energye/lcl/types"
	"github.com/energye/lcl/types/colors"
	"os"
	"path/filepath"
	"sync"
	"widget/wg"
)

type BrowserWindow struct {
	Window
	box lcl.IPanel
	//content      lcl.IPanel
	mainWindowId int32 // 窗口ID
	windowId     int
	chroms       []*Chromium // 当前的chrom列表
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
}

var (
	BW      BrowserWindow
	bgColor = colors.RGBToColor(56, 57, 60)
)

func (m *BrowserWindow) FormCreate(sender lcl.IObject) {
	m.SetWidth(800)
	m.SetHeight(600)
	m.SetDoubleBuffered(true)
	{
		// 控制窗口显示鼠标所在显示器
		centerOnMonitor := func(monitor lcl.IMonitor) {
			m.SetLeft(monitor.Left() + (monitor.Width()-m.Width())/2)
			m.SetTop(monitor.Top() + (monitor.Height()-m.Height())/2)
		}
		mousePos := lcl.Mouse.CursorPos()
		var (
			i         int32 = 0
			defaultOK       = true
		)
		for ; i < lcl.Screen.MonitorCount(); i++ {
			if tempMonitor := lcl.Screen.Monitors(i); tempMonitor.WorkareaRect().PtInRect(mousePos) {
				defaultOK = false
				centerOnMonitor(tempMonitor)
				break
			}
		}
		if defaultOK {
			centerOnMonitor(lcl.Screen.PrimaryMonitor())
		}
	}
	m.SetColor(bgColor)
	if !tool.IsDarwin() {
		m.SetCaption("ENERGY-3.0-浏览器")
	}
	constraints := m.Constraints()
	constraints.SetMinWidth(600)
	constraints.SetMinHeight(400)

	m.box = lcl.NewPanel(m)
	m.box.SetParent(m)
	m.box.SetBevelOuter(types.BvNone)
	m.box.SetDoubleBuffered(true)
	m.box.SetColor(bgColor)
	if tool.IsDarwin() {
		// macos 窗口标题栏自定义了
		// 留高 45
		m.box.SetWidth(m.Width())
		m.box.SetHeight(m.Height() - 45)
		m.box.SetTop(45)
		m.box.SetAnchors(types.NewSet(types.AkLeft, types.AkTop, types.AkRight, types.AkBottom))
	} else {
		m.box.SetWidth(m.Width())
		m.box.SetHeight(m.Height())
		m.box.SetAnchors(types.NewSet(types.AkLeft, types.AkTop, types.AkRight, types.AkBottom))
	}
	// 窗口 拖拽 大小调整
	m.boxDrag()

	m.SetOnShow(func(sender lcl.IObject) {
		m.SetFocus()
		m.SetActiveDefaultControl(m)
	})
	m.SetOnActivate(func(sender lcl.IObject) {
		if !tool.IsDarwin() {
			if iconData, err := os.ReadFile(getResourcePath("window-icon_256x256.png")); err == nil {
				stream := lcl.NewMemoryStream()
				lcl.StreamHelper.Write(stream, iconData)
				stream.SetPosition(0)
				png := lcl.NewPortableNetworkGraphic()
				png.LoadFromStreamWithStream(stream)
				lcl.Application.Icon().Assign(png)
				png.Free()
				stream.Free()
			}
		} else {
			m.TestTool()
		}
		newChromium := m.createChromium("")
		m.OnChromiumCreateTabSheet(newChromium)
		newChromium.createBrowser(nil)
	})

	m.createTitleWidgetControl()

	m.SetOnResize(func(sender lcl.IObject) {
		// 重新计算 tab sheet left 和 width
		m.recalculateTabSheet()
		// 更新窗口控制按钮状态
		m.updateWindowControlBtn()
		if chrom := m.getActiveChrom(); chrom != nil {
			chrom.resize(sender)
		}
		//for _, chrom := range m.chroms {
		//	chrom.resize(sender)
		//}
		if Resize != nil {
			Resize()
		}
	})
	m.SetOnClose(func(sender lcl.IObject, closeAction *types.TCloseAction) {
		println("window.OnClose")
	})
	m.SetOnCloseQuery(func(sender lcl.IObject, canClose *bool) {
		println("window.OnCloseQuery 当前浏览器数量:", len(m.chroms))
		*canClose = len(m.chroms) == 0
		if tool.IsDarwin() {
			for _, chrom := range m.chroms {
				chrom.closeBrowser()
			}
			*canClose = true
		}
	})
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

// 更新窗口控制按钮状态
func (m *BrowserWindow) updateWindowControlBtn() {
	if m.maxBtn == nil {
		return
	}
	if m.WindowState() == types.WsMaximized {
		m.maxBtn.SetHint("向下还原")
		m.maxBtn.SetIcon(getResourcePath("btn-max-re.png"))
	} else if m.WindowState() == types.WsNormal {
		m.maxBtn.SetIcon(getResourcePath("btn-max.png"))
		m.maxBtn.SetHint("最大化")
	}
}

func (m *BrowserWindow) createTitleWidgetControl() {
	// 添加 chromium 按钮
	{
		m.addChromBtn = wg.NewButton(m)
		m.addChromBtn.SetParent(m.box)
		addBtnRect := types.TRect{Left: 5, Top: 5}
		addBtnRect.SetSize(40, 40)
		m.addChromBtn.SetBoundsRect(addBtnRect)
		m.addChromBtn.SetStartColor(bgColor)
		m.addChromBtn.SetEndColor(bgColor)
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
	}
	// 窗口控制按钮 最小化，最大化，关闭
	{
		createWindowControlBtn := func() {
			m.minBtn = wg.NewButton(m)
			m.minBtn.SetParent(m.box)
			m.minBtn.SetShowHint(true)
			m.minBtn.SetHint("最小化")
			minBtnRect := types.TRect{Left: m.box.Width() - 45*3, Top: 5}
			minBtnRect.SetSize(40, 40)
			m.minBtn.SetBoundsRect(minBtnRect)
			m.minBtn.SetStartColor(bgColor)
			m.minBtn.SetEndColor(bgColor)
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
			m.maxBtn.SetStartColor(bgColor)
			m.maxBtn.SetEndColor(bgColor)
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
			m.closeBtn.SetStartColor(bgColor)
			m.closeBtn.SetEndColor(bgColor)
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
		}
		if !tool.IsDarwin() {
			createWindowControlBtn()
		}
	}
	// 浏览器控制按钮
	{
		// 后退
		m.backBtn = wg.NewButton(m)
		m.backBtn.SetParent(m.box)
		m.backBtn.SetShowHint(true)
		m.backBtn.SetHint("单击返回")
		backBtnRect := types.TRect{Left: 5, Top: 47}
		backBtnRect.SetSize(40, 40)
		m.backBtn.SetBoundsRect(backBtnRect)
		m.backBtn.SetStartColor(bgColor)
		m.backBtn.SetEndColor(bgColor)
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
		m.forwardBtn.SetStartColor(bgColor)
		m.forwardBtn.SetEndColor(bgColor)
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
		m.refreshBtn.SetStartColor(bgColor)
		m.refreshBtn.SetEndColor(bgColor)
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
	}
	// 地址栏
	m.createAddrBar()
}

// 浏览器创建完添加一个 tab Sheet
func (m *BrowserWindow) OnChromiumCreateTabSheet(newChromium *Chromium) {
	m.chroms = append(m.chroms, newChromium)
	newChromium.windowId = int32(len(m.chroms))
	//fmt.Println("OnChromiumCreateTabSheet", "当前chromium数量:", len(m.chroms), "新chromiumID:", newChromium.windowId)
	m.AddTabSheet(newChromium)
}

func (m *BrowserWindow) removeTabSheetBrowse(chromium *Chromium) {
	m.browserCloseLock.Lock()
	defer m.browserCloseLock.Unlock()
	var isCloseCurrentActive bool
	if activeChrom := m.getActiveChrom(); activeChrom != nil && activeChrom == chromium {
		isCloseCurrentActive = true
	}
	// 删除当前chrom, 使用 windowId - 1 是当前 chrom 所在下标
	idx := chromium.windowId - 1
	// 删除
	m.chroms = append(m.chroms[:idx], m.chroms[idx+1:]...)
	// 重新设置每个 chromium 的 windowID, 在下次删除时能对应上
	for id, chrom := range m.chroms {
		chrom.windowId = int32(id + 1)
	}
	if len(m.chroms) > 0 {
		// 判断关闭时tabSheet是否为当前激活的
		// 如果是当前激活的，激活最后一个
		if isCloseCurrentActive {
			// 激活最后一个
			lastChrom := m.chroms[len(m.chroms)-1]
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

func (m *BrowserWindow) AddTabSheet(currentChromium *Chromium) {
	// 当前的设置为激活状态（颜色控制）
	var leftSize int32 = 5
	for _, chrom := range m.chroms {
		if chrom.tabSheetBtn != nil {
			leftSize += chrom.tabSheetBtn.Width() + 5
		}
	}

	// 创建新 tabSheet
	newTabSheet := wg.NewButton(m)
	newTabSheet.SetParent(m.box)
	newTabSheet.SetShowHint(true)
	newTabSheet.SetCaption("新建标签页")
	newTabSheet.Font().SetSize(12)
	newTabSheet.Font().SetColor(colors.Cl3DFace)
	newTabSheetRect := types.TRect{Left: leftSize, Top: 5}
	newTabSheetRect.SetSize(0, 0)
	newTabSheet.SetBoundsRect(newTabSheetRect)
	newTabSheet.SetStartColor(colors.RGBToColor(86, 88, 93))
	newTabSheet.SetEndColor(colors.RGBToColor(86, 88, 93))
	newTabSheet.RoundedCorner = newTabSheet.RoundedCorner.Exclude(wg.RcLeftBottom).Exclude(wg.RcRightBottom)
	newTabSheet.SetIconFavorite(getResourcePath("icon.png"))
	newTabSheet.SetIconClose(getResourcePath("sheet_close.png"))
	newTabSheet.SetOnCloseClick(func(sender lcl.IObject) {
		if m.isChromCloseing {
			// 当前有正在关闭的浏览器
			println("有正在关闭的浏览器")
			return
		}
		m.isChromCloseing = true
		go lcl.RunOnMainThreadAsync(func(id uint32) {
			currentChromium.closeBrowser()
		})
	})
	newTabSheet.SetOnClick(func(sender lcl.IObject) {
		// tab sheet 按钮点击
		// 更新其它 tabSheetBtn 不激活, 当前为激活显示
		m.updateOtherTabSheetNoActive(currentChromium)
		// 更新当前 chromium tabSheetBtn激活
		currentChromium.updateTabSheetActive(true)
	})
	currentChromium.isActive = true           // 设置默认激活
	currentChromium.tabSheetBtn = newTabSheet // 绑定到当前 chromium

	// 更新其它tabSheet 非激活状态(颜色控制)
	m.updateOtherTabSheetNoActive(currentChromium)

	// 重新计算 tab sheet left 和 width
	m.recalculateTabSheet()
}

// 清空地址栏 和 还原控制按钮
func (m *BrowserWindow) resetControlBtn() {
	m.addr.SetText("")
	m.backBtn.IsDisable = true
	m.forwardBtn.IsDisable = true
	m.backBtn.SetIcon(getResourcePath("back_disable.png"))
	m.backBtn.Invalidate()
	m.forwardBtn.SetIcon(getResourcePath("forward_disable.png"))
	m.forwardBtn.Invalidate()
	m.refreshBtn.SetIcon(getResourcePath("refresh.png"))
	m.refreshBtn.Invalidate()
	m.updateWindowCaption("")
}

func (m *BrowserWindow) updateWindowCaption(title string) {
	if tool.IsDarwin() {
		return
	}
	lcl.RunOnMainThreadAsync(func(id uint32) {
		if title == "" {
			title = "ENERGY-3.0-浏览器"
		}
		m.SetCaption(title)
	})
}

// 重新计算 tab sheet left 和 width
func (m *BrowserWindow) recalculateTabSheet() {
	var (
		minWidth, maxWidth int32 = 40, 230       // 最小宽度, 最大宽度
		rightSize          int32 = 40 * 6        // 添加按钮，最小化，最大化，关闭按钮的预留位置
		widht                    = m.box.Width() // 当前窗口宽
		leftSize           int32 = 0             // 默认 间距
	)
	areaWidth := widht - rightSize // 区域可用宽度
	count := int32(len(m.chroms))
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

	for _, chrom := range m.chroms {
		if chrom.tabSheetBtn != nil {
			chrom.tabSheetBtn.SetBounds(leftSize+5, 5, avgWidth, 40)
			leftSize += avgWidth
		}
	}

	// 更新添加按钮位置
	m.updateBtnLeft()
}

// 获得当前激活的 chrom
func (m *BrowserWindow) getActiveChrom() *Chromium {
	var result *Chromium
	for _, chrom := range m.chroms {
		if chrom.isActive {
			result = chrom
		}
	}
	return result
}

// 更新其它 tab sheet 状态
func (m *BrowserWindow) updateOtherTabSheetNoActive(currentChromium *Chromium) {
	for _, chrom := range m.chroms {
		if chrom != currentChromium {
			chrom.updateTabSheetActive(false)
		}
	}
}

// 更新 添加按钮位置
func (m *BrowserWindow) updateBtnLeft() {
	var leftSize int32 = 0
	for _, chrom := range m.chroms {
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

var (
	wd, _        = os.Getwd()
	CacheRoot    string
	SiteResource string
)

func getResourcePath(name string) string {
	var sourcePath string
	sourcePath = filepath.Join(wd, "resources", name)
	if tool.IsExist(sourcePath) {
		return sourcePath
	}
	sourcePath = filepath.Join("./", "resources", name)
	if tool.IsExist(sourcePath) {
		return sourcePath
	}
	if tool.IsWindows() {
		sourcePath = filepath.Join("E:\\SWT\\gopath\\src\\github.com\\energye\\workspace\\examples\\cef\\custombrowser\\resources", name)
	} else if tool.IsLinux() {
		sourcePath = filepath.Join("/home/yanghy/app/gopath/src/github.com/energye/workspace/examples/cef/custombrowser/resources", name)
	} else if tool.IsDarwin() {
		sourcePath = filepath.Join("/Users/yanghy/app/workspace/examples/cef/custombrowser/resources", name)
	}
	if tool.IsExist(sourcePath) {
		return sourcePath
	}
	return ""
}

func printIsMainThread() {
	println("printIsMainThread.isMainThread:", api.MainThreadId() == api.CurrentThreadId())
}
