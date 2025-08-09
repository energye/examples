package window

import (
	"fmt"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
	"github.com/energye/lcl/types/colors"
	"net/url"
	"path/filepath"
	"strings"
	"widget/wg"
)

type BrowserWindow struct {
	lcl.TEngForm
	box          lcl.IPanel
	content      lcl.IPanel
	mainWindowId int32 // 窗口ID
	canClose     bool
	oldWndPrc    uintptr
	windowId     int
	chroms       map[int32]*Chromium
	addBtn       *wg.TButton
	closeBtn     *wg.TButton
	backBtn      *wg.TButton
	addr         lcl.IMemo
}

var (
	BW BrowserWindow
)

func (m *BrowserWindow) FormCreate(sender lcl.IObject) {
	m.SetWidth(1024)
	m.SetHeight(768)
	m.SetDoubleBuffered(true)
	//m.SetColor(colors.ClYellow)
	m.SetColor(colors.RGBToColor(56, 57, 60))
	m.ScreenCenter()
	m.SetCaption("ENERGY 3.0 - 自定义浏览器")
	constraints := m.Constraints()
	constraints.SetMinWidth(800)
	constraints.SetMinHeight(600)
	m.chroms = make(map[int32]*Chromium)

	m.box = lcl.NewPanel(m)
	m.box.SetParent(m)
	m.box.SetBevelOuter(types.BvNone)
	m.box.SetDoubleBuffered(true)
	m.box.SetWidth(m.Width())
	m.box.SetHeight(m.Height())
	m.box.SetAnchors(types.NewSet(types.AkLeft, types.AkTop, types.AkRight, types.AkBottom))

	m.content = lcl.NewPanel(m)
	m.content.SetParent(m.box)
	m.content.SetBevelOuter(types.BvNone)
	m.content.SetDoubleBuffered(true)
	m.content.SetAnchors(types.NewSet(types.AkLeft, types.AkTop, types.AkRight, types.AkBottom))
	m.content.SetTop(90)
	m.content.SetLeft(5)
	m.content.SetWidth(m.Width() - 10)
	m.content.SetHeight(m.Height() - (m.content.Top() + 5))

	newChromium := m.createChromium("https://www.baidu.com")
	newChromium.SetOnAfterCreated(m.OnChromiumAfterCreated)
	m.TForm.SetOnActivate(func(sender lcl.IObject) {
		newChromium.createBrowser(nil)
	})

	m.createTitleWidgetControl()

}

func (m *BrowserWindow) createTitleWidgetControl() {
	m.addBtn = wg.NewButton(m)
	m.addBtn.SetParent(m.box)
	addBtnRect := types.TRect{Left: 5, Top: 5}
	addBtnRect.SetSize(40, 40)
	m.addBtn.SetBoundsRect(addBtnRect)
	m.addBtn.SetStartColor(colors.RGBToColor(56, 57, 60))
	m.addBtn.SetEndColor(colors.RGBToColor(56, 57, 60))
	m.addBtn.SetRadius(5)
	m.addBtn.SetAlpha(255)
	m.addBtn.SetIcon(getImageResourcePath("add.png"))
	m.addBtn.SetOnClick(func(sender lcl.IObject) {
		m.addr.SetText("")
		newChromium := m.createChromium("")
		newChromium.SetOnAfterCreated(m.OnChromiumAfterCreated)
		newChromium.createBrowser(nil)
	})

	m.closeBtn = wg.NewButton(m)
	m.closeBtn.SetParent(m.box)
	closeBtnRect := types.TRect{Left: 5, Top: 45}
	closeBtnRect.SetSize(40, 40)
	m.closeBtn.SetBoundsRect(closeBtnRect)
	m.closeBtn.SetStartColor(colors.RGBToColor(56, 57, 60))
	m.closeBtn.SetEndColor(colors.RGBToColor(56, 57, 60))
	m.closeBtn.SetRadius(5)
	m.closeBtn.SetAlpha(255)
	m.closeBtn.SetIcon(getImageResourcePath("stop.png"))

	m.backBtn = wg.NewButton(m)
	m.backBtn.SetParent(m.box)
	backBtnRect := types.TRect{Left: 50, Top: 45}
	backBtnRect.SetSize(40, 40)
	m.backBtn.SetBoundsRect(backBtnRect)
	m.backBtn.SetStartColor(colors.RGBToColor(56, 57, 60))
	m.backBtn.SetEndColor(colors.RGBToColor(56, 57, 60))
	m.backBtn.SetRadius(5)
	m.backBtn.SetAlpha(255)
	m.backBtn.SetIcon(getImageResourcePath("stop.png"))

	m.addr = lcl.NewMemo(m)
	m.addr.SetParent(m.box)
	m.addr.SetLeft(120)
	m.addr.SetTop(50)
	m.addr.SetHeight(33)
	m.addr.SetWidth(m.Width() - (m.addr.Left() + 5))
	//addr.SetBorderStyle(types.BsNone)
	m.addr.SetAnchors(types.NewSet(types.AkLeft, types.AkTop, types.AkRight))
	m.addr.Font().SetSize(14)
	m.addr.Font().SetColor(colors.ClWhite)
	m.addr.SetColor(colors.RGBToColor(56, 57, 60))
	// 阻止 memo 换行
	m.addr.SetOnKeyPress(func(sender lcl.IObject, key *uint16) {
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
					break
				}
			}
		}
	})
	m.addr.SetOnChange(func(sender lcl.IObject) {
		text := m.addr.Text()
		text = strings.ReplaceAll(text, "\r", "")
		text = strings.ReplaceAll(text, "\n", "")
		m.addr.SetText(text)
	})
}

