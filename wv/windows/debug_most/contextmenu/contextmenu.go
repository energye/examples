package contextmenu

import (
	"fmt"
	"github.com/energye/examples/wv/windows/debug_most/devtools"
	"github.com/energye/examples/wv/windows/debug_most/utils"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
	wvTypes "github.com/energye/wv/types/windows"
	wv "github.com/energye/wv/windows"
)

func Contextmenu(form lcl.IForm, browser wv.IWVBrowser) {
	// 右键菜单退出项ID
	var (
		exitItemId     int32
		devtoolsItemId int32
	)
	// 右键菜单图标
	menuExit, err := utils.Assets.ReadFile("assets/menu_exit.png")
	fmt.Println("Contextmenu err:", err)
	menuExitMemory := lcl.NewMemoryStream()
	lcl.StreamHelper.Write(menuExitMemory, menuExit)
	//.LoadFromBytes(menuExit)
	menuExitStreamAdapter := lcl.NewStreamAdapter(menuExitMemory, types.SoOwned)
	baseIntfExitStreamAdapter := lcl.AsStreamAdapter(menuExitStreamAdapter.AsIntfStream())
	browser.SetOnContextMenuRequested(func(sender lcl.IObject, webView wv.ICoreWebView2, args wv.ICoreWebView2ContextMenuRequestedEventArgs) {
		fmt.Println("回调函数 WVBrowser => SetOnContextMenuRequested")
		var TempMenuItemItf wv.ICoreWebView2ContextMenuItem
		tmpArgs := wv.NewCoreWebView2ContextMenuRequestedEventArgs(args)
		menuItemCollection := wv.NewCoreWebView2ContextMenuItemCollection(tmpArgs.MenuItems())
		//menuItemCollection.RemoveAllMenuItems()
		environment := browser.CoreWebView2Environment()

		// 创建菜单项 Exit 带有图标
		if environment.CreateContextMenuItem("EXIT", baseIntfExitStreamAdapter, wvTypes.COREWEBVIEW2_CONTEXT_MENU_ITEM_KIND_COMMAND, &TempMenuItemItf) {
			tmpMenuItem := wv.NewCoreWebView2ContextMenuItem(TempMenuItemItf)
			exitItemId = tmpMenuItem.CommandId()
			fmt.Println("tmpMenuItem", tmpMenuItem.Instance(), TempMenuItemItf.Instance())
			// 设置菜单事件触发对象为delegateEvents, 点击Exit菜单项后，触发 SetOnCustomItemSelected 事件
			tmpMenuItem.AddAllBrowserEvents(browser)
			menuItemCollection.AppendValue(tmpMenuItem.BaseIntf())
		}

		if environment.CreateContextMenuItem("DevTools", nil, wvTypes.COREWEBVIEW2_CONTEXT_MENU_ITEM_KIND_COMMAND, &TempMenuItemItf) {
			tmpMenuItem := wv.NewCoreWebView2ContextMenuItem(TempMenuItemItf)
			devtoolsItemId = tmpMenuItem.CommandId()
			tmpMenuItem.AddAllBrowserEvents(browser)
			menuItemCollection.AppendValue(tmpMenuItem.BaseIntf())
		}

		webView = wv.NewCoreWebView2(webView)
		args = wv.NewCoreWebView2ContextMenuRequestedEventArgs(args)
		menuItems := wv.NewCoreWebView2ContextMenuItemCollection(args.MenuItems())
		contextMenuTarge := wv.NewCoreWebView2ContextMenuTarget(args.ContextMenuTarget())
		fmt.Println("回调函数 WVBrowser => SetOnContextMenuRequested:", menuItems.Count(), contextMenuTarge.PageUri(), webView.BrowserProcessID(), webView.FrameId())
		fmt.Println("回调函数 WVBrowser => SelectedCommandId:", args.SelectedCommandId())
		menuItems.Free()
		contextMenuTarge.Free()
		args.Free()

		// free
		menuItemCollection.Free()
		tmpArgs.Free()
	})
	// 代理事件, 自定义菜单项选择事件回调
	browser.SetOnCustomItemSelected(func(sender lcl.IObject, menuItem wv.ICoreWebView2ContextMenuItem) {
		menuItem = wv.NewCoreWebView2ContextMenuItem(menuItem)
		fmt.Println("SetOnCustomItemSelected", menuItem.CommandId())
		if exitItemId == menuItem.CommandId() {
			menuExitMemory.Free()
			form.Close()
		} else if menuItem.CommandId() == devtoolsItemId {
			devtools.OpenDevtools(browser)
		}
		// free
		menuItem.Free()
	})
}
