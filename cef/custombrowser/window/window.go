package window

import "github.com/energye/lcl/lcl"

type Window struct {
	lcl.TEngForm
	oldWndPrc uintptr
}

func (m *Window) FormAfterCreate(sender lcl.IObject) {
	m.HookWndProcMessage()
}
