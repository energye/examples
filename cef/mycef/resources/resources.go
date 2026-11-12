// ==============================================================================
// embedded resource
// ==============================================================================

package resources

import (
	"embed"
	"github.com/energye/lcl/emfs"
	"github.com/energye/lcl/lcl"
)

//go:embed embed
var icon embed.FS

// Embed Get embedded resources
// Function signature cannot be modified
func Embed(fileName string) []byte {
	data, _ := icon.ReadFile("embed/" + fileName)
	return data
}

// SetIcon Set application icon
// Function signature cannot be modified
func SetIcon() {
	stream := lcl.NewMemoryStream()
	lcl.StreamHelper.Write(stream, Embed("icon.png"))
	stream.SetPosition(0)
	png := lcl.NewPortableNetworkGraphic()
	png.LoadFromStreamWithStream(stream)
	lcl.Application.Icon().Assign(png)
	png.Free()
	stream.Free()
}

func init() {
	emfs.RegisterEmbedFS(emfs.FSName, icon)
	lcl.SetOnBeforeRun(SetIcon)
}
