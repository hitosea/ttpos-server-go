// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// AdminAccess is the golang structure of table ttpos_admin_access for DAO operations like Where/Data.
type AdminAccess struct {
	g.Meta       `orm:"table:ttpos_admin_access, do:true"`
	Id           interface{} // 自增ID
	Name         interface{} // 权限名称
	Path         interface{} // 路由地址
	ApiPath      interface{} // 后端路由地址
	ParentId     interface{} // 父级ID
	Sort         interface{} // 排序(数字越小越靠前)
	Icon         interface{} // 菜单图标
	RedirectName interface{} // 重定向名称
	IsRoute      interface{} // 是否路由 0=不是1=是
	IsMenu       interface{} // 是否菜单 0=不是1=是
	IsShow       interface{} // 是否显示 0=不是1=是
	Remark       interface{} // 描述
	IsSupplier   interface{} // 是否门店菜单 0=不是1=是
	CreateTime   interface{} // 创建时间（时间戳）
	UpdateTime   interface{} // 更新时间（时间戳）
	DeleteTime   interface{} // 删除时间（时间戳）
}
