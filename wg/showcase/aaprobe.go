package main

// --aaprobe is a headless capability/cost probe. It answers with measured
// numbers instead of guesses:
//
//   1. Does TCanvas.AntialiasingMode(AmOn) smooth GDI vector primitives?
//   2. Can TBitmap.ScanLine / CreateIntfImage.PixelData() give a writable
//      raw pixel block, and which one stays in sync with GDI drawing?
//   3. What does a full-window alpha-blended Draw cost vs an opaque BitBlt,
//      and is a "Go pixel buffer -> LoadFromIntfImage -> CopyRect" pipeline
//      cheap enough for 60fps?
//
// Run: go run ./wg/showcase --aaprobe   (no window is created)

import (
	"fmt"
	"time"
	"unsafe"

	"github.com/energye/lcl/api/libname"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
	"github.com/energye/lcl/types/colors"
)

type probeShape int

const (
	psLine probeShape = iota
	psPolygon
	psEllipse
	psArc
	psPie
)

func (k probeShape) String() string {
	return [...]string{"Polyline", "Polygon", "Ellipse", "Arc", "Pie"}[k]
}

func ptrOf(pts []types.TPoint) types.PPoint {
	return types.PPoint(uintptr(unsafe.Pointer(&pts[0])))
}

// countEdgeTones renders one white shape on black and counts pixels that are
// neither pure black nor pure white. Hard GDI edges produce ~0; a real
// antialiasing path produces many grey coverage tones.
func countEdgeTones(mode types.TAntialiasingMode, k probeShape) int {
	const W, H int32 = 260, 140
	b := lcl.NewBitmap()
	b.SetPixelFormat(types.Pf24bit)
	b.SetSize(W, H)
	defer b.Free()
	c := b.Canvas()
	c.SetAntialiasingMode(mode)
	br := c.BrushToBrush()
	br.SetColor(colors.RGBToColor(0, 0, 0))
	br.SetStyle(types.BsSolid)
	c.FillRectWithRect(rect(0, 0, W, H))
	pn := c.PenToPen()
	pn.SetStyle(types.PsSolid)
	pn.SetColor(colors.RGBToColor(255, 255, 255))
	pn.SetWidth(2)
	br.SetColor(colors.RGBToColor(255, 255, 255))
	br.SetStyle(types.BsSolid)
	switch k {
	case psLine:
		pts := []types.TPoint{{X: 12, Y: 16}, {X: 240, Y: 120}}
		c.Polyline(ptrOf(pts), int32(len(pts)))
	case psPolygon:
		pts := []types.TPoint{{X: 20, Y: 120}, {X: 130, Y: 15}, {X: 240, Y: 120}}
		c.Polygon(ptrOf(pts), int32(len(pts)), false)
	case psEllipse:
		c.EllipseWithRect(rect(30, 15, 230, 125))
	case psArc:
		c.ArcWithIntX6(30, 15, 230, 125, 0, 300*16)
	case psPie:
		c.RadialPieWithIntX6(30, 15, 230, 125, 0, 250*16)
	}
	img := b.CreateIntfImage()
	defer img.Free()
	mid := 0
	for y := int32(0); y < H; y++ {
		for x := int32(0); x < W; x++ {
			r := int(img.Colors(x, y).Red >> 8)
			if r > 16 && r < 239 {
				mid++
			}
		}
	}
	return mid
}

func scanBitmap(w, h int32) (lcl.IBitmap, []byte, int) {
	b := lcl.NewBitmap()
	b.SetPixelFormat(types.Pf32bit)
	b.SetSize(w, h)
	p0 := b.ScanLine(0)
	if p0 == 0 {
		return b, nil, 0
	}
	p1 := b.ScanLine(1)
	stride := int(int64(p1) - int64(p0))
	if stride <= 0 {
		return b, nil, 0
	}
	base := int64(p0)
	for y := int32(0); y < h; y += 17 {
		if int64(b.ScanLine(y)) != base+int64(y)*int64(stride) {
			return b, nil, stride
		}
	}
	px := unsafe.Slice((*byte)(unsafe.Pointer(p0)), stride*int(h))
	return b, px, stride
}

func fillRGBA(px []byte, stride, w, h, R, G, B, A int) {
	for y := 0; y < h; y++ {
		o := y * stride
		for x := 0; x < w; x++ {
			px[o+x*4+0] = byte(B)
			px[o+x*4+1] = byte(G)
			px[o+x*4+2] = byte(R)
			px[o+x*4+3] = byte(A)
		}
	}
}

