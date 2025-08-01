package member_req

import (
	"errors"
	"regexp"
)

type MemberNicknameUpdateReq struct {
	Nickname string `json:"nickname"` // 会员昵称
}

func (r *MemberNicknameUpdateReq) Validate() error {
	if r.Nickname == "" {
		return errors.New("昵称不能为空")
	}
	if len(r.Nickname) > 50 {
		return errors.New("昵称长度不能超过50")
	}
	// 正则匹配 不能输入特殊的字符 只允许字母、数字、中文和常见标点符号
	reg := regexp.MustCompile("^[a-zA-Z0-9\u4e00-\u9fa5]+$")
	if !reg.MatchString(r.Nickname) {
		return errors.New("昵称输入字符不合规")
	}
	return nil
}
