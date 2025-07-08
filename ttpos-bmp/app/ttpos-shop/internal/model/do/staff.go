// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Staff is the golang structure of table ttpos_staff for DAO operations like Where/Data.
type Staff struct {
	g.Meta              `orm:"table:ttpos_staff, do:true"`
	Id                  interface{} // 自增ID
	Uuid                interface{} // 员工ID
	CompanyUuid         interface{} // 集团ID
	Username            interface{} // 用户名
	Password            interface{} // 登录密码
	Phone               interface{} // 手机号
	PasswordChangeCount interface{} // 修改密码次数
	PasswordChangeTime  interface{} // 修改密码时间
	RealName            interface{} // 姓名
	IsSuper             interface{} // 是否为超级管理员0不是,1是
	UserType            interface{} // 账号类型0总台1门店
	IsDisable           interface{} // 是否禁用1禁用,0未禁用
	BindKey             interface{} // 绑定的设备key
	CashierOnline       interface{} // 收银员当班 0-不在线 1-在线
	CashierLoginTime    interface{} // 收银员当班登录时间
	DutyNo              interface{} // 当班编号
	CreateTime          interface{} // 创建时间(时间戳)
	UpdateTime          interface{} // 更新时间(时间戳)
	DeleteTime          interface{} // 删除时间(时间戳)
}
