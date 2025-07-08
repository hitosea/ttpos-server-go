// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// AdminUserLoginLog is the golang structure for table admin_user_login_log.
type AdminUserLoginLog struct {
	Id          uint   `json:"id"          orm:"id"            description:"自增ID"`      // 自增ID
	AdminUserId int    `json:"adminUserId" orm:"admin_user_id" description:"用户ID"`      // 用户ID
	Username    string `json:"username"    orm:"username"      description:"用户名"`       // 用户名
	Ip          string `json:"ip"          orm:"ip"            description:"登录ip"`      // 登录ip
	Result      string `json:"result"      orm:"result"        description:"登录结果"`      // 登录结果
	CreateTime  int    `json:"createTime"  orm:"create_time"   description:"创建时间（时间戳）"` // 创建时间（时间戳）
	UpdateTime  int    `json:"updateTime"  orm:"update_time"   description:"更新时间（时间戳）"` // 更新时间（时间戳）
	DeleteTime  int    `json:"deleteTime"  orm:"delete_time"   description:"删除时间（时间戳）"` // 删除时间（时间戳）
}