func probeBlend(label string, src lcl.IBitmap) {
	base := newBackBuffer(64, 64)
	defer base.Free()
	bc := base.Canvas()
	br := bc.BrushToBrush()
	br.SetColor(colors.RGBToColor(0, 0, 255))
	br.SetStyle(types.BsSolid)
	bc.FillRectWithRect(rect(0, 0, 64, 64))
	bc.DrawWithIntX2Graphic(0, 0, src)
	img := base.CreateIntfImage()
	defer img.Free()
	col := img.Colors(32, 32)
	fmt.Printf("[AAPROBE] blend %-22s R=%02x G=%02x B=%02x   (混合期望 ~7f/00/7f)\n",
		label, col.Red>>8, col.Green>>8, col.Blue>>8)
}

// probeFrameSurface checks whether the Pf24bit frame buffer's ScanLine block
// is in sync with GDI drawing (spoiler: it is NOT - LCL keeps the FPImage
// buffer and the HDC bitmap separate until LoadFromIntfImage/CreateIntfImage).
func probeFrameSurface() {
	const W, H int32 = 101, 64
	b := newBackBuffer(W, H)
	defer b.Free()
	p0 := b.ScanLine(0)
	p1 := b.ScanLine(1)
	stride := int(int64(p1) - int64(p0))
	fmt.Printf("[AAPROBE] Pf24bit ScanLine(0)=%X stride=%d (W*3=%d)\n", p0, stride, W*3)
	if p0 == 0 || stride <= 0 {
		fmt.Println("[AAPROBE] 帧缓冲 ScanLine 不可用")
		return
	}
	px := unsafe.Slice((*byte)(unsafe.Pointer(p0)), stride*int(H))
	br := b.Canvas().BrushToBrush()
	br.SetColor(colors.RGBToColor(0x11, 0x22, 0x33))
	br.SetStyle(types.BsSolid)
	b.Canvas().FillRectWithRect(rect(0, 0, W, H))
	o := 10*stride + 10*3
	fmt.Printf("[AAPROBE] GDI 填充后 raw[10,10]=% x  (期望 33 22 11 = BGR)\n", px[o:o+3])
	o = 20*stride + 20*3
	px[o], px[o+1], px[o+2] = 0xFF, 0xFF, 0xFF
	img := b.CreateIntfImage()
	col := img.Colors(20, 20)
	img.Free()
	fmt.Printf("[AAPROBE] ScanLine 直写后 Colors(20,20)=%02x/%02x/%02x (期望 ff/ff/ff)\n",
		col.Red>>8, col.Green>>8, col.Blue>>8)
}

