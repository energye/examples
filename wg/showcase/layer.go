package main

import (
	"math"
	"sync"
	"time"
	"unsafe"

	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
	"github.com/energye/lcl/types/colors"
)

// Theme is the single source of truth for every color in the demo.
type Theme struct {
	Bg, Panel, Edge, Grid              types.TColor
	Accent, Accent2, Accent3, Warm     types.TColor
	Text, Muted, GlassTop, GlassBottom types.TColor
}

func darkTheme() Theme {
	return Theme{
		Bg:          colors.RGBToColor(14, 16, 22),
		Panel:       colors.RGBToColor(24, 27, 35),
		Edge:        colors.RGBToColor(58, 64, 78),
		Grid:        colors.RGBToColor(40, 45, 57),
		Accent:      colors.RGBToColor(76, 154, 255),
		Accent2:     colors.RGBToColor(140, 96, 255),
		Accent3:     colors.RGBToColor(38, 214, 174),
		Warm:        colors.RGBToColor(255, 158, 92),
		Text:        colors.RGBToColor(232, 236, 244),
		Muted:       colors.RGBToColor(126, 136, 154),
		GlassTop:    colors.RGBToColor(72, 84, 116),
		GlassBottom: colors.RGBToColor(32, 38, 58),
	}
}

func lightTheme() Theme {
	return Theme{
		Bg:          colors.RGBToColor(238, 241, 247),
		Panel:       colors.RGBToColor(255, 255, 255),
		Edge:        colors.RGBToColor(196, 204, 218),
		Grid:        colors.RGBToColor(220, 226, 236),
		Accent:      colors.RGBToColor(44, 108, 220),
		Accent2:     colors.RGBToColor(112, 68, 214),
		Accent3:     colors.RGBToColor(16, 158, 126),
		Warm:        colors.RGBToColor(220, 116, 40),
		Text:        colors.RGBToColor(28, 34, 46),
		Muted:       colors.RGBToColor(96, 108, 128),
		GlassTop:    colors.RGBToColor(214, 224, 246),
		GlassBottom: colors.RGBToColor(166, 182, 216),
	}
}

// Surface is a raw view over ILazIntfImage.PixelData(). R/G/B/A are byte
// offsets inside one pixel; they are DETECTED at runtime with a probe pixel,
// so the demo is not hardcoded to Windows BGRA.
type Surface struct {
	Px           []byte
	W, H, Stride int
	R, G, B, A   int
}

func (s *Surface) offset(x, y int) int { return y*s.Stride + x*4 }

// AlphaLayer owns one RGBA offscreen buffer and one 32-bit bitmap. Pixels are
// written in a single Go-side memory loop (O(1) FFI) instead of one
// ILazIntfImage.SetColors syscall per pixel (O(w*h) FFI).
type AlphaLayer struct {
	img    lcl.ILazIntfImage
	bmp    lcl.IBitmap
	surf   Surface
	Width  int32
	Height int32
}

type BakeFunc func(s *Surface)

func NewAlphaLayer(w, h int32) *AlphaLayer {
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	img := lcl.NewLazIntfImageWithIntX2RIQFlags(w, h,
		types.NewSet(types.RiqfRGB, types.RiqfAlpha))
	img.SetSize(w, h)
	bmp := lcl.NewBitmap()
	bmp.SetPixelFormat(types.Pf32bit)
	bmp.SetSize(w, h)

	desc := img.DataDescription()
	bytesPerPixel := int(desc.BitsPerPixel()) / 8
	if bytesPerPixel < 4 {
		bytesPerPixel = 4
	}
	stride := int(w) * bytesPerPixel
	if h > 1 {
		s0 := uintptr(img.GetDataLineStart(0))
		s1 := uintptr(img.GetDataLineStart(1))
		if s1 > s0 {
			stride = int(s1 - s0)
		}
	}
	ptr := uintptr(img.PixelData())
	var px []byte
	if ptr != 0 {
		px = unsafe.Slice((*byte)(unsafe.Pointer(ptr)), stride*int(h))
	}

	r, g, b, a := 2, 1, 0, 3
	if len(px) >= 4 {
		// Probe with 4 distinguishable channels; TFPColor 16-bit values are
		// quantized to their high byte inside the raw buffer.
		img.SetColors(0, 0, lcl.TFPColor{
			Red: 0x1122, Green: 0x3344, Blue: 0x5566, Alpha: 0x7788,
		})
		for i, v := range px[:4] {
			switch v {
			case 0x11:
				r = i
			case 0x33:
				g = i
			case 0x55:
				b = i
			case 0x77:
				a = i
			}
		}
	}
	return &AlphaLayer{
		img: img, bmp: bmp, Width: w, Height: h,
		surf: Surface{Px: px, W: int(w), H: int(h), Stride: stride, R: r, G: g, B: b, A: a},
	}
}

