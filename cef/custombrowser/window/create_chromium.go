package window

import (
	"bytes"
	"fmt"
	"github.com/energye/cef/cef"
	cefTypes "github.com/energye/cef/types"
	"github.com/energye/examples/cef/utils"
	"github.com/energye/examples/cef/utils/draw"
	"github.com/energye/lcl/api"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/tool"
	"github.com/energye/lcl/types"
	"image"
	"image/png"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"widget/wg"
)

type Chromium struct {
	mainWindow *BrowserWindow
	windowId   int32 // 窗口ID
	//timer                              lcl.ITimer
	windowParent                       cef.ICEFWinControl
	chromium                           cef.IChromium
	oldWndPrc                          uintptr
	tabSheetBtn                        *wg.TButton
	tabSheet                           lcl.IPanel
	isActive                           bool
	currentURL                         string
	currentTitle                       string
	siteFavIcon                        map[string]string
	isLoading, canGoBack, canGoForward bool
	isCloseing                         bool
	afterCreate                        func()
}

func (m *Chromium) SetAfterCreate(fn func()) {
	m.afterCreate = fn
}

func (m *Chromium) createBrowser(sender lcl.IObject) {
	rect := m.windowParent.Parent().ClientRect()
	initd := m.chromium.Initialized()
	created := m.chromium.CreateBrowserWithWindowHandleRectStringRequestContextDictionaryValueBool(m.windowParent.Handle(), rect, "", nil, nil, false)
	m.windowParent.UpdateSize()
	fmt.Println("initd created:", initd, created)
}

func (m *Chromium) resize(sender lcl.IObject) {
	if m.chromium != nil {
		m.chromium.NotifyMoveOrResizeStarted()
		if m.windowParent != nil {
			m.windowParent.UpdateSize()
		}
	}
}

func (m *Chromium) closeBrowser() {
	if m.isCloseing {
		return
	}
	m.isCloseing = true
	m.chromium.StopLoad()
	m.tabSheet.SetVisible(false)
	m.chromium.CloseBrowser(true)
	m.windowParent.Free()
	m.tabSheetBtn.Free()
	m.tabSheet.Free()
	println("Chromium.CloseBrowser")
}

func isURLDevtools(url string) bool {
	return strings.Index(url, "devtools://") != -1
}

func (m *Chromium) chromiumBeforeClose(sender lcl.IObject, browser cef.ICefBrowser) {
	println("chromium.OnBeforeClose windowId:", m.windowId, m.chromium.DocumentURL())
	if isURLDevtools(browser.GetMainFrame().GetUrl()) {
		return
	}
	lcl.RunOnMainThreadAsync(func(id uint32) {
		m.mainWindow.removeTabSheetBrowse(m)
	})
}

func (m *Chromium) chromiumClose(sender lcl.IObject, browser cef.ICefBrowser, aAction *cefTypes.TCefCloseBrowserAction) {
	println("chromium.OnClose windowId:", m.windowId)
	if isURLDevtools(browser.GetMainFrame().GetUrl()) {
		return
	}
	if isDarwin {
		m.windowParent.DestroyChildWindow()
		*aAction = cefTypes.CbaClose
	} else if tool.IsLinux() {
		*aAction = cefTypes.CbaClose
	} else {
		*aAction = cefTypes.CbaDelay
		lcl.RunOnMainThreadAsync(func(id uint32) {
			m.windowParent.Free()
		})
	}
}

