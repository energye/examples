package window

/*
#cgo pkg-config: gtk+-2.0
#cgo LDFLAGS: -lX11
#cgo CFLAGS: -DGTK_DISABLE_DEPRECATED=1 -Wno-deprecated-declarations -DGDK_DISABLE_DEPRECATION_WARNINGS

#include <X11/Xlib.h>
#include <X11/Xatom.h>
#include <gdk/gdkx.h>
#include <gtk/gtk.h>


*/
import "C"
import (
	"fmt"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
)

func (m *Window) HookWndProcMessage() {
	m.SetBorderStyleToFormBorderStyle(types.BsNone)
	m.SetOnWindowStateChange(func(sender lcl.IObject) {
		fmt.Println("OnWindowStateChange")
	})
}
