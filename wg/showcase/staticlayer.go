package main

import (
	"time"

	"github.com/energye/lcl/types"
)

// The static layer is the core of the 60fps optimisation.
//
// Profiling the original frame (1440x880, 21 widgets) showed:
//   clear 1.64ms | blit 1.97ms | widgets 31.34ms | total ~35ms (30 FPS)
// of which 9 glass cards + 2 glow blobs = ~24.6ms, i.e. 79% of the frame.
// Those are per-pixel alpha blends of large RGBA bitmaps - and they never
// change. Re-blending them 60 times a second is pure waste.
//
// The fix: rasterise every static widget ONCE into an opaque Pf24bit
// full-window snapshot. Each subsequent frame is:
//   1. one opaque full-window BitBlt of the snapshot into the back buffer
//   2. re-paint only the 12 dynamic widgets (charts, gauges, text, controls)
//   3. one atomic BitBlt of the back buffer onto the visible DC
//
// The snapshot is rebuilt only when the theme, global opacity, window size or
// layer cache changes - all of which are rare, user-driven events.

// ensureStaticBuffer (re)bakes the snapshot when needed.
func (s *Scene) ensureStaticBuffer(r types.TRect) {
	w, h := r.Width(), r.Height()
	if w < 1 || h < 1 {
		return
	}
	// Resize debounce: do not bake for every intermediate WM_SIZE.
	if s.holdStatic && s.staticBuf != nil {
		return
	}
	if s.staticBuf != nil && s.staticW == w && s.staticH == h && !s.staticDirty {
		return
	}
	if s.staticBuf != nil {
		s.staticBuf.Free()
	}
	b := newBackBuffer(w, h)
	s.staticBuf, s.staticW, s.staticH = b, w, h

	t0 := time.Now()
	prev := s.painter.Canvas
	prevFrame := s.painter.frame
	// A re-bake can happen INSIDE OnPaint (slider changes opacity, or a
	// WM_SIZE changes the client rect). painter.frame still points at the
	// previous frame's AA surface there, and that surface may already have
	// been freed by the resize. The static snapshot is a GDI bitmap and MUST
	// be baked with the GDI backend: with frame != nil, FillRect would write
	// into a stale pixel block (segfault), and Text would disappear into the
	// deferred textCmds queue, which Reset() then discards (the "9th card
	// goes blank" bug).
	s.painter.frame = nil
	s.painter.Canvas = b.Canvas()
	s.painter.Reset()
	s.painter.FillRect(r, s.theme.Bg)
	for _, wgt := range s.staticWidgets {
		wgt.Paint(s.painter, s)
	}
	s.painter.Canvas = prev
	s.painter.frame = prevFrame
	s.painter.Reset() // canvas changed: invalidate all cached brush/pen/font state
	s.staticBakeUs = time.Since(t0).Microseconds()
	s.staticDirty = false
	s.staticGen++ // AA backend must re-read the snapshot into Go memory
	s.forceFull = true
}

// beginFrame replaces the old full-screen FillRect with the baked snapshot.
func (s *Scene) beginFrame(r types.TRect) {
	if s.useStatic {
		s.ensureStaticBuffer(r)
		// A stale snapshot from before a resize has different dimensions:
		// blitting it (or a dirty sub-rect of it) would read outside the DIB.
		if s.staticBuf != nil && s.staticW == r.Width() && s.staticH == r.Height() {
			// Pf24bit -> Pf24bit: an opaque BitBlt, no alpha blending.
			s.painter.Canvas.DrawWithIntX2Graphic(r.Left, r.Top, s.staticBuf)
			s.painter.Calls++
			return
		}
	}
	s.painter.FillRect(r, s.theme.Bg)
}

// InvalidateStatic forces a re-bake on the next frame.
func (s *Scene) InvalidateStatic() {
	s.staticDirty = true
	s.forceFull = true
	s.RequestFrame()
}

func (s *Scene) freeStaticBuffer() {
	if s.staticBuf != nil {
		s.staticBuf.Free()
		s.staticBuf = nil
	}
}

// staticInfo feeds the perf panel.
func (s *Scene) staticInfo() (on bool, bakeUs int64, dynamic int) {
	return s.useStatic, s.staticBakeUs, len(s.dynamicWidgets)
}
