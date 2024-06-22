package main

import (
	"fmt"
	_ "github.com/energye/examples/syso"
	"github.com/energye/lcl/inits"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
)

type TMainForm struct {
	lcl.TForm
}

var (
	mainForm TMainForm
)

func main() {
	inits.Init(nil, nil)
	lcl.RunApp(&mainForm)
}

func (m *TMainForm) FormCreate(sender lcl.IObject) {
	m.SetCaption("Window Min Max Size")
	m.WorkAreaCenter()
	m.SetWidth(500)
	m.SetHeight(300)
	// 显示当前窗口大小标签
	sizeLabel := lcl.NewLabel(m)
	sizeLabel.SetParent(m)
	sizeLabel.SetTop(50)
	sizeLabel.SetLeft(50)
	// 设置窗口最大最小
	constraints := m.Constraints()
	constraints.SetMinWidth(400)
	constraints.SetMinHeight(200)
	constraints.SetMaxWidth(600)
	constraints.SetMaxHeight(400)
	m.SetOnResize(func(sender lcl.IObject) {
		rect := m.BoundsRect()
		size := fmt.Sprintf("window size, width: %d - height: %d", rect.Width(), rect.Height())
		sizeLabel.SetCaption(size)
	})
}

func (m *TMainForm) OnFormCloseQuery(Sender lcl.IObject, CanClose *bool) {
	*CanClose = lcl.MessageDlg("是否退出？", types.MtConfirmation, types.MbYes, types.MbNo) == types.IdYes
}
