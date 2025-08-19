// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// AdminUserLoginLog is the golang structure of table ttpos_admin_user_login_log for DAO operations like Where/Data.
type AdminUserLoginLog struct {
	g.Meta      `orm:"table:ttpos_admin_user_login_log, do:true"`
	Id          interface{} // 自增ID
	AdminUserId interface{} // 用户ID
	Username    interface{} // 用户名
	Ip          interface{} // 登录ip
	Result      interface{} // 登录结果
	CreateTime  interface{} // 创建时间（时间戳）
	UpdateTime  interface{} // 更新时间（时间戳）
	DeleteTime  interface{} // 删除时间（时间戳）
}
