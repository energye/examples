# GPUI - Ant Design Style GPU UI Framework

A high-performance, GPU-accelerated UI framework with Ant Design style, built with purego (non-CGo).

## ✨ Features

- **Ant Design Style** - Professional, modern design system
- **GPU Accelerated** - OpenGL rendering at 60 FPS
- **Anti-Aliased** - Smooth rounded corners using SDF
- **Modular Architecture** - Clean separation of concerns
- **Smart Batching** - Automatic draw call optimization
- **Shader Caching** - Uniform locations cached for performance
- **Smooth Animations** - 200-300ms transitions
- **Theme System** - Customizable colors, spacing, typography
- **Focus Management** - Tab cycling with visual feedback
- **Event System** - Proper mouse/keyboard handling

## 📁 Architecture

```
gpui/
├── core/                    # Core Layer
│   ├── gl/                  # OpenGL bindings (purego)
│   ├── math/                # Math utilities (Vec2, Rect, Color, Mat4)
│   └── platform/            # Platform abstraction (events)
│
├── render/                  # Rendering Layer
│   ├── pipeline/            # Renderer, batch manager
│   ├── shader/              # Shader manager with caching
│   └── font/                # Font rendering with texture atlas
│
├── style/                   # Style Layer
│   ├── color/               # Ant Design color system
│   ├── theme/               # Theme system
│   └── animation/           # Animation and easing functions
│
├── widget/                  # Widget Layer
│   ├── base.go              # Base widget interface
│   ├── container.go         # Container and focus manager
│   ├── label.go             # Label widget
│   ├── button.go            # Button with states and animation
│   └── textbox.go           # TextBox with cursor and selection
│
├── ui/                      # UI Engine
│   └── engine.go            # Main engine
│
└── demo/                    # Demo Application
    └── main.go              # Demo program
```

## 🚀 Quick Start

### Prerequisites

- Go 1.20+
- OpenGL 3.0+ support
- Energy LCL framework
- System fonts

### Installation

```bash
go get github.com/energye/examples/lcl/gpui
```

### Basic Usage

```go
package main

import (
    "github.com/energye/examples/lcl/gpui/ui"
    "github.com/energye/examples/lcl/gpui/widget"
)

func main() {
    // Create engine
    engine := ui.NewEngine()
    
    // Initialize
    engine.Init()
    engine.SetSize(800, 600)
    
    // Load font
    engine.LoadDefaultFont(14)
    
    // Create widgets
    label := widget.NewLabel("Hello, World!", engine.Font())
    label.SetPos(20, 20)
    engine.AddWidget(label)
    
    btn := widget.NewButton("Click Me", widget.ButtonPrimary, engine.Font())
    btn.SetPos(20, 60)
    btn.SetOnClick(func() {
        label.SetText("Button clicked!")
    })
    engine.AddWidget(btn)
    
    // Render loop
    for {
        engine.Render()
        // ... swap buffers
    }
}
```

## 🎨 Design System

### Colors (Ant Design)

```go
// Primary colors
Primary      = #1890ff
PrimaryHover = #40a9ff
PrimaryActive = #096dd9

// Semantic colors
Success = #52c41a
Warning = #faad14
Error   = #ff4d4f

// Text colors
TextPrimary   = rgba(0,0,0,0.85)
TextSecondary = rgba(0,0,0,0.45)
TextDisabled  = rgba(0,0,0,0.25)
```

### Spacing (4px base unit)

```go
SpaceXXS = 4px
SpaceXS  = 8px
SpaceSM  = 12px
SpaceMD  = 16px
SpaceLG  = 24px
SpaceXL  = 32px
SpaceXXL = 48px
```

### Border Radius

```go
RadiusSM   = 2px
RadiusMD   = 4px   // Default
RadiusLG   = 6px
RadiusXL   = 8px
RadiusFull = 9999px (pill shape)
```

### Animation

```go
DurationFast   = 150ms
DurationNormal = 200ms
DurationSlow   = 300ms

// Easing functions
EaseOut   = cubic-bezier(0, 0, 0.2, 1)
EaseIn    = cubic-bezier(0.4, 0, 1, 1)
EaseInOut = cubic-bezier(0.4, 0, 0.2, 1)
```

