package window

import (
	"fmt"
	"github.com/energye/examples/cef/custombrowser/window/cocoa"
	"github.com/energye/examples/cef/custombrowser/window/cocoa/toolbar"
	"github.com/energye/lcl/lcl"
	"log"
	"time"
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

	// 创建回调上下文
	//callbackContext := ToolbarCallbackContext{
	//	ClickCallback:       (C.ControlCallback)(C.onButtonClicked),
	//	TextChangedCallback: (C.ControlCallback)(C.onTextChanged),
	//	TextSubmitCallback:  (C.ControlCallback)(C.onTextSubmit),
	//	UserData:            unsafe.Pointer(windowHandle),
	//}

	// 配置窗口工具栏
	config := toolbar.ToolbarConfiguration{
		DisplayMode: toolbar.NSToolbarDisplayModeIconOnly,
		Transparent: true,
		SizeMode:    toolbar.NSToolbarSizeModeSmall,
		//Style:                     NSWindowToolbarStyleUnifiedCompact,
		Style:                     toolbar.NSWindowToolbarStyleUnified,
		IsAllowsUserCustomization: true,
	}
	bar := toolbar.Create(m, config)

	// 创建默认样式
	defaultProperty := toolbar.CreateDefaultControlProperty()
	//defaultProperty.Height = 24
	//defaultProperty.BezelStyle = BezelStyleTexturedRounded // 边框样式
	//defaultProperty.ControlSize = ControlSizeLarge         // 控件大小
	defaultProperty.IsNavigational = true

	item := toolbar.ButtonItem{}
	//bar.AddButton(item, defaultProperty)
	fmt.Println("当前控件总数：", toolbar.GetToolbarItemCount(windowHandle))
	//
	btn1 := bar.NewButton(item, defaultProperty)
	bar.AddControl(btn1)
	btn1.SetOnClick(func(identifier string, owner toolbar.Pointer, sender toolbar.Pointer) *toolbar.GoData {
		fmt.Println("自定义新按钮事件触发了", identifier)
		return nil
	})
	// 添加按钮
	toolbar.AddToolbarButton(windowHandle, "back", "后退", "后退", defaultProperty)
	toolbar.AddToolbarButton(windowHandle, "forwd", "前进", "前进", defaultProperty)
	toolbar.AddToolbarButton(windowHandle, "refs", "刷新", "刷新", defaultProperty)
	//AddToolbarFlexibleSpace(windowHandle)

	// 添加文本框
	textProperty := defaultProperty
	//textProperty.Height = 28
	//textProperty.IsNavigational = true
	textProperty.IsCenteredItem = true
	textProperty.VisibilityPriority = toolbar.NSToolbarItemVisibilityPriorityHigh
	//AddToolbarTextField(windowHandle, "text-field", "text...", textProperty)

	// 添加搜索框
	toolbar.AddToolbarFlexibleSpace(windowHandle)
	//textProperty.MinWidth = 60
	//textProperty.MaxWidth = float64(m.Width() - 250)
	//textProperty.Width = float64(m.Width() - 250)
	sf := toolbar.AddToolbarSearchField(windowHandle, "search-field", "Search...", textProperty)
	println(sf, "textProperty.MaxWidth", textProperty.MaxWidth)
	toolbar.AddToolbarFlexibleSpace(windowHandle)

	// 添加下拉框
	comboProperty := defaultProperty
	comboProperty.IsNavigational = false
	comboProperty.Width = 100
	//comboItems := []string{"Option 1", "Option 2", "Option 3"}
	//AddToolbarCombobox(windowHandle, "options-combo", comboItems, comboProperty)

	// 添加图片按钮
	imageButtonProperty := defaultProperty
	imageButtonProperty.IsNavigational = false
	imageButtonProperty.VisibilityPriority = toolbar.NSToolbarItemVisibilityPriorityHigh
	toolbar.AddToolbarImageButton(windowHandle, "go-back", "arrow.left", "Open settings", imageButtonProperty)
	fmt.Println("当前控件总数：", toolbar.GetToolbarItemCount(windowHandle))
	go func() {
		time.Sleep(time.Second * 2)
		cocoa.RunOnMainThread(func() {
			//SetToolbarControlHidden(windowHandle, "go-back", true)
			//SetToolbarControlValue(windowHandle, "search-field", "Object-c UI线程 设置 Initial value")
			sf.SetText("Object-c UI线程 设置 Initial value")
			fmt.Println("sf.GetText():", sf.GetText())
			//toolbar.NewLCLButton(bar, testbtn)
			//toolbar.SetWindowBackgroundColor(m, toolbar.Color{Red: 56, Green: 57, Blue: 60, Alpha: 255})
		})
		time.Sleep(time.Second * 2)
		lcl.RunOnMainThreadAsync(func(id uint32) {
			//SetToolbarControlValue(windowHandle, "search-field", "lcl.RunOnMainThreadAsync 设置 Initial value")
			sf.SetText("lcl.RunOnMainThreadAsync 设置 Initial value")
			fmt.Println("sf.GetText():", sf.GetText())
		})
		time.Sleep(time.Second * 2)
		lcl.RunOnMainThreadSync(func() {
			//SetToolbarControlValue(windowHandle, "search-field", "lcl.RunOnMainThreadSync 设置 Initial value")
			sf.SetText("lcl.RunOnMainThreadSync 设置 Initial value")
			fmt.Println("sf.GetText():", sf.GetText())
		})
	}()

	fmt.Println("Toolbar created successfully!")

	// 模拟设置控件值
	toolbar.SetToolbarControlValue(windowHandle, "search-field", "Initial value")

	// 模拟获取控件值
	value := toolbar.GetToolbarControlValue(windowHandle, "search-field")
	fmt.Printf("Search field value: %s\n", value)

	bar.SetOnWindowResize(func(identifier string, owner toolbar.Pointer, sender toolbar.Pointer) *toolbar.GoData {
		width := int(m.Width() - 500)
		if width > 700 {
			width = 700
		}
		//fmt.Println("width", width)
		sf.UpdateSearchFieldWidth(width)
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
