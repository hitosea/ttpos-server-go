// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// WarehouseOutForm is the golang structure for table warehouse_out_form.
type WarehouseOutForm struct {
	Id                  uint   `json:"id"                  orm:"id"                    description:"自增ID"`                                                             // 自增ID
	Uuid                uint64 `json:"uuid"                orm:"uuid"                  description:"出库单uuid"`                                                          // 出库单uuid
	FormNo              string `json:"formNo"              orm:"form_no"               description:"编号"`                                                               // 编号
	Scene               int    `json:"scene"               orm:"scene"                 description:"出库类型,0-sales销售出库 1-adjust调整出库 2-loss损耗出库 3-lost丢失出库 4-delete删除出库"` // 出库类型,0-sales销售出库 1-adjust调整出库 2-loss损耗出库 3-lost丢失出库 4-delete删除出库
	Remark              string `json:"remark"              orm:"remark"                description:"备注"`                                                               // 备注
	Status              int    `json:"status"              orm:"status"                description:"状态,0-success已出库 1-canceled已撤销"`                                    // 状态,0-success已出库 1-canceled已撤销
	OperatorUuid        uint64 `json:"operatorUuid"        orm:"operator_uuid"         description:"操作员uuid"`                                                          // 操作员uuid
	AssociatedOrderUuid uint64 `json:"associatedOrderUuid" orm:"associated_order_uuid" description:"关联订单uuid"`                                                         // 关联订单uuid
	RevokeTime          int    `json:"revokeTime"          orm:"revoke_time"           description:"撤销时间(时间戳)"`                                                        // 撤销时间(时间戳)
	CreateTime          uint   `json:"createTime"          orm:"create_time"           description:"创建时间(时间戳)"`                                                        // 创建时间(时间戳)
	UpdateTime          uint   `json:"updateTime"          orm:"update_time"           description:"更新时间(时间戳)"`                                                        // 更新时间(时间戳)
	DeleteTime          uint   `json:"deleteTime"          orm:"delete_time"           description:"删除时间(时间戳)"`                                                        // 删除时间(时间戳)
}
