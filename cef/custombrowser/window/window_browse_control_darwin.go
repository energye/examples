//go:build darwin
// +build darwin

package window

import (
	"fmt"
	"github.com/energye/examples/cef/custombrowser/window/cocoa"
	"github.com/energye/examples/cef/custombrowser/window/cocoa/toolbar"
	"github.com/energye/lcl/api"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types/colors"
	"log"
	"os"
	"strings"
)

const isDarwin = true

func (m *BrowserWindow) createTitleWidgetControl() {

}

var (
	bar        *toolbar.NSToolBar
	backBtn    *toolbar.NSImageButton
	forwardBtn *toolbar.NSImageButton
	refreshBtn *toolbar.NSImageButton
	addr       *toolbar.NSSearchField
)

func init() {
	if isDarwin {
		tabSheetBtnHeight = 30
		tabSheetBtnRightSize = 5
	}
}

func (m *BrowserWindow) macOSToolbar() {
	cocoa.RegisterRunOnMainThreadCallback()
	// 获取窗口句柄
	windowHandle := uintptr(lcl.PlatformWindow(m.Instance()))
	if windowHandle == 0 {
		log.Fatal("Failed to get window handle")
	}
	fmt.Println("windowHandle:", windowHandle)

	// 配置窗口工具栏
	toolbarConfig := toolbar.ToolbarConfiguration{
		DisplayMode: toolbar.NSToolbarDisplayModeIconOnly,
		Transparent: true,
		SizeMode:    toolbar.NSToolbarSizeModeSmall,
		//Style:                     NSWindowToolbarStyleUnifiedCompact,
		Style:                     toolbar.NSWindowToolbarStyleUnified,
		IsAllowsUserCustomization: true,
	}
	bar = toolbar.Create(m, toolbarConfig)
	// 添加按钮
	fmt.Println("当前控件总数:", bar.ItemCount())

	//viewConfig := toolbar.ItemBase{}
	//view := bar.NewView(viewConfig)
	//println("view:", view.Identifier())
	//bar.AddItem(view)

	// 创建默认样式
	defaultProperty := toolbar.CreateDefaultControlProperty()
	//defaultProperty.Height = 24
	//defaultProperty.BezelStyle = BezelStyleTexturedRounded // 边框样式
	//defaultProperty.ControlSize = ControlSizeLarge         // 控件大小
	defaultProperty.IsNavigational = true
	defaultProperty.VisibilityPriority = toolbar.NSToolbarItemVisibilityPriorityHigh

	// 添加按钮
	item := toolbar.ButtonItem{}

	backBtnProperty := defaultProperty
	backBtnConfig := item
	backBtnConfig.Tips = "后退"
	backBtnConfig.IconName = getResourcePath("back.png")
	backBtn = bar.NewImageButtonForImage(backBtnConfig, backBtnProperty)
	backBtn.SetOnClick(func(identifier string, owner toolbar.Pointer, sender toolbar.Pointer) *toolbar.GoArguments {
		fmt.Println("backBtn OnClick", identifier)
		chrom := m.getActiveChrom()
		if chrom != nil && chrom.chromium.CanGoBack() {
			chrom.chromium.GoBack()
		}
		return nil
	})
	bar.AddItem(backBtn)

	forwardBtnProperty := defaultProperty
	forwardBtnConfig := item
	forwardBtnConfig.Tips = "前进"
	forwardBtnConfig.IconName = getResourcePath("forward.png")
	forwardBtn = bar.NewImageButtonForImage(forwardBtnConfig, forwardBtnProperty)
	forwardBtn.SetOnClick(func(identifier string, owner toolbar.Pointer, sender toolbar.Pointer) *toolbar.GoArguments {
		fmt.Println("forwardBtn OnClick", identifier)
		chrom := m.getActiveChrom()
		if chrom != nil && chrom.chromium.CanGoForward() {
			chrom.chromium.GoForward()
		}
		return nil
	})
	bar.AddItem(forwardBtn)

	refreshBtnProperty := defaultProperty
	refreshBtnConfig := item
	refreshBtnConfig.Tips = "刷新"
	refreshBtnConfig.IconName = getResourcePath("refresh.png")
	refreshBtn = bar.NewImageButtonForImage(refreshBtnConfig, refreshBtnProperty)
	refreshBtn.SetOnClick(func(identifier string, owner toolbar.Pointer, sender toolbar.Pointer) *toolbar.GoArguments {
		fmt.Println("refreshBtn OnClick", identifier)
		chrom := m.getActiveChrom()
		if chrom != nil {
			if chrom.isLoading {
				chrom.chromium.StopLoad()
			} else {
				chrom.chromium.Reload()
			}
		}
		return nil
	})
	bar.AddItem(refreshBtn)

	// 添加搜索框
	bar.AddFlexibleSpace()
	textProperty := defaultProperty
	textProperty.IsCenteredItem = true
	textProperty.VisibilityPriority = toolbar.NSToolbarItemVisibilityPriorityHigh
	textItem := toolbar.ControlTextField{}
	textItem.Placeholder = "输入网站地址"
	textItem.Identifier = "SiteAddrSearch"
	//textProperty.MinWidth = 60
	//textProperty.MaxWidth = float64(m.Width() - 250)
	//textProperty.Width = float64(m.Width() - 250)
	addr = bar.NewSearchField(textItem, textProperty)
	addr.SetOnCommit(func(identifier string, value string, owner toolbar.Pointer, sender toolbar.Pointer) *toolbar.GoArguments {
		println("addr OnCommit", identifier, value)
		tempUrl := strings.TrimSpace(value)
		if tempUrl == "" {
			return nil
		}
		for _, chrom := range m.chroms {
			if chrom.isActive {
				chrom.chromium.LoadURLWithStringFrame(tempUrl, chrom.chromium.Browser().GetMainFrame())
			}
		}
		return nil
	})
	bar.AddItem(addr)
	bar.AddFlexibleSpace()

	// 添加图片按钮
	addBtnProperty := defaultProperty
	addBtnProperty.IsNavigational = false
	addBtnProperty.VisibilityPriority = toolbar.NSToolbarItemVisibilityPriorityHigh
	addBtnConfig := item
	addBtnConfig.Tips = "新建标签页"
	addBtnConfig.IconName = getResourcePath("add.png")
	addBtnData, _ := os.ReadFile(addBtnConfig.IconName)
	//addBtn := bar.NewImageButtonForImage(addBtnConfig, addBtnProperty)
	addBtn := bar.NewImageButtonForBytes(addBtnData, addBtnConfig, addBtnProperty)
	addBtn.SetOnClick(func(identifier string, owner toolbar.Pointer, sender toolbar.Pointer) *toolbar.GoArguments {
		fmt.Println("addBtn OnClick", identifier, "isMainThread:", api.MainThreadId() == api.CurrentThreadId())
		m.SetAddrText("")
		newChromium := m.createChromium("")
		m.OnChromiumCreateTabSheet(newChromium)
		newChromium.createBrowser(nil)
		return nil
	})
	bar.AddItem(addBtn)

	// 添加图片按钮
	rightBtnProperty := defaultProperty
	rightBtnProperty.IsNavigational = false
	rightBtnProperty.VisibilityPriority = toolbar.NSToolbarItemVisibilityPriorityHigh
	rightBtnConfig := item
	rightBtnConfig.IconName = getResourcePath("addr-right-btn.png")
	rightBtnData, _ := os.ReadFile(rightBtnConfig.IconName)
	rightBtn := bar.NewImageButtonForBytes(rightBtnData, rightBtnConfig, rightBtnProperty)
	rightBtn.SetOnClick(func(identifier string, owner toolbar.Pointer, sender toolbar.Pointer) *toolbar.GoArguments {
		fmt.Println("rightBtn OnClick", identifier, "isMainThread:", api.MainThreadId() == api.CurrentThreadId())
		if chrom := m.getActiveChrom(); chrom != nil {
			chrom.chromium.LoadURLWithStringFrame("https://energye.github.io", chrom.chromium.Browser().GetMainFrame())
		}
		return nil
	})
	bar.AddItem(rightBtn)

	fmt.Println("当前控件总数:", bar.ItemCount())

	bar.SetOnWindowResize(func(identifier string, owner toolbar.Pointer, sender toolbar.Pointer) *toolbar.GoArguments {
		width := int(m.Width() - 460)
		if width > 700 {
			width = 700
		}
		//fmt.Println("width", width)
		addr.UpdateTextFieldWidth(width)
		return nil
	})
}

func (m *BrowserWindow) SetAddrText(val string) {
	lcl.RunOnMainThreadAsync(func(id uint32) {
		addr.SetText(val)
	})
}

// 清空地址栏 和 还原控制按钮
func (m *BrowserWindow) resetControlBtn() {
	addr.SetText("")
	backBtn.SetEnable(false)
	forwardBtn.SetEnable(false)
	refreshBtn.SetImageFromPath(getResourcePath("refresh.png"))
}

func (m *BrowserWindow) updateRefreshBtn(chromium *Chromium, isLoading bool) {
	if isLoading {
		lcl.RunOnMainThreadAsync(func(id uint32) {
			refreshBtn.SetImageFromPath(getResourcePath("stop.png"))
		})
	} else {
		lcl.RunOnMainThreadAsync(func(id uint32) {
			refreshBtn.SetImageFromPath(getResourcePath("refresh.png"))
		})
	}
}

// 根据当前 chromium 浏览器加载状态更新浏览器控制按钮
func (m *Chromium) updateBrowserControlBtn() {
	backBtn.SetEnable(m.canGoBack)
	forwardBtn.SetEnable(m.canGoForward)
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
	m.updateBrowserControlBtn()
}
