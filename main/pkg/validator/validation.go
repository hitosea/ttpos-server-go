package validator

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
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

// strongPassword 强密码验证
// 规则：不能包括空格，长度为8-16个字符，必须包含字母、数字、符号中至少2种
func strongPassword(fl validator.FieldLevel) bool {
	password, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}

	// 如果密码为空，则跳过验证（通过 omitempty 处理）
	if password == "" {
		return true
	}

	// 1. 检查是否包含空格
	if strings.Contains(password, " ") {
		return false
	}

	// 2. 检查长度是否为8-16个字符
	if len(password) < 8 || len(password) > 16 {
		return false
	}

	// 3. 检查是否包含字母、数字、符号中至少2种
	hasLetter := regexp.MustCompile(`[a-zA-Z]`).MatchString(password)     // 包含字母
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)        // 包含数字
	hasSymbol := regexp.MustCompile(`[^a-zA-Z0-9]`).MatchString(password) // 包含符号

	// 计算符合条件的类型数量
	typeCount := 0
	if hasLetter {
		typeCount++
	}
	if hasNumber {
		typeCount++
	}
	if hasSymbol {
		typeCount++
	}

	// 至少包含2种类型
	return typeCount >= 2
}

// openingHours 营业时间格式验证
// 规则：必须符合 HH:MM-HH:MM 格式，其中 HH 为 00-23，MM 为 00-59
// 支持跨天营业，如：09:00-22:00, 23:00-02:00
func openingHours(fl validator.FieldLevel) bool {
	hours, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}

	// 如果营业时间为空，则跳过验证（通过 omitempty 处理）
	if hours == "" {
		return true
	}

	// 使用正则表达式验证格式：HH:MM-HH:MM
	// HH: 00-23
	// MM: 00-59
	// 中间用连字符分隔
	pattern := `^([01]?[0-9]|2[0-3]):([0-5][0-9])-([01]?[0-9]|2[0-3]):([0-5][0-9])$`

	matched, _ := regexp.MatchString(pattern, hours)
	return matched
}
