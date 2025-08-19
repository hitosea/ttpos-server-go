// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// SaleOrderProductBom is the golang structure for table sale_order_product_bom.
type SaleOrderProductBom struct {
	Id                   uint    `json:"id"                   orm:"id"                      description:"自增ID"`                            // 自增ID
	Uuid                 uint64  `json:"uuid"                 orm:"uuid"                    description:"销售订单商品规格或小料ID"`                   // 销售订单商品规格或小料ID
	Name                 string  `json:"name"                 orm:"name"                    description:"规格或小料名称,不随后台更新"`                  // 规格或小料名称,不随后台更新
	Price                float64 `json:"price"                orm:"price"                   description:"单价,不随后台更新，记录加购时的价格。结账时要校验价格是否变动"` // 单价,不随后台更新，记录加购时的价格。结账时要校验价格是否变动
	SaleOrderUuid        uint64  `json:"saleOrderUuid"        orm:"sale_order_uuid"         description:"销售订单ID"`                          // 销售订单ID
	SaleOrderProductUuid uint64  `json:"saleOrderProductUuid" orm:"sale_order_product_uuid" description:"销售订单商品ID"`                        // 销售订单商品ID
	ProductBomUuid       uint64  `json:"productBomUuid"       orm:"product_bom_uuid"        description:"商品BOM ID"`                        // 商品BOM ID
	IsFlavorBom          int     `json:"isFlavorBom"          orm:"is_flavor_bom"           description:"是否为规格商品BOM, 0-否 1-是"`             // 是否为规格商品BOM, 0-否 1-是
	CreateTime           uint    `json:"createTime"           orm:"create_time"             description:"创建时间(时间戳)"`                       // 创建时间(时间戳)
	UpdateTime           uint    `json:"updateTime"           orm:"update_time"             description:"更新时间(时间戳)"`                       // 更新时间(时间戳)
	DeleteTime           uint    `json:"deleteTime"           orm:"delete_time"             description:"删除时间(时间戳)"`                       // 删除时间(时间戳)
}
