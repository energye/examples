package window

import (
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
)

type Window struct {
	lcl.TEngForm
	oldWndPrc    uintptr
	dx, dy       int32 // down x,y
	mx, my       int32 // move x,y
	wx, wy       int32 // window point
	normalBounds types.TRect
	windowState  types.TWindowState
}

func (m *Window) FormAfterCreate(sender lcl.IObject) {
	m.HookWndProcMessage()
}
