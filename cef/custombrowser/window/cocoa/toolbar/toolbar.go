package toolbar

/*
#cgo CFLAGS: -mmacosx-version-min=11.0 -x objective-c
#cgo LDFLAGS: -mmacosx-version-min=11.0 -framework Cocoa
#include "config.h"

extern void onDelegateEvent(ToolbarCallbackContext *cContext);
*/
import "C"
import (
	"fmt"
	"github.com/energye/lcl/lcl"
	"unsafe"
)

//export onDelegateEvent
func onDelegateEvent(cContext *C.ToolbarCallbackContext) {
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

type NSToolBar struct {
	toolbar  unsafe.Pointer
	delegate unsafe.Pointer
}

func Create(owner lcl.IForm, config ToolbarConfiguration) *NSToolBar {
	windowHandle := uintptr(lcl.PlatformWindow(owner.Instance()))
	if windowHandle == 0 {
		return nil
	}
	cConfig := ToolbarConfigurationToOC(config)
	callback := (C.ControlEventCallback)(C.onDelegateEvent)
	// 使用 C 类型的变量来接收输出
	var delegatePtr, toolbarPtr uintptr
	println(delegatePtr, toolbarPtr)
	// 获取这些变量的指针
	C.CreateToolbar(C.ulong(windowHandle), cConfig, callback,
		(*unsafe.Pointer)(unsafe.Pointer(&delegatePtr)),
		(*unsafe.Pointer)(unsafe.Pointer(&toolbarPtr)),
	)
	println(delegatePtr, toolbarPtr)
	return &NSToolBar{toolbar: unsafe.Pointer(delegatePtr), delegate: unsafe.Pointer(toolbarPtr)}
}
