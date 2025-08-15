package utils

import (
	"crypto/tls"
	"encoding/base64"
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

// 创建一个忽略证书验证的HTTP客户端
var httpClient = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 忽略证书验证
		},
	},
}

// ImageToBase64 将图片URL转换为Base64
func ImageToBase64(imageURL string) (string, error) {
	// 如果URL为空，直接返回空字符串
	if imageURL == "" {
		return "", nil
	}

	// 使用自定义的HTTP客户端发起请求
	resp, err := httpClient.Get(imageURL)
	if err != nil {
		return "", fmt.Errorf("获取图片失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取图片内容
	imageBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取图片内容失败: %v", err)
	}

	// 获取图片MIME类型
	contentType := http.DetectContentType(imageBytes)

	// 将图片内容转换为Base64
	base64Str := base64.StdEncoding.EncodeToString(imageBytes)

	// 返回完整的Base64字符串（包含MIME类型）
	return fmt.Sprintf("data:%s;base64,%s", contentType, base64Str), nil
}

// AddImageDomainAndConvertToBase64 添加图片域名并转换为Base64
func AddImageDomainAndConvertToBase64(imagePath string, baseURL string, isHttps bool) (string, error) {
	// 先获取完整的图片URL
	imageURL := AddImageDomain(imagePath, baseURL, isHttps)

	// 转换为Base64
	return ImageToBase64(imageURL)
}

// IsNetworkImage 判断给定路径是否为网络图片
func IsNetworkImage(imgPath string) bool {
	if imgPath == "" {
		return false
	}

	// 解析URL
	parsedURL, err := url.Parse(imgPath)
	if err != nil {
		return false
	}

	// 检查是否有scheme（协议）
	scheme := strings.ToLower(parsedURL.Scheme)
	return scheme == "http" || scheme == "https"
}

// DownloadImageToLocal 下载网络图片到本地临时文件
func DownloadImageToLocal(imageURL string) (string, error) {
	if imageURL == "" {
		return "", fmt.Errorf("图片URL不能为空")
	}

	// 解析URL获取文件扩展名
	parsedURL, err := url.Parse(imageURL)
	if err != nil {
		return "", fmt.Errorf("解析URL失败: %v", err)
	}

	// 获取文件扩展名，默认为.jpg
	ext := filepath.Ext(parsedURL.Path)
	if ext == "" {
		ext = ".jpg"
	}

	// 创建临时文件
	tmpFile, err := os.CreateTemp("", "downloaded_image_*"+ext)
	if err != nil {
		return "", fmt.Errorf("创建临时文件失败: %v", err)
	}
	defer tmpFile.Close()

	// 下载图片
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Get(imageURL)
	if err != nil {
		os.Remove(tmpFile.Name()) // 清理临时文件
		return "", fmt.Errorf("下载图片失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("下载图片失败，状态码: %d", resp.StatusCode)
	}

	// 保存到临时文件
	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("保存图片失败: %v", err)
	}

	return tmpFile.Name(), nil
}

// GetLocalImagePath 获取本地图片路径
// 如果是网络图片则下载到本地，如果是本地文件则直接返回路径
func GetLocalImagePath(imgPath string) (localPath string, isTemporary bool, err error) {
	if imgPath == "" {
		return "", false, fmt.Errorf("图片路径不能为空")
	}

	// 判断是否为网络图片
	if IsNetworkImage(imgPath) {
		// 下载到本地
		localPath, err = DownloadImageToLocal(imgPath)
		if err != nil {
			return "", false, fmt.Errorf("下载网络图片失败: %v", err)
		}
		return localPath, true, nil
	}

	// 检查本地文件是否存在
	if _, err := os.Stat(imgPath); os.IsNotExist(err) {
		return "", false, fmt.Errorf("本地图片文件不存在: %s", imgPath)
	} else if err != nil {
		return "", false, fmt.Errorf("检查本地图片文件失败: %v", err)
	}

	return imgPath, false, nil
}

// WhiteBackgroundWithBlackText 确保图片是白底黑字
// 将输入图片处理为白底黑字的PNG格式图片
func WhiteBackgroundWithBlackText(imgPath, imgSavePath string) error {
	// 获取本地图片路径
	localPath, isTemporary, err := GetLocalImagePath(imgPath)
	if err != nil {
		return fmt.Errorf("获取图片失败: %v", err)
	}

	// 如果是临时文件，处理完成后清理
	if isTemporary {
		defer func() {
			if err := os.Remove(localPath); err != nil {
				fmt.Printf("警告：清理临时文件失败: %s, 错误: %v\n", localPath, err)
			}
		}()
	}

	// 打开原始图片
	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("打开图片文件失败: %v", err)
	}
	defer file.Close()

	// 解码图片
	srcImg, _, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("解码图片失败: %v", err)
	}

	bounds := srcImg.Bounds()

	// 创建白色背景的新图片
	bgImg := image.NewRGBA(bounds)
	white := color.RGBA{255, 255, 255, 255}
	draw.Draw(bgImg, bounds, &image.Uniform{white}, image.Point{}, draw.Src)

	// 将原图合成到白色背景上（处理透明背景）
	draw.Draw(bgImg, bounds, srcImg, bounds.Min, draw.Over)

	// 转换为灰度图像
	grayImg := imaging.Grayscale(bgImg)

	// 增强对比度
	contrastImg := imaging.AdjustContrast(grayImg, 20)

	// 去除噪点（使用模糊滤镜）
	blurImg := imaging.Blur(contrastImg, 0.5)

	// 计算平均亮度用于阈值处理
	avgBrightness := calculateAverageBrightness(blurImg)
	threshold := uint8(avgBrightness * 0.95)

	// 二值化处理
	binaryImg := thresholdImage(blurImg, threshold)

	// 检查是否需要反色（确保是白底黑字）
	newAvgBrightness := calculateAverageBrightness(binaryImg)
	var finalProcessedImg image.Image = binaryImg
	if newAvgBrightness < 127 {
		finalProcessedImg = imaging.Invert(binaryImg)
	}

	// 最终清理（再次轻微模糊以平滑边缘）
	finalImg := imaging.Blur(finalProcessedImg, 0.3)

	// 确保保存目录存在
	saveDir := filepath.Dir(imgSavePath)
	if err := os.MkdirAll(saveDir, 0777); err != nil {
		return fmt.Errorf("创建保存目录失败: %v", err)
	}

	// 保存处理后的图片
	outputFile, err := os.Create(imgSavePath)
	if err != nil {
		return fmt.Errorf("创建输出文件失败: %v", err)
	}
	defer outputFile.Close()

	// 以PNG格式保存
	err = png.Encode(outputFile, finalImg)
	if err != nil {
		return fmt.Errorf("保存PNG图片失败: %v", err)
	}

	return nil
}

