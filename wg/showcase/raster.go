package main

// raster.go is a pure-Go antialiasing rasteriser.
//
// GDI (the only backend available through this LCL build) draws HARD edges:
// the probe --aaprobe measured exactly ZERO intermediate tones for Polyline,
// Polygon, Ellipse, Arc and Pie with SetAntialiasingMode(AmOn) - the LCL
// property is ignored by this widgetset. Compositing an extra RGBA overlay is
// not an option either: a full-window alpha-blended Draw costs 24.72ms, which
// alone breaks the 16.67ms frame budget.
//
// So instead we blend coverage DIRECTLY into the opaque Pf24bit frame buffer:
//
//   GDI static snapshot --CreateIntfImage().PixelData()--> staticPx (Go []byte)
//   every frame: memcpy(staticPx) -> software geometry with sub-pixel
//   coverage -> LoadFromIntfImage (1.09ms) -> GDI text -> opaque CopyRect
//   (0.48ms) to the visible DC. No per-frame alpha compositing at all.
//
// Every primitive below is a read-modify-write over a BOUNDED number of
// pixels; cost scales with the ink, not with the window area.

import (
	"math"
	"sort"
	"unsafe"

	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
	"github.com/energye/lcl/types/colors"
)

// FramePixels is a raw view over an ILazIntfImage's 24bpp pixel block.
// Byte order is probed at runtime (Windows DIB is B,G,R); the offsets are
// kept explicit so a future 32bpp backend only needs a different bind().
type FramePixels struct {
	px         []byte
	stride     int
	W, H       int
	bO, gO, rO int
	img        lcl.ILazIntfImage // owns px; must outlive the view
	row        []float32         // reusable per-scanline coverage
	xs         []float64         // reusable polygon edge intersections
	edges      []rasterEdge      // reusable edge list, sorted by yLo
	active     []rasterEdge      // reusable active edge table
	clip       types.TRect
	hasClip    bool
	coreHalf   float64 // >0: stroke has a crisp core inside a wider glow halo
	glowAmt    float32 // halo strength; both layers share one distance pass
	bpp        int
	ProbedBGR  bool // whether SetColors actually reached the raw block

	// A stroked polyline is ONE shape. Overlapping round caps used to be
	// alpha-over composited repeatedly at every vertex, which made the edge
	// coverage non-monotonic and produced a visible "bead chain" / jaggies.
	// During a stroke, blend() redirects into this coverage accumulator and
	// takes MAX; strokeEnd() blends each pixel exactly once.
	strokeCov     []float32
	strokeStamp   []uint32 // generation tag: avoids clearing the whole box
	strokeTouched []int32  // indices actually covered, visited once at the end
	strokeGen     uint32
	sx0, sy0      int
	sw            int
	strokeMode    bool
}

// rasterEdge is one non-horizontal polygon edge, pre-divided so a scanline
// intersection is a single multiply-add. An active-edge table keeps the cost
// of an N-point polygon at O(N log N + height*activeEdges) instead of O(N*height).
type rasterEdge struct {
	yLo, yHi float64
	x0, y0   float64
	dxdy     float64
}

const frameSubSamples = 2 // vertical supersampling for polygon edges (2x is visually sufficient)

func unpackColor(c types.TColor) (r, g, b byte) {
	return colors.Red(c), colors.Green(c), colors.Blue(c)
}

func clampCov(v float64) float32 {
	if v <= 0 {
		return 0
	}
	if v >= 1 {
		return 1
	}
	return float32(v)
}

// lineCov folds a wide, dim halo and the crisp core into ONE coverage value.
// Drawing those as two passes doubles the per-pixel distance field for the
// whole curve (the dominant cost of a glowing chart). Because both layers use
// the same colour, ordinary over-compositing collapses them into
// core + glow - core*glow, computed from a single distance d.
func (f *FramePixels) lineCov(d, half float64) float32 {
	outer := clampCov(half + 0.5 - d)
	if outer <= 0 || f.coreHalf <= 0 {
		return outer
	}
	g := outer * f.glowAmt
	core := clampCov(f.coreHalf + 0.5 - d)
	return core + g - core*g
}

