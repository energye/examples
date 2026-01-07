package main

import (
	"github.com/energye/energy/v3/application"
	"github.com/energye/energy/v3/wv"
	"github.com/energye/examples/wv/linux/simple/app"
	_ "github.com/energye/examples/wv/linux/simple/resources"
	"os"
)

func main() {
	// linux webkit2 > gtk3
	os.Setenv("--ws", "gtk3")
	wvApp := wv.Init(nil, nil)
	wvApp.SetOptions(application.Options{DefaultURL: "https://www.baidu.com", Caption: "Test Energy"})
	wv.Run(app.Forms...)
}
