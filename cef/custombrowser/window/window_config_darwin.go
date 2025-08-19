package window

/*
#cgo CFLAGS: -mmacosx-version-min=11.0 -x objective-c
#cgo LDFLAGS: -mmacosx-version-min=11.0 -framework Cocoa
#include "window_config_darwin.h"

extern void onButtonClicked(char *identifier, char *value, void *userData);
extern void onTextChanged(char *identifier, char *value, void *userData);
*/
import "C"
import (
	"fmt"
	"github.com/energye/lcl/lcl"
	"log"
	"unsafe"
)

// Go包装函数
func ConfigureWindow(nsWindowHandle uintptr, config ToolbarConfiguration, callbackContext ToolbarCallbackContext) {
	C.ConfigureWindow(C.ulong(nsWindowHandle), C.ToolbarConfiguration(config), C.ToolbarCallbackContext{
		clickCallback:       callbackContext.ClickCallback,
		textChangedCallback: callbackContext.TextChangedCallback,
		userData:            callbackContext.UserData,
	})
}

func AddToolbarButton(nsWindowHandle uintptr, identifier, title, tooltip string, style ControlStyle) {
	cIdentifier := C.CString(identifier)
	defer C.free(unsafe.Pointer(cIdentifier))

	cTitle := C.CString(title)
	defer C.free(unsafe.Pointer(cTitle))

	var cTooltip *C.char
	if tooltip != "" {
		cTooltip = C.CString(tooltip)
		defer C.free(unsafe.Pointer(cTooltip))
	}

	cStyle := C.ControlStyle{
		width:       C.CGFloat(style.Width),
		height:      C.CGFloat(style.Height),
		bezelStyle:  C.NSBezelStyle(style.BezelStyle),
		controlSize: C.NSControlSize(style.ControlSize),
		font:        (*C.NSFont)(style.Font),
	}

	C.AddToolbarButton(C.ulong(nsWindowHandle), cIdentifier, cTitle, cTooltip, cStyle)
}

func AddToolbarImageButton(nsWindowHandle uintptr, identifier, imageName, tooltip string, style ControlStyle) {
	cIdentifier := C.CString(identifier)
	defer C.free(unsafe.Pointer(cIdentifier))

	cImageName := C.CString(imageName)
	defer C.free(unsafe.Pointer(cImageName))

	var cTooltip *C.char
	if tooltip != "" {
		cTooltip = C.CString(tooltip)
		defer C.free(unsafe.Pointer(cTooltip))
	}

	cStyle := C.ControlStyle{
		width:       C.CGFloat(style.Width),
		height:      C.CGFloat(style.Height),
		bezelStyle:  C.NSBezelStyle(style.BezelStyle),
		controlSize: C.NSControlSize(style.ControlSize),
		font:        (*C.NSFont)(style.Font),
	}

	C.AddToolbarImageButton(C.ulong(nsWindowHandle), cIdentifier, cImageName, cTooltip, cStyle)
}

func AddToolbarTextField(nsWindowHandle uintptr, identifier, placeholder string, style ControlStyle) {
	cIdentifier := C.CString(identifier)
	defer C.free(unsafe.Pointer(cIdentifier))

	var cPlaceholder *C.char
	if placeholder != "" {
		cPlaceholder = C.CString(placeholder)
		defer C.free(unsafe.Pointer(cPlaceholder))
	}

	cStyle := C.ControlStyle{
		width:       C.CGFloat(style.Width),
		height:      C.CGFloat(style.Height),
		bezelStyle:  C.NSBezelStyle(style.BezelStyle),
		controlSize: C.NSControlSize(style.ControlSize),
		font:        (*C.NSFont)(style.Font),
	}

	C.AddToolbarTextField(C.ulong(nsWindowHandle), cIdentifier, cPlaceholder, cStyle)
}

func AddToolbarSearchField(nsWindowHandle uintptr, identifier, placeholder string, style ControlStyle) {
	cIdentifier := C.CString(identifier)
	defer C.free(unsafe.Pointer(cIdentifier))

	var cPlaceholder *C.char
	if placeholder != "" {
		cPlaceholder = C.CString(placeholder)
		defer C.free(unsafe.Pointer(cPlaceholder))
	}

	cStyle := C.ControlStyle{
		width:       C.CGFloat(style.Width),
		height:      C.CGFloat(style.Height),
		bezelStyle:  C.NSBezelStyle(style.BezelStyle),
		controlSize: C.NSControlSize(style.ControlSize),
		font:        (*C.NSFont)(style.Font),
	}

	C.AddToolbarSearchField(C.ulong(nsWindowHandle), cIdentifier, cPlaceholder, cStyle)
}

