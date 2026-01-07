// ==============================================================================
// 📚 内嵌资源
// 📌 不存在时自动创建
// ✏️ 可在此文件中添加业务逻辑
// ==============================================================================

package resources

import (
	"embed"
	"github.com/energye/lcl/lcl"
)

//go:embed embed
var icon embed.FS

// Embed 获取内嵌资源
// 函数签名不能修改
func Embed(fileName string) []byte {
	data, _ := icon.ReadFile("embed/" + fileName)
	return data
}

// SetIcon 设置应用程序图标
// 函数签名不能修改
//
//	app - 应用程序接口对象, 用于设置图标
func SetIcon() {
	stream := lcl.NewMemoryStream()
	lcl.StreamHelper.Write(stream, Embed("icon.png"))
	stream.SetPosition(0)
	png := lcl.NewPortableNetworkGraphic()
	png.LoadFromStreamWithStream(stream)
	//lcl.Application.Icon().Assign(png)
	png.Free()
	stream.Free()
}
