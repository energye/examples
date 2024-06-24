package v8context

import (
	"fmt"
	"github.com/energye/cef/cef"
)

func Context(app cef.ICefApplication) {
	var (
		obj       cef.ICefv8Value
		onHandler cef.IV8Handler
	)
	app.SetOnContextCreated(func(browser cef.ICefBrowser, frame cef.ICefFrame, context cef.ICefv8Context) {
		onHandler = cef.NewV8Handler()
		onHandler.SetOnExecute(func(name string, object cef.ICefv8Value, arguments cef.ICefV8ValueArray) (retVal cef.ICefv8Value, exception string, result bool) {
			fmt.Println("OnExecute name:", name)
			for i := 0; i < arguments.Size(); i++ {
				val := arguments.Get(i)
				fmt.Println("\tvalue:", val.GetStringValue())
			}
			arguments.Free()
			return
		})
		fmt.Println("onHandler:", onHandler.Instance())
		obj = cef.V8ValueRef.NewObject(nil, nil)
		fmt.Println("obj:", obj.IsObject())
		onFunc := cef.V8ValueRef.NewFunction("on", onHandler)
		obj.SetValueByKey("on", onFunc, cef.V8_PROPERTY_ATTRIBUTE_READONLY)
		context.GetGlobal().SetValueByKey("ipc", obj, cef.V8_PROPERTY_ATTRIBUTE_READONLY)
	})
}
