package member_req

import (
	"errors"
	"ttpos-server-go/app/constant"
)

type MemberLoginInfoReq struct {
	CompanyUuid uint64 `json:"company_uuid" form:"company_uuid"` // 集团ID
}

type MemberSendCodeReq struct {
	Phone       string `json:"phone" form:"phone"`               // 手机号
	AreaCode    string `json:"area_code" form:"area_code"`       // 区号
	CompanyUuid uint64 `json:"company_uuid" form:"company_uuid"` // 集团ID
}

func (req *MemberSendCodeReq) Validate() error {
	if req.CompanyUuid == 0 {
		return errors.New("商家ID不能为空")
	}
	// 增加全数字校验
	if !isNumeric(req.Phone) {
		return errors.New("手机号格式不正确")
	}
	// 判断 req.Phone 是否是手机号
	if req.AreaCode == constant.ChinaPrefix {
		if len(req.Phone) != 11 {
			return errors.New("手机号格式不正确")
		}
		if req.Phone[0] != '1' {
			return errors.New("手机号格式不正确")
		}
	} else if req.AreaCode == constant.ThailandPrefix {
		if len(req.Phone) != 10 {
			return errors.New("手机号格式不正确")
		}
		if req.Phone[0] != '0' {
			return errors.New("手机号格式不正确")
		}
	} else {
		return errors.New("区号格式不正确")
	}
	return nil
}

type MemberLoginReq struct {
	Phone       string `json:"phone" form:"phone"`               // 手机号
	CompanyUuid uint64 `json:"company_uuid" form:"company_uuid"` // 集团ID
	Code        string `json:"code" form:"code"`                 // 验证码
}

// 判断是否全为数字
func isNumeric(s string) bool {
	for _, v := range s {
		if v < '0' || v > '9' {
			return false
		}
	}
	return true
}

func (req *MemberLoginReq) Validate() error {
	if req.CompanyUuid == 0 {
		return errors.New("商家ID不能为空")
	}
	if req.Phone == "" {
		return errors.New("手机号不能为空")
	}
	if req.Code == "" {
		return errors.New("验证码不能为空")
	}
	// 增加全数字校验
	if !isNumeric(req.Phone) || len(req.Phone) != 11 && len(req.Phone) != 10 {
		return errors.New("手机号格式不正确")
	}
	// 判断 req.Phone 是否是手机号
	if len(req.Phone) == 11 && req.Phone[0] != '1' {
		return errors.New("手机号格式不正确")
	} else if len(req.Phone) == 10 && req.Phone[0] != '0' {
		return errors.New("手机号格式不正确")
	}
	return nil
}
