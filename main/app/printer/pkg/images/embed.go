package images

import (
	"embed"
)

//go:embed person_in_charge.png
var ImagesFS embed.FS

// 获取图片数据
func GetImageData(imagePath string) ([]byte, error) {
	return ImagesFS.ReadFile(imagePath)
}

// 图片路径常量
const (
	PersonInChargeImg = "person_in_charge.png"
)
