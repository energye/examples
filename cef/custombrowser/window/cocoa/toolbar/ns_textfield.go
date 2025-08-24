package toolbar

/*
#cgo CFLAGS: -mmacosx-version-min=11.0 -x objective-c
#cgo LDFLAGS: -mmacosx-version-min=11.0 -framework Cocoa
#include "config.h"

*/
import "C"

type NSTextField struct {
	Control
	config ControlTextField
}

func NewNSTextField(owner *NSToolBar, config ControlTextField, property ControlProperty) *NSTextField {
	if config.Identifier == "" {
		config.Identifier = nextSerialNumber("TextField")
	}
	var cIdentifier *C.char
	cIdentifier = C.CString(config.Identifier)
	defer C.free(Pointer(cIdentifier))
	var cPlaceholder *C.char
	cPlaceholder = C.CString(config.Placeholder)
	defer C.free(Pointer(cPlaceholder))
	var cTooltip *C.char
	if config.Tips != "" {
		cTooltip = C.CString(config.Tips)
		defer C.free(Pointer(cTooltip))
	}
	cProperty := property.ToOC()
	cTextField := C.NewTextField(owner.delegate, cIdentifier, cPlaceholder, cTooltip, cProperty)
	return &NSTextField{
		Control: Control{
			instance: Pointer(cTextField),
			owner:    owner,
			property: &property,
			item:     config.ItemBase,
		},
		config: config,
	}
}

func (m *NSTextField) SetOnChange(fn TextEvent) {
	RegisterEvent(m.config.Identifier, MakeTextChangeEventEvent(fn))
}

func (m *NSTextField) SetOnCommit(fn TextEvent) {
	RegisterEvent(m.config.Identifier, MakeTextCommitEventEvent(fn))
}
