// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// AdminUser is the golang structure of table ttpos_admin_user for DAO operations like Where/Data.
type AdminUser struct {
	g.Meta      `orm:"table:ttpos_admin_user, do:true"`
	AdminUserId interface{} // 自增ID
	Username    interface{} // 用户名
	Phone       interface{} // 手机号
	Password    interface{} // 登录密码
	RealName    interface{} // 姓名
	IsSuper     interface{} // 是否超级管理员
	Status      interface{} // 状态(0未启用,1已启用)
	CreateTime  interface{} // 创建时间（时间戳）
	UpdateTime  interface{} // 更新时间（时间戳）
	DeleteTime  interface{} // 删除时间（时间戳）
}
