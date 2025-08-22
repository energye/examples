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
	Control
	config ButtonItem
}

func NewNSButton(owner *NSToolBar, config ButtonItem, property ControlProperty) *NSButton {
	if config.Identifier == "" {
		config.Identifier = nextSerialNumber("Button")
	}
	if config.Title == "" {
		config.Title = config.Identifier
	}
	var cIdentifier *C.char
	cIdentifier = C.CString(config.Identifier)
	defer C.free(Pointer(cIdentifier))
	var cTitle *C.char
	cTitle = C.CString(config.Title)
	defer C.free(Pointer(cTitle))
	var cTooltip *C.char
	if config.Tips != "" {
		cTooltip = C.CString(config.Tips)
		defer C.free(Pointer(cTooltip))
	}
	cProperty := property.ToOC()
	cBtn := C.NewButton(owner.delegate, cIdentifier, cTitle, cTooltip, cProperty)
	return &NSButton{Control: Control{instance: Pointer(cBtn), owner: owner, property: &property}, config: config}
}

func (m *NSButton) SetOnClick(fn ButtonAction) {
	registerEvent(m.config.Identifier, fn)
}

func LCLToNSButton(owner *NSToolBar, button lcl.IButton) *NSButton {
	if !button.HandleAllocated() {
		return nil
	}
	handle := button.Handle()
	if cocoa.VerifyWidget(handle) {
		return nil
	}
	btnHandle := unsafe.Pointer(handle)
	//nsButton := (*C.NSButton)(btnHandle)
	return &NSButton{Control: Control{instance: Pointer(btnHandle), owner: owner}}
}
