package window

import "github.com/energye/lcl/types"

func (m *Window) HookWndProcMessage() {
	return
	m.SetBorderStyleToFormBorderStyle(types.BsNone)
}
