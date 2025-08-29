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

// Container is a representation of GTK's GtkContainer.
type Container struct {
	Widget
}

// Bin is a representation of GTK's GtkBin.
type Bin struct {
	Container
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
