package utils

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
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
