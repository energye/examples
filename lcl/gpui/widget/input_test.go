package widget

import (
	"testing"

	"github.com/energye/examples/lcl/gpui/core/math"
)

func TestInputCreation(t *testing.T) {
	input := NewInput("Enter text")
	if input == nil {
		t.Fatal("NewInput returned nil")
	}
	if input.Placeholder() != "Enter text" {
		t.Fatalf("placeholder = %q, want %q", input.Placeholder(), "Enter text")
	}
	if input.Text() != "" {
		t.Fatalf("text = %q, want empty", input.Text())
	}
	if input.InputType() != InputText {
		t.Fatalf("inputType = %v, want InputText", input.InputType())
	}
	if input.Readonly() {
		t.Fatal("input should not be readonly by default")
	}
	if input.MaxLength() != 0 {
		t.Fatalf("maxLength = %d, want 0", input.MaxLength())
	}
}

func TestInputSetText(t *testing.T) {
	input := NewInput("")
	input.SetText("hello")
	if input.Text() != "hello" {
		t.Fatalf("text = %q, want %q", input.Text(), "hello")
	}
	if input.CursorPos() != 5 {
		t.Fatalf("cursorPos = %d, want 5", input.CursorPos())
	}
}

func TestInputMaxLength(t *testing.T) {
	input := NewInput("")
	input.SetMaxLength(5)
	input.SetText("hello world")
	if input.Text() != "hello" {
		t.Fatalf("text = %q, want %q", input.Text(), "hello")
	}
	if len(input.Text()) > 5 {
		t.Fatalf("text length = %d, want <= 5", len(input.Text()))
	}
}

func TestInputInsertText(t *testing.T) {
	input := NewInput("")
	input.InsertText("hello")
	if input.Text() != "hello" {
		t.Fatalf("text = %q, want %q", input.Text(), "hello")
	}
	input.InsertText(" world")
	if input.Text() != "hello world" {
		t.Fatalf("text = %q, want %q", input.Text(), "hello world")
	}
}

func TestInputInsertTextAtCursor(t *testing.T) {
	input := NewInput("")
	input.InsertText("hello")
	input.SetCursorPos(2)
	input.InsertText("XX")
	if input.Text() != "heXXllo" {
		t.Fatalf("text = %q, want %q", input.Text(), "heXXllo")
	}
	if input.CursorPos() != 4 {
		t.Fatalf("cursorPos = %d, want 4", input.CursorPos())
	}
}

func TestInputDeleteSelection(t *testing.T) {
	input := NewInput("")
	input.SetText("hello world")
	input.SetSelection(5, 11)
	input.DeleteSelection()
	if input.Text() != "hello" {
		t.Fatalf("text = %q, want %q", input.Text(), "hello")
	}
}

func TestInputSelectAll(t *testing.T) {
	input := NewInput("")
	input.SetText("hello world")
	input.SelectAll()
	if !input.HasSelection() {
		t.Fatal("should have selection")
	}
	start, end := input.Selection()
	if start != 0 || end != 11 {
		t.Fatalf("selection = (%d,%d), want (0,11)", start, end)
	}
}

func TestInputClear(t *testing.T) {
	input := NewInput("")
	input.SetText("hello")
	input.Clear()
	if input.Text() != "" {
		t.Fatalf("text = %q, want empty", input.Text())
	}
	if input.CursorPos() != 0 {
		t.Fatalf("cursorPos = %d, want 0", input.CursorPos())
	}
}

func TestInputReadonly(t *testing.T) {
	input := NewInput("")
	input.SetReadonly(true)
	input.SetText("hello")
	input.InsertText(" world")
	if input.Text() != "hello" {
		t.Fatalf("text = %q, want %q (readonly should not allow insert)", input.Text(), "hello")
	}
}

func TestInputCallbacks(t *testing.T) {
	input := NewInput("")
	changed := false
	input.SetOnChange(func(text string) {
		changed = true
	})
	input.SetText("hello")
	if !changed {
		t.Fatal("onChange should have been called")
	}
}

func TestInputClearCallback(t *testing.T) {
	input := NewInput("")
	input.SetText("hello")
	cleared := false
	input.SetOnClear(func() {
		cleared = true
	})
	input.Clear()
	if !cleared {
		t.Fatal("onClear should have been called")
	}
}

func TestInputSubmitCallback(t *testing.T) {
	input := NewInput("")
	input.SetText("hello")
	submitted := ""
	input.SetOnSubmit(func(text string) {
		submitted = text
	})
	input.HandleEvent(nil, Event{Type: EventKeyDown, Key: keyEnter})
	if submitted != "hello" {
		t.Fatalf("submitted = %q, want %q", submitted, "hello")
	}
}

