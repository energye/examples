package main

import (
	"github.com/energye/energy/v3/application"
	"github.com/energye/energy/v3/wv"
	"github.com/energye/examples/wv/windows/simple/app"
	. "github.com/energye/examples/wv/windows/simple/resources"
)

func main() {
	wvApp := wv.Init(nil, nil)
	wvApp.SetOptions(application.Options{DefaultURL: "https://www.baidu.com"})
	//wvApp.SetOptions(application.Options{DefaultURL: "about:blank"})
	SetIcon()
	wv.Run(app.Forms...)
}
