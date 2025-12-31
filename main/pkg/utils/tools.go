package utils

import (
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"ttpos-server-go/config"

	"github.com/shopspring/decimal"
)

type VersionInfo struct {
	Version string `json:"version"`
}

func GetLocalIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()
	return conn.LocalAddr().(*net.UDPAddr).IP.String(), nil
}

func GetBaseURL(r *http.Request) string {
	domain := config.Server.Domain
	if domain != "" {
		return domain + "/"
	}

	// 如果没有配置域名，使用默认的
	scheme := "http"

	// 检查是否是HTTPS
	if isHTTPS(r) {
		scheme = "https"
	}

	// 获取主机名
	host := r.Host
	if host == "" {
		host = r.Header.Get("Host")
	}

	// 组合基础URL
	return scheme + "://" + host + "/"
}

// isHTTPS 检查请求是否是HTTPS
func isHTTPS(r *http.Request) bool {
	// 检查多种可能的HTTPS标识
	if r.TLS != nil {
		return true
	}

	// 检查 X-Forwarded-Proto 头
	if r.Header.Get("X-Forwarded-Proto") == "https" {
		return true
	}

	// 检查 Forwarded 头
	if forwarded := r.Header.Get("Forwarded"); strings.Contains(forwarded, "proto=https") {
		return true
	}

	// 检查 Origin 头
	if origin := r.Header.Get("Origin"); strings.HasPrefix(origin, "https:") {
		return true
	}

	// 检查端口
	if r.Header.Get("X-Forwarded-Port") == "443" {
		return true
	}

	return false
}

// AddImageDomain 为图片URL添加域名
// imageURL: 图片URL
// baseURL: 基础域名
// addDomain: 是否添加域名
func AddImageDomain(imageURL, baseURL string, addDomain bool) string {
	// 空URL检查
	if imageURL == "" {
		return ""
	}

	// 特定域名检查
	if strings.Contains(imageURL, "http://qn-cdn.jjjshop.net") {
		return imageURL
	}

	// 解析URL
	urlComponents, err := url.Parse(imageURL)
	if err != nil {
		return imageURL
	}

	// 处理基础URL
	baseURL = strings.TrimRight(baseURL, "/")

	var newURL strings.Builder

	// 添加域名部分
	if addDomain {
		if urlComponents.Path != "" && !strings.HasPrefix(urlComponents.Path, "/") {
			newURL.WriteString(baseURL)
		} else {
			newURL.WriteString(baseURL)
		}
	}

	// 添加路径
	if urlComponents.Path != "" {
		newURL.WriteString(urlComponents.Path)
	}

	// 添加查询参数
	if urlComponents.RawQuery != "" {
		newURL.WriteString("?")
		newURL.WriteString(urlComponents.RawQuery)
	}

	// 添加片段
	if urlComponents.Fragment != "" {
		newURL.WriteString("#")
		newURL.WriteString(urlComponents.Fragment)
	}

	return newURL.String()
}

func RemoveDomain(fileUrl string) string {
	// 解析URL
	parsedURL, err := url.Parse(fileUrl)
	if err != nil {
		return ""
	}
	// 构建新的URL，去掉域名和端口
	newURL := parsedURL.Path
	if parsedURL.RawQuery != "" {
		newURL += "?" + parsedURL.RawQuery
	}
	if parsedURL.Fragment != "" {
		newURL += "#" + parsedURL.Fragment
	}
	return newURL
}

// GetVersion 并返回版本号
func GetVersion() string {
	return config.Version
}

func DecimalAdd(f1 float64, fs ...float64) float64 {
	num := decimal.NewFromFloat(f1)
	for _, f := range fs {
		num = num.Add(decimal.NewFromFloat(f))
	}
	return num.InexactFloat64()
}

func DecimalMul(f1 float64, fs ...float64) float64 {
	num := decimal.NewFromFloat(f1)
	for _, f := range fs {
		num = num.Mul(decimal.NewFromFloat(f))
	}
	return num.InexactFloat64()
}

func FormatFloat(f1 float64) string {
	return strconv.FormatFloat(f1, 'f', -1, 64)
}

func DecimalSub(f1 float64, fs ...float64) float64 {
	num := decimal.NewFromFloat(f1)
	for _, f := range fs {
		num = num.Sub(decimal.NewFromFloat(f))
	}
	return num.InexactFloat64()
}

// ParseFloat 将字符串数字转为数字
func ParseFloat(numStr string) (float64, error) {
	if numStr == "" {
		return 0, nil
	}
	return strconv.ParseFloat(numStr, 64)
}

// Round 保留指定位数的小数（四舍五入）
// value: 需要处理的浮点数
// places: 保留的小数位数
func Round(value float64, places int32) float64 {
	return decimal.NewFromFloat(value).Round(places).InexactFloat64()
}
