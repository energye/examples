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
	var cTitle *C.char
	cTitle = C.CString(config.Title)
	defer C.free(Pointer(cTitle))
	var cTooltip *C.char
	if config.Tips != "" {
		cTooltip = C.CString(config.Tips)
		defer C.free(Pointer(cTooltip))
	}
	cProperty := property.ToOC()
	cBtn := C.NewButton(owner.delegate, cTitle, cTooltip, cProperty)
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
	var cImage *C.char
	cImage = C.CString(config.IconName)
	defer C.free(Pointer(cImage))
	var cTooltip *C.char
	if config.Tips != "" {
		cTooltip = C.CString(config.Tips)
		defer C.free(Pointer(cTooltip))
	}
	cProperty := property.ToOC()
	cBtn := C.NewImageButtonFormImage(owner.delegate, cImage, cTooltip, cProperty)
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
	var cTooltip *C.char
	if config.Tips != "" {
		cTooltip = C.CString(config.Tips)
		defer C.free(Pointer(cTooltip))
	}
	cProperty := property.ToOC()
	cData := (*C.uint8_t)(unsafe.Pointer(&imageBytes[0]))
	cLen := C.size_t(len(imageBytes))
	cBtn := C.NewImageButtonFormBytes(owner.delegate, cData, cLen, cTooltip, cProperty)
	return &NSImageButton{Control: Control{instance: Pointer(cBtn), owner: owner, property: &property, item: config.ItemBase}, config: config}
}

func (m *NSImageButton) SetImageFromPath(imagePath string) {
	cImagePath := C.CString(imagePath)
	defer C.free(unsafe.Pointer(cImagePath))
	C.SetButtonImageFromPath(m.instance, cImagePath)
}

// SetImageFromBytes 设置按钮图片（使用字节数据）
func (m *NSImageButton) SetImageFromBytes(data []byte) {
	if len(data) == 0 {
		return
	}
	cData := (*C.uint8_t)(unsafe.Pointer(&data[0]))
	cLen := C.size_t(len(data))
	C.SetButtonImageFromBytes(m.instance, cData, cLen)
}

func (m *NSImageButton) SetOnClick(fn NotifyEvent) {
	RegisterEvent(m.config.Identifier, MakeNotifyEvent(fn))
}
