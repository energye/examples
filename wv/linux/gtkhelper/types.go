package gtkhelper

/*
#cgo pkg-config: gdk-3.0 gio-2.0 glib-2.0 gobject-2.0 gtk+-3.0
#include <gtk/gtk.h>
#include <gio/gio.h>
#include <stdlib.h>
#include <glib.h>
#include <glib-object.h>
#include "gtk.go.h"
*/
import "C"
import (
	"errors"
	"unsafe"
)

var nilPtrErr = errors.New("cgo returned unexpected nil pointer")

// IWidget is an interface type implemented by all structs
// embedding a Widget.  It is meant to be used as an argument type
// for wrapper functions that wrap around a C GTK function taking a
// GtkWidget.
type IWidget interface {
	toWidget() *C.GtkWidget
	ToWidget() *Widget
	Set(name string, value interface{}) error
}

// Container is a representation of GTK's GtkContainer.
type Container struct {
	Widget
}

// Bin is a representation of GTK's GtkBin.
type Bin struct {
	Container
}

// Widget is a representation of GTK's GtkWidget.
type Widget struct {
	InitiallyUnowned
}

// native returns a pointer to the underlying GtkWidget.
func (v *Widget) native() *C.GtkWidget {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGtkWidget(p)
}

func (v *Widget) toWidget() *C.GtkWidget {
	if v == nil {
		return nil
	}
	return v.native()
}

// ToWidget is a helper getter, e.g.: it returns *gtk.Label as a *gtk.Widget.
// In other cases, where you have a gtk.IWidget, use the type assertion.
func (v *Widget) ToWidget() *Widget {
	return v
}

// InitiallyUnowned is a representation of GLib's GInitiallyUnowned.
type InitiallyUnowned struct {
	// This must be a pointer so copies of the ref-sinked object
	// do not outlive the original object, causing an unref
	// finalizer to prematurely run.
	*Object
}

// Object is a representation of GLib's GObject.
type Object struct {
	GObject *C.GObject
}

// Set calls SetProperty.
func (v *Object) Set(name string, value interface{}) error {
	return nil
}

// Event is a representation of GDK's GdkEvent.
type Event struct {
	GdkEvent *C.GdkEvent
}

func CBool(b bool) C.gboolean {
	if b {
		return C.gboolean(1)
	}
	return C.gboolean(0)
}

func GoBool(b C.gboolean) bool {
	return b != C.FALSE
}

func ToCObject(p unsafe.Pointer) *C.GObject {
	return (*C.GObject)(p)
}
func ToGoObject(instance unsafe.Pointer) *Object {
	cObj := ToCObject(instance)
	return &Object{GObject: cObj}
}

func GoString(cStr *C.gchar) string {
	return C.GoString((*C.char)(cStr))
}
