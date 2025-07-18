package form

import (
	"fmt"
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
)

type TPlayListItem struct {
	Caption  string // 歌曲名
	Singer   string // 歌手
	Length   int32  // 歌曲长度
	Lyric    string // 歌词文件
	FileName string // 歌曲文件名
}

type TPlayControl struct {
	lcl.IDrawGrid
	datas          []TPlayListItem
	focusedColor   types.TColor
	playColor      types.TColor
	mouseMoveColor types.TColor
	mouseMoveIndex int32
	playingIndex   int32
	singerPicR     types.TRect
	singerPic      *lcl.TBitmap
}

func NewPlayControl(owner lcl.IComponent) *TPlayControl {
	m := new(TPlayControl)
	m.IDrawGrid = lcl.NewDrawGrid(owner)
	m.IDrawGrid.SetDefaultDrawing(false)
	m.IDrawGrid.SetDefaultRowHeight(24)
	m.IDrawGrid.SetOptions(types.NewSet(types.GoRangeSelect, types.GoRowSelect))
	m.IDrawGrid.SetRowCount(1)
	m.IDrawGrid.SetColCount(4)
	m.IDrawGrid.SetFixedRows(0)
	m.IDrawGrid.SetFixedCols(0)
	m.IDrawGrid.SetGridLineWidth(0)
	m.IDrawGrid.SetBorderStyleToBorderStyle(types.BsNone)
	m.IDrawGrid.SetScrollBars(types.SsVertical)
	m.IDrawGrid.SetWidth(536)
	m.IDrawGrid.SetHeight(397)
	// 加载时取消第一行永远被选中
	m.IDrawGrid.SetSelection(types.TGridRect{Left: 24, Top: 24, Right: 24, Bottom: 24})
	m.IDrawGrid.SetColWidths(0, int32(float32(m.Width())*0.1))
	m.IDrawGrid.SetColWidths(1, int32(float32(m.Width())*0.4))
	m.IDrawGrid.SetColWidths(2, int32(float32(m.Width())*0.2))
	m.IDrawGrid.SetColWidths(3, int32(float32(m.Width())*0.2))
	m.IDrawGrid.SetColor(0x00EDEEF9)
	m.IDrawGrid.SetDoubleBuffered(true)
	m.focusedColor = 0x00C8CBEB
	m.playColor = m.focusedColor + 12
	m.mouseMoveColor = m.focusedColor - 12
	m.mouseMoveIndex = -1
	m.playingIndex = -1
	m.IDrawGrid.SetOnDblClick(m.onDblClick)
	m.IDrawGrid.SetOnMouseMove(m.onMouseMove)
	//m.IDrawGrid.SetOnMouseDown(m.onMouseDown)
	m.IDrawGrid.SetOnDrawCell(m.onDrawCell)
	//m.TDrawGrid.SetOnMouseEnter(m.onMouseEnter)
	m.IDrawGrid.SetOnMouseLeave(m.onMouseLeave)

	return m
}

func (p *TPlayControl) Add(item TPlayListItem) int32 {
	p.datas = append(p.datas, item)
	p.SetRowCount(int32(len(p.datas)))
	return int32(len(p.datas)) - 1
}

func (p *TPlayControl) onDrawCell(sender lcl.IObject, aCol, aRow int32, rect types.TRect, state types.TGridDrawState) {
	if len(p.datas) > 0 {
		canvas := p.Canvas()
		brush := canvas.BrushToBrush()
		font := canvas.FontToFont()
		if aRow < int32(len(p.datas)) {
			//drawFlags := types.NewSet(types.TfVerticalCenter, types.TfSingleLine, types.TfEndEllipsis)
			item := p.datas[int(aRow)]
			if p.mouseMoveIndex == aRow && p.playingIndex != aRow && !state.In(types.GdFocused) && !state.In(types.GdSelected) {
				brush.SetColor(p.focusedColor - 12)
			} else if state.In(types.GdFocused) || state.In(types.GdSelected) {
				brush.SetColor(p.focusedColor)
			} else {
				brush.SetColor(p.Color())
			}

			if p.playingIndex == aRow {
				brush.SetColor(p.focusedColor + 12)
				p.SetRowHeights(aRow, 50)
				if p.RowHeights(aRow) < 50 {
					p.ScrollBy(0, aRow+2)
				}
			} else {
				p.SetRowHeights(aRow, 24)
			}
			canvas.FillRectWithRect(rect)
			r := p.CellRect(aCol, aRow)
			switch aCol {
			case 0:
				if aRow == p.playingIndex {
					if !p.singerPicR.IsEmpty() {
						r.Left += 1
						r.Top += +1
						r.Bottom -= -1
						//canvas.StretchDraw(r, FSingerPic);
					}
				} else {
					r.Inflate(-10, 0)
					s := fmt.Sprintf("%d.", aRow+1)
					canvas.TextRectWithRectIntX2StringTextStyle(r, r.Left, r.Top, s, lcl.TTextStyle{})
				}

			case 1:
				if aRow == p.playingIndex {
					r.Inflate(-10, 0)
					canvas.FontToFont().SetSize(12)
					font.SetStyle(types.NewSet(types.FsBold))
					canvas.TextRectWithRectIntX2StringTextStyle(r, r.Left, r.Top, item.Caption, lcl.TTextStyle{})
				} else {
					r.Inflate(-5, 0)
					canvas.TextRectWithRectIntX2StringTextStyle(r, r.Left, r.Top, item.Caption, lcl.TTextStyle{})
				}
				canvas.FontToFont().SetSize(9)
				font.SetStyle(0)
			case 2:
				r.Inflate(-5, 0)
				canvas.TextRectWithRectIntX2StringTextStyle(r, r.Left, r.Top, item.Singer, lcl.TTextStyle{})
			case 3:
				r.Inflate(-5, 0)
				canvas.TextRectWithRectIntX2StringTextStyle(r, r.Left, r.Top, p.mediaLengthToTimeStr(item.Length), lcl.TTextStyle{})
			}
		}

	} else {
		p.Canvas().BrushToBrush().SetColor(p.Color())
		p.Canvas().FillRectWithRect(rect)
	}
}

func (p *TPlayControl) onMouseMove(sender lcl.IObject, shift types.TShiftState, x, y int32) {
	if !p.Enabled() {
		return
	}
	var col, row int32
	p.MouseToCellWithIntX4(x, y, &col, &row)
	p.mouseMoveIndex = row
	if p.mouseMoveIndex == -1 {
		return
	}
	p.Invalidate()
}

func (p *TPlayControl) onMouseDown(sender lcl.IObject, button types.TMouseButton, shift types.TShiftState, x, y int32) {

}

func (p *TPlayControl) onDblClick(sender lcl.IObject) {
	if !p.Enabled() {
		return
	}
	row := p.Row()
	if row == -1 {
		return
	}
	p.playingIndex = row
	p.Invalidate()
}

func (p *TPlayControl) onMouseEnter(sender lcl.IObject) {

}

func (p *TPlayControl) onMouseLeave(sender lcl.IObject) {
	if !p.Enabled() {
		return
	}
	p.mouseMoveIndex = -1
	p.Invalidate()
}

func (p *TPlayControl) mediaLengthToTimeStr(alen int32) string {
	return fmt.Sprintf("%.2d:%.2d", int(float32(alen)/1000.0)/60, int(float32(alen)/1000.0)%60)
}
