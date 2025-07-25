package scheme

import (
	"fmt"
	"github.com/energye/cef/cef"
	cefTypes "github.com/energye/cef/types"
	"github.com/energye/examples/cef/debug_most/utils"
	"io/ioutil"
	"net/url"
	"path/filepath"
	"strings"
	"unsafe"
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
	registrar.AddCustomScheme(SchemeName, cefTypes.CEF_SCHEME_OPTION_STANDARD|cefTypes.CEF_SCHEME_OPTION_CORS_ENABLED|cefTypes.CEF_SCHEME_OPTION_SECURE|cefTypes.CEF_SCHEME_OPTION_FETCH_ENABLED)
}

type SliceHeader struct {
	Data uintptr
	Len  int
	Cap  int
}

func ChromiumAfterCreated(browser cef.ICefBrowser) {
	fmt.Println("scheme -> OnAfterCreated")
	schemeHandlerFactory := cef.NewEngSchemeHandlerFactory(0)
	schemeHandlerFactory.SetOnSchemeFactoryNew(func(browser cef.ICefBrowser, frame cef.ICefFrame, schemeName string, request cef.ICefRequest) cef.IEngResourceHandler {
		fmt.Println("scheme -> handlerFactory -> SetOnNew schemeName:", schemeName, "url:", request.GetUrl())
		reqUrl, err := url.Parse(request.GetUrl())
		if err != nil {
			return nil
		}
		if reqUrl.Scheme != SchemeName {
			return nil
		}
		var (
			statusCode   int32 = 404
			statusText         = "Not Found"
			mimeType           = "text/html"
			responseData []byte
			start        int
		)
		var readData = func(dataOut uintptr, bytesToRead int32, bytesRead *int32) bool {
			dataSize := len(responseData)
			if start < dataSize {
				var min = func(x, y int) int {
					if x < y {
						return x
					}
					return y
				}
				space := min(dataSize, int(bytesToRead))
				//把dataOut指针初始化Go类型的切片
				//space切片长度和空间, 使用bytes长度和bytesToRead最小的值
				dataOutByteSlice := &SliceHeader{
					Data: dataOut,
					Len:  space,
					Cap:  space,
				}
				dst := *(*[]byte)(unsafe.Pointer(dataOutByteSlice))
				//end 计算当前读取资源数据的结束位置
				end := start
				//拿出最小的数据长度做为结束位置
				//bytesToRead当前最大读取数量一搬最大值是固定
				if dataSize < int(bytesToRead) {
					end += dataSize
				} else {
					end += int(bytesToRead)
				}
				//如果结束位置大于bytes长度,我们使用bytes长度
				end = min(end, dataSize)
				//把每次分块读取的资源数据复制到dataOut
				c := copy(dst, responseData[start:end])
				start += c            //设置下次读取资源开始位置
				*bytesRead = int32(c) //读取资源读取字节个数
				return *bytesRead > 0
			}
			return false
		}
		resourceHandler := cef.NewEngResourceHandler(browser, frame, schemeName, request)
		resourceHandler.SetOnResourceProcessRequest(func(request cef.ICefRequest, callback cef.ICefCallback) bool {
			fmt.Println("scheme -> handlerFactory -> resourceHandler -> SetOnProcessRequest")
			responseData, err = ioutil.ReadFile(filepath.Join(utils.RootPath(), "assets\\scheme.html"))
			if err == nil {
				statusCode = 200
				statusText = "OK"
				start = 0
			}
			fmt.Println("\tresponseData size:", len(responseData))
			callback.Cont()
			return true
		})
		resourceHandler.SetOnResourceGetResponseHeaders(func(response cef.ICefResponse, outResponseLength *int64, outRedirectUrl *string) {
			fmt.Println("scheme -> handlerFactory -> resourceHandler -> SetOnGetResponseHeaders")
			fmt.Println("\tstatusCode:", statusCode, "statusText:", statusText, "mimeType:", mimeType, "size:", len(responseData))
			response.SetStatus(statusCode)
			response.SetStatusText(statusText)
			response.SetMimeType(mimeType)
			*outResponseLength = int64(len(responseData))
		})

		//resourceHandler.SetOnRead(func(dataOut uintptr, bytesToRead int32, bytesRead *int32, callback cef.ICefResourceReadCallback, outResult *bool) {
		//	fmt.Println("scheme -> handlerFactory -> resourceHandler -> SetOnRead")
		//	*outResult = readData(dataOut, bytesToRead, bytesRead)
		//	if *outResult {
		//		callback.Cont(int64(*bytesRead))
		//	}
		//})
		resourceHandler.SetOnResourceReadResponse(func(dataOut uintptr, bytesToRead int32, bytesRead *int32, callback cef.ICefCallback) bool {
			fmt.Println("scheme -> handlerFactory -> resourceHandler -> SetOnReadResponse")
			r := readData(dataOut, bytesToRead, bytesRead)
			if r {
				callback.Cont()
			}
			return r
		})
		resourceHandler = cef.AsEngResourceHandler(resourceHandler.AsIntfResourceHandler())
		return resourceHandler
	})

	intfSchemeHandlerFactory := cef.AsEngSchemeHandlerFactory(schemeHandlerFactory.AsIntfSchemeHandlerFactory())
	browser.GetHost().GetRequestContext().RegisterSchemeHandlerFactory(SchemeName, DomainName, intfSchemeHandlerFactory)
}
