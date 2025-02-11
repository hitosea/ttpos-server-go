package utils

// BoolToUint 将布尔值转换为无符号整数
func BoolToUint(b bool) uint {
	if b {
		return 1
	}
	return 0
}
