package main

import (
	_ "github.com/energye/examples/syso"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
)

func main() {
	lcl.Init(nil, nil)
	lcl.Application.Initialize()
	lcl.Application.SetMainFormOnTaskBar(true)
	lcl.Application.NewForm(&MainForm)
	lcl.Application.Run()
}

type TMainForm struct {
	lcl.TEngForm
	SynEdit lcl.ISynEdit
}

var MainForm TMainForm

func (m *TMainForm) FormCreate(sender lcl.IObject) {
	m.SetCaption("ENERGY SynEdit Completion")
	m.SetWidth(800)
	m.SetHeight(600)
	m.SetPosition(types.PoScreenCenter)
	m.SynEdit = lcl.NewSynEdit(m)
	m.SynEdit.SetParent(m)
	m.SynEdit.SetWidth(m.Width())
	m.SynEdit.SetHeight(300)
	m.SynEdit.SetAnchors(types.NewSet(types.AkLeft, types.AkTop, types.AkRight))
	m.SynEdit.SetAlign(types.AlTop)
	m.SynEdit.SetBracketHighlightStyle(types.SbhsBoth)
	m.SynEdit.Keystrokes()
}