// bind maps the whole DIB block of img into one Go slice. It is called once
// per (re)allocation of the frame buffer, never per frame.
func (f *FramePixels) bind(img lcl.ILazIntfImage, w, h int32) bool {
	ptr := uintptr(img.PixelData())
	if ptr == 0 {
		return false
	}
	s0 := uintptr(img.GetDataLineStart(0))
	s1 := uintptr(img.GetDataLineStart(1))
	desc := img.DataDescription()
	bpp := int(desc.BitsPerPixel()) / 8
	if bpp < 3 {
		bpp = 3
	}
	stride := int(w) * bpp
	if s1 > s0 {
		stride = int(s1 - s0)
	}
	if stride*int(h) <= 0 {
		return false
	}
	f.img, f.bpp = img, bpp
	f.W, f.H, f.stride = int(w), int(h), stride
	f.px = unsafe.Slice((*byte)(unsafe.Pointer(ptr)), stride*int(h))
	f.bO, f.gO, f.rO = 0, 1, 2
	f.coreHalf, f.glowAmt = 0, 0
	f.hasClip = false
	if cap(f.row) < int(w) {
		f.row = make([]float32, int(w))
	} else {
		f.row = f.row[:int(w)]
	}
	if cap(f.xs) < 64 {
		f.xs = make([]float64, 0, 64)
	}
	f.probeChannels()
	return true
}

// probeChannels writes one pixel with four distinct channels and reads the
// RAW bytes back, exactly like AlphaLayer does for its 32bpp buffer.
func (f *FramePixels) probeChannels() {
	if len(f.px) < 4 {
		return
	}
	orig := [4]byte{}
	copy(orig[:], f.px[:4])
	f.img.SetColors(0, 0, lcl.TFPColor{Red: 0x1122, Green: 0x3344, Blue: 0x5566, Alpha: 0xffff})
	r, g, b := -1, -1, -1
	for i := 0; i < f.bpp && i < 4; i++ {
		switch f.px[i] {
		case 0x11:
			r = i
		case 0x33:
			g = i
		case 0x55:
			b = i
		}
	}
	f.ProbedBGR = r == 2 && g == 1 && b == 0
	if r >= 0 && g >= 0 && b >= 0 {
		f.rO, f.gO, f.bO = r, g, b
	}
	copy(f.px[:4], orig[:])
}

// setClip limits coverage blending to a rectangle (idle dirty frames).
func (f *FramePixels) setClip(r types.TRect) {
	f.clip, f.hasClip = r, true
}
func (f *FramePixels) clearClip() { f.hasClip = false }

// blend is the single hot inner loop: read the existing opaque background,
// mix in the primitive colour by coverage, write back. No alpha channel is
// involved, so the final PRESENT stays a plain BitBlt.
func (f *FramePixels) blend(x, y int, r, g, b byte, cov float32) {
	if len(f.px) == 0 || cov <= 0 || x < 0 || y < 0 || x >= f.W || y >= f.H {
		return
	}
	if f.hasClip && (int32(x) < f.clip.Left || int32(x) >= f.clip.Right ||
		int32(y) < f.clip.Top || int32(y) >= f.clip.Bottom) {
		return
	}
	if f.strokeMode {
		f.putCov(x, y, cov)
		return
	}
	o := y*f.stride + x*f.bpp
	if cov >= 0.999 {
		f.px[o+f.bO], f.px[o+f.gO], f.px[o+f.rO] = b, g, r
		return
	}
	ia := int(cov*256 + 0.5)
	iv := 256 - ia
	pb := &f.px[o+f.bO]
	pg := &f.px[o+f.gO]
	pr := &f.px[o+f.rO]
	*pb = byte((int(*pb)*iv + int(b)*ia) >> 8)
	*pg = byte((int(*pg)*iv + int(g)*ia) >> 8)
	*pr = byte((int(*pr)*iv + int(r)*ia) >> 8)
}

// ---- stroke coverage accumulator ----

