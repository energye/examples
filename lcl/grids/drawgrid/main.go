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
