package window

/*
#cgo CFLAGS: -mmacosx-version-min=11.0 -x objective-c
#cgo LDFLAGS: -mmacosx-version-min=11.0 -framework Cocoa
#include "window_config_darwin.h"

*/
import "C"
import (
	"fmt"
	"github.com/energye/lcl/lcl"
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

func (m *Window) windowShow() {
	nsWindow := unsafe.Pointer(lcl.PlatformWindow(m.Instance()))
	println(nsWindow)
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

// 导出回调函数
//
//export goToolbarHandler
func goToolbarHandler(itemID *C.char) {
	id := C.GoString(itemID)
	fmt.Printf("工具栏事件: %s\n", id)

	// 实际业务处理逻辑
	switch id {
	case "MainIDE.Back":
		fmt.Println("执行后退操作")
	case "MainIDE.Forward":
		fmt.Println("执行前进操作")
	case "MainIDE.Search":
		fmt.Println("执行搜索操作")
	case "MainIDE.Command":
		fmt.Println("执行命令操作")
	}
}
