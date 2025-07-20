package devtools

import (
	"fmt"
	"github.com/energye/cef/cef"
	"github.com/energye/examples/cef/debug_most/utils"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
)

func DevTools(chromium cef.IChromium) {
	chromium.SetOnDevToolsRawMessage(func(sender lcl.IObject, browser cef.ICefBrowser, message uintptr, messageSize uint32, handled *bool) {
		fmt.Println("OnDevToolsRawMessage message:", message, messageSize)
		data := utils.ReadData(message, messageSize)
		fmt.Println("data:", string(data))
		*handled = false
	})
}

func ShowDevtools(chromium cef.IChromium) {
	lcl.RunOnMainThreadAsync(func(id uint32) {
		point := types.TPoint{
			X: 100,
			Y: 100,
		}
		chromium.ShowDevToolsWithPointWinControl(point, nil)
	})
}

var gId int32 = 0

func ExecuteDevToolsMethod(chromium cef.IChromium) {
	//字典对象
	var dict = cef.DictionaryValueRef.New()
	dict.SetBool("mobile", true)
	dict.SetDouble("deviceScaleFactor", 1)
	TempDict := cef.DictionaryValueRef.New()
	TempDict.SetString("type", "portraitPrimary")
	TempDict.SetInt("angle", 0)
	dict.SetDictionary("screenOrientation", TempDict)
	gId = chromium.ExecuteDevToolsMethod(gId, "Emulation.setDeviceMetricsOverride", dict)
	fmt.Println("ExecuteDevToolsMethod - result messageId:", gId)
	//设置浏览器 userAgent
	dict = cef.DictionaryValueRef.New()
	dict.SetString("userAgent", "Mozilla/5.0 (Linux; Android 11; M2102K1G) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.101 Mobile Safari/537.36")
	gId = chromium.ExecuteDevToolsMethod(gId, "Emulation.setUserAgentOverride", dict)
	fmt.Println("ExecuteDevToolsMethod - result messageId:", gId)
	dict.FreeAndNil()
}

func ExecuteJavaScript(chromium cef.IChromium) {
	var jsCode = `document.body.style.background="#999999";`
	chromium.ExecuteJavaScriptWithStringX2FrameInt(jsCode, "", chromium.Browser().GetMainFrame(), 0)
}
