package window

/*
#cgo CFLAGS: -mmacosx-version-min=11.0 -x objective-c
#cgo LDFLAGS: -mmacosx-version-min=11.0 -framework Cocoa
#include "window_config_darwin.h"

extern void onButtonClicked(char *identifier, char *value, void *userData);
extern void onTextChanged(char *identifier, char *value, void *userData);
extern void onTextSubmit(char *identifier, char *value, void *userData);
extern void onRunOnMainThread(long id);

*/
import "C"
import (
	"fmt"
	"github.com/energye/lcl/lcl"
	"log"
	"sync"
	"time"
	"unsafe"
)

func controlPropertyToOC(property ControlProperty) C.ControlProperty {
	cProperty := C.ControlProperty{
		width:          C.CGFloat(property.Width),
		height:         C.CGFloat(property.Height),
		bezelStyle:     C.NSBezelStyle(property.BezelStyle),
		controlSize:    C.NSControlSize(property.ControlSize),
		font:           (*C.NSFont)(property.Font),
		IsNavigational: C.BOOL(property.IsNavigational),
		IsCenteredItem: C.BOOL(property.IsCenteredItem),
	}
	return cProperty
}

func ToolbarConfigurationToOC(config ToolbarConfiguration) C.ToolbarConfiguration {
	cConfig := C.ToolbarConfiguration{
		IsAllowsUserCustomization: C.BOOL(config.IsAllowsUserCustomization),
		IsAutoSavesConfiguration:  C.BOOL(config.IsAutoSavesConfiguration),
		Transparent:               C.BOOL(config.Transparent),
		SeparatorStyle:            C.NSUInteger(config.SeparatorStyle),
		DisplayMode:               C.NSUInteger(config.DisplayMode),
		Style:                     C.NSUInteger(config.Style),
	}
	return cConfig
}

func ConfigureWindow(nsWindowHandle uintptr, config ToolbarConfiguration, callbackContext ToolbarCallbackContext) {
	cConfig := ToolbarConfigurationToOC(config)
	C.ConfigureWindow(C.ulong(nsWindowHandle), cConfig, C.ToolbarCallbackContext{
		clickCallback:       callbackContext.ClickCallback,
		textChangedCallback: callbackContext.TextChangedCallback,
		textSubmitCallback:  callbackContext.TextSubmitCallback,
		userData:            callbackContext.UserData,
	})
}

func AddToolbarButton(nsWindowHandle uintptr, identifier, title, tooltip string, style ControlProperty) {
	cIdentifier := C.CString(identifier)
	defer C.free(unsafe.Pointer(cIdentifier))

	cTitle := C.CString(title)
	defer C.free(unsafe.Pointer(cTitle))

	var cTooltip *C.char
	if tooltip != "" {
		cTooltip = C.CString(tooltip)
		defer C.free(unsafe.Pointer(cTooltip))
	}

	cProperty := controlPropertyToOC(style)

	C.AddToolbarButton(C.ulong(nsWindowHandle), cIdentifier, cTitle, cTooltip, cProperty)
}

func AddToolbarImageButton(nsWindowHandle uintptr, identifier, imageName, tooltip string, style ControlProperty) {
	cIdentifier := C.CString(identifier)
	defer C.free(unsafe.Pointer(cIdentifier))

	cImageName := C.CString(imageName)
	defer C.free(unsafe.Pointer(cImageName))

	var cTooltip *C.char
	if tooltip != "" {
		cTooltip = C.CString(tooltip)
		defer C.free(unsafe.Pointer(cTooltip))
	}

	cProperty := controlPropertyToOC(style)

	C.AddToolbarImageButton(C.ulong(nsWindowHandle), cIdentifier, cImageName, cTooltip, cProperty)
}

