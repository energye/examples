package window

import (
	"github.com/energye/cef/cef"
	cefTypes "github.com/energye/cef/types"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/pkgs/win"
	"github.com/energye/lcl/rtl"
	"github.com/energye/lcl/tool/ptr"
	"github.com/energye/lcl/types"
	"github.com/energye/lcl/types/colors"
	"math"
	"os"
	"unsafe"
)

var CW CommonWindow

type CommonWindow struct {
	lcl.TEngForm
	bufferPanel cef.IBufferPanel
	chromium    cef.IChromium
}

func (m *CommonWindow) FormCreate(sender lcl.IObject) {
	m.SetWidth(1)
	m.SetHeight(1)
	//m.SetLeft(-1)
	//m.SetTop(-1)
	m.SetBorderStyleToFormBorderStyle(types.BsNone)
	m.SetDoubleBuffered(true)
	//m.SetShowInTaskBar(types.StNever)

	m.chromium = cef.NewChromium(m)
	m.chromiumEvent()

	m.bufferPanel = cef.NewBufferPanel(m)
	m.bufferPanel.SetParent(m)
	m.bufferPanel.SetColor(colors.ClAqua)
	m.bufferPanel.SetTop(0)
	m.bufferPanel.SetLeft(0)
	// 这里设置的宽高还未生效，chromium.SetOnGetViewRect 函数里设置生效
	m.bufferPanel.SetWidth(180)
	m.bufferPanel.SetHeight(40)
	m.bufferPanelEvent()

	m.SetOnShow(func(sender lcl.IObject) { //显示窗口时回调
		// 在这里创建初始化和创建chromium
		m.chromium.Initialized()
		m.chromium.CreateBrowserWithWinControlStringRequestContextDictionaryValue(nil, "", nil, nil)
		m.bufferPanel.CreateIMEHandler()
		m.chromium.InitializeDragAndDrop(m.bufferPanel.Handle())
	})
}

var (
	tempBitMap                   lcl.IBitmap
	tempWidth, tempHeight        int32
	tempLineSize                 int
	tempSrcOffset, tempDstOffset int
	src, dst                     uintptr
)

