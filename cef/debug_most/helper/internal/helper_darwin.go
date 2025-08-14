package internal

import (
	"github.com/energye/cef/cef"
	"github.com/energye/examples/cef/application"
)

func InitApplication() cef.ICefApplication {
	app := application.NewApplication()
	app.InitLibLocationFromArgs()
	app.SetUseMockKeyChain(true)
	return app
}