// Bake writes every pixel in one Go loop, then transfers the buffer to the
// bitmap once. BeginUpdate/EndUpdate bracket the raw write.
func (l *AlphaLayer) Bake(fn BakeFunc) {
	if len(l.surf.Px) == 0 {
		return
	}
	l.img.BeginUpdate()
	fn(&l.surf)
	l.img.EndUpdate()
	l.bmp.LoadFromIntfImage(l.img)
}

func (l *AlphaLayer) Draw(c lcl.ICanvas, x, y int32) {
	if c == nil || !c.IsValid() {
		return
	}
	c.DrawWithIntX2Graphic(x, y, l.bmp)
}

func (l *AlphaLayer) Free() {
	if l.img != nil && l.img.IsValid() {
		l.img.Free()
	}
	if l.bmp != nil && l.bmp.IsValid() {
		l.bmp.Free()
	}
	l.img, l.bmp = nil, nil
	l.surf.Px = nil
}

// LayerCache shares baked layers between widgets with the same style/size.
// Without it, 20 identical glass cards would bake 20 separate bitmaps.
type LayerCache struct {
	mu         sync.Mutex
	pool       map[string]*AlphaLayer
	Hits       int
	Misses     int
	BakeMicros int64
	LastBakeUs int64
}

func NewLayerCache() *LayerCache {
	return &LayerCache{pool: make(map[string]*AlphaLayer)}
}

func (c *LayerCache) Get(key string, w, h int32, bake BakeFunc) *AlphaLayer {
	c.mu.Lock()
	defer c.mu.Unlock()
	if l, ok := c.pool[key]; ok {
		c.Hits++
		return l
	}
	c.Misses++
	l := NewAlphaLayer(w, h)
	t := time.Now()
	l.Bake(bake)
	us := time.Since(t).Microseconds()
	c.LastBakeUs = us
	c.BakeMicros += us
	c.pool[key] = l
	return l
}

// InvalidateAll forces a re-bake (theme switch, opacity change, DPI change).
func (c *LayerCache) InvalidateAll() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for k, l := range c.pool {
		l.Free()
		delete(c.pool, k)
	}
}

func (c *LayerCache) Free() {
	c.InvalidateAll()
}

func (c *LayerCache) Stats() (hits, misses, layers int, bakeUs int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Hits, c.Misses, len(c.pool), c.LastBakeUs
}

// ---- bake functions: all per-pixel work happens in Go memory ----

func byteOf(c types.TColor, channel string) byte {
	switch channel {
	case "r":
		return colors.Red(c)
	case "g":
		return colors.Green(c)
	case "b":
		return colors.Blue(c)
	}
	return 0
}

func lerpByte(a, b byte, t float64) byte {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	return byte(float64(a) + (float64(b)-float64(a))*t)
}

// roundedAlpha returns a smooth 0..1 coverage for anti-aliased corners.
func roundedAlpha(x, y, w, h, radius int) float64 {
	if radius <= 0 {
		return 1
	}
	maxR := w / 2
	if h/2 < maxR {
		maxR = h / 2
	}
	if radius > maxR {
		radius = maxR
	}
	var cx, cy int
	switch {
	case x < radius && y < radius:
		cx, cy = radius, radius
	case x >= w-radius && y < radius:
		cx, cy = w-radius-1, radius
	case x < radius && y >= h-radius:
		cx, cy = radius, h-radius-1
	case x >= w-radius && y >= h-radius:
		cx, cy = w-radius-1, h-radius-1
	default:
		return 1
	}
	dx := float64(x - cx)
	dy := float64(y - cy)
	d := dx*dx + dy*dy
	inner := float64(radius-1) * float64(radius-1)
	outer := float64(radius+1) * float64(radius+1)
	if d <= inner {
		return 1
	}
	if d >= outer {
		return 0
	}
	return 1 - (d-inner)/(outer-inner)
}

