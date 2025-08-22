package dock

/*
#cgo CFLAGS: -mmacosx-version-min=11.0 -x objective-c
#cgo LDFLAGS: -mmacosx-version-min=11.0 -framework Cocoa
#import <Cocoa/Cocoa.h>

static void Hide() {
  [[NSApplication sharedApplication] setActivationPolicy:NSApplicationActivationPolicyProhibited];
}

static void Show() {
  [[NSApplication sharedApplication] setActivationPolicy:NSApplicationActivationPolicyRegular];
}
*/
import "C"

func Hide() {
	C.Hide()
}

func Show() {
	C.Show()
}
