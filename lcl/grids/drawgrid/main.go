package main

import (
	"fmt"
	"github.com/energye/examples/lcl/grids/drawgrid/form"
	_ "github.com/energye/examples/syso"
	"github.com/energye/lcl/inits"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
	"math/rand"
)

func main() {
	inits.Init(nil, nil)
	lcl.Application.Initialize()
	lcl.Application.SetMainFormOnTaskBar(true)

	mainForm := lcl.Application.CreateForm()
	mainForm.SetWidth(700)
	mainForm.SetHeight(500)
	mainForm.WorkAreaCenter()
	mainForm.SetCaption("表格自绘")
	mainForm.ScaleSelf()
	grid := form.NewPlayControl(mainForm)
	grid.SetParent(mainForm)
	grid.SetAlign(types.AlClient)
	for i := 1; i <= 100; i++ {
		grid.Add(form.TPlayListItem{fmt.Sprintf("标题%d", i), "张三", 100000 + rand.Int31n(100000), "", ""})

	}
	lcl.Application.Run()
}
