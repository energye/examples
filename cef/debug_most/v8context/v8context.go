package v8context

import (
	"bytes"
	"fmt"
	"github.com/energye/cef/cef"
	"github.com/energye/examples/cef/debug_most/domvisitor"
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
			ctxFrame := v8ctx.GetFrame()
			emitName := arguments.Get(0)
			defer func() {
				object.FreeAndNil()
				ctxFrame.FreeAndNil()
				v8ctx.FreeAndNil()
				emitName.FreeAndNil()
				arguments.Free()
			}()
			fmt.Println("frameId:", ctxFrame.GetIdentifier(), "ProcessType:", process.Args.ProcessType())
			eventName := emitName.GetStringValue()
			if eventName == "domVisitor" {
				domvisitor.DomVisitor()
			} else {
				// 发送消息
				var buf bytes.Buffer
				for i := 1; i < arguments.Size(); i++ {
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
							arg.FreeAndNil()
						}
					}
					val.FreeAndNil()
				}
				dataBytes := buf.Bytes()
				SendBrowserMessage(ctxFrame, eventName, dataBytes)
			}
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
		args := message.GetArgumentList()
		binArgs := args.GetBinary(0)
		fmt.Println("size:", binArgs.GetSize())
		messageDataBytes := make([]byte, int(binArgs.GetSize()))
		binArgs.GetData(uintptr(unsafe.Pointer(&messageDataBytes[0])), binArgs.GetSize(), 0)
		fmt.Println("data:", string(messageDataBytes))
		v8ctx := frame.GetV8Context()
		defer binArgs.FreeAndNil()
		defer args.FreeAndNil()
		defer message.FreeAndNil()
		defer v8ctx.FreeAndNil()
		// 获取当前frame v8context
		// 进入上下文
		if v8ctx.Enter() {
			// 调用JS回调函数
			callFuncArgs := make([]cef.ICefV8Value, 4)
			callFuncArgs[0] = cef.V8ValueRef.NewString("参数数据")
			callFuncArgs[1] = cef.V8ValueRef.NewBool(true)
			callFuncArgs[2] = cef.V8ValueRef.NewInt(9999)
			callFuncArgs[3] = cef.V8ValueRef.NewDouble(100.99)
			// 执行 ipc.on 回调函数
			ret := onCallback.ExecuteFunctionWithContext(v8ctx, nil, callFuncArgs)
			if ret != nil && ret.IsValid() {
				if ret.IsString() {
					fmt.Println("ret-value:", ret.GetStringValue())
					SendBrowserMessage(frame, "jsreturn", []byte(ret.GetStringValue()))
				}
				ret.FreeAndNil()
			}
			for _, v := range callFuncArgs {
				v.FreeAndNil()
			}
			v8ctx.Exit()
		}
	})
}

func SendBrowserMessage(frame cef.ICefFrame, name string, data []byte) {
	processMessage := cef.ProcessMessageRef.New(name)
	messageArgumentList := processMessage.GetArgumentList()
	var dataPtr = uintptr(0)
	if len(data) > 0 {
		dataPtr = uintptr(unsafe.Pointer(&data[0]))
	}
	dataBin := cef.BinaryValueRef.New(dataPtr, uint32(len(data)))
	messageArgumentList.SetBinary(0, dataBin)
	frame.SendProcessMessage(cef.PID_RENDERER, processMessage)
	if dataBin != nil {
		dataBin.FreeAndNil()
	}
	messageArgumentList.Clear()
	messageArgumentList.FreeAndNil()
	processMessage.FreeAndNil()
}
