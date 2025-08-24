package toolbar

/*
#cgo CFLAGS: -mmacosx-version-min=11.0 -x objective-c
#cgo LDFLAGS: -mmacosx-version-min=11.0 -framework Cocoa
#include "config.h"

*/
import "C"

func (m *NSToolBar) AddFlexibleSpace() {
	C.AddToolbarFlexibleSpace(m.toolbar)
	m.controls.Add(GetStringConstValue(C.NSToolbarFlexibleSpaceItemIdentifier), &ControlInfo{})
}

func (m *NSToolBar) AddSpace() {
	C.AddToolbarSpace(m.toolbar)
	m.controls.Add(GetStringConstValue(C.NSToolbarSpaceItemIdentifier), &ControlInfo{})
}