// strokeBegin allocates (or reuses) a coverage bitmap covering the stroke's
// bounding box plus the halo padding, and redirects blend() into it.
func (f *FramePixels) strokeBegin(pts []types.TPoint, half float64) bool {
	minX, minY := int(pts[0].X), int(pts[0].Y)
	maxX, maxY := minX, minY
	for _, p := range pts[1:] {
		if int(p.X) < minX {
			minX = int(p.X)
		}
		if int(p.X) > maxX {
			maxX = int(p.X)
		}
		if int(p.Y) < minY {
			minY = int(p.Y)
		}
		if int(p.Y) > maxY {
			maxY = int(p.Y)
		}
	}
	// segment() walks up to (half+1)/cos+1 pixels perpendicular to the line;
	// at 45 degrees cos=1/sqrt(2). Pad for that worst case, otherwise the
	// halo is clipped exactly at the bends where the slope crosses +/-1.
	pad := int(math.Ceil(math.Sqrt2*(half+1.5))) + 2
	x0, y0 := minX-pad, minY-pad
	x1, y1 := maxX+pad, maxY+pad
	if x0 < 0 {
		x0 = 0
	}
	if y0 < 0 {
		y0 = 0
	}
	if x1 >= f.W {
		x1 = f.W - 1
	}
	if y1 >= f.H {
		y1 = f.H - 1
	}
	if x1 < x0 || y1 < y0 {
		return false
	}
	w := x1 - x0 + 1
	n := w * (y1 - y0 + 1)
	if cap(f.strokeCov) < n {
		f.strokeCov = make([]float32, n)
	} else {
		f.strokeCov = f.strokeCov[:n]
	}
	if cap(f.strokeStamp) < n {
		f.strokeStamp = make([]uint32, n)
	} else {
		f.strokeStamp = f.strokeStamp[:n]
	}
	// Generation tags make stale values irrelevant, so no per-frame clearing
	// of the (potentially 50k entry) box is needed.
	f.strokeGen++
	if f.strokeGen == 0 {
		for i := range f.strokeStamp {
			f.strokeStamp[i] = 0
		}
		f.strokeGen = 1
	}
	f.strokeTouched = f.strokeTouched[:0]
	f.sx0, f.sy0, f.sw = x0, y0, w
	f.strokeMode = true
	return true
}

// putCov records the MAX coverage seen for a pixel. Overlapping round caps
// therefore no longer brighten the vertex: the stroke stays a single shape.
func (f *FramePixels) putCov(x, y int, cov float32) {
	lx, ly := x-f.sx0, y-f.sy0
	if lx < 0 || ly < 0 || lx >= f.sw {
		return
	}
	i := ly*f.sw + lx
	if i >= len(f.strokeCov) {
		return
	}
	if f.strokeStamp[i] != f.strokeGen {
		f.strokeStamp[i] = f.strokeGen
		f.strokeCov[i] = cov
		f.strokeTouched = append(f.strokeTouched, int32(i))
	} else if cov > f.strokeCov[i] {
		f.strokeCov[i] = cov
	}
}

// strokeEnd blends every touched pixel exactly once. The sparse index list
// keeps the final sweep proportional to the stroke area, not the bounding box.
func (f *FramePixels) strokeEnd(r, g, b byte) {
	f.strokeMode = false
	for _, i := range f.strokeTouched {
		if cov := f.strokeCov[i]; cov > 0 {
			lx := int(i) % f.sw
			ly := int(i) / f.sw
			f.blend(f.sx0+lx, f.sy0+ly, r, g, b, cov)
		}
	}
}

// copyFrom restores the baked static background. Passing a rect copies only
// that region (idle dirty frames); nil copies the whole window.
func (f *FramePixels) copyFrom(src []byte, r *types.TRect) {
	if len(f.px) == 0 || len(src) < len(f.px) {
		return
	}
	x0, y0, x1, y1 := 0, 0, f.W, f.H
	if r != nil {
		x0, y0 = int(r.Left), int(r.Top)
		x1, y1 = int(r.Right), int(r.Bottom)
		if x0 < 0 {
			x0 = 0
		}
		if y0 < 0 {
			y0 = 0
		}
		if x1 > f.W {
			x1 = f.W
		}
		if y1 > f.H {
			y1 = f.H
		}
	}
	n := (x1 - x0) * f.bpp
	if n <= 0 {
		return
	}
	for y := y0; y < y1; y++ {
		o := y*f.stride + x0*f.bpp
		copy(f.px[o:o+n], src[o:o+n])
	}
}

