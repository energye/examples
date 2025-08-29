package gtkhelper

/*
#cgo pkg-config: gtk+-3.0
#include <gtk/gtk.h>
*/
import "C"
import "unsafe"

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

type CallbackContext struct {
	widget unsafe.Pointer
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
