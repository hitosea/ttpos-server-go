// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// AdminUserRole is the golang structure for table admin_user_role.
type AdminUserRole struct {
	Id          uint `json:"id"          orm:"id"            description:"自增ID"`      // 自增ID
	AdminUserId int  `json:"adminUserId" orm:"admin_user_id" description:"超管用户ID"`    // 超管用户ID
	RoleId      int  `json:"roleId"      orm:"role_id"       description:"角色ID"`      // 角色ID
	CreateTime  int  `json:"createTime"  orm:"create_time"   description:"创建时间（时间戳）"` // 创建时间（时间戳）
	UpdateTime  int  `json:"updateTime"  orm:"update_time"   description:"更新时间（时间戳）"` // 更新时间（时间戳）
	DeleteTime  int  `json:"deleteTime"  orm:"delete_time"   description:"删除时间（时间戳）"` // 删除时间（时间戳）
}
