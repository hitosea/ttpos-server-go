package utils

import (
	"fmt"
	"strconv"
)

// BoolToUint 将布尔值转换为无符号整数
func BoolToUint(b bool) uint {
	if b {
		return 1
	}
	return 0
}

func Uint64OrStringToString(value any) string {
	switch v := value.(type) {
	case float64:
		return fmt.Sprintf("%.0f", v)
	case uint64:
		return fmt.Sprintf("%d", v)
	case string:
		return v
	default:
		return fmt.Sprintf("%v", value)
	}
}

// StringToUint64 将字符串转换为无符号整数
func StringToUint64(s string) (uint64, error) {
	return strconv.ParseUint(s, 10, 64)
}
