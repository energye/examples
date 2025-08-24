package toolbar

/*
#cgo CFLAGS: -mmacosx-version-min=11.0 -x objective-c
#cgo LDFLAGS: -mmacosx-version-min=11.0 -framework Cocoa
#include "config.h"
*/
import "C"

type NSSearchField struct {
	Control
	config ControlSearchField
}

func NewNSSearchField(owner *NSToolBar, config ControlSearchField, property ControlProperty) *NSSearchField {
	if config.Identifier == "" {
		config.Identifier = nextSerialNumber("SearchField")
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
	cTextField := C.NewSearchField(owner.delegate, cIdentifier, cPlaceholder, cTooltip, cProperty)
	return &NSSearchField{
		Control: Control{
			instance: Pointer(cTextField),
			owner:    owner,
			property: &property,
			item:     config.ItemBase,
		},
		config: config,
	}
}

func (m *NSSearchField) GetText() string {
	cText := C.GetSearchFieldText(m.instance)
	return C.GoString(cText)
}

// SetText 设置搜索框文本
func (m *NSSearchField) SetText(text string) {
	cText := C.CString(text)
	defer C.free(Pointer(cText))
	C.SetSearchFieldText(m.instance, cText)
}

// UpdateSearchFieldWidth
func (m *NSSearchField) UpdateSearchFieldWidth(width int) {
	cWidth := C.CGFloat(width)
	C.UpdateSearchFieldWidth(m.instance, cWidth)
}
