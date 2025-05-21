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
	case strings.Contains(strings.ToLower(userAgent), "fuchsia"):
		return 4
	case strings.Contains(strings.ToLower(userAgent), "ios"):
		return 5
	case strings.Contains(strings.ToLower(userAgent), "linux"):
		return 6
	case strings.Contains(strings.ToLower(userAgent), "macos"):
		return 7
	case strings.Contains(strings.ToLower(userAgent), "windows"):
		return 8
	default:
		return 0
	}
}
