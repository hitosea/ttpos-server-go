// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// SaleOrderDiscountStrategy is the golang structure for table sale_order_discount_strategy.
type SaleOrderDiscountStrategy struct {
	Id            uint    `json:"id"            orm:"id"              description:"自增ID"`                 // 自增ID
	Uuid          uint64  `json:"uuid"          orm:"uuid"            description:"销售订单优惠策略ID"`           // 销售订单优惠策略ID
	Type          int     `json:"type"          orm:"type"            description:"优惠策略类型,0-整单折扣、1-会员折扣"` // 优惠策略类型,0-整单折扣、1-会员折扣
	Name          string  `json:"name"          orm:"name"            description:"优惠策略名称"`               // 优惠策略名称
	Value         float64 `json:"value"         orm:"value"           description:"优惠策略值"`                // 优惠策略值
	JsonField     string  `json:"jsonField"     orm:"json_field"      description:"JSON字段"`               // JSON字段
	SaleOrderUuid uint64  `json:"saleOrderUuid" orm:"sale_order_uuid" description:"销售订单ID"`               // 销售订单ID
	CreateTime    uint    `json:"createTime"    orm:"create_time"     description:"创建时间(时间戳)"`            // 创建时间(时间戳)
	UpdateTime    uint    `json:"updateTime"    orm:"update_time"     description:"更新时间(时间戳)"`            // 更新时间(时间戳)
	DeleteTime    uint    `json:"deleteTime"    orm:"delete_time"     description:"删除时间(时间戳)"`            // 删除时间(时间戳)
}
