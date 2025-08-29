package gtkhelper

// #cgo pkg-config: gdk-3.0 glib-2.0 gobject-2.0
// #include <gdk/gdk.h>
// #include "gdk.go.h"
import "C"

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

// Button constants
type Button uint

const (
	BUTTON_PRIMARY   Button = C.GDK_BUTTON_PRIMARY
	BUTTON_MIDDLE    Button = C.GDK_BUTTON_MIDDLE
	BUTTON_SECONDARY Button = C.GDK_BUTTON_SECONDARY
)
