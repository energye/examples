# Demo: Ant Design Style GPU UI

This demo showcases the new GPUI framework with Ant Design style.

## Features

- ✅ Modern Ant Design style visuals
- ✅ Smooth animations (200-300ms transitions)
- ✅ Professional rounded corners with anti-aliasing
- ✅ Button states (Hover, Pressed, Focused)
- ✅ TextBox with cursor animation
- ✅ Tab key focus cycling
- ✅ Clean, modular architecture

## Widgets Demonstrated

1. **Label** - Title text with primary color
2. **TextBox** - Text input with placeholder
3. **Primary Button** - Blue primary button
4. **Default Button** - Default style button

## Running the Demo

```bash
cd /home/yanghy/app/workspace/examples/lcl/gpui/demo
go run main.go
```

## Expected Output

```
✓ Font loaded: /usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc (XXX bytes)
✓ Engine initialized
✓ Font loaded
✓ UI initialized
```

## Interactions

### Mouse:
- Click TextBox to focus and type
- Hover over buttons to see highlight
- Click buttons to trigger actions
- Tab to switch focus between widgets

### Keyboard:
- Type in TextBox when focused
- Backspace/Delete to delete text
- Left/Right arrows to move cursor
- Home/End to jump to start/end
- Enter to submit (TextBox)
- Tab to cycle focus

## Architecture

This demo uses the new modular architecture:

```
├── core/         # GL bindings, math utilities
├── render/       # Renderer, shaders, font
├── style/        # Theme, colors, animations
├── widget/       # UI widgets
└── ui/           # Engine
```

## Key Improvements

1. **No God Object** - Clean separation of concerns
2. **Shader Caching** - Uniform locations cached for performance
3. **Smart Batching** - Automatic draw call batching
4. **Ant Design Style** - Professional, modern visuals
5. **Smooth Animations** - 200-300ms transitions
6. **Proper Events** - Correct mouse/keyboard handling
7. **Focus Management** - Tab cycling with visual feedback

## Screenshots

The UI features:
- Clean white background
- Blue primary color (#1890ff)
- 4px rounded corners
- Smooth hover effects
- Animated cursor
- Focus ring on active widgets
