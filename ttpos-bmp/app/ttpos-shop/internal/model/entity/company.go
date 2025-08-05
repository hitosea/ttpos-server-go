// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// Company is the golang structure for table company.
type Company struct {
	Id            uint   `json:"id"            orm:"id"              description:"自增ID"`                  // 自增ID
	Uuid          uint64 `json:"uuid"          orm:"uuid"            description:"集团ID"`                  // 集团ID
	Name          string `json:"name"          orm:"name"            description:"集团名称"`                  // 集团名称
	Logo          string `json:"logo"          orm:"logo"            description:"logo"`                  // logo
	ExpireTime    uint   `json:"expireTime"    orm:"expire_time"     description:"过期时间;not null"`         // 过期时间;not null
	AuthDay       int    `json:"authDay"       orm:"auth_day"        description:"授权时间(天) 0为永不过期"`        // 授权时间(天) 0为永不过期
	Status        int    `json:"status"        orm:"status"          description:"状态 1-启用 0-禁用;not null"` // 状态 1-启用 0-禁用;not null
	AuthStartTime uint   `json:"authStartTime" orm:"auth_start_time" description:"授权开始时间(时间戳)"`           // 授权开始时间(时间戳)
	CreateTime    uint   `json:"createTime"    orm:"create_time"     description:"创建时间(时间戳)"`             // 创建时间(时间戳)
	UpdateTime    uint   `json:"updateTime"    orm:"update_time"     description:"更新时间(时间戳)"`             // 更新时间(时间戳)
	DeleteTime    uint   `json:"deleteTime"    orm:"delete_time"     description:"删除时间(时间戳)"`             // 删除时间(时间戳)
}
