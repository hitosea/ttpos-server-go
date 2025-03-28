package utils

import (
	"fmt"
	"strconv"
	"ttpos-server-go/app/errors"
)

func ConvertToFloat64(value any) (float64, error) {
	// 如果 value 为 nil，则返回 0
	// 有的数据是 nil，没有这个字段，所以需要判断
	if value == nil {
		return 0, nil
	}
	numStr := fmt.Sprintf("%v", value)
	price, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0, errors.WithMessage(err, fmt.Sprintf("strconv.ParseFloat failed, price: %v", value))
	}
	return price, nil
}
