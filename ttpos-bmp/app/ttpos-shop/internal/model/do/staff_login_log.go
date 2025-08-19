// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// StaffLoginLog is the golang structure of table ttpos_staff_login_log for DAO operations like Where/Data.
type StaffLoginLog struct {
	g.Meta     `orm:"table:ttpos_staff_login_log, do:true"`
	Id         interface{} // 自增ID
	Uuid       interface{} // UUID
	StaffUuid  interface{} // 员工UUID
	Username   interface{} // 用户名
	Ip         interface{} // 登录ip
	Result     interface{} // 登录结果
	CreateTime interface{} // 创建时间
	UpdateTime interface{} // 更新时间
	DeleteTime interface{} // 删除时间
}
