package main

// --screenshot：无窗口离屏渲染整帧并导出 PNG。
// 用 IBitmap(Pf24bit).Canvas() 当 Painter 的画布，走与真实窗口完全相同的
// PaintAll 路径；再按 showcase 的批量像素读回方式把 BGR 转成 NRGBA 保存。
// 这样可以在没有桌面的环境下验证布局、配色和自绘组件是否正常。

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"unsafe"

	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
)

func headlessScreenshot(path string, w, h int32) error {
	s := NewScene(nil) // form == nil：不创建 HWND，纯 Go 场景
	s.Layout(w, h)

	b := lcl.NewBitmap()
	if b == nil || !b.IsValid() {
		return fmt.Errorf("NewBitmap failed")
	}
	b.SetPixelFormat(types.Pf24bit)
	b.SetSize(w, h)
	defer b.Free()

	c := b.Canvas()
	if c == nil || !c.IsValid() {
		return fmt.Errorf("bitmap canvas unavailable")
	}
	s.painter.Canvas = c
	s.PaintAll(s.painter)

	img, err := bitmapToImage(b, w, h)
	if err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func bitmapToImage(b lcl.IBitmap, w, h int32) (*image.NRGBA, error) {
	ii := b.CreateIntfImage()
	if ii == nil || !ii.IsValid() {
		return nil, fmt.Errorf("CreateIntfImage failed")
	}
	defer ii.Free()
	ptr := uintptr(ii.PixelData())
	desc := ii.DataDescription()
	bpp := int(desc.BitsPerPixel()) / 8
	if bpp < 3 {
		bpp = 3
	}
	s0 := uintptr(ii.GetDataLineStart(0))
	s1 := uintptr(ii.GetDataLineStart(1))
	stride := int(w) * bpp
	if s1 > s0 {
		stride = int(s1 - s0)
	}
	if ptr == 0 || w < 1 || h < 1 || stride < int(w)*3 {
		return nil, fmt.Errorf("pixel data unavailable")
	}
	// 与 showcase/screenshot.go 相同的批量 DIB 读回：PixelData() 返回 uintptr，
	// 这里刻意转回指针一次性读整块像素（go vet 的 possible misuse 属已知误报）。
	px := unsafe.Slice((*byte)(unsafe.Pointer(ptr)), stride*int(h))
	out := image.NewNRGBA(image.Rect(0, 0, int(w), int(h)))
	for y := 0; y < int(h); y++ {
		so := y * stride
		do := y * out.Stride
		for x := 0; x < int(w); x++ {
			out.Pix[do+x*4+0] = px[so+x*bpp+2] // R
			out.Pix[do+x*4+1] = px[so+x*bpp+1] // G
			out.Pix[do+x*4+2] = px[so+x*bpp+0] // B
			out.Pix[do+x*4+3] = 255
		}
	}
	return out, nil
}
