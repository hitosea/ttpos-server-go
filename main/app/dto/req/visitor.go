package req

import "errors"

// VisitorLoginReq 游客登录请求
type VisitorLoginReq struct {
	CompanyUuid uint64 `json:"company_uuid" form:"company_uuid" binding:"required"` // 集团ID
	DeviceId    string `json:"device_id" form:"device_id" binding:"required"`       // 设备ID，用于标识游客
}

// Validate 验证游客登录请求
func (req *VisitorLoginReq) Validate() error {
	if req.CompanyUuid == 0 {
		return errors.New("商家ID不能为空")
	}
	if req.DeviceId == "" {
		return errors.New("设备ID不能为空")
	}
	return nil
}

// VisitorLoginReqMessage 游客登录请求错误提示
var VisitorLoginReqMessage = map[string]string{
	"company_uuid.required": "商家ID不能为空",
	"device_id.required":    "设备ID不能为空",
}

// VisitorBindPhoneReq 游客绑定手机号请求
type VisitorBindPhoneReq struct {
	MemberUuid uint64 `json:"member_uuid" form:"member_uuid" binding:"required"` // 游客会员ID
	Phone      string `json:"phone" form:"phone" binding:"required"`             // 手机号
	Code       string `json:"code" form:"code" binding:"required"`               // 验证码
}

// Validate 验证游客绑定手机号请求
func (req *VisitorBindPhoneReq) Validate() error {
	if req.MemberUuid == 0 {
		return errors.New("会员ID不能为空")
	}
	if req.Phone == "" {
		return errors.New("手机号不能为空")
	}
	if req.Code == "" {
		return errors.New("验证码不能为空")
	}
	return nil
}

// VisitorBindPhoneReqMessage 游客绑定手机号请求错误提示
var VisitorBindPhoneReqMessage = map[string]string{
	"member_uuid.required": "会员ID不能为空",
	"phone.required":       "手机号不能为空",
	"code.required":        "验证码不能为空",
}
