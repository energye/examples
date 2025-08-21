package window

import (
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
)

type Window struct {
	lcl.TEngForm
	oldWndPrc    uintptr
	normalBounds types.TRect
	windowState  types.TWindowState
}

func (m *Window) FormAfterCreate(sender lcl.IObject) {
	//m.HookWndProcMessage()
}
