package v8context

import (
	"bytes"
	"fmt"
	"github.com/energye/cef/cef"
	cefTypes "github.com/energye/cef/types"
	"github.com/energye/examples/cef/debug_most/domvisitor"
	"strconv"
	"unsafe"
)

func Context(app cef.ICefApplication) {
	var (
		ipc         cef.ICefv8Value
		onHandler   cef.IEngV8Handler
		emitHandler cef.IEngV8Handler
		onCallback  cef.ICefv8Value
	)
	app.SetOnContextCreated(func(browser cef.ICefBrowser, frame cef.ICefFrame, context cef.ICefv8Context) {
		onHandler = cef.NewEngV8Handler()
		onHandler.SetOnV8Execute(func(name string, object cef.ICefv8Value, arguments cef.ICefv8ValueArray, retval *cef.ICefv8Value, exception *string) bool {
			fmt.Println("ipc.on Execute name:", name)
			// JS事件名
			//lName := arguments.Get(0)
			// JS事件回调函数
			callFN := arguments.Get(1)
			onCallback = cef.V8ValueRef.UnWrap(callFN.Wrap())
			//callFN.Free()
			//arguments.Free()
			return true
		})
		emitHandler = cef.NewEngV8Handler()
		emitHandler.SetOnV8Execute(func(name string, object cef.ICefv8Value, arguments cef.ICefv8ValueArray, retval *cef.ICefv8Value, exception *string) bool {
			fmt.Println("ipc.emit Execute name:", name)
			v8ctx := cef.V8ContextRef.Current()
			ctxFrame := v8ctx.GetFrame()
			emitName := arguments.Get(0)
			defer func() {
				object.Free()
				ctxFrame.Free()
				v8ctx.Free()
				emitName.Free()
				arguments.Free()
			}()
			fmt.Println("frameId:", ctxFrame.GetIdentifier(), "ProcessType:", app.ProcessType())
			eventName := emitName.GetStringValue()
			if eventName == "domVisitor" {
				domvisitor.DomVisitor()
			} else {
				// 发送消息
				var buf bytes.Buffer
				for i := 1; i < arguments.Count(); i++ {
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
							//arg.Free()
						}
					}
					//val.Free()
				}
				dataBytes := buf.Bytes()
				SendBrowserMessage(ctxFrame, eventName, dataBytes)
			}
			return true
		})
		ipc = cef.V8ValueRef.NewObject(nil, nil)
		onFunc := cef.V8ValueRef.NewFunction("on", cef.AsEngV8Handler(onHandler.AsIntfV8Handler()))
		ipc.SetValueByKey("on", onFunc, cefTypes.V8_PROPERTY_ATTRIBUTE_READONLY)
		emitFunc := cef.V8ValueRef.NewFunction("emit", cef.AsEngV8Handler(emitHandler.AsIntfV8Handler()))
		ipc.SetValueByKey("emit", emitFunc, cefTypes.V8_PROPERTY_ATTRIBUTE_READONLY)
		context.GetGlobal().SetValueByKey("ipc", ipc, cefTypes.V8_PROPERTY_ATTRIBUTE_READONLY)
	})
	app.SetOnProcessMessageReceived(func(browser cef.ICefBrowser, frame cef.ICefFrame, sourceProcess cefTypes.TCefProcessId, message cef.ICefProcessMessage, outResult *bool) {
		fmt.Println("渲染进程 name:", message.GetName())
		args := message.GetArgumentList()
		binArgs := args.GetBinary(0)
		fmt.Println("size:", binArgs.GetSize())
		messageDataBytes := make([]byte, int(binArgs.GetSize()))
		binArgs.GetData(uintptr(unsafe.Pointer(&messageDataBytes[0])), binArgs.GetSize(), 0)
		fmt.Println("data:", string(messageDataBytes))
		v8ctx := frame.GetV8Context()
		//defer binArgs.Free()
		//defer args.Free()
		//defer message.Free()
		//defer v8ctx.Free()
		// 获取当前frame v8context
		// 进入上下文
		if v8ctx.Enter() {
			// 调用JS回调函数
			callFuncArgs := cef.NewCefv8ValueArray(0, 0)
			callFuncArgs.Add(cef.V8ValueRef.NewString("参数数据"))
			callFuncArgs.Add(cef.V8ValueRef.NewBool(true))
			callFuncArgs.Add(cef.V8ValueRef.NewInt(9999))
			callFuncArgs.Add(cef.V8ValueRef.NewDouble(100.99))
			// 执行 ipc.on 回调函数
			ret := onCallback.ExecuteFunctionWithContext(v8ctx, nil, callFuncArgs)
			if ret != nil && ret.IsValid() {
				if ret.IsString() {
					fmt.Println("ret-value:", ret.GetStringValue())
					SendBrowserMessage(frame, "jsreturn", []byte(ret.GetStringValue()))
				}
				//ret.Free()
			}
			//for i := 0; i < callFuncArgs.Count(); i++ {
			//	callFuncArgs.Get(i).Free()
			//}
			v8ctx.Exit()
		}
	})

	app.SetOnWebKitInitialized(func() {
		fmt.Println("SetOnWebKitInitialized")
		var myparamValue string
		v8Handler := cef.NewEngV8Handler()
		v8Handler.SetOnV8Execute(func(name string, object cef.ICefv8Value, arguments cef.ICefv8ValueArray, retval *cef.ICefv8Value, exception *string) bool {
			fmt.Println("v8Handler.Execute", name)
			if name == "GetMyParam" {
				*retval = cef.V8ValueRef.NewString(myparamValue)
			} else if name == "SetMyParam" {
				if arguments.Count() > 0 {
					newValue := arguments.Get(0)
					fmt.Println("value is string:", newValue.IsString())
					fmt.Println("value:", newValue.GetStringValue())
					myparamValue = newValue.GetStringValue()
					//newValue.Free()
				}
			}
			//object.Free()
			//arguments.Free()
			return true
		})
		//注册js
		var jsCode = `
            let test;
            if (!test) {
                test = {};
            }
            (function () {
                test.__defineGetter__('myparam', function () {
                    native function GetMyParam();
                    return GetMyParam();
                });
                test.__defineSetter__('myparam', function (b) {
                    native function SetMyParam();
                    if (b) SetMyParam(b);
                });
            })();
`
		// 注册JS 和v8处理器
		v8Handler = cef.AsEngV8Handler(v8Handler.AsIntfV8Handler())
		cef.MiscFunc.CefRegisterExtension("v8/test", jsCode, v8Handler)
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
	frame.SendProcessMessage(cefTypes.PID_RENDERER, processMessage)
	if dataBin != nil {
		//dataBin.Free()
	}
	messageArgumentList.Clear()
	//messageArgumentList.Free()
	//processMessage.Free()
}
