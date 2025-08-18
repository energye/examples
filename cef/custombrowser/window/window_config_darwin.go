package window

/*
#cgo CFLAGS: -mmacosx-version-min=11.0 -x objective-c
#cgo LDFLAGS: -mmacosx-version-min=11.0 -framework Cocoa
#include "window_config_darwin.h"

extern void onItemClickCallback(char *itemID);
extern void onSearchTextChangedCallback(char *text);
*/
import "C"
import (
	"fmt"
	"github.com/energye/lcl/lcl"
	"log"
	"unsafe"
)

const (
	TransparentTitleBar     = true
	TitleBarSeparatorStyle  = C.NSTitlebarSeparatorStyleAutomatic
	ToolbarStyle            = C.NSWindowToolbarStyleUnifiedCompact
	ToolbarDisplayMode      = C.NSToolbarDisplayModeIconOnly
	AllowsUserCustomization = false
	AutosavesConfiguration  = false
)

// 设置事件处理器
func itemClickHandler(itemID string) {
	log.Printf("工具栏项点击: %s", itemID)
	switch itemID {
	case "MainIDE.Back":
		fmt.Println("执行后退操作")
	case "MainIDE.Forward":
		fmt.Println("执行前进操作")
	case "MainIDE.Command":
		fmt.Println("显示命令面板")
	}
}

func searchTextChangedHandler(text string) {
	log.Printf("搜索文本变化: %s", text)
	// 这里可以添加搜索逻辑
}

// 主动获取搜索框内容
func getSearchText(windowHandle unsafe.Pointer, textName string) string {
	cTextName := C.CString(textName)
	defer C.free(unsafe.Pointer(cTextName))
	cText := C.GetSearchFieldText(C.ulong(uintptr(windowHandle)), cTextName)
	defer C.free(unsafe.Pointer(cText)) // 释放C分配的内存

	return C.GoString(cText)
}

func (m *Window) windowShow() {
	nsWindow := unsafe.Pointer(lcl.PlatformWindow(m.Instance()))
	println(nsWindow)
	// 将 Go 回调函数注册到 C
	C.SetItemClickCallback(C.ItemClickCallback(C.onItemClickCallback))
	C.SetSearchTextChangedCallback(C.SearchTextChangedCallback(C.onSearchTextChangedCallback))

	// 配置标题栏和工具栏
	C.ConfigureWindow(
		C.ulong(uintptr(unsafe.Pointer(nsWindow))),
		C.bool(TransparentTitleBar),
		C.int(TitleBarSeparatorStyle),
		C.int(ToolbarStyle),
		C.int(ToolbarDisplayMode),
		C.bool(AllowsUserCustomization),
		C.bool(AutosavesConfiguration),
	)
}

// 导出给 C 调用的回调函数

//export onItemClickCallback
func onItemClickCallback(itemID *C.char) {
	if itemClickHandler != nil {
		itemClickHandler(C.GoString(itemID))
	}
}

//export onSearchTextChangedCallback
func onSearchTextChangedCallback(text *C.char) {
	if searchTextChangedHandler != nil {
		searchTextChangedHandler(C.GoString(text))
	}
}
