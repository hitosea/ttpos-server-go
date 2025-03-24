package utils

// 三元
func IfInt(is bool, trueVal, falseVal int) int {
	if is {
		return trueVal
	}
	return falseVal
}

func IfString(is bool, trueVal, falseVal string) string {
	if is {
		return trueVal
	}
	return falseVal
}

func IfFloat64(is bool, trueVal, falseVal float64) float64 {
	if is {
		return trueVal
	}
	return falseVal
}
