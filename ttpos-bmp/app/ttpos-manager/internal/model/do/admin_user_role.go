// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// AdminUserRole is the golang structure of table ttpos_admin_user_role for DAO operations like Where/Data.
type AdminUserRole struct {
	g.Meta      `orm:"table:ttpos_admin_user_role, do:true"`
	Id          interface{} // 自增ID
	AdminUserId interface{} // 超管用户ID
	RoleId      interface{} // 角色ID
	CreateTime  interface{} // 创建时间（时间戳）
	UpdateTime  interface{} // 更新时间（时间戳）
	DeleteTime  interface{} // 删除时间（时间戳）
}
