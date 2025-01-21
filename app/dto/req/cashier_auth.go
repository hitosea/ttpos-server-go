package req

type CashierLoginRequest struct {
	Username string `json:"user_name" binding:"required"`
	Password string `json:"password"  binding:"required"`
	Code     string `json:"code"  binding:"required"`
	DeviceId string `json:"device_id"  binding:"required"`
	DeviceIP string `json:"device_ip"`
}

var CashierLoginRequestMessage = map[string]string{
	"username.required": "用户名不能为空",
	"password.required": "密码不能为空",
}
