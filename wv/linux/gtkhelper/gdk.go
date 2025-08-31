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

// NewRectangle helper function to create a GdkRectanlge
func NewRectangle(x, y, width, height int) *Rectangle {
	var r C.GdkRectangle
	r.x = C.int(x)
	r.y = C.int(y)
	r.width = C.int(width)
	r.height = C.int(height)
	return &Rectangle{r}
}

/*
 * GdkRGBA
 */
// To create a GdkRGBA you have to use NewRGBA function.
type RGBA struct {
	rgba *C.GdkRGBA
}

func marshalRGBA(p uintptr) (interface{}, error) {
	c := C.g_value_get_boxed((*C.GValue)(unsafe.Pointer(p)))
	return WrapRGBA(unsafe.Pointer(c)), nil
}

func WrapRGBA(p unsafe.Pointer) *RGBA {
	return wrapRGBA((*C.GdkRGBA)(p))
}

func wrapRGBA(cRgba *C.GdkRGBA) *RGBA {
	if cRgba == nil {
		return nil
	}
	return &RGBA{cRgba}
}

func NewRGBA(values ...float64) *RGBA {
	cRgba := new(C.GdkRGBA)
	for i, value := range values {
		switch i {
		case 0:
			cRgba.red = C.gdouble(value)
		case 1:
			cRgba.green = C.gdouble(value)
		case 2:
			cRgba.blue = C.gdouble(value)
		case 3:
			cRgba.alpha = C.gdouble(value)
		}
	}
	return wrapRGBA(cRgba)
}

func (c *RGBA) Floats() []float64 {
	return []float64{
		float64(c.rgba.red),
		float64(c.rgba.green),
		float64(c.rgba.blue),
		float64(c.rgba.alpha)}
}

func (c *RGBA) Native() uintptr {
	return uintptr(unsafe.Pointer(c.rgba))
}

// SetColors sets all colors values in the RGBA.
func (c *RGBA) SetColors(r, g, b, a float64) {
	c.rgba.red = C.gdouble(r)
	c.rgba.green = C.gdouble(g)
	c.rgba.blue = C.gdouble(b)
	c.rgba.alpha = C.gdouble(a)
}

/*
 * The following methods (Get/Set) are made for
 * more convenient use of the GdkRGBA object
 */
// GetRed get red value from the RGBA.
func (c *RGBA) GetRed() float64 {
	return float64(c.rgba.red)
}

// GetGreen get green value from the RGBA.
func (c *RGBA) GetGreen() float64 {
	return float64(c.rgba.green)
}

// GetBlue get blue value from the RGBA.
func (c *RGBA) GetBlue() float64 {
	return float64(c.rgba.blue)
}

// GetAlpha get alpha value from the RGBA.
func (c *RGBA) GetAlpha() float64 {
	return float64(c.rgba.alpha)
}

// SetRed set red value in the RGBA.
func (c *RGBA) SetRed(red float64) {
	c.rgba.red = C.gdouble(red)
}

// SetGreen set green value in the RGBA.
func (c *RGBA) SetGreen(green float64) {
	c.rgba.green = C.gdouble(green)
}

// SetBlue set blue value in the RGBA.
func (c *RGBA) SetBlue(blue float64) {
	c.rgba.blue = C.gdouble(blue)
}

// SetAlpha set alpha value in the RGBA.
func (c *RGBA) SetAlpha(alpha float64) {
	c.rgba.alpha = C.gdouble(alpha)
}

// Parse is a representation of gdk_rgba_parse().
func (c *RGBA) Parse(spec string) bool {
	cstr := (*C.gchar)(C.CString(spec))
	defer C.free(unsafe.Pointer(cstr))
	return GoBool(C.gdk_rgba_parse(c.rgba, cstr))
}

// String is a representation of gdk_rgba_to_string().
func (c *RGBA) String() string {
	return C.GoString((*C.char)(C.gdk_rgba_to_string(c.rgba)))
}

// free is a representation of gdk_rgba_free().
func (c *RGBA) free() {
	C.gdk_rgba_free(c.rgba)
}

// Equal is a representation of gdk_rgba_equal().
func (c *RGBA) Equal(rgba *RGBA) bool {
	return GoBool(C.gdk_rgba_equal(
		C.gconstpointer(c.rgba),
		C.gconstpointer(rgba.rgba)))
}

// Hash is a representation of gdk_rgba_hash().
func (c *RGBA) Hash() uint {
	return uint(C.gdk_rgba_hash(C.gconstpointer(c.rgba)))
}

/*
 * GdkAtom
 */

// Atom is a representation of GDK's GdkAtom.
type Atom uintptr

// native returns the underlying GdkAtom.
func (v Atom) native() C.GdkAtom {
	return C.toGdkAtom(unsafe.Pointer(uintptr(v)))
}

func (v Atom) Name() string {
	c := C.gdk_atom_name(v.native())
	defer C.g_free(C.gpointer(c))
	return C.GoString((*C.char)(c))
}

// GdkAtomIntern is a wrapper around gdk_atom_intern
func GdkAtomIntern(atomName string, onlyIfExists bool) Atom {
	cstr := C.CString(atomName)
	defer C.free(unsafe.Pointer(cstr))
	c := C.gdk_atom_intern((*C.gchar)(cstr), CBool(onlyIfExists))
	return Atom(uintptr(unsafe.Pointer(c)))
}

// Selections
const (
	SELECTION_PRIMARY       Atom = 1
	SELECTION_SECONDARY     Atom = 2
	SELECTION_CLIPBOARD     Atom = 69
	TARGET_BITMAP           Atom = 5
	TARGET_COLORMAP         Atom = 7
	TARGET_DRAWABLE         Atom = 17
	TARGET_PIXMAP           Atom = 20
	TARGET_STRING           Atom = 31
	SELECTION_TYPE_ATOM     Atom = 4
	SELECTION_TYPE_BITMAP   Atom = 5
	SELECTION_TYPE_COLORMAP Atom = 7
	SELECTION_TYPE_DRAWABLE Atom = 17
	SELECTION_TYPE_INTEGER  Atom = 19
	SELECTION_TYPE_PIXMAP   Atom = 20
	SELECTION_TYPE_WINDOW   Atom = 33
	SELECTION_TYPE_STRING   Atom = 31
)

// VisualType is a representation of GDK's GdkVisualType.
type VisualType int

const (
	VISUAL_STATIC_GRAY  VisualType = C.GDK_VISUAL_STATIC_GRAY
	VISUAL_GRAYSCALE    VisualType = C.GDK_VISUAL_GRAYSCALE
	VISUAL_STATIC_COLOR VisualType = C.GDK_VISUAL_STATIC_COLOR
	ISUAL_PSEUDO_COLOR  VisualType = C.GDK_VISUAL_PSEUDO_COLOR
	VISUAL_TRUE_COLOR   VisualType = C.GDK_VISUAL_TRUE_COLOR
	VISUAL_DIRECT_COLOR VisualType = C.GDK_VISUAL_DIRECT_COLOR
)

func marshalVisualType(p uintptr) (interface{}, error) {
	c := C.g_value_get_enum((*C.GValue)(unsafe.Pointer(p)))
	return VisualType(c), nil
}
