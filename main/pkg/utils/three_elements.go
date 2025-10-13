package utils

import "ttpos-server-go/app/dto"

// 三元
func IfBool(is bool, trueVal, falseVal bool) bool {
	if is {
		return trueVal
	}
	return falseVal
}

// 三元
func IfUint64(is bool, trueVal, falseVal uint64) uint64 {
	if is {
		return trueVal
	}
	return falseVal
}

// 三元
func IfInt64(is bool, trueVal, falseVal int64) int64 {
	if is {
		return trueVal
	}
	return falseVal
}

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

// 三元
func IfFloat64(is bool, trueVal, falseVal float64) float64 {
	if is {
		return trueVal
	}
	return falseVal
}

// 三元
func IfSlice(is bool, trueVal, falseVal []dto.LocaleResponse) []dto.LocaleResponse {
	if is {
		return trueVal
	}
	return falseVal
}

// 三元
func IfUint(is bool, trueVal, falseVal uint) uint {
	if is {
		return trueVal
	}
	return falseVal
}
