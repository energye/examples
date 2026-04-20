package main

import (
	"github.com/energye/lcl/api"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
	"os"
)

type TNotepadForm struct {
	lcl.TEngForm
	Memo     lcl.IMemo
	FileName string
}

var NotepadForm TNotepadForm

func main() {
	lcl.Init(nil, nil)
	lcl.Application.Initialize()
	lcl.Application.SetMainFormOnTaskBar(true)
	lcl.Application.NewForms(&NotepadForm)
	lcl.Application.Run()
}

func (n *TNotepadForm) FormCreate(sender lcl.IObject) {
	n.SetCaption("记事本 - 无标题")
	n.SetPosition(types.PoScreenCenter)
	n.SetWidth(800)
	n.SetHeight(600)

	menu := lcl.NewMainMenu(n)

	fileMenu := lcl.NewMenuItem(menu)
	fileMenu.SetCaption("文件(&F)")
	menu.Items().Add(fileMenu)

	newItem := lcl.NewMenuItem(menu)
	newItem.SetCaption("新建(&N)")
	newItem.SetOnClick(n.OnNewClick)
	fileMenu.Add(newItem)

	openItem := lcl.NewMenuItem(menu)
	openItem.SetCaption("打开(&O)")
	openItem.SetOnClick(n.OnOpenClick)
	fileMenu.Add(openItem)

	saveItem := lcl.NewMenuItem(menu)
	saveItem.SetCaption("保存(&S)")
	saveItem.SetOnClick(n.OnSaveClick)
	fileMenu.Add(saveItem)

	f := lcl.NewMenuItem(menu)
	f.SetCaption("-")
	fileMenu.Add(f)

	exitItem := lcl.NewMenuItem(menu)
	exitItem.SetCaption("退出(&X)")
	exitItem.SetOnClick(func(sender lcl.IObject) {
		lcl.Application.Terminate()
	})
	fileMenu.Add(exitItem)

	editMenu := lcl.NewMenuItem(menu)
	editMenu.SetCaption("编辑(&E)")
	menu.Items().Add(editMenu)

	selectAllItem := lcl.NewMenuItem(menu)
	selectAllItem.SetCaption("全选(&A)")
	selectAllItem.SetOnClick(n.OnSelectAllClick)
	editMenu.Add(selectAllItem)

	n.Memo = lcl.NewMemo(n)
	n.Memo.SetParent(n)
	n.Memo.SetLeft(0)
	n.Memo.SetTop(0)
	n.Memo.SetWidth(n.Width())
	n.Memo.SetHeight(n.Height())
	n.Memo.SetAlign(types.AlClient)
	n.Memo.SetScrollBars(types.SsBoth)
	n.Memo.Font().SetName("Courier New")
	n.Memo.Font().SetSize(12)
}

func (n *TNotepadForm) OnNewClick(sender lcl.IObject) {
	n.Memo.Lines().Clear()
	n.FileName = ""
	n.SetCaption("记事本 - 无标题")
}

func (n *TNotepadForm) OnOpenClick(sender lcl.IObject) {
	openDialog := lcl.NewOpenDialog(n)
	openDialog.SetFilter("文本文件 (*.txt)|*.txt|所有文件 (*.*)|*.*")

	if openDialog.Execute() {
		content, err := os.ReadFile(openDialog.FileName())
		if err == nil {
			n.Memo.Lines().SetTextToStr(string(content))
			n.FileName = openDialog.FileName()
			n.SetCaption("记事本 - " + n.FileName)
		}
	}
}

func (n *TNotepadForm) OnSaveClick(sender lcl.IObject) {
	if n.FileName == "" {
		saveDialog := lcl.NewSaveDialog(n)
		saveDialog.SetFilter("文本文件 (*.txt)|*.txt|所有文件 (*.*)|*.*")
		saveDialog.SetDefaultExt(".txt")

		if saveDialog.Execute() {
			n.FileName = saveDialog.FileName()
		} else {
			return
		}
	}

	content := api.GoStr(n.Memo.Lines().GetText())
	err := os.WriteFile(n.FileName, []byte(content), 0644)
	if err == nil {
		n.SetCaption("记事本 - " + n.FileName)
	}
}

func (n *TNotepadForm) OnSelectAllClick(sender lcl.IObject) {
	n.Memo.SelectAll()
}
