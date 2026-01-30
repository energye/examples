package src

import (
	"fmt"
	"github.com/energye/energy/v3/pkgs/gtk3"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
	"unsafe"
)

type TMainForm struct {
	lcl.TEngForm
}

var MainForm TMainForm

func (m *TMainForm) FormCreate(sender lcl.IObject) {
	fmt.Println("FormCreate")
	m.WorkAreaCenter()
	m.SetBorderStyleToFormBorderStyle(types.BsNone)

	//m.SetColor(0)
	gtkHandle := lcl.PlatformHandle(m.Handle())
	gtkWindow := gtk3.ToGtkWindow(unsafe.Pointer(gtkHandle.Gtk3Window()))
	//gtkWindow.SetDecorated(false)
	fmt.Println(gtkWindow.TypeFromInstance().Name())
	screen := gtkWindow.GetScreen()
	visual, err := screen.GetRGBAVisual()
	if err == nil && visual != nil && screen.IsComposited() {
		//gtkWindow.SetVisual(visual)
		//gtkWindow.SetAppPaintable(true)
	}
	//m.SetOnPaint(func(sender lcl.IObject) {
	//	m.Canvas().SetColors(0, 0, lcl.TFPColor{})
	//})
	lcl.NewButton(m).SetParent(m)
	lcl.NewPanel(m).SetParent(m)

	gtkContainer := gtk3.ToContainer(unsafe.Pointer(gtkHandle.Gtk3Window()))
	GtkContainerType := gtk3.TypeFromName("GtkContainer")
	var iterate func(list *gtk3.List, level int)
	iterate = func(list *gtk3.List, level int) {
		if list == nil {
			return
		}
		for i := uint(0); i < list.Length(); i++ {
			widget := gtk3.ToWidget(list.NthDataRaw(i))
			fmt.Println(widget.TypeFromInstance().Name(), "level:", level)
			if widget.IsA(GtkContainerType) {
				chdWid := gtk3.ToContainer(list.NthDataRaw(i))
				//w, h := chdWid.GetSizeRequest()
				iterate(chdWid.GetChildren(), level+1)
			}
		}
	}
	gtkList := gtkContainer.GetChildren()
	iterate(gtkList, 1)
}

func (m *TMainForm) OnShow(sender lcl.IObject) {
	fmt.Println("OnShow")
}
