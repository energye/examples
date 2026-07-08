// Package widget defines the core widget lifecycle used by the UI engine.
package widget

import (
	"github.com/energye/examples/lcl/gpui/core/math"
	"github.com/energye/examples/lcl/gpui/overlay"
	"github.com/energye/examples/lcl/gpui/render/font"
	"github.com/energye/examples/lcl/gpui/render/pipeline"
	"github.com/energye/examples/lcl/gpui/style/token"
)

// State stores common widget state flags.
type State uint32

const (
	StateNormal State = 0
	StateHover  State = 1 << iota
	StateActive
	StateFocus
	StateDisabled
	StateSelected
	StateChecked
	StateLoading
	StateError
	StateWarning
	StateSuccess
)

// EventType identifies a high-level UI event.
type EventType int

const (
	EventMouseDown EventType = iota
	EventMouseUp
	EventMouseMove
	EventMouseEnter
	EventMouseLeave
	EventMouseWheel
	EventDoubleClick
	EventDragStart
	EventDragMove
	EventDragEnd
	EventKeyDown
	EventCharInput
)

// Event is the common event payload used by the widget lifecycle.
type Event struct {
	Type   EventType
	X, Y   float32
	LocalX float32
	LocalY float32
	Button int
	Key    int
	Mods   int
	Char   rune
	DeltaX float32
	DeltaY float32
	Clicks int
}

// UIEvent is kept as a compatibility alias for engine callers.
type UIEvent = Event

// Constraints describes the available size during measurement.
type Constraints struct {
	Min math.Vec2
	Max math.Vec2
}

// Context contains dependencies needed by widget lifecycle methods.
type Context struct {
	Renderer *pipeline.Renderer
	Tokens   token.Tokens
	Font     *font.Font
	Overlay  *overlay.Manager
	Viewport math.Rect
	Scale    float32
}

// Widget is the common lifecycle interface implemented by every UI node.
type Widget interface {
	Measure(ctx *Context, constraints Constraints) math.Vec2
	Layout(ctx *Context, rect math.Rect)
	Render(ctx *Context)
	HitTest(point math.Vec2) Widget
	HandleEvent(ctx *Context, event Event) bool

	Bounds() math.Rect
	SetBounds(rect math.Rect)
	Parent() Widget
	SetParent(parent Widget)

	Visible() bool
	SetVisible(visible bool)
	Enabled() bool
	SetEnabled(enabled bool)

	Focusable() bool
	SetFocusable(focusable bool)
	Focused() bool
	Focus()
	Blur()

	State() State
	SetState(state State)
	HasState(state State) bool
	SetStateFlag(state State, enabled bool)
}

// ClampSize clamps a measured size to constraints.
func ClampSize(size math.Vec2, constraints Constraints) math.Vec2 {
	if constraints.Min.X > 0 && size.X < constraints.Min.X {
		size.X = constraints.Min.X
	}
	if constraints.Min.Y > 0 && size.Y < constraints.Min.Y {
		size.Y = constraints.Min.Y
	}
	if constraints.Max.X > 0 && size.X > constraints.Max.X {
		size.X = constraints.Max.X
	}
	if constraints.Max.Y > 0 && size.Y > constraints.Max.Y {
		size.Y = constraints.Max.Y
	}
	return size
}
