package cocoa

/*
#cgo CFLAGS: -mmacosx-version-min=11.0 -x objective-c
#cgo LDFLAGS: -mmacosx-version-min=11.0 -framework Cocoa
#import "cocoa_helpers.h"
*/
import "C"
import (
	"fmt"
	"unsafe"
)

// InspectControl 检查控件类型并打印详细信息
func InspectControl(handle uintptr) {
	if handle == 0 {
		fmt.Println("错误：无效的句柄（空指针）")
		return
	}

	cHandle := unsafe.Pointer(handle)

	// 1. 基本类型信息
	className := C.getObjectClassName(cHandle)
	fmt.Printf("控件类型: %s\n", C.GoString(className))

	// 2. 常见类型检查
	commonTypes := []string{
		"NSButton", "NSTextField", "NSComboBox",
		"NSSlider", "NSWindow", "NSView", "NSObject",
	}
	fmt.Println("\n类型匹配检查:")
	for _, typ := range commonTypes {
		cType := C.CString(typ)
		isMatch := bool(C.isObjectOfClass(cHandle, cType))
		C.free(unsafe.Pointer(cType))
		fmt.Printf("  是 %-12s: %v\n", typ, isMatch)
	}

	// 3. 继承链信息
	fmt.Println("\n继承链（从自身到根类）:")
	chain := C.getObjectInheritanceChain(cHandle)
	defer C.freeInheritanceChain(&chain)

	if chain.count == 0 {
		fmt.Println("  无法获取继承链信息")
		return
	}

	for i := 0; i < int(chain.count); i++ {
		classNames := (*[1 << 20]*C.char)(unsafe.Pointer(chain.classNames))
		className := C.GoString(classNames[i])
		fmt.Printf("  第 %2d 层: %s\n", i+1, className)
	}
}

func main() {
	// 注意：这里使用示例句柄，实际使用时需要替换为真实的控件句柄
	// 真实句柄通常从窗口系统或UI框架（如Lazarus）获取
	var sampleHandle uintptr = 0x12345678 // 替换为实际句柄

	fmt.Println("=== Cocoa 控件类型检测工具 ===")
	InspectControl(sampleHandle)
}