func TestInputBackspace(t *testing.T) {
	input := NewInput("")
	input.SetText("hello")
	input.SetCursorPos(3)
	input.HandleEvent(nil, Event{Type: EventKeyDown, Key: keyBackspace})
	if input.Text() != "helo" {
		t.Fatalf("text = %q, want %q", input.Text(), "helo")
	}
	if input.CursorPos() != 2 {
		t.Fatalf("cursorPos = %d, want 2", input.CursorPos())
	}
}

func TestInputDelete(t *testing.T) {
	input := NewInput("")
	input.SetText("hello")
	input.SetCursorPos(2)
	input.HandleEvent(nil, Event{Type: EventKeyDown, Key: keyDelete})
	if input.Text() != "helo" {
		t.Fatalf("text = %q, want %q", input.Text(), "helo")
	}
	if input.CursorPos() != 2 {
		t.Fatalf("cursorPos = %d, want 2", input.CursorPos())
	}
}

func TestInputArrowKeys(t *testing.T) {
	input := NewInput("")
	input.SetText("hello")
	input.SetCursorPos(2)

	input.HandleEvent(nil, Event{Type: EventKeyDown, Key: keyArrowLeft})
	if input.CursorPos() != 1 {
		t.Fatalf("cursorPos = %d, want 1", input.CursorPos())
	}

	input.HandleEvent(nil, Event{Type: EventKeyDown, Key: keyArrowRight})
	if input.CursorPos() != 2 {
		t.Fatalf("cursorPos = %d, want 2", input.CursorPos())
	}

	input.HandleEvent(nil, Event{Type: EventKeyDown, Key: keyHome})
	if input.CursorPos() != 0 {
		t.Fatalf("cursorPos = %d, want 0", input.CursorPos())
	}

	input.HandleEvent(nil, Event{Type: EventKeyDown, Key: keyEnd})
	if input.CursorPos() != 5 {
		t.Fatalf("cursorPos = %d, want 5", input.CursorPos())
	}
}

func TestInputPasswordType(t *testing.T) {
	input := NewInput("")
	input.SetInputType(InputPassword)
	input.SetText("secret")
	if input.InputType() != InputPassword {
		t.Fatalf("inputType = %v, want InputPassword", input.InputType())
	}
}

func TestInputSearchType(t *testing.T) {
	input := NewInput("")
	input.SetInputType(InputSearch)
	if input.InputType() != InputSearch {
		t.Fatalf("inputType = %v, want InputSearch", input.InputType())
	}
}

func TestInputNilSafety(t *testing.T) {
	var input *Input
	// All methods should be nil-safe
	input.SetText("test")
	input.Clear()
	input.InsertText("test")
	input.SelectAll()
	input.SetCursorPos(0)
	input.SetSelection(0, 1)
	input.DeleteSelection()
	input.SetOnChange(nil)
	input.SetOnSubmit(nil)
	input.SetOnFocus(nil)
	input.SetOnBlur(nil)
	input.SetOnClear(nil)
	input.Focus()
	input.Blur()
}

func TestInputMeasure(t *testing.T) {
	input := NewInput("placeholder")
	ctx := &Context{}
	size := input.Measure(ctx, Constraints{Max: math.NewVec2(200, 100)})
	if size.X <= 0 || size.Y <= 0 {
		t.Fatalf("size = (%v,%v), want positive values", size.X, size.Y)
	}
}

func TestInputAllowClear(t *testing.T) {
	input := NewInput("")
	input.SetAllowClear(true)
	if !input.AllowClear() {
		t.Fatal("allowClear should be true")
	}
	input.SetText("hello")
	input.Clear()
	if input.Text() != "" {
		t.Fatalf("text = %q, want empty after clear", input.Text())
	}
}

func TestInputShowCount(t *testing.T) {
	input := NewInput("")
	input.SetShowCount(true)
	input.SetMaxLength(10)
	if !input.ShowCount() {
		t.Fatal("showCount should be true")
	}
}

func TestInputSetMaxLengthTruncates(t *testing.T) {
	input := NewInput("")
	input.SetText("hello world")
	input.SetMaxLength(5)
	if input.Text() != "hello" {
		t.Fatalf("text = %q, want %q", input.Text(), "hello")
	}
	if input.CursorPos() > 5 {
		t.Fatalf("cursorPos = %d, want <= 5", input.CursorPos())
	}
}
