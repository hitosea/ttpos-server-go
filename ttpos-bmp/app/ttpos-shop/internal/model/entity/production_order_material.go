// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// ProductionOrderMaterial is the golang structure for table production_order_material.
type ProductionOrderMaterial struct {
	Id                         uint   `json:"id"                         orm:"id"                            description:"自增ID"`                           // 自增ID
	Uuid                       uint64 `json:"uuid"                       orm:"uuid"                          description:"生产订单原料ID"`                       // 生产订单原料ID
	Name                       string `json:"name"                       orm:"name"                          description:"原料名称,不随后台改变"`                    // 原料名称,不随后台改变
	MaterialUuid               uint64 `json:"materialUuid"               orm:"material_uuid"                 description:"原料ID"`                           // 原料ID
	Num                        int    `json:"num"                        orm:"num"                           description:"原料数量"`                           // 原料数量
	IsProductBom               int    `json:"isProductBom"               orm:"is_product_bom"                description:"是否为商品BOM, 0-否 1-是, 没有原料的规格商品为1"` // 是否为商品BOM, 0-否 1-是, 没有原料的规格商品为1
	Unit                       string `json:"unit"                       orm:"unit"                          description:"单位,不随后台改变"`                      // 单位,不随后台改变
	ProductionOrderProductUuid uint64 `json:"productionOrderProductUuid" orm:"production_order_product_uuid" description:"生产订单商品ID"`                       // 生产订单商品ID
	SaleOrderProductUuid       uint64 `json:"saleOrderProductUuid"       orm:"sale_order_product_uuid"       description:"销售订单商品ID"`                       // 销售订单商品ID
	CreateTime                 uint   `json:"createTime"                 orm:"create_time"                   description:"创建时间(时间戳)"`                      // 创建时间(时间戳)
	UpdateTime                 uint   `json:"updateTime"                 orm:"update_time"                   description:"更新时间(时间戳)"`                      // 更新时间(时间戳)
	DeleteTime                 uint   `json:"deleteTime"                 orm:"delete_time"                   description:"删除时间(时间戳)"`                      // 删除时间(时间戳)
}
