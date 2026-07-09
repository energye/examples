// Package ui provides window integration for LCL
package ui

import (
	"fmt"
	"time"

	renderfont "github.com/energye/examples/lcl/gpui/render/font"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
)

// WindowConfig represents window configuration
type WindowConfig struct {
	Title  string
	Width  int32
	Height int32
}

// Window represents a window with integrated UI engine
type Window struct {
	config    WindowConfig
	form      lcl.IEngForm
	glControl lcl.IOpenGLControl
	engine    *Engine

	// State
	initialized bool
	lastTime    time.Time

	// Mouse state
	mouseDown bool

	// Callbacks
	onShowHandler func()
}

// NewWindow creates a new window with integrated UI engine
func NewWindow(config WindowConfig) *Window {
	w := &Window{
		config:   config,
		lastTime: time.Now(),
	}

	return w
}

// setupWindow sets up the window (called by Run)
func (w *Window) setupWindow(sender lcl.IObject) {
	if w == nil {
		return
	}
	form, ok := sender.(lcl.IEngForm)
	if !ok || form == nil {
		return
	}
	w.form = form
	w.form.SetCaption(w.config.Title)
	w.form.SetWidth(w.config.Width)
	w.form.SetHeight(w.config.Height)
	w.form.ScreenCenter()

	// Create OpenGL control
	w.glControl = lcl.NewOpenGLControl(w.form)
	configureOpenGLControl(w.glControl)
	w.glControl.SetParent(w.form)
	w.glControl.SetAlign(types.AlClient)

	// Set up event handlers
	w.glControl.SetOnPaint(w.onPaint)
	w.glControl.SetOnMouseDown(w.onMouseDown)
	w.glControl.SetOnMouseUp(w.onMouseUp)
	w.glControl.SetOnMouseMove(w.onMouseMove)
	w.glControl.SetOnMouseWheel(w.onMouseWheel)
	w.glControl.SetOnMouseEnter(func(lcl.IObject) { w.glControl.SetFocus() })

	w.form.SetOnKeyDown(w.onKeyDown)
	w.form.SetOnKeyPress(w.onKeyPress)
	w.form.SetOnResize(w.onResize)
	w.form.SetOnClose(w.onClose)
	w.form.SetOnShow(w.onShow)

	// Create engine
	w.engine = NewEngine()
}

// onShow is called when the form is shown
func (w *Window) onShow(sender lcl.IObject) {
	if w == nil || w.glControl == nil {
		return
	}
	// Load font first
	w.loadFont()

	w.glControl.Invalidate()
}

// loadFont loads a suitable font
func (w *Window) loadFont() {
	if w == nil {
		return
	}
	systemFonts, err := renderfont.LoadSystemFontSet()
	if err == nil {
		fallbacks := make([][]byte, 0, len(systemFonts)-1)
		for _, fallback := range systemFonts[1:] {
			fallbacks = append(fallbacks, fallback.Data)
		}
		SetDefaultFontSet(systemFonts[0].Data, fallbacks...)
		fmt.Printf("✓ Font file loaded: %s", systemFonts[0].Path)
		if len(systemFonts) > 1 {
			fmt.Printf(" + %d fallback(s)", len(systemFonts)-1)
		}
		bestCJK := 0
		for _, systemFont := range systemFonts {
			if systemFont.CJKScore > bestCJK {
				bestCJK = systemFont.CJKScore
			}
		}
		if bestCJK < len(renderfont.CJKProbeRunes) {
			fmt.Printf(" (CJK coverage %d/%d)", bestCJK, len(renderfont.CJKProbeRunes))
		}
		fmt.Println()
		return
	}

	fmt.Println("✗ Warning: no freetype-supported font file found")
}

// onPaint handles paint events
func (w *Window) onPaint(sender lcl.IObject) {
	if w == nil || w.glControl == nil || w.form == nil {
		return
	}
	if !w.ensureInitialized() {
		return
	}

	if !w.glControl.MakeCurrent(true) {
		fmt.Println("✗ OpenGL MakeCurrent failed")
		return
	}
	defer w.glControl.ReleaseContext()

	// Update size
	w.engine.SetSize(float32(w.form.Width()), float32(w.form.Height()))

	w.engine.Render()
	w.glControl.SwapBuffers()
}

