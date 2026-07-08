// Package ui provides the main UI engine
package ui

import (
	"fmt"
	"time"

	coremath "github.com/energye/examples/lcl/gpui/core/math"
	"github.com/energye/examples/lcl/gpui/overlay"
	"github.com/energye/examples/lcl/gpui/render/font"
	"github.com/energye/examples/lcl/gpui/render/pipeline"
	"github.com/energye/examples/lcl/gpui/style/token"
	"github.com/energye/examples/lcl/gpui/widget"
)

// RenderHandler is called each frame to render
type RenderHandler func(renderer *pipeline.Renderer)

// Engine is the main UI engine with automatic lifecycle management
type Engine struct {
	// Core components
	renderer *pipeline.Renderer
	root     *widget.Container
	overlay  *overlay.Manager

	// Font management
	font *font.Font

	// Window properties
	width  float32
	height float32

	// State
	initialized bool

	// Callbacks
	onRender RenderHandler

	// Animation
	lastTime   time.Time
	cursorTime float64
}

// NewEngine creates a new UI engine
func NewEngine() *Engine {
	return &Engine{
		renderer: pipeline.NewRenderer(),
		root:     widget.NewContainer(),
		overlay:  overlay.NewManager(),
		lastTime: time.Now(),
	}
}

// Initialize initializes the engine (must be called with GL context current)
func (e *Engine) Initialize() error {
	if e == nil {
		return fmt.Errorf("engine is nil")
	}
	if e.initialized {
		return nil
	}
	if e.renderer == nil {
		e.renderer = pipeline.NewRenderer()
	}
	if e.root == nil {
		e.root = widget.NewContainer()
	}
	if e.overlay == nil {
		e.overlay = overlay.NewManager()
	}

	if err := e.renderer.Init(); err != nil {
		return fmt.Errorf("engine init: %w", err)
	}

	e.initialized = true
	return nil
}

// Render renders a single frame (must be called with GL context current)
func (e *Engine) Render() {
	if e == nil {
		return
	}
	if !e.initialized {
		return
	}

	// Calculate delta time
	now := time.Now()
	dt := now.Sub(e.lastTime).Seconds()
	e.lastTime = now
	e.cursorTime += dt

	// Begin frame
	e.renderer.BeginFrame(e.width, e.height)
	ctx := e.Context()

	// Render root container
	if e.root != nil {
		e.root.Layout(ctx, widgetRootRect(e.width, e.height))
		e.root.Render(ctx)
	}

	// Call custom render handler
	if e.onRender != nil {
		e.onRender(e.renderer)
	}

	// End frame (flushes all batched draw calls)
	e.renderer.EndFrame()
}

// SetSize sets the window size
func (e *Engine) SetSize(width, height float32) {
	if e == nil {
		return
	}
	e.width = width
	e.height = height
}

// Size returns the window size
func (e *Engine) Size() (float32, float32) {
	if e == nil {
		return 0, 0
	}
	return e.width, e.height
}

// SetFont sets the default font
func (e *Engine) SetFont(f *font.Font) {
	if e == nil {
		return
	}
	e.font = f
}

// Font returns the default font
func (e *Engine) Font() *font.Font {
	if e == nil {
		return nil
	}
	return e.font
}

// Root returns the root container
func (e *Engine) Root() *widget.Container {
	if e == nil {
		return nil
	}
	return e.root
}

// Renderer returns the renderer
func (e *Engine) Renderer() *pipeline.Renderer {
	if e == nil {
		return nil
	}
	return e.renderer
}

// Overlay returns the overlay manager.
func (e *Engine) Overlay() *overlay.Manager {
	if e == nil {
		return nil
	}
	return e.overlay
}

// Context creates the current widget lifecycle context.
func (e *Engine) Context() *widget.Context {
	if e == nil {
		return nil
	}
	scale := float32(1)
	return &widget.Context{
		Renderer: e.renderer,
		Tokens:   token.Current(),
		Font:     e.font,
		Overlay:  e.overlay,
		Viewport: widgetRootRect(e.width, e.height),
		Scale:    scale,
	}
}

