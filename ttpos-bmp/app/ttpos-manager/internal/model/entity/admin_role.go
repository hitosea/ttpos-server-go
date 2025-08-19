// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// AdminRole is the golang structure for table admin_role.
type AdminRole struct {
	Id         uint   `json:"id"         orm:"id"          description:"自增ID"`        // 自增ID
	RoleName   string `json:"roleName"   orm:"role_name"   description:"角色名称"`        // 角色名称
	Sort       int    `json:"sort"       orm:"sort"        description:"排序(数字越小越靠前)"` // 排序(数字越小越靠前)
	CreateTime int    `json:"createTime" orm:"create_time" description:"创建时间（时间戳）"`   // 创建时间（时间戳）
	UpdateTime int    `json:"updateTime" orm:"update_time" description:"更新时间（时间戳）"`   // 更新时间（时间戳）
	DeleteTime int    `json:"deleteTime" orm:"delete_time" description:"删除时间（时间戳）"`   // 删除时间（时间戳）
}
