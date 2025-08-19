package window

/*
#cgo CFLAGS: -mmacosx-version-min=11.0 -x objective-c
#cgo LDFLAGS: -mmacosx-version-min=11.0 -framework Cocoa
#include "window_config_darwin.h"
*/
import "C"
import "unsafe"

// ToolbarConfiguration 定义Go对应的类型和常量
type ToolbarConfiguration C.ToolbarConfiguration

const (
	ToolbarConfigurationNone                   ToolbarConfiguration = C.ToolbarConfigurationNone
	ToolbarConfigurationAllowUserCustomization ToolbarConfiguration = C.ToolbarConfigurationAllowUserCustomization
	ToolbarConfigurationAutoSaveConfiguration  ToolbarConfiguration = C.ToolbarConfigurationAutoSaveConfiguration
	ToolbarConfigurationShowSeparator          ToolbarConfiguration = C.ToolbarConfigurationShowSeparator
	ToolbarConfigurationDisplayModeIconOnly    ToolbarConfiguration = C.ToolbarConfigurationDisplayModeIconOnly
	ToolbarConfigurationDisplayModeTextOnly    ToolbarConfiguration = C.ToolbarConfigurationDisplayModeTextOnly
	ToolbarConfigurationDisplayModeIconAndText ToolbarConfiguration = C.ToolbarConfigurationDisplayModeIconAndText
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

// ControlStyle 的Go包装
type ControlStyle struct {
	Width       float64
	Height      float64
	BezelStyle  NSBezelStyle
	ControlSize NSControlSize
	Font        unsafe.Pointer
	position    ToolbarPosition
}

// ToolbarCallbackContext 的Go包装
type ToolbarCallbackContext struct {
	ClickCallback       C.ControlCallback
	TextChangedCallback C.ControlCallback
	UserData            unsafe.Pointer
}
