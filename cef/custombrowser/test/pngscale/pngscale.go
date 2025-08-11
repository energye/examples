package main

import (
	"github.com/energye/examples/cef/utils/draw"
	"image"
	"image/png"
	"os"
)

func main() {
	// 1. 读取原始图片
	file, err := os.Open("E:\\SWT\\gopath\\src\\github.com\\energye\\workspace\\examples\\cef\\custombrowser\\test\\xxx.png")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	img, err := png.Decode(file)
	if err != nil {
		panic(err)
	}

	// 2. 计算目标尺寸（例如缩小到 50%）
	originalBounds := img.Bounds()
	newWidth := 16
	newHeight := 16

	// 3. 创建目标图像（RGBA 格式）
	scaledImg := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	// 4. 使用 CatmullRom 插值（比双线性更平滑）
	draw.CatmullRom.Scale(scaledImg, scaledImg.Bounds(), img, originalBounds, draw.Over, nil)

	// 5. 保存结果
	outFile, err := os.Create("E:\\SWT\\gopath\\src\\github.com\\energye\\workspace\\examples\\cef\\custombrowser\\test\\output.png")
	if err != nil {
		panic(err)
	}
	defer outFile.Close()

	png.Encode(outFile, scaledImg)
}
