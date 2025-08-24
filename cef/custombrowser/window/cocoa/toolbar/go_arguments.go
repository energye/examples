package toolbar

/*
#cgo CFLAGS: -mmacosx-version-min=11.0 -x objective-c
#cgo LDFLAGS: -mmacosx-version-min=11.0 -framework Cocoa
#include "go_arguments.h"

*/
import "C"
import (
	"fmt"
	"os"
)

// Go 端数据类型映射
const (
	GoArgsType_None    = C.ArgsType_None
	GoArgsType_Int     = C.ArgsType_Int
	GoArgsType_Float   = C.ArgsType_Float
	GoArgsType_Bool    = C.ArgsType_Bool
	GoArgsType_String  = C.ArgsType_String
	GoArgsType_Object  = C.ArgsType_Object
	GoArgsType_Pointer = C.ArgsType_Pointer
)

type OCGoArguments = C.GoArguments
type OCGoArgsItem = C.GoArgsItem
type OCGoArgumentsType = C.GoArgumentsType

type GoArguments struct {
	Items []any
}

func (m *GoArguments) Free() {

}

func (m *GoArguments) ToOC() *OCGoArguments {
	if len(m.Items) == 0 {
		return nil
	}
	goArgs := (*OCGoArguments)(C.malloc(C.sizeof_GoArguments))
	goArgs.Count = C.int(len(m.Items))
	goArgs.Items = (*OCGoArgsItem)(C.malloc(C.size_t(goArgs.Count) * C.sizeof_GoArgsItem))
	items := (*[1 << 6]*OCGoArgsItem)(Pointer(goArgs.Items))[:goArgs.Count:goArgs.Count]
	for i, arg := range m.Items {
		item := items[i] // 直接访问数组元素
		switch v := arg.(type) {
		case int:
			item.Type = GoArgsType_Int
			// 为 int 值分配 C 堆内存（确保 ObjC 侧可释放）
			val := (*C.int)(C.malloc(C.sizeof_int))
			*val = C.int(v)
			item.Value = Pointer(val)
		case float32:
			item.Type = GoArgsType_Float
			val := (*C.float)(C.malloc(C.sizeof_float))
			*val = C.float(v)
			item.Value = Pointer(val)
		case bool:
			item.Type = GoArgsType_Bool
			val := (*C.bool)(C.malloc(C.sizeof_bool))
			*val = C.bool(v)
			item.Value = Pointer(val)
		case string:
			item.Type = GoArgsType_String
			// C.CString 内部调用 malloc 分配内存，ObjC 侧可用 free 释放
			item.Value = Pointer(C.CString(v))
		case uintptr:
			item.Type = GoArgsType_Pointer
			// 直接传递指针（由外部管理生命周期）
			item.Value = Pointer(v)
		default:
			fmt.Println("[ERROR] CreateGoArguments 不支持的类型参数 index:", i, "value:", arg)
			os.Exit(1)
		}
	}
	return goArgs
}

func GetFromGoArguments(data *OCGoArguments, index int, expectedType int) Pointer {
	if data == nil || index < 0 || index >= int(data.Count) {
		return nil
	}
	return C.GetFromGoArguments(data, C.int(index), OCGoArgumentsType(expectedType))
}

func GetIntFromGoArguments(data *OCGoArguments, index int) int {
	ptr := GetFromGoArguments(data, index, GoArgsType_Int)
	if ptr == nil {
		return 0
	}
	return int(*(*C.int)(ptr))
}

func GetFloatFromGoArguments(data *OCGoArguments, index int) float64 {
	ptr := GetFromGoArguments(data, index, GoArgsType_Float)
	if ptr == nil {
		return 0.0
	}
	return float64(*(*C.double)(ptr))
}

func GetBoolFromGoArguments(data *OCGoArguments, index int) bool {
	ptr := GetFromGoArguments(data, index, GoArgsType_Bool)
	if ptr == nil {
		return false
	}
	return bool(*(*C.bool)(ptr))
}

func GetStringFromGoArguments(data *OCGoArguments, index int) string {
	ptr := GetFromGoArguments(data, index, GoArgsType_String)
	if ptr == nil {
		return ""
	}
	return C.GoString((*C.char)(ptr))
}

func GetPointerFromGoArguments(data *OCGoArguments, index int) Pointer {
	return GetFromGoArguments(data, index, GoArgsType_Pointer)
}

func GetObjectFromGoArguments(data *OCGoArguments, index int) Pointer {
	return GetFromGoArguments(data, index, GoArgsType_Object)
}

// FreeGoArguments 调用 ObjC 侧的释放函数（在 Go 侧触发释放）
func FreeGoArguments(args *OCGoArguments) {
	if args == nil {
		return
	}
	C.FreeGoArguments(args)
}
