// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// WarehouseForm is the golang structure for table warehouse_form.
type WarehouseForm struct {
	Id                uint   `json:"id"                orm:"id"                  description:"自增ID"`                                              // 自增ID
	Uuid              uint64 `json:"uuid"              orm:"uuid"                description:"库存入库单ID"`                                           // 库存入库单ID
	FormNo            string `json:"formNo"            orm:"form_no"             description:"编号"`                                                // 编号
	Scene             int    `json:"scene"             orm:"scene"               description:"交易类型,0-purchase采购入库 1-add添加入库 2-adjust调整入库 3-退菜入库"` // 交易类型,0-purchase采购入库 1-add添加入库 2-adjust调整入库 3-退菜入库
	Num               int    `json:"num"               orm:"num"                 description:"数量"`                                                // 数量
	Remark            string `json:"remark"            orm:"remark"              description:"备注"`                                                // 备注
	Status            int    `json:"status"            orm:"status"              description:"状态,0-success已入库 1-canceled已撤销"`                     // 状态,0-success已入库 1-canceled已撤销
	ProductBomUuid    uint64 `json:"productBomUuid"    orm:"product_bom_uuid"    description:"商品BOM表uuid"`                                        // 商品BOM表uuid
	MaterialUuid      uint64 `json:"materialUuid"      orm:"material_uuid"       description:"材料uuid"`                                            // 材料uuid
	PurchaseOrderUuid uint64 `json:"purchaseOrderUuid" orm:"purchase_order_uuid" description:"采购订单uuid"`                                          // 采购订单uuid
	OperatorUuid      uint64 `json:"operatorUuid"      orm:"operator_uuid"       description:"操作员uuid"`                                           // 操作员uuid
	RevokeTime        int    `json:"revokeTime"        orm:"revoke_time"         description:"撤销时间(时间戳)"`                                         // 撤销时间(时间戳)
	CreateTime        uint   `json:"createTime"        orm:"create_time"         description:"创建时间(时间戳)"`                                         // 创建时间(时间戳)
	UpdateTime        uint   `json:"updateTime"        orm:"update_time"         description:"更新时间(时间戳)"`                                         // 更新时间(时间戳)
	DeleteTime        uint   `json:"deleteTime"        orm:"delete_time"         description:"删除时间(时间戳)"`                                         // 删除时间(时间戳)
}
