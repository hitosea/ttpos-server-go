// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// AdminUserOptLog is the golang structure of table ttpos_admin_user_opt_log for DAO operations like Where/Data.
type AdminUserOptLog struct {
	g.Meta      `orm:"table:ttpos_admin_user_opt_log, do:true"`
	Id          interface{} // 自增ID
	AdminUserId interface{} // 用户ID
	Title       interface{} // 标题
	Url         interface{} // 访问url
	RequestType interface{} // 请求类型
	Browser     interface{} // 浏览器
	Agent       interface{} // 浏览器信息
	Content     interface{} // 操作内容
	Ip          interface{} // 登录ip
	CreateTime  interface{} // 创建时间（时间戳）
	UpdateTime  interface{} // 更新时间（时间戳）
	DeleteTime  interface{} // 删除时间（时间戳）
}
