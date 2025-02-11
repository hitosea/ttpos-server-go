package utils

import (
	"strings"
)

// GetPlatform 根据User-Agent获取平台信息
func GetPlatform(userAgent string) int {
	switch {
	case strings.Contains(userAgent, "Android"):
		return 1
	case strings.Contains(userAgent, "iPhone"):
		return 2
	case strings.Contains(userAgent, "Mobile"):
		return 3
	default:
		return 0
	}
}
