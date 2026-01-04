package main

import (
	"embed"
	"fmt"
	. "github.com/energye/examples/syso"
	"github.com/energye/examples/wv/windows/application"
	"github.com/energye/lcl/api"
	"github.com/energye/lcl/api/exception"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/tool/exec"
	"github.com/energye/lcl/types"
	"github.com/energye/lcl/types/colors"
	"github.com/energye/lcl/types/messages"
	wvTypes "github.com/energye/wv/types/windows"
	wv "github.com/energye/wv/windows"
	"path/filepath"
)

var mainForm TMainForm
var load wv.IWVLoader

//go:embed resources
var resources embed.FS

func init() {
	TestLoadLibPath()
}
func main() {
	fmt.Println("Go ENERGY Run Main")
	lcl.Init(nil, nil)
	wv.Init()
	exception.SetOnException(func(exception int32, message string) {
		fmt.Println("[ERROR] exception:", exception, "message:", message)
	})
	fmt.Println(api.MainThreadId(), api.CurrentThreadId())

	// GlobalWebView2Loader
	load = application.NewWVLoader()
	fmt.Println("当前目录:", exec.CurrentDir)
	fmt.Println("WebView2Loader.dll目录:", application.WV2LoaderDllPath())
	fmt.Println("用户缓存目录:", filepath.Join(application.WVCachePath(), "webview2Cache"))
	load.SetUserDataFolder(application.WVCachePath())
	load.SetLoaderDllPath(application.WV2LoaderDllPath())
	if r := load.StartWebView2(); r {
		fmt.Println("StartWebView2", r)
	}
	api.SetOnReleaseCallback(func() {
		wv.DestroyGlobalWebView2Loader()
	})
	lcl.Application.Initialize()
	lcl.Application.SetMainFormOnTaskBar(true)
	lcl.Application.NewForm(&mainForm)
	lcl.Application.Run()
}

type TMainForm struct {
	lcl.TEngForm
	windowParent   wv.IWVWindowParent
	browser        wv.IWVBrowser
	getCookieBtn   lcl.IButton
	addCookieBtn   lcl.IButton
	delCookieBtn   lcl.IButton
	cookieList     lcl.IListBox
	menuExitMemory lcl.IMemoryStream
}