// 浏览器创建完添加一个 tab Sheet
func (m *BrowserWindow) OnChromiumAfterCreated(newChromium *Chromium) {
	m.chroms[newChromium.windowId] = newChromium
	fmt.Println("OnChromiumAfterCreated", "当前chromium数量:", len(m.chroms), "新chromiumID:", newChromium.windowId)
	m.AddTabSheet(newChromium)
}

func (m *BrowserWindow) AddTabSheet(newChromium *Chromium) {
	// 当前的设置为激活状态（颜色控制）
	var leftSize int32 = 5
	for _, chrom := range m.chroms {
		if chrom.tabSheet != nil {
			leftSize += chrom.tabSheet.Width() + 5
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
	newTabSheetRect.SetSize(230, 40)
	newTabSheet.SetBoundsRect(newTabSheetRect)
	newTabSheet.SetStartColor(colors.RGBToColor(86, 88, 93))
	newTabSheet.SetEndColor(colors.RGBToColor(86, 88, 93))
	newTabSheet.RoundedCorner = newTabSheet.RoundedCorner.Exclude(wg.RcLeftBottom).Exclude(wg.RcRightBottom)
	newTabSheet.SetOnCloseClick(func(sender lcl.IObject) {
		fmt.Println("点击了 X")
	})
	newTabSheet.SetIconFavorite(getImageResourcePath("icon.png"))
	newTabSheet.SetIconClose(getImageResourcePath("sheet_close.png"))
	newTabSheet.SetOnClick(func(sender lcl.IObject) {
		m.updateTabSheetActive(newChromium)
		newChromium.updateTabSheetActive(true)
	})
	newChromium.isActive = true
	newChromium.tabSheet = newTabSheet

	// 更新添加按钮位置
	m.updateAddBtnLeft()

	// 更新其它tabSheet 非激活状态(颜色控制)
	m.updateTabSheetActive(newChromium)
}

// 更新其它 tab sheet 状态
func (m *BrowserWindow) updateTabSheetActive(currentChromium *Chromium) {
	for _, chrom := range m.chroms {
		if chrom != currentChromium {
			chrom.updateTabSheetActive(false)
		}
	}
}

func (m *BrowserWindow) updateAddBtnLeft() {
	var leftSize int32 = 5
	for _, chrom := range m.chroms {
		if chrom.tabSheet != nil {
			leftSize += chrom.tabSheet.Width() + 5
		}
	}
	// 更新 添加按钮位置
	m.addBtn.SetLeft(leftSize)
}

func (m *BrowserWindow) FormAfterCreate(sender lcl.IObject) {
	//m.HookWndProcMessage()
}

func getImageResourcePath(imageName string) string {
	return filepath.Join("E:\\SWT\\gopath\\src\\github.com\\energye\\workspace\\examples\\cef\\custombrowser\\resources", imageName)
}
