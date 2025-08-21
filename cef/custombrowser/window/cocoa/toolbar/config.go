package toolbar

/*
#cgo CFLAGS: -mmacosx-version-min=11.0 -x objective-c
#cgo LDFLAGS: -mmacosx-version-min=11.0 -framework Cocoa
#include "config.h"

extern void onControlEvent(ToolbarCallbackContext *cContext);
*/
import "C"
import (
	"fmt"
	"unsafe"
)

func controlPropertyToOC(property ControlProperty) C.ControlProperty {
	cProperty := C.ControlProperty{
		width:              C.CGFloat(property.Width),
		height:             C.CGFloat(property.Height),
		minWidth:           C.CGFloat(property.MinWidth),
		maxWidth:           C.CGFloat(property.MaxWidth),
		bezelStyle:         C.NSBezelStyle(property.BezelStyle),
		controlSize:        C.NSControlSize(property.ControlSize),
		font:               (*C.NSFont)(property.Font),
		IsNavigational:     C.BOOL(property.IsNavigational),
		IsCenteredItem:     C.BOOL(property.IsCenteredItem),
		VisibilityPriority: C.NSInteger(property.VisibilityPriority),
	}
	return cProperty
}

func ToolbarConfigurationToOC(config ToolbarConfiguration) C.ToolbarConfiguration {
	cConfig := C.ToolbarConfiguration{
		IsAllowsUserCustomization: C.BOOL(config.IsAllowsUserCustomization),
		IsAutoSavesConfiguration:  C.BOOL(config.IsAutoSavesConfiguration),
		Transparent:               C.BOOL(config.Transparent),
		ShowsToolbarButton:        C.BOOL(config.ShowsToolbarButton),
		SeparatorStyle:            C.NSUInteger(config.SeparatorStyle),
		DisplayMode:               C.NSUInteger(config.DisplayMode),
		SizeMode:                  C.NSUInteger(config.SizeMode),
		Style:                     C.NSUInteger(config.Style),
	}
	return cConfig
}

func ConfigureWindow(nsWindowHandle uintptr, config ToolbarConfiguration, owner unsafe.Pointer) {
	cConfig := ToolbarConfigurationToOC(config)
	callback := (C.ControlEventCallback)(C.onControlEvent)
	C.ConfigureWindow(C.ulong(nsWindowHandle), cConfig, callback, owner)
}

func AddToolbarButton(nsWindowHandle uintptr, identifier, title, tooltip string, property ControlProperty) {
	cIdentifier := C.CString(identifier)
	defer C.free(unsafe.Pointer(cIdentifier))

	cTitle := C.CString(title)
	defer C.free(unsafe.Pointer(cTitle))

	var cTooltip *C.char
	if tooltip != "" {
		cTooltip = C.CString(tooltip)
		defer C.free(unsafe.Pointer(cTooltip))
	}

	cProperty := controlPropertyToOC(property)

	C.AddToolbarButton(C.ulong(nsWindowHandle), cIdentifier, cTitle, cTooltip, cProperty)
}

func AddToolbarImageButton(nsWindowHandle uintptr, identifier, imageName, tooltip string, property ControlProperty) {
	cIdentifier := C.CString(identifier)
	defer C.free(unsafe.Pointer(cIdentifier))

	cImageName := C.CString(imageName)
	defer C.free(unsafe.Pointer(cImageName))

	var cTooltip *C.char
	if tooltip != "" {
		cTooltip = C.CString(tooltip)
		defer C.free(unsafe.Pointer(cTooltip))
	}

	cProperty := controlPropertyToOC(property)

	C.AddToolbarImageButton(C.ulong(nsWindowHandle), cIdentifier, cImageName, cTooltip, cProperty)
}

func AddToolbarTextField(nsWindowHandle uintptr, identifier, placeholder string, property ControlProperty) {
	cIdentifier := C.CString(identifier)
	defer C.free(unsafe.Pointer(cIdentifier))

	var cPlaceholder *C.char
	if placeholder != "" {
		cPlaceholder = C.CString(placeholder)
		defer C.free(unsafe.Pointer(cPlaceholder))
	}

	cProperty := controlPropertyToOC(property)

	C.AddToolbarTextField(C.ulong(nsWindowHandle), cIdentifier, cPlaceholder, cProperty)
}

func AddToolbarSearchField(nsWindowHandle uintptr, identifier, placeholder string, property ControlProperty) *NSSearchField {
	cIdentifier := C.CString(identifier)
	defer C.free(unsafe.Pointer(cIdentifier))

	var cPlaceholder *C.char
	if placeholder != "" {
		cPlaceholder = C.CString(placeholder)
		defer C.free(unsafe.Pointer(cPlaceholder))
	}

	cProperty := controlPropertyToOC(property)

	cSF := C.AddToolbarSearchField(C.ulong(nsWindowHandle), cIdentifier, cPlaceholder, cProperty)
	return &NSSearchField{instance: unsafe.Pointer(cSF)}
}

func AddToolbarCombobox(nsWindowHandle uintptr, identifier string, items []string, property ControlProperty) {
	cIdentifier := C.CString(identifier)
	defer C.free(unsafe.Pointer(cIdentifier))

	// 转换Go字符串切片为C字符串数组
	cItems := make([]*C.char, len(items))
	for i, item := range items {
		cItems[i] = C.CString(item)
	}
	cProperty := controlPropertyToOC(property)
	C.AddToolbarCombobox(C.ulong(nsWindowHandle), cIdentifier, (**C.char)(unsafe.Pointer(&cItems[0])), C.int(len(items)), cProperty)
	for i, _ := range items {
		C.free(unsafe.Pointer(cItems[i]))
	}
}

func AddToolbarCustomView(nsWindowHandle uintptr, identifier string, property ControlProperty) {
	cIdentifier := C.CString(identifier)
	defer C.free(unsafe.Pointer(cIdentifier))
	cProperty := controlPropertyToOC(property)
	C.AddToolbarCustomView(C.ulong(nsWindowHandle), cIdentifier, cProperty)
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

func SetToolbarControlHidden(nsWindowHandle uintptr, identifier string, hidden bool) {
	cIdentifier := C.CString(identifier)
	defer C.free(unsafe.Pointer(cIdentifier))
	C.SetToolbarControlHidden(C.ulong(nsWindowHandle), cIdentifier, C.bool(hidden))
}

func GetToolbarItemCount(nsWindowHandle uintptr) int {
	return int(C.GetToolbarItemCount(C.ulong(nsWindowHandle)))
}

func CreateDefaultControlProperty() ControlProperty {
	cProperty := C.CreateDefaultControlProperty()
	return ControlProperty{
		Width:              float64(cProperty.width),
		Height:             float64(cProperty.height),
		MinWidth:           float64(cProperty.minWidth),
		MaxWidth:           float64(cProperty.maxWidth),
		BezelStyle:         NSBezelStyle(cProperty.bezelStyle),
		ControlSize:        NSControlSize(cProperty.controlSize),
		Font:               unsafe.Pointer(cProperty.font),
		VisibilityPriority: int(cProperty.VisibilityPriority),
	}
}

// 导出Go回调函数供C调用

//export onControlEvent
func onControlEvent(cContext *C.ToolbarCallbackContext) {
	event := ToolbarCallbackContext{
		Type:       TccType(cContext.type_),
		Identifier: C.GoString(cContext.identifier),
		Value:      C.GoString(cContext.value),
		Index:      int(cContext.index),
		Owner:      cContext.owner,
		Sender:     cContext.sender,
	}
	fmt.Println("onControlEvent:", event)
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
