package scheme

import (
	"fmt"
	"github.com/energye/cef/cef"
	"strings"
)

const (
	SchemeName = "fs"
	DomainName = "energy"
)

func ApplicationOnRegCustomSchemes(registrar cef.ICefSchemeRegistrarRef) {
	fmt.Println("scheme -> OnRegCustomSchemes")
	switch strings.ToUpper(SchemeName) {
	case "HTTP", "HTTPS", "FILE", "FTP", "ABOUT", "DATA":
		return
	}
	registrar.AddCustomScheme(SchemeName, cef.CEF_SCHEME_OPTION_STANDARD|cef.CEF_SCHEME_OPTION_CORS_ENABLED|cef.CEF_SCHEME_OPTION_SECURE|cef.CEF_SCHEME_OPTION_FETCH_ENABLED)
}

func ChromiumAfterCreated(browser cef.ICefBrowser) {
	fmt.Println("scheme -> OnAfterCreated")
	handlerFactory := cef.NewSchemeHandlerFactory(0)
	handlerFactory.SetOnNew(func(browser cef.ICefBrowser, frame cef.ICefFrame, schemeName string, request cef.ICefRequest) (result cef.ICefResourceHandler) {
		fmt.Println("scheme -> handlerFactory -> SetOnNew schemeName:", schemeName)
		resourceHandler := cef.NewResourceHandler(browser, frame, schemeName, request)
		resourceHandler.SetOnProcessRequest(func(request cef.ICefRequest, callback cef.ICefCallback, outResult *bool) {
			fmt.Println("scheme -> handlerFactory -> resourceHandler -> SetOnProcessRequest")
		})
		resourceHandler.SetOnGetResponseHeaders(func(response cef.ICefResponse, outResponseLength *int64, outRedirectUrl *string) {
			fmt.Println("scheme -> handlerFactory -> resourceHandler -> SetOnGetResponseHeaders")
		})
		resourceHandler.SetOnReadResponse(func(dataOut uintptr, bytesToRead int32, bytesRead *int32, callback cef.ICefCallback, outResult *bool) {
			fmt.Println("scheme -> handlerFactory -> resourceHandler -> SetOnReadResponse")
		})
		return resourceHandler.AsInterface()
	})
	browser.GetHost().GetRequestContext().RegisterSchemeHandlerFactory(SchemeName, DomainName, handlerFactory.AsInterface())
	handlerFactory.Free()
}
