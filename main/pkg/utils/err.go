package utils

import (
	"regexp"
	"strings"
)

func IsNotFoundRecord(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "record not found")
}

// 简短erpnext的报错信息。通过`line 数字通配符,`（如"line 525,"） 来截取报错信息
func ShortenErpnextError(errorMsg string) string {
	if errorMsg == "" {
		return ""
	}

	// 使用正则表达式匹配 "line 数字," 模式
	re := regexp.MustCompile(`line \d+,`)
	matches := re.FindAllString(errorMsg, -1)

	if len(matches) == 0 {
		// 如果没有找到 line 数字模式，返回原始错误信息
		return errorMsg
	}

	// 使用最后一个匹配的 "line 数字," 来分割字符串
	firstMatch := matches[len(matches)-1]
	parts := strings.Split(errorMsg, firstMatch)

	if len(parts) > 1 {
		return parts[1]
	}

	return errorMsg
}

// "Item WPR数字通配符 is disabled" 判断是否是这个错误，并识别出erp_code
func IsItemDisabledError(errorMsg string) (bool, string) {
	if strings.Contains(errorMsg, "Item WPR") {
		re := regexp.MustCompile(`Item (WPR\d+) is disabled`)
		matches := re.FindStringSubmatch(errorMsg)
		if len(matches) > 1 {
			return true, matches[1]
		}
	}
	return false, ""
}
