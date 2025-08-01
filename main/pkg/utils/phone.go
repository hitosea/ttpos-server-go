package utils

import (
	"errors"
	"regexp"
)

// ValidatePhone 验证手机号格式
// 支持9位泰国手机号（不能以0开头）和11位中国手机号（必须以1开头）
func ValidatePhone(phone string) error {
	// 检查是否为空
	if phone == "" {
		return errors.New("手机号不能为空")
	}

	// 检查是否全为数字
	matched, _ := regexp.MatchString(`^\d+$`, phone)
	if !matched {
		return errors.New("手机号只能包含数字")
	}

	// 获取手机号长度
	length := len(phone)

	switch length {
	case 9:
		// 9位泰国手机号，不能以0开头
		if phone[0] == '0' {
			return errors.New("9位手机号不能以0开头")
		}
		return nil
	case 11:
		return nil
	default:
		return errors.New("手机号长度必须为9位或11位")
	}
}
