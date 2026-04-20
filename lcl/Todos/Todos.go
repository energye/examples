package main

import (
	"fmt"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
)

type TTodosForm struct {
	lcl.TEngForm
	InputEdit    lcl.IEdit
	AddButton    lcl.IButton
	TodoList     lcl.IListBox
	RemoveButton lcl.IButton
	ClearButton  lcl.IButton
}

var TodosForm TTodosForm

func main() {
	lcl.Init(nil, nil)
	lcl.Application.Initialize()
	lcl.Application.SetMainFormOnTaskBar(true)
	lcl.Application.NewForms(&TodosForm)
	lcl.Application.Run()
}

func (t *TTodosForm) FormCreate(sender lcl.IObject) {
	t.SetCaption("待办事项列表")
	t.SetPosition(types.PoScreenCenter)
	t.SetWidth(500)
	t.SetHeight(600)

	label := lcl.NewLabel(t)
	label.SetParent(t)
	label.SetCaption("添加新任务:")
	label.SetLeft(20)
	label.SetTop(20)
	label.SetWidth(100)
	label.SetHeight(25)

	t.InputEdit = lcl.NewEdit(t)
	t.InputEdit.SetParent(t)
	t.InputEdit.SetLeft(20)
	t.InputEdit.SetTop(50)
	t.InputEdit.SetWidth(350)
	t.InputEdit.SetHeight(30)

	t.AddButton = lcl.NewButton(t)
	t.AddButton.SetParent(t)
	t.AddButton.SetCaption("添加")
	t.AddButton.SetLeft(380)
	t.AddButton.SetTop(50)
	t.AddButton.SetWidth(90)
	t.AddButton.SetHeight(30)
	t.AddButton.SetOnClick(t.OnAddClick)

	t.TodoList = lcl.NewListBox(t)
	t.TodoList.SetParent(t)
	t.TodoList.SetLeft(20)
	t.TodoList.SetTop(100)
	t.TodoList.SetWidth(460)
	t.TodoList.SetHeight(400)

	t.RemoveButton = lcl.NewButton(t)
	t.RemoveButton.SetParent(t)
	t.RemoveButton.SetCaption("删除选中")
	t.RemoveButton.SetLeft(20)
	t.RemoveButton.SetTop(520)
	t.RemoveButton.SetWidth(120)
	t.RemoveButton.SetHeight(35)
	t.RemoveButton.SetOnClick(t.OnRemoveClick)

	t.ClearButton = lcl.NewButton(t)
	t.ClearButton.SetParent(t)
	t.ClearButton.SetCaption("清空全部")
	t.ClearButton.SetLeft(360)
	t.ClearButton.SetTop(520)
	t.ClearButton.SetWidth(120)
	t.ClearButton.SetHeight(35)
	t.ClearButton.SetOnClick(t.OnClearClick)
}

func (t *TTodosForm) OnAddClick(sender lcl.IObject) {
	text := t.InputEdit.Text()
	if text != "" {
		t.TodoList.Items().Add(text)
		t.InputEdit.SetText("")
		t.InputEdit.SetFocus()
	}
}

func (t *TTodosForm) OnRemoveClick(sender lcl.IObject) {
	index := t.TodoList.ItemIndex()
	if index >= 0 {
		t.TodoList.Items().Delete(index)
	}
}

func (t *TTodosForm) OnClearClick(sender lcl.IObject) {
	t.TodoList.Items().Clear()
}

func (t *TTodosForm) FormCloseQuery(sender lcl.IObject, canClose *bool) {
	count := t.TodoList.Items().Count()
	if count > 0 {
		fmt.Printf("您还有 %d 个待办事项未完成\n", count)
	}
}
