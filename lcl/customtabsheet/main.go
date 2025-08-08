package main

import (
	"embed"
	"fmt"
	. "github.com/energye/examples/syso"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
	"github.com/energye/lcl/types/colors"
	"time"
	"widget/wg"
)

func init() {
	TestLoadLibPath()
	Chdir("lcl/action")
}

type TMainForm struct {
	lcl.TEngForm
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

func (m *TMainForm) FormCreate(sender lcl.IObject) {
	m.SetCaption("ENERGY 自定义(自绘)控件")
	m.SetPosition(types.PoScreenCenter)
	m.SetWidth(800)
	m.SetHeight(600)
	m.SetDoubleBuffered(true)
	//m.SetColor(colors.ClYellow)
	m.SetColor(colors.RGBToColor(143, 143, 143))

	{
		cus := wg.NewButton(m)
		cus.SetParent(m)
		cus.SetCaption("上圆角")
		cus.SetShowHint(true)
		cus.SetHint("上圆角上圆角")
		cus.Font().SetSize(12)
		cus.Font().SetColor(colors.Cl3DFace)
		cus.SetBoundsRect(types.TRect{Left: 50, Top: 50, Right: 250, Bottom: 100})
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
