package toolbar

type IView interface {
	Instance() Pointer
	Identifier() string
}

type View struct {
	instance Pointer
	item     ItemBase
}

func (m *View) Instance() Pointer {
	return m.instance
}

func (m *View) Identifier() string {
	return m.item.Identifier
}
