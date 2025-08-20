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

type TccType = int

const (
	TCCClicked            TccType = 1
	TCCTextDidChange      TccType = 2
	TCCTextDidEndEditing  TccType = 3
	TCCSelectionChanged   TccType = 4
	TCCSelectionDidChange TccType = 5
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

type ItemVisibilityPriority = int

const (
	NSToolbarItemVisibilityPriorityStandard ItemVisibilityPriority = 0
	NSToolbarItemVisibilityPriorityLow      ItemVisibilityPriority = -1000
	NSToolbarItemVisibilityPriorityHigh     ItemVisibilityPriority = 1000
	NSToolbarItemVisibilityPriorityUser     ItemVisibilityPriority = 2000
)

type TitlebarSeparatorStyle = int

const (
	NSTitlebarSeparatorStyleAutomatic TitlebarSeparatorStyle = 0
	NSTitlebarSeparatorStyleNone      TitlebarSeparatorStyle = 1
	NSTitlebarSeparatorStyleLine      TitlebarSeparatorStyle = 2
	NSTitlebarSeparatorStyleShadow    TitlebarSeparatorStyle = 3
)

type NSBezelStyle C.NSBezelStyle
type NSControlSize C.NSControlSize

type ControlItemBase struct {
	Identifier   string
	Priority     int
	Navigational bool
}

type ControlItemUI struct {
	ControlItemBase
	IconName string
	Title    string
	Tips     string
	Bordered bool
}

type ControlItemAction struct {
	ControlItemUI
	OnAction func(identifier, value string, userData uintptr)
}

type ControlItem ControlItemAction

type ControlItemSearch struct {
	ControlItemAction
	SendWhole         bool
	SendImmediately   bool
	ResignsWithCancel bool
	PreferredWidth    float32
}

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

// ToolbarConfiguration 的Go包装
type ToolbarConfiguration struct {
	IsAllowsUserCustomization bool
	IsAutoSavesConfiguration  bool
	Transparent               bool
	SeparatorStyle            TitlebarSeparatorStyle
	DisplayMode               ToolbarDisplayMode
	Style                     ToolbarStyle
}

type ToolbarCallbackContext struct {
	Type       TccType        // 事件类型
	Identifier string         // 控件标识
	Value      string         // 控件值
	Index      int            // 值索引
	Owner      unsafe.Pointer // 所属对象
	Sender     unsafe.Pointer // 控件对象
}
