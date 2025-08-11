package main

import (
	"github.com/energye/examples/cef/utils"
	"image/png"
	"os"
)

func main() {
	// 1. 打开ICO文件
	icoFile, err := os.Open("E:\\SWT\\gopath\\src\\github.com\\energye\\workspace\\examples\\EnergyCache\\SiteResource\\www.bilibili.com_favicon.ico")
	if err != nil {
		panic(err)
	}
	defer icoFile.Close()

	//buf := &bytes.Buffer{}
	// 2. 解码ICO（自动选择最佳尺寸）
	img, err := utils.Decode(icoFile)
	if err != nil {
		panic(err)
	}

	// 3. 创建PNG输出文件
	pngFile, err := os.Create("E:\\SWT\\gopath\\src\\github.com\\energye\\workspace\\examples\\cef\\custombrowser\\test\\xxx.png")
	if err != nil {
		panic(err)
	}
	defer pngFile.Close()
	//buf := &bytes.Buffer{}

	// 4. 编码为PNG格式
	if err := png.Encode(pngFile, img); err != nil {
		panic(err)
	}
}
