package pipeline

import (
	"strings"
	"unicode/utf8"

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
	lastSpaceByte := -1
	i := 0

	for _, r := range text {
		if r == ' ' || r == '\t' {
			lastSpace = i
			lastSpaceByte = i
		}
		width += f.RuneAdvance(r)
		if width > maxWidth {
			if lastSpace >= 0 {
				return text[:lastSpaceByte], strings.TrimLeft(text[lastSpaceByte+len(" "):], " \t")
			}
			if i == 0 {
				// Return at least one character
				size := utf8.RuneLen(r)
				return text[:size], text[size:]
			}
			return text[:i], text[i:]
		}
		i += utf8.RuneLen(r)
	}

	return text, ""
}

func ellipsize(text string, f *font.Font, maxWidth float32) string {
	const suffix = "..."
	if f.TextWidth(text) <= maxWidth {
		return text
	}

	// Find the longest prefix that fits with suffix
	for i := len(text); i > 0; i-- {
		// Find a valid UTF-8 boundary
		if i < len(text) && (text[i]&0xC0) == 0x80 {
			continue // Skip continuation bytes
		}
		candidate := text[:i] + suffix
		if f.TextWidth(candidate) <= maxWidth {
			return candidate
		}
	}
	return suffix
}
