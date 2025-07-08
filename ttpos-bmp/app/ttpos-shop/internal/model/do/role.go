// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Role is the golang structure of table ttpos_role for DAO operations like Where/Data.
type Role struct {
	g.Meta     `orm:"table:ttpos_role, do:true"`
	Id         interface{} // 自增ID
	Uuid       interface{} // 角色ID
	Name       interface{} // 角色名称
	Sort       interface{} // 排序(数字越小越靠前)
	CreateTime interface{} // 创建时间(时间戳)
	UpdateTime interface{} // 更新时间(时间戳)
	DeleteTime interface{} // 删除时间(时间戳)
}
