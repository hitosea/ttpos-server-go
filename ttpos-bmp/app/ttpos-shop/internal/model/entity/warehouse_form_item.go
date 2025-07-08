// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// WarehouseFormItem is the golang structure for table warehouse_form_item.
type WarehouseFormItem struct {
	Id                   uint    `json:"id"                   orm:"id"                      description:"自增ID"`                                                             // 自增ID
	Uuid                 uint64  `json:"uuid"                 orm:"uuid"                    description:"入库单明细uuid"`                                                        // 入库单明细uuid
	Num                  float64 `json:"num"                  orm:"num"                     description:"入库数量"`                                                             // 入库数量
	Scene                int     `json:"scene"                orm:"scene"                   description:"场景,0-采购 1-添加入库 2-调整入库 3-退菜入库、反结账入库,这个场景不显示在入库记录页面"`                // 场景,0-采购 1-添加入库 2-调整入库 3-退菜入库、反结账入库,这个场景不显示在入库记录页面
	AddStock             int     `json:"addStock"             orm:"add_stock"               description:"是否已经加库存,0-未加库存 1-已加库存。用于判断该入库记录是否已经将对应的货物加库存，若没加库存将在下次检查时加该货物的库存"` // 是否已经加库存,0-未加库存 1-已加库存。用于判断该入库记录是否已经将对应的货物加库存，若没加库存将在下次检查时加该货物的库存
	MaterialUuid         uint64  `json:"materialUuid"         orm:"material_uuid"           description:"材料uuid"`                                                           // 材料uuid
	ProductBomUuid       uint64  `json:"productBomUuid"       orm:"product_bom_uuid"        description:"商品BOM表uuid"`                                                       // 商品BOM表uuid
	WarehouseFormUuid    uint64  `json:"warehouseFormUuid"    orm:"warehouse_form_uuid"     description:"入库单uuid"`                                                          // 入库单uuid
	SaleOrderProductUuid uint64  `json:"saleOrderProductUuid" orm:"sale_order_product_uuid" description:"销售订单商品uuid,用于退菜入库"`                                                // 销售订单商品uuid,用于退菜入库
	SaleBillUuid         uint64  `json:"saleBillUuid"         orm:"sale_bill_uuid"          description:"销售账单uuid,用于退菜入库"`                                                  // 销售账单uuid,用于退菜入库
	CreateTime           uint    `json:"createTime"           orm:"create_time"             description:"创建时间(时间戳)"`                                                        // 创建时间(时间戳)
	UpdateTime           uint    `json:"updateTime"           orm:"update_time"             description:"更新时间(时间戳)"`                                                        // 更新时间(时间戳)
	DeleteTime           uint    `json:"deleteTime"           orm:"delete_time"             description:"删除时间(时间戳)"`                                                        // 删除时间(时间戳)
}
