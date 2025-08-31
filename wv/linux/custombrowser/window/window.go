package window

import "C"
import (
	"fmt"
	"github.com/energye/examples/wv/assets"
	"github.com/energye/examples/wv/linux/gtkhelper"
	"github.com/energye/lcl/api"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
	"github.com/energye/lcl/types/colors"
	"sync"
	"widget/wg"
)

var (
	CacheRoot    string
	SiteResource string
	Window       BrowserWindow
	bgColor      = colors.RGBToColor(56, 57, 60)
)

type BrowserWindow struct {
	lcl.TEngForm
	box          lcl.IPanel
	mainWindowId int32 // 窗口ID
	windowId     int
	browses      []*Browser  // 当前的chrom列表
	addChromBtn  *wg.TButton // 添加浏览器按钮
	// 浏览器控制按钮
	backBtn    *wg.TButton
	forwardBtn *wg.TButton
	refreshBtn *wg.TButton
	// 地址栏
	addr         lcl.IMemo
	addrRightBtn *wg.TButton
	// 窗口控制按钮
	minBtn   *wg.TButton
	maxBtn   *wg.TButton
	closeBtn *wg.TButton
	// 标题栏相关
	titleHeight           int32 // 标题栏高度
	borderWidth           int32 // 边框宽
	isDown                bool  // 鼠标按下和抬起
	isTitleBar, isDarging bool  // 窗口标题栏
	borderHT              uintptr
	// 窗口关闭锁，一个一个关闭
	browserCloseLock    sync.Mutex
	isWindowButtonClose bool // 点击的窗口关闭按钮
	isChromCloseing     bool // 当前是否有正在关闭的 chrom
	oldWndPrc           uintptr
	normalBounds        types.TRect
	windowState         types.TWindowState
}

func (m *BrowserWindow) FormCreate(sender lcl.IObject) {
	png := lcl.NewPortableNetworkGraphic()
	png.LoadFromFile(assets.GetResourcePath("window-icon_64x64.png"))
	m.Icon().Assign(png)
	png.Free()

	m.SetWidth(1024)
	m.SetHeight(600)
	m.ScreenCenter()
	//m.SetDoubleBuffered(true)

	m.SetOnShow(func(sender lcl.IObject) {
	})

	m.SetOnCloseQuery(func(sender lcl.IObject, canClose *bool) {

	})

	edit := lcl.NewEdit(m)
	edit.SetParent(m)
	edit.SetOnKeyPress(func(sender lcl.IObject, key *uint16) {
		fmt.Println("SetOnKeyPress key:", *key)
	})
	edit.SetOnKeyUp(func(sender lcl.IObject, key *uint16, shift types.TShiftState) {
		fmt.Println("SetOnKeyUp key:", *key)
	})
	edit.SetOnKeyDown(func(sender lcl.IObject, key *uint16, shift types.TShiftState) {

	})
	m.Toolbar()
}

func (m *BrowserWindow) Toolbar() {
	gtkHandle := lcl.PlatformHandle(m.Handle())
	gtkWindowPtr := gtkHandle.Gtk3Window()
	fmt.Println("gtkWindowPtr:", gtkWindowPtr)

	headerBar, err := gtkhelper.NewHeaderBar()
	if err != nil {
		return
	}
	headerBar.SetShowCloseButton(true)

	gtkWindow := gtkhelper.ToGtkWindow(uintptr(gtkWindowPtr))
	gtkWindow.SetTitlebar(headerBar)

	btn := gtkhelper.NewButton() // .ButtonNewWithLabel("button")
	btn.SetRelief(gtkhelper.RELIEF_NONE)
	cssProvid := gtkhelper.NewCssProvider()
	defer cssProvid.Unref()
	cssProvid.LoadFromData(`
button {
	background: transparent;
	border: none;
	padding: 2px; /* 减小点击区域内边距 */
}
button:hover {
	background: rgba(128, 128, 128, 0.2); /* 悬停时轻微灰色背景 */
	border-radius: 2px;
}
button:active {
	background: rgba(128, 128, 128, 0.4); /* 点击时加深背景 */
}
`)

	btnStyleCtx := btn.GetStyleContext()
	btnStyleCtx.AddProvider(cssProvid, gtkhelper.STYLE_PROVIDER_PRIORITY_APPLICATION)

	var sh *gtkhelper.SignalHandler
	sh = btn.SetOnClick(func(sender *gtkhelper.Widget) {
		println("btn.SetOnClick", sender, "IsMainThread:", api.MainThreadId() == api.CurrentThreadId())
		//sh.Disconnect()
	})
	println("OnClick handlerId:", sh.HandlerID(), "eventId:", sh.ID())
	btnIcon := gtkhelper.NewImageFromIconName("open-menu-symbolic", gtkhelper.ICON_SIZE_BUTTON)
	btn.SetImage(btnIcon)

	headerBar.PackStart(btn)

	entry := gtkhelper.NewEntry()
	entry.SetPlaceholderText("请输入")
	entry.SetSizeRequest(250, -1)
	entry.SetOnChanged(func(sender *gtkhelper.Widget, text string) {
		println("entry.SetOnChanged text:", text)
	})
	entry.SetOnCommit(func(sender *gtkhelper.Widget, text string) {
		println("entry.SetOnCommit text:", text)
	})
	entry.SetOnKeyRelease(func(sender *gtkhelper.Widget, key *gtkhelper.EventKey) bool {
		println("entry.SetOnKeyPress key:", key.KeyVal(), gtkhelper.KEY_Return, gtkhelper.KEY_KP_Enter)
		if key.KeyVal() == gtkhelper.KEY_Return || key.KeyVal() == gtkhelper.KEY_KP_Enter {
			println("entry.SetOnKeyPress text:", entry.GetText())
			return true
		}
		return false
	})
	headerBar.SetCustomTitle(entry)
}
