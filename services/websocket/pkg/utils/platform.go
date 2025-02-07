package utils

import (
	"strings"
)

// GetPlatform 根据User-Agent获取平台信息
func GetPlatform(userAgent string) string {
	switch {
	case strings.Contains(userAgent, "Mobile"):
		return "Mobile"
	case strings.Contains(userAgent, "Android"):
		return "Android"
	case strings.Contains(userAgent, "iPhone"):
		return "iPhone"
	default:
		return "Web"
	}
}
