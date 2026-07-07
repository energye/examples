package pipeline

import (
	"strings"

	"github.com/energye/examples/lcl/gpui/core/math"
	"github.com/energye/examples/lcl/gpui/render/font"
)

// TextAlign controls horizontal text placement within a rectangle.
type TextAlign int

const (
	TextAlignLeft TextAlign = iota
	TextAlignCenter
	TextAlignRight
)

// TextOptions controls rectangle text rendering.
type TextOptions struct {
	Font       *font.Font
	Color      math.Color
	Align      TextAlign
	MaxLines   int
	Ellipsis   bool
	LineHeight float32
}

// DrawTextInRect draws text constrained to a rectangle.
func (r *Renderer) DrawTextInRect(text string, rect math.Rect, opts TextOptions) {
	if text == "" || opts.Font == nil || rect.W <= 0 || rect.H <= 0 {
		return
	}

	lineHeight := opts.LineHeight
	if lineHeight <= 0 {
		lineHeight = opts.Font.LineHeight()
	}

	maxLines := opts.MaxLines
	if maxLines <= 0 {
		maxLines = int(rect.H / lineHeight)
		if maxLines < 1 {
			maxLines = 1
		}
	}

	lines := wrapText(text, opts.Font, rect.W, maxLines, opts.Ellipsis)
	r.PushClip(rect)
	for i, line := range lines {
		y := rect.Y + float32(i)*lineHeight
		if y+lineHeight > rect.Y+rect.H {
			break
		}
		x := alignedTextX(line, rect, opts.Font, opts.Align)
		r.DrawText(line, x, y, opts.Font, opts.Color)
	}
	r.PopClip()
}

func alignedTextX(text string, rect math.Rect, f *font.Font, align TextAlign) float32 {
	width := f.TextWidth(text)
	switch align {
	case TextAlignCenter:
		return rect.X + (rect.W-width)/2
	case TextAlignRight:
		return rect.X + rect.W - width
	default:
		return rect.X
	}
}

func wrapText(text string, f *font.Font, maxWidth float32, maxLines int, ellipsis bool) []string {
	rawLines := strings.Split(text, "\n")
	lines := make([]string, 0, maxLines)

	for _, raw := range rawLines {
		for len(raw) > 0 {
			if len(lines) >= maxLines {
				return lines
			}

			line, rest := takeLine(raw, f, maxWidth)
			if rest == "" {
				lines = append(lines, line)
				break
			}

			if len(lines) == maxLines-1 && ellipsis {
				lines = append(lines, ellipsize(line, f, maxWidth))
				return lines
			}

			lines = append(lines, line)
			raw = rest
		}
		if raw == "" && len(lines) < maxLines {
			continue
		}
	}

	return lines
}

func takeLine(text string, f *font.Font, maxWidth float32) (line, rest string) {
	var width float32
	lastSpace := -1
	lastSpaceWidth := float32(0)
	runes := []rune(text)

	for i, r := range runes {
		if r == ' ' || r == '\t' {
			lastSpace = i
			lastSpaceWidth = width
		}
		g, ok := f.GetGlyph(r)
		if ok {
			width += g.Advance
		}
		if width > maxWidth {
			if lastSpace > 0 {
				return string(runes[:lastSpace]), strings.TrimLeft(string(runes[lastSpace+1:]), " \t")
			}
			if i == 0 {
				return string(runes[:1]), string(runes[1:])
			}
			_ = lastSpaceWidth
			return string(runes[:i]), string(runes[i:])
		}
	}

	return text, ""
}

func ellipsize(text string, f *font.Font, maxWidth float32) string {
	const suffix = "..."
	if f.TextWidth(text) <= maxWidth {
		return text
	}

	runes := []rune(text)
	for len(runes) > 0 {
		candidate := string(runes) + suffix
		if f.TextWidth(candidate) <= maxWidth {
			return candidate
		}
		runes = runes[:len(runes)-1]
	}
	return suffix
}