func (m *CommonWindow) chromiumEvent() {
	m.chromium.SetOnCursorChange(func(sender lcl.IObject, browser cef.ICefBrowser, cursor cefTypes.TCefCursorHandle, cursorType cefTypes.TCefCursorType, customCursorInfo cef.TCefCursorInfo, result *bool) {
		m.bufferPanel.SetCursor(cef.MiscFunc.CefCursorToWindowsCursor(cursorType))
		*result = true
	})
	// 得到显示大小, 这样bufferPanel就显示实际大小
	m.chromium.SetOnGetViewRect(func(sender lcl.IObject, browser cef.ICefBrowser, rect *cef.TCefRect) {
		var scale = float64(m.bufferPanel.ScreenScale())
		*rect = cef.TCefRect{}
		rect.X = 0
		rect.Y = 0
		rect.Width = cef.MiscFunc.DeviceToLogicalWithIntDouble(m.bufferPanel.Width(), scale)
		rect.Height = cef.MiscFunc.DeviceToLogicalWithIntDouble(m.bufferPanel.Height(), scale)
	})
	// 获取设置屏幕信息
	m.chromium.SetOnGetScreenInfo(func(sender lcl.IObject, browser cef.ICefBrowser, screenInfo *cef.TCefScreenInfo, outResult *bool) {
		var scale = float64(m.bufferPanel.ScreenScale())
		var rect = &cef.TCefRect{}
		screenInfo = new(cef.TCefScreenInfo)
		rect.Width = cef.MiscFunc.DeviceToLogicalWithIntDouble(m.bufferPanel.Width(), scale)
		rect.Height = cef.MiscFunc.DeviceToLogicalWithIntDouble(m.bufferPanel.Height(), scale)
		screenInfo.DeviceScaleFactor = float32(scale)
		screenInfo.Depth = 0
		screenInfo.DepthPerComponent = 0
		screenInfo.IsMonochrome = 0
		screenInfo.Rect = *rect
		screenInfo.AvailableRect = *rect
		*outResult = true
	})
	// 获取设置屏幕点
	m.chromium.SetOnGetScreenPoint(func(sender lcl.IObject, browser cef.ICefBrowser, viewX int32, viewY int32, screenX *int32, screenY *int32, outResult *bool) {
		var scale = float64(m.bufferPanel.ScreenScale())
		var viewPoint = types.TPoint{}
		viewPoint.X = cef.MiscFunc.DeviceToLogicalWithIntDouble(viewX, scale)
		viewPoint.Y = cef.MiscFunc.DeviceToLogicalWithIntDouble(viewY, scale)
		var screenPoint = m.bufferPanel.ClientToScreenWithPoint(viewPoint)
		*outResult = true
		*screenX = screenPoint.X
		*screenY = screenPoint.Y
	})
	m.chromium.SetOnAfterCreated(func(sender lcl.IObject, browser cef.ICefBrowser) {
		html, _ := os.ReadFile("E:\\SWT\\gopath\\src\\github.com\\energye\\energy\\examples\\osr\\alienwindow\\alien.html")
		m.chromium.LoadStringWithStringFrame(string(html), browser.GetMainFrame())
	})
	m.chromium.SetOnPopupShow(func(sender lcl.IObject, browser cef.ICefBrowser, show bool) {
		if m.chromium != nil {
			m.chromium.Invalidate(cefTypes.PET_VIEW)
		}
	})
	m.chromium.SetOnPopupSize(func(sender lcl.IObject, browser cef.ICefBrowser, rect cef.TCefRect) {
		screenScale := m.bufferPanel.ScreenScale()
		cef.MiscFunc.LogicalToDeviceWithRectDouble(&rect, float64(screenScale))
	})
	// 在Paint内展示内容到窗口中
	// 使用 bitmap
	m.chromium.SetOnPaint(func(sender lcl.IObject, browser cef.ICefBrowser, type_ cefTypes.TCefPaintElementType, dirtyRectsCount cefTypes.NativeUInt, dirtyRects cef.ICefRectArray, buffer uintptr, width int32, height int32) {
		if m.bufferPanel.BeginBufferDraw() {
			m.bufferPanel.UpdateBufferDimensions(width, height)
			m.bufferPanel.BufferIsResized(false)
			tempBitMap = m.bufferPanel.Buffer()
			tempBitMap.BeginUpdate(false)
			tempWidth = m.bufferPanel.BufferWidth()
			tempHeight = m.bufferPanel.BufferHeight()
			rgbSizeOf := int(unsafe.Sizeof(types.TRGBQuad{}))
			srcStride := int(width) * rgbSizeOf
			for i := 0; i < dirtyRects.Count(); i++ {
				rect := dirtyRects.Get(i)
				if rect.X >= 0 && rect.Y >= 0 {
					tempLineSize = int(math.Min(float64(rect.Width), float64(tempWidth-rect.X))) * rgbSizeOf
					if tempLineSize > 0 {
						tempSrcOffset = int((rect.Y*width)+rect.X) * rgbSizeOf
						tempDstOffset = int(rect.X) * rgbSizeOf
						//src := @pbyte(buffer)[TempSrcOffset];
						src = uintptr(ptr.GetParamPtr(buffer, tempSrcOffset)) // 拿到src指针, 实际是 byte 指针
						j := int(math.Min(float64(rect.Height), float64(tempHeight-rect.Y)))
						for ii := 0; ii < j; ii++ {
							tempBufferBits := tempBitMap.ScanLine(rect.Y + int32(ii))
							dst = uintptr(ptr.GetParamPtr(tempBufferBits, tempDstOffset)) //拿到dst指针, 实际是 byte 指针
							rtl.Move(src, dst, tempLineSize)                              //  也可以自己实现字节复制
							src = src + uintptr(srcStride)
						}
					}
				}
			}
			tempBitMap.EndUpdate(false)

			m.bufferPanel.EndBufferDraw()
			if m.HandleAllocated() {
				m.bufferPanel.Invalidate()
			}

			size := types.TSize{Cx: tempBitMap.Width(), Cy: tempBitMap.Height()}
			srcPt := types.TPoint{}
			m.SetWidth(size.Cx)
			m.SetHeight(size.Cy)
			memDC := win.CreateCompatibleDC(0)
			win.SelectObject(memDC, tempBitMap.Handle())

			blendFunc := win.TBlendFunction{BlendOp: win.AC_SRC_OVER, BlendFlags: 0, SourceConstantAlpha: 255, AlphaFormat: win.AC_SRC_ALPHA}

			win.SetWindowLong(m.Handle(), win.GWL_EXSTYLE, uintptr(win.GetWindowLong(m.Handle(), win.GWL_EXSTYLE)|win.WS_EX_LAYERED))

			win.UpdateLayeredWindow(m.Handle(), 0, nil, &size, memDC, &srcPt, 0, &blendFunc, win.ULW_ALPHA)

			win.DeleteDC(memDC)
		}
	})
}

func (m *CommonWindow) bufferPanelEvent() {
	m.bufferPanel.SetOnClick(func(sender lcl.IObject) {
		m.bufferPanel.SetFocus()
	})
	m.bufferPanel.SetOnEnter(func(sender lcl.IObject) {
		m.chromium.SetFocus(true)
	})
	m.bufferPanel.SetOnExit(func(sender lcl.IObject) {
		m.chromium.SetFocus(false)
	})
	// panel Align 设置为 client 时， 如果调整窗口大小
	// 该函数被回调, 需要调用 WasResized 调整页面同步和主窗口一样
	m.bufferPanel.SetOnResize(func(sender lcl.IObject) {
		if m.bufferPanel.BufferIsResized(false) {
			m.chromium.Invalidate(cefTypes.PET_VIEW)
		} else {
			m.chromium.WasResized()
		}
	})
}
