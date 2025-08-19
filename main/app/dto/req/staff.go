package req

type UpdateStaffReq struct {
	Uuid            uint64   `json:"uuid" binding:"required"`                               // 管理员ID
	Username        string   `json:"username" binding:"required,email"`                     // 账号，邮箱 邮箱格式
	RealName        string   `json:"real_name" binding:"required"`                          // 姓名
	Phone           string   `json:"phone" binding:"required,max=20"`                       // 手机号，最多20位
	Roles           []uint64 `json:"roles" binding:"required"`                              // 角色ID列表
	Password        string   `json:"password" binding:"omitempty,strong_password"`          // 密码，如果不为空，则不能包括空格，长度为8-16个字符必须包含字母、数字、符号中至少2种
	ConfirmPassword string   `json:"confirm_password" binding:"omitempty,eqfield=Password"` // 确认密码
}

var UpdateStaffRequestMessage = map[string]string{
	"uuid.required":            "员工ID不能为空",
	"username.required":        "用户名不能为空",
	"username.email":           "用户名必须是有效的邮箱格式",
	"real_name.required":       "姓名不能为空",
	"phone.required":           "手机号不能为空",
	"roles.required":           "角色不能为空",
	"password.strong_password": "密码不符合要求：不能包含空格，长度为8-16个字符，必须包含字母、数字、符号中至少2种",
	"confirm_password.eqfield": "两次密码输入不一致",
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
	Username        string   `json:"username" binding:"required,email"`                    // 账号，邮箱 邮箱格式
	RealName        string   `json:"real_name" binding:"required"`                         // 姓名
	Phone           string   `json:"phone" binding:"required"`                             // 手机号
	Roles           []uint64 `json:"roles" binding:"required"`                             // 角色ID列表
	Password        string   `json:"password" binding:"required,strong_password"`          // 密码，如果不为空，则不能包括空格，长度为8-16个字符必须包含字母、数字、符号中至少2种
	ConfirmPassword string   `json:"confirm_password" binding:"required,eqfield=Password"` // 确认密码
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
