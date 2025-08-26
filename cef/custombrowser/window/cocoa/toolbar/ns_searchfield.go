package toolbar

/*
#cgo CFLAGS: -mmacosx-version-min=11.0 -x objective-c
#cgo LDFLAGS: -mmacosx-version-min=11.0 -framework Cocoa
#include "config.h"
*/
import "C"

type NSSearchField struct {
	TextField
}

func NewNSSearchField(owner *NSToolBar, config ControlTextField, property ControlProperty) *NSSearchField {
	if config.Identifier == "" {
		config.Identifier = nextSerialNumber("SearchField")
	}
	var cPlaceholder *C.char
	cPlaceholder = C.CString(config.Placeholder)
	defer C.free(Pointer(cPlaceholder))
	var cTooltip *C.char
	if config.Tips != "" {
		cTooltip = C.CString(config.Tips)
		defer C.free(Pointer(cTooltip))
	}
	cProperty := property.ToOC()
	cTextField := C.NewSearchField(owner.delegate, cPlaceholder, cTooltip, cProperty)
	m := &NSSearchField{}
	m.config = config
	m.Control = Control{
		instance: Pointer(cTextField),
		owner:    owner,
		property: &property,
		item:     config.ItemBase,
	}
	return m
}
