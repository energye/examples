package gtkhelper

// #include <gdk/gdk.h>
// #include "gdk.go.h"
import "C"
import (
	"unsafe"
)

// Screen is a representation of GDK's GdkScreen.
type Screen struct {
	*Object
}

// native returns a pointer to the underlying GdkScreen.
func (v *Screen) native() *C.GdkScreen {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGdkScreen(p)
}

// Native returns a pointer to the underlying GdkScreen.
func (v *Screen) Native() uintptr {
	return uintptr(unsafe.Pointer(v.native()))
}

func marshalScreen(p uintptr) (interface{}, error) {
	c := C.g_value_get_object((*C.GValue)(unsafe.Pointer(p)))
	obj := &Object{ToCObject(unsafe.Pointer(c))}
	return &Screen{obj}, nil
}

func toScreen(s *C.GdkScreen) (*Screen, error) {
	if s == nil {
		return nil, nilPtrErr
	}
	return &Screen{ToGoObject(unsafe.Pointer(s))}, nil
}

// GetRGBAVisual is a wrapper around gdk_screen_get_rgba_visual().
func (v *Screen) GetRGBAVisual() (*Visual, error) {
	c := C.gdk_screen_get_rgba_visual(v.native())
	if c == nil {
		return nil, nilPtrErr
	}

	return &Visual{ToGoObject(unsafe.Pointer(c))}, nil
}

// GetSystemVisual is a wrapper around gdk_screen_get_system_visual().
func (v *Screen) GetSystemVisual() (*Visual, error) {
	c := C.gdk_screen_get_system_visual(v.native())
	if c == nil {
		return nil, nilPtrErr
	}

	return &Visual{ToGoObject(unsafe.Pointer(c))}, nil
}

// ScreenGetDefault is a wrapper around gdk_screen_get_default().
func ScreenGetDefault() (*Screen, error) {
	return toScreen(C.gdk_screen_get_default())
}

// IsComposited is a wrapper around gdk_screen_is_composited().
func (v *Screen) IsComposited() bool {
	return GoBool(C.gdk_screen_is_composited(v.native()))
}

// GetRootWindow is a wrapper around gdk_screen_get_root_window().
func (v *Screen) GetRootWindow() *Window {
	return toWindow(C.gdk_screen_get_root_window(v.native()))
}

// GetDisplay is a wrapper around gdk_screen_get_display().
func (v *Screen) GetDisplay() (*Display, error) {
	return toDisplay(C.gdk_screen_get_display(v.native()))
}

func toString(c *C.gchar) (string, error) {
	if c == nil {
		return "", nilPtrErr
	}
	return C.GoString((*C.char)(c)), nil
}

// GetResolution is a wrapper around gdk_screen_get_resolution().
func (v *Screen) GetResolution() float64 {
	return float64(C.gdk_screen_get_resolution(v.native()))
}

// SetResolution is a wrapper around gdk_screen_set_resolution().
func (v *Screen) SetResolution(r float64) {
	C.gdk_screen_set_resolution(v.native(), C.gdouble(r))
}