// ---- lines (distance-field Wu-style coverage, any width, round caps) ----

func (f *FramePixels) disc(cx, cy, half float64, r, g, b byte) {
	outer := half + 0.5
	xa, xb := int(math.Floor(cx-outer)), int(math.Ceil(cx+outer))
	ya, yb := int(math.Floor(cy-outer)), int(math.Ceil(cy+outer))
	outer2 := outer * outer
	for y := ya; y <= yb; y++ {
		dy := float64(y) + 0.5 - cy
		for x := xa; x <= xb; x++ {
			dx := float64(x) + 0.5 - cx
			d2 := dx*dx + dy*dy
			if d2 >= outer2 {
				continue
			}
			f.blend(x, y, r, g, b, f.lineCov(math.Sqrt(d2), half))
		}
	}
}

// axisLine is the common case (chart grids, dividers, bar edges): a purely
// horizontal or vertical segment has CONSTANT coverage along its length, so
// the per-pixel distance field can be skipped entirely.
func (f *FramePixels) axisLine(x0, y0, x1, y1 float64, r, g, b byte, half float64) bool {
	if y0 == y1 {
		xa, xb := int(math.Floor(math.Min(x0, x1))), int(math.Ceil(math.Max(x0, x1)))
		ya, yb := int(math.Floor(y0-half-0.5)), int(math.Ceil(y0+half+0.5))
		for y := ya; y <= yb; y++ {
			cov := f.lineCov(math.Abs(float64(y)+0.5-y0), half)
			if cov <= 0 {
				continue
			}
			for x := xa; x <= xb; x++ {
				f.blend(x, y, r, g, b, cov)
			}
		}
		return true
	}
	if x0 == x1 {
		ya, yb := int(math.Floor(math.Min(y0, y1))), int(math.Ceil(math.Max(y0, y1)))
		xa, xb := int(math.Floor(x0-half-0.5)), int(math.Ceil(x0+half+0.5))
		for x := xa; x <= xb; x++ {
			cov := f.lineCov(math.Abs(float64(x)+0.5-x0), half)
			if cov <= 0 {
				continue
			}
			for y := ya; y <= yb; y++ {
				f.blend(x, y, r, g, b, cov)
			}
		}
		return true
	}
	return false
}

