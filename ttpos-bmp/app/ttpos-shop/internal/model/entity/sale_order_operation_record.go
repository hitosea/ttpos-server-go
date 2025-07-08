// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// SaleOrderOperationRecord is the golang structure for table sale_order_operation_record.
type SaleOrderOperationRecord struct {
	Id            uint   `json:"id"            orm:"id"              description:"自增ID"`                                              // 自增ID
	Uuid          uint64 `json:"uuid"          orm:"uuid"            description:"桌台账单记录ID"`                                          // 桌台账单记录ID
	Source        string `json:"source"        orm:"source"          description:"操作来源 cashier-收银端 assistant-点餐助手 shop-商家后台 h5-扫码点餐"` // 操作来源 cashier-收银端 assistant-点餐助手 shop-商家后台 h5-扫码点餐
	Action        string `json:"action"        orm:"action"          description:"操作行为"`                                              // 操作行为
	Data          string `json:"data"          orm:"data"            description:"数据"`                                                // 数据
	Remark        string `json:"remark"        orm:"remark"          description:"备注"`                                                // 备注
	SaleBillUuid  uint64 `json:"saleBillUuid"  orm:"sale_bill_uuid"  description:"销售账单ID"`                                            // 销售账单ID
	SaleOrderUuid uint64 `json:"saleOrderUuid" orm:"sale_order_uuid" description:"销售订单ID"`                                            // 销售订单ID
	H5OrderUuid   uint64 `json:"h5OrderUuid"   orm:"h5_order_uuid"   description:"h5订单Uuid"`                                          // h5订单Uuid
	OperatorUuid  uint64 `json:"operatorUuid"  orm:"operator_uuid"   description:"操作员ID"`                                             // 操作员ID
	CreateTime    uint   `json:"createTime"    orm:"create_time"     description:"创建时间(时间戳)"`                                         // 创建时间(时间戳)
	UpdateTime    uint   `json:"updateTime"    orm:"update_time"     description:"更新时间(时间戳)"`                                         // 更新时间(时间戳)
	DeleteTime    uint   `json:"deleteTime"    orm:"delete_time"     description:"删除时间(时间戳)"`                                         // 删除时间(时间戳)
}
