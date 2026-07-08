package overlay

import "github.com/energye/examples/lcl/gpui/core/math"

// PlacementOptions controls popup positioning.
type PlacementOptions struct {
	Offset math.Vec2
	Flip   bool
	Clamp  bool
}

// Place computes popup bounds inside a viewport.
func Place(anchor math.Rect, popupSize math.Vec2, viewport math.Rect, placement Placement, opts PlacementOptions) math.Rect {
	rect := placeRaw(anchor, popupSize, placement, opts.Offset)
	if opts.Flip && !fits(rect, viewport) {
		flipped := placeRaw(anchor, popupSize, flipPlacement(placement), opts.Offset)
		if fits(flipped, viewport) || visibleArea(flipped, viewport) > visibleArea(rect, viewport) {
			rect = flipped
		}
	}
	if opts.Clamp {
		rect = clampToViewport(rect, viewport)
	}
	return rect
}

func placeRaw(anchor math.Rect, popupSize math.Vec2, placement Placement, offset math.Vec2) math.Rect {
	switch placement {
	case Top:
		return math.NewRect(anchor.X+(anchor.W-popupSize.X)/2+offset.X, anchor.Y-popupSize.Y-offset.Y, popupSize.X, popupSize.Y)
	case Bottom:
		return math.NewRect(anchor.X+(anchor.W-popupSize.X)/2+offset.X, anchor.Y+anchor.H+offset.Y, popupSize.X, popupSize.Y)
	case Left:
		return math.NewRect(anchor.X-popupSize.X-offset.X, anchor.Y+(anchor.H-popupSize.Y)/2+offset.Y, popupSize.X, popupSize.Y)
	case Right:
		return math.NewRect(anchor.X+anchor.W+offset.X, anchor.Y+(anchor.H-popupSize.Y)/2+offset.Y, popupSize.X, popupSize.Y)
	case BottomRight:
		return math.NewRect(anchor.X+anchor.W-popupSize.X+offset.X, anchor.Y+anchor.H+offset.Y, popupSize.X, popupSize.Y)
	case TopLeft:
		return math.NewRect(anchor.X+offset.X, anchor.Y-popupSize.Y-offset.Y, popupSize.X, popupSize.Y)
	case TopRight:
		return math.NewRect(anchor.X+anchor.W-popupSize.X+offset.X, anchor.Y-popupSize.Y-offset.Y, popupSize.X, popupSize.Y)
	case LeftTop:
		return math.NewRect(anchor.X-popupSize.X-offset.X, anchor.Y+offset.Y, popupSize.X, popupSize.Y)
	case LeftBottom:
		return math.NewRect(anchor.X-popupSize.X-offset.X, anchor.Y+anchor.H-popupSize.Y+offset.Y, popupSize.X, popupSize.Y)
	case RightTop:
		return math.NewRect(anchor.X+anchor.W+offset.X, anchor.Y+offset.Y, popupSize.X, popupSize.Y)
	case RightBottom:
		return math.NewRect(anchor.X+anchor.W+offset.X, anchor.Y+anchor.H-popupSize.Y+offset.Y, popupSize.X, popupSize.Y)
	case Center:
		return math.NewRect(anchor.X+(anchor.W-popupSize.X)/2+offset.X, anchor.Y+(anchor.H-popupSize.Y)/2+offset.Y, popupSize.X, popupSize.Y)
	default:
		return math.NewRect(anchor.X+offset.X, anchor.Y+anchor.H+offset.Y, popupSize.X, popupSize.Y)
	}
}

func flipPlacement(placement Placement) Placement {
	switch placement {
	case Top:
		return Bottom
	case Bottom:
		return Top
	case BottomLeft:
		return TopLeft
	case BottomRight:
		return TopRight
	case TopLeft:
		return BottomLeft
	case TopRight:
		return BottomRight
	case LeftTop:
		return RightTop
	case RightTop:
		return LeftTop
	case Left:
		return Right
	case Right:
		return Left
	case LeftBottom:
		return RightBottom
	case RightBottom:
		return LeftBottom
	default:
		return placement
	}
}

func fits(rect, viewport math.Rect) bool {
	return rect.X >= viewport.X &&
		rect.Y >= viewport.Y &&
		rect.X+rect.W <= viewport.X+viewport.W &&
		rect.Y+rect.H <= viewport.Y+viewport.H
}

func visibleArea(rect, viewport math.Rect) float32 {
	intersection := rect.Intersect(viewport)
	if intersection.W <= 0 || intersection.H <= 0 {
		return 0
	}
	return intersection.W * intersection.H
}

func clampToViewport(rect, viewport math.Rect) math.Rect {
	if rect.W > viewport.W {
		rect.X = viewport.X
	} else if rect.X < viewport.X {
		rect.X = viewport.X
	} else if rect.X+rect.W > viewport.X+viewport.W {
		rect.X = viewport.X + viewport.W - rect.W
	}

	if rect.H > viewport.H {
		rect.Y = viewport.Y
	} else if rect.Y < viewport.Y {
		rect.Y = viewport.Y
	} else if rect.Y+rect.H > viewport.Y+viewport.H {
		rect.Y = viewport.Y + viewport.H - rect.H
	}
	return rect
}
