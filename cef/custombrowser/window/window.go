package window

import (
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/tool"
	"github.com/energye/lcl/types"
)

type Window struct {
	lcl.TEngForm
	oldWndPrc    uintptr
	normalBounds types.TRect
	windowState  types.TWindowState
}

func (m *Window) FormAfterCreate(sender lcl.IObject) {
	if !tool.IsDarwin() {
		m.HookWndProcMessage()
	}
}
