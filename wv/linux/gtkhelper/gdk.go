package gtkhelper

// #cgo pkg-config: gdk-3.0 glib-2.0 gobject-2.0
// #include <gdk/gdk.h>
// #include "gdk.go.h"
import "C"
import (
	"unsafe"
)

// WindowEdge is a representation of GDK's GdkWindowEdge
type WindowEdge int

const (
	WINDOW_EDGE_NORTH_WEST WindowEdge = C.GDK_WINDOW_EDGE_NORTH_WEST
	WINDOW_EDGE_NORTH      WindowEdge = C.GDK_WINDOW_EDGE_NORTH
	WINDOW_EDGE_NORTH_EAST WindowEdge = C.GDK_WINDOW_EDGE_NORTH_EAST
	WINDOW_EDGE_WEST       WindowEdge = C.GDK_WINDOW_EDGE_WEST
	WINDOW_EDGE_EAST       WindowEdge = C.GDK_WINDOW_EDGE_EAST
	WINDOW_EDGE_SOUTH_WEST WindowEdge = C.GDK_WINDOW_EDGE_SOUTH_WEST
	WINDOW_EDGE_SOUTH      WindowEdge = C.GDK_WINDOW_EDGE_SOUTH
	WINDOW_EDGE_SOUTH_EAST WindowEdge = C.GDK_WINDOW_EDGE_SOUTH_EAST
)

// ButtonType constants
type ButtonType uint

const (
	BUTTON_PRIMARY   ButtonType = C.GDK_BUTTON_PRIMARY
	BUTTON_MIDDLE    ButtonType = C.GDK_BUTTON_MIDDLE
	BUTTON_SECONDARY ButtonType = C.GDK_BUTTON_SECONDARY
)

// Event is a representation of GDK's GdkEvent.
type Event struct {
	GdkEvent *C.GdkEvent
}

// native returns a pointer to the underlying GdkEvent.
func (v *Event) native() *C.GdkEvent {
	if v == nil {
		return nil
	}
	return v.GdkEvent
}

// Native returns a pointer to the underlying GdkEvent.
func (v *Event) Native() uintptr {
	return uintptr(unsafe.Pointer(v.native()))
}

func marshalEvent(p uintptr) (interface{}, error) {
	c := C.g_value_get_boxed((*C.GValue)(unsafe.Pointer(p)))
	return &Event{(*C.GdkEvent)(unsafe.Pointer(c))}, nil
}

func (v *Event) free() {
	C.gdk_event_free(v.native())
}

func (v *Event) ScanCode() int {
	return int(C.gdk_event_get_scancode(v.native()))
}

// Rectangle is a representation of GDK's GdkRectangle type.
type Rectangle struct {
	GdkRectangle C.GdkRectangle
}

// RectangleNew helper function to create a GdkRectanlge
func RectangleNew(x, y, width, height int) *Rectangle {
	var r C.GdkRectangle
	r.x = C.int(x)
	r.y = C.int(y)
	r.width = C.int(width)
	r.height = C.int(height)
	return &Rectangle{r}
}
