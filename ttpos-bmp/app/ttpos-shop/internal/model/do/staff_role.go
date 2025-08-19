// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// StaffRole is the golang structure of table ttpos_staff_role for DAO operations like Where/Data.
type StaffRole struct {
	g.Meta     `orm:"table:ttpos_staff_role, do:true"`
	Id         interface{} // 自增ID
	Uuid       interface{} // 员工角色关系ID
	StaffUuid  interface{} // 超管用户ID
	RoleUuid   interface{} // 角色ID
	CreateTime interface{} // 创建时间(时间戳)
	UpdateTime interface{} // 更新时间(时间戳)
	DeleteTime interface{} // 删除时间(时间戳)
}
