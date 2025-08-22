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
	"sync"
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
	fmt.Printf("onDelegateEvent event: %+v\n", event)
	fn := eventList[event.Identifier]
	if fn == nil {
		return
	}
	switch fn.(type) {
	case ButtonAction:
		fn.(ButtonAction)(event.Identifier, event.Owner, event.Sender)
	}
}

var (
	eventList = make(map[string]any)
	eventLock sync.Mutex
)

func registerEvent(identifier string, fn any) {
	eventLock.Lock()
	defer eventLock.Unlock()
	eventList[identifier] = fn
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
		delegate: Pointer(delegatePtr), toolbar: Pointer(toolbarPtr),
		config: &config}
}

func (m *NSToolBar) AddControl(control IControl) {
	var identifier *C.char
	identifier = C.CString("TestBtn1")
	defer C.free(Pointer(identifier))
	cProperty := control.Property().ToOC()
	C.ToolbarAddControl(m.delegate, m.toolbar, Pointer(control.Instance()), identifier, cProperty)
}

func (m *NSToolBar) NewButton(config ButtonItem, property ControlProperty) *NSButton {
	return NewNSButton(m, config, property)
}

func (m *NSToolBar) NewLCLButton() *NSButton {
	button := lcl.NewButton(m.owner)
	button.SetParent(m.partBox)
	nsButton := LCLToNSButton(m, button)
	return nsButton
}
