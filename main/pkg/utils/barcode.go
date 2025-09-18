package utils

import (
	"regexp"
	"strings"
)

// ProcessBarcode 处理条码，按照以下规则：
// 1. 条码大于13位：超出部分截取
// 2. 条码非纯数字：过滤非数字字符
// 3. 条码前/后/中间包含空格：过滤空格
func ProcessBarcode(barcode string) string {
	if barcode == "" {
		return ""
	}

	// 1. 过滤空格
	processed := strings.ReplaceAll(barcode, " ", "")

	// 2. 过滤非数字字符
	reg := regexp.MustCompile(`[^0-9]`)
	processed = reg.ReplaceAllString(processed, "")

	// 3. 如果长度大于13位，截取前13位
	if len(processed) > 13 {
		processed = processed[:13]
	}

	return processed
}
