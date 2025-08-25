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
	GoArgsType_None    = OCGoArgumentsType(C.ArgsType_None)    // 未使用类型
	GoArgsType_Int     = OCGoArgumentsType(C.ArgsType_Int)     // 基础类型 int
	GoArgsType_Float   = OCGoArgumentsType(C.ArgsType_Float)   // 基础类型 float64
	GoArgsType_Bool    = OCGoArgumentsType(C.ArgsType_Bool)    // 基础类型 bool
	GoArgsType_String  = OCGoArgumentsType(C.ArgsType_String)  // 基础类型 string
	GoArgsType_Object  = OCGoArgumentsType(C.ArgsType_Object)  // 对象类型 NS 里创建的对象指针 (void*)[obj retain]
	GoArgsType_Pointer = OCGoArgumentsType(C.ArgsType_Pointer) // 指针类型 NS 里创建的 [NSValue valueWithPointer:customData]
)

type OCGoArgsItem struct {
	item Pointer
}

type OCGoArguments struct {
	arguments Pointer
	count     int
}

type GoArgsItem struct {
	Value Pointer
	Type  OCGoArgumentsType
}

type GoArguments struct {
	Items []any
}

func (m *GoArguments) ToOC() *C.GoArguments {
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

	goArgs := (*C.GoArguments)(C.malloc(C.sizeof_GoArguments))
	goArgs.Count = C.int(len(m.Items))
	goArgs.Items = (*C.GoArgsItem)(C.malloc(C.size_t(goArgs.Count) * C.sizeof_GoArgsItem))
	// 64
	items := (*[1 << 6]*C.GoArgsItem)(Pointer(goArgs.Items))[:goArgs.Count:goArgs.Count]
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

func (m *OCGoArguments) GetItem(index int) *OCGoArgsItem {
	item := C.GetItemFromGoArguments((*C.GoArguments)(m.arguments), C.int(index))
	return &OCGoArgsItem{item: Pointer(item)}
}

func (m *OCGoArgsItem) Type() OCGoArgumentsType {
	item := (*C.GoArgsItem)(m.item)
	return OCGoArgumentsType(item.Type)
}

func (m *OCGoArgsItem) Value() Pointer {
	item := (*C.GoArgsItem)(m.item)
	return Pointer(item.Value)
}

func (m *OCGoArgsItem) IntValue() int {
	if m.Type() == GoArgsType_Int {
		item := (*C.GoArgsItem)(m.item)
		return int(*(*C.int)(item.Value))
	}
	return 0
}

func (m *OCGoArguments) GetInt(index int) int {
	item := C.GetItemFromGoArguments((*C.GoArguments)(m.arguments), C.int(index))
	if item == nil || item.Type != GoArgsType_Int {
		return 0
	}
	return int(*(*C.int)(item.Value))
}

func (m *OCGoArguments) GetFloat(index int) float64 {
	item := C.GetItemFromGoArguments((*C.GoArguments)(m.arguments), C.int(index))
	if item == nil || item.Type != GoArgsType_Float {
		return 0
	}
	return float64(*(*C.double)(item.Value))
}

func (m *OCGoArguments) GetBool(index int) bool {
	item := C.GetItemFromGoArguments((*C.GoArguments)(m.arguments), C.int(index))
	if item == nil || item.Type != GoArgsType_Bool {
		return false
	}
	return bool(*(*C.bool)(item.Value))
}

func (m *OCGoArguments) GetString(index int) string {
	item := C.GetItemFromGoArguments((*C.GoArguments)(m.arguments), C.int(index))
	if item == nil || item.Type != GoArgsType_String {
		return ""
	}
	return C.GoString((*C.char)(item.Value))
}

func (m *OCGoArguments) GetPointer(index int) Pointer {
	item := C.GetItemFromGoArguments((*C.GoArguments)(m.arguments), C.int(index))
	if item == nil || item.Type != GoArgsType_Pointer {
		return nil
	}
	return Pointer(item.Value)
}

func (m *OCGoArguments) GetObject(index int) Pointer {
	item := C.GetItemFromGoArguments((*C.GoArguments)(m.arguments), C.int(index))
	if item == nil || item.Type != GoArgsType_Object {
		return nil
	}
	return Pointer(item.Value)
}

// FreeGoArguments 调用 ObjC 侧的释放函数（在 Go 侧触发释放）
func FreeGoArguments(args *C.GoArguments) {
	if args == nil {
		return
	}
	C.FreeGoArguments(args)
}
