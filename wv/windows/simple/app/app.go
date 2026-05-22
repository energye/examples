// ==============================================================================
// 📚 窗体维护列表
// 🔥 ENERGY GUI 设计器自动生成代码. 不能编辑
// ==============================================================================

package app

import (
	"github.com/energye/energy/v3/application/pack"
	"github.com/energye/lcl/api/libname"
	"github.com/energye/lcl/lcl"
	"runtime"
)

var (
	// Info app pack info
	Info = pack.Info
)

// Forms 应用使用的窗体列表
var Forms = []lcl.IEngForm{
	&Form1,
}

func init() {
	if runtime.GOOS == "linux" {
		// linux webkit2 > gtk3
		libname.UseWS = "gtk3"
	}
	if runtime.GOOS == "darwin" {
		// macOS universal-binary
		// os.Setenv("ENERGY_UNIVERSAL_BINARY", "universal")
	}
}
