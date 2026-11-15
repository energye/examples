//go:build windows

package main

// 真实窗口测试的 Win32 绑定：投递鼠标 / 键盘 / 滚轮消息，并把客户区像素
// 抓成 PNG。全部走真实的 HWND，不绕过 LCL 的消息分发。

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	wtUser32 = windows.NewLazySystemDLL("user32.dll")
	wtGdi32  = windows.NewLazySystemDLL("gdi32.dll")

	wtPostMessageW  = wtUser32.NewProc("PostMessageW")
	wtSendMessageW  = wtUser32.NewProc("SendMessageW")
	wtGetClientRect = wtUser32.NewProc("GetClientRect")
	wtGetDC         = wtUser32.NewProc("GetDC")
	wtReleaseDC     = wtUser32.NewProc("ReleaseDC")
	wtPrintWindow   = wtUser32.NewProc("PrintWindow")

	wtCreateCompatibleDC = wtGdi32.NewProc("CreateCompatibleDC")
	wtCreateDIBSection   = wtGdi32.NewProc("CreateDIBSection")
	wtSelectObject       = wtGdi32.NewProc("SelectObject")
	wtDeleteObject       = wtGdi32.NewProc("DeleteObject")
	wtDeleteDC           = wtGdi32.NewProc("DeleteDC")
	wtBitBlt             = wtGdi32.NewProc("BitBlt")
)

func winTestSupported() bool { return true }

func winPostMouse(hwnd uintptr, msg uint32, x, y int32, down bool) bool {
	var wp uintptr
	if down {
		wp = 0x0001 // MK_LBUTTON
	}
	lp := uintptr(uint16(x)) | uintptr(uint16(y))<<16
	r, _, _ := wtPostMessageW.Call(hwnd, uintptr(msg), wp, lp)
	return r != 0
}

func winPostKey(hwnd uintptr, vk uint16) {
	wtPostMessageW.Call(hwnd, winWMKeyDown, uintptr(vk), 1)
	wtPostMessageW.Call(hwnd, winWMKeyUp, uintptr(vk), 0xC0000001)
}

// winPostWheel 用 SendMessage 同步投递：滚轮处理会立即改缩放，便于下一步断言。
func winPostChar(hwnd uintptr, ch uint32) {
	wtPostMessageW.Call(hwnd, winWMChar, uintptr(ch), 1)
}

func winPostWheel(hwnd uintptr, screenX, screenY int32, delta int16) {
	wp := uintptr(uint16(delta)) << 16
	lp := uintptr(uint16(screenX)) | uintptr(uint16(screenY))<<16
	wtSendMessageW.Call(hwnd, winWMMouseWheel, wp, lp)
}

type wtBitmapInfoHeader struct {
	Size          uint32
	Width         int32
	Height        int32
	Planes        uint16
	BitCount      uint16
	Compression   uint32
	SizeImage     uint32
	XPelsPerMeter int32
	YPelsPerMeter int32
	ClrUsed       uint32
	ClrImportant  uint32
}

type wtBitmapInfo struct {
	Header wtBitmapInfoHeader
	Colors [1]uint32
}

// winCaptureClient 抓取窗口客户区的真实像素（PrintWindow 优先，BitBlt 兜底）。
func winCaptureClient(hwnd uintptr, path string) error {
	var rc wRect
	if r, _, _ := wtGetClientRect.Call(hwnd, uintptr(unsafe.Pointer(&rc))); r == 0 {
		return fmt.Errorf("GetClientRect failed")
	}
	w, h := int(rc.Right-rc.Left), int(rc.Bottom-rc.Top)
	if w <= 0 || h <= 0 {
		return fmt.Errorf("empty client rect %dx%d", w, h)
	}
	hdc, _, _ := wtGetDC.Call(hwnd)
	if hdc == 0 {
		return fmt.Errorf("GetDC failed")
	}
	defer wtReleaseDC.Call(hwnd, hdc)

	mem, _, _ := wtCreateCompatibleDC.Call(hdc)
	if mem == 0 {
		return fmt.Errorf("CreateCompatibleDC failed")
	}
	defer wtDeleteDC.Call(mem)

	// 用 DIB Section 直接拿到像素缓冲，避免 GetDIBits 在不同驱动上的兼容问题。
	bi := wtBitmapInfo{}
	bi.Header.Size = uint32(unsafe.Sizeof(bi.Header))
	bi.Header.Width = int32(w)
	bi.Header.Height = -int32(h) // 负数 = top-down
	bi.Header.Planes = 1
	bi.Header.BitCount = 32
	bi.Header.Compression = 0 // BI_RGB
	var bits unsafe.Pointer
	bmp, _, _ := wtCreateDIBSection.Call(hdc, uintptr(unsafe.Pointer(&bi)), 0,
		uintptr(unsafe.Pointer(&bits)), 0, 0)
	if bmp == 0 || bits == nil {
		return fmt.Errorf("CreateDIBSection failed")
	}
	defer wtDeleteObject.Call(bmp)

	old, _, _ := wtSelectObject.Call(mem, bmp)
	ok, _, _ := wtPrintWindow.Call(hwnd, mem, 2) // PW_RENDERFULLCONTENT
	if ok == 0 {
		const srcCopy = 0x00CC0020
		if r, _, _ := wtBitBlt.Call(mem, 0, 0, uintptr(w), uintptr(h), hdc, 0, 0, srcCopy); r == 0 {
			wtSelectObject.Call(mem, old)
			return fmt.Errorf("PrintWindow / BitBlt failed")
		}
	}
	buf := unsafe.Slice((*byte)(bits), w*h*4)
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for i := 0; i < w*h; i++ {
		img.Pix[i*4+0] = buf[i*4+2] // R
		img.Pix[i*4+1] = buf[i*4+1] // G
		img.Pix[i*4+2] = buf[i*4+0] // B
		img.Pix[i*4+3] = 255
	}
	wtSelectObject.Call(mem, old)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}