func AddToolbarCombobox(nsWindowHandle uintptr, identifier string, items []string, style ControlStyle) {
	cIdentifier := C.CString(identifier)
	defer C.free(unsafe.Pointer(cIdentifier))

	// 转换Go字符串切片为C字符串数组
	cItems := make([]*C.char, len(items))
	for i, item := range items {
		cItems[i] = C.CString(item)
		defer C.free(unsafe.Pointer(cItems[i]))
	}

	cStyle := C.ControlStyle{
		width:       C.CGFloat(style.Width),
		height:      C.CGFloat(style.Height),
		bezelStyle:  C.NSBezelStyle(style.BezelStyle),
		controlSize: C.NSControlSize(style.ControlSize),
		font:        (*C.NSFont)(style.Font),
	}

	C.AddToolbarCombobox(C.ulong(nsWindowHandle), cIdentifier, (**C.char)(unsafe.Pointer(&cItems[0])), C.int(len(items)), cStyle)
}

func AddToolbarCustomView(nsWindowHandle uintptr, identifier string, style ControlStyle) {
	cIdentifier := C.CString(identifier)
	defer C.free(unsafe.Pointer(cIdentifier))

	cStyle := C.ControlStyle{
		width:       C.CGFloat(style.Width),
		height:      C.CGFloat(style.Height),
		bezelStyle:  C.NSBezelStyle(style.BezelStyle),
		controlSize: C.NSControlSize(style.ControlSize),
		font:        (*C.NSFont)(style.Font),
	}

	C.AddToolbarCustomView(C.ulong(nsWindowHandle), cIdentifier, cStyle)
}

func AddToolbarFlexibleSpace(nsWindowHandle uintptr) {
	C.AddToolbarFlexibleSpace(C.ulong(nsWindowHandle))
}

func AddToolbarSpace(nsWindowHandle uintptr) {
	C.AddToolbarSpace(C.ulong(nsWindowHandle))
}

func AddToolbarSpaceByWidth(nsWindowHandle uintptr, width float32) {
	C.AddToolbarSpaceByWidth(C.ulong(nsWindowHandle), C.CGFloat(width))
}

func RemoveToolbarItem(nsWindowHandle uintptr, identifier string) {
	cIdentifier := C.CString(identifier)
	defer C.free(unsafe.Pointer(cIdentifier))

	C.RemoveToolbarItem(C.ulong(nsWindowHandle), cIdentifier)
}

func GetToolbarControlValue(nsWindowHandle uintptr, identifier string) string {
	cIdentifier := C.CString(identifier)
	defer C.free(unsafe.Pointer(cIdentifier))

	cValue := C.GetToolbarControlValue(C.ulong(nsWindowHandle), cIdentifier)
	if cValue == nil {
		return ""
	}
	return C.GoString(cValue)
}

func SetToolbarControlValue(nsWindowHandle uintptr, identifier, value string) {
	cIdentifier := C.CString(identifier)
	defer C.free(unsafe.Pointer(cIdentifier))

	cValue := C.CString(value)
	defer C.free(unsafe.Pointer(cValue))

	C.SetToolbarControlValue(C.ulong(nsWindowHandle), cIdentifier, cValue)
}

func SetToolbarControlEnabled(nsWindowHandle uintptr, identifier string, enabled bool) {
	cIdentifier := C.CString(identifier)
	defer C.free(unsafe.Pointer(cIdentifier))

	C.SetToolbarControlEnabled(C.ulong(nsWindowHandle), cIdentifier, C.bool(enabled))
}

func CreateDefaultControlStyle() ControlStyle {
	cStyle := C.CreateDefaultControlStyle()
	return ControlStyle{
		Width:       float64(cStyle.width),
		Height:      float64(cStyle.height),
		BezelStyle:  NSBezelStyle(cStyle.bezelStyle),
		ControlSize: NSControlSize(cStyle.controlSize),
		Font:        unsafe.Pointer(cStyle.font),
	}
}

