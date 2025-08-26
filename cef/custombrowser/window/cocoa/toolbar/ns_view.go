package toolbar

/*
#cgo CFLAGS: -mmacosx-version-min=11.0 -x objective-c
#cgo LDFLAGS: -mmacosx-version-min=11.0 -framework Cocoa
#include "config.h"
#include "ns_view.h"
*/
import "C"

type IView interface {
	Instance() Pointer
	Identifier() string
}

type View struct {
	instance Pointer
	config   ItemBase
}

func NewNSView(config ItemBase) *View {
	if config.Identifier == "" {
		config.Identifier = nextSerialNumber("NSView")
	}
	cView := C.NewCustomView()
	m := &View{
		instance: Pointer(cView),
	}
	m.config = config
	return m
}

func (m *View) Instance() Pointer {
	return m.instance
}

func (m *View) Identifier() string {
	return m.config.Identifier
}
