package cocoa

/*
#cgo CFLAGS: -mmacosx-version-min=11.0 -x objective-c
#cgo LDFLAGS: -mmacosx-version-min=11.0 -framework Cocoa
#include "cocoa.h"

extern void onRunOnMainThread(long id);

*/
import "C"
import (
	"sync"
	"unsafe"
)

//export onRunOnMainThread
func onRunOnMainThread(id C.long) {
	doRunOnMainThread(int(id))
}

type runOnMainThreadFn func()

var (
	callbackFuncList     = make(map[int]runOnMainThreadFn)
	callbackFuncListLock = sync.Mutex{}
	isRROMTC             bool
)

func doRunOnMainThread(id int) {
	fn, ok := callbackFuncList[id]
	if ok {
		delete(callbackFuncList, id)
		fn()
	}
}

func RegisterRunOnMainThreadCallback() {
	if isRROMTC {
		return
	}
	isRROMTC = true
	C.RegisterRunOnMainThreadCallback(C.RunOnMainThreadCallback(C.onRunOnMainThread))
}

func RunOnMainThread(fn runOnMainThreadFn) {
	callbackFuncListLock.Lock()
	defer callbackFuncListLock.Unlock()
	id := int(uintptr(unsafe.Pointer(&fn)))
	callbackFuncList[id] = fn
	C.ExecuteRunOnMainThread(C.long(id))
}
