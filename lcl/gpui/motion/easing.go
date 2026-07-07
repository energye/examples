package motion

// Easing maps progress [0,1] to eased progress [0,1].
type Easing func(t float32) float32

var (
	Linear Easing = func(t float32) float32 { return clamp01(t) }

	EaseOut Easing = func(t float32) float32 {
		t = clamp01(t) - 1
		return t*t*t + 1
	}

	EaseIn Easing = func(t float32) float32 {
		t = clamp01(t)
		return t * t * t
	}

	EaseInOut Easing = func(t float32) float32 {
		t = clamp01(t)
		if t < 0.5 {
			return 4 * t * t * t
		}
		p := 2*t - 2
		return 0.5*p*p*p + 1
	}
)

func clamp01(v float32) float32 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
