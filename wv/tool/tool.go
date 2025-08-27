package tool

import (
	"bytes"
	"github.com/energye/examples/cef/utils"
	"github.com/energye/examples/cef/utils/draw"
	"image"
	"image/png"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

func DownloadFavicon(savePath, host, icoURL string, fn func(iconPath string)) {
	if _, err := url.Parse(icoURL); err == nil {
		// 下载 favicon.ico
		go func() {
			resp, err := http.Get(icoURL)
			if err == nil {
				defer resp.Body.Close()
				data, err := io.ReadAll(resp.Body)
				if err == nil {
					// png 或 ico 缩放至 16x16
					// 检测图片真实格式
					if imageFormat, err := utils.DetectImageFormatByte(data); err == nil {
						// 缩放图片
						// 把 ico 转 png
						// 把 png 缩放至 16x16
						if imageFormat == "ico" {
							icoBuf := &bytes.Buffer{}
							icoBuf.Write(data)
							// 解码ICO（自动选择最佳尺寸）
							icoImg, err := utils.Decode(icoBuf)
							if err != nil {
								println("[ERROR] OnFavIconUrlChange ICO Decode:", err.Error())
								return
							}
							pngBuf := &bytes.Buffer{}
							// 编码为PNG格式
							if err := png.Encode(pngBuf, icoImg); err != nil {
								println("[ERROR] OnFavIconUrlChange ICO To PNG:", err.Error())
								return
							}
							// 解码 png 到 image
							pngImg, err := png.Decode(pngBuf)
							if err != nil {
								println("[ERROR] OnFavIconUrlChange PNG Decode:", err.Error())
								return
							}
							pngBounds := pngImg.Bounds()
							// 存放缩放后的图像 16x16
							scaledImg := image.NewRGBA(image.Rect(0, 0, 16, 16))
							// 使用 CatmullRom 插值（比双线性更平滑）
							draw.CatmullRom.Scale(scaledImg, scaledImg.Bounds(), pngImg, pngBounds, draw.Over, nil)
							// 最后保存缩放 png
							scalePngBuf := &bytes.Buffer{}
							if err := png.Encode(scalePngBuf, scaledImg); err != nil {
								println("[ERROR] OnFavIconUrlChange PNG Encode Save Buffer:", err.Error())
								return
							}
							data = scalePngBuf.Bytes()
						} else {
							pngBuf := &bytes.Buffer{}
							pngBuf.Write(data)
							// 解码 png 到 image
							pngImg, err := png.Decode(pngBuf)
							if err != nil {
								println("[ERROR] OnFavIconUrlChange PNG Decode:", err.Error())
								return
							}
							pngBounds := pngImg.Bounds()
							// 存放缩放后的图像 16x16
							scaledImg := image.NewRGBA(image.Rect(0, 0, 16, 16))
							// 使用 CatmullRom 插值（比双线性更平滑）
							draw.CatmullRom.Scale(scaledImg, scaledImg.Bounds(), pngImg, pngBounds, draw.Over, nil)
							// 最后保存缩放 png
							scalePngBuf := &bytes.Buffer{}
							if err := png.Encode(scalePngBuf, scaledImg); err != nil {
								println("[ERROR] OnFavIconUrlChange PNG Encode Save Buffer:", err.Error())
								return
							}
							data = scalePngBuf.Bytes()
						}

						// 创建保存目录
						if err = os.MkdirAll(savePath, fs.ModePerm); err != nil {
							println("[ERROR] OnFavIconUrlChange MkdirAll:", err.Error())
						}
						// 保存图标目录
						saveIcoPath := filepath.Join(savePath, host+"_favicon.png")
						// 保存 logo
						if err = os.WriteFile(saveIcoPath, data, fs.ModePerm); err == nil {
							// callback
							fn(saveIcoPath)
						} else {
							println("[ERROR] OnFavIconUrlChange WriteFile:", err.Error())
						}
					} else {
						println("[ERROR] OnFavIconUrlChange DetectImageFormatByte:", err.Error())
					}
				}
			}
		}()
	}
}
