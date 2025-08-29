package gtkhelper

// #cgo pkg-config: gdk-3.0 gio-2.0 glib-2.0 gobject-2.0 gtk+-3.0
// #include <gio/gio.h>
// #include <gtk/gtk.h>
// #include "gtk.go.h"
import "C"
import (
	"unsafe"
)

// WindowType is a representation of GTK's GtkWindowType.
type WindowType int

const (
	WINDOW_TOPLEVEL WindowType = C.GTK_WINDOW_TOPLEVEL
	WINDOW_POPUP    WindowType = C.GTK_WINDOW_POPUP
)

/*
 * GdkGravity
 */

type Gravity int

const (
	GDK_GRAVITY_NORTH_WEST = C.GDK_GRAVITY_NORTH_WEST
	GDK_GRAVITY_NORTH      = C.GDK_GRAVITY_NORTH
	GDK_GRAVITY_NORTH_EAST = C.GDK_GRAVITY_NORTH_EAST
	GDK_GRAVITY_WEST       = C.GDK_GRAVITY_WEST
	GDK_GRAVITY_CENTER     = C.GDK_GRAVITY_CENTER
	GDK_GRAVITY_EAST       = C.GDK_GRAVITY_EAST
	GDK_GRAVITY_SOUTH_WEST = C.GDK_GRAVITY_SOUTH_WEST
	GDK_GRAVITY_SOUTH      = C.GDK_GRAVITY_SOUTH
	GDK_GRAVITY_SOUTH_EAST = C.GDK_GRAVITY_SOUTH_EAST
	GDK_GRAVITY_STATIC     = C.GDK_GRAVITY_STATIC
)

// WindowTypeHint is a representation of GDK's GdkWindowTypeHint
type WindowTypeHint int

const (
	WINDOW_TYPE_HINT_NORMAL        WindowTypeHint = C.GDK_WINDOW_TYPE_HINT_NORMAL
	WINDOW_TYPE_HINT_DIALOG        WindowTypeHint = C.GDK_WINDOW_TYPE_HINT_DIALOG
	WINDOW_TYPE_HINT_MENU          WindowTypeHint = C.GDK_WINDOW_TYPE_HINT_MENU
	WINDOW_TYPE_HINT_TOOLBAR       WindowTypeHint = C.GDK_WINDOW_TYPE_HINT_TOOLBAR
	WINDOW_TYPE_HINT_SPLASHSCREEN  WindowTypeHint = C.GDK_WINDOW_TYPE_HINT_SPLASHSCREEN
	WINDOW_TYPE_HINT_UTILITY       WindowTypeHint = C.GDK_WINDOW_TYPE_HINT_UTILITY
	WINDOW_TYPE_HINT_DOCK          WindowTypeHint = C.GDK_WINDOW_TYPE_HINT_DOCK
	WINDOW_TYPE_HINT_DESKTOP       WindowTypeHint = C.GDK_WINDOW_TYPE_HINT_DESKTOP
	WINDOW_TYPE_HINT_DROPDOWN_MENU WindowTypeHint = C.GDK_WINDOW_TYPE_HINT_DROPDOWN_MENU
	WINDOW_TYPE_HINT_POPUP_MENU    WindowTypeHint = C.GDK_WINDOW_TYPE_HINT_POPUP_MENU
	WINDOW_TYPE_HINT_TOOLTIP       WindowTypeHint = C.GDK_WINDOW_TYPE_HINT_TOOLTIP
	WINDOW_TYPE_HINT_NOTIFICATION  WindowTypeHint = C.GDK_WINDOW_TYPE_HINT_NOTIFICATION
	WINDOW_TYPE_HINT_COMBO         WindowTypeHint = C.GDK_WINDOW_TYPE_HINT_COMBO
	WINDOW_TYPE_HINT_DND           WindowTypeHint = C.GDK_WINDOW_TYPE_HINT_DND
)

// ModifierType is a representation of GDK's GdkModifierType.
type ModifierType uint

const (
	SHIFT_MASK    ModifierType = C.GDK_SHIFT_MASK
	LOCK_MASK                  = C.GDK_LOCK_MASK
	CONTROL_MASK               = C.GDK_CONTROL_MASK
	MOD1_MASK                  = C.GDK_MOD1_MASK
	MOD2_MASK                  = C.GDK_MOD2_MASK
	MOD3_MASK                  = C.GDK_MOD3_MASK
	MOD4_MASK                  = C.GDK_MOD4_MASK
	MOD5_MASK                  = C.GDK_MOD5_MASK
	BUTTON1_MASK               = C.GDK_BUTTON1_MASK
	BUTTON2_MASK               = C.GDK_BUTTON2_MASK
	BUTTON3_MASK               = C.GDK_BUTTON3_MASK
	BUTTON4_MASK               = C.GDK_BUTTON4_MASK
	BUTTON5_MASK               = C.GDK_BUTTON5_MASK
	SUPER_MASK                 = C.GDK_SUPER_MASK
	HYPER_MASK                 = C.GDK_HYPER_MASK
	META_MASK                  = C.GDK_META_MASK
	RELEASE_MASK               = C.GDK_RELEASE_MASK
	MODIFIER_MASK              = C.GDK_MODIFIER_MASK
)

func marshalModifierType(p uintptr) (interface{}, error) {
	c := C.g_value_get_flags((*C.GValue)(unsafe.Pointer(p)))
	return ModifierType(c), nil
}

// castWidget takes a native GtkWidget and casts it to the appropriate Go struct.
func castWidget(c *C.GtkWidget) IWidget {
	wdt := new(Widget)
	wdt.Object = ToGoObject(unsafe.Pointer(c))
	return wdt
}
