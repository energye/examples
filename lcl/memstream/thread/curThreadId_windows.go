package thread

import "github.com/energye/lcl/pkgs/win"

func GetCurrentThreadId() uintptr {
	return win.GetCurrentThreadId()
}
