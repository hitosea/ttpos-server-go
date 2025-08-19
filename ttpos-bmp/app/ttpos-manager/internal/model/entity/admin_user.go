// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// AdminUser is the golang structure for table admin_user.
type AdminUser struct {
	AdminUserId uint   `json:"adminUserId" orm:"admin_user_id" description:"自增ID"`          // 自增ID
	Username    string `json:"username"    orm:"username"      description:"用户名"`           // 用户名
	Phone       string `json:"phone"       orm:"phone"         description:"手机号"`           // 手机号
	Password    string `json:"password"    orm:"password"      description:"登录密码"`          // 登录密码
	RealName    string `json:"realName"    orm:"real_name"     description:"姓名"`            // 姓名
	IsSuper     int    `json:"isSuper"     orm:"is_super"      description:"是否超级管理员"`       // 是否超级管理员
	Status      int    `json:"status"      orm:"status"        description:"状态(0未启用,1已启用)"` // 状态(0未启用,1已启用)
	CreateTime  int    `json:"createTime"  orm:"create_time"   description:"创建时间（时间戳）"`     // 创建时间（时间戳）
	UpdateTime  int    `json:"updateTime"  orm:"update_time"   description:"更新时间（时间戳）"`     // 更新时间（时间戳）
	DeleteTime  int    `json:"deleteTime"  orm:"delete_time"   description:"删除时间（时间戳）"`     // 删除时间（时间戳）
}
