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
}

func NewNSButton(owner *NSToolBar, config ButtonItem, property ControlProperty) *NSButton {
	if config.Identifier == "" {
		config.Identifier = nextSerialNumber("Button")
	}
	if config.Title == "" {
		config.Title = config.Identifier
	}
	var cTitle *C.char
	cTitle = C.CString(config.Title)
	defer C.free(Pointer(cTitle))
	var cTooltip *C.char
	if config.Tips != "" {
		cTooltip = C.CString(config.Tips)
		defer C.free(Pointer(cTooltip))
	}
	cProperty := property.ToOC()
	cBtn := C.NewButton(cTitle, cTooltip, cProperty)
	return &NSButton{Control: Control{instance: Pointer(cBtn), owner: owner, property: &property}}
}

func (m NSButton) SetOnClick(fn ButtonAction) {

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
