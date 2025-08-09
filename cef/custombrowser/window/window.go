package window

import (
	"fmt"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
	"github.com/energye/lcl/types/colors"
	"strings"
)

type BrowserWindow struct {
	lcl.TEngForm
	box          lcl.IPanel
	content      lcl.IPanel
	mainWindowId int32 // 窗口ID
	canClose     bool
	oldWndPrc    uintptr
	windowId     int
	chroms       map[int32]*Chromium
}

var (
	BW BrowserWindow
)

func (m *BrowserWindow) FormCreate(sender lcl.IObject) {
	m.SetWidth(1024)
	m.SetHeight(768)
	m.SetDoubleBuffered(true)
	//m.SetColor(colors.ClYellow)
	m.SetColor(colors.RGBToColor(56, 57, 60))
	m.ScreenCenter()
	m.SetCaption("ENERGY 3.0 - 自定义浏览器")
	m.chroms = make(map[int32]*Chromium)

	m.box = lcl.NewPanel(m)
	m.box.SetParent(m)
	m.box.SetBevelOuter(types.BvNone)
	m.box.SetDoubleBuffered(true)
	m.box.SetWidth(m.Width())
	m.box.SetHeight(m.Height())
	m.box.SetAnchors(types.NewSet(types.AkLeft, types.AkTop, types.AkRight, types.AkBottom))

	m.content = lcl.NewPanel(m)
	m.content.SetParent(m.box)
	m.content.SetBevelOuter(types.BvNone)
	m.content.SetDoubleBuffered(true)
	m.content.SetAnchors(types.NewSet(types.AkLeft, types.AkTop, types.AkRight, types.AkBottom))
	m.content.SetTop(90)
	m.content.SetLeft(5)
	m.content.SetWidth(m.Width() - 10)
	m.content.SetHeight(m.Height() - (m.content.Top() + 5))

	newChromium := m.createChromium("https://www.baidu.com")
	newChromium.SetOnAfterCreated(m.OnChromiumAfterCreated)
	m.TForm.SetOnActivate(func(sender lcl.IObject) {
		newChromium.createBrowser(nil)
	})

	m.createTitleWidget()

}

func (m *BrowserWindow) createTitleWidget() {
	btn := lcl.NewButton(m)
	btn.SetParent(m)
	btn.SetCaption("测试创建")
	btn.SetOnClick(func(sender lcl.IObject) {
		newChromium := m.createChromium("https://www.baidu.com")
		newChromium.SetOnAfterCreated(m.OnChromiumAfterCreated)
		newChromium.createBrowser(nil)
	})

	//cus := wg.NewButton(m)
	//cus.SetParent(m.box)
	//cus.SetBoundsRect(types.TRect{Left: 108, Top: 48, Right: 250, Bottom: 90})
	//cus.SetStartColor(colors.ClGray)
	//cus.SetEndColor(colors.ClGray)
	//cus.SetRadius(10)

	addr := lcl.NewMemo(m)
	addr.SetParent(m.box)
	addr.SetLeft(110)
	addr.SetTop(50)
	addr.SetHeight(33)
	addr.SetWidth(m.Width() - (addr.Left() + 5))
	//addr.SetBorderStyle(types.BsNone)
	addr.SetAnchors(types.NewSet(types.AkLeft, types.AkTop, types.AkRight))
	addr.Font().SetSize(14)
	addr.Font().SetColor(colors.ClWhite)
	addr.SetColor(colors.RGBToColor(56, 57, 60))
	// 阻止 memo 换行
	addr.SetOnKeyPress(func(sender lcl.IObject, key *uint16) {
		k := *key
		if k == 13 || k == 10 {
			*key = 0
		}
	})
	addr.SetOnChange(func(sender lcl.IObject) {
		text := addr.Text()
		text = strings.ReplaceAll(text, "\r", "")
		text = strings.ReplaceAll(text, "\n", "")
		addr.SetText(text)
	})
}
func (m *BrowserWindow) OnChromiumAfterCreated(newChromium *Chromium) {
	m.chroms[newChromium.windowId] = newChromium
	fmt.Println("OnChromiumAfterCreated", "当前chromium数量:", len(m.chroms), "新chromiumID:", newChromium.windowId)
}

func (m *BrowserWindow) FormAfterCreate(sender lcl.IObject) {
	//m.HookWndProcMessage()
}
