package window

import (
	"fmt"
	"github.com/energye/cef/cef"
	cefTypes "github.com/energye/cef/types"
	"github.com/energye/examples/cef/utils"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/tool"
	"github.com/energye/lcl/types"
	"github.com/energye/lcl/types/colors"
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
	mainWindow                         *BrowserWindow
	windowId                           int32 // 窗口ID
	timer                              lcl.ITimer
	windowParent                       cef.ICEFWinControl
	chromium                           cef.IChromium
	canClose                           bool
	oldWndPrc                          uintptr
	tabSheetBtn                        *wg.TButton
	isActive                           bool
	currentURL                         string
	currentTitle                       string
	siteFavIcon                        map[string]string
	isLoading, canGoBack, canGoForward bool
	isClose                            bool
}

func (m *Chromium) createBrowser(sender lcl.IObject) {
	if m.timer == nil {
		return
	}
	m.timer.SetEnabled(false)
	rect := m.windowParent.Parent().ClientRect()
	m.chromium.Initialized()
	created := m.chromium.CreateBrowserWithWindowHandleRectStringRequestContextDictionaryValueBool(m.windowParent.Handle(), rect, "", nil, nil, false)
	//fmt.Println("createBrowser rect:", rect, "init:", init, "create:", created)
	if !created {
		m.timer.SetEnabled(true)
	} else {
		m.windowParent.UpdateSize()
		m.timer.Free()
		m.timer = nil
	}
}

func (m *Chromium) resize(sender lcl.IObject) {
	if m.chromium != nil {
		m.chromium.NotifyMoveOrResizeStarted()
		if m.windowParent != nil {
			m.windowParent.UpdateSize()
		}
	}
}

func (m *Chromium) chromiumClose(sender lcl.IObject, browser cef.ICefBrowser, aAction *cefTypes.TCefCloseBrowserAction) {
	//fmt.Println("chromium.Close")
	if tool.IsDarwin() {
		m.windowParent.DestroyChildWindow()
		*aAction = cefTypes.CbaClose
	} else {
		*aAction = cefTypes.CbaDelay
		lcl.RunOnMainThreadAsync(func(id uint32) {
			m.windowParent.Free()
		})
	}
}

func (m *Chromium) chromiumBeforeClose(sender lcl.IObject, browser cef.ICefBrowser) {
	//fmt.Println("chromium.BeforeClose")
	m.canClose = true
	m.isClose = true
}

func (m *Chromium) updateTabSheetActive(isActive bool) {
	if m.tabSheetBtn == nil {
		return
	}
	if isActive {
		activeColor := colors.RGBToColor(86, 88, 93)
		m.tabSheetBtn.SetStartColor(activeColor)
		m.tabSheetBtn.SetEndColor(activeColor)
		m.windowParent.SetVisible(true)
		m.isActive = true
		lcl.RunOnMainThreadAsync(func(id uint32) {
			m.mainWindow.addr.SetText(m.currentURL)
		})
		m.mainWindow.updateWindowCaption(m.currentTitle)
	} else {
		notActiveColor := colors.RGBToColor(56, 57, 60)
		m.tabSheetBtn.SetStartColor(notActiveColor)
		m.tabSheetBtn.SetEndColor(notActiveColor)
		m.windowParent.SetVisible(false)
		m.isActive = false
	}
	m.tabSheetBtn.Invalidate()
	// 根据当前 chromium 浏览器加载状态更新浏览器控制按钮
	m.updateBrowserControlBtn()
}

// 根据当前 chromium 浏览器加载状态更新浏览器控制按钮
func (m *Chromium) updateBrowserControlBtn() {
	m.mainWindow.backBtn.IsDisable = !m.canGoBack
	m.mainWindow.forwardBtn.IsDisable = !m.canGoForward
	lcl.RunOnMainThreadAsync(func(id uint32) {
		if !m.canGoBack {
			// 禁用
			m.mainWindow.backBtn.SetIcon(getResourcePath("back_disable.png"))
		} else {
			m.mainWindow.backBtn.SetIcon(getResourcePath("back.png"))
		}
		m.mainWindow.backBtn.Invalidate()
		if !m.canGoForward {
			// 禁用
			m.mainWindow.forwardBtn.SetIcon(getResourcePath("forward_disable.png"))
		} else {
			m.mainWindow.forwardBtn.SetIcon(getResourcePath("forward.png"))
		}
		m.mainWindow.forwardBtn.Invalidate()
	})
}

func (m *Chromium) closeBrowser() {
	m.chromium.StopLoad()
	m.windowParent.SetVisible(false)
	m.chromium.CloseBrowser(true)
	m.tabSheetBtn.Free()
}

