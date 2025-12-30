package utils

import (
	"crypto/md5"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
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

// DownloadImageToLocal 下载网络图片到本地文件
// permanent: true表示永久保存到缓存目录，false表示保存到临时文件
func DownloadImageToLocal(imageURL string, permanent bool) (string, error) {
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

	var filePath string

	if permanent {
		// 永久保存：使用URL的MD5哈希作为文件名
		hash := md5.Sum([]byte(imageURL))
		fileName := hex.EncodeToString(hash[:]) + ext

		// 创建缓存目录
		cacheDir := "./tmp/image_cache"
		err := os.MkdirAll(cacheDir, 0755)
		if err != nil {
			return "", fmt.Errorf("创建缓存目录失败: %v", err)
		}

		filePath = filepath.Join(cacheDir, fileName)

		// 检查文件是否已存在
		if _, err := os.Stat(filePath); err == nil {
			// 文件已存在，直接返回路径
			return filePath, nil
		}
	} else {
		// 临时文件：创建临时文件
		tmpFile, err := os.CreateTemp("", "downloaded_image_*"+ext)
		if err != nil {
			return "", fmt.Errorf("创建临时文件失败: %v", err)
		}
		tmpFile.Close() // 关闭文件句柄，稍后重新打开写入
		filePath = tmpFile.Name()
	}

	// 下载图片
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Get(imageURL)
	if err != nil {
		if !permanent {
			os.Remove(filePath) // 清理临时文件
		}
		return "", fmt.Errorf("下载图片失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		if !permanent {
			os.Remove(filePath)
		}
		return "", fmt.Errorf("下载图片失败，状态码: %d", resp.StatusCode)
	}

	// 创建目标文件
	outFile, err := os.Create(filePath)
	if err != nil {
		if !permanent {
			os.Remove(filePath)
		}
		return "", fmt.Errorf("创建文件失败: %v", err)
	}
	defer outFile.Close()

	// 保存图片内容
	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		os.Remove(filePath) // 下载失败时清理文件
		return "", fmt.Errorf("保存图片失败: %v", err)
	}

	return filePath, nil
}
