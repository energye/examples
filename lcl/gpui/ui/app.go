// Package ui provides high-level application framework
package ui

import (
	"fmt"
	"os"
	"time"

	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
)

// Application is a high-level application wrapper
type Application struct {
	form          lcl.IEngForm
	glControl     lcl.IOpenGLControl
	engine        *Engine
	ready         bool
	initialized   bool
	onSetup       func(*Engine)
	lastTime      time.Time
	ticker        *time.Ticker
	title         string
	width         int32
	height        int32
	snapshotSaved bool
}

// NewApplication creates a new application
func NewApplication(title string, width, height int32) *Application {
	app := &Application{
		lastTime: time.Now(),
		title:    title,
		width:    width,
		height:   height,
	}

	// Store for use in FormCreate
	currentApp = app

	return app
}

// currentApp is the current application instance
var currentApp *Application

// OnSetup sets the UI setup function
func (a *Application) OnSetup(fn func(*Engine)) {
	a.onSetup = fn
}

// Run runs the application
func (a *Application) Run() {
	lcl.Init()
	lcl.RunApp(&appForm{})
}

// appForm implements lcl.IEngForm
type appForm struct {
	lcl.TEngForm
}

// FormCreate is called when the form is created
func (f *appForm) FormCreate(sender lcl.IObject) {
	app := currentApp
	if app == nil {
		return
	}

	app.form = sender.(lcl.IEngForm)
	app.form.SetCaption(app.title)
	app.form.SetWidth(app.width)
	app.form.SetHeight(app.height)

	// Create GL control
	app.glControl = lcl.NewOpenGLControl(app.form)
	configureOpenGLControl(app.glControl)
	app.glControl.SetParent(app.form)
	app.glControl.SetAlign(types.AlClient)

	// Setup event handlers
	app.setupEvents()

	// Set show handler
	app.form.SetOnShow(func(sender lcl.IObject) {
		loadSystemFont()
		app.ready = true
		app.startRenderLoop()
		app.glControl.Invalidate()
	})
}

func (a *Application) setupEvents() {
	// Paint
	a.glControl.SetOnPaint(func(sender lcl.IObject) {
		if !a.ready || !a.ensureInitialized() {
			return
		}
		if !a.glControl.MakeCurrent(true) {
			fmt.Println("OpenGL MakeCurrent failed")
			return
		}
		defer a.glControl.ReleaseContext()
		a.engine.SetSize(float32(a.form.Width()), float32(a.form.Height()))
		a.engine.Render()
		a.saveSnapshotIfRequested()
		a.glControl.SwapBuffers()
	})

	// Mouse
	a.glControl.SetOnMouseDown(func(sender lcl.IObject, button types.TMouseButton, shift types.TShiftState, x, y int32) {
		if a.engine != nil {
			btn := 0
			if button == types.MbRight {
				btn = 2
			}
			a.engine.HandleMouseDown(float32(x), float32(y), btn)
		}
	})

	a.glControl.SetOnMouseUp(func(sender lcl.IObject, button types.TMouseButton, shift types.TShiftState, x, y int32) {
		if a.engine != nil {
			btn := 0
			if button == types.MbRight {
				btn = 2
			}
			a.engine.HandleMouseUp(float32(x), float32(y), btn)
		}
	})

	a.glControl.SetOnMouseMove(func(sender lcl.IObject, shift types.TShiftState, x, y int32) {
		if a.engine != nil {
			a.engine.HandleMouseMove(float32(x), float32(y))
		}
	})

	a.glControl.SetOnMouseEnter(func(lcl.IObject) {
		a.glControl.SetFocus()
	})

	// Keyboard
	a.form.SetOnKeyDown(func(sender lcl.IObject, key *uint16, shift types.TShiftState) {
		if a.engine != nil {
			mods := 0
			if shift.In(types.SsShift) {
				mods |= 1
			}
			if shift.In(types.SsCtrl) {
				mods |= 2
			}
			a.engine.HandleKeyDown(int(*key), mods)
		}
	})

	a.form.SetOnKeyPress(func(sender lcl.IObject, key *uint16) {
		if a.engine != nil {
			a.engine.HandleCharInput(rune(*key))
		}
	})

	// Resize
	a.form.SetOnResize(func(sender lcl.IObject) {
		if a.engine != nil && a.ready {
			a.engine.SetSize(float32(a.form.Width()), float32(a.form.Height()))
		}
	})

	// Close
	a.form.SetOnClose(func(sender lcl.IObject, action *types.TCloseAction) {
		a.stop()
		*action = types.CaFree
	})
}

func (a *Application) ensureInitialized() bool {
	if a.initialized {
		return true
	}
	if a.glControl == nil || !a.glControl.HandleAllocated() {
		return false
	}
	a.glControl.RealizeBounds()
	if !a.glControl.MakeCurrent(true) {
		return false
	}
	defer a.glControl.ReleaseContext()

	a.engine = NewEngine()
	if err := a.engine.Initialize(); err != nil {
		fmt.Println("Engine init error:", err)
		return false
	}

	a.engine.SetSize(float32(a.form.Width()), float32(a.form.Height()))

	// Load font
	if DefaultFontData != nil {
		font, err := LoadDefaultFont(14)
		if err != nil {
			fmt.Println("Font load error:", err)
		} else {
			a.engine.SetFont(font)
			fmt.Println("✓ Font loaded")
		}
	}

	if a.onSetup != nil {
		a.onSetup(a.engine)
	}
	a.initialized = true
	fmt.Println("✓ Engine initialized")
	return true
}

func (a *Application) startRenderLoop() {
	if a.ticker != nil {
		return
	}
	a.ticker = time.NewTicker(time.Second / 60)
	go func() {
		for range a.ticker.C {
			if a.ready && a.glControl != nil {
				lcl.RunOnMainThreadSync(func() {
					a.glControl.Invalidate()
				})
			}
		}
	}()
}

func (a *Application) stop() {
	if a.ticker != nil {
		a.ticker.Stop()
	}
	if a.engine != nil {
		a.glControl.MakeCurrent(true)
		a.engine.Delete()
		a.glControl.ReleaseContext()
		a.engine = nil
	}
}

func (a *Application) saveSnapshotIfRequested() {
	if a.snapshotSaved || a.engine == nil {
		return
	}
	path := os.Getenv("GPUI_GPU_SNAPSHOT")
	if path == "" {
		return
	}
	if err := a.engine.Renderer().SavePNG(path); err != nil {
		fmt.Println("GPU snapshot error:", err)
		return
	}
	a.snapshotSaved = true
	fmt.Println("✓ GPU snapshot:", path)
	a.form.Close()
}

// loadSystemFont loads a system font
func loadSystemFont() {
	paths := []string{
		"/usr/share/fonts/opentype/noto/NotoSansCJK-Light.ttc",
		"/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc",
		"/usr/share/fonts/truetype/wqy/wqy-microhei.ttc",
		"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
		"/System/Library/Fonts/PingFang.ttc",
		"C:/Windows/Fonts/msyh.ttc",
	}

	for _, p := range paths {
		d, err := os.ReadFile(p)
		if err == nil {
			SetDefaultFontData(d)
			fmt.Printf("✓ Font: %s\n", p)
			return
		}
	}
	fmt.Println("✗ No font found")
}
