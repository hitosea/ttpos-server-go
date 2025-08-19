// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// WarehouseOutFormItem is the golang structure for table warehouse_out_form_item.
type WarehouseOutFormItem struct {
	Id                   uint    `json:"id"                   orm:"id"                      description:"自增ID"`                                                             // 自增ID
	Uuid                 uint64  `json:"uuid"                 orm:"uuid"                    description:"出库单明细uuid"`                                                        // 出库单明细uuid
	Num                  float64 `json:"num"                  orm:"num"                     description:"数量"`                                                               // 数量
	Scene                int     `json:"scene"                orm:"scene"                   description:"场景,0-sales销售 1-adjust调整 2-loss损耗 3-lost丢失 4-delete删除"`             // 场景,0-sales销售 1-adjust调整 2-loss损耗 3-lost丢失 4-delete删除
	Status               int     `json:"status"               orm:"status"                  description:"状态,0-预出库 1-已出库。预出库时，表示库存扣减但未在出库记录页面显示.已出库时才在出库记录页面显示"`             // 状态,0-预出库 1-已出库。预出库时，表示库存扣减但未在出库记录页面显示.已出库时才在出库记录页面显示
	ReduceStock          int     `json:"reduceStock"          orm:"reduce_stock"            description:"是否已经减库存,0-未减库存 1-已减库存。用于判断该出库记录是否已经将对应的货物减库存，若没减库存将在下次检查时减该货物的库存"` // 是否已经减库存,0-未减库存 1-已减库存。用于判断该出库记录是否已经将对应的货物减库存，若没减库存将在下次检查时减该货物的库存
	RevokeTime           int     `json:"revokeTime"           orm:"revoke_time"             description:"撤销时间(时间戳)"`                                                        // 撤销时间(时间戳)
	MaterialUuid         uint64  `json:"materialUuid"         orm:"material_uuid"           description:"材料uuid"`                                                           // 材料uuid
	WarehouseOutFormUuid uint64  `json:"warehouseOutFormUuid" orm:"warehouse_out_form_uuid" description:"出库单uuid"`                                                          // 出库单uuid
	ProductBomUuid       uint64  `json:"productBomUuid"       orm:"product_bom_uuid"        description:"商品BOM表uuid"`                                                       // 商品BOM表uuid
	SaleOrderProductUuid uint64  `json:"saleOrderProductUuid" orm:"sale_order_product_uuid" description:"销售订单商品uuid,用于结账完成时判断订单的每个商品是否都已有对应的出库记录"`                          // 销售订单商品uuid,用于结账完成时判断订单的每个商品是否都已有对应的出库记录
	SaleOrderUuid        uint64  `json:"saleOrderUuid"        orm:"sale_order_uuid"         description:"销售订单uuid,用于结账完成时判断订单的每个商品是否都已有对应的出库记录"`                            // 销售订单uuid,用于结账完成时判断订单的每个商品是否都已有对应的出库记录
	SaleBillUuid         uint64  `json:"saleBillUuid"         orm:"sale_bill_uuid"          description:"销售账单uuid,用于结账完成时判断订单的每个商品是否都已有对应的出库记录"`                            // 销售账单uuid,用于结账完成时判断订单的每个商品是否都已有对应的出库记录
	CreateTime           uint    `json:"createTime"           orm:"create_time"             description:"创建时间(时间戳)"`                                                        // 创建时间(时间戳)
	UpdateTime           uint    `json:"updateTime"           orm:"update_time"             description:"更新时间(时间戳)"`                                                        // 更新时间(时间戳)
	DeleteTime           uint    `json:"deleteTime"           orm:"delete_time"             description:"删除时间(时间戳)"`                                                        // 删除时间(时间戳)
}
