package utils

import (
	"bytes"
	"errors"
	"os"
)

// 图片格式的魔数签名
var magicTable = map[string]string{
	"\xff\xd8\xff":      "jpeg",
	"\x89PNG\r\n\x1a\n": "png",
	"GIF87a":            "gif",
	"GIF89a":            "gif",
	"BM":                "bmp",
	"\x00\x00\x01\x00":  "ico",
}

// DetectImageFormat 检测图片真实格式
func DetectImageFormat(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// 读取文件前16字节（足够覆盖常见格式）
	buffer := make([]byte, 16)
	if _, err := file.Read(buffer); err != nil {
		return "", err
	}

	return DetectImageFormatByte(buffer)
}

// DetectImageFormatByte 检测图片真实格式
func DetectImageFormatByte(imageData []byte) (string, error) {
	if len(imageData) < 16 {
		return "", errors.New("图片太小")
	}
	// 读取文件前16字节（足够覆盖常见格式）
	buffer := imageData[:16]

	// 特殊处理ICO中嵌套PNG的情况
	if bytes.HasPrefix(buffer, []byte("\x00\x00\x01\x00")) {
		// 如果ICO文件内嵌PNG，从偏移量6开始检测
		if bytes.HasPrefix(buffer[6:], []byte("\x89PNG")) {
			return "png", nil
		}
		return "ico", nil
	}

	// 检查其他格式
	for magic, format := range magicTable {
		if bytes.HasPrefix(buffer, []byte(magic)) {
			return format, nil
		}
	}

	return "", errors.New("不支持的格式")
}