// CursorTime returns the cursor animation time
func (e *Engine) CursorTime() float64 {
	if e == nil {
		return 0
	}
	return e.cursorTime
}

// SetRenderHandler sets a custom render handler
func (e *Engine) SetRenderHandler(handler RenderHandler) {
	if e == nil {
		return
	}
	e.onRender = handler
}

// AddWidget adds a widget to the root container
func (e *Engine) AddWidget(w widget.Widget) {
	if e == nil || e.root == nil || w == nil {
		return
	}
	e.root.Add(w)
}

// FocusManager returns the focus manager
func (e *Engine) FocusManager() *widget.FocusManager {
	if e == nil || e.root == nil {
		return nil
	}
	return e.root.FocusManager()
}

// SetFocus sets focus to a widget
func (e *Engine) SetFocus(w widget.Widget) {
	if e == nil || e.root == nil {
		return
	}
	e.root.FocusManager().SetFocus(w)
}

// HandleMouseDown handles mouse down event
func (e *Engine) HandleMouseDown(x, y float32, button int) {
	if e == nil || e.root == nil {
		return
	}
	e.root.HandleEvent(e.Context(), widget.Event{Type: widget.EventMouseDown, X: x, Y: y, LocalX: x, LocalY: y, Button: button})
}

// HandleMouseUp handles mouse up event
func (e *Engine) HandleMouseUp(x, y float32, button int) {
	if e == nil || e.root == nil {
		return
	}
	e.root.HandleEvent(e.Context(), widget.Event{Type: widget.EventMouseUp, X: x, Y: y, LocalX: x, LocalY: y, Button: button})
}

// HandleMouseMove handles mouse move event
func (e *Engine) HandleMouseMove(x, y float32) {
	if e == nil || e.root == nil {
		return
	}
	e.root.HandleEvent(e.Context(), widget.Event{Type: widget.EventMouseMove, X: x, Y: y, LocalX: x, LocalY: y})
}

// HandleKeyDown handles key down event
func (e *Engine) HandleKeyDown(key int, mods int) {
	if e == nil || e.root == nil {
		return
	}
	focusMgr := e.root.FocusManager()

	// Handle Tab for focus cycling
	if key == 9 { // Tab
		if mods&1 != 0 { // Shift
			focusMgr.Prev()
		} else {
			focusMgr.Next()
		}
		return
	}

	// Pass to focused widget
	if focused := focusMgr.Current(); focused != nil {
		focused.HandleEvent(e.Context(), widget.Event{Type: widget.EventKeyDown, Key: key, Mods: mods})
	}
}

// HandleCharInput handles character input event
func (e *Engine) HandleCharInput(char rune) {
	if e == nil || e.root == nil {
		return
	}
	focusMgr := e.root.FocusManager()
	if focused := focusMgr.Current(); focused != nil {
		focused.HandleEvent(e.Context(), widget.Event{Type: widget.EventCharInput, Char: char})
	}
}

// Delete deletes all resources
func (e *Engine) Delete() {
	if e == nil {
		return
	}
	if e.renderer != nil {
		e.renderer.Delete()
	}
	if e.font != nil {
		e.font.Delete()
	}
}

// IsInitialized returns whether the engine is initialized
func (e *Engine) IsInitialized() bool {
	if e == nil {
		return false
	}
	return e.initialized
}

// DefaultFontData holds the default font data
var DefaultFontData []byte

// SetDefaultFontData sets the default font data
func SetDefaultFontData(data []byte) {
	DefaultFontData = data
}

// LoadFont loads a font from TTF data
func LoadFont(ttfData []byte, size float64) (*font.Font, error) {
	return font.NewFont(ttfData, size)
}

// LoadDefaultFont loads the default font
func LoadDefaultFont(size float64) (*font.Font, error) {
	if DefaultFontData == nil {
		return nil, fmt.Errorf("no font data available")
	}
	return font.NewFont(DefaultFontData, size)
}

func widgetRootRect(width, height float32) coremath.Rect {
	return coremath.NewRect(0, 0, width, height)
}
