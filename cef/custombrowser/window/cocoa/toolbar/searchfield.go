package toolbar

/*
#cgo CFLAGS: -mmacosx-version-min=11.0 -x objective-c
#cgo LDFLAGS: -mmacosx-version-min=11.0 -framework Cocoa
#include "config.h"
*/
import "C"

type NSSearchField struct {
	instance Pointer
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
