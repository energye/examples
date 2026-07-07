// Package ui provides window integration for LCL
package ui

import (
	"fmt"
	"os"
	"time"

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
	w.form = sender.(lcl.IEngForm)
	w.form.SetCaption(w.config.Title)
	w.form.SetWidth(w.config.Width)
	w.form.SetHeight(w.config.Height)
	w.form.ScreenCenter()

	// Create OpenGL control
	w.glControl = lcl.NewOpenGLControl(w.form)
	w.glControl.SetParent(w.form)
	w.glControl.SetAlign(types.AlClient)

	// Set up event handlers
	w.glControl.SetOnPaint(w.onPaint)
	w.glControl.SetOnMouseDown(w.onMouseDown)
	w.glControl.SetOnMouseUp(w.onMouseUp)
	w.glControl.SetOnMouseMove(w.onMouseMove)
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
	// Load font first
	w.loadFont()

	// Initialize engine in main thread
	lcl.RunOnMainThreadSync(func() {
		w.glControl.MakeCurrent(true)
		defer w.glControl.ReleaseContext()

		if err := w.engine.Initialize(); err != nil {
			fmt.Println("✗ Engine init error:", err)
			return
		}
		fmt.Println("✓ Engine initialized")

		// Set initial size
		w.engine.SetSize(float32(w.form.Width()), float32(w.form.Height()))

		// Load font
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

		// Call user's show handler after initialization
		if w.onShowHandler != nil {
			w.onShowHandler()
		}
	})
}

// loadFont loads a suitable font
func (w *Window) loadFont() {
	paths := []string{
		// High quality CJK fonts
		"/usr/share/fonts/opentype/noto/NotoSansCJK-Light.ttc",
		"/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc",
		"/usr/share/fonts/truetype/wqy/wqy-microhei.ttc",
		// Fallback
		"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
		"/usr/share/fonts/truetype/liberation/LiberationSans-Regular.ttf",
		"/System/Library/Fonts/PingFang.ttc",
		"C:/Windows/Fonts/msyh.ttc",
		"C:/Windows/Fonts/arial.ttf",
	}

	for _, p := range paths {
		d, err := os.ReadFile(p)
		if err == nil {
			SetDefaultFontData(d)
			fmt.Printf("✓ Font file loaded: %s\n", p)
			return
		}
	}

	fmt.Println("✗ Warning: no font file found")
}

// onPaint handles paint events
func (w *Window) onPaint(sender lcl.IObject) {
	if !w.initialized || w.engine == nil {
		return
	}

	w.glControl.MakeCurrent(true)
	defer w.glControl.ReleaseContext()

	// Update size
	w.engine.SetSize(float32(w.form.Width()), float32(w.form.Height()))

	w.engine.Render()
	w.glControl.SwapBuffers()
}

// onMouseDown handles mouse down events
func (w *Window) onMouseDown(sender lcl.IObject, button types.TMouseButton, shift types.TShiftState, x, y int32) {
	if w.engine == nil {
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
	if w.engine == nil {
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
	if w.engine == nil {
		return
	}
	w.engine.HandleMouseMove(float32(x), float32(y))
}

// onKeyDown handles key down events
func (w *Window) onKeyDown(sender lcl.IObject, key *uint16, shift types.TShiftState) {
	if w.engine == nil {
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
	if w.engine == nil {
		return
	}
	w.engine.HandleCharInput(rune(*key))
}

// onResize handles resize events
func (w *Window) onResize(sender lcl.IObject) {
	if w.engine != nil && w.initialized {
		w.engine.SetSize(float32(w.form.Width()), float32(w.form.Height()))
	}
}

// onClose handles close events
func (w *Window) onClose(sender lcl.IObject, action *types.TCloseAction) {
	if w.engine != nil {
		w.glControl.MakeCurrent(true)
		w.engine.Delete()
		w.glControl.ReleaseContext()
		w.engine = nil
	}
	*action = types.CaFree
}

// Engine returns the UI engine
func (w *Window) Engine() *Engine {
	return w.engine
}

// Form returns the form
func (w *Window) Form() lcl.IEngForm {
	return w.form
}

// GLControl returns the OpenGL control
func (w *Window) GLControl() lcl.IOpenGLControl {
	return w.glControl
}

// OnShow sets the show handler
func (w *Window) OnShow(handler func()) {
	w.onShowHandler = handler
}

// Run runs the application
func (w *Window) Run() {
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
