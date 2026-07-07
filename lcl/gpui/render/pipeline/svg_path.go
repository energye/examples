package pipeline

import (
	"fmt"
	"strconv"
	"unicode"

	"github.com/energye/examples/lcl/gpui/core/math"
)

const curveSegments = 16

// ParseSVGPath parses common SVG path commands into a flattened Path.
func ParseSVGPath(data string) (*Path, error) {
	tokens := svgPathTokens(data)
	if len(tokens) == 0 {
		return NewPath(), nil
	}

	path := NewPath()
	var current math.Vec2
	var start math.Vec2
	var cmd byte
	i := 0

	for i < len(tokens) {
		if isCommandToken(tokens[i]) {
			cmd = tokens[i][0]
			i++
		} else if cmd == 0 {
			return nil, fmt.Errorf("svg path: missing command before %q", tokens[i])
		}

		relative := unicode.IsLower(rune(cmd))
		switch unicode.ToUpper(rune(cmd)) {
		case 'M':
			first := true
			for i+1 < len(tokens) && !isCommandToken(tokens[i]) {
				x, y, next, err := readPair(tokens, i)
				if err != nil {
					return nil, err
				}
				i = next
				pos := resolvePoint(current, x, y, relative)
				if first {
					path.MoveTo(pos.X, pos.Y)
					start = pos
					first = false
				} else {
					path.LineTo(pos.X, pos.Y)
				}
				current = pos
			}

		case 'L':
			for i+1 < len(tokens) && !isCommandToken(tokens[i]) {
				x, y, next, err := readPair(tokens, i)
				if err != nil {
					return nil, err
				}
				i = next
				current = resolvePoint(current, x, y, relative)
				path.LineTo(current.X, current.Y)
			}

		case 'H':
			for i < len(tokens) && !isCommandToken(tokens[i]) {
				x, err := parseFloat32(tokens[i])
				if err != nil {
					return nil, err
				}
				i++
				if relative {
					current.X += x
				} else {
					current.X = x
				}
				path.LineTo(current.X, current.Y)
			}

		case 'V':
			for i < len(tokens) && !isCommandToken(tokens[i]) {
				y, err := parseFloat32(tokens[i])
				if err != nil {
					return nil, err
				}
				i++
				if relative {
					current.Y += y
				} else {
					current.Y = y
				}
				path.LineTo(current.X, current.Y)
			}

		case 'C':
			for i+5 < len(tokens) && !isCommandToken(tokens[i]) {
				values, next, err := readValues(tokens, i, 6)
				if err != nil {
					return nil, err
				}
				i = next
				c1 := resolvePoint(current, values[0], values[1], relative)
				c2 := resolvePoint(current, values[2], values[3], relative)
				end := resolvePoint(current, values[4], values[5], relative)
				flattenCubic(path, current, c1, c2, end)
				current = end
			}

		case 'Q':
			for i+3 < len(tokens) && !isCommandToken(tokens[i]) {
				values, next, err := readValues(tokens, i, 4)
				if err != nil {
					return nil, err
				}
				i = next
				c := resolvePoint(current, values[0], values[1], relative)
				end := resolvePoint(current, values[2], values[3], relative)
				flattenQuadratic(path, current, c, end)
				current = end
			}

		case 'Z':
			path.Close()
			current = start

		default:
			return nil, fmt.Errorf("svg path: unsupported command %q", cmd)
		}
	}

	return path, nil
}

func svgPathTokens(data string) []string {
	var tokens []string
	for i := 0; i < len(data); {
		r := rune(data[i])
		if unicode.IsSpace(r) || data[i] == ',' {
			i++
			continue
		}
		if isCommandByte(data[i]) {
			tokens = append(tokens, data[i:i+1])
			i++
			continue
		}
		if unicode.IsLetter(r) {
			tokens = append(tokens, data[i:i+1])
			i++
			continue
		}

		start := i
		if data[i] == '-' || data[i] == '+' {
			i++
		}
		for i < len(data) && ((data[i] >= '0' && data[i] <= '9') || data[i] == '.') {
			i++
		}
		if i < len(data) && (data[i] == 'e' || data[i] == 'E') {
			i++
			if i < len(data) && (data[i] == '-' || data[i] == '+') {
				i++
			}
			for i < len(data) && data[i] >= '0' && data[i] <= '9' {
				i++
			}
		}
		if start == i {
			i++
			continue
		}
		tokens = append(tokens, data[start:i])
	}
	return tokens
}

func isCommandToken(token string) bool {
	return len(token) == 1 && unicode.IsLetter(rune(token[0]))
}

func isCommandByte(b byte) bool {
	switch b {
	case 'M', 'm', 'L', 'l', 'H', 'h', 'V', 'v', 'C', 'c', 'Q', 'q', 'Z', 'z':
		return true
	default:
		return false
	}
}

func readPair(tokens []string, i int) (float32, float32, int, error) {
	values, next, err := readValues(tokens, i, 2)
	if err != nil {
		return 0, 0, i, err
	}
	return values[0], values[1], next, nil
}

func readValues(tokens []string, i, count int) ([]float32, int, error) {
	if i+count > len(tokens) {
		return nil, i, fmt.Errorf("svg path: expected %d values", count)
	}
	values := make([]float32, count)
	for n := 0; n < count; n++ {
		if isCommandToken(tokens[i+n]) {
			return nil, i, fmt.Errorf("svg path: expected number, got command %q", tokens[i+n])
		}
		value, err := parseFloat32(tokens[i+n])
		if err != nil {
			return nil, i, err
		}
		values[n] = value
	}
	return values, i + count, nil
}

func parseFloat32(token string) (float32, error) {
	value, err := strconv.ParseFloat(token, 32)
	if err != nil {
		return 0, fmt.Errorf("svg path: parse number %q: %w", token, err)
	}
	return float32(value), nil
}

func resolvePoint(current math.Vec2, x, y float32, relative bool) math.Vec2 {
	if relative {
		return math.NewVec2(current.X+x, current.Y+y)
	}
	return math.NewVec2(x, y)
}

func flattenCubic(path *Path, p0, p1, p2, p3 math.Vec2) {
	for i := 1; i <= curveSegments; i++ {
		t := float32(i) / curveSegments
		mt := 1 - t
		x := mt*mt*mt*p0.X + 3*mt*mt*t*p1.X + 3*mt*t*t*p2.X + t*t*t*p3.X
		y := mt*mt*mt*p0.Y + 3*mt*mt*t*p1.Y + 3*mt*t*t*p2.Y + t*t*t*p3.Y
		path.LineTo(x, y)
	}
}

func flattenQuadratic(path *Path, p0, p1, p2 math.Vec2) {
	for i := 1; i <= curveSegments; i++ {
		t := float32(i) / curveSegments
		mt := 1 - t
		x := mt*mt*p0.X + 2*mt*t*p1.X + t*t*p2.X
		y := mt*mt*p0.Y + 2*mt*t*p1.Y + t*t*p2.Y
		path.LineTo(x, y)
	}
}
