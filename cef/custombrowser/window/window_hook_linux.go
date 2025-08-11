package window

import "github.com/energye/lcl/types"

func (m *Window) HookWndProcMessage() {
	m.SetBorderStyleToFormBorderStyle(types.BsNone)
}
