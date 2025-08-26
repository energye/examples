package toolbar

/*
#cgo CFLAGS: -mmacosx-version-min=11.0 -x objective-c
#cgo LDFLAGS: -mmacosx-version-min=11.0 -framework Cocoa
#include "config.h"

*/
import "C"

func (m *NSToolBar) AddFlexibleSpace() {
	m.items.Add(GetStringConstValue(C.NSToolbarFlexibleSpaceItemIdentifier), nil)
	C.AddToolbarFlexibleSpace(m.toolbar)
}

func (m *NSToolBar) AddSpace() {
	m.items.Add(GetStringConstValue(C.NSToolbarSpaceItemIdentifier), nil)
	C.AddToolbarSpace(m.toolbar)
}