func CreateControlStyle(width, height float64, bezelStyle NSBezelStyle, controlSize NSControlSize, font unsafe.Pointer) ControlStyle {
	cStyle := C.CreateControlStyle(
		C.CGFloat(width),
		C.CGFloat(height),
		C.NSBezelStyle(bezelStyle),
		C.NSControlSize(controlSize),
		font,
	)
	return ControlStyle{
		Width:       float64(cStyle.width),
		Height:      float64(cStyle.height),
		BezelStyle:  NSBezelStyle(cStyle.bezelStyle),
		ControlSize: NSControlSize(cStyle.controlSize),
		Font:        unsafe.Pointer(cStyle.font),
	}
}

// 导出Go回调函数供C调用

//export onButtonClicked
func onButtonClicked(identifier *C.char, value *C.char, userData unsafe.Pointer) {
	id := C.GoString(identifier)
	val := C.GoString(value)
	fmt.Println("clicked id:", id, "val:", val, "userData:", uintptr(userData))
}

//export onTextChanged
func onTextChanged(identifier *C.char, value *C.char, userData unsafe.Pointer) {
	id := C.GoString(identifier)
	val := C.GoString(value)
	fmt.Println("clicked id:", id, "val:", val, "userData:", uintptr(userData))
}

// //export releaseFont
//
//	func releaseFont(font *C.NSFont) {
//		C.CFRelease(C.CFTypeRef(font))
//	}
//
// 定义工具栏位置枚举
type ToolbarPosition int

const (
	ToolbarLeft   ToolbarPosition = iota // 左侧
	ToolbarCenter                        // 中间
	ToolbarRight                         // 右侧
)

// 计算插入索引
func calculateInsertIndex(nsWindowHandle uintptr, position ToolbarPosition) int {
	// 获取当前工具栏项数（需要新增一个Objective-C函数获取项数）
	itemCount := int(C.GetToolbarItemCount(C.ulong(nsWindowHandle)))

	switch position {
	case ToolbarLeft:
		return 0 // 左侧：插入到最前面
	case ToolbarCenter:
		// 中间：插入到现有项数的一半位置（向下取整）
		return itemCount / 2
	case ToolbarRight:
		return itemCount // 右侧：插入到末尾
	default:
		return itemCount // 默认右侧
	}
}

func (m *Window) TestTool() {
	// 获取窗口句柄
	windowHandle := uintptr(lcl.PlatformWindow(m.Instance()))
	if windowHandle == 0 {
		log.Fatal("Failed to get window handle")
	}
	fmt.Println("windowHandle:", windowHandle)

	// 创建回调上下文
	callbackContext := ToolbarCallbackContext{
		ClickCallback:       (C.ControlCallback)(C.onButtonClicked),
		TextChangedCallback: (C.ControlCallback)(C.onTextChanged),
		UserData:            unsafe.Pointer(windowHandle),
	}

	// 配置窗口工具栏
	config := ToolbarConfigurationAllowUserCustomization |
		//ToolbarConfigurationAutoSaveConfiguration |
		ToolbarConfigurationDisplayModeIconAndText

	ConfigureWindow(windowHandle, config, callbackContext)

	// 创建默认样式
	defaultStyle := CreateDefaultControlStyle()
	defaultStyle.Height = 24
	defaultStyle.BezelStyle = BezelStyleTexturedRounded // 边框样式
	defaultStyle.ControlSize = ControlSizeLarge         // 控件大小

	itemCount := int(C.GetToolbarItemCount(C.ulong(windowHandle)))
	fmt.Println("当前控件总数：", itemCount)
	// 添加按钮
	AddToolbarButton(windowHandle, "run-button", "Run", "Run the program", defaultStyle)
	AddToolbarFlexibleSpace(windowHandle)

	// 添加图片按钮
	AddToolbarImageButton(windowHandle, "settings-button", "NSPreferencesGeneral", "Open settings", defaultStyle)
	AddToolbarFlexibleSpace(windowHandle)

	// 添加文本框
	textFieldStyle := defaultStyle
	textFieldStyle.Width = 400
	textFieldStyle.Height = 24
	AddToolbarTextField(windowHandle, "search-field", "Search...", textFieldStyle)
	AddToolbarSpace(windowHandle)

	// 添加下拉框
	comboItems := []string{"Option 1", "Option 2", "Option 3"}
	AddToolbarCombobox(windowHandle, "options-combo", comboItems, defaultStyle)

	fmt.Println("Toolbar created successfully!")

	// 模拟设置控件值
	SetToolbarControlValue(windowHandle, "search-field", "Initial value")

	// 模拟获取控件值
	value := GetToolbarControlValue(windowHandle, "search-field")
	fmt.Printf("Search field value: %s\n", value)
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
