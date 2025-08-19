// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// AdminRoleAccess is the golang structure of table ttpos_admin_role_access for DAO operations like Where/Data.
type AdminRoleAccess struct {
	g.Meta     `orm:"table:ttpos_admin_role_access, do:true"`
	Id         interface{} // 自增ID
	RoleId     interface{} // 角色ID
	AccessId   interface{} // 权限ID
	CreateTime interface{} // 创建时间（时间戳）
	UpdateTime interface{} // 更新时间（时间戳）
	DeleteTime interface{} // 删除时间（时间戳）
}
