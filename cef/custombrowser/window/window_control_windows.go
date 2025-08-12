package window

import (
	"github.com/energye/lcl/lcl"
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
	m.normalBounds = m.BoundsRect()
	monitorRect := m.Monitor().BoundsRect()
	win.SetWindowPos(m.Handle(), win.HWND_TOP, monitorRect.Left, monitorRect.Top, monitorRect.Width(), monitorRect.Height(), win.SWP_NOOWNERZORDER|win.SWP_FRAMECHANGED)
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
	if m.isTitleBar {
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
}

func (m *BrowserWindow) boxMouseMove(sender lcl.IObject, shift types.TShiftState, x, y int32) {
	lcl.Screen.SetCursor(types.CrDefault)
	// 判断鼠标所在区域
	rect := m.BoundsRect()
	if x > m.borderWidth && y > m.borderWidth && x < rect.Width()-m.borderWidth && y < rect.Height()-m.borderWidth && y < m.titleHeight {
		// 标题栏部分
		if m.isDown {
			if win.ReleaseCapture() {
				win.PostMessage(m.Handle(), messages.WM_NCLBUTTONDOWN, messages.HTCAPTION, 0)
			}
		}
		m.borderHT = 0 // 重置边框标记
		m.isTitleBar = true
	} else {
		m.isTitleBar = false
		// 边框区域判断 (8个区域)
		switch {
		// 角落区域 (优先判断)
		case x < m.borderWidth && y < m.borderWidth:
			m.borderHT = messages.HTTOPLEFT
			lcl.Screen.SetCursor(types.CrSizeNWSE)
		case x > rect.Width()-m.borderWidth && y < m.borderWidth:
			m.borderHT = messages.HTTOPRIGHT
			lcl.Screen.SetCursor(types.CrSizeNESW)
		case x < m.borderWidth && y > rect.Height()-m.borderWidth:
			m.borderHT = messages.HTBOTTOMLEFT
			lcl.Screen.SetCursor(types.CrSizeNESW)
		case x > rect.Width()-m.borderWidth && y > rect.Height()-m.borderWidth:
			m.borderHT = messages.HTBOTTOMRIGHT
			lcl.Screen.SetCursor(types.CrSizeNWSE)
		// 边缘区域
		case y < m.borderWidth:
			m.borderHT = messages.HTTOP
			lcl.Screen.SetCursor(types.CrSizeNS)
		case y > rect.Height()-m.borderWidth:
			m.borderHT = messages.HTBOTTOM
			lcl.Screen.SetCursor(types.CrSizeNS)
		case x < m.borderWidth:
			m.borderHT = messages.HTLEFT
			lcl.Screen.SetCursor(types.CrSizeWE)
		case x > rect.Width()-m.borderWidth:
			m.borderHT = messages.HTRIGHT
			lcl.Screen.SetCursor(types.CrSizeWE)
		default:
			m.borderHT = 0 // 客户区
		}
	}
}

func (m *BrowserWindow) boxMouseDown(sender lcl.IObject, button types.TMouseButton, shift types.TShiftState, x, y int32) {
	m.isDown = true
	if m.borderHT != 0 {
		if win.ReleaseCapture() {
			win.PostMessage(m.Handle(), messages.WM_NCLBUTTONDOWN, m.borderHT, 0)
		}
	}
}
