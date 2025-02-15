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
	Source      string
	DeviceId    string
	CompanyUuid uint64
	StaffUuid   uint64
	UrlPath     string
	Assistant   Assistant
}
