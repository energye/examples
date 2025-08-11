package window

import (
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/tool"
	"github.com/energye/lcl/types"
)

func (m *BrowserWindow) Minimize() {
	m.SetWindowState(types.WsMinimized)
}

func (m *BrowserWindow) Maximize() {
	if m.WindowState() == types.WsNormal {
		m.SetWindowState(types.WsMaximized)
	} else {
		m.SetWindowState(types.WsNormal)
		if tool.IsDarwin() { //要这样重复设置2次不然不启作用
			m.SetWindowState(types.WsMaximized)
			m.SetWindowState(types.WsNormal)
		}
	}
}

func (m *BrowserWindow) FullScreen() {
	if m.WindowState() == types.WsMinimized || m.WindowState() == types.WsMaximized {

	}
	m.windowState = types.WsFullScreen
	m.previousWindowPlacement = m.BoundsRect()
	//monitorRect := m.Monitor().BoundsRect()
}

func (m *BrowserWindow) ExitFullScreen() {
	if m.IsFullScreen() {
		m.windowState = types.WsNormal
		m.SetWindowState(types.WsNormal)
		m.SetBoundsRect(m.previousWindowPlacement)
	}
}

func (m *BrowserWindow) IsFullScreen() bool {
	if tool.IsDarwin() {
		return m.windowState == types.WsFullScreen && m.WindowState() == types.WsFullScreen
	}
	return m.windowState == types.WsFullScreen
}

func (m *BrowserWindow) boxMouseMove(sender lcl.IObject, shift types.TShiftState, x, y int32) {

}

func (m *BrowserWindow) boxMouseDown(sender lcl.IObject, button types.TMouseButton, shift types.TShiftState, x, y int32) {

}
