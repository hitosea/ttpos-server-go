// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Access is the golang structure of table ttpos_access for DAO operations like Where/Data.
type Access struct {
	g.Meta           `orm:"table:ttpos_access, do:true"`
	Id               interface{} // 自增ID
	Uuid             interface{} // 权限ID
	Name             interface{} // 权限名称
	Path             interface{} // 路由地址
	ApiPath          interface{} // 后端路由地址
	ParentUuid       interface{} // 父级ID
	Sort             interface{} // 排序(数字越小越靠前)
	Icon             interface{} // 菜单图标
	RedirectName     interface{} // 重定向名称
	IsRoute          interface{} // 是否是路由 0=不是1=是
	IsMenu           interface{} // 是否是菜单 0不是 1是
	IsShow           interface{} // 是否显示1=显示0=不显示
	PlusCategoryUuid interface{} // 插件分类ID
	Remark           interface{} // 描述
	IsSupplier       interface{} // 是否门店菜单0否1是
	CreateTime       interface{} // 创建时间(时间戳)
	UpdateTime       interface{} // 更新时间(时间戳)
	DeleteTime       interface{} // 删除时间(时间戳)
}