func (f *FramePixels) segment(x0, y0, x1, y1 float64, r, g, b byte, half float64) {
	dx, dy := x1-x0, y1-y0
	l2 := dx*dx + dy*dy
	if l2 < 1e-12 {
		f.disc(x0, y0, half, r, g, b)
		return
	}
	if f.axisLine(x0, y0, x1, y1, r, g, b, half) {
		return
	}
	L := math.Sqrt(l2)
	pad := half + 1.5
	if math.Abs(dx) >= math.Abs(dy) {
		if x0 > x1 {
			x0, y0, x1, y1 = x1, y1, x0, y0
			dx, dy = -dx, -dy
		}
		cos := math.Abs(dx) / L
		xa := int(math.Floor(x0 - pad))
		xb := int(math.Ceil(x1 + pad))
		span := (half+1)/cos + 1
		for x := xa; x <= xb; x++ {
			px := float64(x) + 0.5
			t := (px - x0) / dx
			if t < 0 {
				t = 0
			} else if t > 1 {
				t = 1
			}
			yc := y0 + t*dy
			ya, yb := int(math.Floor(yc-span)), int(math.Ceil(yc+span))
			for y := ya; y <= yb; y++ {
				py := float64(y) + 0.5
				ux, uy := px-x0, py-y0
				// Distance to the INFINITE line is a cross product / L (no
				// sqrt). Only the rare pixels beyond an endpoint need Hypot.
				t := (ux*dx + uy*dy) / l2
				var d float64
				if t < 0 {
					d = math.Hypot(ux, uy)
				} else if t > 1 {
					d = math.Hypot(px-x1, py-y1)
				} else {
					d = math.Abs(ux*dy-uy*dx) / L
				}
				// MUST use lineCov here too. The other branch did, but this one
				// used the raw halo coverage: horizontal stretches of a glowing
				// curve became a fully opaque 6-8px band while steep stretches
				// rendered as a 2px core plus a dim halo. The width flipped at
				// |dx|==|dy|, so every bend looked broken / jagged.
				if cov := f.lineCov(d, half); cov > 0 {
					f.blend(x, y, r, g, b, cov)
				}
			}
		}
		return
	}
	if y0 > y1 {
		x0, y0, x1, y1 = x1, y1, x0, y0
		dx, dy = -dx, -dy
	}
	cos := math.Abs(dy) / L
	ya := int(math.Floor(y0 - pad))
	yb := int(math.Ceil(y1 + pad))
	span := (half+1)/cos + 1
	for y := ya; y <= yb; y++ {
		py := float64(y) + 0.5
		t := (py - y0) / dy
		if t < 0 {
			t = 0
		} else if t > 1 {
			t = 1
		}
		xc := x0 + t*dx
		xa, xb := int(math.Floor(xc-span)), int(math.Ceil(xc+span))
		for x := xa; x <= xb; x++ {
			px := float64(x) + 0.5
			ux, uy := px-x0, py-y0
			t := (ux*dx + uy*dy) / l2
			var d float64
			if t < 0 {
				d = math.Hypot(ux, uy)
			} else if t > 1 {
				d = math.Hypot(px-x1, py-y1)
			} else {
				d = math.Abs(ux*dy-uy*dx) / L
			}
			if cov := f.lineCov(d, half); cov > 0 {
				f.blend(x, y, r, g, b, cov)
			}
		}
	}
}

func (f *FramePixels) polyline(pts []types.TPoint, c types.TColor, width int32) {
	if len(pts) < 2 {
		return
	}
	r, g, b := unpackColor(c)
	half := float64(width) / 2
	for i := 0; i+1 < len(pts); i++ {
		f.segment(float64(pts[i].X), float64(pts[i].Y),
			float64(pts[i+1].X), float64(pts[i+1].Y), r, g, b, half)
	}
}

// polylineGlow draws a halo + crisp core in a SINGLE distance-field pass.
// The halo half-width drives the traversal; lineCov adds the core for free.
// Coverage is accumulated with MAX before blending, so the 180+ round caps
// of a dense curve no longer over-composite into a bead chain / jagged edge.
func (f *FramePixels) polylineGlow(pts []types.TPoint, c types.TColor, coreW, glowW int32, amt float32) {
	if len(pts) < 2 {
		return
	}
	r, g, b := unpackColor(c)
	half := float64(glowW) / 2
	if !f.strokeBegin(pts, half) {
		return
	}
	f.coreHalf = float64(coreW) / 2
	f.glowAmt = amt
	for i := 0; i+1 < len(pts); i++ {
		f.segment(float64(pts[i].X), float64(pts[i].Y),
			float64(pts[i+1].X), float64(pts[i+1].Y), r, g, b, half)
	}
	f.coreHalf, f.glowAmt = 0, 0
	f.strokeEnd(r, g, b)
}

// ---- polygons (scanline, 4x vertical supersampling + analytic x coverage) ----

func (f *FramePixels) polygon(pts []types.TPoint, fill, border types.TColor, borderW int32) {
	f.polygonImpl(pts, fill, fill, border, borderW, false)
}

// polygonGradient adds a vertical two-stop gradient for free: the scanline
// loop already visits every row, so the colour is just one lerp per row.
func (f *FramePixels) polygonGradient(pts []types.TPoint, top, bottom, border types.TColor, borderW int32) {
	f.polygonImpl(pts, top, bottom, border, borderW, true)
}

