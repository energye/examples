package gtkhelper

/*
#cgo pkg-config: gdk-3.0 gio-2.0 glib-2.0 gobject-2.0 gtk+-3.0
#include <gio/gio.h>
#include <gtk/gtk.h>
#include "gtk.go.h"
*/
import "C"
import (
	"unsafe"
)

// EventBox is a representation of GTK's GtkEventBox.
type EventBox struct {
	Bin
}

// native returns a pointer to the underlying GtkEventBox.
func (v *EventBox) native() *C.GtkEventBox {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGtkEventBox(p)
}

func marshalEventBox(p uintptr) (interface{}, error) {
	c := C.g_value_get_object((*C.GValue)(unsafe.Pointer(p)))
	obj := ToGoObject(unsafe.Pointer(c))
	return wrapEventBox(obj), nil
}

func wrapEventBox(obj *Object) *EventBox {
	if obj == nil {
		return nil
	}

	return &EventBox{Bin{Container{Widget{InitiallyUnowned{obj}}}}}
}

// NewEventBox is a wrapper around gtk_event_box_new().
func NewEventBox() *EventBox {
	c := C.gtk_event_box_new()
	if c == nil {
		return nil
	}
	obj := ToGoObject(unsafe.Pointer(c))
	return wrapEventBox(obj)
}

// SetAboveChild is a wrapper around gtk_event_box_set_above_child().
func (v *EventBox) SetAboveChild(aboveChild bool) {
	C.gtk_event_box_set_above_child(v.native(), CBool(aboveChild))
}

// GetAboveChild is a wrapper around gtk_event_box_get_above_child().
func (v *EventBox) GetAboveChild() bool {
	c := C.gtk_event_box_get_above_child(v.native())
	return GoBool(c)
}

// SetVisibleWindow is a wrapper around gtk_event_box_set_visible_window().
func (v *EventBox) SetVisibleWindow(visibleWindow bool) {
	C.gtk_event_box_set_visible_window(v.native(), CBool(visibleWindow))
}

// GetVisibleWindow is a wrapper around gtk_event_box_get_visible_window().
func (v *EventBox) GetVisibleWindow() bool {
	c := C.gtk_event_box_get_visible_window(v.native())
	return GoBool(c)
}

func (v *EventBox) SetOnClick(fn TButtonPressEvent) *SignalHandler {
	return registerAction(v, EsnButtonPressEvent, MakeButtonPressEvent(fn))
}

func (v *EventBox) SetOnLeave(fn TLeaveEnterNotifyEvent) *SignalHandler {
	return registerAction(v, EsnLeaveNotifyEvent, MakeLeaveEnterNotifyEvent(fn))
}

func (v *EventBox) SetOnEnter(fn TLeaveEnterNotifyEvent) *SignalHandler {
	return registerAction(v, EsnEnterNotifyEvent, MakeLeaveEnterNotifyEvent(fn))
}