// probePipeline validates the proposed AA pipeline:
//
//	GDI static snapshot -> CreateIntfImage().PixelData() (bulk read)
//	-> Go-side coverage blending -> LoadFromIntfImage (bulk write)
//	-> opaque CopyRect to the visible DC
func probePipeline() {
	const W, H int32 = 1440, 880
	const iters = 8

	st := newBackBuffer(W, H)
	defer st.Free()
	br := st.Canvas().BrushToBrush()
	br.SetColor(colors.RGBToColor(0x11, 0x22, 0x33))
	br.SetStyle(types.BsSolid)
	st.Canvas().FillRectWithRect(rect(0, 0, W, H))

	// 1) Bulk read of the GDI-painted snapshot.
	t := time.Now()
	si := st.CreateIntfImage()
	readUs := time.Since(t).Microseconds()
	ptr := uintptr(si.PixelData())
	desc := si.DataDescription()
	s0 := uintptr(si.GetDataLineStart(0))
	s1 := uintptr(si.GetDataLineStart(1))
	stride := int(W) * 4
	if s1 > s0 {
		stride = int(s1 - s0)
	}
	fmt.Printf("[AAPROBE] CreateIntfImage: %.2fms  PixelData=%X stride=%d bpp=%d  line0=%X line1=%X\n",
		float64(readUs)/1000, ptr, stride, desc.BitsPerPixel(), s0, s1)
	col := si.Colors(100, 100)
	fmt.Printf("[AAPROBE] 快照 Colors(100,100)=%02x/%02x/%02x (期望 11/22/33)\n",
		col.Red>>8, col.Green>>8, col.Blue>>8)
	if ptr == 0 {
		si.Free()
		return
	}
	px := unsafe.Slice((*byte)(unsafe.Pointer(ptr)), stride*int(H))
	fmt.Printf("[AAPROBE] 快照 raw[100,100]=% x\n", px[100*stride+100*4:100*stride+100*4+4])

	// Copy the snapshot into a stable Go buffer before freeing si.
	staticPx := make([]byte, len(px))
	copy(staticPx, px)

	// 2) Modify pixels in Go, push them back into the bitmap.
	t = time.Now()
	for i := 0; i < 200000; i++ {
		q := (i%800)*stride + (i%1400)*4
		px[q], px[q+1], px[q+2] = 0xE0, 0x60, 0x20
	}
	blendUs := time.Since(t).Microseconds()
	si.BeginUpdate()
	si.EndUpdate()
	t = time.Now()
	st.LoadFromIntfImage(si)
	loadUs := time.Since(t).Microseconds()
	si.Free()
	check := st.CreateIntfImage()
	c2 := check.Colors(0, 0)
	check.Free()
	fmt.Printf("[AAPROBE] Go 混合 200000 像素=%.2fms  LoadFromIntfImage(全窗)=%.2fms  回读(0,0)=%02x/%02x/%02x\n",
		float64(blendUs)/1000, float64(loadUs)/1000, c2.Red>>8, c2.Green>>8, c2.Blue>>8)

	// 3) Opaque CopyRect (pure BitBlt) instead of alpha-blended Draw.
	dst := newBackBuffer(W, H)
	defer dst.Free()
	all := rect(0, 0, W, H)
	dst.Canvas().CopyRectWithRectX2Canvas(all, st.Canvas(), all)
	t = time.Now()
	for i := 0; i < iters; i++ {
		dst.Canvas().CopyRectWithRectX2Canvas(all, st.Canvas(), all)
	}
	copyUs := time.Since(t).Microseconds() / iters
	fmt.Printf("[AAPROBE] 不透明 CopyRect 全窗=%.2fms/次 (对比 alpha Draw 整窗 24.72ms)\n",
		float64(copyUs)/1000)

	// 4) Go-side memcpy of the static snapshot (the per-frame background).
	dest := make([]byte, len(staticPx))
	t = time.Now()
	for i := 0; i < iters; i++ {
		copy(dest, staticPx)
	}
	memUs := time.Since(t).Microseconds() / iters
	fmt.Printf("[AAPROBE] Go memcpy 静态快照=%.2fms/次  合计预算≈%.2fms\n",
		float64(memUs)/1000, float64(memUs+blendUs+loadUs+copyUs)/1000)
}

func runAAProbe() int {
	lcl.Init()
	fmt.Println("[AAPROBE] runtime:", libname.LibName)

	for _, k := range []probeShape{psLine, psPolygon, psEllipse, psArc, psPie} {
		off := countEdgeTones(types.AmOff, k)
		on := countEdgeTones(types.AmOn, k)
		dont := countEdgeTones(types.AmDontCare, k)
		v := "NO AA (GDI 硬边)"
		if on > off+50 && on > off*3/2 {
			v = "AA ACTIVE"
		}
		fmt.Printf("[AAPROBE] %-9s AmOff=%-5d AmOn=%-5d AmDontCare=%-5d  -> %s\n",
			k.String(), off, on, dont, v)
	}

	b, px, stride := scanBitmap(64, 64)
	if len(px) > 0 {
		o := 10*stride + 10*4
		px[o], px[o+1], px[o+2], px[o+3] = 0x11, 0x22, 0x33, 0x80
		img := b.CreateIntfImage()
		col := img.Colors(10, 10)
		img.Free()
		fmt.Printf("[AAPROBE] Pf32bit ScanLine stride=%d 直写回读 R=%02x G=%02x B=%02x A=%02x (期望 33/22/11/80)\n",
			stride, col.Red>>8, col.Green>>8, col.Blue>>8, col.Alpha>>8)
		fillRGBA(px, stride, 64, 64, 0xFF, 0, 0, 0x80)
		probeBlend("ScanLine 直写 Pf32bit", b)
	} else {
		fmt.Println("[AAPROBE] ScanLine 不可用")
	}
	b.Free()

	ref := NewAlphaLayer(64, 64)
	ref.Bake(func(s *Surface) {
		for y := 0; y < s.H; y++ {
			o := y * s.Stride
			for x := 0; x < s.W; x++ {
				s.Px[o+x*4+s.B], s.Px[o+x*4+s.G] = 0, 0
				s.Px[o+x*4+s.R], s.Px[o+x*4+s.A] = 0xFF, 0x80
			}
		}
	})
	probeBlend("AlphaLayer(基准)", ref.bmp)
	ref.Free()

	probeFrameSurface()
	probePipeline()
	return 0
}
