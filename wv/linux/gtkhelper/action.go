package gtkhelper

/*
#cgo pkg-config: gdk-3.0 gio-2.0 glib-2.0 gobject-2.0 gtk+-3.0
#include <stdlib.h>
#include <string.h>
#include <gtk/gtk.h>

extern void go_on_clicked(GtkWidget* widget, gpointer user_data);
extern void go_on_activated(GtkWidget* widget, gpointer user_data);

static void remove_signal_handler(GtkWidget* widget, gulong handler_id) {
    if (handler_id > 0 && widget != NULL) {
        g_signal_handler_disconnect(widget, handler_id);
    }
}
*/
import "C"
import (
	"fmt"
	"unsafe"
)

func doOnClick(widget, userData unsafe.Pointer) {
	fmt.Println("doOnClick userData:", uintptr(userData))
}

//export go_on_clicked
func go_on_clicked(widget *C.GtkWidget, user_data C.gpointer) {
	doOnClick(unsafe.Pointer(widget), unsafe.Pointer(user_data))
}

//export go_on_activated
func go_on_activated(widget *C.GtkWidget, user_data C.gpointer) {
}

type signalHandler struct {
	widget    C.gpointer
	handlerID C.gulong
}

func (m *signalHandler) Disconnect() {
	if m != nil && m.handlerID > 0 {
		C.remove_signal_handler((*C.GtkWidget)(unsafe.Pointer(m.widget)), m.handlerID)
		m.handlerID = 0
	}
}

func registerAction(widget, userData C.gpointer, signal string) *signalHandler {
	cb := C.GCallback(C.go_on_clicked)
	name := C.CString(signal)
	defer C.free(unsafe.Pointer(name))
	handlerId := C.g_signal_connect_data(widget, name, cb, userData, nil, 0)
	return &signalHandler{
		widget:    widget,
		handlerID: handlerId,
	}
}

func registerClickAction(widget, userData C.gpointer) *signalHandler {
	return registerAction(widget, userData, "clicked")
}