func (w *Window) ensureInitialized() bool {
	if w == nil {
		return false
	}
	if w.initialized {
		return true
	}
	if w.engine == nil {
		w.engine = NewEngine()
	}
	if w.glControl == nil || !w.glControl.HandleAllocated() {
		return false
	}

	w.glControl.RealizeBounds()
	if !w.glControl.MakeCurrent(true) {
		return false
	}
	defer w.glControl.ReleaseContext()

	if err := w.engine.Initialize(); err != nil {
		fmt.Println("✗ Engine init error:", err)
		return false
	}
	fmt.Println("✓ Engine initialized")

	w.engine.SetSize(float32(w.form.Width()), float32(w.form.Height()))

	if DefaultFontData != nil {
		f, err := LoadDefaultFont(14)
		if err != nil {
			fmt.Println("✗ Font load error:", err)
		} else {
			w.engine.SetFont(f)
			fmt.Println("✓ Font loaded")
		}
	}

	w.initialized = true
	if w.onShowHandler != nil {
		w.onShowHandler()
	}
	return true
}

// onMouseDown handles mouse down events
func (w *Window) onMouseDown(sender lcl.IObject, button types.TMouseButton, shift types.TShiftState, x, y int32) {
	if w == nil || w.engine == nil {
		return
	}
	btn := 0
	if button == types.MbRight {
		btn = 2
	}
	w.mouseDown = true
	w.engine.HandleMouseDown(float32(x), float32(y), btn)
}

// onMouseUp handles mouse up events
func (w *Window) onMouseUp(sender lcl.IObject, button types.TMouseButton, shift types.TShiftState, x, y int32) {
	if w == nil || w.engine == nil {
		return
	}
	btn := 0
	if button == types.MbRight {
		btn = 2
	}
	w.mouseDown = false
	w.engine.HandleMouseUp(float32(x), float32(y), btn)
}

// onMouseMove handles mouse move events
func (w *Window) onMouseMove(sender lcl.IObject, shift types.TShiftState, x, y int32) {
	if w == nil || w.engine == nil {
		return
	}
	w.engine.HandleMouseMove(float32(x), float32(y))
}

// onMouseWheel handles mouse wheel events
func (w *Window) onMouseWheel(sender lcl.IObject, shift types.TShiftState, wheelDelta int32, mousePos types.TPoint, handled *bool) {
	if w == nil || w.engine == nil {
		return
	}
	w.engine.HandleMouseWheel(float32(mousePos.X), float32(mousePos.Y), 0, float32(wheelDelta))
	if handled != nil {
		*handled = true
	}
}

// onKeyDown handles key down events
func (w *Window) onKeyDown(sender lcl.IObject, key *uint16, shift types.TShiftState) {
	if w == nil || w.engine == nil || key == nil {
		return
	}
	mods := 0
	if shift.In(types.SsShift) {
		mods |= 1
	}
	if shift.In(types.SsCtrl) {
		mods |= 2
	}
	w.engine.HandleKeyDown(int(*key), mods)
}

// onKeyPress handles key press events
func (w *Window) onKeyPress(sender lcl.IObject, key *uint16) {
	if w == nil || w.engine == nil || key == nil {
		return
	}
	w.engine.HandleCharInput(rune(*key))
}

// onResize handles resize events
func (w *Window) onResize(sender lcl.IObject) {
	if w != nil && w.engine != nil && w.initialized && w.form != nil {
		w.engine.SetSize(float32(w.form.Width()), float32(w.form.Height()))
	}
}

// onClose handles close events
func (w *Window) onClose(sender lcl.IObject, action *types.TCloseAction) {
	if w != nil && w.engine != nil && w.glControl != nil && w.glControl.MakeCurrent(true) {
		w.engine.Delete()
		w.glControl.ReleaseContext()
	}
	if w != nil {
		w.engine = nil
		w.initialized = false
	}
	if action != nil {
		*action = types.CaFree
	}
}

// Engine returns the UI engine
func (w *Window) Engine() *Engine {
	if w == nil {
		return nil
	}
	return w.engine
}

// Form returns the form
func (w *Window) Form() lcl.IEngForm {
	if w == nil {
		return nil
	}
	return w.form
}

// GLControl returns the OpenGL control
func (w *Window) GLControl() lcl.IOpenGLControl {
	if w == nil {
		return nil
	}
	return w.glControl
}

// OnShow sets the show handler
func (w *Window) OnShow(handler func()) {
	if w == nil {
		return
	}
	w.onShowHandler = handler
}

// Run runs the application
func (w *Window) Run() {
	if w == nil {
		return
	}
	// Create a form wrapper
	formWrapper := &windowForm{window: w}
	lcl.Init()
	lcl.RunApp(formWrapper)
}

// windowForm wraps the Window to implement lcl.IEngForm
type windowForm struct {
	lcl.TEngForm
	window *Window
}

// FormCreate is called when the form is created
func (f *windowForm) FormCreate(sender lcl.IObject) {
	fmt.Println("✓ FormCreate called")
	f.window.setupWindow(sender)
	fmt.Println("✓ Window setup complete")
}
