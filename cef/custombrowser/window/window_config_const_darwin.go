package window

/*
#cgo CFLAGS: -mmacosx-version-min=11.0 -x objective-c
#cgo LDFLAGS: -mmacosx-version-min=11.0 -framework Cocoa
#include "window_config_darwin.h"
*/
import "C"
import "unsafe"

type ToolbarDisplayMode = int

const (
	NSToolbarDisplayModeDefault      ToolbarDisplayMode = 0
	NSToolbarDisplayModeIconAndLabel ToolbarDisplayMode = 1
	NSToolbarDisplayModeIconOnly     ToolbarDisplayMode = 2
	NSToolbarDisplayModeLabelOnly    ToolbarDisplayMode = 3
)

type ToolbarStyle = int

const (
	NSWindowToolbarStyleAutomatic      ToolbarStyle = 0
	NSWindowToolbarStyleExpanded       ToolbarStyle = 1
	NSWindowToolbarStylePreference     ToolbarStyle = 2
	NSWindowToolbarStyleUnified        ToolbarStyle = 3
	NSWindowToolbarStyleUnifiedCompact ToolbarStyle = 4
)

// 边框样式
const (
	BezelStyleRounded           NSBezelStyle = C.NSBezelStyleRounded
	BezelStyleRegularSquare     NSBezelStyle = C.NSBezelStyleRegularSquare
	BezelStyleDisclosure        NSBezelStyle = C.NSBezelStyleDisclosure
	BezelStyleShadowlessSquare  NSBezelStyle = C.NSBezelStyleShadowlessSquare
	BezelStyleCircular          NSBezelStyle = C.NSBezelStyleCircular
	BezelStyleTexturedSquare    NSBezelStyle = C.NSBezelStyleTexturedSquare
	BezelStyleHelpButton        NSBezelStyle = C.NSBezelStyleHelpButton
	BezelStyleSmallSquare       NSBezelStyle = C.NSBezelStyleSmallSquare
	BezelStyleTexturedRounded   NSBezelStyle = C.NSBezelStyleTexturedRounded
	BezelStyleRoundRect         NSBezelStyle = C.NSBezelStyleRoundRect
	BezelStyleRecessed          NSBezelStyle = C.NSBezelStyleRecessed
	BezelStyleRoundedDisclosure NSBezelStyle = C.NSBezelStyleRoundedDisclosure
	BezelStyleInline            NSBezelStyle = C.NSBezelStyleInline
)

// 控件大小
const (
	ControlSizeMini    NSControlSize = C.NSControlSizeMini
	ControlSizeSmall   NSControlSize = C.NSControlSizeSmall
	ControlSizeRegular NSControlSize = C.NSControlSizeRegular
	ControlSizeLarge   NSControlSize = C.NSControlSizeLarge
)

type NSBezelStyle C.NSBezelStyle
type NSControlSize C.NSControlSize

// ControlProperty 的Go包装
type ControlProperty struct {
	Width          float64
	Height         float64
	BezelStyle     NSBezelStyle
	ControlSize    NSControlSize
	Font           unsafe.Pointer
	IsNavigational bool
	IsCenteredItem bool
}

// ToolbarCallbackContext 的Go包装
type ToolbarCallbackContext struct {
	ClickCallback       C.ControlCallback
	TextChangedCallback C.ControlCallback
	UserData            unsafe.Pointer
}

// ToolbarConfiguration 的Go包装
type ToolbarConfiguration struct {
	IsAllowsUserCustomization bool
	IsAutoSavesConfiguration  bool
	DisplayMode               ToolbarDisplayMode
	Style                     ToolbarStyle
}
