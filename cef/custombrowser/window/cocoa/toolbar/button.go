package toolbar

/*
#cgo CFLAGS: -mmacosx-version-min=11.0 -x objective-c
#cgo LDFLAGS: -mmacosx-version-min=11.0 -framework Cocoa
#include "config.h"

*/
import "C"
import (
	"github.com/energye/examples/cef/custombrowser/window/cocoa"
	"github.com/energye/lcl/lcl"
	"unsafe"
)

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
	cBtn := C.AddToolbarButton(C.ulong(nsWindow), cIdentifier, cTitle, cTooltip, cProperty)
	return &NSButton{instance: Pointer(cBtn)}
}

func LCLToNSButton(button lcl.IButton) *NSButton {
	if !button.HandleAllocated() {
		return nil
	}
	handle := button.Handle()
	if cocoa.VerifyWidget(handle) {
		return nil
	}
	ptr := unsafe.Pointer(handle)
	//nsButton := (*C.NSButton)(ptr)
	return &NSButton{instance: ptr}
}
