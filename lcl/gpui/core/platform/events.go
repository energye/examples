// Package platform provides platform abstraction for events and windows
package platform

// MouseButton represents a mouse button
type MouseButton int

const (
	MouseButtonLeft   MouseButton = 0
	MouseButtonMiddle MouseButton = 1
	MouseButtonRight  MouseButton = 2
)

// KeyAction represents a key action
type KeyAction int

const (
	KeyActionPress   KeyAction = 0
	KeyActionRelease KeyAction = 1
	KeyActionRepeat  KeyAction = 2
)

// KeyModifier represents keyboard modifiers
type KeyModifier int

const (
	ModShift   KeyModifier = 1 << 0
	ModControl KeyModifier = 1 << 1
	ModAlt     KeyModifier = 1 << 2
	ModSuper   KeyModifier = 1 << 3
)

// Key represents a keyboard key
type Key int

const (
	KeyNone Key = iota

	// Printable keys
	KeySpace
	KeyApostrophe
	KeyComma
	KeyMinus
	KeyPeriod
	KeySlash
	Key0
	Key1
	Key2
	Key3
	Key4
	Key5
	Key6
	Key7
	Key8
	Key9
	KeySemicolon
	KeyEqual
	KeyA
	KeyB
	KeyC
	KeyD
	KeyE
	KeyF
	KeyG
	KeyH
	KeyI
	KeyJ
	KeyK
	KeyL
	KeyM
	KeyN
	KeyO
	KeyP
	KeyQ
	KeyR
	KeyS
	KeyT
	KeyU
	KeyV
	KeyW
	KeyX
	KeyY
	KeyZ
	KeyLeftBracket
	KeyBackslash
	KeyRightBracket
	KeyGraveAccent

	// Function keys
	KeyEscape
	KeyEnter
	KeyTab
	KeyBackspace
	KeyInsert
	KeyDelete
	KeyRight
	KeyLeft
	KeyDown
	KeyUp
	KeyPageUp
	KeyPageDown
	KeyHome
	KeyEnd
	KeyCapsLock
	KeyScrollLock
	KeyNumLock
	KeyPrintScreen
	KeyPause
	KeyF1
	KeyF2
	KeyF3
	KeyF4
	KeyF5
	KeyF6
	KeyF7
	KeyF8
	KeyF9
	KeyF10
	KeyF11
	KeyF12

	// Keypad
	KeyKP0
	KeyKP1
	KeyKP2
	KeyKP3
	KeyKP4
	KeyKP5
	KeyKP6
	KeyKP7
	KeyKP8
	KeyKP9
	KeyKPDecimal
	KeyKPDivide
	KeyKPMultiply
	KeyKPSubtract
	KeyKPAdd
	KeyKPEnter
	KeyKPEqual

	KeyLeftShift
	KeyLeftControl
	KeyLeftAlt
	KeyLeftSuper
	KeyRightShift
	KeyRightControl
	KeyRightAlt
	KeyRightSuper
)

// MouseEvent represents a mouse event
type MouseEvent struct {
	X, Y   float32
	Button MouseButton
	Action KeyAction
	Mods   KeyModifier
}

// KeyEvent represents a keyboard event
type KeyEvent struct {
	Key    Key
	Action KeyAction
	Mods   KeyModifier
}

// CharEvent represents a character input event
type CharEvent struct {
	Char rune
}

// ScrollEvent represents a scroll event
type ScrollEvent struct {
	XOffset, YOffset float32
}

// WindowResizeEvent represents a window resize event
type WindowResizeEvent struct {
	Width, Height int
}

// EventHandler interface for handling events
type EventHandler interface {
	OnMouseMove(event MouseEvent)
	OnMouseButton(event MouseEvent)
	OnKey(event KeyEvent)
	OnChar(event CharEvent)
	OnScroll(event ScrollEvent)
	OnResize(event WindowResizeEvent)
}
