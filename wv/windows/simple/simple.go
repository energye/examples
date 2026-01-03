package main

import (
	"github.com/energye/energy/v3/application"
	"github.com/energye/energy/v3/wv"
	_ "github.com/energye/examples/syso"
	"github.com/energye/examples/wv/windows/simple/app"
)

func main() {
	wvApp := wv.Init(nil, nil)
	wvApp.SetOptions(application.Options{DefaultURL: "https://www.baidu.com"})

	wv.Run(app.Forms...)
}
