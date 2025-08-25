package toolbar

type IControl interface {
	Instance() Pointer
	Owner() *NSToolBar
	Property() *ControlProperty
	Identifier() string
}

type Control struct {
	item     ItemBase
	owner    *NSToolBar
	instance Pointer
	property *ControlProperty
}

func (m *Control) Identifier() string {
	return m.item.Identifier
}

func (m *Control) Instance() Pointer {
	return m.instance
}

func (m *Control) Owner() *NSToolBar {
	return m.owner
}

func (m *Control) Property() *ControlProperty {
	return m.property
}
