package toolbar

type IControl interface {
	Instance() uintptr
	Owner() *NSToolBar
	Property() *ControlProperty
	Identifier() string
}

type Control struct {
	item ItemBase
	//type_    ControlType
	owner    *NSToolBar
	instance Pointer
	property *ControlProperty
}

func (m *Control) Identifier() string {
	return m.item.Identifier
}

//func (m *Control) IsCocoa() bool {
//	return m.type_ == CtCocoa
//}
//
//func (m *Control) IsLCL() bool {
//	return m.type_ == CtLCL
//}

func (m *Control) Instance() uintptr {
	return uintptr(m.instance)
}

func (m *Control) Owner() *NSToolBar {
	return m.owner
}

func (m *Control) Property() *ControlProperty {
	return m.property
}
