package toolbar

/*
#cgo CFLAGS: -mmacosx-version-min=11.0 -x objective-c
#cgo LDFLAGS: -mmacosx-version-min=11.0 -framework Cocoa
#include "config.h"
*/
import "C"
import (
	"strconv"
	"unsafe"
)

type Pointer = unsafe.Pointer

type ButtonAction func(identifier string, owner Pointer, sender Pointer)
type SearchFieldAction func(identifier string, owner Pointer, sender Pointer)

type ItemBase struct {
	Identifier   string
	Priority     ItemVisibilityPriority
	Navigational bool
}

type ItemUI struct {
	ItemBase
	IconName string
	Title    string
	Tips     string
	Bordered bool
}

type ButtonItem struct {
	ItemUI
	OnAction ButtonAction
}

func (m *ButtonItem) ToOC() {

}

type ControlSearchField struct {
	ItemUI
	SendWhole         bool
	SendImmediately   bool
	ResignsWithCancel bool
	PreferredWidth    float32
	OnAction          SearchFieldAction
}

// ControlProperty 的Go包装
type ControlProperty struct {
	Width              float64
	Height             float64
	MinWidth           float64
	MaxWidth           float64
	BezelStyle         NSBezelStyle
	ControlSize        NSControlSize
	Font               Pointer
	IsNavigational     bool
	IsCenteredItem     bool
	VisibilityPriority ItemVisibilityPriority
}

func (m *ControlProperty) ToOC() C.ControlProperty {
	cProperty := C.ControlProperty{
		width:              C.CGFloat(m.Width),
		height:             C.CGFloat(m.Height),
		minWidth:           C.CGFloat(m.MinWidth),
		maxWidth:           C.CGFloat(m.MaxWidth),
		bezelStyle:         C.NSBezelStyle(m.BezelStyle),
		controlSize:        C.NSControlSize(m.ControlSize),
		font:               (*C.NSFont)(m.Font),
		IsNavigational:     C.BOOL(m.IsNavigational),
		IsCenteredItem:     C.BOOL(m.IsCenteredItem),
		VisibilityPriority: C.NSInteger(m.VisibilityPriority),
	}
	return cProperty
}

type ToolbarCallbackContext struct {
	Type       TccType // 事件类型
	Identifier string  // 控件标识
	Value      string  // 控件值
	Index      int     // 值索引
	Owner      Pointer // 所属对象
	Sender     Pointer // 控件对象
}

// ToolbarConfiguration 的Go包装
type ToolbarConfiguration struct {
	IsAllowsUserCustomization bool
	IsAutoSavesConfiguration  bool
	Transparent               bool
	ShowsToolbarButton        bool // 隐藏工具栏默认的"显示/隐藏"按钮（右侧）
	SeparatorStyle            TitlebarSeparatorStyle
	DisplayMode               ToolbarDisplayMode
	SizeMode                  ToolbarSizeMode
	Style                     ToolbarStyle
}

func ToolbarConfigurationToOC(config ToolbarConfiguration) C.ToolbarConfiguration {
	cConfig := C.ToolbarConfiguration{
		IsAllowsUserCustomization: C.BOOL(config.IsAllowsUserCustomization),
		IsAutoSavesConfiguration:  C.BOOL(config.IsAutoSavesConfiguration),
		Transparent:               C.BOOL(config.Transparent),
		ShowsToolbarButton:        C.BOOL(config.ShowsToolbarButton),
		SeparatorStyle:            C.NSUInteger(config.SeparatorStyle),
		DisplayMode:               C.NSUInteger(config.DisplayMode),
		SizeMode:                  C.NSUInteger(config.SizeMode),
		Style:                     C.NSUInteger(config.Style),
	}
	return cConfig
}

var serialNumber = make(map[string]int)

func nextSerialNumber(type_ string) string {
	var r int
	if sn, ok := serialNumber[type_]; ok {
		r = sn
	} else {
		r = 0
	}
	r++
	serialNumber[type_] = r
	return type_ + strconv.Itoa(r)
}
