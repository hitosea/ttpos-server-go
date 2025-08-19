// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// RoleAccess is the golang structure of table ttpos_role_access for DAO operations like Where/Data.
type RoleAccess struct {
	g.Meta     `orm:"table:ttpos_role_access, do:true"`
	Id         interface{} // 自增ID
	Uuid       interface{} // 角色权限关系ID
	RoleUuid   interface{} // 角色ID
	AccessUuid interface{} // 权限ID
	CreateTime interface{} // 创建时间(时间戳)
	UpdateTime interface{} // 更新时间(时间戳)
	DeleteTime interface{} // 删除时间(时间戳)
}
