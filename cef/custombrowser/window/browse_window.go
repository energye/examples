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
	m.SetWidth(1024)
	m.SetHeight(700)
	m.SetDoubleBuffered(true)
	{
		// 控制窗口显示鼠标所在显示器
		centerOnMonitor := func(monitor lcl.IMonitor) {
			m.SetLeft(monitor.Left() + (monitor.Width()-m.Width())/2)
			top := monitor.Top() + (monitor.Height()-m.Height())/2
			m.SetTop(top)
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
	if !isDarwin {
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
	if isDarwin {
		// macos 窗口标题栏自定义
		m.box.SetWidth(m.Width())
		m.box.SetHeight(m.Height())
		m.box.SetAlign(types.AlClient)
	} else {
		m.box.SetWidth(m.Width())
		m.box.SetHeight(m.Height())
		m.box.SetAnchors(types.NewSet(types.AkLeft, types.AkTop, types.AkRight, types.AkBottom))
		// 窗口 拖拽 大小调整
		m.boxDrag()
	}

	m.SetOnActivate(func(sender lcl.IObject) {
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
		m.macOSToolbar()
		newChromium := m.createChromium("")
		m.OnChromiumCreateTabSheet(newChromium)
		newChromium.createBrowser(nil)
	})
	// 创建标题栏控件
	m.createTitleWidgetControl()
	m.SetOnResize(func(sender lcl.IObject) {
		// 重新计算 tab sheet left 和 width
		m.recalculateTabSheet()
		if !isDarwin {
			// 更新窗口控制按钮状态
			m.updateWindowControlBtn()
		}
		if chrom := m.getActiveChrom(); chrom != nil {
			chrom.resize(sender)
		}
		//for _, chrom := range m.chroms {
		//	chrom.resize(sender)
		//}
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
	if m.WindowState() == types.WsMaximized {
		m.maxBtn.SetHint("向下还原")
		m.maxBtn.SetIcon(getResourcePath("btn-max-re.png"))
	} else if m.WindowState() == types.WsNormal {
		m.maxBtn.SetIcon(getResourcePath("btn-max.png"))
		m.maxBtn.SetHint("最大化")
	}
}

// 浏览器创建完添加一个 tab Sheet
func (m *BrowserWindow) OnChromiumCreateTabSheet(newChromium *Chromium) {
	m.chroms = append(m.chroms, newChromium)
	newChromium.windowId = int32(len(m.chroms))
	//fmt.Println("OnChromiumCreateTabSheet", "当前chromium数量:", len(m.chroms), "新chromiumID:", newChromium.windowId)
	m.AddTabSheetBtn(newChromium)
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

func (m *BrowserWindow) AddTabSheetBtn(currentChromium *Chromium) {
	// 当前的设置为激活状态（颜色控制）
	var leftSize int32 = 5
	for _, chrom := range m.chroms {
		if chrom.tabSheetBtn != nil {
			leftSize += chrom.tabSheetBtn.Width() + 5
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
	newTabSheetBtn.SetIconFavorite(getResourcePath("icon.png"))
	newTabSheetBtn.SetIconClose(getResourcePath("sheet_close.png"))
	newTabSheetBtn.SetOnCloseClick(func(sender lcl.IObject) {
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
	newTabSheetBtn.SetOnClick(func(sender lcl.IObject) {
		// tab sheet 按钮点击
		// 更新其它 tabSheetBtn 不激活, 当前为激活显示
		m.updateOtherTabSheetNoActive(currentChromium)
		// 更新当前 chromium tabSheetBtn激活
		currentChromium.updateTabSheetActive(true)
	})
	currentChromium.isActive = true              // 设置默认激活
	currentChromium.tabSheetBtn = newTabSheetBtn // 绑定到当前 chromium

	// 更新其它tabSheet 非激活状态(颜色控制)
	m.updateOtherTabSheetNoActive(currentChromium)

	// 重新计算 tab sheet left 和 width
	m.recalculateTabSheet()
}

func (m *BrowserWindow) updateWindowCaption(title string) {
	if isDarwin {
		return
	}
	lcl.RunOnMainThreadAsync(func(id uint32) {
		if title == "" {
			title = "ENERGY-3.0-浏览器"
		}
		m.SetCaption(title)
	})
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
			chrom.tabSheetBtn.SetBounds(leftSize+5, 5, avgWidth, tabSheetBtnHeight)
			leftSize += avgWidth
		}
	}

	// 更新添加按钮位置
	if !isDarwin {
		m.updateBtnLeft()
	}
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
