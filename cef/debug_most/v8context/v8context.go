package v8context

import (
	"fmt"
	"github.com/energye/cef/cef"
)

func Context(app cef.ICefApplication) {
	var (
		obj       cef.ICefV8Value
		onHandler cef.IV8Handler
	)
	app.SetOnContextCreated(func(browser cef.ICefBrowser, frame cef.ICefFrame, context cef.ICefV8Context) {
		onHandler = cef.NewV8Handler()
		onHandler.SetOnExecute(func(name string, object cef.ICefV8Value, arguments cef.ICefV8ValueArray) (retVal cef.ICefV8Value, exception string, result bool) {
			fmt.Println("OnExecute name:", name)
			for i := 0; i < arguments.Size(); i++ {
				val := arguments.Get(i)
				if val.IsString() {
					fmt.Println("\tvalue:", val.GetStringValue())
				} else if val.IsInt() {
					fmt.Println("\tvalue:", val.GetStringValue())
				}
			}
			arguments.Free()
			return
		})
		fmt.Println("onHandler:", onHandler.Instance())
		obj = cef.V8ValueRef.NewObject(nil, nil)
		fmt.Println("obj:", obj.IsObject())
		onFunc := cef.V8ValueRef.NewFunction("on", onHandler.AsInterface())
		obj.SetValueByKey("on", onFunc, cef.V8_PROPERTY_ATTRIBUTE_READONLY)
		context.GetGlobal().SetValueByKey("ipc", obj, cef.V8_PROPERTY_ATTRIBUTE_READONLY)
	})
}