func (f *FramePixels) polygonImpl(pts []types.TPoint, fill, fill2, border types.TColor, borderW int32, gradient bool) {
	n := len(pts)
	if n < 3 {
		return
	}
	minX, minY := int(pts[0].X), int(pts[0].Y)
	maxX, maxY := minX, minY
	for _, p := range pts[1:] {
		if int(p.X) < minX {
			minX = int(p.X)
		}
		if int(p.X) > maxX {
			maxX = int(p.X)
		}
		if int(p.Y) < minY {
			minY = int(p.Y)
		}
		if int(p.Y) > maxY {
			maxY = int(p.Y)
		}
	}
	minX--
	minY--
	maxX++
	maxY++
	if minX < 0 {
		minX = 0
	}
	if minY < 0 {
		minY = 0
	}
	if maxX >= f.W {
		maxX = f.W - 1
	}
	if maxY >= f.H {
		maxY = f.H - 1
	}
	if minX > maxX || minY > maxY {
		return
	}
	fr, fg, fb := unpackColor(fill)
	fr2, fg2, fb2 := unpackColor(fill2)
	row := f.row[minX : maxX+1]

	// Build, then sort, the non-horizontal edges ONCE. During the scanline
	// loop only the active subset (typically 2-4 edges) is examined.
	f.edges = f.edges[:0]
	for i := 0; i < n; i++ {
		a, b := pts[i], pts[(i+1)%n]
		if a.Y == b.Y {
			continue
		}
		y0, y1 := float64(a.Y), float64(b.Y)
		lo, hi := y0, y1
		if lo > hi {
			lo, hi = hi, lo
		}
		f.edges = append(f.edges, rasterEdge{
			yLo: lo, yHi: hi, x0: float64(a.X), y0: y0,
			dxdy: float64(b.X-a.X) / (y1 - y0),
		})
	}
	sort.Slice(f.edges, func(i, j int) bool { return f.edges[i].yLo < f.edges[j].yLo })
	f.active = f.active[:0]
	next := 0
	for y := minY; y <= maxY; y++ {
		cr, cg, cb := fr, fg, fb
		if gradient {
			t := float64(y-minY) / float64(maxInt(1, maxY-minY))
			cr = lerpByte(fr, fr2, t)
			cg = lerpByte(fg, fg2, t)
			cb = lerpByte(fb, fb2, t)
		}
		for next < len(f.edges) && f.edges[next].yLo <= float64(y)+1 {
			f.active = append(f.active, f.edges[next])
			next++
		}
		w := 0
		for _, e := range f.active {
			if e.yHi >= float64(y) {
				f.active[w] = e
				w++
			}
		}
		f.active = f.active[:w]
		for i := range row {
			row[i] = 0
		}
		if len(f.active) < 2 {
			continue
		}
		for k := 0; k < frameSubSamples; k++ {
			sy := float64(y) + (float64(k)+0.5)/frameSubSamples
			f.xs = f.xs[:0]
			for _, e := range f.active {
				if sy >= e.yLo && sy < e.yHi {
					f.xs = append(f.xs, e.x0+(sy-e.y0)*e.dxdy)
				}
			}
			if len(f.xs) < 2 {
				continue
			}
			sort.Float64s(f.xs)
			for i := 0; i+1 < len(f.xs); i += 2 {
				x0, x1 := f.xs[i], f.xs[i+1]
				if x1 <= x0 {
					continue
				}
				xa, xb := int(math.Floor(x0)), int(math.Ceil(x1))
				if xa < minX {
					xa = minX
				}
				if xb > maxX {
					xb = maxX
				}
				for x := xa; x <= xb; x++ {
					cov := math.Min(float64(x+1), x1) - math.Max(float64(x), x0)
					if cov > 1 {
						cov = 1
					}
					if cov > 0 {
						row[x-minX] += float32(cov / frameSubSamples)
					}
				}
			}
		}
		for x := minX; x <= maxX; x++ {
			if row[x-minX] > 0 {
				f.blend(x, y, cr, cg, cb, clampCov(float64(row[x-minX])))
			}
		}
	}
	if borderW > 0 {
		br, bg, bb := unpackColor(border)
		half := float64(borderW) / 2
		for i := 0; i < n; i++ {
			a, b := pts[i], pts[(i+1)%n]
			f.segment(float64(a.X), float64(a.Y), float64(b.X), float64(b.Y), br, bg, bb, half)
		}
	}
}

