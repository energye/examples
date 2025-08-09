package window

import (
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/pkgs/win"
	"github.com/energye/lcl/types"
	"github.com/energye/lcl/types/messages"
	"syscall"
	"unsafe"
)

func (m *BrowserWindow) wndProc(hwnd types.HWND, message uint32, wParam, lParam uintptr) uintptr {
	switch message {
	case messages.WM_DPICHANGED:
		if !lcl.Application.Scaled() {
			newWindowSize := (*types.TRect)(unsafe.Pointer(lParam))
			win.SetWindowPos(m.Handle(), uintptr(0),
				newWindowSize.Left, newWindowSize.Top, newWindowSize.Right-newWindowSize.Left, newWindowSize.Bottom-newWindowSize.Top,
				win.SWP_NOZORDER|win.SWP_NOACTIVATE)
		}
		return 0 // 确保处理WM_DPICHANGED后返回

	case messages.WM_ACTIVATE:
		win.ExtendFrameIntoClientArea(m.Handle(), win.Margins{CxLeftWidth: 1, CxRightWidth: 1, CyTopHeight: 1, CyBottomHeight: 1})
		return 0

	case messages.WM_NCCALCSIZE:
		if wParam != 0 {
			isMaximize := uint32(win.GetWindowLong(m.Handle(), win.GWL_STYLE))&win.WS_MAXIMIZE != 0
			if isMaximize {
				rect := (*types.TRect)(unsafe.Pointer(lParam))
				monitor := win.MonitorFromRect(rect, win.MONITOR_DEFAULTTONULL)
				if monitor != 0 {
					var monitorInfo types.TMonitorInfo
					monitorInfo.CbSize = types.DWORD(unsafe.Sizeof(monitorInfo))
					if win.GetMonitorInfo(monitor, &monitorInfo) {
						*rect = monitorInfo.RcWork
					}
				}
			}
			return 0 // 移除标准边框
		}

		//case messages.WM_NCHITTEST: // 新增：处理鼠标命中测试
		//	x := int32(lParam & 0xFFFF)
		//	y := int32(lParam >> 16)
		//	var rect types.TRect
		//	win.GetWindowRect(m.Handle(), &rect)
		//
		//	borderWidth := int32(5) // 边缘检测宽度
		//	left := x - rect.Left
		//	right := rect.Right - x
		//	top := y - rect.Top
		//	bottom := rect.Bottom - y
		//
		//	// 检测角落区域
		//	if left < borderWidth && top < borderWidth {
		//		return messages.HTTOPLEFT
		//	} else if right < borderWidth && top < borderWidth {
		//		return messages.HTTOPRIGHT
		//	} else if left < borderWidth && bottom < borderWidth {
		//		return messages.HTBOTTOMLEFT
		//	} else if right < borderWidth && bottom < borderWidth {
		//		return messages.HTBOTTOMRIGHT
		//	}
		//
		//	// 检测边缘区域
		//	if left < borderWidth {
		//		return messages.HTLEFT
		//	} else if right < borderWidth {
		//		return messages.HTRIGHT
		//	} else if top < borderWidth {
		//		return messages.HTTOP
		//	} else if bottom < borderWidth {
		//		return messages.HTBOTTOM
		//	}
		//
		//	// 检测标题栏区域（假设标题栏高度为30）
		//	titleBarHeight := int32(30)
		//	if top < titleBarHeight {
		//		return messages.HTCAPTION // 允许拖动窗口
		//	}
	}

	return win.CallWindowProc(m.oldWndPrc, uintptr(hwnd), message, wParam, lParam)
}

func (m *BrowserWindow) HookWndProcMessage() {
	wndProcCallback := syscall.NewCallback(m.wndProc)
	m.oldWndPrc = win.SetWindowLongPtr(m.Handle(), win.GWL_WNDPROC, wndProcCallback)
	// trigger WM_NCCALCSIZE
	// https://learn.microsoft.com/en-us/windows/win32/dwm/customframe#removing-the-standard-frame
	clientRect := m.BoundsRect()
	win.SetWindowPos(m.Handle(), 0, clientRect.Left, clientRect.Top, clientRect.Right-clientRect.Left, clientRect.Bottom-clientRect.Top, win.SWP_FRAMECHANGED|win.SWP_NOACTIVATE)
}