func AddToolbarTextField(nsWindowHandle uintptr, identifier, placeholder string, style ControlProperty) {
	cIdentifier := C.CString(identifier)
	defer C.free(unsafe.Pointer(cIdentifier))

	var cPlaceholder *C.char
	if placeholder != "" {
		cPlaceholder = C.CString(placeholder)
		defer C.free(unsafe.Pointer(cPlaceholder))
	}

	cProperty := controlPropertyToOC(style)

	C.AddToolbarTextField(C.ulong(nsWindowHandle), cIdentifier, cPlaceholder, cProperty)
}

func AddToolbarSearchField(nsWindowHandle uintptr, identifier, placeholder string, style ControlProperty) {
	cIdentifier := C.CString(identifier)
	defer C.free(unsafe.Pointer(cIdentifier))

	var cPlaceholder *C.char
	if placeholder != "" {
		cPlaceholder = C.CString(placeholder)
		defer C.free(unsafe.Pointer(cPlaceholder))
	}

	cProperty := controlPropertyToOC(style)

	C.AddToolbarSearchField(C.ulong(nsWindowHandle), cIdentifier, cPlaceholder, cProperty)
}

func AddToolbarCombobox(nsWindowHandle uintptr, identifier string, items []string, style ControlProperty) {
	cIdentifier := C.CString(identifier)
	defer C.free(unsafe.Pointer(cIdentifier))

	// 转换Go字符串切片为C字符串数组
	cItems := make([]*C.char, len(items))
	for i, item := range items {
		cItems[i] = C.CString(item)
	}
	cProperty := controlPropertyToOC(style)
	C.AddToolbarCombobox(C.ulong(nsWindowHandle), cIdentifier, (**C.char)(unsafe.Pointer(&cItems[0])), C.int(len(items)), cProperty)
	for i, _ := range items {
		C.free(unsafe.Pointer(cItems[i]))
	}
}

