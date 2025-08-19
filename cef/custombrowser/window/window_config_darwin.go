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

// 定义Go对应的类型和常量
type ToolbarConfiguration C.ToolbarConfiguration

const (
	ToolbarConfigurationNone                   ToolbarConfiguration = C.ToolbarConfigurationNone
	ToolbarConfigurationAllowUserCustomization ToolbarConfiguration = C.ToolbarConfigurationAllowUserCustomization
	ToolbarConfigurationAutoSaveConfiguration  ToolbarConfiguration = C.ToolbarConfigurationAutoSaveConfiguration
	ToolbarConfigurationShowSeparator          ToolbarConfiguration = C.ToolbarConfigurationShowSeparator
	ToolbarConfigurationDisplayModeIconOnly    ToolbarConfiguration = C.ToolbarConfigurationDisplayModeIconOnly
	ToolbarConfigurationDisplayModeTextOnly    ToolbarConfiguration = C.ToolbarConfigurationDisplayModeTextOnly
	ToolbarConfigurationDisplayModeIconAndText ToolbarConfiguration = C.ToolbarConfigurationDisplayModeIconAndText
)

type NSBezelStyle C.NSBezelStyle
type NSControlSize C.NSControlSize

// ControlStyle 的Go包装
type ControlStyle struct {
	Width       float64
	Height      float64
	BezelStyle  NSBezelStyle
	ControlSize NSControlSize
	Font        unsafe.Pointer
}

// ToolbarCallbackContext 的Go包装
type ToolbarCallbackContext struct {
	ClickCallback       C.ControlCallback
	TextChangedCallback C.ControlCallback
	UserData            unsafe.Pointer
}

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
	fmt.Printf("Button clicked: %s, Value: %s\n", id, val)
}

//export onTextChanged
func onTextChanged(identifier *C.char, value *C.char, userData unsafe.Pointer) {
	id := C.GoString(identifier)
	val := C.GoString(value)
	fmt.Printf("Text changed: %s, Value: %s\n", id, val)
}

func (m *Window) TestTool() {
	// 获取窗口句柄
	windowHandle := uintptr(lcl.PlatformWindow(m.Instance()))
	if windowHandle == 0 {
		log.Fatal("Failed to get window handle")
	}

	// 创建回调上下文
	callbackContext := ToolbarCallbackContext{
		ClickCallback:       (C.ControlCallback)(C.onButtonClicked),
		TextChangedCallback: (C.ControlCallback)(C.onTextChanged),
		UserData:            unsafe.Pointer(windowHandle),
	}

	// 配置窗口工具栏
	config := ToolbarConfigurationAllowUserCustomization |
		ToolbarConfigurationAutoSaveConfiguration |
		ToolbarConfigurationDisplayModeIconAndText

	ConfigureWindow(windowHandle, config, callbackContext)

	// 创建默认样式
	defaultStyle := CreateDefaultControlStyle()

	// 添加按钮
	AddToolbarButton(windowHandle, "run-button", "Run", "Run the program", defaultStyle)
	AddToolbarSpace(windowHandle)

	// 添加图片按钮
	AddToolbarImageButton(windowHandle, "settings-button", "NSPreferencesGeneral", "Open settings", defaultStyle)
	AddToolbarFlexibleSpace(windowHandle)

	// 添加文本框
	textFieldStyle := defaultStyle
	textFieldStyle.Width = 200
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
