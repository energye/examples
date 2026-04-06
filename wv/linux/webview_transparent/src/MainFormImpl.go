package src

import (
	"fmt"
	"github.com/energye/energy/v3/application"
	"github.com/energye/energy/v3/pkgs/gtk3"
	"github.com/energye/lcl/lcl"
	"unsafe"
)

type TMainForm struct {
	lcl.TEngForm
}

var MainForm TMainForm

func (m *TMainForm) FormCreate(sender lcl.IObject) {
	fmt.Println("FormCreate")
	m.WorkAreaCenter()
	//m.SetBorderStyleToFormBorderStyle(types.BsNone)

	//m.SetColor(0)
	gtkHandle := lcl.PlatformHandle(m.Handle())
	gtkWindow := gtk3.ToGtkWindow(unsafe.Pointer(gtkHandle.Gtk3Window()))
	options := application.GApplication.Options
	if options.WindowTransparent {
		screen := gtkWindow.GetScreen()
		visual, err := screen.GetRGBAVisual()
		isComposited := screen.IsComposited()
		fmt.Println("isComposited:", err == nil && visual != nil && isComposited)
		if err == nil && visual != nil && isComposited {
			gtkWindow.SetVisual(visual)
			gtkWindow.SetAppPaintable(true)
		}
	}

	//gtkWindow.SetDecorated(false)
	fmt.Println(gtkWindow.TypeFromInstance().Name())

	//lcl.NewButton(m).SetParent(m)
	//lcl.NewPanel(m).SetParent(m)
	//mainMenu := lcl.NewMainMenu(m)
	//mainMenu.Items().Add(lcl.NewMenuItem(m))

}

func (m *TMainForm) OnShow(sender lcl.IObject) {
	fmt.Println("OnShow")
	m.iterateWidget()
}

func (m *TMainForm) iterateWidget() {
	gtkHandle := lcl.PlatformHandle(m.Handle())
	gtkContainer := gtk3.ToContainer(unsafe.Pointer(gtkHandle.Gtk3Window()))
	GtkContainerType := gtk3.TypeFromName("GtkContainer")
	var iterate func(list *gtk3.List, level int)
	iterate = func(list *gtk3.List, level int) {
		if list == nil {
			return
		}
		for i := uint(0); i < list.Length(); i++ {
			widget := gtk3.ToWidget(list.NthDataRaw(i))
			//x := widget.GetAllocation().GetX()
			//y := widget.GetAllocation().GetY()
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
