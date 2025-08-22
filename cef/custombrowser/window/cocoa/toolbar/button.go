package toolbar

/*
#cgo CFLAGS: -mmacosx-version-min=11.0 -x objective-c
#cgo LDFLAGS: -mmacosx-version-min=11.0 -framework Cocoa
#include "config.h"

void verifyNSButton(void* btn) {
    NSObject* obj = (NSObject*)btn;
    if ([obj isKindOfClass:[NSButton class]]) {
        NSLog(@"验证成功：是 NSButton 实例");
    } else {
        NSLog(@"验证失败：不是 NSButton 实例");
    }
}

*/
import "C"
import (
	"github.com/energye/lcl/lcl"
	"unsafe"
)

func verifyNSButton(nsButton Pointer) {
	C.verifyNSButton(nsButton)
}

type NSButton struct {
	instance Pointer
	property *ControlProperty
}

func AddNSButton(nsWindow uintptr, config ButtonItem, property ControlProperty) *NSButton {
	if config.Identifier == "" {
		config.Identifier = nextSerialNumber("Button")
	}
	cIdentifier := C.CString(config.Identifier)
	defer C.free(Pointer(cIdentifier))
	var cTitle *C.char
	if config.Title != "" {
		cTitle = C.CString(config.Title)
		defer C.free(Pointer(cTitle))
	}
	var cTooltip *C.char
	if config.Tips != "" {
		cTooltip = C.CString(config.Tips)
		defer C.free(Pointer(cTooltip))
	}
	cProperty := property.ToOC()
	C.AddToolbarButton(C.ulong(nsWindow), cIdentifier, cTitle, cTooltip, cProperty)

	return &NSButton{}
}

func LCLToNSButton(button lcl.IButton) {
	if button.HandleAllocated() {

	}
	handle := button.Handle()
	ptr := unsafe.Pointer(handle)
	//nsButton := (*C.NSButton)(ptr)
	verifyNSButton(ptr)
}
