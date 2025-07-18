package main

import (
	"fmt"
	"github.com/energye/examples/lcl/grids/drawgrid/form"
	. "github.com/energye/examples/syso"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
	"math/rand"
)

func init() {
	TestLoadLibPath()
}

func main() {
	lcl.Init(nil, nil)
	lcl.Application.Initialize()
	lcl.Application.SetMainFormOnTaskBar(true)

	var mainForm lcl.TEngForm
	lcl.Application.NewForm(&mainForm)
	mainForm.SetWidth(700)
	mainForm.SetHeight(500)
	mainForm.WorkAreaCenter()
	mainForm.SetCaption("表格自绘")
	ScaleSelf(&mainForm)
	grid := form.NewPlayControl(&mainForm)
	grid.SetParent(&mainForm)
	grid.SetAlign(types.AlClient)
	for i := 1; i <= 100; i++ {
		grid.Add(form.TPlayListItem{Caption: fmt.Sprintf("标题%d", i), Singer: "张三", Length: 100000 + rand.Int31n(100000)})
	}
	lcl.Application.Run()
}

// ScaleSelf : 这个方法主要是用于当不使用资源窗口创建时用，这个方法要用于设置了Width, Height或者ClientWidth、ClientHeight之后
func ScaleSelf(f lcl.IEngForm) {
	if lcl.Application.Scaled() {
		f.SetClientWidth(int32(float64(f.ClientWidth()) * (float64(lcl.Screen.PixelsPerInch()) / 96.0)))
		f.SetClientHeight(int32(float64(f.ClientHeight()) * (float64(lcl.Screen.PixelsPerInch()) / 96.0)))
	}
}
