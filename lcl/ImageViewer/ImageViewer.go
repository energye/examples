package main

import (
	"github.com/energye/lcl/lcl"
	"github.com/energye/lcl/types"
)

type TImageViewerForm struct {
	lcl.TEngForm
	Image      lcl.IImage
	OpenButton lcl.IButton
	ZoomInBtn  lcl.IButton
	ZoomOutBtn lcl.IButton
	FitBtn     lcl.IButton
	Scale      float64
}

var ImageViewerForm TImageViewerForm

func main() {
	lcl.Init(nil, nil)
	lcl.Application.Initialize()
	lcl.Application.SetMainFormOnTaskBar(true)
	lcl.Application.NewForms(&ImageViewerForm)
	lcl.Application.Run()
}

func (i *TImageViewerForm) FormCreate(sender lcl.IObject) {
	i.SetCaption("图片查看器")
	i.SetPosition(types.PoScreenCenter)
	i.SetWidth(900)
	i.SetHeight(700)

	panel := lcl.NewPanel(i)
	panel.SetParent(i)
	panel.SetLeft(0)
	panel.SetTop(0)
	panel.SetWidth(i.Width())
	panel.SetHeight(40)
	panel.SetAlign(types.AlTop)
	panel.SetBevelOuter(types.BvNone)

	i.OpenButton = lcl.NewButton(panel)
	i.OpenButton.SetParent(panel)
	i.OpenButton.SetCaption("打开图片")
	i.OpenButton.SetLeft(10)
	i.OpenButton.SetTop(5)
	i.OpenButton.SetWidth(100)
	i.OpenButton.SetHeight(30)
	i.OpenButton.SetOnClick(i.OnOpenClick)

	i.ZoomInBtn = lcl.NewButton(panel)
	i.ZoomInBtn.SetParent(panel)
	i.ZoomInBtn.SetCaption("放大 (+)")
	i.ZoomInBtn.SetLeft(120)
	i.ZoomInBtn.SetTop(5)
	i.ZoomInBtn.SetWidth(80)
	i.ZoomInBtn.SetHeight(30)
	i.ZoomInBtn.SetOnClick(i.OnZoomInClick)

	i.ZoomOutBtn = lcl.NewButton(panel)
	i.ZoomOutBtn.SetParent(panel)
	i.ZoomOutBtn.SetCaption("缩小 (-)")
	i.ZoomOutBtn.SetLeft(210)
	i.ZoomOutBtn.SetTop(5)
	i.ZoomOutBtn.SetWidth(80)
	i.ZoomOutBtn.SetHeight(30)
	i.ZoomOutBtn.SetOnClick(i.OnZoomOutClick)

	i.FitBtn = lcl.NewButton(panel)
	i.FitBtn.SetParent(panel)
	i.FitBtn.SetCaption("适应窗口")
	i.FitBtn.SetLeft(300)
	i.FitBtn.SetTop(5)
	i.FitBtn.SetWidth(80)
	i.FitBtn.SetHeight(30)
	i.FitBtn.SetOnClick(i.OnFitClick)

	i.Image = lcl.NewImage(i)
	i.Image.SetParent(i)
	i.Image.SetLeft(0)
	i.Image.SetTop(40)
	i.Image.SetWidth(i.Width())
	i.Image.SetHeight(i.Height() - 40)
	i.Image.SetAlign(types.AlClient)
	i.Image.SetStretch(false)
	i.Image.SetCenter(true)

	i.Scale = 1.0
}

func (i *TImageViewerForm) OnOpenClick(sender lcl.IObject) {
	openDialog := lcl.NewOpenDialog(i)
	openDialog.SetFilter("图片文件 (*.png;*.jpg;*.jpeg;*.bmp;*.gif)|*.png;*.jpg;*.jpeg;*.bmp;*.gif|所有文件 (*.*)|*.*")

	if openDialog.Execute() {
		picture := lcl.NewPicture()
		picture.LoadFromFile(openDialog.FileName())
		i.Image.Picture().Assign(picture)
		i.SetCaption("图片查看器 - " + openDialog.FileName())
		i.Scale = 1.0
		picture.Free()
	}
}

func (i *TImageViewerForm) OnZoomInClick(sender lcl.IObject) {
	i.Scale *= 1.2
	i.ApplyZoom()
}

func (i *TImageViewerForm) OnZoomOutClick(sender lcl.IObject) {
	i.Scale /= 1.2
	i.ApplyZoom()
}

func (i *TImageViewerForm) OnFitClick(sender lcl.IObject) {
	i.Image.SetStretch(true)
	i.Scale = 1.0
}

func (i *TImageViewerForm) ApplyZoom() {
	i.Image.SetStretch(false)
	width := int32(float64(i.Image.Picture().Width()) * i.Scale)
	height := int32(float64(i.Image.Picture().Height()) * i.Scale)
	i.Image.SetWidth(width)
	i.Image.SetHeight(height)
}
