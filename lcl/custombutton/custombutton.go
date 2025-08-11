package main

import (
	"embed"
	"fmt"
	. "github.com/energye/examples/syso"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
	"github.com/energye/lcl/types/colors"
	"os"
	"path/filepath"
	"time"
	"widget/wg"
)

func init() {
	TestLoadLibPath()
	Chdir("lcl/custombutton")
}

type TMainForm struct {
	lcl.TEngForm
	oldWndPrc uintptr
	box       lcl.IPanel
}

var MainForm TMainForm

//go:embed resources
var resources embed.FS

var (
	wd, _       = os.Getwd()
	examplePath = filepath.Join(wd, "lcl", "custombutton")
)

func main() {
	lcl.Init(nil, nil)
	lcl.Application.Initialize()
	lcl.Application.SetMainFormOnTaskBar(true)
	lcl.Application.NewForm(&MainForm)
	lcl.Application.Run()
}

func (m *TMainForm) FormCreate(sender lcl.IObject) {
	m.SetCaption("ENERGY 自绘(自定义)按钮")
	m.SetPosition(types.PoScreenCenter)
	m.SetWidth(800)
	m.SetHeight(600)
	m.SetDoubleBuffered(true)
	//m.SetColor(colors.ClYellow)
	m.SetColor(colors.RGBToColor(56, 57, 60))

	//m.box = lcl.NewPanel(m)
	//m.box.SetParent(m)
	////m.box.SetAlign(types.AlClient)
	//m.box.SetDoubleBuffered(true)
	//m.box.SetBounds(5, 5, m.Width()-10, m.Height()-10)
	//m.box.SetAnchors(types.NewSet(types.AkLeft, types.AkTop, types.AkRight, types.AkBottom))
	//m.box.SetColor(colors.RGBToColor(56, 57, 60))
	//m.box.SetBevelOuter(types.BvNone)

	{
		click := func(sender lcl.IObject) {
			fmt.Println(lcl.AsGraphicControl(sender).Caption())
		}
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
		cus.SetIconFavorite(filepath.Join(examplePath, "resources", "icon.png"))
		cus.SetIconClose(filepath.Join(examplePath, "resources", "close.png"))
		cus.SetOnClick(click)

		cus2 := wg.NewButton(m)
		cus2.SetParent(m)
		cus2.SetCaption("大圆角")
		cus2.SetBoundsRect(types.TRect{Left: 50, Top: 150, Right: 250, Bottom: 220})
		cus2.SetStartColor(colors.RGBToColor(255, 100, 0))
		cus2.SetEndColor(colors.RGBToColor(69, 81, 143))
		//cus2.SetEndColor(colors.RGBToColor(180, 0, 0))
		cus2.Font().SetColor(colors.ClWhite)
		cus2.SetRadius(20)
		cus2.SetAlpha(255)
		cus2.SetOnClick(click)

		cus3 := wg.NewButton(m)
		cus3.SetParent(m)
		cus3.SetCaption("小圆角")
		cus3.SetBoundsRect(types.TRect{Left: 50, Top: 250, Right: 250, Bottom: 320})
		cus3.SetStartColor(colors.RGBToColor(0, 180, 0))
		cus3.SetEndColor(colors.RGBToColor(0, 100, 0))
		cus3.Font().SetColor(colors.ClYellow)
		cus3.SetRadius(8)
		cus3.SetAlpha(255)
		cus3.SetOnClick(click)

		cus4 := wg.NewButton(m)
		cus4.SetParent(m)
		cus4.SetCaption("大大圆角")
		cus4.Font().SetColor(colors.ClWhite)
		cus4.SetBoundsRect(types.TRect{Left: 50, Top: 350, Right: 250, Bottom: 420})
		cus4.SetStartColor(colors.RGBToColor(41, 42, 43))
		cus4.SetEndColor(colors.RGBToColor(80, 81, 82))
		cus4.SetRadius(35)
		cus4.SetAlpha(255)
		cus4.SetOnClick(click)

		cus5 := wg.NewButton(m)
		cus5.SetParent(m)
		cus5.SetCaption("X")
		cus5.Font().SetColor(colors.ClWhite)
		cus5.Font().SetSize(14)
		rect5 := types.TRect{Left: 50, Top: 450}
		rect5.SetSize(50, 50)
		cus5.SetBoundsRect(rect5)
		cus5.SetStartColor(colors.ClBlue)
		cus5.SetEndColor(colors.ClYellow)
		cus5.SetRadius(35)
		cus5.SetAlpha(255)
		cus5.SetOnClick(click)

		cus6 := wg.NewButton(m)
		cus6.SetParent(m)
		cus6.SetCaption("< X >")
		cus6.Font().SetColor(colors.ClWhite)
		cus6.Font().SetSize(14)
		rect6 := types.TRect{Left: 150, Top: 450}
		rect6.SetSize(50, 50)
		cus6.SetBoundsRect(rect6)
		cus6.SetStartColor(colors.ClGray)
		cus6.SetEndColor(colors.ClLtGray)
		cus6.SetRadius(5)
		cus6.SetAlpha(255)
		cus6.SetOnClick(click)

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
