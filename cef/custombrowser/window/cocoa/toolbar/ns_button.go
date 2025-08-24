package toolbar

/*
#cgo CFLAGS: -mmacosx-version-min=11.0 -x objective-c
#cgo LDFLAGS: -mmacosx-version-min=11.0 -framework Cocoa
#include "config.h"

*/
import "C"
import "unsafe"

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

func (m *NSButton) SetOnClick(fn NotifyEvent) {
	RegisterEvent(m.config.Identifier, MakeNotifyEvent(fn))
}

type NSImageButton struct {
	Control
	config ButtonItem
}

func NewNSImageButtonForImage(owner *NSToolBar, config ButtonItem, property ControlProperty) *NSImageButton {
	if config.Identifier == "" {
		config.Identifier = nextSerialNumber("ImageButton")
	}
	var cIdentifier *C.char
	cIdentifier = C.CString(config.Identifier)
	defer C.free(Pointer(cIdentifier))
	var cImage *C.char
	cImage = C.CString(config.IconName)
	defer C.free(Pointer(cImage))
	var cTooltip *C.char
	if config.Tips != "" {
		cTooltip = C.CString(config.Tips)
		defer C.free(Pointer(cTooltip))
	}
	cProperty := property.ToOC()
	cBtn := C.NewImageButtonFormImage(owner.delegate, cIdentifier, cImage, cTooltip, cProperty)
	return &NSImageButton{Control: Control{instance: Pointer(cBtn), owner: owner, property: &property, item: config.ItemBase}, config: config}
}

func NewNSImageButtonForBytes(owner *NSToolBar, imageBytes []byte, config ButtonItem, property ControlProperty) *NSImageButton {
	// 将字节数组传递给Objective-C创建图片按钮
	if len(imageBytes) == 0 {
		return nil
	}
	if config.Identifier == "" {
		config.Identifier = nextSerialNumber("ImageButton")
	}
	var cIdentifier *C.char
	cIdentifier = C.CString(config.Identifier)
	defer C.free(Pointer(cIdentifier))
	var cTooltip *C.char
	if config.Tips != "" {
		cTooltip = C.CString(config.Tips)
		defer C.free(Pointer(cTooltip))
	}
	cProperty := property.ToOC()
	cData := (*C.uint8_t)(unsafe.Pointer(&imageBytes[0]))
	cLen := C.size_t(len(imageBytes))
	cBtn := C.NewImageButtonFormBytes(owner.delegate, cIdentifier, cData, cLen, cTooltip, cProperty)
	return &NSImageButton{Control: Control{instance: Pointer(cBtn), owner: owner, property: &property, item: config.ItemBase}, config: config}
}

func (m *NSImageButton) SetOnClick(fn NotifyEvent) {
	RegisterEvent(m.config.Identifier, MakeNotifyEvent(fn))
}
