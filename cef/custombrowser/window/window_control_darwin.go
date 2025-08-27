package window

import (
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
)

func (m *BrowserWindow) Minimize() {
}

func (m *BrowserWindow) Maximize() {
}

func (m *BrowserWindow) FullScreen() {
}

func (m *BrowserWindow) ExitFullScreen() {
}

func (m *BrowserWindow) IsFullScreen() bool {
	return false
}

func (m *BrowserWindow) boxDblClick(sender lcl.IObject) {
}

func (m *BrowserWindow) boxMouseMove(sender lcl.IObject, shift types.TShiftState, x, y int32) {
}

func (m *BrowserWindow) boxMouseDown(sender lcl.IObject, button types.TMouseButton, shift types.TShiftState, x, y int32) {
}
