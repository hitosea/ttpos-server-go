package req

type UpdateStaffReq struct {
	Uuid            uint64   `json:"uuid" binding:"required"`                               // 员工ID
	RealName        string   `json:"real_name" binding:"required,max=100"`                  // 姓名，限制100个字符
	Username        string   `json:"username" binding:"required,max=64,email"`              // 邮箱，限制64个字符
	Phone           string   `json:"phone" binding:"required,max=20"`                       // 手机号，限制20个字符
	Roles           []uint64 `json:"roles" binding:"required"`                              // 角色ID列表
	Password        string   `json:"password" binding:"omitempty,strong_password"`          // 密码，如果不为空，则不能包括空格，长度为8-16个字符必须包含字母、数字、符号中至少2种
	ConfirmPassword string   `json:"confirm_password" binding:"omitempty,eqfield=Password"` // 确认密码
}

var UpdateStaffRequestMessage = map[string]string{
	"uuid.required":            "员工ID不能为空",
	"real_name.required":       "姓名不能为空",
	"real_name.max":            "姓名不能超过100个字",
	"username.required":        "邮箱不能为空",
	"username.max":             "邮箱不能超过64个字",
	"username.email":           "邮箱必须是有效的邮箱格式",
	"phone.required":           "手机号不能为空",
	"phone.max":                "手机号不能超过20个字",
	"roles.required":           "角色不能为空",
	"password.strong_password": "密码不符合要求：不能包含空格，长度为8-16个字符，必须包含字母、数字、符号中至少2种",
	"confirm_password.eqfield": "两次密码不一致",
}

type UpdateStaffStatusReq struct {
	Uuid   uint64 `json:"uuid" binding:"required"`             // 员工ID
	Status *int   `json:"status" binding:"required,oneof=1 0"` // 状态 1:启用 0:禁用
}

var UpdateStaffStatusRequestMessage = map[string]string{
	"uuid.required":   "员工ID不能为空",
	"status.required": "状态不能为空",
}

type DeleteStaffReq struct {
	Uuid uint64 `json:"uuid" binding:"required"` // 员工ID
}

var DeleteStaffRequestMessage = map[string]string{
	"uuid.required": "员工ID不能为空",
}

type AddStaffReq struct {
	RealName        string   `json:"real_name" binding:"required,max=100"`                 // 姓名，限制100个字符
	Username        string   `json:"username" binding:"required,max=64,email"`             // 邮箱，限制64个字符
	Phone           string   `json:"phone" binding:"required,max=20"`                      // 手机号，限制20个字符
	Roles           []uint64 `json:"roles" binding:"required"`                             // 角色ID列表
	Password        string   `json:"password" binding:"required,strong_password"`          // 密码，如果不为空，则不能包括空格，长度为8-16个字符必须包含字母、数字、符号中至少2种
	ConfirmPassword string   `json:"confirm_password" binding:"required,eqfield=Password"` // 确认密码
}

var AddStaffRequestMessage = map[string]string{
	"real_name.required":        "姓名不能为空",
	"real_name.max":             "姓名不能超过100个字",
	"username.required":         "邮箱不能为空",
	"username.max":              "邮箱不能超过64个字",
	"username.email":            "邮箱必须是有效的邮箱格式",
	"phone.required":            "手机号不能为空",
	"phone.max":                 "手机号不能超过20个字",
	"roles.required":            "角色不能为空",
	"password.required":         "密码不能为空",
	"password.strong_password":  "密码不符合要求：不能包含空格，长度为8-16个字符，必须包含字母、数字、符号中至少2种",
	"confirm_password.required": "确认密码不能为空",
	"confirm_password.eqfield":  "两次密码不一致",
}

type AddRoleReq struct {
	Name        string   `json:"name" binding:"required"`         // 角色名称
	AccessUuids []uint64 `json:"access_uuids" binding:"required"` // 权限ID列表
}

type UpdateRoleReq struct {
	Uuid        uint64   `json:"uuid" binding:"required"`         // 角色ID
	Name        string   `json:"name" binding:"required"`         // 角色名称
	AccessUuids []uint64 `json:"access_uuids" binding:"required"` // 权限ID列表
}

type DeleteRoleReq struct {
	Uuid uint64 `json:"uuid" binding:"required"` // 角色ID
}

type GetRoleReq struct {
	Uuid uint64 `form:"uuid" binding:"required"` // 角色ID
}
