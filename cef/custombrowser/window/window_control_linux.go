package window

import (
	"github.com/energye/lcl/api"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/tool"
	"github.com/energye/lcl/types"
)

func (m *BrowserWindow) Minimize() {
	m.SetWindowState(types.WsMinimized)
}

func (m *BrowserWindow) Maximize() {
	if m.WindowState() == types.WsNormal {
		m.normalBounds = m.BoundsRect()
		m.SetWindowState(types.WsMaximized)
		workAreaRect := lcl.Screen.WorkAreaRect()
		m.SetBoundsRect(workAreaRect)
	} else if m.WindowState() == types.WsMaximized {
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
	lcl.RunOnMainThreadAsync(func(id uint32) {
		m.Maximize()
	})
}

func (m *BrowserWindow) boxMouseMove(sender lcl.IObject, shift types.TShiftState, x, y int32) {
	lcl.Screen.SetCursor(types.CrDefault)
	// 判断鼠标所在区域
	rect := m.BoundsRect()
	if x > m.borderWidth && y > m.borderWidth && x < rect.Width()-m.borderWidth && y < rect.Height()-m.borderWidth && y < m.titleHeight {
		// 标题栏部分
		m.borderHT = 0 // 重置边框标记
		m.isTitleBar = true
		if m.isDown {
			m.isDown = false
			if m.WindowState() == types.WsMaximized {
				// 拖拽时 最大化状态重新计算窗口 Rect
				m.SetWindowState(types.WsNormal)
				workAreaRect := lcl.Screen.WorkAreaRect()
				curPos := lcl.Mouse.CursorPos()
				rect := m.normalBounds
				rect.Top = workAreaRect.Top
				rect.Left = curPos.X - (rect.Width() / 2)
				if rect.Left < workAreaRect.Left {
					rect.Left = workAreaRect.Left
				}
				rect.SetWidth(m.normalBounds.Width())
				rect.SetHeight(m.normalBounds.Height())
				m.SetBoundsRect(rect)
			}
			lcl.RunOnMainThreadAsync(func(id uint32) {
				lcl.DragWindow(m.Handle(), m.Left(), m.Top(), 1, api.GDK_WINDOW_EDGE_NORTH_WEST)
				lcl.Mouse.SetCapture(m.Handle())
			})
		}
	} else {
		m.isTitleBar = false
		// 边框区域判断 (8个区域)
		switch {
		// 角落区域 (优先判断)
		case x < m.borderWidth && y < m.borderWidth:
			m.borderHT = uintptr(api.GDK_WINDOW_EDGE_NORTH_WEST) // messages.HTTOPLEFT
			lcl.Screen.SetCursor(types.CrSizeNWSE)
		case x > rect.Width()-m.borderWidth && y < m.borderWidth:
			m.borderHT = uintptr(api.GDK_WINDOW_EDGE_NORTH_EAST) // messages.HTTOPRIGHT
			lcl.Screen.SetCursor(types.CrSizeNESW)
		case x < m.borderWidth && y > rect.Height()-m.borderWidth:
			m.borderHT = uintptr(api.GDK_WINDOW_EDGE_SOUTH_WEST) // messages.HTBOTTOMLEFT
			lcl.Screen.SetCursor(types.CrSizeNESW)
		case x > rect.Width()-m.borderWidth && y > rect.Height()-m.borderWidth:
			m.borderHT = uintptr(api.GDK_WINDOW_EDGE_SOUTH_EAST) // messages.HTBOTTOMRIGHT
			lcl.Screen.SetCursor(types.CrSizeNWSE)
		// 边缘区域
		case y < m.borderWidth:
			m.borderHT = uintptr(api.GDK_WINDOW_EDGE_NORTH) // messages.HTTOP
			lcl.Screen.SetCursor(types.CrSizeNS)
		case y > rect.Height()-m.borderWidth:
			m.borderHT = uintptr(api.GDK_WINDOW_EDGE_SOUTH) // messages.HTBOTTOM
			lcl.Screen.SetCursor(types.CrSizeNS)
		case x < m.borderWidth:
			m.borderHT = uintptr(api.GDK_WINDOW_EDGE_WEST) // messages.HTLEFT
			lcl.Screen.SetCursor(types.CrSizeWE)
		case x > rect.Width()-m.borderWidth:
			m.borderHT = uintptr(api.GDK_WINDOW_EDGE_EAST) // messages.HTRIGHT
			lcl.Screen.SetCursor(types.CrSizeWE)
		default:
			m.borderHT = 0 // 客户区
		}
	}
}

func (m *BrowserWindow) boxMouseDown(sender lcl.IObject, button types.TMouseButton, shift types.TShiftState, x, y int32) {
	if button == types.MbLeft {
		m.isDown = true
		if m.borderHT != 0 {
			lcl.DragWindow(m.Handle(), m.Left(), m.Top(), 1, api.TGdkWindowEdge(m.borderHT))
			lcl.Mouse.SetCapture(m.Handle())
		}
	}
}
