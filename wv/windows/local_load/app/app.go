// ==============================================================================
// 📚 窗体维护列表
// 🔥 ENERGY GUI 设计器自动生成代码. 不能编辑
// ==============================================================================

package app

import (
	"github.com/energye/lcl/api/libname"
	"github.com/energye/lcl/lcl"
)

// Forms 应用使用的窗体列表
var Forms = []lcl.IEngForm{
	&Form1Window,
}

func init() {
	// linux webkit2 > gtk3
	libname.UseWS = "gtk3"
}