func (m *TMainForm) FormCreate(sender lcl.IObject) {
	m.SetCaption("Webview2 Cookie 管理")
	m.SetPosition(types.PoScreenCenter)
	m.SetWidth(1024)
	m.SetHeight(768)
	m.SetDoubleBuffered(true)
	//return
	m.createRightBoxLayout()

	m.windowParent = wv.NewWindowParent(m)
	m.windowParent.SetParent(m)
	//重新调整browser窗口的Parent属性
	//重新设置了上边距，宽，高
	m.windowParent.SetAlign(types.AlCustom) //重置对齐,默认是整个客户端
	m.windowParent.SetHeight(m.Height())
	m.windowParent.SetWidth(m.Width() - m.Width()/4)
	m.windowParent.SetAnchors(types.NewSet(types.AkLeft, types.AkTop, types.AkRight, types.AkBottom))

	m.browser = wv.NewBrowser(m)
	m.browser.SetDefaultURL("https://www.baidu.com")
	m.browser.SetTargetCompatibleBrowserVersion("95.0.1020.44") // 设置
	fmt.Println("TargetCompatibleBrowserVersion:", m.browser.TargetCompatibleBrowserVersion())
	m.browser.SetOnAfterCreated(func(sender lcl.IObject) {
		fmt.Println("回调函数 WVBrowser => SetOnAfterCreated")
		m.windowParent.UpdateSize()
		settings := m.browser.CoreWebView2Settings()
		settings.SetAreDevToolsEnabled(false)
	})
	// 右键菜单图标
	menuExit, _ := resources.ReadFile("resources/menu_exit.png")
	m.menuExitMemory = lcl.NewMemoryStream()
	lcl.StreamHelper.Write(m.menuExitMemory, menuExit)
	menuExitStreamAdapter := lcl.NewStreamAdapter(m.menuExitMemory, types.SoOwned)
	intfMenuExitStreamAdapter := lcl.AsStreamAdapter(menuExitStreamAdapter.AsIntfStream())
	// 代理事件对象
	//delegateEvents := wv.NewWVBrowserDelegateEvents()
	// 右键菜单退出项ID
	var (
		exitItemId    int32
		refreshItemId int32
	)
	m.browser.SetOnContextMenuRequested(func(sender lcl.IObject, webView wv.ICoreWebView2, args wv.ICoreWebView2ContextMenuRequestedEventArgs) {
		fmt.Println("回调函数 WVBrowser => SetOnContextMenuRequested")
		var TempMenuItemItf wv.ICoreWebView2ContextMenuItem
		tmpArgs := wv.NewCoreWebView2ContextMenuRequestedEventArgs(args)
		menuItemCollection := wv.NewCoreWebView2ContextMenuItemCollection(tmpArgs.MenuItems())
		menuItemCollection.RemoveAllMenuItems()
		environment := m.browser.CoreWebView2Environment()

		// 创建菜单项 Exit 带有图标
		if environment.CreateContextMenuItem("EXIT", intfMenuExitStreamAdapter, wvTypes.COREWEBVIEW2_CONTEXT_MENU_ITEM_KIND_COMMAND, &TempMenuItemItf) {
			tmpMenuItem := wv.NewCoreWebView2ContextMenuItem(TempMenuItemItf)
			exitItemId = tmpMenuItem.CommandId()
			fmt.Println("tmpMenuItem", tmpMenuItem.Instance(), TempMenuItemItf.Instance())
			// 设置菜单事件触发对象为delegateEvents, 点击Exit菜单项后，触发 SetOnCustomItemSelected 事件
			tmpMenuItem.AddAllBrowserEvents(m.browser) // .AddCustomItemSelectedEvent(delegateEvents)
			fmt.Println("Initialized", tmpMenuItem.Initialized())
			menuItemCollection.InsertValueAtIndex(0, tmpMenuItem.BaseIntf())
			fmt.Println("exitItemId", exitItemId)
		}
		if environment.CreateContextMenuItem("刷新", nil, wvTypes.COREWEBVIEW2_CONTEXT_MENU_ITEM_KIND_COMMAND, &TempMenuItemItf) {
			tmpMenuItem := wv.NewCoreWebView2ContextMenuItem(TempMenuItemItf)
			refreshItemId = tmpMenuItem.CommandId()
			fmt.Println("tmpMenuItem", tmpMenuItem.Instance(), TempMenuItemItf.Instance())
			// 设置菜单事件触发对象为delegateEvents, 点击Exit菜单项后，触发 SetOnCustomItemSelected 事件
			tmpMenuItem.AddAllBrowserEvents(m.browser) //(delegateEvents)
			menuItemCollection.InsertValueAtIndex(0, tmpMenuItem.BaseIntf())
			fmt.Println("refreshItemId", refreshItemId)

		}

		// free
		menuItemCollection.Free()
		tmpArgs.Free()
	})
	// 代理事件, 自定义菜单项选择事件回调
	m.browser.SetOnCustomItemSelected(func(sender lcl.IObject, menuItem wv.ICoreWebView2ContextMenuItem) {
		menuItem = wv.NewCoreWebView2ContextMenuItem(menuItem)
		fmt.Println("SetOnCustomItemSelected", menuItem.CommandId())
		if exitItemId == menuItem.CommandId() {
			m.Close()
		} else if refreshItemId == menuItem.CommandId() {
			m.browser.Refresh()
		}
		// free
		menuItem.Free()
	})
	m.browser.SetOnGetCookiesCompleted(func(sender lcl.IObject, errorCode types.HRESULT, result wv.ICoreWebView2CookieList) {
		result = wv.NewCoreWebView2CookieList(result)
		count := int(result.Count())
		fmt.Println("count", count, api.MainThreadId() == api.CurrentThreadId())
		m.cookieList.Clear()
		for i := 0; i < count; i++ {
			cookie := result.Items(uint32(i))
			cookie = wv.NewCoreWebView2Cookie(cookie)
			fmt.Println("count", cookie.Name(), cookie.Domain())
			m.cookieList.Items().Add(fmt.Sprintf("%s - %s", cookie.Name(), cookie.Domain()))
		}
	})
	m.browser.SetOnNewWindowRequested(func(sender lcl.IObject, webView wv.ICoreWebView2, args wv.ICoreWebView2NewWindowRequestedEventArgs) {
		fmt.Println("回调函数 =》 SetOnNewWindowRequested")
		args = wv.NewCoreWebView2NewWindowRequestedEventArgs(args)
		fmt.Println("阻止弹出窗口, 在当前页面打开链接:", args.URI())
		args.SetHandled(true)
		m.browser.Navigate(args.URI())
		args.Free()
	})
	// 设置browser到window parent
	m.windowParent.SetBrowser(m.browser)

	// 窗口显示时创建browser
	m.SetOnShow(func(sender lcl.IObject) {
		if load.InitializationError() {
			fmt.Println("回调函数 => SetOnShow 初始化失败")
		} else {
			if load.Initialized() {
				fmt.Println("回调函数 => SetOnShow 初始化成功")
				m.browser.CreateBrowserWithHandleBool(m.windowParent.Handle(), true)
			}
		}
	})
	m.SetOnWndProc(func(msg *types.TMessage) {
		m.InheritedWndProc(msg)
		switch msg.Msg {
		case messages.WM_SIZE, messages.WM_MOVE, messages.WM_MOVING:
			m.browser.NotifyParentWindowPositionChanged()
		}
	})
	m.SetOnDestroy(m.OnFormDestroy)
}

