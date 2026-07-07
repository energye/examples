package main

import (
	"flag"
	"fmt"
	"image"
	_ "image/png"
	"os"
)

func main() {
	file := flag.String("file", "", "PNG file to validate")
	width := flag.Int("width", 0, "expected image width")
	height := flag.Int("height", 0, "expected image height")
	minNonBackground := flag.Int("min-non-bg", 1, "minimum pixels that differ from the top-left background color")
	minColors := flag.Int("min-colors", 2, "minimum distinct RGBA colors")
	flag.Parse()

	if *file == "" {
		fail("missing -file")
	}

	f, err := os.Open(*file)
	if err != nil {
		fail("open %s: %v", *file, err)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		fail("decode %s: %v", *file, err)
	}

	bounds := img.Bounds()
	if *width > 0 && bounds.Dx() != *width {
		fail("width = %d, want %d", bounds.Dx(), *width)
	}
	if *height > 0 && bounds.Dy() != *height {
		fail("height = %d, want %d", bounds.Dy(), *height)
	}

	bg := rgbaKey(img.At(bounds.Min.X, bounds.Min.Y).RGBA())
	colors := make(map[uint64]struct{})
	nonBackground := 0
	nonTransparent := 0

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			key := rgbaKey(img.At(x, y).RGBA())
			colors[key] = struct{}{}
			if key != bg {
				nonBackground++
			}
			if key&0xffff != 0 {
				nonTransparent++
			}
		}
	}

	if nonBackground < *minNonBackground {
		fail("non-background pixels = %d, want >= %d", nonBackground, *minNonBackground)
	}
	if len(colors) < *minColors {
		fail("distinct colors = %d, want >= %d", len(colors), *minColors)
	}
	if nonTransparent == 0 {
		fail("image is fully transparent")
	}

	fmt.Printf("snapshot ok: %s size=%dx%d colors=%d nonBackground=%d\n", *file, bounds.Dx(), bounds.Dy(), len(colors), nonBackground)
}

func rgbaKey(r, g, b, a uint32) uint64 {
	return uint64(r)<<48 | uint64(g)<<32 | uint64(b)<<16 | uint64(a)
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "snapshot validation failed: "+format+"\n", args...)
	os.Exit(1)
}
