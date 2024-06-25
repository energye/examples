package v8context

import (
	"bytes"
	"fmt"
	"github.com/energye/cef/cef"
	"github.com/energye/lcl/process"
	"strconv"
	"unsafe"
)

func Context(app cef.ICefApplication) {
	fmt.Println("ProcessType:", process.Args.ProcessType())
	var (
		ipc         cef.ICefV8Value
		onHandler   cef.IV8Handler
		emitHandler cef.IV8Handler
		onCallback  cef.ICefV8Value
	)
	app.SetOnContextCreated(func(browser cef.ICefBrowser, frame cef.ICefFrame, context cef.ICefV8Context) {
		onHandler = cef.NewV8Handler()
		onHandler.SetOnExecute(func(name string, object cef.ICefV8Value, arguments cef.ICefV8ValueArray) (retVal cef.ICefV8Value, exception string, result bool) {
			fmt.Println("ipc.on Execute name:", name)
			// JS事件名
			//lName := arguments.Get(0)
			// JS事件回调函数
			callFN := arguments.Get(1)
			onCallback = cef.V8ValueRef.UnWrap(callFN.Wrap())
			callFN.Free()
			arguments.Free()
			return
		})
		emitHandler = cef.NewV8Handler()
		emitHandler.SetOnExecute(func(name string, object cef.ICefV8Value, arguments cef.ICefV8ValueArray) (retVal cef.ICefV8Value, exception string, result bool) {
			fmt.Println("ipc.emit Execute name:", name)
			v8ctx := cef.V8ContextRef.Current()
			frame := v8ctx.GetFrame()
			fmt.Println("frameId:", frame.GetIdentifier(), "ProcessType:", process.Args.ProcessType())
			// 发送消息
			var buf bytes.Buffer
			for i := 0; i < arguments.Size(); i++ {
				val := arguments.Get(i)
				if val.IsString() {
					buf.WriteString(val.GetStringValue())
				} else if val.IsInt() {
					buf.WriteString(strconv.Itoa(int(val.GetIntValue())))
				} else if val.IsArray() {
					lenh := int(val.GetArrayLength())
					for i := 0; i < lenh; i++ {
						arg := val.GetValueByIndex(int32(i))
						if arg.IsString() {
							buf.WriteString(arg.GetStringValue())
						} else if arg.IsInt() {
							buf.WriteString(strconv.Itoa(int(arg.GetIntValue())))
						} else if arg.IsUInt() {
							buf.WriteString(strconv.Itoa(int(arg.GetUIntValue())))
						} else if arg.IsDouble() {
							buf.WriteString(fmt.Sprintf("%v", arg.GetDoubleValue()))
						}
					}
				}
			}
			dataBytes := buf.Bytes()
			processMessage := cef.ProcessMessageRef.New("send-browser")
			messageArgumentList := processMessage.GetArgumentList()
			dataBin := cef.BinaryValueRef.New(uintptr(unsafe.Pointer(&dataBytes[0])), uint32(len(dataBytes)))
			messageArgumentList.SetBinary(0, dataBin)
			frame.SendProcessMessage(cef.PID_RENDERER, processMessage)
			messageArgumentList.Clear()
			arguments.Free()
			return
		})
		ipc = cef.V8ValueRef.NewObject(nil, nil)
		onFunc := cef.V8ValueRef.NewFunction("on", onHandler.AsInterface())
		ipc.SetValueByKey("on", onFunc, cef.V8_PROPERTY_ATTRIBUTE_READONLY)
		emitFunc := cef.V8ValueRef.NewFunction("emit", emitHandler.AsInterface())
		ipc.SetValueByKey("emit", emitFunc, cef.V8_PROPERTY_ATTRIBUTE_READONLY)
		context.GetGlobal().SetValueByKey("ipc", ipc, cef.V8_PROPERTY_ATTRIBUTE_READONLY)
	})
	app.SetOnProcessMessageReceived(func(browser cef.ICefBrowser, frame cef.ICefFrame, sourceProcess cef.TCefProcessId, message cef.ICefProcessMessage, outResult *bool) {
		fmt.Println("渲染进程 name:", message.GetName())
		binArgs := message.GetArgumentList().GetBinary(0)
		fmt.Println("size:", binArgs.GetSize())
		messageDataBytes := make([]byte, int(binArgs.GetSize()))
		binArgs.GetData(uintptr(unsafe.Pointer(&messageDataBytes[0])), binArgs.GetSize(), 0)
		fmt.Println("data:", string(messageDataBytes))
		// 调用JS回调函数
		onCallback.ExecuteFunctionWithContext(frame.GetV8Context(), nil, nil)
	})
}