## 🧩 Widgets

### Label

Static text display.

```go
label := widget.NewLabel("Hello", font)
label.SetPos(20, 20)
label.SetColor(color.TextPrimary)
label.SetFont(customFont)
```

### Button

Clickable button with states.

```go
btn := widget.NewButton("Click", widget.ButtonPrimary, font)
btn.SetPos(20, 60)
btn.SetOnClick(func() {
    // Handle click
})
```

**Button Types:**
- `ButtonDefault` - Default style
- `ButtonPrimary` - Primary blue
- `ButtonSuccess` - Success green
- `ButtonWarning` - Warning yellow
- `ButtonDanger` - Danger red

**Button States:**
- Normal
- Hovered (lighter background)
- Pressed (darker background)
- Focused (blue ring)
- Disabled (50% opacity)

### TextBox

Text input with cursor and selection.

```go
textbox := widget.NewTextBox("Placeholder...", font)
textbox.SetPos(20, 100)
textbox.SetOnChange(func(text string) {
    fmt.Println("Text:", text)
})
textbox.SetOnSubmit(func(text string) {
    fmt.Println("Submit:", text)
})
```

**Features:**
- Cursor animation (blinking)
- Text selection
- Keyboard navigation (Left/Right/Home/End)
- Delete/Backspace
- Focus ring
- Placeholder text

### Container

Container for child widgets.

```go
container := widget.NewContainer()
container.Add(label)
container.Add(button)
container.Add(textbox)
```

## 🎯 Key Improvements Over Old Architecture

### 1. No God Object

**Before:** 724-line `engine.go` doing everything

**After:** Clean separation:
- `core/` - Platform abstraction
- `render/` - Rendering pipeline
- `style/` - Visual styling
- `widget/` - UI components
- `ui/` - Engine coordination

### 2. Shader Uniform Caching

**Before:** `glGetUniformLocation` called every frame

**After:** Cached after first call:
```go
func (sm *ShaderManager) GetUniformLocation(name string) int32 {
    if loc, ok := sm.uniformLocs[name]; ok {
        return loc
    }
    loc := glGetUniformLocation(...)
    sm.uniformLocs[name] = loc
    return loc
}
```

### 3. Smart Batching

**Before:** Flush on every shader change

**After:** Automatic batching:
```go
func (bm *BatchManager) AddQuad(shader, texture, verts) {
    if bm.current.shader != shader || bm.current.texture != texture {
        bm.flushCurrent() // Flush only when needed
        bm.current = newBatch()
    }
    bm.current.verts = append(bm.current.verts, verts...)
}
```

### 4. Ant Design Style

**Before:** Hardcoded colors, inconsistent styling

**After:** Professional design system:
- Unified color palette
- Consistent spacing
- Smooth animations
- Clear state feedback

### 5. Proper Event Handling

**Before:** Events lost, coordinates wrong

**After:** Correct implementation:
- Proper coordinate transformation
- Focus management with Tab cycling
- Correct mouse up/down handling
- Keyboard event propagation

## 📊 Performance

- **60 FPS** smooth rendering
- **GPU-accelerated** anti-aliased rounded corners
- **Batched rendering** minimizes draw calls
- **Shader caching** eliminates redundant lookups
- **Efficient font atlas** reduces texture switches

## 🔧 Dependencies

- `github.com/energye/lcl` - LCL window framework
- `github.com/ebitengine/purego` - Non-CGo OpenGL bindings
- `golang.org/x/image` - Font rendering

## 📚 Documentation

- `README.md` - This file
- `demo/README.md` - Demo documentation
- Source code comments - Detailed API documentation

## 🎯 Future Plans

- [ ] More widgets (Checkbox, Radio, Dropdown, Slider)
- [ ] Layout managers (HBox, VBox, Grid)
- [ ] Scroll support
- [ ] Drag and drop
- [ ] More themes (Dark mode, Custom themes)
- [ ] Animation framework improvements
- [ ] IME support for CJK input

## 📄 License

See project root LICENSE file.

## 🤝 Contributing

Contributions welcome! Please read the contributing guidelines first.

---

**GPUI** - Professional GPU-accelerated UI with Ant Design style 🚀
