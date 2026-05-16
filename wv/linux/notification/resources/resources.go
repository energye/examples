// ==============================================================================
// embedded resource
// ==============================================================================

package resources

import (
	_ "embed"
	"github.com/energye/lcl/lcl"
)

//go:embed icon.png
var icon []byte

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
