package main

// The scene was originally laid out in hard-coded 1440x880 coordinates. That
// meant (a) the whole content block sat low in the window and (b) a resize
// left the content clipped in the top-left corner instead of following the
// client area. Layout maps design-space rects into the current client rect:
// uniform scale + centring, so every card stays visible at any window size.

import (
	"math"

	"github.com/energye/lcl/types"
)

const (
	designW = 1440.0
	designH = 880.0
	// The original grid (header 16..82, cards 96..828) left 16px above and
	// 52px below once the header was accounted for, so all blocks read as
	// "sitting too low". Lift the whole design block to balance the margins.
	designVShift = -16.0
)

type Layout struct {
	scale   float64
	dx, dy  float64
	cw, ch  int32
	initial bool
}

// update returns true when the client size changed and all widget rects must
// be recomputed. It is called at most once per size change, never per frame.
func (l *Layout) update(cw, ch int32) bool {
	if cw < 1 || ch < 1 {
		return false
	}
	if l.initial && l.cw == cw && l.ch == ch {
		return false
	}
	s := math.Min(float64(cw)/designW, float64(ch)/designH)
	if s <= 0 {
		return false
	}
	l.scale = s
	l.dx = (float64(cw) - designW*s) / 2
	// dy never falls below designVShift*s: when height is the limiting axis
	// s=ch/designH and the centring term is zero. The header's design Top is
	// 16, so 16*s + dy >= 0 always - nothing is clipped at the top.
	l.dy = (float64(ch)-designH*s)/2 + designVShift*s
	l.cw, l.ch, l.initial = cw, ch, true
	return true
}

func (l *Layout) rect(d types.TRect) types.TRect {
	s := l.scale
	return types.TRect{
		Left:   int32(math.Round(l.dx + float64(d.Left)*s)),
		Top:    int32(math.Round(l.dy + float64(d.Top)*s)),
		Right:  int32(math.Round(l.dx + float64(d.Right)*s)),
		Bottom: int32(math.Round(l.dy + float64(d.Bottom)*s)),
	}
}
