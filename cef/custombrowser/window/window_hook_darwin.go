package window

import (
	"fmt"
	"github.com/energye/examples/cef/custombrowser/window/cocoa"
	"github.com/energye/examples/cef/custombrowser/window/cocoa/toolbar"
	"github.com/energye/lcl/lcl"
	"log"
	"os"
)

var Resize func() = nil

func (m *Window) TestTool() {
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

//现代 macOS 工具栏开发最佳实践总结
//
//理解“统一工具栏”：从 macOS 11 (Big Sur) 开始，工具栏和标题栏在视觉上融合。使用 isNavigational 和 allowedAligned 属性来正确放置你的项。
//明确项的角色：
//导航类 (isNavigational = true)：如前进、后退、侧边栏切换。靠左放置。
//主要操作/搜索 (principalItem)：如搜索栏。居中放置。
//内容相关操作 (allowedAligned = .trailing)：如分享、排序、查看选项。靠右放置。
//灵活空间 (.flexibleSpace, .space)：用于布局和对齐。
//优先使用 SF Symbols：确保图标在不同主题和状态下的一致性。
//善用分组：对于相关的操作（如视图切换：列表、图标、分栏），使用 NSToolbarItemGroup 并以 collapsed 模式显示，以节省空间。
//响应式显示：正确设置 visibilityPriority，确保在窗口变窄时，最重要的项仍然可见，不重要的项会被自动隐藏到溢出菜单中。
