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
	if m.windowState == types.WsNormal {
		m.normalBounds = m.BoundsRect()
		m.windowState = types.WsMaximized
		m.SetWindowState(types.WsMaximized)
		workAreaRect := lcl.Screen.WorkAreaRect()
		m.SetBoundsRect(workAreaRect)
	} else if m.windowState == types.WsMaximized {
		m.windowState = types.WsNormal
		m.SetWindowState(types.WsNormal)
		m.SetBoundsRect(m.normalBounds)
	}
}

func (m *BrowserWindow) FullScreen() {
	m.windowState = types.WsFullScreen
}

func (m *BrowserWindow) ExitFullScreen() {
	if m.IsFullScreen() {
		m.windowState = types.WsNormal
		m.SetWindowState(types.WsNormal)
		m.SetBoundsRect(m.normalBounds)
	}
}

func (m *BrowserWindow) IsFullScreen() bool {
	if tool.IsDarwin() {
		return m.windowState == types.WsFullScreen && m.WindowState() == types.WsFullScreen
	}
	return m.windowState == types.WsFullScreen
}

func (m *BrowserWindow) boxDblClick(sender lcl.IObject) {

}

func (m *BrowserWindow) boxMouseMove(sender lcl.IObject, shift types.TShiftState, x, y int32) {
}

func (m *BrowserWindow) boxMouseDown(sender lcl.IObject, button types.TMouseButton, shift types.TShiftState, x, y int32) {

}
