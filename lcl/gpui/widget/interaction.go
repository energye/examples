package widget

const (
	keyEnter      = 13
	keySpace      = 32
	keyEscape     = 27
	keyTab        = 9
	keyBackspace  = 8
	keyDelete     = 46
	keyArrowLeft  = 37
	keyArrowUp    = 38
	keyArrowRight = 39
	keyArrowDown  = 40
	keyHome       = 36
	keyEnd        = 35
	keyPageUp     = 33
	keyPageDown   = 34
)

// InteractionOptions controls the default pointer and keyboard behavior.
type InteractionOptions struct {
	Pointer               bool
	Keyboard              bool
	ClickOnMouseDown      bool
	ClickOnMouseUp        bool
	EnterActivates        bool
	SpaceActivates        bool
	KeyboardRequiresFocus bool
	ConsumeLoading        bool
}

// DefaultInteractionOptions returns the standard control interaction behavior.
func DefaultInteractionOptions() InteractionOptions {
	return InteractionOptions{
		Pointer:               true,
		Keyboard:              true,
		ClickOnMouseDown:      false,
		ClickOnMouseUp:        true,
		EnterActivates:        true,
		SpaceActivates:        true,
		KeyboardRequiresFocus: true,
		ConsumeLoading:        true,
	}
}

// InteractionController centralizes common control state transitions.
type InteractionController struct {
	target  Widget
	options InteractionOptions
	onClick func(Event)
	pressed bool
}

// NewInteractionController creates an interaction controller for a widget.
func NewInteractionController(target Widget) *InteractionController {
	return &InteractionController{
		target:  target,
		options: DefaultInteractionOptions(),
	}
}

// Target returns the widget being controlled.
func (c *InteractionController) Target() Widget {
	if c == nil {
		return nil
	}
	return c.target
}

// SetTarget updates the controlled widget.
func (c *InteractionController) SetTarget(target Widget) {
	if c == nil {
		return
	}
	clearInteractionState(c.target)
	c.pressed = false
	c.target = target
	clearInteractionState(c.target)
}

// Options returns the current interaction options.
func (c *InteractionController) Options() InteractionOptions {
	if c == nil {
		return DefaultInteractionOptions()
	}
	return c.options
}

// SetOptions replaces the interaction options.
func (c *InteractionController) SetOptions(options InteractionOptions) {
	if c == nil {
		return
	}
	c.options = options
}

// SetOnClick sets the activation callback.
func (c *InteractionController) SetOnClick(handler func(Event)) {
	if c == nil {
		return
	}
	c.onClick = handler
}

// Pressed reports whether a pointer press is active.
func (c *InteractionController) Pressed() bool {
	return c != nil && c.pressed
}

// Reset clears transient pointer interaction state.
func (c *InteractionController) Reset() {
	if c == nil {
		return
	}
	c.pressed = false
	clearInteractionState(c.target)
}

// SetHover updates hover state.
func (c *InteractionController) SetHover(hover bool) {
	if c == nil || c.target == nil {
		return
	}
	if !c.canInteract() {
		hover = false
	}
	c.target.SetStateFlag(StateHover, hover)
}

// HandleEvent applies common pointer and keyboard interaction behavior.
func (c *InteractionController) HandleEvent(ctx *Context, event Event) bool {
	if c == nil || c.target == nil {
		return false
	}
	if c.isLoading() {
		c.target.SetStateFlag(StateActive, false)
		c.pressed = false
		return c.options.ConsumeLoading && isActivationEvent(event)
	}
	if !c.canInteract() {
		c.Reset()
		return false
	}

	switch event.Type {
	case EventMouseEnter:
		if !c.options.Pointer {
			return false
		}
		c.SetHover(true)
		return false
	case EventMouseLeave:
		if !c.options.Pointer {
			return false
		}
		c.SetHover(false)
		return false
	case EventMouseMove:
		if !c.options.Pointer {
			return false
		}
		inside := c.eventInsideTarget(event)
		c.SetHover(inside)
		if c.pressed {
			c.target.SetStateFlag(StateActive, true)
		}
		return false
	case EventMouseDown:
		if !c.options.Pointer {
			return false
		}
		c.pressed = true
		c.target.SetStateFlag(StateActive, true)
		if c.options.ClickOnMouseDown && c.eventInsideTarget(event) {
			c.activate(event)
		}
		return true
	case EventMouseUp:
		if !c.options.Pointer {
			return false
		}
		wasPressed := c.pressed || c.target.HasState(StateActive)
		inside := c.eventInsideTarget(event)
		c.pressed = false
		c.target.SetStateFlag(StateActive, false)
		if wasPressed && inside && c.options.ClickOnMouseUp {
			c.activate(event)
		}
		return wasPressed
	case EventDoubleClick:
		if !c.options.Pointer {
			return false
		}
		if c.eventInsideTarget(event) {
			c.pressed = false
			c.target.SetStateFlag(StateActive, false)
			c.activate(event)
			return true
		}
		return false
	case EventKeyDown:
		if !c.shouldHandleKey(event.Key) {
			return false
		}
		c.target.SetStateFlag(StateActive, true)
		c.activate(event)
		c.target.SetStateFlag(StateActive, false)
		return true
	default:
		return false
	}
}

func (c *InteractionController) activate(event Event) {
	if c == nil || c.onClick == nil {
		return
	}
	c.onClick(event)
}

func (c *InteractionController) shouldHandleKey(key int) bool {
	if c == nil || c.target == nil || !c.options.Keyboard {
		return false
	}
	if c.options.KeyboardRequiresFocus && !c.target.Focused() {
		return false
	}
	return (key == keyEnter && c.options.EnterActivates) || (key == keySpace && c.options.SpaceActivates)
}

func (c *InteractionController) canInteract() bool {
	return c != nil && c.target != nil && c.target.Visible() && c.target.Enabled() && !c.isLoading()
}

func (c *InteractionController) isLoading() bool {
	return c != nil && c.target != nil && c.target.HasState(StateLoading)
}

func (c *InteractionController) eventInsideTarget(event Event) bool {
	if c == nil || c.target == nil {
		return false
	}
	bounds := c.target.Bounds()
	if bounds.W < 0 || bounds.H < 0 {
		return false
	}
	return event.LocalX >= 0 && event.LocalY >= 0 && event.LocalX <= bounds.W && event.LocalY <= bounds.H
}

func isActivationEvent(event Event) bool {
	switch event.Type {
	case EventMouseDown, EventMouseUp:
		return true
	case EventKeyDown:
		return event.Key == keyEnter || event.Key == keySpace
	default:
		return false
	}
}

func clearInteractionState(target Widget) {
	if target == nil {
		return
	}
	target.SetStateFlag(StateHover, false)
	target.SetStateFlag(StateActive, false)
}