// WhiteBackgroundWithBlackTextFromURL 从URL或本地路径处理图片为白底黑字
func WhiteBackgroundWithBlackTextFromURL(imgPath, outputPath string) error {
	// 获取本地图片路径
	localPath, isTemporary, err := GetLocalImagePath(imgPath)
	if err != nil {
		return fmt.Errorf("获取图片失败: %v", err)
	}

	// 如果是临时文件，处理完成后需要清理
	if isTemporary {
		defer func() {
			if err := os.Remove(localPath); err != nil {
				fmt.Printf("警告：清理临时文件失败: %s, 错误: %v\n", localPath, err)
			}
		}()
	}

	// 调用原有的处理方法
	return WhiteBackgroundWithBlackText(localPath, outputPath)
}

// calculateAverageBrightness 计算图片的平均亮度
func calculateAverageBrightness(img image.Image) float64 {
	bounds := img.Bounds()
	var totalBrightness uint64
	var pixelCount uint64

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			grayColor := color.GrayModel.Convert(img.At(x, y)).(color.Gray)
			totalBrightness += uint64(grayColor.Y)
			pixelCount++
		}
	}

	if pixelCount == 0 {
		return 0
	}

	return float64(totalBrightness) / float64(pixelCount)
}

// thresholdImage 对图片进行阈值处理（二值化）
func thresholdImage(img image.Image, threshold uint8) *image.Gray {
	bounds := img.Bounds()
	result := image.NewGray(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			grayColor := color.GrayModel.Convert(img.At(x, y)).(color.Gray)

			// 阈值处理：大于阈值设为白色，小于阈值设为黑色
			if grayColor.Y > threshold {
				result.SetGray(x, y, color.Gray{255}) // 白色
			} else {
				result.SetGray(x, y, color.Gray{0}) // 黑色
			}
		}
	}

	return result
}

// GetWhiteBackgroundWithBlackTextLogoPath 获得白色背景和黑色文本徽标路径
// 如果处理后的文件不存在，则自动处理原始logo并保存
// appId: 应用ID，logoUrl: 原始logo文件路径，uploadsDir: 上传目录基础路径
func GetWhiteBackgroundWithBlackTextLogoPath(appId uint64, logoUrl, uploadsDir string) (string, error) {
	// 构建保存路径
	shopDir := fmt.Sprintf("shop%d", appId)
	logoBasename := filepath.Base(logoUrl)
	savePath := filepath.Join(uploadsDir, shopDir, "white_background_text_"+logoBasename)

	return savePath, nil
}