// ---- circles / ellipses (signed distance field) ----

func (f *FramePixels) ellipse(rr types.TRect, bg, border types.TColor, borderW float64) {
	cx := (float64(rr.Left) + float64(rr.Right)) / 2
	cy := (float64(rr.Top) + float64(rr.Bottom)) / 2
	rx := float64(rr.Width()) / 2
	ry := float64(rr.Height()) / 2
	if rx < 0.5 || ry < 0.5 {
		return
	}
	scale := math.Min(rx, ry)
	fr, fg, fb := unpackColor(bg)
	br, bg2, bb := unpackColor(border)
	// Only a ~1px annulus needs a sqrt; the interior is a straight solid fill.
	rim := (borderW/2 + 1) / scale
	outer := 1 + rim
	inner := 1 - rim
	if inner < 0 {
		inner = 0
	}
	outer2, inner2 := outer*outer, inner*inner
	x0, x1 := int(math.Floor(cx-rx-1)), int(math.Ceil(cx+rx+1))
	y0, y1 := int(math.Floor(cy-ry-1)), int(math.Ceil(cy+ry+1))
	for y := y0; y <= y1; y++ {
		ny := (float64(y) + 0.5 - cy) / ry
		ny2 := ny * ny
		for x := x0; x <= x1; x++ {
			nx := (float64(x) + 0.5 - cx) / rx
			d2 := nx*nx + ny2
			if d2 > outer2 {
				continue
			}
			if d2 < inner2 {
				f.blend(x, y, fr, fg, fb, 1)
				continue
			}
			edge := (1 - math.Sqrt(d2)) * scale // px inside the rim
			if cov := edge + 0.5; cov > 0 {
				f.blend(x, y, fr, fg, fb, clampCov(cov))
			}
			if borderW > 0 {
				if cov := borderW/2 + 0.5 - math.Abs(edge); cov > 0 {
					f.blend(x, y, br, bg2, bb, clampCov(cov))
				}
			}
		}
	}
}

// sector replaces a per-pixel atan2 with two cross products. Precomputed once
// per primitive; contains() is then four multiplies.
type sector struct {
	sx, sy, ex, ey float64
	start, end     float64
	wide, full     bool
}

func newSector(startDeg, sweepDeg float64) sector {
	start := startDeg * math.Pi / 180
	end := (startDeg + sweepDeg) * math.Pi / 180
	abs := math.Abs(sweepDeg)
	return sector{
		sx: math.Cos(start), sy: math.Sin(start),
		ex: math.Cos(end), ey: math.Sin(end),
		start: start, end: end,
		wide: abs > 180, full: abs >= 359.9,
	}
}

func (s sector) contains(dx, dy float64) bool {
	if s.full {
		return true
	}
	c1 := s.sx*dy - s.sy*dx // cross(start, v) >= 0
	c2 := dx*s.ey - dy*s.ex // cross(v, end) >= 0
	if s.wide {
		return c1 >= 0 || c2 >= 0
	}
	return c1 >= 0 && c2 >= 0
}

// arc draws a thick circular stroke over [startDeg, startDeg+sweepDeg] using
// a radial distance field (no sqrt outside the narrow annulus).
func (f *FramePixels) arc(rr types.TRect, startDeg, sweepDeg float64, c types.TColor, width int32) {
	cx := (float64(rr.Left) + float64(rr.Right)) / 2
	cy := (float64(rr.Top) + float64(rr.Bottom)) / 2
	rad := math.Min(float64(rr.Width()), float64(rr.Height()))/2 - float64(width)/2
	if rad < 1 {
		return
	}
	half := float64(width) / 2
	outer := rad + half + 1
	inner2 := (rad - half - 1) * (rad - half - 1)
	outer2 := outer * outer
	r, g, b := unpackColor(c)
	sec := newSector(startDeg, sweepDeg)
	x0, x1 := int(math.Floor(cx-outer)), int(math.Ceil(cx+outer))
	y0, y1 := int(math.Floor(cy-outer)), int(math.Ceil(cy+outer))
	for y := y0; y <= y1; y++ {
		dy := float64(y) + 0.5 - cy
		for x := x0; x <= x1; x++ {
			dx := float64(x) + 0.5 - cx
			d2 := dx*dx + dy*dy
			if d2 < inner2 || d2 > outer2 || !sec.contains(dx, dy) {
				continue
			}
			d := math.Sqrt(d2)
			if cov := half + 0.5 - math.Abs(d-rad); cov > 0 {
				f.blend(x, y, r, g, b, clampCov(cov))
			}
		}
	}
}

