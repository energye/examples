package toolbar

/*
#cgo CFLAGS: -mmacosx-version-min=11.0 -x objective-c
#cgo LDFLAGS: -mmacosx-version-min=11.0 -framework Cocoa
#include "config.h"

*/
import "C"
import (
	"github.com/energye/lcl/lcl"
)

type NSToolBar struct {
	owner    lcl.IForm
	toolbar  Pointer
	delegate Pointer
	config   *ToolbarConfiguration
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
	registerEvent("__doWindowResize", makeWindowDidResizeAction(toolbar.doWindowResize))
	registerEvent("__doToolbarDefaultItemIdentifiers", makeToolbarDefaultItemIdentifiers(toolbar.doToolbarDefaultItemIdentifiers))
	return toolbar
}

func (m *NSToolBar) doWindowResize(identifier string, owner Pointer, sender Pointer) *GoData {
	return nil
}

func (m *NSToolBar) doToolbarDefaultItemIdentifiers(identifier string, owner Pointer, sender Pointer) *GoData {
	println("doToolbarDefaultItemIdentifiers identifier:", identifier)
	return &GoData{}
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
}

func (m *NSToolBar) NewButton(config ButtonItem, property ControlProperty) *NSButton {
	return NewNSButton(m, config, property)
}
