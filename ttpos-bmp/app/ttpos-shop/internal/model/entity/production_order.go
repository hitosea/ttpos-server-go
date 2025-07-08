// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// ProductionOrder is the golang structure for table production_order.
type ProductionOrder struct {
	Id            uint   `json:"id"            orm:"id"              description:"自增ID"`                                                                 // 自增ID
	Uuid          uint64 `json:"uuid"          orm:"uuid"            description:"生产订单ID"`                                                               // 生产订单ID
	DeskUuid      uint64 `json:"deskUuid"      orm:"desk_uuid"       description:"桌台ID"`                                                                 // 桌台ID
	SaleOrderUuid uint64 `json:"saleOrderUuid" orm:"sale_order_uuid" description:"销售订单ID"`                                                               // 销售订单ID
	SaleBillUuid  uint64 `json:"saleBillUuid"  orm:"sale_bill_uuid"  description:"销售账单ID"`                                                               // 销售账单ID
	Source        string `json:"source"        orm:"source"          description:"操作来源 shop-商家、cashier-收银机、tablet-平板端、kitchen-厨显端、assistant-点餐助手、h5-H5"` // 操作来源 shop-商家、cashier-收银机、tablet-平板端、kitchen-厨显端、assistant-点餐助手、h5-H5
	CreateTime    uint   `json:"createTime"    orm:"create_time"     description:"创建时间(时间戳)"`                                                            // 创建时间(时间戳)
	UpdateTime    uint   `json:"updateTime"    orm:"update_time"     description:"更新时间(时间戳)"`                                                            // 更新时间(时间戳)
	DeleteTime    uint   `json:"deleteTime"    orm:"delete_time"     description:"删除时间(时间戳)"`                                                            // 删除时间(时间戳)
}
