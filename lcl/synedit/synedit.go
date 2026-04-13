package main

import (
	_ "github.com/energye/examples/syso"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
	"github.com/energye/lcl/types/font"
	"os"
	"strings"
)

func main() {
	os.Setenv("--ws", "gtk3")
	lcl.Init(nil, nil)
	lcl.Application.Initialize()
	lcl.Application.SetMainFormOnTaskBar(true)
	lcl.Application.NewForm(&MainForm)
	lcl.Application.Run()
}

type TMainForm struct {
	lcl.TEngForm
	SynEdit   lcl.ISynEdit
	SynEdit2  lcl.ISynEdit
	SynAnySyn lcl.ISynAnySyn
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
	synEditFont := m.SynEdit.Font()
	synEditFont.SetName("Menlo") // Macos
	synEditFont.SetQuality(types.FqCleartype)
	synEditFont.SetCharSet(font.DEFAULT_CHARSET)
	//m.SynEdit.Canvas().SetFontToFont(synEditFont)

	m.SynEdit2 = lcl.NewSynEdit(m)
	m.SynEdit2.SetParent(m)
	m.SynEdit2.SetWidth(m.Width())
	m.SynEdit2.SetHeight(300)
	m.SynEdit2.SetTop(300)
	m.SynEdit2.SetAnchors(types.NewSet(types.AkLeft, types.AkTop, types.AkBottom, types.AkRight))
	m.SynEdit2.SetAlign(types.AlBottom)
	m.SynEdit2.SetBracketHighlightStyle(types.SbhsBoth)
	m.SynEdit2.SetHighlighter(lcl.NewSynHTMLSyn(m))

	m.SynAnySyn = lcl.NewSynAnySyn(m)
	m.SynAnySyn.KeyWords().Clear()
	m.SynAnySyn.KeyWords().SetTextToStr("example")
	m.SynAnySyn.Constants().Clear()
	m.SynAnySyn.Constants().SetTextToStr("highlighter")
	constants := lcl.NewStringList()
	constants.Add("reprehenderit")
	constants.Add("elit")
	m.SynAnySyn.Constants().SetStringsWithStrings(constants)
	m.SynEdit.SetHighlighter(m.SynAnySyn)

	lines := strings.Split("This is an example how to write a highlighter from scratch.\nSee the units for each example highlighter.\n--\nThis is NOT about extending the IDE. This is about SynEdit and it's Highlighter only.\nTherefore this does not include: ++\n- registration in the component pallette.\n- Using the Object Inspector\nThose steps are the same as they would be for any other self wriiten componont.\n\n-(-\nLorem ipsum dolor sit amet, \nconsectetur adipisicing elit, \nsed -- do eiusmod tempor incididunt ut labore et dolore magna aliqua. \n-- (Nested double -) Ut enim ad minim veniam, \nquis nostrud exercitation ++ (Nested double +) ullamco laboris nisi ut aliquip ex ea commodo consequat. \nDuis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. \n++ Excepteur sint occaecat cupidatat non proident, \nsunt in culpa qui officia deserunt mollit anim id est laborum.\n-(- --\nLorem ipsum dolor sit amet, \nconsectetur adipisicing elit, \nsed do eiusmod tempor incididunt ut labore et dolore magna aliqua. \n++ Ut enim ad minim veniam, \nquis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. \nDuis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. \nExcepteur sint occaecat cupidatat non proident, \nsunt in culpa qui officia deserunt mollit anim id est laborum.\n-)-\n-)-\n\n", "\n")
	for _, line := range lines {
		m.SynEdit.Lines().Add(line)
	}
}
