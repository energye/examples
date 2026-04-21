//go:build cgo

package src

import (
	"github.com/energye/energy/v3/platform/darwin/cocoa/cgo/notification"
)

func (m *TMainForm) CreateNotify() {
	// 初始化通知服务
	m.notifService = notification.New()
}
