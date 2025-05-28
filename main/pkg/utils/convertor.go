package utils

import "fmt"

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