func (m *TMainForm) createRightBoxLayout() {
	messagePanel := lcl.NewPanel(m)
	messagePanel.SetParent(m)
	messagePanel.SetParentDoubleBuffered(true)
	messagePanel.SetLeft(m.Width() - m.Width()/4)
	messagePanel.SetHeight(m.Height())
	messagePanel.SetWidth(m.Width() / 4)
	messagePanel.SetBorderStyleToBorderStyle(types.BsNone)
	messagePanel.SetBorderWidth(1)
	messagePanel.SetColor(colors.ClWhite)
	messagePanel.SetAnchors(types.NewSet(types.AkTop, types.AkRight, types.AkBottom))
	// get
	m.getCookieBtn = lcl.NewButton(messagePanel)
	m.getCookieBtn.SetParent(messagePanel)
	m.getCookieBtn.SetLeft(5)
	m.getCookieBtn.SetTop(5)
	m.getCookieBtn.SetWidth(messagePanel.Width() - 10)
	m.getCookieBtn.SetCaption("获取Cookie")
	m.getCookieBtn.SetOnClick(func(sender lcl.IObject) {
		// 这是一个异步调用，将触发TWVBrowser.OnGetCookies已完成带有Cookie的事件
		fmt.Println("GetCookies", api.MainThreadId() == api.CurrentThreadId())
		m.browser.GetCookies("")
	})
	// list
	m.cookieList = lcl.NewListBox(messagePanel)
	m.cookieList.SetParent(messagePanel)
	m.cookieList.SetLeft(5)
	m.cookieList.SetTop(m.getCookieBtn.Top() + 35)
	m.cookieList.SetWidth(m.getCookieBtn.Width())
	m.cookieList.SetHeight(400)

	// add
	m.addCookieBtn = lcl.NewButton(messagePanel)
	m.addCookieBtn.SetParent(messagePanel)
	m.addCookieBtn.SetLeft(5)
	m.addCookieBtn.SetTop(m.cookieList.Top() + m.cookieList.Height() + 25)
	m.addCookieBtn.SetWidth(messagePanel.Width() - 10)
	m.addCookieBtn.SetCaption("添加Cookie")
	m.addCookieBtn.SetOnClick(func(sender lcl.IObject) {
		cookie := m.browser.CreateCookie("mycustomcookie", "123456", "example.com", "/")
		//cookie = wv.NewCoreWebView2Cookie(cookie)
		m.browser.AddOrUpdateCookie(cookie)
		// free is nil
		cookie.SetInstance(nil)
	})

	// del
	m.delCookieBtn = lcl.NewButton(messagePanel)
	m.delCookieBtn.SetParent(messagePanel)
	m.delCookieBtn.SetLeft(5)
	m.delCookieBtn.SetTop(m.addCookieBtn.Top() + 35)
	m.delCookieBtn.SetWidth(messagePanel.Width() - 10)
	m.delCookieBtn.SetCaption("删除Cookie")
	m.delCookieBtn.SetOnClick(func(sender lcl.IObject) {
		m.browser.DeleteAllCookies()
	})
}

func (m *TMainForm) OnFormDestroy(sender lcl.IObject) {
	fmt.Println("OnFormDestroy")
	m.menuExitMemory.Free()
}
