package main

// --screenshot exports the exact composited frame (software AA geometry +
// GDI ClearType text) straight from the offscreen back buffer, bypassing all
// Win32 screen-capture quirks. Useful for A/B comparing WG_AA=0 vs default.

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"strings"
	"time"
	"unsafe"

	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
)

// frameToImage reads the frame bitmap back in bulk (one FFI call for the
// whole pixel block) and converts B,G,R -> NRGBA.
func frameToImage(s *Scene) (*image.NRGBA, error) {
	if s.backBuf == nil || !s.backBuf.IsValid() {
		return nil, fmt.Errorf("back buffer unavailable")
	}
	img := s.backBuf.CreateIntfImage()
	if img == nil || !img.IsValid() {
		return nil, fmt.Errorf("CreateIntfImage failed")
	}
	defer img.Free()
	r := types.TRect{Left: 0, Top: 0, Right: int32(s.aaW), Bottom: int32(s.aaH)}
	if r.Right <= 0 {
		r = s.host.ClientRect()
	}
	w, h := int(r.Width()), int(r.Height())
	ptr := uintptr(img.PixelData())
	if ptr == 0 || w < 1 || h < 1 {
		return nil, fmt.Errorf("PixelData unavailable")
	}
	desc := img.DataDescription()
	bpp := int(desc.BitsPerPixel()) / 8
	if bpp < 3 {
		bpp = 3
	}
	s0 := uintptr(img.GetDataLineStart(0))
	s1 := uintptr(img.GetDataLineStart(1))
	stride := w * bpp
	if s1 > s0 {
		stride = int(s1 - s0)
	}
	px := unsafe.Slice((*byte)(unsafe.Pointer(ptr)), stride*h)
	out := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		so := y * stride
		do := y * out.Stride
		for x := 0; x < w; x++ {
			out.Pix[do+x*4+0] = px[so+x*bpp+2] // R
			out.Pix[do+x*4+1] = px[so+x*bpp+1] // G
			out.Pix[do+x*4+2] = px[so+x*bpp+0] // B
			out.Pix[do+x*4+3] = 255
		}
	}
	return out, nil
}

func savePNG(img *image.NRGBA, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

// zoomRegion produces a nearest-neighbour enlarged crop for inspecting edges.
func zoomRegion(src *image.NRGBA, x0, y0, x1, y1, scale int) *image.NRGBA {
	if x1 > src.Bounds().Dx() {
		x1 = src.Bounds().Dx()
	}
	if y1 > src.Bounds().Dy() {
		y1 = src.Bounds().Dy()
	}
	w, h := (x1-x0)*scale, (y1-y0)*scale
	out := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < y1-y0; y++ {
		for x := 0; x < x1-x0; x++ {
			c := src.NRGBAAt(x0+x, y0+y)
			for dy := 0; dy < scale; dy++ {
				base := (y*scale+dy)*out.Stride + (x * scale * 4)
				for dx := 0; dx < scale; dx++ {
					o := base + dx*4
					out.Pix[o], out.Pix[o+1], out.Pix[o+2], out.Pix[o+3] = c.R, c.G, c.B, 255
				}
			}
		}
	}
	return out
}

func startScreenshot(s *Scene, path string) {
	go func() {
		// Wait for the frame pump to be running and a few real frames to land.
		for s.Frame < 30 {
			time.Sleep(10 * time.Millisecond)
		}
		// WG_SIZE=WxH resizes the real window before exporting, so the
		// adaptive layout can be verified at a non-design size.
		if sz := os.Getenv("WG_SIZE"); sz != "" {
			var w, h int32
			if _, err := fmt.Sscanf(sz, "%dx%d", &w, &h); err == nil && w > 0 && h > 0 {
				done := make(chan struct{})
				lcl.RunOnMainThreadAsync(func(id uint32) {
					mainForm.SetWidth(w)
					mainForm.SetHeight(h)
					close(done)
				})
				<-done
				want := s.Frame + 12
				for s.Frame < want {
					time.Sleep(10 * time.Millisecond)
				}
				fmt.Printf("[SCREENSHOT] resized to %dx%d, layout scale=%.3f\n", w, h, s.layout.scale)
			}
		}
		// WG_BAKE=1 reproduces the "drag the opacity slider then the 9th card
		// goes blank" path: a static re-bake INSIDE OnPaint. Export the frame
		// AFTER it so the screenshot can prove the snapshot is complete.
		if os.Getenv("WG_BAKE") != "" {
			done := make(chan struct{})
			lcl.RunOnMainThreadAsync(func(id uint32) {
				s.SetOpacity(0.45)
				close(done)
			})
			<-done
			want := s.Frame + 10
			for s.Frame < want {
				time.Sleep(10 * time.Millisecond)
			}
			fmt.Printf("[SCREENSHOT] static re-bake done (gen=%d captured=%d opacity=%.2f)\n",
				s.staticGen, s.capturedGen, s.Opacity)
		}
		lcl.RunOnMainThreadAsync(func(id uint32) {
			img, err := frameToImage(s)
			if err != nil {
				fmt.Println("[SCREENSHOT] error:", err)
				requestExit(1)
				return
			}
			if err := savePNG(img, path); err != nil {
				fmt.Println("[SCREENSHOT] save:", err)
				requestExit(1)
				return
			}
			// A 3x crop of the waveform/gauge area makes edge quality obvious.
			zoomPath := strings.TrimSuffix(path, ".png") + "_zoom.png"
			if err := savePNG(zoomRegion(img, 30, 130, 470, 340, 3), zoomPath); err != nil {
				fmt.Println("[SCREENSHOT] zoom:", err)
			}
			fmt.Printf("[SCREENSHOT] %dx%d -> %s (+ %s)\n",
				img.Bounds().Dx(), img.Bounds().Dy(), path, zoomPath)
			requestExit(0)
		})
	}()
}
