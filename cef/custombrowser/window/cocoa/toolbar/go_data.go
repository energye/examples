package toolbar

/*
#cgo CFLAGS: -mmacosx-version-min=11.0 -x objective-c
#cgo LDFLAGS: -mmacosx-version-min=11.0 -framework Cocoa
#include "go_data.h"
*/
import "C"
import "unsafe"

type GoDataType = int

const (
	GDtNone GoDataType = iota
	GDtString
	GDtStringArray
	GDtPointer
)

type StringArray struct {
	Items []string
	Count int
}

type GoData struct {
	Type        GoDataType
	String      string
	StringArray StringArray
	Pointer     Pointer
}

func (m *GoData) ToOC() *C.GoData {
	cData := (*C.GoData)(C.malloc(C.sizeof_GoData))
	if m.Type == GDtString {
		cData.Type = C.DataType_String
		cData.DtString = C.CString(m.String)
	} else if m.Type == GDtStringArray {
		cData.Type = C.DataType_StringArray
		if len(m.StringArray.Items) > 0 {
			items := m.StringArray.Items
			cArray := (**C.char)(C.malloc(C.size_t(len(items)) * C.size_t(unsafe.Sizeof(uintptr(0)))))
			for i, s := range items {
				ptr := (**C.char)(Pointer(uintptr(Pointer(cArray)) + uintptr(i)*unsafe.Sizeof(uintptr(0))))
				*ptr = C.CString(s)
			}
			cData.DtStringArray.Items = cArray
			cData.DtStringArray.Count = C.int(len(items))
		}
	} else if m.Type == GDtStringArray {
		cData.Type = C.DataType_Pointer
		cData.DtPointer = m.Pointer
	} else {
		cData.Type = C.DataType_None
	}
	return cData
}

//export GoFreeGoData
func GoFreeGoData(data *C.GoData) {
	println("[INFO] GoFreeGoData")
	switch data.Type {
	case C.DataType_String:
		if data.DtString != nil {
			C.free(Pointer(data.DtString))
		}

	case C.DataType_StringArray:
		if data.DtStringArray.Items != nil && data.DtStringArray.Count > 0 {
			//items := (*[1 << 20]*C.char)(unsafe.Pointer(data.DtStringArray.Items))[:data.DtStringArray.Count:data.DtStringArray.Count]
			items := unsafe.Slice(data.DtStringArray.Items, data.DtStringArray.Count)
			for i := 0; i < int(data.DtStringArray.Count); i++ {
				if items[i] != nil {
					C.free(Pointer(items[i]))
				}
			}
			C.free(Pointer(data.DtStringArray.Items))
		}
	case C.DataType_Pointer:
		data.DtPointer = nil
	}
	C.free(Pointer(data))
}
