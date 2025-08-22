package toolbar

/*
#cgo CFLAGS: -mmacosx-version-min=11.0 -x objective-c
#cgo LDFLAGS: -mmacosx-version-min=11.0 -framework Cocoa
#include "config.h"

extern void onDelegateEvent(ToolbarCallbackContext *cContext);
*/
import "C"
import (
	"fmt"
	"github.com/energye/lcl/lcl"
)

//export onDelegateEvent
func onDelegateEvent(cContext *C.ToolbarCallbackContext) {
	event := ToolbarCallbackContext{
		Type:       TccType(cContext.type_),
		Identifier: C.GoString(cContext.identifier),
		Value:      C.GoString(cContext.value),
		Index:      int(cContext.index),
		Owner:      cContext.owner,
		Sender:     cContext.sender,
	}
	fmt.Println("onControlEvent:", event)
}

type NSToolBar struct {
	owner    lcl.IForm
	partBox  lcl.IPanel
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
	callback := (C.ControlEventCallback)(C.onDelegateEvent)
	var delegatePtr, toolbarPtr uintptr
	C.CreateToolbar(C.ulong(nsWindow), cConfig, callback,
		(*Pointer)(Pointer(&delegatePtr)),
		(*Pointer)(Pointer(&toolbarPtr)),
	)
	partBox := lcl.NewPanel(owner)
	partBox.SetParent(owner)
	partBox.SetBounds(0, 0, 1, 1)
	partBox.SetVisible(false)
	return &NSToolBar{owner: owner, partBox: partBox,
		toolbar: Pointer(delegatePtr), delegate: Pointer(toolbarPtr),
		config: &config}
}

func (m *NSToolBar) AddControl(control IControl) {

}

func (m *NSToolBar) AddButton(config ButtonItem, property ControlProperty) *NSButton {
	nsWindow := uintptr(lcl.PlatformWindow(m.owner.Instance()))
	if nsWindow == 0 {
		return nil
	}
	return AddNSButton(nsWindow, config, property)
}

func (m *NSToolBar) AddLCLButton() *NSButton {
	button := lcl.NewButton(m.owner)
	button.SetParent(m.partBox)
	nsButton := LCLToNSButton(button)
	return nsButton
}
