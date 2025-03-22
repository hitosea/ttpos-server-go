package utils

// 三元
func IfInt(is bool, trueVal, falseVal int) int {
	if is {
		return trueVal
	}
	return falseVal
}

// 三元
func IfString(is bool, trueVal, falseVal string) string {
	if is {
		return trueVal
	}
	return falseVal
}
