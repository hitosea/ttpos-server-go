// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// AdminRole is the golang structure of table ttpos_admin_role for DAO operations like Where/Data.
type AdminRole struct {
	g.Meta     `orm:"table:ttpos_admin_role, do:true"`
	Id         interface{} // 自增ID
	RoleName   interface{} // 角色名称
	Sort       interface{} // 排序(数字越小越靠前)
	CreateTime interface{} // 创建时间（时间戳）
	UpdateTime interface{} // 更新时间（时间戳）
	DeleteTime interface{} // 删除时间（时间戳）
}