func (m *BrowserWindow) createChromium(defaultUrl string) *Chromium {
	newChromium := &Chromium{mainWindow: m, siteFavIcon: make(map[string]string)}

	newChromium.chromium = cef.NewChromium(m)
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
	newChromium.windowParent.SetParent(m.content)
	newChromium.windowParent.SetDoubleBuffered(true)
	newChromium.windowParent.SetAlign(types.AlClient)
	// 创建一个定时器, 用来createBrowser
	newChromium.timer = lcl.NewTimer(m)
	newChromium.timer.SetEnabled(false)
	newChromium.timer.SetInterval(200)
	newChromium.timer.SetOnTimer(newChromium.createBrowser)

	m.content.SetOnResize(newChromium.resize)
	m.content.SetOnEnter(func(sender lcl.IObject) {
		newChromium.chromium.Initialized()
		newChromium.chromium.FrameIsFocused()
		newChromium.chromium.SetFocus(true)
	})

	newChromium.windowParent.SetOnExit(func(sender lcl.IObject) {
		newChromium.chromium.SendCaptureLostEvent()
	})

	// 2. 触发后控制延迟关闭, 在UI线程中调用 windowParent.Free() 释放对象，然后触发 chromium.SetOnBeforeClose
	newChromium.chromium.SetOnClose(newChromium.chromiumClose)
	// 3. 触发后将canClose设置为true, 发送消息到主窗口关闭，触发 m.SetOnCloseQuery
	newChromium.chromium.SetOnBeforeClose(newChromium.chromiumBeforeClose)

	newChromium.chromium.SetOnAfterCreated(func(sender lcl.IObject, browser cef.ICefBrowser) {
		//fmt.Println("SetOnAfterCreated", browser.GetIdentifier(), browser.GetHost().HasDevTools())
		newChromium.windowParent.UpdateSize()
	})
	newChromium.chromium.SetOnBeforeBrowse(func(sender lcl.IObject, browser cef.ICefBrowser, frame cef.ICefFrame, request cef.ICefRequest,
		userGesture, isRedirect bool, result *bool) {
		//fmt.Println("SetOnBeforeBrowse", browser.GetIdentifier(), browser.GetHost().HasDevTools())
		newChromium.windowParent.UpdateSize()
	})
	newChromium.chromium.SetOnBeforePopup(func(sender lcl.IObject, browser cef.ICefBrowser, frame cef.ICefFrame,
		popupId int32, targetUrl string, targetFrameName string, targetDisposition cefTypes.TCefWindowOpenDisposition, userGesture bool,
		popupFeatures cef.TCefPopupFeatures, windowInfo *cef.TCefWindowInfo, client *cef.IEngClient, settings *cef.TCefBrowserSettings,
		extraInfo *cef.ICefDictionaryValue, noJavascriptAccess *bool, result *bool) {
		*result = true
		lcl.RunOnMainThreadAsync(func(id uint32) {
			// 创建新的 tab
			m.addr.SetText("")
			newChromium := m.createChromium(targetUrl)
			m.OnChromiumCreateTabSheet(newChromium)
			newChromium.createBrowser(nil)
		})
	})
	newChromium.chromium.SetOnTitleChange(func(sender lcl.IObject, browser cef.ICefBrowser, title string) {
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
		m.updateWindowCaption(title)
	})
	newChromium.chromium.SetOnLoadingStateChange(func(sender lcl.IObject, browser cef.ICefBrowser, isLoading bool, canGoBack bool, canGoForward bool) {
		newChromium.isLoading = isLoading
		newChromium.canGoBack = canGoBack
		newChromium.canGoForward = canGoForward
		//fmt.Println("OnLoadingStateChange isLoading:", isLoading)
		if isLoading {
			lcl.RunOnMainThreadAsync(func(id uint32) {
				newChromium.mainWindow.refreshBtn.SetIcon(getResourcePath("stop.png"))
			})
		} else {
			lcl.RunOnMainThreadAsync(func(id uint32) {
				newChromium.mainWindow.refreshBtn.SetIcon(getResourcePath("refresh.png"))
			})
		}
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
		//fmt.Println("OnLoadStart URL:", tempUrl)
		newChromium.currentURL = tempUrl
		if newChromium.isActive {
			lcl.RunOnMainThreadAsync(func(id uint32) {
				m.addr.SetText(tempUrl)
				m.addr.SetFocus()
			})
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
		}
		fmt.Println("OnFavIconUrlChange:", icoURL)
		if icoURL != "" {
			if tempURL, err := url.Parse(icoURL); err == nil {
				_, ok := newChromium.siteFavIcon[tempURL.Host]
				if !ok {
					// 下载 favicon.ico
					go func() {
						resp, err := http.Get(icoURL)
						if err == nil {
							host := resp.Request.Host
							defer resp.Body.Close()
							data, err := io.ReadAll(resp.Body)
							if err == nil {
								if imageFormat, err := utils.DetectImageFormatByte(data); err == nil {
									saveIcoPath := filepath.Join(SiteResource, host+"_favicon."+imageFormat)
									_ = os.MkdirAll(SiteResource, fs.ModeDir)
									if err = os.WriteFile(saveIcoPath, data, fs.ModePerm); err == nil {
										newChromium.siteFavIcon[tempURL.Host] = saveIcoPath
										// 在此保证更新一次图标
										lcl.RunOnMainThreadAsync(func(id uint32) {
											newChromium.tabSheetBtn.SetIconFavorite(saveIcoPath)
											newChromium.tabSheetBtn.Invalidate()
										})
									}
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
