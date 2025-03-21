// Package image 提供图像处理相关的辅助函数
package image

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/disintegration/imaging"
)

// ImageHelp 提供图像处理的辅助函数
type ImageHelp struct{}

// WhiteBackgroundWithBlackText 确保图像是白底黑字
// 处理图像使其成为白底黑字，适合打印
// imgPath: 输入图像路径或URL
// imgSavePath: 输出图像保存路径
func (h *ImageHelp) WhiteBackgroundWithBlackText(imgPath, imgSavePath string) error {
	var src image.Image
	var err error

	// 检查是否是URL
	if strings.HasPrefix(imgPath, "http://") || strings.HasPrefix(imgPath, "https://") {
		// 从URL下载图像
		src, err = h.DownloadImage(imgPath)
	} else {
		// 读取本地图像
		src, err = imaging.Open(imgPath)
	}

	if err != nil {
		fmt.Println("读取图像失败:", err, imgPath)
		return err
	}

	// 处理透明背景
	bounds := src.Bounds()
	dst := image.NewRGBA(bounds)

	// 创建白色背景
	draw.Draw(dst, bounds, image.NewUniform(color.White), image.Point{}, draw.Src)

	// 将原图合并到白色背景上
	draw.Draw(dst, bounds, src, image.Point{}, draw.Over)

	// 处理彩色图像，保留重要的颜色信息
	enhancedImg := imaging.AdjustContrast(dst, 20) // 增加20%的对比度

	// 使用高斯模糊去除噪点
	blurredImg := imaging.Blur(enhancedImg, 0.1)

	// 创建一个新的灰度图像用于最终输出
	binaryImg := image.NewGray(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := blurredImg.At(x, y)
			r, g, b, _ := c.RGBA()

			// 检测黄色（r高，g高，b低）
			isYellow := (r > 20000 && g > 20000 && b < 35000)

			// 计算亮度
			brightness := (r + g + b) / 3

			// 如果是黄色或亮度较低，则设为黑色（保留文字）
			if isYellow || brightness < 32768 {
				binaryImg.SetGray(x, y, color.Gray{Y: 0}) // 黑色
			} else {
				binaryImg.SetGray(x, y, color.Gray{Y: 255}) // 白色
			}
		}
	}

	// 确保是白底黑字（计算黑白像素比例）
	blackCount := 0
	totalPixels := bounds.Dx() * bounds.Dy()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if binaryImg.GrayAt(x, y).Y == 0 {
				blackCount++
			}
		}
	}

	// 如果黑色像素超过50%，则反转图像（确保是白底黑字）
	finalImg := binaryImg
	if blackCount > totalPixels/2 {
		finalImg = image.NewGray(bounds)
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				v := binaryImg.GrayAt(x, y).Y
				finalImg.SetGray(x, y, color.Gray{Y: 255 - v})
			}
		}
	}

	// 确保目录存在
	dir := filepath.Dir(imgSavePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0777)
		if err != nil {
			return err
		}
	}

	// 保存图像
	outFile, err := os.Create(imgSavePath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	// 以PNG格式保存
	return png.Encode(outFile, finalImg)
}

// NewImageHelp 创建一个新的ImageHelp实例
func NewImageHelp() *ImageHelp {
	return &ImageHelp{}
}

// RemoveImageDomain 移除图像URL中的域名部分
func (h *ImageHelp) RemoveImageDomain(imageURL string) string {
	// 尝试解析URL
	parsedURL, err := url.Parse(imageURL)
	if err != nil {
		// 如果解析失败，使用简单的字符串处理方法
		return h.simpleRemoveDomain(imageURL)
	}

	// 返回路径部分
	return parsedURL.Path
}

// simpleRemoveDomain 使用简单的字符串处理方法移除域名
func (h *ImageHelp) simpleRemoveDomain(imageURL string) string {
	// 移除协议部分
	for _, prefix := range []string{"http://", "https://"} {
		if strings.HasPrefix(imageURL, prefix) {
			imageURL = imageURL[len(prefix):]
			break
		}
	}

	// 找到第一个斜杠位置
	slashIndex := strings.Index(imageURL, "/")
	if slashIndex >= 0 {
		return imageURL[slashIndex:]
	}

	return imageURL
}

// downloadImage 从URL下载图像
func (h *ImageHelp) DownloadImage(imageURL string) (image.Image, error) {
	// 创建HTTP客户端，设置超时
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// 发送GET请求
	resp, err := client.Get(imageURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("下载图像失败，HTTP状态码：" + resp.Status)
	}

	// 解码图像
	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return nil, err
	}

	return img, nil
}

// DownloadFile 从URL下载文件到本地
func (h *ImageHelp) DownloadFile(url string, destPath string) error {
	// 创建HTTP客户端，设置超时
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// 发送GET请求
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return errors.New("下载文件失败，HTTP状态码：" + resp.Status)
	}

	// 确保目录存在
	dir := filepath.Dir(destPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0777)
		if err != nil {
			return err
		}
	}

	// 创建文件
	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	// 写入文件
	_, err = io.Copy(out, resp.Body)
	return err
}
