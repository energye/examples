package window

import (
	"fmt"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/tool"
	"github.com/energye/lcl/types"
	"github.com/energye/lcl/types/colors"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"widget/wg"
)

type BrowserWindow struct {
	Window
	box          lcl.IPanel
	content      lcl.IPanel
	mainWindowId int32 // 窗口ID
	canClose     bool
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
	// 窗口大小变化记录
	previousWindowPlacement types.TRect
	windowState             types.TWindowState
	//
	titleHeight        int32 // 标题栏高度
	borderWidth        int32 // 边框宽
	isDown, isTitleBar bool  // 鼠标按下和抬起
	borderHT           uintptr
}

var (
	BW BrowserWindow
)

func (m *BrowserWindow) FormCreate(sender lcl.IObject) {
	m.SetWidth(1200)
	m.SetHeight(800)
	m.SetDoubleBuffered(true)
	//m.SetColor(colors.ClYellow)
	m.SetColor(colors.RGBToColor(56, 57, 60))
	m.WorkAreaCenter()
	m.SetCaption("ENERGY-3.0-浏览器")

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

	constraints := m.Constraints()
	constraints.SetMinWidth(800)
	constraints.SetMinHeight(600)

	m.box = lcl.NewPanel(m)
	m.box.SetParent(m)
	m.box.SetBevelOuter(types.BvNone)
	m.box.SetDoubleBuffered(true)
	m.box.SetWidth(m.Width())
	m.box.SetHeight(m.Height())
	m.box.SetColor(colors.RGBToColor(56, 57, 60))
	m.box.SetAnchors(types.NewSet(types.AkLeft, types.AkTop, types.AkRight, types.AkBottom))

	// 窗口 拖拽 大小调整
	m.boxDrag()

	m.content = lcl.NewPanel(m)
	m.content.SetParent(m.box)
	m.content.SetBevelOuter(types.BvNone)
	m.content.SetDoubleBuffered(true)
	m.content.SetAnchors(types.NewSet(types.AkLeft, types.AkTop, types.AkRight, types.AkBottom))
	m.content.SetTop(90)
	m.content.SetLeft(5)
	m.content.SetWidth(m.Width() - 10)
	m.content.SetHeight(m.Height() - (m.content.Top() + 5))

	newChromium := m.createChromium("")
	m.OnChromiumCreateTabSheet(newChromium)
	m.TForm.SetOnActivate(func(sender lcl.IObject) {
		newChromium.createBrowser(nil)
	})

	m.createTitleWidgetControl()

	m.SetOnResize(func(sender lcl.IObject) {
		m.recalculateTabSheet()
		m.updateWindowControlBtn()
	})
	m.SetOnShow(func(sender lcl.IObject) {
		m.SetFocus()
		m.SetActiveDefaultControl(m)
	})
}

// box 容器 窗口 拖拽 大小调整
func (m *BrowserWindow) boxDrag() {
	m.titleHeight = 45 // 标题栏高度
	m.borderWidth = 5  // 边框宽

	m.box.SetOnMouseMove(m.boxMouseMove)
	m.box.SetOnDblClick(func(sender lcl.IObject) {
		if m.isTitleBar {
			if m.WindowState() == types.WsNormal {
				m.SetWindowState(types.WsMaximized)
			} else {
				m.SetWindowState(types.WsNormal)
				if tool.IsDarwin() { //要这样重复设置2次不然不启作用
					m.SetWindowState(types.WsMaximized)
					m.SetWindowState(types.WsNormal)
				}
			}
		}
	})
	m.box.SetOnMouseDown(m.boxMouseDown)
	m.box.SetOnMouseUp(func(sender lcl.IObject, button types.TMouseButton, shift types.TShiftState, x, y int32) {
		m.isDown = false
	})
}

