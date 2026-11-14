// ==============================================================================
// Application startup portal
// ==============================================================================

package main

import (
	"embed"
	"fmt"
	"github.com/energye/energy/v3/application"
	"github.com/energye/energy/v3/ipc"
	"github.com/energye/energy/v3/wv"
	"github.com/energye/examples/wv/mywv/app"
	_ "github.com/energye/examples/wv/mywv/resources"
)

//go:embed web
var web embed.FS

func main() {
	// Global Initialization
	wvApp := wv.Init()
	wvApp.SetOptions(application.Options{
		DefaultURL: "fs://energy/index.html",
	})
	// Local resource loading
	wvApp.SetLocalLoad(application.LocalLoad{
		Scheme:     "fs",
		Domain:     "energy",
		ResRootDir: "web",
		FS:         web,
	})
	// IPC
	ipc.On("counter:change", func(context ipc.IContext) {
		data := context.Data().([]any)
		fmt.Println("data:", data)
		context.Result(data[0])
	})
	wv.Run(app.Forms...)
}
