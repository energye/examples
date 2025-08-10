package window

import (
	"github.com/energye/lcl/pkgs/win"
	"github.com/energye/lcl/tool"
	"github.com/energye/lcl/types"
	"github.com/energye/lcl/types/messages"
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
		if win.ReleaseCapture() {
			win.SendMessage(m.Handle(), messages.WM_SYSCOMMAND, messages.SC_RESTORE, 0)
		}
	}
	m.windowState = types.WsFullScreen
	m.previousWindowPlacement = m.BoundsRect()
	monitorRect := m.Monitor().BoundsRect()
	win.SetWindowPos(m.Handle(), win.HWND_TOP, monitorRect.Left, monitorRect.Top, monitorRect.Width(), monitorRect.Height(), win.SWP_NOOWNERZORDER|win.SWP_FRAMECHANGED)
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
