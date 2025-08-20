package window

/*
#cgo CFLAGS: -mmacosx-version-min=11.0 -x objective-c
#cgo LDFLAGS: -mmacosx-version-min=11.0 -framework Cocoa

#import <Cocoa/Cocoa.h>

// 设置窗口标题栏透明
void setTitlebarTransparent(void* window) {
    NSWindow *nsWindow = (NSWindow*)window;
    nsWindow.titlebarAppearsTransparent = YES;
    nsWindow.backgroundColor = [NSColor clearColor];
}

void setFullSizeContentView(void* window) {
    NSWindow *nsWindow = (NSWindow*)window;
    // 启用全尺寸内容视图，让客户区覆盖标题栏
    nsWindow.styleMask |= NSWindowStyleMaskFullSizeContentView;
}

*/
import "C"
import (
	"unsafe"
)

type NSButton = unsafe.Pointer

type TrafficLightButtons struct {
	CloseButton    NSButton
	MinimizeButton NSButton
	ZoomButton     NSButton
}

type TrafficLightRect struct {
	X, Y, Width, Height float32
}

func (m *Window) HookWndProcMessage() {
	//nsWindow := unsafe.Pointer(lcl.PlatformWindow(m.Instance()))
	//
	//C.setFullSizeContentView(nsWindow) // 设置客户区填充整个窗口（包括标题栏）
	//C.setTitlebarTransparent(nsWindow) // 设置窗口标题栏透明
	//
	//edit := lcl.NewEdit(m)
	//edit.SetParent(m)
	//edit.SetLeft(400)
}
