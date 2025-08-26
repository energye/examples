package window

import (
	"fmt"
	"github.com/energye/examples/cef/custombrowser/window/cocoa"
	"github.com/energye/examples/cef/custombrowser/window/cocoa/toolbar"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
	"log"
	"os"
)

func (m *BrowserWindow) Minimize() {
}

func (m *BrowserWindow) Maximize() {
}

func (m *BrowserWindow) FullScreen() {
}

func (m *BrowserWindow) ExitFullScreen() {
}

func (m *BrowserWindow) IsFullScreen() bool {
	return false
}

func (m *BrowserWindow) boxDblClick(sender lcl.IObject) {
}

func (m *BrowserWindow) boxMouseMove(sender lcl.IObject, shift types.TShiftState, x, y int32) {
}

func (m *BrowserWindow) boxMouseDown(sender lcl.IObject, button types.TMouseButton, shift types.TShiftState, x, y int32) {
}

func (m *BrowserWindow) toolbar() {
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
	bar := toolbar.Create(m, toolbarConfig)
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
	backBtnConfig.IconName = "/Users/yanghy/app/workspace/examples/cef/custombrowser/resources/back.png"
	backBtn := bar.NewImageButtonForImage(backBtnConfig, backBtnProperty)
	backBtn.SetOnClick(func(identifier string, owner toolbar.Pointer, sender toolbar.Pointer) *toolbar.GoArguments {
		fmt.Println("OnClick", identifier)
		return nil
	})
	bar.AddItem(backBtn)

	forwardBtnProperty := defaultProperty
	forwardBtnConfig := item
	forwardBtnConfig.Tips = "前进"
	forwardBtnConfig.IconName = "/Users/yanghy/app/workspace/examples/cef/custombrowser/resources/forward.png"
	forwardBtn := bar.NewImageButtonForImage(forwardBtnConfig, forwardBtnProperty)
	forwardBtn.SetOnClick(func(identifier string, owner toolbar.Pointer, sender toolbar.Pointer) *toolbar.GoArguments {
		fmt.Println("OnClick", identifier)
		backBtn.SetEnable(!backBtn.Enable())
		//backBtn.SetHidden(!backBtn.Hidden())
		return nil
	})
	bar.AddItem(forwardBtn)

	refreshBtnProperty := defaultProperty
	refreshBtnConfig := item
	refreshBtnConfig.Tips = "刷新"
	refreshBtnConfig.IconName = "/Users/yanghy/app/workspace/examples/cef/custombrowser/resources/refresh.png"
	refreshBtn := bar.NewImageButtonForImage(refreshBtnConfig, refreshBtnProperty)
	refreshBtn.SetOnClick(func(identifier string, owner toolbar.Pointer, sender toolbar.Pointer) *toolbar.GoArguments {
		fmt.Println("OnClick", identifier)
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
	search := bar.NewSearchField(textItem, textProperty)
	search.SetOnCommit(func(identifier string, value string, owner toolbar.Pointer, sender toolbar.Pointer) *toolbar.GoArguments {
		println("OnCommit", identifier, value)
		return nil
	})
	bar.AddItem(search)
	bar.AddFlexibleSpace()

	// 添加图片按钮
	addBtnProperty := defaultProperty
	addBtnProperty.IsNavigational = false
	addBtnProperty.VisibilityPriority = toolbar.NSToolbarItemVisibilityPriorityHigh
	addBtnConfig := item
	addBtnConfig.IconName = "/Users/yanghy/app/workspace/examples/cef/custombrowser/resources/add.png"
	addBtnData, _ := os.ReadFile(addBtnConfig.IconName)
	//addBtn := bar.NewImageButtonForImage(addBtnConfig, addBtnProperty)
	addBtn := bar.NewImageButtonForBytes(addBtnData, addBtnConfig, addBtnProperty)
	addBtn.SetOnClick(func(identifier string, owner toolbar.Pointer, sender toolbar.Pointer) *toolbar.GoArguments {
		fmt.Println("OnClick", identifier)
		return nil
	})
	bar.AddItem(addBtn)

	// 添加图片按钮
	rightBtnProperty := defaultProperty
	rightBtnProperty.IsNavigational = false
	rightBtnProperty.VisibilityPriority = toolbar.NSToolbarItemVisibilityPriorityHigh
	rightBtnConfig := item
	rightBtnConfig.IconName = "/Users/yanghy/app/workspace/examples/cef/custombrowser/resources/addr-right-btn.png"
	rightBtnData, _ := os.ReadFile(rightBtnConfig.IconName)
	rightBtn := bar.NewImageButtonForBytes(rightBtnData, rightBtnConfig, rightBtnProperty)
	rightBtn.SetOnClick(func(identifier string, owner toolbar.Pointer, sender toolbar.Pointer) *toolbar.GoArguments {
		fmt.Println("OnClick", identifier)
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
		search.UpdateTextFieldWidth(width)
		return nil
	})
}
