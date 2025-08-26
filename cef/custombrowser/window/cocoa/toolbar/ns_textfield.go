package toolbar

/*
#cgo CFLAGS: -mmacosx-version-min=11.0 -x objective-c
#cgo LDFLAGS: -mmacosx-version-min=11.0 -framework Cocoa
#include "config.h"

*/
import "C"

type TextField struct {
	Control
	config ControlTextField
}

type NSTextField struct {
	TextField
}

func NewNSTextField(owner *NSToolBar, config ControlTextField, property ControlProperty) *NSTextField {
	if config.Identifier == "" {
		config.Identifier = nextSerialNumber("TextField")
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
	cTextField := C.NewTextField(owner.delegate, cPlaceholder, cTooltip, cProperty)
	m := &NSTextField{}
	m.config = config
	m.Control = Control{
		instance: Pointer(cTextField),
		owner:    owner,
		property: &property,
		item:     config.ItemBase,
	}
	return m
}

func (m *TextField) SetOnChange(fn TextEvent) {
	RegisterEvent(m.config.Identifier, MakeTextChangeEvent(fn))
}

func (m *TextField) SetOnCommit(fn TextEvent) {
	RegisterEvent(m.config.Identifier, MakeTextCommitEvent(fn))
}

func (m *TextField) GetText() string {
	cText := C.GetTextFieldText(m.instance)
	return C.GoString(cText)
}

// SetText 设置搜索框文本
func (m *TextField) SetText(text string) {
	cText := C.CString(text)
	defer C.free(Pointer(cText))
	C.SetTextFieldText(m.instance, cText)
}

func (m *TextField) UpdateTextFieldWidth(width int) {
	cWidth := C.CGFloat(width)
	C.UpdateTextFieldWidth(m.instance, cWidth)
}
