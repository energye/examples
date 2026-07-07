package pipeline

import (
	"fmt"
	stdmath "math"
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
	var lastCubicControl math.Vec2
	var lastQuadraticControl math.Vec2
	hasCubicControl := false
	hasQuadraticControl := false
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
			hasCubicControl = false
			hasQuadraticControl = false

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
			hasCubicControl = false
			hasQuadraticControl = false

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
			hasCubicControl = false
			hasQuadraticControl = false

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
			hasCubicControl = false
			hasQuadraticControl = false

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
				lastCubicControl = c2
				hasCubicControl = true
				hasQuadraticControl = false
				current = end
			}

		case 'S':
			for i+3 < len(tokens) && !isCommandToken(tokens[i]) {
				values, next, err := readValues(tokens, i, 4)
				if err != nil {
					return nil, err
				}
				i = next
				c1 := current
				if hasCubicControl {
					c1 = reflectPoint(lastCubicControl, current)
				}
				c2 := resolvePoint(current, values[0], values[1], relative)
				end := resolvePoint(current, values[2], values[3], relative)
				flattenCubic(path, current, c1, c2, end)
				lastCubicControl = c2
				hasCubicControl = true
				hasQuadraticControl = false
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
				lastQuadraticControl = c
				hasQuadraticControl = true
				hasCubicControl = false
				current = end
			}

		case 'T':
			for i+1 < len(tokens) && !isCommandToken(tokens[i]) {
				x, y, next, err := readPair(tokens, i)
				if err != nil {
					return nil, err
				}
				i = next
				c := current
				if hasQuadraticControl {
					c = reflectPoint(lastQuadraticControl, current)
				}
				end := resolvePoint(current, x, y, relative)
				flattenQuadratic(path, current, c, end)
				lastQuadraticControl = c
				hasQuadraticControl = true
				hasCubicControl = false
				current = end
			}

		case 'A':
			for i+6 < len(tokens) && !isCommandToken(tokens[i]) {
				values, next, err := readValues(tokens, i, 7)
				if err != nil {
					return nil, err
				}
				i = next
				end := resolvePoint(current, values[5], values[6], relative)
				flattenArc(path, current, values[0], values[1], values[2], values[3] != 0, values[4] != 0, end)
				current = end
				hasCubicControl = false
				hasQuadraticControl = false
			}

		case 'Z':
			path.Close()
			current = start
			hasCubicControl = false
			hasQuadraticControl = false

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
	case 'M', 'm', 'L', 'l', 'H', 'h', 'V', 'v', 'C', 'c', 'S', 's', 'Q', 'q', 'T', 't', 'A', 'a', 'Z', 'z':
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

func reflectPoint(point, around math.Vec2) math.Vec2 {
	return math.NewVec2(around.X*2-point.X, around.Y*2-point.Y)
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

func flattenArc(path *Path, start math.Vec2, rx, ry, rotation float32, largeArc, sweep bool, end math.Vec2) {
	rx = absFloat32(rx)
	ry = absFloat32(ry)
	if rx == 0 || ry == 0 || (start.X == end.X && start.Y == end.Y) {
		path.LineTo(end.X, end.Y)
		return
	}

	phi := float64(rotation) * stdmath.Pi / 180
	cosPhi := stdmath.Cos(phi)
	sinPhi := stdmath.Sin(phi)

	dx := float64(start.X-end.X) / 2
	dy := float64(start.Y-end.Y) / 2
	x1p := cosPhi*dx + sinPhi*dy
	y1p := -sinPhi*dx + cosPhi*dy

	rxf := float64(rx)
	ryf := float64(ry)
	lambda := (x1p*x1p)/(rxf*rxf) + (y1p*y1p)/(ryf*ryf)
	if lambda > 1 {
		scale := stdmath.Sqrt(lambda)
		rxf *= scale
		ryf *= scale
	}

	rx2 := rxf * rxf
	ry2 := ryf * ryf
	x1p2 := x1p * x1p
	y1p2 := y1p * y1p
	denom := rx2*y1p2 + ry2*x1p2
	if denom == 0 {
		path.LineTo(end.X, end.Y)
		return
	}

	sign := 1.0
	if largeArc == sweep {
		sign = -1
	}
	coef := sign * stdmath.Sqrt(stdmath.Max(0, (rx2*ry2-rx2*y1p2-ry2*x1p2)/denom))
	cxp := coef * (rxf * y1p / ryf)
	cyp := coef * (-ryf * x1p / rxf)

	cx := cosPhi*cxp - sinPhi*cyp + float64(start.X+end.X)/2
	cy := sinPhi*cxp + cosPhi*cyp + float64(start.Y+end.Y)/2

	theta1 := vectorAngle(1, 0, (x1p-cxp)/rxf, (y1p-cyp)/ryf)
	delta := vectorAngle((x1p-cxp)/rxf, (y1p-cyp)/ryf, (-x1p-cxp)/rxf, (-y1p-cyp)/ryf)
	if !sweep && delta > 0 {
		delta -= 2 * stdmath.Pi
	}
	if sweep && delta < 0 {
		delta += 2 * stdmath.Pi
	}

	segments := int(stdmath.Ceil(stdmath.Abs(delta) / (stdmath.Pi / 8)))
	if segments < 1 {
		segments = 1
	}
	if segments > 64 {
		segments = 64
	}

	for i := 1; i <= segments; i++ {
		t := theta1 + delta*float64(i)/float64(segments)
		xp := rxf * stdmath.Cos(t)
		yp := ryf * stdmath.Sin(t)
		x := cosPhi*xp - sinPhi*yp + cx
		y := sinPhi*xp + cosPhi*yp + cy
		path.LineTo(float32(x), float32(y))
	}
}

func absFloat32(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}

func vectorAngle(ux, uy, vx, vy float64) float64 {
	dot := ux*vx + uy*vy
	length := stdmath.Sqrt((ux*ux + uy*uy) * (vx*vx + vy*vy))
	if length == 0 {
		return 0
	}
	value := dot / length
	if value < -1 {
		value = -1
	}
	if value > 1 {
		value = 1
	}
	angle := stdmath.Acos(value)
	if ux*vy-uy*vx < 0 {
		return -angle
	}
	return angle
}
