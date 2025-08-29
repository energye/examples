package gtkhelper

/*
#cgo pkg-config: gdk-3.0 gio-2.0 glib-2.0 gobject-2.0 gtk+-3.0
#include <stdlib.h>
#include <string.h>
#include <gtk/gtk.h>

extern void go_on_clicked(GtkWidget* widget, gpointer user_data);
extern void go_on_activated(GtkWidget* widget, gpointer user_data);

static void remove_signal_handler(GtkWidget* widget, gulong handler_id) {
  	g_print("尝试移除信号处理器: handler_id=%lu, widget=%p\n", handler_id, widget);
    if (handler_id > 0 && widget != NULL) {
        g_signal_handler_disconnect(widget, handler_id);
        g_print("移除信号处理器成功: handler_id=%lu\n", handler_id);
    }
}
*/
import "C"
import (
	"fmt"
	"sync"
	"unsafe"
)

func doOnClick(widget, userData unsafe.Pointer) {
	fmt.Println("doOnClick userData:", uintptr(userData))
	id := uintptr(userData)
	if cb, ok := eventList[id]; ok {
		context := &CallbackContext{widget: widget}
		cb.cb(context)
	}
}

//export go_on_clicked
func go_on_clicked(widget *C.GtkWidget, user_data C.gpointer) {
	doOnClick(unsafe.Pointer(widget), unsafe.Pointer(user_data))
}

//export go_on_activated
func go_on_activated(widget *C.GtkWidget, user_data C.gpointer) {
}

// 事件列表
var (
	eventList = make(map[uintptr]*Callback)
	eventLock sync.Mutex
)

// RegisterEvent 事件注册，使用控件唯一标识 + 事件类型做为事件唯一id
func RegisterEvent(id uintptr, fn *Callback) {
	eventLock.Lock()
	defer eventLock.Unlock()
	eventList[id] = fn
}

type SignalHandler struct {
	widget    *C.GtkWidget
	handlerID C.gulong
	id        uintptr
}

func (m *SignalHandler) Disconnect() {
	if m != nil && m.handlerID > 0 {
		C.remove_signal_handler(m.widget, m.handlerID)
		m.handlerID = 0
		delete(eventList, m.id)
	}
}

func (m *SignalHandler) HandlerID() uint64 {
	return uint64(m.handlerID)
}

func (m *SignalHandler) ID() int {
	return int(m.id)
}

func registerSignal(widget *C.GtkWidget, signal string) *SignalHandler {
	cb := C.GCallback(C.go_on_clicked)
	name := C.CString(signal)
	defer C.free(unsafe.Pointer(name))
	pointer := C.gpointer(widget)
	handlerId := C.g_signal_connect_data(pointer, name, cb, pointer, nil, 0)
	return &SignalHandler{
		widget:    widget,
		handlerID: handlerId,
		id:        uintptr(unsafe.Pointer(pointer)),
	}
}

func registerAction(widget IWidget, signal string, cb *Callback) *SignalHandler {
	cWidget := widget.toWidget()
	sh := registerSignal(cWidget, signal)
	RegisterEvent(sh.id, cb)
	return sh
}
