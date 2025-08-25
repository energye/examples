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

type OCGoArgumentsType = C.GoArgumentsType

const (
	GoArgsType_None    = C.ArgsType_None    // 未使用类型
	GoArgsType_Int     = C.ArgsType_Int     // 基础类型 int
	GoArgsType_Float   = C.ArgsType_Float   // 基础类型 float64
	GoArgsType_Bool    = C.ArgsType_Bool    // 基础类型 bool
	GoArgsType_String  = C.ArgsType_String  // 基础类型 string
	GoArgsType_Object  = C.ArgsType_Object  // 对象类型 NS 里创建的对象指针 (void*)[obj retain]
	GoArgsType_Pointer = C.ArgsType_Pointer // 指针类型 NS 里创建的 [NSValue valueWithPointer:customData]
)

type OCGoArguments = C.GoArguments
type OCGoArgsItem = C.GoArgsItem

type GoArgsItem struct {
	Value Pointer
	Type  OCGoArgumentsType
}

type GoArguments struct {
	Items []any
}

func (m *GoArguments) Free() {

}

func (m *GoArguments) ToOC() *OCGoArguments {
	if len(m.Items) == 0 {
		return nil
	}
	toInt := func(value any) int {
		switch v := value.(type) {
		case int:
			return v
		case int8:
			return int(v)
		case int16:
			return int(v)
		case int32:
			return int(v)
		case int64:
			return int(v)
		case uint:
			return int(v)
		case uint8:
			return int(v)
		case uint16:
			return int(v)
		case uint32:
			return int(v)
		case uint64:
			return int(v)
		default:
			return 0
		}
	}

	toDouble := func(value any) float64 {
		switch v := value.(type) {
		case float32:
			return float64(v)
		case float64:
			return v
		default:
			return 0
		}
	}

	goArgs := (*OCGoArguments)(C.malloc(C.sizeof_GoArguments))
	goArgs.Count = C.int(len(m.Items))
	goArgs.Items = (*OCGoArgsItem)(C.malloc(C.size_t(goArgs.Count) * C.sizeof_GoArgsItem))
	items := (*[1 << 6]*OCGoArgsItem)(Pointer(goArgs.Items))[:goArgs.Count:goArgs.Count]
	for i, arg := range m.Items {
		item := items[i] // 直接访问数组元素
		switch v := arg.(type) {
		// malloc 分配内存 确保 ObjC 侧可释放
		case int, int8, uint8, int16, uint16, int32, uint32, int64, uint64:
			item.Type = GoArgsType_Int
			val := (*C.int)(C.malloc(C.sizeof_int))
			*val = C.int(toInt(v))
			item.Value = Pointer(val)
		case float32, float64:
			item.Type = GoArgsType_Float
			val := (*C.double)(C.malloc(C.sizeof_double))
			*val = C.double(toDouble(v))
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
			item.Type = GoArgsType_Object // NS 里创建的对象
			item.Value = Pointer(v)
		default:
			fmt.Println("[ERROR] CreateGoArguments 不支持的类型参数 index:", i, "value:", arg)
			os.Exit(1)
		}
	}
	return goArgs
}

func GetItemFromGoArguments(data *OCGoArguments, index int) *OCGoArgsItem {
	item := C.GetItemFromGoArguments(data, C.int(index))
	return item
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
