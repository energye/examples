package src

import (
	"fmt"
	"github.com/energye/energy/v3/pkgs/gtk3"
	"github.com/energye/lcl/lcl"
)

type TMainForm struct {
	lcl.TEngForm
}

var MainForm TMainForm

func (m *TMainForm) FormCreate(sender lcl.IObject) {
	fmt.Println("FormCreate")
	m.WorkAreaCenter()
	m.SetVisible(false)
	//m.SetColor(0)
	gtkHandle := lcl.PlatformHandle(m.Handle())
	gtkWindow := gtk3.ToGtkWindow(uintptr(gtkHandle.Gtk3Window()))
	gtkWindow.SetDecorated(false)
	fmt.Println(gtkWindow.TypeFromInstance().Name())
	screen := gtkWindow.GetScreen()
	visual, err := screen.GetRGBAVisual()
	if err == nil && visual != nil && screen.IsComposited() {
		gtkWindow.SetVisual(visual)
		gtkWindow.SetAppPaintable(true)
	}
	//m.SetOnPaint(func(sender lcl.IObject) {
	//	m.Canvas().SetColors(0, 0, lcl.TFPColor{})
	//})
}

func (m *TMainForm) OnShow(sender lcl.IObject) {
	fmt.Println("OnShow")
	gtkHandle := lcl.PlatformHandle(m.Handle())
	gtkContainer := gtk3.ToGtkContainer(uintptr(gtkHandle.Gtk3Window()))
	gtkList := gtkContainer.GetChildren()
	fmt.Println(gtkList.Length())
	for i := uint(0); i < gtkList.Length(); i++ {
		chdWid := gtk3.ToGtkBox(uintptr(gtkList.NthDataRaw(i)))
		fmt.Println(chdWid.TypeFromInstance().Name())
		chdWid.SetSizeRequest(50, 50)
	}
}