func (m *BrowserWindow) updateWindowControlBtn() {
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
	m.addChromBtn = wg.NewButton(m)
	m.addChromBtn.SetParent(m.box)
	addBtnRect := types.TRect{Left: 5, Top: 5}
	addBtnRect.SetSize(40, 40)
	m.addChromBtn.SetBoundsRect(addBtnRect)
	m.addChromBtn.SetStartColor(colors.RGBToColor(56, 57, 60))
	m.addChromBtn.SetEndColor(colors.RGBToColor(56, 57, 60))
	m.addChromBtn.SetRadius(5)
	m.addChromBtn.SetAlpha(255)
	m.addChromBtn.SetIcon(getResourcePath("add.png"))
	m.addChromBtn.SetOnClick(func(sender lcl.IObject) {
		m.addr.SetText("")
		newChromium := m.createChromium("")
		m.OnChromiumCreateTabSheet(newChromium)
		newChromium.createBrowser(nil)
	})
	// 窗口控制按钮 最小化，最大化，关闭
	{
		m.minBtn = wg.NewButton(m)
		m.minBtn.SetParent(m.box)
		m.minBtn.SetShowHint(true)
		m.minBtn.SetHint("最小化")
		minBtnRect := types.TRect{Left: m.box.Width() - 45*3, Top: 5}
		minBtnRect.SetSize(40, 40)
		m.minBtn.SetBoundsRect(minBtnRect)
		m.minBtn.SetStartColor(colors.RGBToColor(56, 57, 60))
		m.minBtn.SetEndColor(colors.RGBToColor(56, 57, 60))
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
		m.maxBtn.SetStartColor(colors.RGBToColor(56, 57, 60))
		m.maxBtn.SetEndColor(colors.RGBToColor(56, 57, 60))
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
		m.closeBtn.SetStartColor(colors.RGBToColor(56, 57, 60))
		m.closeBtn.SetEndColor(colors.RGBToColor(56, 57, 60))
		m.closeBtn.SetRadius(5)
		m.closeBtn.SetAlpha(255)
		m.closeBtn.SetIcon(getResourcePath("btn-close.png"))
		m.closeBtn.SetOnClick(func(sender lcl.IObject) {
			for _, chrom := range m.chroms {
				chrom.closeBrowser()
			}
			m.Close()
		})
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
		m.backBtn.SetStartColor(colors.RGBToColor(56, 57, 60))
		m.backBtn.SetEndColor(colors.RGBToColor(56, 57, 60))
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
		m.forwardBtn.SetStartColor(colors.RGBToColor(56, 57, 60))
		m.forwardBtn.SetEndColor(colors.RGBToColor(56, 57, 60))
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
		m.refreshBtn.SetStartColor(colors.RGBToColor(56, 57, 60))
		m.refreshBtn.SetEndColor(colors.RGBToColor(56, 57, 60))
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
	m.addr = lcl.NewMemo(m)
	m.addr.SetParent(m.box)
	m.addr.SetLeft(140)
	m.addr.SetTop(50)
	m.addr.SetHeight(33)
	m.addr.SetWidth(m.Width() - (m.addr.Left() + 50))
	//addr.SetBorderStyle(types.BsNone)
	m.addr.SetAnchors(types.NewSet(types.AkLeft, types.AkTop, types.AkRight))
	m.addr.Font().SetSize(14)
	m.addr.Font().SetColor(colors.ClWhite)
	m.addr.SetColor(colors.RGBToColor(56, 57, 60))
	// 阻止 memo 换行
	m.addr.SetOnKeyPress(func(sender lcl.IObject, key *uint16) {
		k := *key
		fmt.Println("addr.onkeypress", k)
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

	// 地址栏右边的 logo 按钮
	m.addrRightBtn = wg.NewButton(m)
	m.addrRightBtn.SetParent(m.box)
	m.addrRightBtn.SetShowHint(true)
	m.addrRightBtn.SetHint("   GO  \nENERGY")
	addrRightBtnRect := types.TRect{Left: m.addr.Left() + m.addr.Width() + 40/2, Top: 47}
	addrRightBtnRect.SetSize(40, 40)
	m.addrRightBtn.SetBoundsRect(addrRightBtnRect)
	m.addrRightBtn.SetStartColor(colors.RGBToColor(56, 57, 60))
	m.addrRightBtn.SetEndColor(colors.RGBToColor(56, 57, 60))
	m.addrRightBtn.SetRadius(35)
	m.addrRightBtn.SetAlpha(100)
	m.addrRightBtn.SetIcon(getResourcePath("addr-right-btn.png"))
	m.addrRightBtn.SetOnClick(func(sender lcl.IObject) {
		if chrom := m.getActiveChrom(); chrom != nil {
			chrom.chromium.LoadURLWithStringFrame("https://energye.github.io", chrom.chromium.Browser().GetMainFrame())
		}
	})
}

// 浏览器创建完添加一个 tab Sheet
func (m *BrowserWindow) OnChromiumCreateTabSheet(newChromium *Chromium) {
	m.chroms = append(m.chroms, newChromium)
	newChromium.windowId = int32(len(m.chroms))
	//fmt.Println("OnChromiumCreateTabSheet", "当前chromium数量:", len(m.chroms), "新chromiumID:", newChromium.windowId)
	m.AddTabSheet(newChromium)
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
	newTabSheet.IsScaled = true
	newTabSheet.ScaledWidth = 16
	newTabSheet.ScaledHeight = 16
	newTabSheet.SetOnCloseClick(func(sender lcl.IObject) {
		currentChromium.closeBrowser()
		// 删除当前chrom, 使用 windowId - 1 是当前 chrom 所在下标
		idx := currentChromium.windowId - 1
		// 删除
		m.chroms = append(m.chroms[:idx], m.chroms[idx+1:]...)
		// 重新设置每个 chromium 的 windowID, 在下次删除时能对应上
		for id, chrom := range m.chroms {
			chrom.windowId = int32(id + 1)
		}
		if len(m.chroms) > 0 {
			lastChrom := m.chroms[len(m.chroms)-1]
			lastChrom.updateTabSheetActive(true)
			m.updateTabSheetActive(lastChrom)
		} else {
			// 没有 chrom 清空和还原控制按钮、地址栏
			m.resetControlBtn()
		}
		// 重新计算 tab sheet left 和 width
		m.recalculateTabSheet()
	})
	newTabSheet.SetOnMouseLeave(func(sender lcl.IObject) {

	})
	newTabSheet.SetOnMouseDown(func(sender lcl.IObject, button types.TMouseButton, shift types.TShiftState, X int32, Y int32) {
		fmt.Println("TabSheet.OnMouseDown button:", button)
		//CW.Show()
	})
	newTabSheet.SetIconFavorite(getResourcePath("icon.png"))
	newTabSheet.SetIconClose(getResourcePath("sheet_close.png"))
	newTabSheet.SetOnClick(func(sender lcl.IObject) {
		m.updateTabSheetActive(currentChromium)
		currentChromium.updateTabSheetActive(true)
	})
	currentChromium.isActive = true           // 设置默认激活
	currentChromium.tabSheetBtn = newTabSheet // 绑定到当前 chromium

	// 更新其它tabSheet 非激活状态(颜色控制)
	m.updateTabSheetActive(currentChromium)

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
func (m *BrowserWindow) updateTabSheetActive(currentChromium *Chromium) {
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
	// 地址栏右侧按钮
	if m.addrRightBtn != nil {
		m.addrRightBtn.SetLeft(m.addr.Left() + m.addr.Width() + 5)
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
	}
	if tool.IsExist(sourcePath) {
		return sourcePath
	}
	return ""
}
