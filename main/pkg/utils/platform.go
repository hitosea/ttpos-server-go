package utils

import (
	"strings"
)

// GetPlatform 根据User-Agent获取平台信息
func GetPlatform(userAgent string) int {
	switch {
	case strings.Contains(strings.ToLower(userAgent), "android"):
		return 1
	case strings.Contains(strings.ToLower(userAgent), "iphone"):
		return 2
	case strings.Contains(strings.ToLower(userAgent), "mobile"):
		return 3
	default:
		return 0
	}
}
