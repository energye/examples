package gtkhelper

/*
#cgo pkg-config: gtk+-3.0
#include <gtk/gtk.h>
*/
import "C"
import "unsafe"

type EventSignalName = string

const (
	EsnClicked       EventSignalName = "clicked"
	EsnChanged       EventSignalName = "changed"
	EsnActivate      EventSignalName = "activate"
	EsnKeyPressEvent EventSignalName = "key-press-event"
)

// TccType 事件类型, 用于区分普通通知事件, 还是特殊事件
type TccType = int

const (
	TCCNotify TccType = iota
	TCCClicked
	TCCTextDidChange
	TCCTextDidEndEditing
	TCCSelectionChanged
	TCCSelectionDidChange
)

type TNotifyEvent func(sender *Widget)
type TTextChangedEvent func(sender *Widget, text string)
type TTextCommitEvent func(sender *Widget, text string)
type TTextKeyPressEvent func(sender *Widget, key *EventKey) bool

type CallbackContext struct {
	widget unsafe.Pointer
	input  any
	result any
}

type Callback struct {
	type_ TccType
	cb    func(ctx *CallbackContext)
}

func MakeNotifyEvent(cb TNotifyEvent) *Callback {
	return &Callback{
		type_: TCCNotify,
		cb: func(ctx *CallbackContext) {
			cb(wrapWidget(ToGoObject(ctx.widget)))
		},
	}
}

func MakeTextChangedEvent(cb TTextChangedEvent) *Callback {
	return &Callback{
		type_: TCCTextDidChange,
		cb: func(ctx *CallbackContext) {
			text := C.gtk_entry_get_text((*C.GtkEntry)(ctx.widget))
			cb(wrapWidget(ToGoObject(ctx.widget)), C.GoString(text))
		},
	}
}

func MakeTextCommitEvent(cb TTextCommitEvent) *Callback {
	return &Callback{
		type_: TCCTextDidEndEditing,
		cb: func(ctx *CallbackContext) {
			text := C.gtk_entry_get_text((*C.GtkEntry)(ctx.widget))
			cb(wrapWidget(ToGoObject(ctx.widget)), C.GoString(text))
		},
	}
}

func MakeTextKeyPressEvent(cb TTextKeyPressEvent) *Callback {
	return &Callback{
		type_: TCCTextDidEndEditing,
		cb: func(ctx *CallbackContext) {
			keyPtr := ctx.input.(unsafe.Pointer)
			key := ToKeyEvent(keyPtr)
			result := cb(wrapWidget(ToGoObject(ctx.widget)), key)
			ctx.result = result
		},
	}
}
