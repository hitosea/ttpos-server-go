// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// UploadGroup is the golang structure for table upload_group.
type UploadGroup struct {
	GroupId    uint   `json:"groupId"    orm:"group_id"    description:"自增ID"`          // 自增ID
	GroupType  string `json:"groupType"  orm:"group_type"  description:"文件类型"`          // 文件类型
	GroupName  string `json:"groupName"  orm:"group_name"  description:"分类名称"`          // 分类名称
	Sort       uint   `json:"sort"       orm:"sort"        description:"分类排序(数字越小越靠前)"` // 分类排序(数字越小越靠前)
	CreateTime int    `json:"createTime" orm:"create_time" description:"创建时间（时间戳）"`     // 创建时间（时间戳）
	UpdateTime int    `json:"updateTime" orm:"update_time" description:"更新时间（时间戳）"`     // 更新时间（时间戳）
	DeleteTime int    `json:"deleteTime" orm:"delete_time" description:"删除时间（时间戳）"`     // 删除时间（时间戳）
}
