package main

import (
	"github.com/energye/energy/v3/pkgs/gtk3"
	"log"
)

func main() {
	gtk3.Init(nil)

	win, err := gtk3.NewWindow(gtk3.WINDOW_TOPLEVEL)
	if err != nil {
		log.Fatal("Unable to create window:", err)
	}
	win.SetTitle("Simple Example")
	//win.Connect("destroy", func() {
	//	gtk.MainQuit()
	//})

	Toolbar(win)
	l := gtk3.NewLabel("Hello, gotk3!")
	win.Add(l)

	screen := win.GetScreen()
	visual, err := screen.GetRGBAVisual()
	if err == nil && visual != nil && screen.IsComposited() {
		win.SetVisual(visual)
		win.SetAppPaintable(true)
	}
	win.SetDecorated(false)

	win.SetDefaultSize(800, 600)
	win.ShowAll()
	gtk3.Main()
}

func Toolbar(gtkWindow *gtk3.Window) {

	headerBar, err := gtk3.NewHeaderBar()
	if err != nil {
		return
	}
	headerBar.SetShowCloseButton(true)

	gtkWindow.SetTitlebar(headerBar)

	btn := gtk3.NewButton() // .ButtonNewWithLabel("button")
	btn.SetRelief(gtk3.RELIEF_NONE)
	cssProvid := gtk3.NewCssProvider()
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
	btnStyleCtx.AddProvider(cssProvid, gtk3.STYLE_PROVIDER_PRIORITY_APPLICATION)

	headerBar.PackStart(btn)

}
