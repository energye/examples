// ==============================================================================
// 📚 窗体维护列表
// 🔥 ENERGY GUI 设计器自动生成代码. 不能编辑
// ==============================================================================

package app

import (
	"github.com/energye/energy/v3/application/pack"
	"github.com/energye/lcl/lcl"
	"os"
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
	if "WV" == "WV" {
		// linux webkit2 > gtk3
		os.Setenv("--ws", "gtk3")
	}
	if runtime.GOOS == "darwin" {
		// macOS universal-binary
		// os.Setenv("--universal-binary", "universal")
	}
}
