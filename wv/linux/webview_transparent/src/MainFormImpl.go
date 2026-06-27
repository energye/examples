package src

import (
	"fmt"
	"github.com/energye/energy/v3/application"
	"github.com/energye/energy/v3/platform/linux/gtk3"
	gtk3Types "github.com/energye/energy/v3/platform/linux/types"
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
	gtkWindow := gtk3.AsWindow(unsafe.Pointer(gtkHandle.Gtk3Window()))
	//gdkWindow := gtk3.AsGdkWindow(unsafe.Pointer(gtkHandle.Gtk3Window()))
	options := application.GApplication.Options
	if options.WindowTransparent {
		screen := gtkWindow.GetScreen()
		visual := screen.GetRGBAVisual()
		isComposited := screen.IsComposited()
		fmt.Println("isComposited:", visual != nil && isComposited)
		if visual != nil && isComposited {
			gtkWindow.SetVisual(visual)
			gtkWindow.SetAppPaintable(true)
		}
	}
	//m.SetWindowState(types.WsMaximized)
	//gtkWindow.SetDecorated(false)
	fmt.Println(gtkWindow.GetName())
	//gtkWindow.Realize()
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
	gtkContainer := gtk3.AsContainer(unsafe.Pointer(gtkHandle.Gtk3Window()))
	var iterate func(list gtk3Types.IList, level int)
	iterate = func(list gtk3Types.IList, level int) {
		if list == nil {
			return
		}
		for i := uint(0); i < list.Length(); i++ {
			widget := gtk3.AsWidget(list.NthDataRaw(i))
			//x := widget.GetAllocation().GetX()
			//y := widget.GetAllocation().GetY() 注释
			fmt.Println(widget.GetName(), "level:", level)
			if widget.IsContainer() {
				chdWid := gtk3.AsContainer(list.NthDataRaw(i))
				//w, h := chdWid.GetSizeRequest()
				iterate(chdWid.GetChildren(), level+1)
			}
		}
	}
	gtkList := gtkContainer.GetChildren()
	iterate(gtkList, 1)
}
