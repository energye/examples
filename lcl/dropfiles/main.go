package main

import (
	"fmt"
	. "github.com/energye/examples/syso"
	"github.com/energye/lcl/lcl"
)

type TMainForm struct {
	lcl.TEngForm
}

var mainForm TMainForm

func init() {
	TestLoadLibPath()
}
func main() {
	lcl.Init(nil, nil)
	lcl.RunApp(&mainForm)
}

func (f *TMainForm) FormCreate(object lcl.IObject) {
	f.SetCaption("drop files")
	f.SetWidth(300)
	f.SetHeight(200)
	f.ScreenCenter()
	f.EnabledMaximize(false)

	// allow drop file
	f.SetAllowDropFiles(true)

	f.SetOnDropFiles(func(sender lcl.IObject, fileNames lcl.IStringArray) {
		fmt.Println("当前拖放文件事件执行，文件数：", fileNames.Count())
		for i := 0; i < fileNames.Count(); i++ {
			fmt.Println("index:", i, ", filename:", fileNames.GetValue(i))
		}
	})
}
