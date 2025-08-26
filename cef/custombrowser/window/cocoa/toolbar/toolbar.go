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
	"strings"
)

// NSToolBar 绑定到指定窗口
// 具有 toolbar delegate 实例
type NSToolBar struct {
	owner        lcl.IForm
	toolbar      Pointer
	delegate     Pointer
	config       *ToolbarConfiguration
	windowResize NotifyEvent
	items        tool.ArrayMap[IView]
}

//type ControlInfo struct {
//	control  IControl
//	property *ControlProperty
//}

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
	RegisterEvent("__doDelegateToolbar", MakeDelegateToolbarEvent(toolbar.doDelegateToolbar))
	return toolbar
}

func (m *NSToolBar) doWindowResize(identifier string, owner Pointer, sender Pointer) *GoArguments {
	if m.windowResize != nil {
		return m.windowResize(identifier, owner, sender)
	}
	return nil
}

func (m *NSToolBar) doToolbarDefaultItemIdentifiers(identifier string, owner Pointer, sender Pointer) *GoArguments {
	println("doToolbarDefaultItemIdentifiers 事件ID:", identifier, "控件IDs:", strings.Join(m.items.Keys(), " "))
	result := &GoArguments{}
	m.items.Iterate(func(key string, value IView) bool {
		result.Add(key)
		return false
	})
	return result
}

func (m *NSToolBar) doToolbarAllowedItemIdentifiers(identifier string, owner Pointer, sender Pointer) *GoArguments {
	println("doToolbarAllowedItemIdentifiers 事件ID:", identifier, "控件IDs:", strings.Join(m.items.Keys(), " "))
	result := &GoArguments{}
	m.items.Iterate(func(key string, value IView) bool {
		result.Add(key)
		return false
	})
	// 系统项
	result.Add(GetStringConstValue(C.NSToolbarFlexibleSpaceItemIdentifier))
	result.Add(GetStringConstValue(C.NSToolbarSpaceItemIdentifier))
	return result
}

func (m *NSToolBar) doDelegateToolbar(arguments *OCGoArguments, owner Pointer, sender Pointer) *GoArguments {
	itemIdentifier := arguments.GetString(0)
	println("doDelegateToolbar itemIdentifier:", itemIdentifier, "ControlCount:", m.items.Count())
	item := m.items.Get(itemIdentifier)
	if item == nil {
		println("control 是空的？？")
		return nil
	}

	if control, ok := item.(IControl); ok {
		cProperty := control.Property().ToOCMalloc()
		result := &GoArguments{}
		result.Add(control.Instance())               // 0 控件
		result.AddCStruct(CStructPointer(cProperty)) // 1 属性 C结构体
		println(cProperty)
		return result
	} else if _, ok := item.(IView); ok {

	}
	return nil
}

func (m *NSToolBar) SetOnWindowResize(fn NotifyEvent) {
	m.windowResize = fn
}

func (m *NSToolBar) AddItem(item IView) {
	if item == nil {
		println("[ERROR] AddItem 控件是 nil")
		return
	}
	if control, ok := item.(IControl); ok {
		var identifier *C.char
		identifier = C.CString(control.Identifier())
		defer C.free(Pointer(identifier))
		cProperty := control.Property().ToOC()
		// 先保存控件
		m.items.Add(control.Identifier(), control)
		// 然后添加到 toolbar
		C.ToolbarAddItem(m.delegate, m.toolbar, control.Instance(), identifier, cProperty)
	} else if _, ok := item.(IView); ok {
		var identifier *C.char
		identifier = C.CString(control.Identifier())
		defer C.free(Pointer(identifier))
		// 先保存控件
		m.items.Add(control.Identifier(), nil)
		// 然后添加到 toolbar
		temp := &ControlProperty{}
		C.ToolbarAddItem(m.delegate, m.toolbar, control.Instance(), identifier, temp.ToOC())
	}
}

func (m *NSToolBar) NewButton(config ButtonItem, property ControlProperty) *NSButton {
	return NewNSButton(m, config, property)
}

func (m *NSToolBar) NewImageButtonForImage(config ButtonItem, property ControlProperty) *NSImageButton {
	return NewNSImageButtonForImage(m, config, property)
}

func (m *NSToolBar) NewImageButtonForBytes(imageBytes []byte, config ButtonItem, property ControlProperty) *NSImageButton {
	return NewNSImageButtonForBytes(m, imageBytes, config, property)
}

func (m *NSToolBar) NewTextField(config ControlTextField, property ControlProperty) *NSTextField {
	return NewNSTextField(m, config, property)
}

func (m *NSToolBar) NewSearchField(config ControlTextField, property ControlProperty) *NSSearchField {
	return NewNSSearchField(m, config, property)
}

func (m *NSToolBar) ItemCount() int {
	return int(C.ToolbarItemCount(m.toolbar))
}
