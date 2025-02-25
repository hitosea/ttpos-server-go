package validator

import (
	"github.com/go-playground/validator/v10"
	"strconv"
)

func autoOrderLimit(fl validator.FieldLevel) bool {
	if val, isString := fl.Field().Interface().(string); !isString {
		return false
	} else {
		limit, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return false
		} else if limit < 0.01 || limit > 100000000 {
			return false
		}
		return true
	}
}
