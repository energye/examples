package toolbar

/*
#cgo CFLAGS: -mmacosx-version-min=11.0 -x objective-c
#cgo LDFLAGS: -mmacosx-version-min=11.0 -framework Cocoa
#include "config.h"

*/
import "C"
import (
	"github.com/energye/examples/cef/custombrowser/window/cocoa/tool"
	"github.com/energye/lcl/lcl"
)

// NSToolBar 绑定到指定窗口
// 具有 toolbar delegate 实例
type NSToolBar struct {
	owner        lcl.IForm
	toolbar      Pointer
	delegate     Pointer
	config       *ToolbarConfiguration
	windowResize NotifyEvent
	controls     tool.HashMap[IControl] // NSView
	property     tool.HashMap[*ControlProperty]
	identifiers  tool.ArrayMap[string]
}

func Create(owner lcl.IForm, config ToolbarConfiguration) *NSToolBar {
	nsWindow := uintptr(lcl.PlatformWindow(owner.Instance()))
	if nsWindow == 0 {
		return nil
	}
	cConfig := ToolbarConfigurationToOC(config)
	callback := cControlEventCallback() //(C.ControlEventCallback)(C.onDelegateEvent)
	var delegatePtr, toolbarPtr uintptr
	C.CreateToolbar(C.ulong(nsWindow), cConfig, callback,
		(*Pointer)(Pointer(&delegatePtr)),
		(*Pointer)(Pointer(&toolbarPtr)),
	)
	toolbar := &NSToolBar{owner: owner,
		delegate: Pointer(delegatePtr), toolbar: Pointer(toolbarPtr),
		config: &config}
	// 注册默认事件
	RegisterEvent("__doWindowResize", MakeNotifyEvent(toolbar.doWindowResize))
	RegisterEvent("__doToolbarDefaultItemIdentifiers", MakeNotifyEvent(toolbar.doToolbarDefaultItemIdentifiers))
	RegisterEvent("__doToolbarAllowedItemIdentifiers", MakeNotifyEvent(toolbar.doToolbarAllowedItemIdentifiers))
	return toolbar
}

func (m *NSToolBar) doWindowResize(identifier string, owner Pointer, sender Pointer) *GoData {
	if m.windowResize != nil {
		return m.windowResize(identifier, owner, sender)
	}
	return nil
}

func (m *NSToolBar) doToolbarDefaultItemIdentifiers(identifier string, owner Pointer, sender Pointer) *GoData {
	println("doToolbarDefaultItemIdentifiers identifier:", identifier)
	ids := m.identifiers.Values()
	return &GoData{Type: GDtStringArray, StringArray: StringArray{Items: ids, Count: 3}}
}

func (m *NSToolBar) doToolbarAllowedItemIdentifiers(identifier string, owner Pointer, sender Pointer) *GoData {
	println("doToolbarAllowedItemIdentifiers identifier:", identifier)
	ids := m.identifiers.Values()
	ids = append(ids, GetStringConstValue(C.NSToolbarFlexibleSpaceItemIdentifier))
	ids = append(ids, GetStringConstValue(C.NSToolbarSpaceItemIdentifier))
	return &GoData{Type: GDtStringArray, StringArray: StringArray{Items: ids, Count: 3}}
}

func (m *NSToolBar) SetOnWindowResize(fn NotifyEvent) {
	m.windowResize = fn
}

func (m *NSToolBar) AddControl(control IControl) {
	if control == nil {
		println("[ERROR] AddControl 控件是 nil")
		return
	}
	var identifier *C.char
	identifier = C.CString(control.Identifier())
	defer C.free(Pointer(identifier))
	cProperty := control.Property().ToOC()
	C.ToolbarAddControl(m.delegate, m.toolbar, Pointer(control.Instance()), identifier, cProperty)
	m.identifiers.Add(control.Identifier(), control.Identifier())
	m.controls.Add(control.Identifier(), control)
}

func (m *NSToolBar) NewButton(config ButtonItem, property ControlProperty) *NSButton {
	return NewNSButton(m, config, property)
}
