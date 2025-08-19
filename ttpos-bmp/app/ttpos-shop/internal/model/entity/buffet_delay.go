// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// BuffetDelay is the golang structure for table buffet_delay.
type BuffetDelay struct {
	Id         uint    `json:"id"         orm:"id"          description:"自增ID"`         // 自增ID
	Uuid       uint64  `json:"uuid"       orm:"uuid"        description:"自助餐加钟价格ID"`    // 自助餐加钟价格ID
	Name       string  `json:"name"       orm:"name"        description:"自助餐加钟商品名称"`    // 自助餐加钟商品名称
	DelayTime  int     `json:"delayTime"  orm:"delay_time"  description:"加钟时间(分钟)"`     // 加钟时间(分钟)
	Price      float64 `json:"price"      orm:"price"       description:"价格"`           // 价格
	Status     int     `json:"status"     orm:"status"      description:"状态 0-禁用 1-启用"` // 状态 0-禁用 1-启用
	CreateTime uint    `json:"createTime" orm:"create_time" description:"创建时间(时间戳)"`    // 创建时间(时间戳)
	UpdateTime uint    `json:"updateTime" orm:"update_time" description:"更新时间(时间戳)"`    // 更新时间(时间戳)
	DeleteTime uint    `json:"deleteTime" orm:"delete_time" description:"删除时间(时间戳)"`    // 删除时间(时间戳)
}