func AddToolbarCustomView(nsWindowHandle uintptr, identifier string, style ControlProperty) {
	cIdentifier := C.CString(identifier)
	defer C.free(unsafe.Pointer(cIdentifier))
	cProperty := controlPropertyToOC(style)
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

func CreateDefaultControlProperty() ControlProperty {
	cProperty := C.CreateDefaultControlProperty()
	return ControlProperty{
		Width:       float64(cProperty.width),
		Height:      float64(cProperty.height),
		BezelStyle:  NSBezelStyle(cProperty.bezelStyle),
		ControlSize: NSControlSize(cProperty.controlSize),
		Font:        unsafe.Pointer(cProperty.font),
	}
}

func CreateControlProperty(width, height float64, bezelStyle NSBezelStyle, controlSize NSControlSize, font unsafe.Pointer) ControlProperty {
	cProperty := C.CreateControlProperty(
		C.CGFloat(width),
		C.CGFloat(height),
		C.NSBezelStyle(bezelStyle),
		C.NSControlSize(controlSize),
		font,
	)
	return ControlProperty{
		Width:       float64(cProperty.width),
		Height:      float64(cProperty.height),
		BezelStyle:  NSBezelStyle(cProperty.bezelStyle),
		ControlSize: NSControlSize(cProperty.controlSize),
		Font:        unsafe.Pointer(cProperty.font),
	}
}

// 导出Go回调函数供C调用

//export onButtonClicked
func onButtonClicked(identifier *C.char, value *C.char, userData unsafe.Pointer) {
	id := C.GoString(identifier)
	val := C.GoString(value)
	fmt.Println("onButtonClicked id:", id, "val:", val, "userData:", uintptr(userData))
}

//export onTextChanged
func onTextChanged(identifier *C.char, value *C.char, userData unsafe.Pointer) {
	id := C.GoString(identifier)
	val := C.GoString(value)
	fmt.Println("onTextChanged id:", id, "val:", val, "userData:", uintptr(userData))
}

//export onTextSubmit
func onTextSubmit(identifier *C.char, value *C.char, userData unsafe.Pointer) {
	id := C.GoString(identifier)
	val := C.GoString(value)
	fmt.Println("onTextSubmit id:", id, "val:", val, "userData:", uintptr(userData))
}

//export onRunOnMainThread
func onRunOnMainThread(id C.long) {
	doRunOnMainThread(int(id))
}

func registerRunOnMainThreadCallback() {
	C.RegisterRunOnMainThreadCallback(C.RunOnMainThreadCallback(C.onRunOnMainThread))
}

type runOnMainThreadFn func()

var (
	callbackFuncList     = make(map[int]runOnMainThreadFn)
	callbackFuncListLock = sync.Mutex{}
)

func doRunOnMainThread(id int) {
	fn, ok := callbackFuncList[id]
	delete(callbackFuncList, id)
	if ok {
		fn()
	}
}

func RunOnManThread(fn runOnMainThreadFn) {
	callbackFuncListLock.Lock()
	defer callbackFuncListLock.Unlock()
	id := int(uintptr(unsafe.Pointer(&fn)))
	callbackFuncList[id] = fn
	C.ExecuteRunOnMainThread(C.long(id))
}

func (m *Window) TestTool() {
	registerRunOnMainThreadCallback()
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
		TextSubmitCallback:  (C.ControlCallback)(C.onTextSubmit),
		UserData:            unsafe.Pointer(windowHandle),
	}

	// 配置窗口工具栏
	config := ToolbarConfiguration{
		DisplayMode: NSToolbarDisplayModeIconOnly,
		Transparent: true,
	}

	ConfigureWindow(windowHandle, config, callbackContext)

	// 创建默认样式
	defaultProperty := CreateDefaultControlProperty()
	defaultProperty.Height = 24
	//defaultProperty.BezelStyle = BezelStyleTexturedRounded // 边框样式
	//defaultProperty.ControlSize = ControlSizeLarge         // 控件大小
	defaultProperty.IsNavigational = true

	fmt.Println("当前控件总数：", int(C.GetToolbarItemCount(C.ulong(windowHandle))))
	// 添加按钮
	AddToolbarButton(windowHandle, "run-button", "Run", "Run the program", defaultProperty)
	//AddToolbarFlexibleSpace(windowHandle)

	// 添加文本框
	textFieldProperty := defaultProperty
	textFieldProperty.Height = 28
	textFieldProperty.IsNavigational = false
	textFieldProperty.IsCenteredItem = true
	//AddToolbarTextField(windowHandle, "text-field", "text...", textFieldProperty)
	AddToolbarSearchField(windowHandle, "search-field", "Search...", textFieldProperty)
	AddToolbarFlexibleSpace(windowHandle)

	// 添加下拉框
	comboProperty := defaultProperty
	comboProperty.IsNavigational = false
	comboProperty.Width = 100
	comboItems := []string{"Option 1", "Option 2", "Option 3"}
	AddToolbarCombobox(windowHandle, "options-combo", comboItems, comboProperty)

	// 添加图片按钮
	imageButtonProperty := defaultProperty
	imageButtonProperty.IsNavigational = false
	AddToolbarImageButton(windowHandle, "go-back", "arrow.left", "Open settings", imageButtonProperty)
	fmt.Println("当前控件总数：", int(C.GetToolbarItemCount(C.ulong(windowHandle))))
	SetToolbarControlHidden(windowHandle, "go-back", false)
	go func() {
		time.Sleep(time.Second * 2)
		RunOnManThread(func() {
			//SetToolbarControlHidden(windowHandle, "go-back", true)
			SetToolbarControlValue(windowHandle, "search-field", "Object-c UI线程 设置 Initial value")
		})
		time.Sleep(time.Second * 2)
		lcl.RunOnMainThreadAsync(func(id uint32) {
			SetToolbarControlValue(windowHandle, "search-field", "lcl.RunOnMainThreadAsync 设置 Initial value")
		})
		time.Sleep(time.Second * 2)
		lcl.RunOnMainThreadSync(func() {
			SetToolbarControlValue(windowHandle, "search-field", "lcl.RunOnMainThreadSync 设置 Initial value")

		})
	}()

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