func (m *BrowserWindow) createChromium(defaultUrl string) *Chromium {
	var tabSheetTop int32 = 90
	if isDarwin {
		tabSheetTop = tabSheetBtnHeight + 8
	}
	newChromium := &Chromium{mainWindow: m, siteFavIcon: make(map[string]string)}
	{
		newChromium.tabSheet = lcl.NewPanel(m)
		newChromium.tabSheet.SetParent(m.box)
		newChromium.tabSheet.SetBevelOuter(types.BvNone)
		newChromium.tabSheet.SetDoubleBuffered(true)
		newChromium.tabSheet.SetAnchors(types.NewSet(types.AkLeft, types.AkTop, types.AkRight, types.AkBottom))
		newChromium.tabSheet.SetTop(tabSheetTop)
		newChromium.tabSheet.SetLeft(5)
		newChromium.tabSheet.SetWidth(m.box.Width() - 10)
		newChromium.tabSheet.SetHeight(m.box.Height() - (newChromium.tabSheet.Top() + 5))
	}
	{
		newChromium.chromium = cef.NewChromium(m)
		//newChromium.chromium.SetRuntimeStyle(cefTypes.CEF_RUNTIME_STYLE_ALLOY)
		options := newChromium.chromium.Options()
		options.SetChromeStatusBubble(cefTypes.STATE_DISABLED)
		//options.SetWebgl(cefTypes.STATE_ENABLED)
	}

	if defaultUrl == "" {
		defaultHtmlPath := getResourcePath("default.html")
		newChromium.chromium.SetDefaultUrl("file://" + defaultHtmlPath)
	} else {
		newChromium.chromium.SetDefaultUrl(defaultUrl)
	}
	if tool.IsWindows() {
		newChromium.windowParent = cef.NewWindowParent(m)
	} else {
		windowParent := cef.NewLinkedWindowParent(m)
		windowParent.SetChromium(newChromium.chromium)
		newChromium.windowParent = windowParent
	}

	newChromium.windowParent.SetParent(newChromium.tabSheet)
	newChromium.windowParent.SetDoubleBuffered(true)
	newChromium.windowParent.SetAlign(types.AlClient)

	// window parent event
	newChromium.windowParent.SetOnEnter(func(sender lcl.IObject) {
		if !newChromium.chromium.FrameIsFocused() {
			newChromium.chromium.SetFocus(true)
		}
	})
	newChromium.windowParent.SetOnExit(func(sender lcl.IObject) {
		newChromium.chromium.SendCaptureLostEvent()
	})

	// chromium event

	// 2. 触发后控制延迟关闭, 在UI线程中调用 windowParent.Free() 释放对象，然后触发 chromium.SetOnBeforeClose
	newChromium.chromium.SetOnClose(newChromium.chromiumClose)
	// 3. 触发后将canClose设置为true, 发送消息到主窗口关闭，触发 m.SetOnCloseQuery
	newChromium.chromium.SetOnBeforeClose(newChromium.chromiumBeforeClose)
	newChromium.chromium.SetOnGotFocus(func(sender lcl.IObject, browser cef.ICefBrowser) {
		if isURLDevtools(browser.GetMainFrame().GetUrl()) {
			return
		}
		if tool.IsLinux() {
			lcl.RunOnMainThreadAsync(func(id uint32) {
				newChromium.windowParent.SetFocus()
			})
		}
	})
	newChromium.chromium.SetOnAfterCreated(func(sender lcl.IObject, browser cef.ICefBrowser) {
		if isURLDevtools(browser.GetMainFrame().GetUrl()) {
			return
		}
		//fmt.Println("SetOnAfterCreated", browser.GetIdentifier(), browser.GetHost().HasDevTools())
		lcl.RunOnMainThreadAsync(func(id uint32) {
			newChromium.windowParent.UpdateSize()
		})
		if newChromium.afterCreate != nil {
			newChromium.afterCreate()
		}
	})

	newChromium.chromium.SetOnBeforePopup(func(sender lcl.IObject, browser cef.ICefBrowser, frame cef.ICefFrame, targetUrl string, targetFrameName string, targetDisposition cefTypes.TCefWindowOpenDisposition, userGesture bool, popupFeatures cef.TCefPopupFeatures, windowInfo *cef.TCefWindowInfo, client *cef.IEngClient, settings *cef.TCefBrowserSettings, extraInfo *cef.ICefDictionaryValue, noJavascriptAccess *bool, result *bool) {
		if isURLDevtools(targetUrl) {
			return
		}
		*result = true
		println("chromium.OnBeforePopup isMainThread:", api.MainThreadId() == api.CurrentThreadId())
		m.SetAddrText("")
		lcl.RunOnMainThreadAsync(func(id uint32) {
			// 创建新的 tab
			newChromium := m.createChromium(targetUrl)
			m.OnChromiumCreateTabSheet(newChromium)
			newChromium.createBrowser(nil)
		})
	})
	newChromium.chromium.SetOnTitleChange(func(sender lcl.IObject, browser cef.ICefBrowser, title string) {
		printIsMainThread()
		if newChromium.tabSheetBtn != nil {
			if isDefaultResourceHTML(title) {
				title = "新建标签页"
			}

			lcl.RunOnMainThreadAsync(func(id uint32) {
				newChromium.tabSheetBtn.SetCaption(title)
				newChromium.tabSheetBtn.SetHint(title)
				newChromium.tabSheetBtn.Invalidate()
			})
		}
		newChromium.currentTitle = title
		if newChromium.isActive && !isDarwin {
			m.updateWindowCaption(title)
		}
	})
	newChromium.chromium.SetOnLoadingStateChange(func(sender lcl.IObject, browser cef.ICefBrowser, isLoading bool, canGoBack bool, canGoForward bool) {
		newChromium.isLoading = isLoading
		newChromium.canGoBack = canGoBack
		newChromium.canGoForward = canGoForward
		//fmt.Println("OnLoadingStateChange isLoading:", isLoading)
		newChromium.mainWindow.updateRefreshBtn(newChromium, isLoading)
		if !isLoading {
			// 加载完 尝试获取已缓存的图标
			loadUrl := browser.GetMainFrame().GetUrl()
			go func() {
				// 设置图标到 tab sheet
				if tempURL, err := url.Parse(loadUrl); err == nil {
					if icoPath, ok := newChromium.siteFavIcon[tempURL.Host]; ok {
						lcl.RunOnMainThreadAsync(func(id uint32) {
							newChromium.tabSheetBtn.SetIconFavorite(icoPath)
							newChromium.tabSheetBtn.Invalidate()
						})
						return
					}
				}
				// 使用默认图标
				lcl.RunOnMainThreadAsync(func(id uint32) {
					newChromium.tabSheetBtn.SetIconFavorite(getResourcePath("icon.png"))
					newChromium.tabSheetBtn.Invalidate()
				})
			}()
		}
		newChromium.updateBrowserControlBtn()
	})
	newChromium.chromium.SetOnLoadStart(func(sender lcl.IObject, browser cef.ICefBrowser, frame cef.ICefFrame, transitionType cefTypes.TCefTransitionType) {
		tempUrl := browser.GetMainFrame().GetUrl()
		if isDefaultResourceHTML(tempUrl) {
			tempUrl = ""
		}
		//println("OnLoadStart URL:", tempUrl)
		newChromium.currentURL = tempUrl
		if newChromium.isActive {
			m.SetAddrText(tempUrl)
		}
	})
	newChromium.chromium.SetOnFavIconUrlChange(func(sender lcl.IObject, browser cef.ICefBrowser, iconUrls lcl.IStrings) {
		var icoURL string
		for i := 0; i < int(iconUrls.Count()); i++ {
			tempUrl := iconUrls.Strings(int32(i))
			if strings.LastIndex(strings.ToLower(tempUrl), ".ico") != -1 {
				icoURL = tempUrl
				break
			}
			if strings.LastIndex(strings.ToLower(tempUrl), ".png") != -1 {
				icoURL = tempUrl
				break
			}
		}
		browserUrl := browser.GetMainFrame().GetUrl()
		//println("OnFavIconUrlChange icoURL:", icoURL, "browserUrl:", browserUrl)
		var host string
		if tempUrl, err := url.Parse(browserUrl); err != nil {
			println("[ERROR] OnFavIconUrlChange ICON Parse URL:", err.Error())
			return
		} else {
			host = tempUrl.Host
		}
		if icoURL != "" {
			if tempURL, err := url.Parse(icoURL); err == nil {
				_, ok := newChromium.siteFavIcon[tempURL.Host]
				if !ok {
					// 下载 favicon.ico
					go func() {
						resp, err := http.Get(icoURL)
						if err == nil {
							defer resp.Body.Close()
							data, err := io.ReadAll(resp.Body)
							if err == nil {
								// png 或 ico 缩放至 16x16
								// 检测图片真实格式
								if imageFormat, err := utils.DetectImageFormatByte(data); err == nil {
									// 缩放图片
									// 把 ico 转 png
									// 把 png 缩放至 16x16
									if imageFormat == "ico" {
										icoBuf := &bytes.Buffer{}
										icoBuf.Write(data)
										// 解码ICO（自动选择最佳尺寸）
										icoImg, err := utils.Decode(icoBuf)
										if err != nil {
											println("[ERROR] OnFavIconUrlChange ICO Decode:", err.Error())
											return
										}
										pngBuf := &bytes.Buffer{}
										// 编码为PNG格式
										if err := png.Encode(pngBuf, icoImg); err != nil {
											println("[ERROR] OnFavIconUrlChange ICO To PNG:", err.Error())
											return
										}
										// 解码 png 到 image
										pngImg, err := png.Decode(pngBuf)
										if err != nil {
											println("[ERROR] OnFavIconUrlChange PNG Decode:", err.Error())
											return
										}
										pngBounds := pngImg.Bounds()
										// 存放缩放后的图像 16x16
										scaledImg := image.NewRGBA(image.Rect(0, 0, 16, 16))
										// 使用 CatmullRom 插值（比双线性更平滑）
										draw.CatmullRom.Scale(scaledImg, scaledImg.Bounds(), pngImg, pngBounds, draw.Over, nil)
										// 最后保存缩放 png
										scalePngBuf := &bytes.Buffer{}
										if err := png.Encode(scalePngBuf, scaledImg); err != nil {
											println("[ERROR] OnFavIconUrlChange PNG Encode Save Buffer:", err.Error())
											return
										}
										data = scalePngBuf.Bytes()
									} else {
										pngBuf := &bytes.Buffer{}
										pngBuf.Write(data)
										// 解码 png 到 image
										pngImg, err := png.Decode(pngBuf)
										if err != nil {
											println("[ERROR] OnFavIconUrlChange PNG Decode:", err.Error())
											return
										}
										pngBounds := pngImg.Bounds()
										// 存放缩放后的图像 16x16
										scaledImg := image.NewRGBA(image.Rect(0, 0, 16, 16))
										// 使用 CatmullRom 插值（比双线性更平滑）
										draw.CatmullRom.Scale(scaledImg, scaledImg.Bounds(), pngImg, pngBounds, draw.Over, nil)
										// 最后保存缩放 png
										scalePngBuf := &bytes.Buffer{}
										if err := png.Encode(scalePngBuf, scaledImg); err != nil {
											println("[ERROR] OnFavIconUrlChange PNG Encode Save Buffer:", err.Error())
											return
										}
										data = scalePngBuf.Bytes()
									}

									// 创建保存目录
									if err = os.MkdirAll(SiteResource, fs.ModePerm); err != nil {
										println("[ERROR] OnFavIconUrlChange MkdirAll:", err.Error())
									}
									// 保存图标目录
									saveIcoPath := filepath.Join(SiteResource, host+"_favicon.png")
									// 保存 logo
									if err = os.WriteFile(saveIcoPath, data, fs.ModePerm); err == nil {
										newChromium.siteFavIcon[tempURL.Host] = saveIcoPath
										// 在此保证更新一次图标到 tabSheetBtn
										lcl.RunOnMainThreadAsync(func(id uint32) {
											newChromium.tabSheetBtn.SetIconFavorite(saveIcoPath)
											newChromium.tabSheetBtn.Invalidate()
										})
									} else {
										println("[ERROR] OnFavIconUrlChange WriteFile:", err.Error())
									}
								} else {
									println("[ERROR] OnFavIconUrlChange DetectImageFormatByte:", err.Error())
								}
							}
						}
					}()
				}
			}
		}
	})
	return newChromium
}

// 过滤 掉一些特定的 url , 在浏览器首页加载时使用的
func isDefaultResourceHTML(v string) bool {
	return v == "about:blank" || v == "DevTools" ||
		(strings.Index(v, "file://") != -1 && strings.Index(v, "resources") != -1) ||
		strings.Index(v, "default.html") != -1 ||
		strings.Index(v, "view-source:file://") != -1 ||
		strings.Index(v, "devtools://") != -1
}
