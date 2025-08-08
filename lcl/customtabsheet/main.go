package main

import (
	"embed"
	"fmt"
	. "github.com/energye/examples/syso"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/pkgs/win"
	"github.com/energye/lcl/types"
	"github.com/energye/lcl/types/colors"
	"github.com/energye/lcl/types/messages"
	"syscall"
	"time"
	"unsafe"
	"widget/wg"
)

func init() {
	TestLoadLibPath()
	Chdir("lcl/action")
}

type TMainForm struct {
	lcl.TEngForm
	oldWndPrc uintptr
}

var MainForm TMainForm

//go:embed resources
var resources embed.FS

func main() {
	lcl.Init(nil, nil)
	lcl.Application.Initialize()
	lcl.Application.SetMainFormOnTaskBar(true)
	lcl.Application.NewForm(&MainForm)
	lcl.Application.Run()
}

type TTabState = int32

const (
	tsNormal TTabState = iota
	tsHover
	tsActive
)

func ReadImgData(name string) []byte {
	data, err := resources.ReadFile("resources/" + name)
	if err != nil {
		panic(err)
	}
	return data
}

func (m *TMainForm) wndProc(hwnd types.HWND, message uint32, wParam, lParam uintptr) uintptr {
	switch message {
	case messages.WM_DPICHANGED:
		if !lcl.Application.Scaled() {
			newWindowSize := (*types.TRect)(unsafe.Pointer(lParam))
			win.SetWindowPos(m.Handle(), uintptr(0),
				newWindowSize.Left, newWindowSize.Top, newWindowSize.Right-newWindowSize.Left, newWindowSize.Bottom-newWindowSize.Top,
				win.SWP_NOZORDER|win.SWP_NOACTIVATE)
		}
	}
	switch message {
	case messages.WM_ACTIVATE:
		// If we want to have a frameless window but with the default frame decorations, extend the DWM client area.
		// This Option is not affected by returning 0 in WM_NCCALCSIZE.
		// As a result we have hidden the titlebar but still have the default window frame styling.
		// See: https://docs.microsoft.com/en-us/windows/win32/api/dwmapi/nf-dwmapi-dwmextendframeintoclientarea#remarks
		win.ExtendFrameIntoClientArea(m.Handle(), win.Margins{CxLeftWidth: 1, CxRightWidth: 1, CyTopHeight: 1, CyBottomHeight: 1})
	case messages.WM_NCCALCSIZE:
		// Trigger condition: Change the window size
		// Disable the standard frame by allowing the client area to take the full window size.
		// See: https://docs.microsoft.com/en-us/windows/win32/winmsg/wm-nccalcsize#remarks
		// This hides the titlebar and also disables the resizing from user interaction because the standard frame is not
		// shown. We still need the WS_THICKFRAME style to enable resizing from the frontend.
		if wParam != 0 {
			// Content overflow screen issue when maximizing borderless windows
			// See: https://github.com/MicrosoftEdge/WebView2Feedback/issues/2549
			//isMinimize := uint32(win.GetWindowLong(m.Handle(), win.GWL_STYLE))&win.WS_MINIMIZE != 0
			isMaximize := uint32(win.GetWindowLong(m.Handle(), win.GWL_STYLE))&win.WS_MAXIMIZE != 0
			if isMaximize {
				rect := (*types.TRect)(unsafe.Pointer(lParam))
				// m.Monitor().WorkareaRect(): When minimizing windows and restoring windows on multiple monitors, the main monitor is obtained.
				// Need to obtain correct monitor information to prevent error freezing message loops from occurring
				monitor := win.MonitorFromRect(rect, win.MONITOR_DEFAULTTONULL)
				if monitor != 0 {
					var monitorInfo types.TMonitorInfo
					monitorInfo.CbSize = types.DWORD(unsafe.Sizeof(monitorInfo))
					if win.GetMonitorInfo(monitor, &monitorInfo) {
						*rect = monitorInfo.RcWork
					}
				}
			}
			return 0
		}
	}

	return win.CallWindowProc(m.oldWndPrc, uintptr(hwnd), message, wParam, lParam)
}
func (m *TMainForm) FormCreate(sender lcl.IObject) {
	m.SetCaption("ENERGY 自定义(自绘)控件")
	m.SetPosition(types.PoScreenCenter)
	m.SetWidth(800)
	m.SetHeight(600)
	m.SetDoubleBuffered(true)
	//m.SetColor(colors.ClYellow)
	m.SetColor(colors.RGBToColor(56, 57, 60))

	{
		wndProcCallback := syscall.NewCallback(m.wndProc)
		m.oldWndPrc = win.SetWindowLongPtr(m.Handle(), win.GWL_WNDPROC, wndProcCallback)
	}

	{
		cus := wg.NewButton(m)
		cus.SetParent(m)
		cus.SetShowHint(true)
		cus.SetCaption("上圆角")
		cus.SetHint("上圆角上圆角")
		cus.Font().SetSize(12)
		cus.Font().SetColor(colors.Cl3DFace)
		cus.SetBoundsRect(types.TRect{Left: 50, Top: 50, Right: 250, Bottom: 90})
		cus.SetStartColor(colors.RGBToColor(86, 88, 93))
		cus.SetEndColor(colors.RGBToColor(86, 88, 93))
		cus.RoundedCorner = cus.RoundedCorner.Exclude(wg.RcLeftBottom).Exclude(wg.RcRightBottom)
		cus.SetOnCloseClick(func(sender lcl.IObject) {
			fmt.Println("点击了 X")
		})
		cus.SetIconFavorite("E:\\SWT\\gopath\\src\\github.com\\energye\\workspace\\examples\\lcl\\customtabsheet\\resources\\icon.png")
		cus.SetIconClose("E:\\SWT\\gopath\\src\\github.com\\energye\\workspace\\examples\\lcl\\customtabsheet\\resources\\close.png")

		//cus2 := wg.NewButton(m)
		//cus2.SetParent(m)
		//cus2.SetCaption("大圆角")
		//cus2.SetBoundsRect(types.TRect{Left: 50, Top: 150, Right: 250, Bottom: 220})
		//cus2.SetStartColor(colors.RGBToColor(255, 100, 0))
		//cus2.SetEndColor(colors.RGBToColor(69, 81, 143))
		////cus2.SetEndColor(colors.RGBToColor(180, 0, 0))
		//cus2.Font().SetColor(colors.ClWhite)
		//cus2.SetRadius(20)
		//cus2.SetAlpha(255)
		//
		//cus3 := wg.NewButton(m)
		//cus3.SetParent(m)
		//cus3.SetCaption("小圆角")
		//cus3.SetBoundsRect(types.TRect{Left: 50, Top: 250, Right: 250, Bottom: 320})
		//cus3.SetStartColor(colors.RGBToColor(0, 180, 0))
		//cus3.SetEndColor(colors.RGBToColor(0, 100, 0))
		//cus3.Font().SetColor(colors.ClYellow)
		//cus3.SetRadius(8)
		//cus3.SetAlpha(255)
		//
		//cus4 := wg.NewButton(m)
		//cus4.SetParent(m)
		//cus4.SetCaption("大大圆角")
		//cus4.Font().SetColor(colors.ClWhite)
		//cus4.SetBoundsRect(types.TRect{Left: 50, Top: 350, Right: 250, Bottom: 420})
		//cus4.SetStartColor(colors.RGBToColor(41, 42, 43))
		//cus4.SetEndColor(colors.RGBToColor(80, 81, 82))
		//cus4.SetRadius(35)
		//cus4.SetAlpha(255)
	}
	{
		if false {
			bgColors := []colors.TColor{colors.ClBlue, colors.ClRed, colors.ClGreen, colors.ClYellow}
			go func() {
				i := 0
				for {
					time.Sleep(time.Second)
					lcl.RunOnMainThreadAsync(func(id uint32) {
						m.SetColor(bgColors[i])
					})
					i++
					if i >= len(bgColors) {
						i = 0
					}
				}
			}()
		}
	}

	//{
	//
	//	png := lcl.NewPortableNetworkGraphic()
	//	png.LoadFromFile("D:\\Energy-Doc\\energy-icon-198x198.png")
	//	//png.GetSize(&tabWidth, &tabHeight)
	//	tabWidth, tabHeight := int32(192), int32(56)
	//	tabBuf := lcl.NewBitmap()
	//	//tabBuf.SetPixelFormat(types.Pf32bit)
	//	//tabBuf.SetSize(tabWidth, tabHeight)
	//	//tabBuf.Canvas().StretchDrawWithRectGraphic()
	//
	//	//tabBuf.Canvas().SetAntialiasingMode(types.AmOn)
	//	//tabBuf.Canvas().BrushToBrush().SetColor()
	//	tabBuf.LoadFromFile("E:\\SWT\\gopath\\src\\github.com\\energye\\energy\\examples\\osr\\alienwindow\\1ttt.bmp")
	//	//tabBuf.Assign(png)
	//
	//	tab := lcl.NewPanel(m)
	//	tab.SetParent(m)
	//	// 设置面板默认属性
	//	tab.SetBevelOuter(types.BvNone)
	//	tab.SetParentBackground(false)
	//	tab.SetWidth(tabWidth)
	//	tab.SetHeight(tabHeight)
	//
	//	tab.SetOnPaint(func(sender lcl.IObject) {
	//		fmt.Println("tab.OnPaint")
	//		tab.Canvas().DrawWithIntX2Graphic(0, 0, tabBuf)
	//	})
	//	tab.SetOnResize(func(sender lcl.IObject) {
	//		fmt.Println("tab.OnResize")
	//		tab.Invalidate()
	//	})
	//}
}
