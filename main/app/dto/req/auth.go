package req

type LoginReq struct {
	Username string `json:"username" binding:"required"`   // 用户名
	Password string `json:"password"  binding:"required"`  // 密码
	Code     string `json:"code"  binding:"required"`      // 验证码
	DeviceId string `json:"device_id"  binding:"required"` // 设备ID
	Brand    string `json:"brand"`                         // 品牌名称
	Source   string `json:"-"`
}

type BindCashierReq struct {
	DeviceId    string `json:"device_id" binding:"required"` // 收银机设备ID
	CashierUuid uint64 `json:"cashier_uuid"`                 // 收银员Uuid
}

var LoginRequestMessage = map[string]string{
	"username.required": "用户名不能为空",
	"password.required": "密码不能为空",
}

type Assistant struct {
	DeviceId  string `json:"device_id"`  // 收银设备ID
	StaffUuid uint64 `json:"staff_uuid"` // 收银员工ID
}

type Authenticate struct {
	Source   string
	DeviceId string

	CompanyUuid uint64
	StaffUuid   uint64
	UrlPath     string
	Assistant   Assistant

	TokenIssuedAt int64
}

type ChangePasswordReq struct {
	OldPassword     string `json:"old_password" binding:"required"`                         // 旧密码
	NewPassword     string `json:"new_password" binding:"required,strong_password"`         // 新密码，不能包括空格，长度为8-16个字符必须包含字母、数字、符号中至少2种
	ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=NewPassword"` // 确认密码
}

var ChangePasswordRequestMessage = map[string]string{
	"old_password.required":        "旧密码不能为空",
	"new_password.required":        "新密码不能为空",
	"new_password.strong_password": "新密码不符合要求：不能包含空格，长度为8-16个字符，必须包含字母、数字、符号中至少2种",
	"confirm_password.required":    "确认密码不能为空",
	"confirm_password.eqfield":     "两次密码输入不一致",
}

// StoreSwitchReq 门店切换请求
type StoreSwitchReq struct {
	CompanyUuid uint64 `json:"company_uuid" binding:"required"` // 要切换到的门店UUID
}
