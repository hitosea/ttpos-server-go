// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// SaleOrderProductAttribute is the golang structure for table sale_order_product_attribute.
type SaleOrderProductAttribute struct {
	Id                   uint   `json:"id"                   orm:"id"                      description:"自增ID"`          // 自增ID
	Uuid                 uint64 `json:"uuid"                 orm:"uuid"                    description:"商品属性ID"`        // 商品属性ID
	Name                 string `json:"name"                 orm:"name"                    description:"商品属性名称,不随后台更新"` // 商品属性名称,不随后台更新
	SaleOrderUuid        uint64 `json:"saleOrderUuid"        orm:"sale_order_uuid"         description:"销售订单ID"`        // 销售订单ID
	SaleOrderProductUuid uint64 `json:"saleOrderProductUuid" orm:"sale_order_product_uuid" description:"销售订单商品ID"`      // 销售订单商品ID
	ProductAttributeUuid uint64 `json:"productAttributeUuid" orm:"product_attribute_uuid"  description:"商品属性ID"`        // 商品属性ID
	CreateTime           uint   `json:"createTime"           orm:"create_time"             description:"创建时间(时间戳)"`     // 创建时间(时间戳)
	UpdateTime           uint   `json:"updateTime"           orm:"update_time"             description:"更新时间(时间戳)"`     // 更新时间(时间戳)
	DeleteTime           uint   `json:"deleteTime"           orm:"delete_time"             description:"删除时间(时间戳)"`     // 删除时间(时间戳)
}
