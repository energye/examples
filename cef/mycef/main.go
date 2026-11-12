// ==============================================================================
// Application startup portal
// ==============================================================================

package main

import (
	"embed"
	"github.com/energye/energy/v3/application"
	"github.com/energye/energy/v3/cef"
	"github.com/energye/energy/v3/ipc"
	"github.com/energye/examples/cef/mycef/app"
	_ "github.com/energye/examples/cef/mycef/resources"
)

//go:embed web
var web embed.FS

func main() {
	// Global Initialization
	cefApp := cef.Init()
	cefApp.SetOptions(application.Options{
		DefaultURL: "fs://energy/index.html",
	})
	// Local resource loading
	cefApp.SetLocalLoad(application.LocalLoad{
		Scheme:     "fs",
		Domain:     "energy",
		ResRootDir: "web",
		FS:         web,
	})
	// IPC
	ipc.On("counter:change", func(context ipc.IContext) {
		data := context.Data().([]any)
		context.Result(data[0])
	})
	// Convert the form list to the CEF window list and start the application
	forms := make([]cef.IAppWindow, 0, len(app.Forms))
	for _, form := range app.Forms {
		forms = append(forms, form)
	}
	cef.Run(forms...)
}
