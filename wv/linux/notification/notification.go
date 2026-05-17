package main

import (
	"embed"
	_ "github.com/energye/examples/syso"
	"github.com/energye/examples/wv/linux/notification/src"
	"github.com/energye/lcl/api"
	"github.com/energye/lcl/emfs"
	"github.com/energye/lcl/lcl"
	"os"
)

func main() {
	api.SetDebug(true)
	os.Setenv("--ws", "gtk3")
	lcl.Init()
	println(api.Widget().IsGTK3())
	lcl.Application.Initialize()
	lcl.Application.SetMainFormOnTaskBar(true)
	SetIcon()
	lcl.Application.NewForm(&src.MainForm)
	lcl.Application.Run()
}

//go:embed embed/icon.png
var icon []byte

//go:embed embed
var iconFS embed.FS

func SetIcon() {
	stream := lcl.NewMemoryStream()
	lcl.StreamHelper.Write(stream, icon)
	stream.SetPosition(0)
	png := lcl.NewPortableNetworkGraphic()
	png.LoadFromStreamWithStream(stream)
	lcl.Application.Icon().Assign(png)
	png.Free()
	stream.Free()

}

func init() {
	emfs.RegisterEmbedFS(emfs.FSName, iconFS)
}
