//go:build !cgo

package src

import (
	"github.com/energye/energy/v3/platform/darwin/cocoa/nocgo/notification"
)

func (m *TMainForm) CreateNotify() {
	m.notifService = notification.New()
}
