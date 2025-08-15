package contextmenu

import (
	"fmt"
	"github.com/energye/cef/cef"
	cefTypes "github.com/energye/cef/types"
	"github.com/energye/examples/cef/debug_most/devtools"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types/colors"
)

func ContextMenu(chromium cef.IChromium) {
	var (
		menuId01           cefTypes.TCefMenuId
		menuId02           cefTypes.TCefMenuId
		menuId03           cefTypes.TCefMenuId
		menuId0301         cefTypes.TCefMenuId
		menuId0302         cefTypes.TCefMenuId
		menuIdCheck        cefTypes.TCefMenuId
		isMenuIdCheck      = true
		menuIdEnable       cefTypes.TCefMenuId
		isMenuIdEnable     = true
		menuIdEnableCtl    cefTypes.TCefMenuId
		menuIdRadio101     cefTypes.TCefMenuId
		menuIdRadio102     cefTypes.TCefMenuId
		menuIdRadio103     cefTypes.TCefMenuId
		radioDefault1Check cefTypes.TCefMenuId
		menuIdRadio201     cefTypes.TCefMenuId
		menuIdRadio202     cefTypes.TCefMenuId
		menuIdRadio203     cefTypes.TCefMenuId
		radioDefault2Check cefTypes.TCefMenuId
		refresh            cefTypes.TCefMenuId
		devtoolsId         cefTypes.TCefMenuId
	)
	nextMenuId := cefTypes.MENU_ID_USER_FIRST
	var nextCommandId = func(reset ...bool) cefTypes.TCefMenuId {
		if len(reset) > 0 {
			nextMenuId = cefTypes.MENU_ID_USER_FIRST
		}
		nextMenuId++
		return cefTypes.TCefMenuId(nextMenuId)
	}
	chromium.SetOnBeforeContextMenu(func(sender lcl.IObject, browser cef.ICefBrowser, frame cef.ICefFrame, params cef.ICefContextMenuParams, model cef.ICefMenuModel) {
		fmt.Println("OnBeforeContextMenu")
		model.AddSeparator()
		menuId01 = cefTypes.MENU_ID_USER_FIRST + 1
		menuId01 = nextCommandId(true)
		model.AddItem(menuId01, "菜单一 html 文字变红色")
		menuId02 = nextCommandId()
		model.AddItem(menuId02, "菜单二 html 文字变绿色")
		menuId03 = nextCommandId()
		menu03 := model.AddSubMenu(menuId03, "菜单三 带有子菜单")
		menuId0301 = nextCommandId()
		menu03.AddItem(menuId0301, "菜单三的子菜单一 ")
		menuId0302 = nextCommandId()
		menu03.AddItem(menuId0302, "菜单三的子菜单二")
		model.AddSeparator()
		//check
		menuIdCheck = nextCommandId()
		model.AddCheckItem(menuIdCheck, "这是一个checkItem-好像就windows有效")
		model.SetChecked(menuIdCheck, isMenuIdCheck)
		//enable
		model.AddSeparator()
		menuIdEnable = nextCommandId()
		if isMenuIdEnable {
			model.AddItem(menuIdEnable, "菜单-已启用")
			model.SetColor(menuIdEnable, cefTypes.CEF_MENU_COLOR_TEXT, colors.NewARGB(255, 111, 12, 200).ARGB())
		} else {
			model.AddItem(menuIdEnable, "菜单-已禁用")
		}
		model.SetEnabled(menuIdEnable, isMenuIdEnable)
		menuIdEnableCtl = nextCommandId()
		model.AddItem(menuIdEnableCtl, "启用上面菜单")
		//为什么要用Visible而不是不创建这个菜单? 因为菜单项的ID是动态的啊。
		model.SetVisible(menuIdEnableCtl, !isMenuIdEnable)
		if !isMenuIdEnable {
			model.SetColor(menuIdEnableCtl, cefTypes.CEF_MENU_COLOR_TEXT, colors.NewARGB(255, 222, 111, 0).ARGB())
		}
		model.AddSeparator()
		//radio 1组
		menuIdRadio101 = nextCommandId()
		menuIdRadio102 = nextCommandId()
		menuIdRadio103 = nextCommandId()
		model.AddRadioItem(menuIdRadio101, "单选按钮 1 1组", 1001)
		model.AddRadioItem(menuIdRadio102, "单选按钮 2 1组", 1001)
		model.AddRadioItem(menuIdRadio103, "单选按钮 3 1组", 1001)
		if radioDefault1Check == 0 {
			radioDefault1Check = menuIdRadio101
		}
		model.SetChecked(radioDefault1Check, true)
		model.AddSeparator()
		//radio 2组
		menuIdRadio201 = nextCommandId()
		menuIdRadio202 = nextCommandId()
		menuIdRadio203 = nextCommandId()
		model.AddRadioItem(menuIdRadio201, "单选按钮 1 2组", 1002)
		model.AddRadioItem(menuIdRadio202, "单选按钮 2 2组", 1002)
		model.AddRadioItem(menuIdRadio203, "单选按钮 3 2组", 1002)
		if radioDefault2Check == 0 {
			radioDefault2Check = menuIdRadio201
		}
		model.SetChecked(radioDefault2Check, true)
		refresh = nextCommandId()
		model.AddItem(refresh, "强制刷新")
		devtoolsId = nextCommandId()
		model.AddItem(devtoolsId, "开发者工具")
	})
	chromium.SetOnContextMenuCommand(func(sender lcl.IObject, browser cef.ICefBrowser, frame cef.ICefFrame, params cef.ICefContextMenuParams, commandId cefTypes.TCefMenuId, eventFlags uint32, result *bool) {
		fmt.Println("OnContextMenuCommand commandId:", commandId)
		switch commandId {
		case menuId01:
		case menuId02:
		case menuIdEnable:
			isMenuIdEnable = !isMenuIdEnable
		case menuIdCheck:
			isMenuIdCheck = !isMenuIdCheck
		case menuIdEnableCtl:
			isMenuIdEnable = true
		case menuIdRadio101, menuIdRadio102, menuIdRadio103:
			radioDefault1Check = commandId
		case menuIdRadio201, menuIdRadio202, menuIdRadio203:
			radioDefault2Check = commandId
		case refresh:
			chromium.ReloadIgnoreCache()
		case devtoolsId:
			devtools.ShowDevtools(chromium)
		}
	})
}