// pie fills a sector with an antialiased outer edge, then strokes the two
// radial borders and the outer rim.
func (f *FramePixels) pie(rr types.TRect, startDeg, sweepDeg float64, fill, border types.TColor) {
	cx := (float64(rr.Left) + float64(rr.Right)) / 2
	cy := (float64(rr.Top) + float64(rr.Bottom)) / 2
	rad := math.Min(float64(rr.Width()), float64(rr.Height())) / 2
	if rad < 1 {
		return
	}
	r, g, b := unpackColor(fill)
	sec := newSector(startDeg, sweepDeg)
	outer2 := (rad + 1) * (rad + 1)
	x0, x1 := int(math.Floor(cx-rad-1)), int(math.Ceil(cx+rad+1))
	y0, y1 := int(math.Floor(cy-rad-1)), int(math.Ceil(cy+rad+1))
	for y := y0; y <= y1; y++ {
		dy := float64(y) + 0.5 - cy
		for x := x0; x <= x1; x++ {
			dx := float64(x) + 0.5 - cx
			d2 := dx*dx + dy*dy
			if d2 > outer2 || !sec.contains(dx, dy) {
				continue
			}
			if cov := rad - math.Sqrt(d2) + 0.5; cov > 0 {
				f.blend(x, y, r, g, b, clampCov(cov))
			}
		}
	}
	br, bg, bb := unpackColor(border)
	half := 0.7
	// radial separators + outer rim
	f.segment(cx, cy, cx+rad*math.Cos(sec.start), cy+rad*math.Sin(sec.start), br, bg, bb, half)
	f.segment(cx, cy, cx+rad*math.Cos(sec.end), cy+rad*math.Sin(sec.end), br, bg, bb, half)
	f.arc(rr, startDeg, sweepDeg, border, 1)
}

// ---- rounded rectangles (analytic corner coverage) ----

func (f *FramePixels) fillRect(rr types.TRect, c types.TColor) {
	r, g, b := unpackColor(c)
	x0, y0 := int(rr.Left), int(rr.Top)
	x1, y1 := int(rr.Right), int(rr.Bottom)
	if x0 < 0 {
		x0 = 0
	}
	if y0 < 0 {
		y0 = 0
	}
	if x1 > f.W {
		x1 = f.W
	}
	if y1 > f.H {
		y1 = f.H
	}
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			f.blend(x, y, r, g, b, 1)
		}
	}
}

func (f *FramePixels) roundRect(rr types.TRect, radius int32, bg, border types.TColor) {
	w, h := int(rr.Width()), int(rr.Height())
	if w < 2 || h < 2 {
		return
	}
	fr, fg, fb := unpackColor(bg)
	br, bg2, bb := unpackColor(border)
	for y := 0; y < h; y++ {
		py := int(rr.Top) + y
		for x := 0; x < w; x++ {
			a := roundedAlpha(x, y, w, h, int(radius))
			if a <= 0 {
				continue
			}
			px := int(rr.Left) + x
			f.blend(px, py, fr, fg, fb, clampCov(a))
			// 1px rim: the difference between the outer and an inset shape.
			var inner float64
			if x >= 1 && y >= 1 && x < w-1 && y < h-1 {
				inner = roundedAlpha(x-1, y-1, w-2, h-2, maxInt(0, int(radius)-1))
			}
			if edge := a - inner; edge > 0 && edge < 1.001 {
				f.blend(px, py, br, bg2, bb, clampCov(edge))
			}
		}
	}
}
