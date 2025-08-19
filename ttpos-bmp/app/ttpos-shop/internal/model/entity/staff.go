// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// Staff is the golang structure for table staff.
type Staff struct {
	Id                  uint   `json:"id"                  orm:"id"                    description:"自增ID"`             // 自增ID
	Uuid                uint64 `json:"uuid"                orm:"uuid"                  description:"员工ID"`             // 员工ID
	CompanyUuid         uint64 `json:"companyUuid"         orm:"company_uuid"          description:"集团ID"`             // 集团ID
	Username            string `json:"username"            orm:"username"              description:"用户名"`              // 用户名
	Password            string `json:"password"            orm:"password"              description:"登录密码"`             // 登录密码
	Phone               string `json:"phone"               orm:"phone"                 description:"手机号"`              // 手机号
	PasswordChangeCount int    `json:"passwordChangeCount" orm:"password_change_count" description:"修改密码次数"`           // 修改密码次数
	PasswordChangeTime  uint   `json:"passwordChangeTime"  orm:"password_change_time"  description:"修改密码时间"`           // 修改密码时间
	RealName            string `json:"realName"            orm:"real_name"             description:"姓名"`               // 姓名
	IsSuper             int    `json:"isSuper"             orm:"is_super"              description:"是否为超级管理员0不是,1是"`   // 是否为超级管理员0不是,1是
	UserType            int    `json:"userType"            orm:"user_type"             description:"账号类型0总台1门店"`       // 账号类型0总台1门店
	IsDisable           int    `json:"isDisable"           orm:"is_disable"            description:"是否禁用1禁用,0未禁用"`     // 是否禁用1禁用,0未禁用
	BindKey             string `json:"bindKey"             orm:"bind_key"              description:"绑定的设备key"`         // 绑定的设备key
	CashierOnline       int    `json:"cashierOnline"       orm:"cashier_online"        description:"收银员当班 0-不在线 1-在线"` // 收银员当班 0-不在线 1-在线
	CashierLoginTime    uint   `json:"cashierLoginTime"    orm:"cashier_login_time"    description:"收银员当班登录时间"`        // 收银员当班登录时间
	DutyNo              string `json:"dutyNo"              orm:"duty_no"               description:"当班编号"`             // 当班编号
	CreateTime          uint   `json:"createTime"          orm:"create_time"           description:"创建时间(时间戳)"`        // 创建时间(时间戳)
	UpdateTime          uint   `json:"updateTime"          orm:"update_time"           description:"更新时间(时间戳)"`        // 更新时间(时间戳)
	DeleteTime          uint   `json:"deleteTime"          orm:"delete_time"           description:"删除时间(时间戳)"`        // 删除时间(时间戳)
}