// bakeGlass paints an anti-aliased rounded, semi-transparent gradient card
// with a top highlight and a 1px inner rim. Every pixel carries its own alpha.
func bakeGlass(s *Surface, th Theme, radius int, opacity, glow float64) {
	topR, topG, topB := byteOf(th.GlassTop, "r"), byteOf(th.GlassTop, "g"), byteOf(th.GlassTop, "b")
	botR, botG, botB := byteOf(th.GlassBottom, "r"), byteOf(th.GlassBottom, "g"), byteOf(th.GlassBottom, "b")
	ar, ag, ab := byteOf(th.Accent, "r"), byteOf(th.Accent, "g"), byteOf(th.Accent, "b")
	for y := 0; y < s.H; y++ {
		ratio := float64(y) / float64(maxInt(1, s.H-1))
		r := lerpByte(topR, botR, ratio)
		g := lerpByte(topG, botG, ratio)
		b := lerpByte(topB, botB, ratio)
		for x := 0; x < s.W; x++ {
			a := roundedAlpha(x, y, s.W, s.H, radius)
			cr, cg, cb := r, g, b
			// diagonal sheen + top highlight
			sheen := float64(x+y) / float64(maxInt(1, s.W+s.H))
			if sheen < 0.16 {
				k := (0.16 - sheen) / 0.16
				cr = lerpByte(cr, 255, k*0.20)
				cg = lerpByte(cg, 255, k*0.20)
				cb = lerpByte(cb, 255, k*0.20)
			}
			if y < 2 && a > 0 {
				cr, cg, cb = lerpByte(cr, ar, 0.35), lerpByte(cg, ag, 0.35), lerpByte(cb, ab, 0.35)
			}
			// rim glow just inside the rounded edge
			if a > 0 && a < 0.9 {
				cr, cg, cb = lerpByte(cr, ar, glow), lerpByte(cg, ag, glow), lerpByte(cb, ab, glow)
			}
			o := s.offset(x, y)
			s.Px[o+s.B] = cb
			s.Px[o+s.G] = cg
			s.Px[o+s.R] = cr
			s.Px[o+s.A] = byte(a * opacity * 255)
		}
	}
}

// bakeGlow paints a radial, irregular (non-rectangular) translucent blob.
// It is used for the soft light shapes behind the cards.
func bakeGlow(s *Surface, th Theme, opacity float64, lobes int) {
	ar, ag, ab := byteOf(th.Accent2, "r"), byteOf(th.Accent2, "g"), byteOf(th.Accent2, "b")
	wr, wg, wb := byteOf(th.Warm, "r"), byteOf(th.Warm, "g"), byteOf(th.Warm, "b")
	cx, cy := float64(s.W)/2, float64(s.H)/2
	maxD := float64(minInt(s.W, s.H)) / 2
	for y := 0; y < s.H; y++ {
		for x := 0; x < s.W; x++ {
			dx := float64(x) - cx
			dy := float64(y) - cy
			ang := math.Atan2(dy, dx)
			d := math.Sqrt(dx*dx + dy*dy)
			// lobed boundary -> an irregular, non-circular shape
			edge := maxD * (0.66 + 0.20*math.Cos(float64(lobes)*ang))
			var a float64
			if d < edge {
				a = 1 - d/edge
				a = a * a
			}
			t := d / maxD
			o := s.offset(x, y)
			s.Px[o+s.B] = lerpByte(ab, wb, t)
			s.Px[o+s.G] = lerpByte(ag, wg, t)
			s.Px[o+s.R] = lerpByte(ar, wr, t)
			s.Px[o+s.A] = byte(a * opacity * 255)
		}
	}
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
