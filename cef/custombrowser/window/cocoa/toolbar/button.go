package toolbar

/*
#cgo CFLAGS: -mmacosx-version-min=11.0 -x objective-c
#cgo LDFLAGS: -mmacosx-version-min=11.0 -framework Cocoa
#include "config.h"

*/
import "C"

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
	return &NSButton{Control: Control{instance: Pointer(cBtn), owner: owner, property: &property, item: config.ItemBase}, config: config}
}

func (m *NSButton) SetOnClick(fn ButtonAction) {
	registerEvent(m.config.Identifier, fn)
}

type NSImageButton struct {
	Control
	config ButtonItem
}

func NewNSImageButton(owner *NSToolBar, config ButtonItem, property ControlProperty) *NSImageButton {
	if config.Identifier == "" {
		config.Identifier = nextSerialNumber("ImageButton")
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
	return &NSImageButton{Control: Control{instance: Pointer(cBtn), owner: owner, property: &property, item: config.ItemBase}, config: config}
}

func (m *NSImageButton) SetOnClick(fn ButtonAction) {
	registerEvent(m.config.Identifier, fn)
}
